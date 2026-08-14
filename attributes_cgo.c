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

igraph_error_t go_igraph_cattribute_list_scope(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    igraph_strvector_t *names,
    igraph_vector_int_t *types) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_GRAPH:
        GO_IGRAPH_CALL(igraph_cattribute_list(graph, names, types, NULL, NULL, NULL, NULL));
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_list(graph, NULL, NULL, names, types, NULL, NULL));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_list(graph, NULL, NULL, NULL, NULL, names, types));
    default:
        return IGRAPH_EINVAL;
    }
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

igraph_error_t go_igraph_cattribute_numeric_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_vector_t *result) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(
            igraph_cattribute_VANV(graph, name, igraph_vss_all(), result));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EANV(
            graph, name, igraph_ess_all(IGRAPH_EDGEORDER_ID), result));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_string_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_strvector_t *result) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(
            igraph_cattribute_VASV(graph, name, igraph_vss_all(), result));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EASV(
            graph, name, igraph_ess_all(IGRAPH_EDGEORDER_ID), result));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_boolean_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_vector_bool_t *result) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(
            igraph_cattribute_VABV(graph, name, igraph_vss_all(), result));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EABV(
            graph, name, igraph_ess_all(IGRAPH_EDGEORDER_ID), result));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_numeric_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    igraph_real_t value) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAN_set(graph, name, id, value));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAN_set(graph, name, id, value));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_string_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    const char *value) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAS_set(graph, name, id, value));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAS_set(graph, name, id, value));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_boolean_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    igraph_bool_t value) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAB_set(graph, name, id, value));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAB_set(graph, name, id, value));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_numeric_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_vector_t *values) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAN_setv(graph, name, values));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAN_setv(graph, name, values));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_string_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_strvector_t *values) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAS_setv(graph, name, values));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAS_setv(graph, name, values));
    default:
        return IGRAPH_EINVAL;
    }
}

igraph_error_t go_igraph_cattribute_boolean_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_vector_bool_t *values) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        GO_IGRAPH_CALL(igraph_cattribute_VAB_setv(graph, name, values));
    case IGRAPH_ATTRIBUTE_EDGE:
        GO_IGRAPH_CALL(igraph_cattribute_EAB_setv(graph, name, values));
    default:
        return IGRAPH_EINVAL;
    }
}

void go_igraph_cattribute_remove(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name) {
    switch (scope) {
    case IGRAPH_ATTRIBUTE_VERTEX:
        igraph_cattribute_remove_v(graph, name);
        break;
    case IGRAPH_ATTRIBUTE_EDGE:
        igraph_cattribute_remove_e(graph, name);
        break;
    default:
        break;
    }
}
