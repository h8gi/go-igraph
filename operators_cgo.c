#include "operators_cgo.h"
#include "igraph_error_cgo.h"

#include <stdlib.h>

igraph_t *go_igraph_graph_array_alloc(igraph_int_t count) {
    if (count == 0) {
        return NULL;
    }
    return calloc((size_t) count, sizeof(igraph_t));
}

void go_igraph_graph_array_set(
    igraph_t *graphs, igraph_int_t index, const igraph_t *graph) {
    graphs[index] = *graph;
}

typedef igraph_error_t (*go_igraph_many_operator_t)(
    igraph_t *, const igraph_vector_ptr_t *, igraph_vector_int_list_t *);

static igraph_error_t go_igraph_call_many_operator(
    igraph_t *result, const igraph_t *graphs, igraph_int_t count,
    igraph_vector_int_list_t *maps, go_igraph_many_operator_t operation) {
    igraph_vector_ptr_t pointers;
    igraph_error_t code = igraph_vector_ptr_init(&pointers, count);
    if (code != IGRAPH_SUCCESS) {
        return code;
    }
    for (igraph_int_t index = 0; index < count; ++index) {
        VECTOR(pointers)[index] = (void *) &graphs[index];
    }
    code = operation(result, &pointers, maps);
    igraph_vector_ptr_destroy(&pointers);
    return code;
}

static igraph_error_t go_igraph_disjoint_union_many_adapter(
    igraph_t *result, const igraph_vector_ptr_t *graphs,
    igraph_vector_int_list_t *unused) {
    (void) unused;
    return igraph_disjoint_union_many(result, graphs);
}

igraph_error_t go_igraph_disjoint_union_many(
    igraph_t *result, const igraph_t *graphs, igraph_int_t count) {
    GO_IGRAPH_CALL(go_igraph_call_many_operator(
        result, graphs, count, NULL, go_igraph_disjoint_union_many_adapter));
}

igraph_error_t go_igraph_union_many(
    igraph_t *result, const igraph_t *graphs, igraph_int_t count,
    igraph_vector_int_list_t *maps) {
    GO_IGRAPH_CALL(go_igraph_call_many_operator(
        result, graphs, count, maps, igraph_union_many));
}

igraph_error_t go_igraph_intersection_many(
    igraph_t *result, const igraph_t *graphs, igraph_int_t count,
    igraph_vector_int_list_t *maps) {
    GO_IGRAPH_CALL(go_igraph_call_many_operator(
        result, graphs, count, maps, igraph_intersection_many));
}

igraph_error_t go_igraph_graph_power(
    const igraph_t *graph, igraph_t *result, igraph_int_t order,
    igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_graph_power(graph, result, order, directed));
}

igraph_error_t go_igraph_connect_neighborhood(
    igraph_t *graph, igraph_int_t order, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_connect_neighborhood(graph, order, mode));
}

igraph_error_t go_igraph_contract_vertices(
    igraph_t *graph, const igraph_vector_int_t *mapping,
    const igraph_attribute_combination_t *vertex_combination) {
    GO_IGRAPH_CALL(igraph_contract_vertices(
        graph, mapping, vertex_combination));
}

igraph_error_t go_igraph_reverse_edges(
    igraph_t *graph, igraph_es_t edges) {
    GO_IGRAPH_CALL(igraph_reverse_edges(graph, edges));
}

igraph_error_t go_igraph_join(
    igraph_t *result, const igraph_t *left, const igraph_t *right) {
    GO_IGRAPH_CALL(igraph_join(result, left, right));
}

igraph_error_t go_igraph_product(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_product_t type) {
    GO_IGRAPH_CALL(igraph_product(result, left, right, type));
}

igraph_error_t go_igraph_rooted_product(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_int_t root) {
    GO_IGRAPH_CALL(igraph_rooted_product(result, left, right, root));
}

igraph_error_t go_igraph_disjoint_union(
    igraph_t *result, const igraph_t *left, const igraph_t *right) {
    GO_IGRAPH_CALL(igraph_disjoint_union(result, left, right));
}

igraph_error_t go_igraph_union(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_CALL(igraph_union(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_intersection(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_CALL(igraph_intersection(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_difference(
    igraph_t *result, const igraph_t *left, const igraph_t *right) {
    GO_IGRAPH_CALL(igraph_difference(result, left, right));
}

igraph_error_t go_igraph_compose(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_CALL(igraph_compose(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_complementer(
    igraph_t *result, const igraph_t *graph, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_complementer(result, graph, loops));
}

igraph_error_t go_igraph_has_multiple(
    const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_has_multiple(graph, result));
}
