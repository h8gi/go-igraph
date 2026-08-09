#ifndef GO_IGRAPH_MODULARITY_CGO_H
#define GO_IGRAPH_MODULARITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_coreness(
    const igraph_t *graph,
    igraph_vector_int_t *cores,
    igraph_neimode_t mode);

igraph_error_t go_igraph_trussness(
    const igraph_t *graph,
    igraph_vector_int_t *trussness);

igraph_error_t go_igraph_modularity(
    const igraph_t *graph,
    const igraph_vector_int_t *membership,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_bool_t directed,
    igraph_real_t *modularity);

igraph_error_t go_igraph_modularity_matrix(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_real_t resolution,
    igraph_matrix_t *modmat,
    igraph_bool_t directed);

#endif
