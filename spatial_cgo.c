#include "spatial_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_convex_hull_2d(
        const igraph_matrix_t *points, igraph_vector_int_t *point_indices,
        igraph_matrix_t *coordinates) {
    GO_IGRAPH_CALL(igraph_convex_hull_2d(
        points, point_indices, coordinates));
}

igraph_error_t go_igraph_spatial_edge_lengths(
        const igraph_t *graph, igraph_vector_t *lengths,
        const igraph_matrix_t *points, igraph_metric_t metric) {
    GO_IGRAPH_CALL(igraph_spatial_edge_lengths(
        graph, lengths, points, metric));
}

igraph_error_t go_igraph_nearest_neighbor_graph(
        igraph_t *graph, const igraph_matrix_t *points,
        igraph_metric_t metric, igraph_int_t max_neighbors,
        igraph_real_t cutoff, igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_nearest_neighbor_graph(
        graph, points, metric, max_neighbors, cutoff, directed));
}
