#include "modularity_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_coreness(
    const igraph_t *graph,
    igraph_vector_int_t *cores,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_coreness(graph, cores, mode));
}

igraph_error_t go_igraph_trussness(
    const igraph_t *graph,
    igraph_vector_int_t *trussness) {
    GO_IGRAPH_CALL(igraph_trussness(graph, trussness));
}

igraph_error_t go_igraph_modularity(
    const igraph_t *graph,
    const igraph_vector_int_t *membership,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_bool_t directed,
    igraph_real_t *modularity) {
    GO_IGRAPH_CALL(igraph_modularity(
        graph, membership, weights, resolution, directed, modularity));
}

igraph_error_t go_igraph_modularity_matrix(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_matrix_t *modmat,
    igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_modularity_matrix(
        graph, weights, resolution, modmat, directed));
}
