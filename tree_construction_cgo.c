#include "tree_construction_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_from_prufer(igraph_t *graph, const igraph_vector_int_t *prufer) {
  GO_IGRAPH_CALL(igraph_from_prufer(graph, prufer));
}

igraph_error_t go_igraph_to_prufer(const igraph_t *graph, igraph_vector_int_t *prufer) {
  GO_IGRAPH_CALL(igraph_to_prufer(graph, prufer));
}

igraph_error_t go_igraph_tree_from_parent_vector(igraph_t *graph, const igraph_vector_int_t *parents, igraph_tree_mode_t mode) {
  GO_IGRAPH_CALL(igraph_tree_from_parent_vector(graph, parents, mode));
}

igraph_error_t go_igraph_regular_tree(igraph_t *graph, igraph_int_t height, igraph_int_t degree, igraph_tree_mode_t mode) {
  GO_IGRAPH_CALL(igraph_regular_tree(graph, height, degree, mode));
}

igraph_error_t go_igraph_symmetric_tree(igraph_t *graph, const igraph_vector_int_t *branches, igraph_tree_mode_t mode) {
  GO_IGRAPH_CALL(igraph_symmetric_tree(graph, branches, mode));
}
