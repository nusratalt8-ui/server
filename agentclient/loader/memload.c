#include "memload.h"
#include <stdlib.h>
#include <string.h>

typedef struct {
    BYTE   *base;
    SIZE_T  size;
} MOD;

static DWORD sec_prot(DWORD c) {
    if (c & IMAGE_SCN_MEM_EXECUTE) {
        if (c & IMAGE_SCN_MEM_WRITE) return PAGE_EXECUTE_READWRITE;
        if (c & IMAGE_SCN_MEM_READ)  return PAGE_EXECUTE_READ;
        return PAGE_EXECUTE;
    }
    if (c & IMAGE_SCN_MEM_WRITE) return PAGE_READWRITE;
    return PAGE_READONLY;
}

void *mod_load(const void *data, size_t size) {
    if (!data || size < sizeof(IMAGE_DOS_HEADER)) return NULL;
    IMAGE_DOS_HEADER   *dos = (IMAGE_DOS_HEADER *)data;
    if (dos->e_magic != IMAGE_DOS_SIGNATURE) return NULL;
    if (size < dos->e_lfanew + sizeof(IMAGE_NT_HEADERS)) return NULL;
    IMAGE_NT_HEADERS   *nt  = (IMAGE_NT_HEADERS *)((BYTE *)data + dos->e_lfanew);
    if (nt->Signature  != IMAGE_NT_SIGNATURE)  return NULL;
    if (!(nt->FileHeader.Characteristics & IMAGE_FILE_EXECUTABLE_IMAGE)) return NULL;

    BYTE *base = (BYTE *)VirtualAlloc(
        (LPVOID)(ULONG_PTR)nt->OptionalHeader.ImageBase,
        nt->OptionalHeader.SizeOfImage,
        MEM_RESERVE | MEM_COMMIT, PAGE_READWRITE);
    if (!base)
        base = (BYTE *)VirtualAlloc(NULL,
            nt->OptionalHeader.SizeOfImage,
            MEM_RESERVE | MEM_COMMIT, PAGE_READWRITE);
    if (!base) return NULL;

    memcpy(base, data, nt->OptionalHeader.SizeOfHeaders);

    IMAGE_SECTION_HEADER *sec = IMAGE_FIRST_SECTION(nt);
    for (int i = 0; i < nt->FileHeader.NumberOfSections; i++, sec++) {
        if (!sec->SizeOfRawData) continue;
        memcpy(base + sec->VirtualAddress,
               (BYTE *)data + sec->PointerToRawData,
               sec->SizeOfRawData);
    }

    ULONG_PTR delta = (ULONG_PTR)base - (ULONG_PTR)nt->OptionalHeader.ImageBase;
    if (delta) {
        IMAGE_DATA_DIRECTORY *rd = &nt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_BASERELOC];
        if (rd->Size > 0 && rd->VirtualAddress > 0) {
            IMAGE_BASE_RELOCATION *r = (IMAGE_BASE_RELOCATION *)(base + rd->VirtualAddress);
            while (r->VirtualAddress) {
                if (!r->SizeOfBlock) break;
                WORD *e = (WORD *)((BYTE *)r + sizeof(IMAGE_BASE_RELOCATION));
                int n = (r->SizeOfBlock - sizeof(IMAGE_BASE_RELOCATION)) / sizeof(WORD);
                for (int i = 0; i < n; i++) {
                    int t = e[i] >> 12, off = e[i] & 0xFFF;
                    if (t == IMAGE_REL_BASED_DIR64)
                        *(ULONG_PTR *)(base + r->VirtualAddress + off) += delta;
                    else if (t == IMAGE_REL_BASED_HIGHLOW)
                        *(DWORD *)(base + r->VirtualAddress + off) += (DWORD)delta;
                }
                r = (IMAGE_BASE_RELOCATION *)((BYTE *)r + r->SizeOfBlock);
            }
        }
    }

    IMAGE_DATA_DIRECTORY *id = &nt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_IMPORT];
    if (id->Size) {
        IMAGE_IMPORT_DESCRIPTOR *imp = (IMAGE_IMPORT_DESCRIPTOR *)(base + id->VirtualAddress);
        for (; imp->Name; imp++) {
            HMODULE lib = LoadLibraryA((LPCSTR)(base + imp->Name));
            if (!lib) { VirtualFree(base, 0, MEM_RELEASE); return NULL; }
            ULONG_PTR *oft = (ULONG_PTR *)(base + (imp->OriginalFirstThunk ? imp->OriginalFirstThunk : imp->FirstThunk));
            ULONG_PTR *ft  = (ULONG_PTR *)(base + imp->FirstThunk);
            for (; *oft; oft++, ft++) {
                FARPROC fn;
                if (*oft & IMAGE_ORDINAL_FLAG64)
                    fn = GetProcAddress(lib, (LPCSTR)IMAGE_ORDINAL64(*oft));
                else
                    fn = GetProcAddress(lib, ((IMAGE_IMPORT_BY_NAME *)(base + *oft))->Name);
                if (!fn) { VirtualFree(base, 0, MEM_RELEASE); return NULL; }
                *ft = (ULONG_PTR)fn;
            }
        }
    }

    sec = IMAGE_FIRST_SECTION(nt);
    for (int i = 0; i < nt->FileHeader.NumberOfSections; i++, sec++) {
        if (!sec->SizeOfRawData) continue;
        DWORD old;
        VirtualProtect(base + sec->VirtualAddress, sec->SizeOfRawData, sec_prot(sec->Characteristics), &old);
    }

    typedef BOOL (WINAPI *DllEntry)(HINSTANCE, DWORD, LPVOID);
    ((DllEntry)(base + nt->OptionalHeader.AddressOfEntryPoint))((HINSTANCE)base, DLL_PROCESS_ATTACH, NULL);

    MOD *m = (MOD *)malloc(sizeof(MOD));
    m->base = base;
    m->size = nt->OptionalHeader.SizeOfImage;
    return m;
}

FARPROC mod_sym(void *mod, const char *name) {
    if (!mod) return NULL;
    BYTE *base = ((MOD *)mod)->base;
    IMAGE_NT_HEADERS *nt = (IMAGE_NT_HEADERS *)(base + ((IMAGE_DOS_HEADER *)base)->e_lfanew);
    IMAGE_DATA_DIRECTORY *ed = &nt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT];
    if (!ed->Size) return NULL;
    IMAGE_EXPORT_DIRECTORY *exp = (IMAGE_EXPORT_DIRECTORY *)(base + ed->VirtualAddress);
    DWORD *names = (DWORD *)(base + exp->AddressOfNames);
    WORD  *ords  = (WORD  *)(base + exp->AddressOfNameOrdinals);
    DWORD *funcs = (DWORD *)(base + exp->AddressOfFunctions);
    for (DWORD i = 0; i < exp->NumberOfNames; i++)
        if (!strcmp((char *)(base + names[i]), name))
            return (FARPROC)(base + funcs[ords[i]]);
    return NULL;
}

void mod_free(void *mod) {
    if (!mod) return;
    MOD *m = (MOD *)mod;
    IMAGE_NT_HEADERS *nt = (IMAGE_NT_HEADERS *)(m->base + ((IMAGE_DOS_HEADER *)m->base)->e_lfanew);
    typedef BOOL (WINAPI *DllEntry)(HINSTANCE, DWORD, LPVOID);
    ((DllEntry)(m->base + nt->OptionalHeader.AddressOfEntryPoint))((HINSTANCE)m->base, DLL_PROCESS_DETACH, NULL);
    VirtualFree(m->base, 0, MEM_RELEASE);
    free(m);
}