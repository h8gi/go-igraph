#include "embedding_cgo.h"
#include "algorithm_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_strength(
    const igraph_t *graph,
    igraph_vector_t *res,
    igraph_vs_t vids,
    igraph_neimode_t mode,
    igraph_loops_t loops,
    const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_strength(graph, res, vids, mode, loops, weights));
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
    go_igraph_arpack_options(&options, max_iterations, tolerance);
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
    go_igraph_arpack_options(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_laplacian_spectral_embedding(
        graph, no, weights, which, type, scaled, x, y, d, &options));
}

igraph_error_t go_igraph_dim_select(
    const igraph_vector_t *sv,
    igraph_integer_t *dim) {
    GO_IGRAPH_CALL(igraph_dim_select(sv, dim));
}
