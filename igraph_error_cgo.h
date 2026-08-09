#ifndef GO_IGRAPH_ERROR_CGO_H
#define GO_IGRAPH_ERROR_CGO_H

#include <igraph.h>

/*
 * The pinned thread-safe igraph build stores handlers in thread-local state.
 * Installing and restoring them around one upstream operation in one cgo call
 * keeps handler state on the same OS thread and turns igraph failures into
 * return codes instead of process aborts.
 */
#define GO_IGRAPH_CALL(expression)                                        \
    do {                                                                  \
        igraph_error_handler_t *old_error =                               \
            igraph_set_error_handler(&igraph_error_handler_ignore);       \
        igraph_warning_handler_t *old_warning =                           \
            igraph_set_warning_handler(&igraph_warning_handler_ignore);   \
        igraph_error_t code = (expression);                               \
        igraph_set_warning_handler(old_warning);                          \
        igraph_set_error_handler(old_error);                              \
        return code;                                                      \
    } while (0)

#endif
