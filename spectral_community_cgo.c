#include "spectral_community_cgo.h"
#include "igraph_error_cgo.h"

static void go_igraph_arpack_options_local(
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

igraph_error_t go_igraph_community_leading_eigenvector(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_matrix_int_t *merges,
    igraph_vector_int_t *membership,
    igraph_int_t steps,
    int max_iterations,
    igraph_real_t tolerance,
    igraph_real_t *modularity,
    igraph_bool_t start) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options_local(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_community_leading_eigenvector(
        graph, weights, merges, membership, steps, &options, modularity, start,
        NULL, NULL, NULL, NULL, NULL));
}

igraph_error_t go_igraph_community_spinglass(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t *modularity,
    igraph_real_t *temperature,
    igraph_vector_int_t *membership,
    igraph_vector_int_t *csize,
    igraph_int_t spins,
    igraph_bool_t parupdate,
    igraph_real_t starttemp,
    igraph_real_t stoptemp,
    igraph_real_t coolfact,
    igraph_spincomm_update_t update_rule,
    igraph_real_t gamma,
    igraph_spinglass_implementation_t implementation,
    igraph_real_t lambda) {
    GO_IGRAPH_CALL(igraph_community_spinglass(
        graph, weights, modularity, temperature, membership, csize, spins,
        parupdate, starttemp, stoptemp, coolfact, update_rule, gamma,
        implementation, lambda));
}

igraph_error_t go_igraph_community_spinglass_single(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_integer_t vertex,
    igraph_vector_int_t *community,
    igraph_real_t *cohesion,
    igraph_real_t *adhesion,
    igraph_real_t *inner_links,
    igraph_real_t *outer_links,
    igraph_int_t spins,
    igraph_spincomm_update_t update_rule,
    igraph_real_t gamma) {
    GO_IGRAPH_CALL(igraph_community_spinglass_single(
        graph, weights, vertex, community, cohesion, adhesion, inner_links,
        outer_links, spins, update_rule, gamma));
}

igraph_error_t go_igraph_community_optimal_modularity(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_real_t *modularity,
    igraph_vector_int_t *membership) {
    GO_IGRAPH_CALL(igraph_community_optimal_modularity(
        graph, weights, resolution, modularity, membership));
}
