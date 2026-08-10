#include "graph_results_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_residual_graph(
    const igraph_t *graph,
    const igraph_vector_t *capacity,
    igraph_t *residual,
    igraph_vector_t *residual_capacity,
    const igraph_vector_t *flow) {
    GO_IGRAPH_CALL(igraph_residual_graph(graph, capacity, residual, residual_capacity, flow));
}

igraph_error_t go_igraph_reverse_residual_graph(
    const igraph_t *graph,
    const igraph_vector_t *capacity,
    igraph_t *residual,
    const igraph_vector_t *flow) {
    GO_IGRAPH_CALL(igraph_reverse_residual_graph(graph, capacity, residual, flow));
}

igraph_error_t go_igraph_gomory_hu_tree(
    const igraph_t *graph,
    igraph_t *tree,
    igraph_vector_t *flows,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_gomory_hu_tree(graph, tree, flows, capacity));
}

igraph_error_t go_igraph_dominator_tree(
    const igraph_t *graph,
    igraph_int_t root,
    igraph_vector_int_t *dom,
    igraph_t *domtree,
    igraph_vector_int_t *leftout,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_dominator_tree(graph, root, dom, domtree, leftout, mode));
}

igraph_error_t go_igraph_even_tarjan_reduction(
    const igraph_t *graph,
    igraph_t *graphbar,
    igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_even_tarjan_reduction(graph, graphbar, capacity));
}
