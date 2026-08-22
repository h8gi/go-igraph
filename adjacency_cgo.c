#include "adjacency_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_adjacency(
    igraph_t *graph, const igraph_matrix_t *matrix,
    igraph_adjacency_t mode, igraph_loops_t loops) {
  GO_IGRAPH_CALL(igraph_adjacency(graph, matrix, mode, loops));
}

igraph_error_t go_igraph_weighted_adjacency(
    igraph_t *graph, const igraph_matrix_t *matrix, igraph_adjacency_t mode,
    igraph_vector_t *weights, igraph_loops_t loops) {
  GO_IGRAPH_CALL(
      igraph_weighted_adjacency(graph, matrix, mode, weights, loops));
}
