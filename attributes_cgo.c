#include "attributes_cgo.h"
#include "igraph_error_cgo.h"

/*
 * The attribute handler is process-global. This function is called during Go
 * package initialization, before callers can construct a graph, and never
 * replaces a handler installed by another C consumer.
 */
igraph_error_t go_igraph_install_cattribute_table(void) {
    if (igraph_has_attribute_table()) {
        return IGRAPH_EINVAL;
    }

    igraph_set_attribute_table(&igraph_cattribute_table);
    return IGRAPH_SUCCESS;
}

igraph_bool_t go_igraph_cattribute_table_is_installed(void) {
    return igraph_has_attribute_table();
}

igraph_error_t go_igraph_attribute_record_init(
    igraph_attribute_record_t *record,
    const char *name,
    igraph_attribute_type_t type) {
    GO_IGRAPH_CALL(igraph_attribute_record_init(record, name, type));
}

igraph_error_t go_igraph_attribute_record_check_type(
    const igraph_attribute_record_t *record,
    igraph_attribute_type_t type) {
    GO_IGRAPH_CALL(igraph_attribute_record_check_type(record, type));
}

igraph_error_t go_igraph_attribute_record_resize(
    igraph_attribute_record_t *record,
    igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_attribute_record_resize(record, size));
}

igraph_error_t go_igraph_attribute_record_list_init(
    igraph_attribute_record_list_t *list,
    igraph_int_t size) {
    GO_IGRAPH_CALL(igraph_attribute_record_list_init(list, size));
}

igraph_error_t go_igraph_cattribute_list_graph(
    const igraph_t *graph,
    igraph_strvector_t *names,
    igraph_vector_int_t *types) {
    GO_IGRAPH_CALL(igraph_cattribute_list(graph, names, types, NULL, NULL, NULL, NULL));
}

igraph_error_t go_igraph_cattribute_GAN_set(
    igraph_t *graph,
    const char *name,
    igraph_real_t value) {
    GO_IGRAPH_CALL(igraph_cattribute_GAN_set(graph, name, value));
}

igraph_error_t go_igraph_cattribute_GAS_set(
    igraph_t *graph,
    const char *name,
    const char *value) {
    GO_IGRAPH_CALL(igraph_cattribute_GAS_set(graph, name, value));
}

igraph_error_t go_igraph_cattribute_GAB_set(
    igraph_t *graph,
    const char *name,
    igraph_bool_t value) {
    GO_IGRAPH_CALL(igraph_cattribute_GAB_set(graph, name, value));
}
