#ifndef GO_IGRAPH_ISOMORPHISM_CGO_H
#define GO_IGRAPH_ISOMORPHISM_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_isomorphic(
    const igraph_t *left, const igraph_t *right, igraph_bool_t *result);

igraph_error_t go_igraph_subisomorphic(
    const igraph_t *pattern, const igraph_t *target, igraph_bool_t *result);

igraph_error_t go_igraph_isomorphic_vf2(
    const igraph_t *source, const igraph_t *target,
    const igraph_vector_int_t *source_vertex_colors,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *source_edge_colors,
    const igraph_vector_int_t *target_edge_colors,
    igraph_bool_t *result, igraph_vector_int_t *source_to_target,
    igraph_vector_int_t *target_to_source);

igraph_error_t go_igraph_subisomorphic_vf2(
    const igraph_t *target, const igraph_t *pattern,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *pattern_vertex_colors,
    const igraph_vector_int_t *target_edge_colors,
    const igraph_vector_int_t *pattern_edge_colors,
    igraph_bool_t *result, igraph_vector_int_t *target_to_pattern,
    igraph_vector_int_t *pattern_to_target);

igraph_error_t go_igraph_is_simple(
    const igraph_t *graph, igraph_bool_t *result);

igraph_error_t go_igraph_count_isomorphisms_vf2(
    const igraph_t *source, const igraph_t *target,
    const igraph_vector_int_t *source_vertex_colors,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *source_edge_colors,
    const igraph_vector_int_t *target_edge_colors,
    igraph_int_t *count);

igraph_error_t go_igraph_count_subisomorphisms_vf2(
    const igraph_t *target, const igraph_t *pattern,
    const igraph_vector_int_t *target_vertex_colors,
    const igraph_vector_int_t *pattern_vertex_colors,
    const igraph_vector_int_t *target_edge_colors,
    const igraph_vector_int_t *pattern_edge_colors,
    igraph_int_t *count);

igraph_error_t go_igraph_enumerate_isomorphisms_vf2(
    const igraph_t *first, const igraph_t *second,
    const igraph_vector_int_t *first_vertex_colors,
    const igraph_vector_int_t *second_vertex_colors,
    const igraph_vector_int_t *first_edge_colors,
    const igraph_vector_int_t *second_edge_colors,
    igraph_int_t max_mappings, igraph_bool_t subgraph,
    igraph_vector_int_list_t *mappings, igraph_bool_t *truncated);

igraph_error_t go_igraph_vector_int_list_push_back_copy(
    igraph_vector_int_list_t *list, const igraph_vector_int_t *vector);

igraph_error_t go_igraph_subisomorphic_lad(
    const igraph_t *pattern, const igraph_t *target,
    const igraph_vector_int_list_t *domains,
    igraph_bool_t *result, igraph_vector_int_t *mapping,
    igraph_bool_t induced);

igraph_error_t go_igraph_canonical_permutation(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    igraph_vector_int_t *permutation);

igraph_error_t go_igraph_permute_vertices(
    const igraph_t *graph, igraph_t *result,
    const igraph_vector_int_t *permutation);

igraph_error_t go_igraph_automorphism_group(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    igraph_vector_int_list_t *generators);

igraph_error_t go_igraph_count_automorphisms_exact(
    const igraph_t *graph, const igraph_vector_int_t *colors,
    char **decimal);

void go_igraph_free(void *pointer);

#endif
