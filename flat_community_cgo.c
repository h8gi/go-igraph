#include "flat_community_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_community_multilevel(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_vector_int_t *membership,
    igraph_matrix_int_t *memberships,
    igraph_vector_t *modularity) {
    GO_IGRAPH_CALL(igraph_community_multilevel(
        graph, weights, resolution, membership, memberships, modularity));
}

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
    igraph_real_t *quality) {
    GO_IGRAPH_CALL(igraph_community_leiden(
        graph, edge_weights, vertex_out_weights, vertex_in_weights, resolution, beta, start, n_iterations, membership, nb_clusters, quality));
}

igraph_error_t go_igraph_community_label_propagation(
    const igraph_t *graph,
    igraph_vector_int_t *membership,
    igraph_neimode_t mode,
    const igraph_vector_t *weights,
    const igraph_vector_int_t *initial,
    const igraph_vector_bool_t *fixed,
    igraph_lpa_variant_t variant) {
    GO_IGRAPH_CALL(igraph_community_label_propagation(
        graph, membership, mode, weights, initial, fixed, variant));
}

igraph_error_t go_igraph_community_infomap(
    const igraph_t *graph,
    const igraph_vector_t *edge_weights,
    const igraph_vector_t *vertex_weights,
    igraph_int_t nb_trials,
    igraph_bool_t is_regularized,
    igraph_real_t regularization_strength,
    igraph_vector_int_t *membership,
    igraph_real_t *codelength) {
    GO_IGRAPH_CALL(igraph_community_infomap(
        graph, edge_weights, vertex_weights, nb_trials, is_regularized, regularization_strength, membership, codelength));
}

igraph_error_t go_igraph_community_fluid_communities(
    const igraph_t *graph,
    igraph_int_t no_of_communities,
    igraph_vector_int_t *membership) {
    GO_IGRAPH_CALL(igraph_community_fluid_communities(
        graph, no_of_communities, membership));
}
