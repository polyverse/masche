#ifndef LISTLIBS_DARWIN_H

#include "../cresponse/response.h"

#include <stdint.h>

#include "../cresponse/response.h"

#include <mach/mach.h>

// Mac specific process handle.
typedef task_t process_handle_t;


/**
 * Returns a dynamically allocated list of absolute paths (as null-terminated
 * strings) to the libraries loaded by the process.
 **/
response_t *list_loaded_libraries(process_handle_t handle, char ***libs,
        size_t *count);

/**
 * Frees the list allocated by the previous function.
 **/
void free_loaded_libraries_list(char **list, size_t count);

#endif /* LISTLIBS_DARWIN_H */
