#include "mixing_cgo.h"
#include "igraph_error_cgo.h"

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
