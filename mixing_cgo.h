#ifndef GO_IGRAPH_MIXING_CGO_H
#define GO_IGRAPH_MIXING_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_assortativity_nominal(
    const igraph_t *, const igraph_vector_int_t *, igraph_real_t *,
    igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_assortativity(
    const igraph_t *, const igraph_vector_t *, const igraph_vector_t *,
    const igraph_vector_t *, igraph_real_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_assortativity_degree(
    const igraph_t *, igraph_real_t *, igraph_bool_t);

#endif
