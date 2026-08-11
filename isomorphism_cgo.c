#include "isomorphism_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_isomorphic(
    const igraph_t *left, const igraph_t *right, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_isomorphic(left, right, result));
}

igraph_error_t go_igraph_subisomorphic(
    const igraph_t *pattern, const igraph_t *target, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_subisomorphic(pattern, target, result));
}

igraph_error_t go_igraph_isomorphic_vf2(
    const igraph_t *source, const igraph_t *target,
    const igraph_vector_int_t *source_vertex_colors,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *source_edge_colors,
    const igraph_vector_int_t *target_edge_colors,
    igraph_bool_t *result, igraph_vector_int_t *source_to_target,
    igraph_vector_int_t *target_to_source) {
    GO_IGRAPH_CALL(igraph_isomorphic_vf2(
        source, target, source_vertex_colors, target_vertex_colors,
        source_edge_colors, target_edge_colors, result, source_to_target,
        target_to_source, NULL, NULL, NULL));
}

igraph_error_t go_igraph_subisomorphic_vf2(
    const igraph_t *target, const igraph_t *pattern,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *pattern_vertex_colors,
    const igraph_vector_int_t *target_edge_colors,
    const igraph_vector_int_t *pattern_edge_colors,
    igraph_bool_t *result, igraph_vector_int_t *target_to_pattern,
    igraph_vector_int_t *pattern_to_target) {
    GO_IGRAPH_CALL(igraph_subisomorphic_vf2(
        target, pattern, target_vertex_colors, pattern_vertex_colors,
        target_edge_colors, pattern_edge_colors, result, target_to_pattern,
        pattern_to_target, NULL, NULL, NULL));
}

igraph_error_t go_igraph_is_simple(
    const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_simple(graph, result, IGRAPH_DIRECTED));
}
