#ifndef GO_IGRAPH_GRAPHLETS_CGO_H
#define GO_IGRAPH_GRAPHLETS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_graphlets_is_simple(
    const igraph_t *graph,
    igraph_bool_t *simple);

igraph_error_t go_igraph_graphlets(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_list_t *cliques,
    igraph_vector_t *mu,
    igraph_int_t niter);

igraph_error_t go_igraph_graphlets_candidate_basis(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_list_t *cliques,
    igraph_vector_t *thresholds);

igraph_error_t go_igraph_graphlets_project(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    const igraph_vector_int_list_t *cliques,
    igraph_vector_t *mu,
    igraph_bool_t start_mu,
    igraph_int_t niter);

#endif
