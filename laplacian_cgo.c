#include "laplacian_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_get_laplacian(const igraph_t *graph, igraph_matrix_t *result, igraph_neimode_t mode, igraph_laplacian_normalization_t normalization, const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_get_laplacian(graph, result, mode, normalization, weights));
}
