#include "hierarchical_community_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_community_fastgreedy(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_matrix_int_t *merges,
    igraph_vector_t *modularity,
    igraph_vector_int_t *membership) {
    GO_IGRAPH_CALL(igraph_community_fastgreedy(
        graph, weights, merges, modularity, membership));
}

igraph_error_t go_igraph_community_walktrap(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_int_t steps,
    igraph_matrix_int_t *merges,
    igraph_vector_t *modularity,
    igraph_vector_int_t *membership) {
    GO_IGRAPH_CALL(igraph_community_walktrap(
        graph, weights, steps, merges, modularity, membership));
}

igraph_error_t go_igraph_community_edge_betweenness(
    const igraph_t *graph,
    igraph_vector_int_t *removed_edges,
    igraph_vector_t *edge_betweenness,
    igraph_matrix_int_t *merges,
    igraph_vector_int_t *bridges,
    igraph_vector_t *modularity,
    igraph_vector_int_t *membership,
    igraph_bool_t directed,
    const igraph_vector_t *weights,
    const igraph_vector_t *lengths) {
    GO_IGRAPH_CALL(igraph_community_edge_betweenness(
        graph, removed_edges, edge_betweenness, merges, bridges, modularity, membership, directed, weights, lengths));
}

igraph_error_t go_igraph_community_eb_get_merges(
    const igraph_t *graph,
    igraph_bool_t directed,
    const igraph_vector_int_t *edges,
    const igraph_vector_t *weights,
    igraph_matrix_int_t *merges,
    igraph_vector_int_t *bridges,
    igraph_vector_t *modularity,
    igraph_vector_int_t *membership) {
    GO_IGRAPH_CALL(igraph_community_eb_get_merges(
        graph, directed, edges, weights, merges, bridges, modularity, membership));
}
