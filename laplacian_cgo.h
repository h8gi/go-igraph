#ifndef GO_IGRAPH_LAPLACIAN_CGO_H
#define GO_IGRAPH_LAPLACIAN_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_get_laplacian(const igraph_t *, igraph_matrix_t *, igraph_neimode_t, igraph_laplacian_normalization_t, const igraph_vector_t *);

#endif
