#include "subgraph_cgo.h"

/*
 * Keep handler installation, the fallible upstream operation, and handler
 * restoration on the same OS thread and inside one cgo call.
 */
#define GO_IGRAPH_SUBGRAPH_CALL(expression)                                \
    do {                                                                    \
        igraph_error_handler_t *old_error =                                 \
            igraph_set_error_handler(&igraph_error_handler_ignore);         \
        igraph_warning_handler_t *old_warning =                             \
            igraph_set_warning_handler(&igraph_warning_handler_ignore);     \
        igraph_error_t code = (expression);                                 \
        igraph_set_warning_handler(old_warning);                            \
        igraph_set_error_handler(old_error);                                \
        return code;                                                        \
    } while (0)

igraph_error_t go_igraph_induced_subgraph_map(
    const igraph_t *graph, igraph_t *result, igraph_vs_t vertices,
    igraph_vector_int_t *map, igraph_vector_int_t *inverse_map) {
    GO_IGRAPH_SUBGRAPH_CALL(igraph_induced_subgraph_map(
        graph, result, vertices, IGRAPH_SUBGRAPH_AUTO, map, inverse_map));
}

igraph_error_t go_igraph_subgraph_from_edges(
    const igraph_t *graph, igraph_t *result, igraph_es_t edges,
    igraph_bool_t delete_vertices) {
    GO_IGRAPH_SUBGRAPH_CALL(igraph_subgraph_from_edges(
        graph, result, edges, delete_vertices));
}

igraph_error_t go_igraph_decompose(
    const igraph_t *graph, igraph_graph_list_t *components,
    igraph_connectedness_t mode, igraph_int_t maximum_components,
    igraph_int_t minimum_vertices) {
    GO_IGRAPH_SUBGRAPH_CALL(igraph_decompose(
        graph, components, mode, maximum_components, minimum_vertices));
}
