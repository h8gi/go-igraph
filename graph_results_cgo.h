#ifndef GO_IGRAPH_GRAPH_RESULTS_CGO_H
#define GO_IGRAPH_GRAPH_RESULTS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_residual_graph(
    const igraph_t *graph,
    const igraph_vector_t *capacity,
    igraph_t *residual,
    igraph_vector_t *residual_capacity,
    const igraph_vector_t *flow);

igraph_error_t go_igraph_reverse_residual_graph(
    const igraph_t *graph,
    const igraph_vector_t *capacity,
    igraph_t *residual,
    const igraph_vector_t *flow);

igraph_error_t go_igraph_gomory_hu_tree(
    const igraph_t *graph,
    igraph_t *tree,
    igraph_vector_t *flows,
    const igraph_vector_t *capacity);

igraph_error_t go_igraph_dominator_tree(
    const igraph_t *graph,
    igraph_int_t root,
    igraph_vector_int_t *dom,
    igraph_t *domtree,
    igraph_vector_int_t *leftout,
    igraph_neimode_t mode);

igraph_error_t go_igraph_even_tarjan_reduction(
    const igraph_t *graph,
    igraph_t *graphbar,
    igraph_vector_t *capacity);

#endif
