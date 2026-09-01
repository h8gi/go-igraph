#ifndef GO_IGRAPH_POWER_LAW_CGO_H
#define GO_IGRAPH_POWER_LAW_CGO_H

#include <stdint.h>
#include <igraph.h>

igraph_error_t go_igraph_power_law_fit(
    const igraph_vector_t *, igraph_real_t, igraph_bool_t,
    igraph_bool_t *, igraph_real_t *, igraph_real_t *, igraph_real_t *, igraph_real_t *);

igraph_error_t go_igraph_power_law_p_value(
    const igraph_vector_t *, igraph_bool_t, igraph_real_t, igraph_real_t,
    igraph_real_t, igraph_real_t, igraph_real_t, igraph_bool_t, uint64_t, igraph_real_t *);

#endif
