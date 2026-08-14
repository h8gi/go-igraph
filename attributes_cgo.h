#ifndef GO_IGRAPH_ATTRIBUTES_CGO_H
#define GO_IGRAPH_ATTRIBUTES_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_install_cattribute_table(void);
igraph_bool_t go_igraph_cattribute_table_is_installed(void);

igraph_error_t go_igraph_attribute_combination_init(
    igraph_attribute_combination_t *combination);
igraph_error_t go_igraph_attribute_combination_add(
    igraph_attribute_combination_t *combination,
    const char *name,
    igraph_attribute_combination_type_t type);

igraph_error_t go_igraph_attribute_record_init(
    igraph_attribute_record_t *record,
    const char *name,
    igraph_attribute_type_t type);
igraph_error_t go_igraph_attribute_record_check_type(
    const igraph_attribute_record_t *record,
    igraph_attribute_type_t type);
igraph_error_t go_igraph_attribute_record_resize(
    igraph_attribute_record_t *record,
    igraph_int_t size);

igraph_error_t go_igraph_attribute_record_list_init(
    igraph_attribute_record_list_t *list,
    igraph_int_t size);

igraph_error_t go_igraph_cattribute_list_scope(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    igraph_strvector_t *names,
    igraph_vector_int_t *types);
igraph_error_t go_igraph_cattribute_GAN_set(
    igraph_t *graph,
    const char *name,
    igraph_real_t value);
igraph_error_t go_igraph_cattribute_GAS_set(
    igraph_t *graph,
    const char *name,
    const char *value);
igraph_error_t go_igraph_cattribute_GAB_set(
    igraph_t *graph,
    const char *name,
    igraph_bool_t value);

igraph_error_t go_igraph_cattribute_numeric_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_vector_t *result);
igraph_error_t go_igraph_cattribute_string_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_strvector_t *result);
igraph_error_t go_igraph_cattribute_boolean_values(
    const igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_vector_bool_t *result);
igraph_error_t go_igraph_cattribute_numeric_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    igraph_real_t value);
igraph_error_t go_igraph_cattribute_string_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    const char *value);
igraph_error_t go_igraph_cattribute_boolean_set(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    igraph_int_t id,
    igraph_bool_t value);
igraph_error_t go_igraph_cattribute_numeric_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_vector_t *values);
igraph_error_t go_igraph_cattribute_string_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_strvector_t *values);
igraph_error_t go_igraph_cattribute_boolean_setv(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name,
    const igraph_vector_bool_t *values);
void go_igraph_cattribute_remove(
    igraph_t *graph,
    igraph_attribute_elemtype_t scope,
    const char *name);

#endif
