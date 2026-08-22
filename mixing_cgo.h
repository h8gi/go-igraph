#ifndef GO_IGRAPH_MIXING_CGO_H
#define GO_IGRAPH_MIXING_CGO_H

#include <igraph.h>

int go_igraph_is_directed(const igraph_t *);

igraph_error_t go_igraph_assortativity_nominal(
    const igraph_t *, const igraph_vector_int_t *, igraph_real_t *,
    igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_assortativity(
    const igraph_t *, const igraph_vector_t *, const igraph_vector_t *,
    const igraph_vector_t *, igraph_real_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_assortativity_degree(
    const igraph_t *, igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_joint_type_distribution(
    const igraph_t *, const igraph_vector_t *, igraph_matrix_t *,
    const igraph_vector_int_t *, const igraph_vector_int_t *,
    igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_joint_degree_distribution(
    const igraph_t *, const igraph_vector_t *, igraph_matrix_t *,
    igraph_neimode_t, igraph_neimode_t, igraph_bool_t, igraph_bool_t,
    igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_joint_degree_matrix(
    const igraph_t *, const igraph_vector_t *, igraph_matrix_t *,
    igraph_int_t, igraph_int_t);

#endif
