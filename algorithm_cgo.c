#include "algorithm_cgo.h"

/*
 * The pinned thread-safe igraph build stores handlers in thread-local state.
 * Installing and restoring them around one upstream operation in one cgo call
 * keeps handler state on the same OS thread and turns igraph failures into
 * return codes instead of process aborts.
 */
#define GO_IGRAPH_CALL(expression)                                          \
    do {                                                                    \
        igraph_error_handler_t *old_error =                                 \
            igraph_set_error_handler(&igraph_error_handler_ignore);         \
        igraph_warning_handler_t *old_warning =                             \
            igraph_set_warning_handler(&igraph_warning_handler_ignore);     \
        igraph_error_t code = (expression);                                 \
        igraph_set_warning_handler(old_warning);                            \
        igraph_set_error_handler(old_error);                                \
        return code;                                                        \
    } while (0)

igraph_error_t go_igraph_vector_int_init(igraph_vector_int_t *value,
                                         igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_vector_int_init(value, size));
}

igraph_error_t go_igraph_vector_init(igraph_vector_t *value,
                                     igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_vector_init(value, size));
}

igraph_error_t go_igraph_matrix_init(igraph_matrix_t *value,
                                     igraph_int_t rows,
                                     igraph_int_t columns) {
    GO_IGRAPH_CALL(igraph_matrix_init(value, rows, columns));
}

igraph_error_t go_igraph_vector_int_list_init(igraph_vector_int_list_t *value,
                                              igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_vector_int_list_init(value, size));
}

igraph_error_t go_igraph_vs_vector_copy(igraph_vs_t *value,
                                        const igraph_vector_int_t *ids) {
    GO_IGRAPH_CALL(igraph_vs_vector_copy(value, ids));
}

igraph_error_t go_igraph_es_vector_copy(igraph_es_t *value,
                                        const igraph_vector_int_t *ids) {
    GO_IGRAPH_CALL(igraph_es_vector_copy(value, ids));
}

igraph_error_t go_igraph_vit_create(const igraph_t *graph,
                                    igraph_vs_t selector,
                                    igraph_vit_t *iterator) {
    GO_IGRAPH_CALL(igraph_vit_create(graph, selector, iterator));
}

igraph_error_t go_igraph_eit_create(const igraph_t *graph,
                                    igraph_es_t selector,
                                    igraph_eit_t *iterator) {
    GO_IGRAPH_CALL(igraph_eit_create(graph, selector, iterator));
}

igraph_error_t go_igraph_get_eid(const igraph_t *graph,
                                 igraph_int_t *edge_id,
                                 igraph_int_t from,
                                 igraph_int_t to,
                                 igraph_bool_t directed,
                                 igraph_bool_t error_on_missing) {
    GO_IGRAPH_CALL(igraph_get_eid(
        graph, edge_id, from, to, directed, error_on_missing));
}

igraph_error_t go_igraph_degree(const igraph_t *graph,
                                igraph_vector_int_t *result,
                                igraph_vs_t vertices,
                                igraph_neimode_t mode,
                                igraph_loops_t loops) {
    GO_IGRAPH_CALL(igraph_degree(graph, result, vertices, mode, loops));
}

igraph_error_t go_igraph_neighborhood_size(
    const igraph_t *graph, igraph_vector_int_t *result, igraph_vs_t vertices,
    igraph_int_t order, igraph_neimode_t mode, igraph_int_t min_distance) {
    GO_IGRAPH_CALL(igraph_neighborhood_size(
        graph, result, vertices, order, mode, min_distance));
}

igraph_error_t go_igraph_neighborhood(
    const igraph_t *graph, igraph_vector_int_list_t *result,
    igraph_vs_t vertices, igraph_int_t order, igraph_neimode_t mode,
    igraph_int_t min_distance) {
    GO_IGRAPH_CALL(igraph_neighborhood(
        graph, result, vertices, order, mode, min_distance));
}

igraph_error_t go_igraph_connected_components(
    const igraph_t *graph, igraph_vector_int_t *membership,
    igraph_vector_int_t *sizes, igraph_int_t *count,
    igraph_connectedness_t mode) {
    GO_IGRAPH_CALL(igraph_connected_components(
        graph, membership, sizes, count, mode));
}

igraph_error_t go_igraph_is_connected(const igraph_t *graph,
                                      igraph_bool_t *result,
                                      igraph_connectedness_t mode) {
    GO_IGRAPH_CALL(igraph_is_connected(graph, result, mode));
}

igraph_error_t go_igraph_bfs(
    const igraph_t *graph, igraph_int_t root,
    const igraph_vector_int_t *roots, igraph_neimode_t mode,
    igraph_bool_t unreachable, const igraph_vector_int_t *restricted,
    igraph_vector_int_t *order, igraph_vector_int_t *rank,
    igraph_vector_int_t *parents, igraph_vector_int_t *predecessors,
    igraph_vector_int_t *successors, igraph_vector_int_t *distances,
    igraph_bfshandler_t *callback, void *extra) {
    GO_IGRAPH_CALL(igraph_bfs(
        graph, root, roots, mode, unreachable, restricted, order, rank,
        parents, predecessors, successors, distances, callback, extra));
}

igraph_error_t go_igraph_dfs(
    const igraph_t *graph, igraph_int_t root, igraph_neimode_t mode,
    igraph_bool_t unreachable, igraph_vector_int_t *order,
    igraph_vector_int_t *finish_order, igraph_vector_int_t *parents,
    igraph_vector_int_t *distances, igraph_dfshandler_t *in_callback,
    igraph_dfshandler_t *out_callback, void *extra) {
    GO_IGRAPH_CALL(igraph_dfs(
        graph, root, mode, unreachable, order, finish_order, parents,
        distances, in_callback, out_callback, extra));
}

igraph_error_t go_igraph_distances(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_matrix_t *result, igraph_vs_t sources, igraph_vs_t targets,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_distances(
        graph, weights, result, sources, targets, mode));
}

igraph_error_t go_igraph_get_shortest_path(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_int_t *vertices, igraph_vector_int_t *edges,
    igraph_int_t source, igraph_int_t target, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_get_shortest_path(
        graph, weights, vertices, edges, source, target, mode));
}

igraph_error_t go_igraph_density(const igraph_t *graph,
                                 const igraph_vector_t *weights,
                                 igraph_real_t *result,
                                 igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_density(graph, weights, result, loops));
}

igraph_error_t go_igraph_diameter(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *length, igraph_int_t *from, igraph_int_t *to,
    igraph_vector_int_t *vertices, igraph_vector_int_t *edges,
    igraph_bool_t directed, igraph_bool_t unconnected) {
    GO_IGRAPH_CALL(igraph_diameter(
        graph, weights, length, from, to, vertices, edges, directed,
        unconnected));
}

igraph_error_t go_igraph_average_path_length(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *result, igraph_real_t *unconnected_pairs,
    igraph_bool_t directed, igraph_bool_t unconnected) {
    GO_IGRAPH_CALL(igraph_average_path_length(
        graph, weights, result, unconnected_pairs, directed, unconnected));
}

igraph_error_t go_igraph_transitivity_undirected(
    const igraph_t *graph, igraph_real_t *result,
    igraph_transitivity_mode_t mode) {
    GO_IGRAPH_CALL(igraph_transitivity_undirected(graph, result, mode));
}

igraph_error_t go_igraph_transitivity_local_undirected(
    const igraph_t *graph, igraph_vector_t *result, igraph_vs_t vertices,
    igraph_transitivity_mode_t mode) {
    GO_IGRAPH_CALL(igraph_transitivity_local_undirected(
        graph, result, vertices, mode));
}

igraph_error_t go_igraph_transitivity_avglocal_undirected(
    const igraph_t *graph, igraph_real_t *result,
    igraph_transitivity_mode_t mode) {
    GO_IGRAPH_CALL(igraph_transitivity_avglocal_undirected(
        graph, result, mode));
}

igraph_error_t go_igraph_closeness(
    const igraph_t *graph, igraph_vector_t *result,
    igraph_vector_int_t *reachable_count, igraph_bool_t *all_reachable,
    igraph_vs_t vertices, igraph_neimode_t mode,
    const igraph_vector_t *weights, igraph_bool_t normalized,
    igraph_bool_t has_cutoff, igraph_real_t cutoff) {
    if (has_cutoff) {
        GO_IGRAPH_CALL(igraph_closeness_cutoff(
            graph, result, reachable_count, all_reachable, vertices, mode,
            weights, normalized, cutoff));
    }
    GO_IGRAPH_CALL(igraph_closeness(
        graph, result, reachable_count, all_reachable, vertices, mode,
        weights, normalized));
}

igraph_error_t go_igraph_harmonic_centrality(
    const igraph_t *graph, igraph_vector_t *result, igraph_vs_t vertices,
    igraph_neimode_t mode, const igraph_vector_t *weights,
    igraph_bool_t normalized, igraph_bool_t has_cutoff,
    igraph_real_t cutoff) {
    if (has_cutoff) {
        GO_IGRAPH_CALL(igraph_harmonic_centrality_cutoff(
            graph, result, vertices, mode, weights, normalized, cutoff));
    }
    GO_IGRAPH_CALL(igraph_harmonic_centrality(
        graph, result, vertices, mode, weights, normalized));
}
