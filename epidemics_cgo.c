#include <stdlib.h>

#include "epidemics_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_sir_results_init(igraph_vector_ptr_t *result) {
    GO_IGRAPH_CALL(igraph_vector_ptr_init(result, 0));
}

igraph_error_t go_igraph_sir_run(const igraph_t *graph, igraph_real_t beta,
                                 igraph_real_t gamma, igraph_int_t runs,
                                 igraph_vector_ptr_t *result) {
    GO_IGRAPH_CALL(igraph_sir(graph, beta, gamma, runs, result));
}

igraph_int_t go_igraph_sir_results_size(const igraph_vector_ptr_t *result) {
    return igraph_vector_ptr_size(result);
}

igraph_sir_t *go_igraph_sir_results_get(const igraph_vector_ptr_t *result,
                                        igraph_int_t index) {
    return (igraph_sir_t *) VECTOR(*result)[index];
}

void go_igraph_sir_results_destroy(igraph_vector_ptr_t *result) {
    igraph_int_t i, n = igraph_vector_ptr_size(result);
    for (i = 0; i < n; ++i) {
        igraph_sir_t *sir = (igraph_sir_t *) VECTOR(*result)[i];
        if (sir != NULL) {
            igraph_sir_destroy(sir);
            free(sir);
        }
    }
    igraph_vector_ptr_destroy(result);
}
