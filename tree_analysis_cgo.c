#include "tree_analysis_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_tree(const igraph_t *g, igraph_bool_t *result, igraph_int_t *root, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_is_tree(g, result, root, mode));
}
igraph_error_t go_igraph_is_forest(const igraph_t *g, igraph_bool_t *result, igraph_vector_int_t *roots, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_is_forest(g, result, roots, mode));
}
igraph_error_t go_igraph_minimum_spanning_tree(const igraph_t *g, igraph_vector_int_t *result, const igraph_vector_t *weights, igraph_mst_algorithm_t method) {
    GO_IGRAPH_CALL(igraph_minimum_spanning_tree(g, result, weights, method));
}
igraph_error_t go_igraph_unfold_tree(const igraph_t *g, igraph_t *tree, igraph_neimode_t mode, const igraph_vector_int_t *roots, igraph_vector_int_t *vertex_index) {
    GO_IGRAPH_CALL(igraph_unfold_tree(g, tree, mode, roots, vertex_index));
}
