#ifndef GO_IGRAPH_COMMUNITY_CGO_H
#define GO_IGRAPH_COMMUNITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_reindex_membership(
    igraph_vector_int_t *membership,
    igraph_vector_int_t *new_to_old,
    igraph_int_t *nb_clusters);

igraph_error_t go_igraph_community_to_membership(
    const igraph_matrix_int_t *merges,
    igraph_int_t nodes,
    igraph_int_t steps,
    igraph_vector_int_t *membership,
    igraph_vector_int_t *csize);

igraph_error_t go_igraph_le_community_to_membership(
    const igraph_matrix_int_t *merges,
    igraph_int_t steps,
    igraph_vector_int_t *membership,
    igraph_vector_int_t *csize);

igraph_error_t go_igraph_rng_seed(uint64_t seed);

igraph_error_t go_igraph_matrix_int_init(
    igraph_matrix_int_t *m,
    igraph_int_t nrow,
    igraph_int_t ncol);

void go_igraph_matrix_int_destroy(igraph_matrix_int_t *m);

void go_igraph_matrix_int_set(
    igraph_matrix_int_t *m,
    igraph_int_t row,
    igraph_int_t col,
    igraph_int_t value);

igraph_int_t go_igraph_matrix_int_get(
    const igraph_matrix_int_t *m,
    igraph_int_t row,
    igraph_int_t col);

igraph_int_t go_igraph_matrix_int_nrow(const igraph_matrix_int_t *m);
igraph_int_t go_igraph_matrix_int_ncol(const igraph_matrix_int_t *m);

#endif
