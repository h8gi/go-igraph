#ifndef GO_IGRAPH_EDGE_DIAGNOSTICS_CGO_H
#define GO_IGRAPH_EDGE_DIAGNOSTICS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_has_loop(const igraph_t *, igraph_bool_t *);
igraph_error_t go_igraph_count_loops(const igraph_t *, igraph_int_t *);
igraph_error_t go_igraph_is_loop(
    const igraph_t *, igraph_vector_bool_t *, igraph_es_t);
igraph_error_t go_igraph_count_multiple(
    const igraph_t *, igraph_vector_int_t *, igraph_es_t);
igraph_error_t go_igraph_is_multiple(
    const igraph_t *, igraph_vector_bool_t *, igraph_es_t);
igraph_error_t go_igraph_has_mutual(
    const igraph_t *, igraph_bool_t *, igraph_bool_t);
igraph_error_t go_igraph_is_mutual(
    const igraph_t *, igraph_vector_bool_t *, igraph_es_t, igraph_bool_t);

#endif
