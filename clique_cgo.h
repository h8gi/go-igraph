#ifndef GO_IGRAPH_CLIQUE_CGO_H
#define GO_IGRAPH_CLIQUE_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_complete(const igraph_t *, igraph_bool_t *);
igraph_error_t go_igraph_is_clique(const igraph_t *, igraph_vs_t,
                                   igraph_bool_t, igraph_bool_t *);
igraph_error_t go_igraph_is_independent_vertex_set(
    const igraph_t *, igraph_vs_t, igraph_bool_t *);
igraph_error_t go_igraph_clique_number(const igraph_t *, igraph_int_t *);
igraph_error_t go_igraph_independence_number(const igraph_t *, igraph_int_t *);

#endif
