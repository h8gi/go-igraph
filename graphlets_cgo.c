#include "graphlets_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_graphlets_is_simple(
    const igraph_t *graph,
    igraph_bool_t *simple) {
    GO_IGRAPH_CALL(igraph_is_simple(graph, simple, IGRAPH_UNDIRECTED));
}

igraph_error_t go_igraph_graphlets(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_list_t *cliques,
    igraph_vector_t *mu,
    igraph_int_t niter) {
    GO_IGRAPH_CALL(igraph_graphlets(graph, weights, cliques, mu, niter));
}

igraph_error_t go_igraph_graphlets_candidate_basis(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_list_t *cliques,
    igraph_vector_t *thresholds) {
    GO_IGRAPH_CALL(igraph_graphlets_candidate_basis(
        graph, weights, cliques, thresholds));
}

igraph_error_t go_igraph_graphlets_project(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    const igraph_vector_int_list_t *cliques,
    igraph_vector_t *mu,
    igraph_bool_t start_mu,
    igraph_int_t niter) {
    GO_IGRAPH_CALL(igraph_graphlets_project(
        graph, weights, cliques, mu, start_mu, niter));
}
