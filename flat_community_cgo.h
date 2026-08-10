#ifndef GO_IGRAPH_FLAT_COMMUNITY_CGO_H
#define GO_IGRAPH_FLAT_COMMUNITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_community_multilevel(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_vector_int_t *membership,
    igraph_matrix_int_t *memberships,
    igraph_vector_t *modularity);

igraph_error_t go_igraph_community_leiden(
    const igraph_t *graph,
    const igraph_vector_t *edge_weights,
    const igraph_vector_t *vertex_out_weights,
    const igraph_vector_t *vertex_in_weights,
    igraph_real_t resolution,
    igraph_real_t beta,
    igraph_bool_t start,
    igraph_int_t n_iterations,
    igraph_vector_int_t *membership,
    igraph_int_t *nb_clusters,
    igraph_real_t *quality);

igraph_error_t go_igraph_community_label_propagation(
    const igraph_t *graph,
    igraph_vector_int_t *membership,
    igraph_neimode_t mode,
    const igraph_vector_t *weights,
    const igraph_vector_int_t *initial,
    const igraph_vector_bool_t *fixed,
    igraph_lpa_variant_t variant);

igraph_error_t go_igraph_community_infomap(
    const igraph_t *graph,
    const igraph_vector_t *edge_weights,
    const igraph_vector_t *vertex_weights,
    igraph_int_t nb_trials,
    igraph_bool_t is_regularized,
    igraph_real_t regularization_strength,
    igraph_vector_int_t *membership,
    igraph_real_t *codelength);

igraph_error_t go_igraph_community_fluid_communities(
    const igraph_t *graph,
    igraph_int_t no_of_communities,
    igraph_vector_int_t *membership);

#endif
