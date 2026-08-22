#include "mixing_cgo.h"
#include "igraph_error_cgo.h"

int go_igraph_is_directed(const igraph_t *graph) {
    return igraph_is_directed(graph) ? 1 : 0;
}

igraph_error_t go_igraph_assortativity_nominal(
        const igraph_t *graph, const igraph_vector_int_t *types,
        igraph_real_t *result, igraph_bool_t directed,
        igraph_bool_t normalized) {
    GO_IGRAPH_CALL(igraph_assortativity_nominal(
        graph, NULL, types, result, directed, normalized));
}

igraph_error_t go_igraph_assortativity(
        const igraph_t *graph, const igraph_vector_t *weights,
        const igraph_vector_t *values, const igraph_vector_t *target_values,
        igraph_real_t *result, igraph_bool_t directed,
        igraph_bool_t normalized) {
    GO_IGRAPH_CALL(igraph_assortativity(
        graph, weights, values, target_values, result, directed, normalized));
}

igraph_error_t go_igraph_assortativity_degree(
        const igraph_t *graph, igraph_real_t *result,
        igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_assortativity_degree(graph, result, directed));
}

igraph_error_t go_igraph_joint_type_distribution(
        const igraph_t *graph, const igraph_vector_t *weights,
        igraph_matrix_t *result, const igraph_vector_int_t *from_types,
        const igraph_vector_int_t *to_types, igraph_bool_t directed,
        igraph_bool_t normalized) {
    GO_IGRAPH_CALL(igraph_joint_type_distribution(
        graph, weights, result, from_types, to_types, directed, normalized));
}

igraph_error_t go_igraph_joint_degree_distribution(
        const igraph_t *graph, const igraph_vector_t *weights,
        igraph_matrix_t *result, igraph_neimode_t from_mode,
        igraph_neimode_t to_mode, igraph_bool_t directed_neighbors,
        igraph_bool_t normalized, igraph_int_t max_from_degree,
        igraph_int_t max_to_degree) {
    GO_IGRAPH_CALL(igraph_joint_degree_distribution(
        graph, weights, result, from_mode, to_mode, directed_neighbors,
        normalized, max_from_degree, max_to_degree));
}

igraph_error_t go_igraph_joint_degree_matrix(
        const igraph_t *graph, const igraph_vector_t *weights,
        igraph_matrix_t *result, igraph_int_t max_out_degree,
        igraph_int_t max_in_degree) {
    GO_IGRAPH_CALL(igraph_joint_degree_matrix(
        graph, weights, result, max_out_degree, max_in_degree));
}
