package igraph

// #cgo pkg-config: igraph
// #include <stdlib.h>
// #include <igraph.h>
// #include "attributes_cgo.h"
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"
)

// AttributeScope identifies whether metadata belongs to the graph itself, its
// vertices, or its edges. Its zero value is AttributeGraph.
type AttributeScope uint8

const (
	AttributeGraph AttributeScope = iota
	AttributeVertex
	AttributeEdge
)

func (scope AttributeScope) cValue() (C.igraph_attribute_elemtype_t, error) {
	switch scope {
	case AttributeGraph:
		return C.IGRAPH_ATTRIBUTE_GRAPH, nil
	case AttributeVertex:
		return C.IGRAPH_ATTRIBUTE_VERTEX, nil
	case AttributeEdge:
		return C.IGRAPH_ATTRIBUTE_EDGE, nil
	default:
		return 0, fmt.Errorf("igraph: invalid attribute scope: %d", scope)
	}
}

// AttributeType identifies one of the scalar metadata types supported by the
// package. The zero value is invalid so omitted or unknown types are rejected.
type AttributeType uint8

const (
	AttributeNumeric AttributeType = iota + 1
	AttributeBoolean
	AttributeString
)

func (attributeType AttributeType) cValue() (C.igraph_attribute_type_t, error) {
	switch attributeType {
	case AttributeNumeric:
		return C.IGRAPH_ATTRIBUTE_NUMERIC, nil
	case AttributeBoolean:
		return C.IGRAPH_ATTRIBUTE_BOOLEAN, nil
	case AttributeString:
		return C.IGRAPH_ATTRIBUTE_STRING, nil
	default:
		return 0, fmt.Errorf("igraph: invalid attribute type: %d", attributeType)
	}
}

func attributeTypeFromC(value int) (AttributeType, error) {
	switch C.igraph_attribute_type_t(value) {
	case C.IGRAPH_ATTRIBUTE_NUMERIC:
		return AttributeNumeric, nil
	case C.IGRAPH_ATTRIBUTE_BOOLEAN:
		return AttributeBoolean, nil
	case C.IGRAPH_ATTRIBUTE_STRING:
		return AttributeString, nil
	default:
		return 0, fmt.Errorf("igraph: unsupported attribute type: %d", value)
	}
}

// AttributeMetadata describes one named typed attribute. Values returned by
// package APIs are Go-owned; Name never aliases C storage.
type AttributeMetadata struct {
	Name  string
	Scope AttributeScope
	Type  AttributeType
}

type attributeRuntimeState struct {
	once sync.Once
	err  error
}

var globalAttributeRuntime attributeRuntimeState

func init() {
	if err := ensureAttributeRuntime(); err != nil {
		panic(err)
	}
}

// ensureAttributeRuntime installs the process-global C attribute table before
// callers can construct graphs. sync.Once makes repeated and concurrent setup
// calls read-only after the first installation.
//
//igraph:internal igraph_set_attribute_table
//igraph:internal igraph_has_attribute_table
func ensureAttributeRuntime() error {
	return ensureAttributeRuntimeWithInstaller(
		&globalAttributeRuntime,
		func() int { return int(C.go_igraph_install_cattribute_table()) },
	)
}

func ensureAttributeRuntimeWithInstaller(
	state *attributeRuntimeState,
	install func() int,
) error {
	state.once.Do(func() {
		if install == nil {
			state.err = fmt.Errorf("igraph: attribute runtime installer is nil")
			return
		}
		if code := install(); code != int(C.IGRAPH_SUCCESS) {
			state.err = igraphError("install attribute runtime", code)
		}
	})
	return state.err
}

func attributeRuntimeInstalled() bool {
	return C.go_igraph_cattribute_table_is_installed() != booltoint(false)
}

func validateAttributeName(name string) error {
	if name == "" {
		return fmt.Errorf("igraph: attribute name must not be empty")
	}
	return validateAttributeString("attribute name", name)
}

func validateAttributeString(description, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("igraph: %s is not valid UTF-8", description)
	}
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("igraph: %s contains an embedded NUL byte", description)
	}
	return nil
}

func attributeMetadataFromSlices(
	scope AttributeScope,
	names []string,
	types []int,
) ([]AttributeMetadata, error) {
	if _, err := scope.cValue(); err != nil {
		return nil, err
	}
	if len(names) != len(types) {
		return nil, fmt.Errorf(
			"igraph: attribute metadata name/type length mismatch: %d != %d",
			len(names), len(types),
		)
	}

	metadata := make([]AttributeMetadata, len(names))
	seen := make(map[string]struct{}, len(names))
	for i, name := range names {
		if err := validateAttributeName(name); err != nil {
			return nil, fmt.Errorf("igraph: attribute metadata at index %d: %w", i, err)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("igraph: duplicate attribute name %q in scope %d", name, scope)
		}
		seen[name] = struct{}{}
		attributeType, err := attributeTypeFromC(types[i])
		if err != nil {
			return nil, fmt.Errorf("igraph: attribute metadata at index %d: %w", i, err)
		}
		metadata[i] = AttributeMetadata{
			Name:  strings.Clone(name),
			Scope: scope,
			Type:  attributeType,
		}
	}
	return metadata, nil
}

type attributeRecord struct {
	value       C.igraph_attribute_record_t
	initialized bool
}

type attributeRecordInitializer func(*attributeRecord, string, AttributeType) int

//igraph:internal igraph_attribute_record_init
func newAttributeRecord(name string, attributeType AttributeType) (*attributeRecord, error) {
	return newAttributeRecordWithInitializer(name, attributeType, initializeAttributeRecord)
}

func newAttributeRecordWithInitializer(
	name string,
	attributeType AttributeType,
	initialize attributeRecordInitializer,
) (*attributeRecord, error) {
	if err := validateAttributeName(name); err != nil {
		return nil, err
	}
	if _, err := attributeType.cValue(); err != nil {
		return nil, err
	}
	if initialize == nil {
		return nil, fmt.Errorf("igraph: attribute record initializer is nil")
	}

	record := &attributeRecord{}
	if code := initialize(record, name, attributeType); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize attribute record", code)
	}
	record.initialized = true
	return record, nil
}

func initializeAttributeRecord(
	record *attributeRecord,
	name string,
	attributeType AttributeType,
) int {
	cType, err := attributeType.cValue()
	if err != nil {
		return int(C.IGRAPH_EINVAL)
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return int(C.go_igraph_attribute_record_init(&record.value, cName, cType))
}

//igraph:internal igraph_attribute_record_check_type
func (record *attributeRecord) checkType(attributeType AttributeType) error {
	if record == nil || !record.initialized {
		return ErrClosed
	}
	cType, err := attributeType.cValue()
	if err != nil {
		return err
	}
	if code := C.go_igraph_attribute_record_check_type(&record.value, cType); code != C.IGRAPH_SUCCESS {
		return igraphError("check attribute record type", int(code))
	}
	return nil
}

//igraph:internal igraph_attribute_record_size
func (record *attributeRecord) size() (int, error) {
	if record == nil || !record.initialized {
		return 0, ErrClosed
	}
	return igraphIntToInt(C.igraph_attribute_record_size(&record.value), "attribute record length")
}

//igraph:internal igraph_attribute_record_resize
func (record *attributeRecord) resize(size int) error {
	if record == nil || !record.initialized {
		return ErrClosed
	}
	cSize, err := intToIgraphInt(size, "attribute record length")
	if err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("igraph: attribute record length must be non-negative: %d", size)
	}
	if code := C.go_igraph_attribute_record_resize(&record.value, cSize); code != C.IGRAPH_SUCCESS {
		return igraphError("resize attribute record", int(code))
	}
	return nil
}

//igraph:internal igraph_attribute_record_destroy
func (record *attributeRecord) close() {
	if record == nil || !record.initialized {
		return
	}
	C.igraph_attribute_record_destroy(&record.value)
	record.initialized = false
}

type attributeRecordList struct {
	value       C.igraph_attribute_record_list_t
	initialized bool
}

type attributeRecordListInitializer func(*attributeRecordList, int) int

//igraph:internal igraph_attribute_record_list_init
func newAttributeRecordList(size int) (*attributeRecordList, error) {
	return newAttributeRecordListWithInitializer(size, initializeAttributeRecordList)
}

func newAttributeRecordListWithInitializer(
	size int,
	initialize attributeRecordListInitializer,
) (*attributeRecordList, error) {
	if size < 0 {
		return nil, fmt.Errorf("igraph: attribute record list length must be non-negative: %d", size)
	}
	if _, err := intToIgraphInt(size, "attribute record list length"); err != nil {
		return nil, err
	}
	if initialize == nil {
		return nil, fmt.Errorf("igraph: attribute record list initializer is nil")
	}

	list := &attributeRecordList{}
	if code := initialize(list, size); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize attribute record list", code)
	}
	list.initialized = true
	return list, nil
}

func initializeAttributeRecordList(list *attributeRecordList, size int) int {
	return int(C.go_igraph_attribute_record_list_init(&list.value, C.igraph_int_t(size)))
}

//igraph:internal igraph_attribute_record_list_size
func (list *attributeRecordList) size() (int, error) {
	if list == nil || !list.initialized {
		return 0, ErrClosed
	}
	return igraphIntToInt(
		C.igraph_attribute_record_list_size(&list.value),
		"attribute record list length",
	)
}

//igraph:internal igraph_attribute_record_list_destroy
func (list *attributeRecordList) close() {
	if list == nil || !list.initialized {
		return
	}
	C.igraph_attribute_record_list_destroy(&list.value)
	list.initialized = false
}
