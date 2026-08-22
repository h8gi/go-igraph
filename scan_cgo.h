#ifndef GO_IGRAPH_SCAN_CGO_H
#define GO_IGRAPH_SCAN_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_local_scan_0(
    const igraph_t *, igraph_vector_t *, const igraph_vector_t *, igraph_neimode_t);
igraph_error_t go_igraph_local_scan_1_ecount(
    const igraph_t *, igraph_vector_t *, const igraph_vector_t *, igraph_neimode_t);
igraph_error_t go_igraph_local_scan_k_ecount(
    const igraph_t *, igraph_int_t, igraph_vector_t *, const igraph_vector_t *, igraph_neimode_t);
igraph_error_t go_igraph_local_scan_subset_ecount(
    const igraph_t *, igraph_vector_t *, const igraph_vector_t *,
    const igraph_vector_int_list_t *);
igraph_error_t go_igraph_scan_list_append_copy(
    igraph_vector_int_list_t *, const igraph_vector_int_t *);
igraph_error_t go_igraph_local_scan_0_them(
    const igraph_t *, const igraph_t *, igraph_vector_t *,
    const igraph_vector_t *, igraph_neimode_t);
igraph_error_t go_igraph_local_scan_1_ecount_them(
    const igraph_t *, const igraph_t *, igraph_vector_t *,
    const igraph_vector_t *, igraph_neimode_t);
igraph_error_t go_igraph_local_scan_k_ecount_them(
    const igraph_t *, const igraph_t *, igraph_int_t, igraph_vector_t *,
    const igraph_vector_t *, igraph_neimode_t);

#endif
