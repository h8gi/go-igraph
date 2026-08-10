#ifndef GO_IGRAPH_SPECTRAL_COMMUNITY_CGO_H
#define GO_IGRAPH_SPECTRAL_COMMUNITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_community_leading_eigenvector(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_matrix_int_t *merges,
    igraph_vector_int_t *membership,
    igraph_int_t steps,
    int max_iterations,
    igraph_real_t tolerance,
    igraph_real_t *modularity,
    igraph_bool_t start);

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
    igraph_real_t lambda);

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
    igraph_real_t gamma);

igraph_error_t go_igraph_community_optimal_modularity(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_real_t *modularity,
    igraph_vector_int_t *membership);

#endif
