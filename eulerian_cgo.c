#include "eulerian_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_eulerian(
    const igraph_t *graph, igraph_bool_t *has_path,
    igraph_bool_t *has_cycle) {
    GO_IGRAPH_CALL(igraph_is_eulerian(graph, has_path, has_cycle));
}

igraph_error_t go_igraph_eulerian_path(
    const igraph_t *graph, igraph_vector_int_t *edges,
    igraph_vector_int_t *vertices) {
    GO_IGRAPH_CALL(igraph_eulerian_path(graph, edges, vertices));
}

igraph_error_t go_igraph_eulerian_cycle(
    const igraph_t *graph, igraph_vector_int_t *edges,
    igraph_vector_int_t *vertices) {
    GO_IGRAPH_CALL(igraph_eulerian_cycle(graph, edges, vertices));
}
