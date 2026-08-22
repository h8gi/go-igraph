#ifndef GO_IGRAPH_ADJACENCY_CGO_H
#define GO_IGRAPH_ADJACENCY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_adjacency(
    igraph_t *graph, const igraph_matrix_t *matrix,
    igraph_adjacency_t mode, igraph_loops_t loops);

igraph_error_t go_igraph_weighted_adjacency(
    igraph_t *graph, const igraph_matrix_t *matrix, igraph_adjacency_t mode,
    igraph_vector_t *weights, igraph_loops_t loops);

igraph_error_t go_igraph_get_adjacency(
    const igraph_t *graph, igraph_matrix_t *matrix,
    igraph_get_adjacency_t type, const igraph_vector_t *weights,
    igraph_loops_t loops);

igraph_error_t go_igraph_get_stochastic(
    const igraph_t *graph, igraph_matrix_t *matrix,
    igraph_bool_t column_wise, const igraph_vector_t *weights);

#endif
