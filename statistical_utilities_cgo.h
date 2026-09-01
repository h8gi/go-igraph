#ifndef GO_IGRAPH_STATISTICAL_UTILITIES_CGO_H
#define GO_IGRAPH_STATISTICAL_UTILITIES_CGO_H

#include <stdint.h>
#include <igraph.h>

igraph_error_t go_igraph_running_mean(const igraph_vector_t *, igraph_int_t, igraph_vector_t *);
igraph_error_t go_igraph_random_sample(
    igraph_int_t, igraph_int_t, igraph_int_t, igraph_bool_t, uint64_t, igraph_vector_int_t *);

#endif
