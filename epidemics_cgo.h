#ifndef GO_IGRAPH_EPIDEMICS_CGO_H
#define GO_IGRAPH_EPIDEMICS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_sir_results_init(igraph_vector_ptr_t *result);
igraph_error_t go_igraph_sir_run(const igraph_t *graph, igraph_real_t beta,
                                 igraph_real_t gamma, igraph_int_t runs,
                                 igraph_vector_ptr_t *result);
igraph_int_t go_igraph_sir_results_size(const igraph_vector_ptr_t *result);
igraph_sir_t *go_igraph_sir_results_get(const igraph_vector_ptr_t *result,
                                        igraph_int_t index);
void go_igraph_sir_results_destroy(igraph_vector_ptr_t *result);

#endif
