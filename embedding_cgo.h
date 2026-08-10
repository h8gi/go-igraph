#ifndef GO_IGRAPH_EMBEDDING_CGO_H
#define GO_IGRAPH_EMBEDDING_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_adjacency_spectral_embedding(
    const igraph_t *graph,
    igraph_integer_t no,
    const igraph_vector_t *weights,
    igraph_eigen_which_position_t which,
    igraph_bool_t scaled,
    igraph_matrix_t *x,
    igraph_matrix_t *y,
    igraph_vector_t *d,
    const igraph_vector_t *cvec,
    int max_iterations,
    igraph_real_t tolerance);

igraph_error_t go_igraph_laplacian_spectral_embedding(
    const igraph_t *graph,
    igraph_integer_t no,
    const igraph_vector_t *weights,
    igraph_eigen_which_position_t which,
    igraph_laplacian_spectral_embedding_type_t type,
    igraph_bool_t scaled,
    igraph_matrix_t *x,
    igraph_matrix_t *y,
    igraph_vector_t *d,
    int max_iterations,
    igraph_real_t tolerance);

igraph_error_t go_igraph_dim_select(
    const igraph_vector_t *sv,
    igraph_integer_t *dim);

#endif
