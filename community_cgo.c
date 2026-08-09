#include "community_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_reindex_membership(
    igraph_vector_int_t *membership,
    igraph_vector_int_t *new_to_old,
    igraph_int_t *nb_clusters) {
    GO_IGRAPH_CALL(igraph_reindex_membership(membership, new_to_old, nb_clusters));
}

igraph_error_t go_igraph_community_to_membership(
    const igraph_matrix_int_t *merges,
    igraph_int_t nodes,
    igraph_int_t steps,
    igraph_vector_int_t *membership,
    igraph_vector_int_t *csize) {
    GO_IGRAPH_CALL(igraph_community_to_membership(merges, nodes, steps, membership, csize));
}

igraph_error_t go_igraph_le_community_to_membership(
    const igraph_matrix_int_t *merges,
    igraph_int_t steps,
    igraph_vector_int_t *membership,
    igraph_vector_int_t *csize) {
    GO_IGRAPH_CALL(igraph_le_community_to_membership(merges, steps, membership, csize));
}

igraph_error_t go_igraph_rng_seed(uint64_t seed) {
    GO_IGRAPH_CALL(igraph_rng_seed(igraph_rng_default(), (igraph_uint_t)seed));
}

igraph_error_t go_igraph_matrix_int_init(
    igraph_matrix_int_t *m,
    igraph_int_t nrow,
    igraph_int_t ncol) {
    GO_IGRAPH_CALL(igraph_matrix_int_init(m, nrow, ncol));
}

void go_igraph_matrix_int_destroy(igraph_matrix_int_t *m) {
    igraph_matrix_int_destroy(m);
}

void go_igraph_matrix_int_set(
    igraph_matrix_int_t *m,
    igraph_int_t row,
    igraph_int_t col,
    igraph_int_t value) {
    igraph_matrix_int_set(m, row, col, value);
}

igraph_int_t go_igraph_matrix_int_get(
    const igraph_matrix_int_t *m,
    igraph_int_t row,
    igraph_int_t col) {
    return igraph_matrix_int_get(m, row, col);
}

igraph_int_t go_igraph_matrix_int_nrow(const igraph_matrix_int_t *m) {
    return m ? m->nrow : 0;
}

igraph_int_t go_igraph_matrix_int_ncol(const igraph_matrix_int_t *m) {
    return m ? m->ncol : 0;
}
