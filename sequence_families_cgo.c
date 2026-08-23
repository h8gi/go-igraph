#include "sequence_families_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_hexagonal_lattice(igraph_t *g, const igraph_vector_int_t *d, igraph_bool_t directed, igraph_bool_t mutual) {
  GO_IGRAPH_CALL(igraph_hexagonal_lattice(g, d, directed, mutual));
}
igraph_error_t go_igraph_triangular_lattice(igraph_t *g, const igraph_vector_int_t *d, igraph_bool_t directed, igraph_bool_t mutual) {
  GO_IGRAPH_CALL(igraph_triangular_lattice(g, d, directed, mutual));
}
igraph_error_t go_igraph_de_bruijn(igraph_t *g, igraph_int_t alphabet, igraph_int_t length) {
  GO_IGRAPH_CALL(igraph_de_bruijn(g, alphabet, length));
}
igraph_error_t go_igraph_kautz(igraph_t *g, igraph_int_t degree, igraph_int_t order) {
  GO_IGRAPH_CALL(igraph_kautz(g, degree, order));
}
igraph_error_t go_igraph_lcf(igraph_t *g, igraph_int_t vertices, const igraph_vector_int_t *shifts, igraph_int_t repeats) {
  GO_IGRAPH_CALL(igraph_lcf(g, vertices, shifts, repeats));
}
