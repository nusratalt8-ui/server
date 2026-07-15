#ifndef MEMLOAD_H
#define MEMLOAD_H
#include <windows.h>
void *mod_load(const void *data, size_t size);
FARPROC mod_sym(void *mod, const char *name);
void mod_free(void *mod);
void install_crash_handler(void);
#endif