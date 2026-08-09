#include "subgraph_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_induced_subgraph_map(
    const igraph_t *graph, igraph_t *result, igraph_vs_t vertices,
    igraph_vector_int_t *map, igraph_vector_int_t *inverse_map) {
    GO_IGRAPH_CALL(igraph_induced_subgraph_map(
        graph, result, vertices, IGRAPH_SUBGRAPH_AUTO, map, inverse_map));
}

igraph_error_t go_igraph_subgraph_from_edges(
    const igraph_t *graph, igraph_t *result, igraph_es_t edges,
    igraph_bool_t delete_vertices) {
    GO_IGRAPH_CALL(igraph_subgraph_from_edges(
        graph, result, edges, delete_vertices));
}

igraph_error_t go_igraph_decompose(
    const igraph_t *graph, igraph_graph_list_t *components,
    igraph_connectedness_t mode, igraph_int_t maximum_components,
    igraph_int_t minimum_vertices) {
    GO_IGRAPH_CALL(igraph_decompose(
        graph, components, mode, maximum_components, minimum_vertices));
}
