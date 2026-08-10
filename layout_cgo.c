#include "layout_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_layout_circle(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_vs_t order) {
    GO_IGRAPH_CALL(igraph_layout_circle(graph, res, order));
}

igraph_error_t go_igraph_layout_star(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t center,
    const igraph_vector_int_t *order) {
    GO_IGRAPH_CALL(igraph_layout_star(graph, res, center, order));
}

igraph_error_t go_igraph_layout_grid(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width) {
    GO_IGRAPH_CALL(igraph_layout_grid(graph, res, width));
}

igraph_error_t go_igraph_layout_random(
    const igraph_t *graph,
    igraph_matrix_t *res) {
    GO_IGRAPH_CALL(igraph_layout_random(graph, res));
}

igraph_error_t go_igraph_layout_reingold_tilford(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel) {
    GO_IGRAPH_CALL(igraph_layout_reingold_tilford(graph, res, mode, roots, rootlevel));
}

igraph_error_t go_igraph_layout_reingold_tilford_circular(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel) {
    GO_IGRAPH_CALL(igraph_layout_reingold_tilford_circular(graph, res, mode, roots, rootlevel));
}

igraph_error_t go_igraph_layout_bipartite(
    const igraph_t *graph,
    const igraph_vector_bool_t *types,
    igraph_matrix_t *res,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter) {
    GO_IGRAPH_CALL(igraph_layout_bipartite(graph, types, res, hgap, vgap, maxiter));
}

igraph_error_t go_igraph_layout_sugiyama(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_matrix_list_t *routing,
    const igraph_vector_int_t *layers,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter,
    const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_layout_sugiyama(
        graph, res, routing, layers, hgap, vgap, maxiter, weights));
}
