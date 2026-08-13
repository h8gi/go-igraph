#include "algorithm_cgo.h"
#include "igraph_error_cgo.h"

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

igraph_error_t go_igraph_graph_list_init(igraph_graph_list_t *value,
                                         igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_graph_list_init(value, size));
}

igraph_error_t go_igraph_graph_list_push_back_copy(igraph_graph_list_t *list,
                                                   const igraph_t *graph) {
    GO_IGRAPH_CALL(igraph_graph_list_push_back_copy(list, graph));
}

igraph_error_t go_igraph_graph_list_remove(igraph_graph_list_t *list,
                                           igraph_int_t index,
                                           igraph_t *graph) {
    GO_IGRAPH_CALL(igraph_graph_list_remove(list, index, graph));
}

igraph_error_t go_igraph_copy(igraph_t *to, const igraph_t *from) {
    GO_IGRAPH_CALL(igraph_copy(to, from));
}

igraph_error_t go_igraph_delete_edges(igraph_t *graph, igraph_es_t edges) {
    GO_IGRAPH_CALL(igraph_delete_edges(graph, edges));
}

igraph_error_t go_igraph_delete_vertices_map(
        igraph_t *graph, igraph_vs_t vertices, igraph_vector_int_t *map,
        igraph_vector_int_t *invmap) {
    GO_IGRAPH_CALL(igraph_delete_vertices_map(
        graph, vertices, map, invmap));
}

igraph_error_t go_igraph_get_edgelist(
        const igraph_t *graph, igraph_vector_int_t *edges,
        igraph_bool_t by_column) {
    GO_IGRAPH_CALL(igraph_get_edgelist(graph, edges, by_column));
}

igraph_error_t go_igraph_simplify(igraph_t *graph,
                                  igraph_bool_t remove_multiple,
                                  igraph_bool_t remove_loops) {
    GO_IGRAPH_CALL(igraph_simplify(
        graph, remove_multiple, remove_loops, NULL));
}

igraph_error_t go_igraph_to_directed(igraph_t *graph,
                                     igraph_to_directed_t mode) {
    GO_IGRAPH_CALL(igraph_to_directed(graph, mode));
}

igraph_error_t go_igraph_to_undirected(igraph_t *graph,
                                       igraph_to_undirected_t mode) {
    GO_IGRAPH_CALL(igraph_to_undirected(graph, mode, NULL));
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

igraph_error_t go_igraph_articulation_points(
    const igraph_t *graph, igraph_vector_int_t *result) {
    GO_IGRAPH_CALL(igraph_articulation_points(graph, result));
}

igraph_error_t go_igraph_bridges(const igraph_t *graph,
                                 igraph_vector_int_t *result) {
    GO_IGRAPH_CALL(igraph_bridges(graph, result));
}

igraph_error_t go_igraph_biconnected_components(
    const igraph_t *graph, igraph_int_t *count,
    igraph_vector_int_list_t *component_edges,
    igraph_vector_int_list_t *component_vertices,
    igraph_vector_int_t *articulation_points) {
    GO_IGRAPH_CALL(igraph_biconnected_components(
        graph, count, NULL, component_edges, component_vertices,
        articulation_points));
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

igraph_error_t go_igraph_get_shortest_paths(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_int_list_t *vertices, igraph_vector_int_list_t *edges,
    igraph_int_t source, igraph_vs_t targets, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_get_shortest_paths(
        graph, weights, vertices, edges, source, targets, mode, NULL, NULL));
}

igraph_error_t go_igraph_get_k_shortest_paths(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_int_list_t *vertices, igraph_vector_int_list_t *edges,
    igraph_int_t k, igraph_int_t source, igraph_int_t target,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_get_k_shortest_paths(
        graph, weights, vertices, edges, k, source, target, mode));
}

igraph_error_t go_igraph_get_all_simple_paths(
    const igraph_t *graph, igraph_vector_int_list_t *paths,
    igraph_int_t source, igraph_vs_t targets, igraph_neimode_t mode,
    igraph_int_t min_length, igraph_int_t max_length,
    igraph_int_t max_results) {
    GO_IGRAPH_CALL(igraph_get_all_simple_paths(
        graph, paths, source, targets, mode, min_length, max_length,
        max_results));
}

igraph_error_t go_igraph_distances_cutoff(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_matrix_t *result, igraph_vs_t sources, igraph_vs_t targets,
    igraph_neimode_t mode, igraph_real_t cutoff) {
    GO_IGRAPH_CALL(igraph_distances_cutoff(
        graph, weights, result, sources, targets, mode, cutoff));
}

igraph_error_t go_igraph_eccentricity(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_t *result, igraph_vs_t vertices, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_eccentricity(graph, weights, result, vertices, mode));
}

igraph_error_t go_igraph_radius(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *result, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_radius(graph, weights, result, mode));
}

igraph_error_t go_igraph_graph_center(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_int_t *result, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_graph_center(graph, weights, result, mode));
}

igraph_error_t go_igraph_pseudo_diameter(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *diameter, igraph_int_t start, igraph_int_t *from,
    igraph_int_t *to, igraph_bool_t directed, igraph_bool_t unconnected) {
    GO_IGRAPH_CALL(igraph_pseudo_diameter(
        graph, weights, diameter, start, from, to, directed, unconnected));
}

igraph_error_t go_igraph_global_efficiency(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *result, igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_global_efficiency(graph, weights, result, directed));
}

igraph_error_t go_igraph_local_efficiency(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_t *result, igraph_vs_t vertices, igraph_bool_t directed,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_efficiency(
        graph, weights, result, vertices, directed, mode));
}

igraph_error_t go_igraph_average_local_efficiency(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *result, igraph_bool_t directed, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_average_local_efficiency(
        graph, weights, result, directed, mode));
}

igraph_error_t go_igraph_path_length_hist(
    const igraph_t *graph, igraph_vector_t *result,
    igraph_real_t *unconnected, igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_path_length_hist(
        graph, result, unconnected, directed));
}

igraph_error_t go_igraph_get_widest_paths(
    const igraph_t *graph, igraph_vector_int_list_t *vertices,
    igraph_vector_int_list_t *edges, igraph_int_t from, igraph_vs_t to,
    const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_get_widest_paths(
        graph, vertices, edges, from, to, weights, mode, NULL, NULL));
}

igraph_error_t go_igraph_widest_path_widths(
    const igraph_t *graph, igraph_matrix_t *result, igraph_vs_t from,
    igraph_vs_t to, const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_widest_path_widths_dijkstra(
        graph, result, from, to, weights, mode));
}

igraph_error_t go_igraph_voronoi(
    const igraph_t *graph, igraph_vector_int_t *membership,
    igraph_vector_t *distances, const igraph_vector_int_t *generators,
    const igraph_vector_t *weights, igraph_neimode_t mode,
    igraph_voronoi_tiebreaker_t tiebreaker) {
    GO_IGRAPH_CALL(igraph_voronoi(
        graph, membership, distances, generators, weights, mode, tiebreaker));
}

igraph_error_t go_igraph_spanner(
    const igraph_t *graph, igraph_vector_int_t *edges,
    igraph_real_t stretch, const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_spanner(graph, edges, stretch, weights));
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

igraph_error_t go_igraph_betweenness(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_t *result, igraph_vs_t vertices, igraph_bool_t directed,
    igraph_bool_t normalized, igraph_bool_t has_cutoff,
    igraph_real_t cutoff) {
    if (has_cutoff) {
        GO_IGRAPH_CALL(igraph_betweenness_cutoff(
            graph, weights, result, vertices, directed, normalized, cutoff));
    }
    GO_IGRAPH_CALL(igraph_betweenness(
        graph, weights, result, vertices, directed, normalized));
}

igraph_error_t go_igraph_edge_betweenness(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_t *result, igraph_es_t edges, igraph_bool_t directed,
    igraph_bool_t normalized, igraph_bool_t has_cutoff,
    igraph_real_t cutoff) {
    if (has_cutoff) {
        GO_IGRAPH_CALL(igraph_edge_betweenness_cutoff(
            graph, weights, result, edges, directed, normalized, cutoff));
    }
    GO_IGRAPH_CALL(igraph_edge_betweenness(
        graph, weights, result, edges, directed, normalized));
}

void go_igraph_arpack_options(
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

igraph_error_t go_igraph_eigenvector_centrality(
    const igraph_t *graph, igraph_vector_t *result, igraph_real_t *value,
    igraph_neimode_t mode, const igraph_vector_t *weights,
    int max_iterations, igraph_real_t tolerance) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_eigenvector_centrality(
        graph, result, value, mode, weights, &options));
}

igraph_error_t go_igraph_hub_and_authority_scores(
    const igraph_t *graph, igraph_vector_t *hub_result,
    igraph_vector_t *authority_result, igraph_real_t *value,
    const igraph_vector_t *weights, int max_iterations,
    igraph_real_t tolerance) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_hub_and_authority_scores(
        graph, hub_result, authority_result, value, weights, &options));
}

igraph_error_t go_igraph_pagerank(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_t *result, igraph_real_t *value,
    const igraph_vector_t *reset, igraph_vs_t reset_vertices, int reset_kind,
    igraph_real_t damping, igraph_bool_t directed, igraph_vs_t vertices,
    igraph_pagerank_algo_t algorithm, int max_iterations,
    igraph_real_t tolerance) {
    igraph_arpack_options_t options;
    go_igraph_arpack_options(&options, max_iterations, tolerance);
    if (reset_kind == 1) {
        GO_IGRAPH_CALL(igraph_personalized_pagerank(
            graph, weights, result, value, reset, damping, directed, vertices,
            algorithm, &options));
    }
    if (reset_kind == 2) {
        GO_IGRAPH_CALL(igraph_personalized_pagerank_vs(
            graph, weights, result, value, reset_vertices, damping, directed,
            vertices, algorithm, &options));
    }
    GO_IGRAPH_CALL(igraph_pagerank(
        graph, weights, result, value, damping, directed, vertices, algorithm,
        &options));
}

igraph_error_t go_igraph_centralization_degree(
    const igraph_t *graph, igraph_vector_t *result, igraph_neimode_t mode,
    igraph_loops_t loops, igraph_real_t *centralization,
    igraph_bool_t normalized) {
    igraph_real_t ignored_theoretical_max;
    GO_IGRAPH_CALL(igraph_centralization_degree(
        graph, result, mode, loops, centralization,
        &ignored_theoretical_max, normalized));
}

igraph_error_t go_igraph_centralization_degree_tmax(
    const igraph_t *graph, igraph_real_t *result, igraph_neimode_t mode,
    igraph_loops_t loops) {
    GO_IGRAPH_CALL(igraph_centralization_degree_tmax(
        graph, igraph_vcount(graph), mode, loops, result));
}

igraph_error_t go_igraph_centralization_betweenness(
    const igraph_t *graph, igraph_vector_t *result, igraph_bool_t directed,
    igraph_real_t *centralization, igraph_bool_t normalized) {
    igraph_real_t ignored_theoretical_max;
    GO_IGRAPH_CALL(igraph_centralization_betweenness(
        graph, result, directed, centralization, &ignored_theoretical_max,
        normalized));
}

igraph_error_t go_igraph_centralization_betweenness_tmax(
    const igraph_t *graph, igraph_real_t *result, igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_centralization_betweenness_tmax(
        graph, igraph_vcount(graph), directed, result));
}

igraph_error_t go_igraph_centralization_closeness(
    const igraph_t *graph, igraph_vector_t *result, igraph_neimode_t mode,
    igraph_real_t *centralization, igraph_bool_t normalized) {
    igraph_real_t ignored_theoretical_max;
    GO_IGRAPH_CALL(igraph_centralization_closeness(
        graph, result, mode, centralization, &ignored_theoretical_max,
        normalized));
}

igraph_error_t go_igraph_centralization_closeness_tmax(
    const igraph_t *graph, igraph_real_t *result, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_centralization_closeness_tmax(
        graph, igraph_vcount(graph), mode, result));
}

igraph_error_t go_igraph_centralization_eigenvector(
    const igraph_t *graph, igraph_vector_t *result, igraph_neimode_t mode,
    int max_iterations, igraph_real_t tolerance,
    igraph_real_t *centralization, igraph_bool_t normalized) {
    igraph_arpack_options_t options;
    igraph_real_t ignored_eigenvalue, ignored_theoretical_max;
    go_igraph_arpack_options(&options, max_iterations, tolerance);
    GO_IGRAPH_CALL(igraph_centralization_eigenvector_centrality(
        graph, result, &ignored_eigenvalue, mode, &options, centralization,
        &ignored_theoretical_max, normalized));
}

igraph_error_t go_igraph_centralization_eigenvector_tmax(
    const igraph_t *graph, igraph_real_t *result, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_centralization_eigenvector_centrality_tmax(
        graph, igraph_vcount(graph), mode, result));
}
