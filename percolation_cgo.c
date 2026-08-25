#include "percolation_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_bond_percolation(
        const igraph_t *graph, igraph_vector_int_t *giant_size,
        igraph_vector_int_t *vertex_count,
        const igraph_vector_int_t *edge_order) {
    GO_IGRAPH_CALL(igraph_bond_percolation(
        graph, giant_size, vertex_count, edge_order));
}

igraph_error_t go_igraph_site_percolation(
        const igraph_t *graph, igraph_vector_int_t *giant_size,
        igraph_vector_int_t *edge_count,
        const igraph_vector_int_t *vertex_order) {
    GO_IGRAPH_CALL(igraph_site_percolation(
        graph, giant_size, edge_count, vertex_order));
}

igraph_error_t go_igraph_edgelist_percolation(
        const igraph_vector_int_t *edges,
        igraph_vector_int_t *giant_size,
        igraph_vector_int_t *vertex_count) {
    GO_IGRAPH_CALL(igraph_edgelist_percolation(
        edges, giant_size, vertex_count));
}
