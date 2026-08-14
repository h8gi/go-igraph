package igraph

// #cgo pkg-config: igraph
// #include <stdlib.h>
// #include <igraph.h>
// #include "attributes_cgo.h"
import "C"

import (
	"fmt"
	"sort"
	"unsafe"
)

// AttributeCombination selects how values are combined when several source
// elements become one result element. The zero value drops the attribute.
// Operations unsupported for an attribute's type are rejected before the
// graph is changed.
type AttributeCombination uint8

const (
	AttributeCombineDrop AttributeCombination = iota
	AttributeCombineFirst
	AttributeCombineLast
	AttributeCombineSum
	AttributeCombineProduct
	AttributeCombineMinimum
	AttributeCombineMaximum
	AttributeCombineMean
	AttributeCombineConcat
)

// AttributeCombinationPolicy supplies a default rule and optional per-name
// overrides. The map and its strings are borrowed only for the call. A nil
// policy means that no merge policy was supplied; operations that would merge
// attributed elements reject it. The zero policy explicitly drops attributes.
type AttributeCombinationPolicy struct {
	Default   AttributeCombination
	Overrides map[string]AttributeCombination
}

func (combination AttributeCombination) cValue() (C.igraph_attribute_combination_type_t, error) {
	switch combination {
	case AttributeCombineDrop:
		return C.IGRAPH_ATTRIBUTE_COMBINE_IGNORE, nil
	case AttributeCombineFirst:
		return C.IGRAPH_ATTRIBUTE_COMBINE_FIRST, nil
	case AttributeCombineLast:
		return C.IGRAPH_ATTRIBUTE_COMBINE_LAST, nil
	case AttributeCombineSum:
		return C.IGRAPH_ATTRIBUTE_COMBINE_SUM, nil
	case AttributeCombineProduct:
		return C.IGRAPH_ATTRIBUTE_COMBINE_PROD, nil
	case AttributeCombineMinimum:
		return C.IGRAPH_ATTRIBUTE_COMBINE_MIN, nil
	case AttributeCombineMaximum:
		return C.IGRAPH_ATTRIBUTE_COMBINE_MAX, nil
	case AttributeCombineMean:
		return C.IGRAPH_ATTRIBUTE_COMBINE_MEAN, nil
	case AttributeCombineConcat:
		return C.IGRAPH_ATTRIBUTE_COMBINE_CONCAT, nil
	default:
		return 0, fmt.Errorf("igraph: invalid attribute combination: %d", combination)
	}
}

func validateCombinationType(combination AttributeCombination, attributeType AttributeType) error {
	if _, err := combination.cValue(); err != nil {
		return err
	}
	switch combination {
	case AttributeCombineDrop, AttributeCombineFirst, AttributeCombineLast:
		return nil
	case AttributeCombineSum, AttributeCombineProduct, AttributeCombineMinimum,
		AttributeCombineMaximum, AttributeCombineMean:
		if attributeType == AttributeNumeric {
			return nil
		}
	case AttributeCombineConcat:
		if attributeType == AttributeString {
			return nil
		}
	}
	return fmt.Errorf("igraph: attribute combination %d is not valid for attribute type %d", combination, attributeType)
}

func validateCombinationPolicy(policy *AttributeCombinationPolicy, metadata []AttributeMetadata) error {
	if policy == nil {
		if len(metadata) == 0 {
			return nil
		}
		return fmt.Errorf("igraph: an explicit attribute combination policy is required")
	}
	types := make(map[string]AttributeType, len(metadata))
	for _, attribute := range metadata {
		types[attribute.Name] = attribute.Type
		combination := policy.Default
		if override, ok := policy.Overrides[attribute.Name]; ok {
			combination = override
		}
		if err := validateCombinationType(combination, attribute.Type); err != nil {
			return fmt.Errorf("igraph: attribute %q: %w", attribute.Name, err)
		}
	}
	for name := range policy.Overrides {
		if err := validateAttributeName(name); err != nil {
			return err
		}
		_, ok := types[name]
		if !ok {
			return fmt.Errorf("%w: attribute combination override %q", ErrAttributeNotFound, name)
		}
	}
	_, err := policy.Default.cValue()
	return err
}

type attributeCombination struct {
	value       C.igraph_attribute_combination_t
	initialized bool
	destroy     func()
}

type attributeCombinationHooks struct {
	init    func() error
	add     func(string, AttributeCombination) error
	destroy func()
}

//igraph:internal igraph_attribute_combination_init
//igraph:internal igraph_attribute_combination_add
//igraph:internal igraph_attribute_combination_destroy
func newAttributeCombination(policy *AttributeCombinationPolicy, metadata []AttributeMetadata) (*attributeCombination, error) {
	return newAttributeCombinationWithHooks(policy, metadata, attributeCombinationHooks{})
}

func newAttributeCombinationWithHooks(policy *AttributeCombinationPolicy, metadata []AttributeMetadata, hooks attributeCombinationHooks) (*attributeCombination, error) {
	if err := validateCombinationPolicy(policy, metadata); err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	combination := &attributeCombination{}
	if hooks.init != nil {
		if err := hooks.init(); err != nil {
			return nil, err
		}
		combination.initialized = true
		combination.destroy = hooks.destroy
	} else {
		if code := C.go_igraph_attribute_combination_init(&combination.value); code != C.IGRAPH_SUCCESS {
			return nil, igraphError("initialize attribute combination", int(code))
		}
		combination.initialized = true
	}
	failed := true
	defer func() {
		if failed {
			combination.close()
		}
	}()

	names := make([]string, 0, len(policy.Overrides))
	for name := range policy.Overrides {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := combination.add(name, policy.Overrides[name], hooks.add); err != nil {
			return nil, err
		}
	}
	if err := combination.addDefault(policy.Default, hooks.add); err != nil {
		return nil, err
	}
	failed = false
	return combination, nil
}

func (combination *attributeCombination) add(name string, rule AttributeCombination, hook func(string, AttributeCombination) error) error {
	if hook != nil {
		return hook(name, rule)
	}
	cRule, _ := rule.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_attribute_combination_add(&combination.value, cName, cRule); code != C.IGRAPH_SUCCESS {
		return igraphError("add named attribute combination", int(code))
	}
	return nil
}

func (combination *attributeCombination) addDefault(rule AttributeCombination, hook func(string, AttributeCombination) error) error {
	if hook != nil {
		return hook("", rule)
	}
	cRule, _ := rule.cValue()
	if code := C.go_igraph_attribute_combination_add(&combination.value, nil, cRule); code != C.IGRAPH_SUCCESS {
		return igraphError("add default attribute combination", int(code))
	}
	return nil
}

func (combination *attributeCombination) close() {
	if combination != nil && combination.initialized {
		if combination.destroy != nil {
			combination.destroy()
		} else {
			C.igraph_attribute_combination_destroy(&combination.value)
		}
		combination.initialized = false
	}
}
