#ifndef GO_IGRAPH_PERCOLATION_CGO_H
#define GO_IGRAPH_PERCOLATION_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_bond_percolation(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    const igraph_vector_int_t *);
igraph_error_t go_igraph_site_percolation(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    const igraph_vector_int_t *);
igraph_error_t go_igraph_edgelist_percolation(
    const igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_vector_int_t *);

#endif
