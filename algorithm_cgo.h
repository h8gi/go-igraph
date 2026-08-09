#ifndef GO_IGRAPH_ALGORITHM_CGO_H
#define GO_IGRAPH_ALGORITHM_CGO_H

#include <igraph.h>

/* Fallible temporary-resource operations used by algorithm bindings. */
igraph_error_t go_igraph_vector_int_init(igraph_vector_int_t *, igraph_int_t);
igraph_error_t go_igraph_vector_init(igraph_vector_t *, igraph_int_t);
igraph_error_t go_igraph_matrix_init(igraph_matrix_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_vector_int_list_init(igraph_vector_int_list_t *, igraph_int_t);
igraph_error_t go_igraph_graph_list_init(igraph_graph_list_t *, igraph_int_t);
igraph_error_t go_igraph_graph_list_push_back_copy(igraph_graph_list_t *, const igraph_t *);
igraph_error_t go_igraph_graph_list_remove(igraph_graph_list_t *, igraph_int_t, igraph_t *);
igraph_error_t go_igraph_copy(igraph_t *, const igraph_t *);
igraph_error_t go_igraph_delete_edges(igraph_t *, igraph_es_t);
igraph_error_t go_igraph_delete_vertices_map(
    igraph_t *, igraph_vs_t, igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_get_edgelist(
    const igraph_t *, igraph_vector_int_t *, igraph_bool_t);
igraph_error_t go_igraph_simplify(igraph_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_to_directed(igraph_t *, igraph_to_directed_t);
igraph_error_t go_igraph_to_undirected(igraph_t *, igraph_to_undirected_t);
igraph_error_t go_igraph_vs_vector_copy(igraph_vs_t *, const igraph_vector_int_t *);
igraph_error_t go_igraph_es_vector_copy(igraph_es_t *, const igraph_vector_int_t *);
igraph_error_t go_igraph_vit_create(const igraph_t *, igraph_vs_t, igraph_vit_t *);
igraph_error_t go_igraph_eit_create(const igraph_t *, igraph_es_t, igraph_eit_t *);
igraph_error_t go_igraph_get_eid(const igraph_t *, igraph_int_t *, igraph_int_t,
                                 igraph_int_t, igraph_bool_t, igraph_bool_t);

/* Milestone 3 algorithms. */
igraph_error_t go_igraph_degree(const igraph_t *, igraph_vector_int_t *,
                                igraph_vs_t, igraph_neimode_t, igraph_loops_t);
igraph_error_t go_igraph_neighborhood_size(
    const igraph_t *, igraph_vector_int_t *, igraph_vs_t, igraph_int_t,
    igraph_neimode_t, igraph_int_t);
igraph_error_t go_igraph_neighborhood(
    const igraph_t *, igraph_vector_int_list_t *, igraph_vs_t, igraph_int_t,
    igraph_neimode_t, igraph_int_t);
igraph_error_t go_igraph_connected_components(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_int_t *, igraph_connectedness_t);
igraph_error_t go_igraph_is_connected(const igraph_t *, igraph_bool_t *,
                                      igraph_connectedness_t);
igraph_error_t go_igraph_articulation_points(const igraph_t *,
                                             igraph_vector_int_t *);
igraph_error_t go_igraph_bridges(const igraph_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_biconnected_components(
    const igraph_t *, igraph_int_t *, igraph_vector_int_list_t *,
    igraph_vector_int_list_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_bfs(
    const igraph_t *, igraph_int_t, const igraph_vector_int_t *,
    igraph_neimode_t, igraph_bool_t, const igraph_vector_int_t *,
    igraph_vector_int_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_vector_int_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_bfshandler_t *, void *);
igraph_error_t go_igraph_dfs(
    const igraph_t *, igraph_int_t, igraph_neimode_t, igraph_bool_t,
    igraph_vector_int_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_vector_int_t *, igraph_dfshandler_t *, igraph_dfshandler_t *, void *);
igraph_error_t go_igraph_distances(
    const igraph_t *, const igraph_vector_t *, igraph_matrix_t *, igraph_vs_t,
    igraph_vs_t, igraph_neimode_t);
igraph_error_t go_igraph_get_shortest_path(
    const igraph_t *, const igraph_vector_t *, igraph_vector_int_t *,
    igraph_vector_int_t *, igraph_int_t, igraph_int_t, igraph_neimode_t);
igraph_error_t go_igraph_density(const igraph_t *, const igraph_vector_t *,
                                 igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_diameter(
    const igraph_t *, const igraph_vector_t *, igraph_real_t *, igraph_int_t *,
    igraph_int_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_average_path_length(
    const igraph_t *, const igraph_vector_t *, igraph_real_t *, igraph_real_t *,
    igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_transitivity_undirected(
    const igraph_t *, igraph_real_t *, igraph_transitivity_mode_t);
igraph_error_t go_igraph_transitivity_local_undirected(
    const igraph_t *, igraph_vector_t *, igraph_vs_t,
    igraph_transitivity_mode_t);
igraph_error_t go_igraph_transitivity_avglocal_undirected(
    const igraph_t *, igraph_real_t *, igraph_transitivity_mode_t);

/* Milestone 4 centrality algorithms. */
igraph_error_t go_igraph_closeness(
    const igraph_t *, igraph_vector_t *, igraph_vector_int_t *,
    igraph_bool_t *, igraph_vs_t, igraph_neimode_t,
    const igraph_vector_t *, igraph_bool_t, igraph_bool_t, igraph_real_t);
igraph_error_t go_igraph_harmonic_centrality(
    const igraph_t *, igraph_vector_t *, igraph_vs_t, igraph_neimode_t,
    const igraph_vector_t *, igraph_bool_t, igraph_bool_t, igraph_real_t);
igraph_error_t go_igraph_betweenness(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *,
    igraph_vs_t, igraph_bool_t, igraph_bool_t, igraph_bool_t, igraph_real_t);
igraph_error_t go_igraph_edge_betweenness(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *,
    igraph_es_t, igraph_bool_t, igraph_bool_t, igraph_bool_t, igraph_real_t);
igraph_error_t go_igraph_eigenvector_centrality(
    const igraph_t *, igraph_vector_t *, igraph_real_t *, igraph_neimode_t,
    const igraph_vector_t *, int, igraph_real_t);
igraph_error_t go_igraph_hub_and_authority_scores(
    const igraph_t *, igraph_vector_t *, igraph_vector_t *, igraph_real_t *,
    const igraph_vector_t *, int, igraph_real_t);
igraph_error_t go_igraph_pagerank(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *,
    igraph_real_t *, const igraph_vector_t *, igraph_vs_t, int,
    igraph_real_t, igraph_bool_t, igraph_vs_t, igraph_pagerank_algo_t,
    int, igraph_real_t);
igraph_error_t go_igraph_centralization_degree(
    const igraph_t *, igraph_vector_t *, igraph_neimode_t, igraph_loops_t,
    igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_centralization_degree_tmax(
    const igraph_t *, igraph_real_t *, igraph_neimode_t, igraph_loops_t);
igraph_error_t go_igraph_centralization_betweenness(
    const igraph_t *, igraph_vector_t *, igraph_bool_t, igraph_real_t *,
    igraph_bool_t);
igraph_error_t go_igraph_centralization_betweenness_tmax(
    const igraph_t *, igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_centralization_closeness(
    const igraph_t *, igraph_vector_t *, igraph_neimode_t, igraph_real_t *,
    igraph_bool_t);
igraph_error_t go_igraph_centralization_closeness_tmax(
    const igraph_t *, igraph_real_t *, igraph_neimode_t);
igraph_error_t go_igraph_centralization_eigenvector(
    const igraph_t *, igraph_vector_t *, igraph_neimode_t, int,
    igraph_real_t, igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_centralization_eigenvector_tmax(
    const igraph_t *, igraph_real_t *, igraph_neimode_t);

#endif
