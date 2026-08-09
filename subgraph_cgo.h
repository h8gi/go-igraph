#ifndef GO_IGRAPH_SUBGRAPH_CGO_H
#define GO_IGRAPH_SUBGRAPH_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_induced_subgraph_map(
    const igraph_t *, igraph_t *, igraph_vs_t, igraph_vector_int_t *,
    igraph_vector_int_t *);
igraph_error_t go_igraph_subgraph_from_edges(
    const igraph_t *, igraph_t *, igraph_es_t, igraph_bool_t);
igraph_error_t go_igraph_decompose(
    const igraph_t *, igraph_graph_list_t *, igraph_connectedness_t,
    igraph_int_t, igraph_int_t);

#endif
