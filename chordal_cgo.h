#ifndef GO_IGRAPH_CHORDAL_CGO_H
#define GO_IGRAPH_CHORDAL_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_maximum_cardinality_search(const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_is_chordal(const igraph_t *, const igraph_vector_int_t *, const igraph_vector_int_t *, igraph_bool_t *, igraph_vector_int_t *, igraph_t *);
igraph_error_t go_igraph_is_perfect(const igraph_t *, igraph_bool_t *);
igraph_error_t go_igraph_is_simple_for_perfect(const igraph_t *, igraph_bool_t *);

#endif
