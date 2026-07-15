#include <windows.h>
#include <dbghelp.h>
#include <stdio.h>

static LONG WINAPI crash_handler(EXCEPTION_POINTERS *ep) {
    char base[MAX_PATH];
    char dir[MAX_PATH];
    char path[MAX_PATH];
    char dmppath[MAX_PATH];

    if (!GetEnvironmentVariableA("APPDATA", base, sizeof(base))) {
        GetTempPathA(sizeof(base), base);
    }

    _snprintf(dir, sizeof(dir), "%s\\WindowsUpdate\\crashes", base);
    CreateDirectoryA(dir, NULL);

    SYSTEMTIME st;
    GetLocalTime(&st);
    _snprintf(path, sizeof(path), "%s\\crash_%04d%02d%02d_%02d%02d%02d_%03d.log",
        dir, st.wYear, st.wMonth, st.wDay,
        st.wHour, st.wMinute, st.wSecond, st.wMilliseconds);
    _snprintf(dmppath, sizeof(dmppath), "%s\\crash_%04d%02d%02d_%02d%02d%02d_%03d.dmp",
        dir, st.wYear, st.wMonth, st.wDay,
        st.wHour, st.wMinute, st.wSecond, st.wMilliseconds);

    FILE *f = fopen(path, "w");
    if (f) {
        fprintf(f, "code=0x%08lX addr=%p\n",
            ep->ExceptionRecord->ExceptionCode,
            ep->ExceptionRecord->ExceptionAddress);
        DWORD nparams = ep->ExceptionRecord->NumberParameters;
        for (DWORD i = 0; i < nparams && i < EXCEPTION_MAXIMUM_PARAMETERS; i++) {
            fprintf(f, "param[%lu]=0x%p\n", i, (void*)ep->ExceptionRecord->ExceptionInformation[i]);
        }
        fclose(f);
    }

    HANDLE hFile = CreateFileA(dmppath, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hFile != INVALID_HANDLE_VALUE) {
        MINIDUMP_EXCEPTION_INFORMATION mei = {0};
        mei.ThreadId = GetCurrentThreadId();
        mei.ExceptionPointers = ep;
        mei.ClientPointers = FALSE;
        HMODULE hDbgHelp = LoadLibraryA("dbghelp.dll");
        if (hDbgHelp) {
            typedef BOOL (WINAPI *MiniDumpWriteDump_t)(HANDLE, DWORD, HANDLE, MINIDUMP_TYPE,
                PMINIDUMP_EXCEPTION_INFORMATION, PMINIDUMP_USER_STREAM_INFORMATION,
                PMINIDUMP_CALLBACK_INFORMATION);
            MiniDumpWriteDump_t pDump = (MiniDumpWriteDump_t)GetProcAddress(hDbgHelp, "MiniDumpWriteDump");
            if (pDump) {
                pDump(GetCurrentProcess(), GetCurrentProcessId(), hFile,
                    MiniDumpWithDataSegs, &mei, NULL, NULL);
            }
        }
        CloseHandle(hFile);
    }

    return EXCEPTION_EXECUTE_HANDLER;
}

void install_crash_handler(void) {
    SetUnhandledExceptionFilter(crash_handler);
}