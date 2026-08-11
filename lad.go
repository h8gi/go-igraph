package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "isomorphism_cgo.h"
*/
import "C"

import "fmt"

// LADOptions configures LAD subgraph matching. Domains, when non-nil, is
// indexed by pattern vertex and lists the permitted target vertex IDs. Domains
// are borrowed and copied for the synchronous call.
type LADOptions struct {
	Induced bool
	Domains [][]int
}

// LADEnumerationOptions configures bounded LAD mapping enumeration.
// MaxMappings must be positive.
type LADEnumerationOptions struct {
	LADOptions
	MaxMappings int
}

// ContainsSubgraphIsomorphicToLAD finds the first LAD mapping of pattern into
// g (the target). LAD supports directed and undirected graphs with matching
// directedness and allows loops, but does not support parallel edges. Returned
// mappings are non-nil and Go-owned.
//
//igraph:bind igraph_subisomorphic_lad
func (g *Graph) ContainsSubgraphIsomorphicToLAD(pattern *Graph, options LADOptions) (SubgraphIsomorphismResult, error) {
	result := emptySubgraphIsomorphismResult()
	err := withLockedGraphs([]*Graph{g, pattern}, func(graphs []*C.igraph_t) error {
		domains, impossible, err := validateLADDomains(graphs[1], graphs[0], options.Domains)
		if err != nil {
			return err
		}
		if err := validateLADGraphs(graphs[1], graphs[0]); err != nil {
			return err
		}
		if impossible {
			return nil
		}
		found, mapping, err := ladFirstMapping(graphs[1], graphs[0], domains, options.Induced)
		if err != nil || !found {
			return err
		}
		result, err = subgraphResultFromPatternMapping(mapping, int(C.igraph_vcount(graphs[0])))
		return err
	})
	return result, err
}

// EnumerateSubgraphIsomorphismsLAD returns at most MaxMappings
// pattern-to-target mappings. It partitions the remaining domain search space
// after each match, so the upstream unbounded map-list output is never used.
// Returned outer and nested slices are independently Go-owned.
func (g *Graph) EnumerateSubgraphIsomorphismsLAD(pattern *Graph, options LADEnumerationOptions) (MappingEnumerationResult, error) {
	result := MappingEnumerationResult{Mappings: make([][]int, 0)}
	if options.MaxMappings <= 0 {
		return result, fmt.Errorf("igraph: MaxMappings must be positive: %d", options.MaxMappings)
	}
	err := withLockedGraphs([]*Graph{g, pattern}, func(graphs []*C.igraph_t) error {
		domains, impossible, err := validateLADDomains(graphs[1], graphs[0], options.Domains)
		if err != nil {
			return err
		}
		if err := validateLADGraphs(graphs[1], graphs[0]); err != nil {
			return err
		}
		if impossible {
			return nil
		}
		if domains == nil {
			domains = unrestrictedLADDomains(int(C.igraph_vcount(graphs[1])), int(C.igraph_vcount(graphs[0])))
		}
		pending := [][][]int{domains}
		for len(pending) > 0 {
			last := len(pending) - 1
			searchDomains := pending[last]
			pending = pending[:last]
			found, mapping, err := ladFirstMapping(graphs[1], graphs[0], searchDomains, options.Induced)
			if err != nil {
				return err
			}
			if !found {
				continue
			}
			if len(result.Mappings) == options.MaxMappings {
				result.Truncated = true
				return nil
			}
			result.Mappings = append(result.Mappings, append([]int{}, mapping...))
			pending = append(pending, partitionLADDomains(searchDomains, mapping)...)
		}
		return nil
	})
	return result, err
}

func emptySubgraphIsomorphismResult() SubgraphIsomorphismResult {
	return SubgraphIsomorphismResult{
		PatternToTarget: make([]int, 0),
		TargetToPattern: make([]int, 0),
	}
}

func subgraphResultFromPatternMapping(mapping []int, targetCount int) (SubgraphIsomorphismResult, error) {
	result := SubgraphIsomorphismResult{
		Found:           true,
		PatternToTarget: append([]int{}, mapping...),
		TargetToPattern: make([]int, targetCount),
	}
	for index := range result.TargetToPattern {
		result.TargetToPattern[index] = RemovedID
	}
	for patternID, targetID := range mapping {
		if targetID < 0 || targetID >= targetCount {
			return emptySubgraphIsomorphismResult(), fmt.Errorf("igraph: LAD target mapping ID %d out of range [0, %d)", targetID, targetCount)
		}
		if result.TargetToPattern[targetID] != RemovedID {
			return emptySubgraphIsomorphismResult(), fmt.Errorf("igraph: LAD mapping repeats target vertex %d", targetID)
		}
		result.TargetToPattern[targetID] = patternID
	}
	return result, nil
}

func validateLADGraphs(pattern, target *C.igraph_t) error {
	if C.igraph_is_directed(pattern) != C.igraph_is_directed(target) {
		return fmt.Errorf("igraph: LAD pattern and target directedness must match")
	}
	patternMultiple, err := operatorGraphHasMultiple(pattern)
	if err != nil {
		return err
	}
	targetMultiple, err := operatorGraphHasMultiple(target)
	if err != nil {
		return err
	}
	if patternMultiple || targetMultiple {
		return fmt.Errorf("igraph: LAD does not support parallel edges")
	}
	return nil
}

func validateLADDomains(pattern, target *C.igraph_t, domains [][]int) ([][]int, bool, error) {
	if domains == nil {
		return nil, false, nil
	}
	patternCount := int(C.igraph_vcount(pattern))
	targetCount := int(C.igraph_vcount(target))
	if len(domains) != patternCount {
		return nil, false, fmt.Errorf("igraph: LAD domain count %d must match pattern vertex count %d", len(domains), patternCount)
	}
	copyDomains := make([][]int, len(domains))
	impossible := false
	for patternID, domain := range domains {
		copyDomains[patternID] = make([]int, len(domain))
		seen := make(map[int]struct{}, len(domain))
		for index, targetID := range domain {
			if targetID < 0 || targetID >= targetCount {
				return nil, false, fmt.Errorf("igraph: LAD domain %d target ID %d out of range [0, %d)", patternID, targetID, targetCount)
			}
			if _, exists := seen[targetID]; exists {
				return nil, false, fmt.Errorf("igraph: LAD domain %d repeats target ID %d", patternID, targetID)
			}
			if _, err := intToIgraphInt(targetID, fmt.Sprintf("LAD domain %d value %d", patternID, index)); err != nil {
				return nil, false, err
			}
			seen[targetID] = struct{}{}
			copyDomains[patternID][index] = targetID
		}
		if len(domain) == 0 {
			impossible = true
		}
	}
	return copyDomains, impossible, nil
}

func unrestrictedLADDomains(patternCount, targetCount int) [][]int {
	domains := make([][]int, patternCount)
	for patternID := range domains {
		domains[patternID] = make([]int, targetCount)
		for targetID := range domains[patternID] {
			domains[patternID][targetID] = targetID
		}
	}
	return domains
}

func partitionLADDomains(domains [][]int, mapping []int) [][][]int {
	partitions := make([][][]int, 0, len(mapping))
	for split := range mapping {
		partition := cloneLADDomains(domains)
		for prefix := 0; prefix < split; prefix++ {
			partition[prefix] = []int{mapping[prefix]}
		}
		partition[split] = removeLADDomainValue(partition[split], mapping[split])
		if len(partition[split]) != 0 {
			partitions = append(partitions, partition)
		}
	}
	return partitions
}

func cloneLADDomains(domains [][]int) [][]int {
	result := make([][]int, len(domains))
	for index := range domains {
		result[index] = append([]int(nil), domains[index]...)
	}
	return result
}

func removeLADDomainValue(domain []int, value int) []int {
	result := make([]int, 0, len(domain)-1)
	for _, candidate := range domain {
		if candidate != value {
			result = append(result, candidate)
		}
	}
	return result
}

func ladFirstMapping(pattern, target *C.igraph_t, domains [][]int, induced bool) (bool, []int, error) {
	var domainList *intVectorList
	var err error
	if domains != nil {
		domainList, err = newLADDomainList(domains)
		if err != nil {
			return false, nil, err
		}
		defer domainList.close()
	}
	mapping, err := newIntVector(nil)
	if err != nil {
		return false, nil, err
	}
	defer mapping.close()
	var domainPointer *C.igraph_vector_int_list_t
	if domainList != nil {
		domainPointer = &domainList.value
	}
	var found C.igraph_bool_t
	code := C.go_igraph_subisomorphic_lad(pattern, target, domainPointer, &found, &mapping.value, booltoint(induced))
	if code != C.IGRAPH_SUCCESS {
		return false, nil, igraphError("run LAD subgraph isomorphism", int(code))
	}
	values, err := mapping.slice()
	if err != nil {
		return false, nil, err
	}
	if found == booltoint(false) {
		return false, make([]int, 0), nil
	}
	return true, values, nil
}

//igraph:internal igraph_vector_int_list_push_back_copy
func newLADDomainList(domains [][]int) (*intVectorList, error) {
	list, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	for _, domain := range domains {
		vector, err := newIntVector(domain)
		if err != nil {
			list.close()
			return nil, err
		}
		code := C.go_igraph_vector_int_list_push_back_copy(&list.value, &vector.value)
		vector.close()
		if code != C.IGRAPH_SUCCESS {
			list.close()
			return nil, igraphError("copy LAD domain", int(code))
		}
	}
	return list, nil
}
