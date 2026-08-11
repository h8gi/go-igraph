#include "isomorphism_cgo.h"
#include "igraph_error_cgo.h"
#include <string.h>

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

igraph_error_t go_igraph_count_isomorphisms_vf2(
    const igraph_t *source, const igraph_t *target,
    const igraph_vector_int_t *source_vertex_colors,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *source_edge_colors,
    const igraph_vector_int_t *target_edge_colors,
    igraph_int_t *count) {
    GO_IGRAPH_CALL(igraph_count_isomorphisms_vf2(
        source, target, source_vertex_colors, target_vertex_colors,
        source_edge_colors, target_edge_colors, count, NULL, NULL, NULL));
}

igraph_error_t go_igraph_count_subisomorphisms_vf2(
    const igraph_t *target, const igraph_t *pattern,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *pattern_vertex_colors,
    const igraph_vector_int_t *target_edge_colors,
    const igraph_vector_int_t *pattern_edge_colors,
    igraph_int_t *count) {
    GO_IGRAPH_CALL(igraph_count_subisomorphisms_vf2(
        target, pattern, target_vertex_colors, pattern_vertex_colors,
        target_edge_colors, pattern_edge_colors, count, NULL, NULL, NULL));
}

typedef struct {
    igraph_vector_int_list_t *mappings;
    igraph_int_t max_mappings;
    igraph_bool_t subgraph;
    igraph_bool_t truncated;
} go_igraph_vf2_collector_t;

static igraph_error_t go_igraph_vf2_collect(
    const igraph_vector_int_t *map12,
    const igraph_vector_int_t *map21,
    void *arg) {
    go_igraph_vf2_collector_t *collector = arg;
    if (igraph_vector_int_list_size(collector->mappings) >= collector->max_mappings) {
        collector->truncated = true;
        return IGRAPH_STOP;
    }
    return igraph_vector_int_list_push_back_copy(
        collector->mappings, collector->subgraph ? map21 : map12);
}

static igraph_error_t go_igraph_run_enumerate_isomorphisms_vf2(
    const igraph_t *first, const igraph_t *second,
    const igraph_vector_int_t *first_vertex_colors,
    const igraph_vector_int_t *second_vertex_colors,
    const igraph_vector_int_t *first_edge_colors,
    const igraph_vector_int_t *second_edge_colors,
    igraph_bool_t subgraph, go_igraph_vf2_collector_t *collector) {
    if (subgraph) {
        GO_IGRAPH_CALL(igraph_get_subisomorphisms_vf2_callback(
            first, second, first_vertex_colors, second_vertex_colors,
            first_edge_colors, second_edge_colors, NULL, NULL,
            &go_igraph_vf2_collect, NULL, NULL, collector));
    } else {
        GO_IGRAPH_CALL(igraph_get_isomorphisms_vf2_callback(
            first, second, first_vertex_colors, second_vertex_colors,
            first_edge_colors, second_edge_colors, NULL, NULL,
            &go_igraph_vf2_collect, NULL, NULL, collector));
    }
}

igraph_error_t go_igraph_enumerate_isomorphisms_vf2(
    const igraph_t *first, const igraph_t *second,
    const igraph_vector_int_t *first_vertex_colors,
    const igraph_vector_int_t *second_vertex_colors,
    const igraph_vector_int_t *first_edge_colors,
    const igraph_vector_int_t *second_edge_colors,
    igraph_int_t max_mappings, igraph_bool_t subgraph,
    igraph_vector_int_list_t *mappings, igraph_bool_t *truncated) {
    go_igraph_vf2_collector_t collector = {
        .mappings = mappings,
        .max_mappings = max_mappings,
        .subgraph = subgraph,
        .truncated = false
    };
    igraph_error_t code = go_igraph_run_enumerate_isomorphisms_vf2(
        first, second, first_vertex_colors, second_vertex_colors,
        first_edge_colors, second_edge_colors, subgraph, &collector);
    *truncated = collector.truncated;
    return code;
}

igraph_error_t go_igraph_vector_int_list_push_back_copy(
    igraph_vector_int_list_t *list, const igraph_vector_int_t *vector) {
    GO_IGRAPH_CALL(igraph_vector_int_list_push_back_copy(list, vector));
}

igraph_error_t go_igraph_subisomorphic_lad(
    const igraph_t *pattern, const igraph_t *target,
    const igraph_vector_int_list_t *domains,
    igraph_bool_t *result, igraph_vector_int_t *mapping,
    igraph_bool_t induced) {
    GO_IGRAPH_CALL(igraph_subisomorphic_lad(
        pattern, target, domains, result, mapping, NULL, induced));
}

igraph_error_t go_igraph_canonical_permutation(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    igraph_vector_int_t *permutation) {
    GO_IGRAPH_CALL(igraph_canonical_permutation(graph, colors, permutation));
}

igraph_error_t go_igraph_permute_vertices(
    const igraph_t *graph, igraph_t *result,
    const igraph_vector_int_t *permutation) {
    GO_IGRAPH_CALL(igraph_permute_vertices(graph, result, permutation));
}

igraph_error_t go_igraph_automorphism_group(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    igraph_vector_int_list_t *generators) {
    GO_IGRAPH_CALL(igraph_automorphism_group(graph, colors, generators));
}

static igraph_error_t go_igraph_run_count_automorphisms_exact(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    igraph_bliss_info_t *info) {
    GO_IGRAPH_CALL(igraph_count_automorphisms_bliss(
        graph, colors, IGRAPH_BLISS_FL, info));
}

igraph_error_t go_igraph_count_automorphisms_exact(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    char **decimal) {
    igraph_bliss_info_t info;
    memset(&info, 0, sizeof(info));
    igraph_error_t code = go_igraph_run_count_automorphisms_exact(
        graph, colors, &info);
    if (code != IGRAPH_SUCCESS) {
        igraph_free(info.group_size);
        return code;
    }
    *decimal = info.group_size;
    return IGRAPH_SUCCESS;
}

void go_igraph_free(void *pointer) {
    igraph_free(pointer);
}
