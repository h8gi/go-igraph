#ifndef GO_IGRAPH_TREE_ANALYSIS_CGO_H
#define GO_IGRAPH_TREE_ANALYSIS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_tree(const igraph_t *, igraph_bool_t *, igraph_int_t *, igraph_neimode_t);
igraph_error_t go_igraph_is_forest(const igraph_t *, igraph_bool_t *, igraph_vector_int_t *, igraph_neimode_t);
igraph_error_t go_igraph_minimum_spanning_tree(const igraph_t *, igraph_vector_int_t *, const igraph_vector_t *, igraph_mst_algorithm_t);
igraph_error_t go_igraph_unfold_tree(const igraph_t *, igraph_t *, igraph_neimode_t, const igraph_vector_int_t *, igraph_vector_int_t *);

#endif
