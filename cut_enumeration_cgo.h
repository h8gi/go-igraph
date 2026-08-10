#ifndef GO_IGRAPH_CUT_ENUMERATION_CGO_H
#define GO_IGRAPH_CUT_ENUMERATION_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_all_st_cuts(
    const igraph_t *graph,
    igraph_vector_int_list_t *cuts,
    igraph_vector_int_list_t *partition1s,
    igraph_int_t source,
    igraph_int_t target);

igraph_error_t go_igraph_all_st_mincuts(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_list_t *cuts,
    igraph_vector_int_list_t *partition1s,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity);

#endif
