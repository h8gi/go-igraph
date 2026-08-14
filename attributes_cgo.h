#ifndef GO_IGRAPH_ATTRIBUTES_CGO_H
#define GO_IGRAPH_ATTRIBUTES_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_install_cattribute_table(void);
igraph_bool_t go_igraph_cattribute_table_is_installed(void);

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

#endif
