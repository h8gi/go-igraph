#include "assignment_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_solve_lsap(const igraph_matrix_t *costs, igraph_int_t size, igraph_vector_int_t *result) {
    GO_IGRAPH_CALL(igraph_solve_lsap(costs, size, result));
}
