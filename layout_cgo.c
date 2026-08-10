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
