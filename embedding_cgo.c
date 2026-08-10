#include "embedding_cgo.h"
#include "igraph_error_cgo.h"

static void go_igraph_arpack_options_embedding(
    igraph_arpack_options_t *options, int max_iterations,
    igraph_real_t tolerance) {
    igraph_arpack_options_init(options);
    if (max_iterations > 0) {
        options->mxiter = max_iterations;
    }
    if (tolerance > 0) {
        options->tol = tolerance;
    }
}

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
    igraph_real_t tolerance) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options_embedding(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_adjacency_spectral_embedding(
        graph, no, weights, which, scaled, x, y, d, cvec, &options));
}

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
    igraph_real_t tolerance) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options_embedding(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_laplacian_spectral_embedding(
        graph, no, weights, which, type, scaled, x, y, d, &options));
}

igraph_error_t go_igraph_dim_select(
    const igraph_vector_t *sv,
    igraph_integer_t *dim) {
    GO_IGRAPH_CALL(igraph_dim_select(sv, dim));
}
