#ifndef GO_IGRAPH_TREE_CONSTRUCTION_CGO_H
#define GO_IGRAPH_TREE_CONSTRUCTION_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_from_prufer(igraph_t *graph, const igraph_vector_int_t *prufer);
igraph_error_t go_igraph_to_prufer(const igraph_t *graph, igraph_vector_int_t *prufer);
igraph_error_t go_igraph_tree_from_parent_vector(igraph_t *graph, const igraph_vector_int_t *parents, igraph_tree_mode_t mode);
igraph_error_t go_igraph_regular_tree(igraph_t *graph, igraph_int_t height, igraph_int_t degree, igraph_tree_mode_t mode);
igraph_error_t go_igraph_symmetric_tree(igraph_t *graph, const igraph_vector_int_t *branches, igraph_tree_mode_t mode);

#endif
