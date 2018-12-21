#ifndef _LIST_LIBS_WINDOWS_H_
#define _LIST_LIBS_WINDOWS_H_

#include <windows.h>
#include <psapi.h>
#include "../process/process.h"

#include <windows.h>

/**
 * Windows specific process handle.
 *
 * NOTE: We use uintptr_t instead of HANDLE because Go doesn't allow
 * pointers with invalid values. Windows' HANDLE is a PVOID internally and
 * sometimes it is used as an integer.
 **/
typedef uintptr_t process_handle_t;

typedef struct t_ModuleInfo {
    char *filename;
    MODULEINFO info;
} ModuleInfo;

typedef struct t_EnumProcessModulesResponse {
    DWORD error;
    DWORD length;
    ModuleInfo *modules;
} EnumProcessModulesResponse;

EnumProcessModulesResponse *getModules(process_handle_t handle);
void EnumProcessModulesResponse_Free(EnumProcessModulesResponse *r);

#endif // _LIST_LIBS_WINDOWS_H_
