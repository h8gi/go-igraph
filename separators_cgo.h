#ifndef GO_IGRAPH_SEPARATORS_CGO_H
#define GO_IGRAPH_SEPARATORS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_separator(
    const igraph_t *, igraph_vs_t, igraph_bool_t *);
igraph_error_t go_igraph_is_minimal_separator(
    const igraph_t *, igraph_vs_t, igraph_bool_t *);

#endif
