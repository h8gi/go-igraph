#include "cut_enumeration_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_all_st_cuts(
    const igraph_t *graph,
    igraph_vector_int_list_t *cuts,
    igraph_vector_int_list_t *partition1s,
    igraph_int_t source,
    igraph_int_t target) {
    GO_IGRAPH_CALL(igraph_all_st_cuts(graph, cuts, partition1s, source, target));
}

igraph_error_t go_igraph_all_st_mincuts(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_list_t *cuts,
    igraph_vector_int_list_t *partition1s,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_all_st_mincuts(graph, value, cuts, partition1s, source, target, capacity));
}
