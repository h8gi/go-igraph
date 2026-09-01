#ifndef GO_IGRAPH_ASSIGNMENT_CGO_H
#define GO_IGRAPH_ASSIGNMENT_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_solve_lsap(const igraph_matrix_t *, igraph_int_t, igraph_vector_int_t *);

#endif
