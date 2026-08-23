#ifndef GO_IGRAPH_SEQUENCE_FAMILIES_CGO_H
#define GO_IGRAPH_SEQUENCE_FAMILIES_CGO_H
#include <igraph.h>
igraph_error_t go_igraph_hexagonal_lattice(igraph_t *, const igraph_vector_int_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_triangular_lattice(igraph_t *, const igraph_vector_int_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_de_bruijn(igraph_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_kautz(igraph_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_lcf(igraph_t *, igraph_int_t, const igraph_vector_int_t *, igraph_int_t);
#endif
