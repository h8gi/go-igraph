#ifndef GO_IGRAPH_REACHABILITY_CGO_H
#define GO_IGRAPH_REACHABILITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_reachability(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *,
    igraph_int_t *, igraph_vector_int_list_t *, igraph_neimode_t);
igraph_error_t go_igraph_count_reachable(
    const igraph_t *, igraph_vector_int_t *, igraph_neimode_t);
igraph_error_t go_igraph_neighborhood_graphs(
    const igraph_t *, igraph_graph_list_t *, igraph_vs_t, igraph_int_t,
    igraph_neimode_t, igraph_int_t);
igraph_error_t go_igraph_transitive_closure(const igraph_t *, igraph_t *);

#endif
