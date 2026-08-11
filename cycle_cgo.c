#include "cycle_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_acyclic(
        const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_acyclic(graph, result));
}

igraph_error_t go_igraph_is_dag(
        const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_dag(graph, result));
}

igraph_error_t go_igraph_topological_sorting(
        const igraph_t *graph, igraph_vector_int_t *result,
        igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_topological_sorting(graph, result, mode));
}

igraph_error_t go_igraph_find_cycle(
        const igraph_t *graph, igraph_vector_int_t *vertices,
        igraph_vector_int_t *edges, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_find_cycle(graph, vertices, edges, mode));
}

igraph_error_t go_igraph_girth(
        const igraph_t *graph, igraph_real_t *girth,
        igraph_vector_int_t *vertices) {
    GO_IGRAPH_CALL(igraph_girth(graph, girth, vertices));
}
