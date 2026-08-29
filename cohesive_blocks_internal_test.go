package igraph

import (
	"errors"
	"testing"
)

func TestCollectCohesiveBlocksFailuresAndCleanup(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name         string
		failInit     int
		queryInit    bool
		failQuery    bool
		failConvert  int
		failTreeInfo bool
		wantLists    int
		wantVectors  int
		wantTrees    int
	}{
		{name: "list initialization", failInit: 1},
		{name: "cohesion initialization", failInit: 2, wantLists: 1},
		{name: "parent initialization", failInit: 3, wantLists: 1, wantVectors: 1},
		{name: "upstream before graph", failQuery: true, wantLists: 1, wantVectors: 2},
		{name: "upstream after graph", failQuery: true, queryInit: true, wantLists: 1, wantVectors: 2, wantTrees: 1},
		{name: "blocks conversion", queryInit: true, failConvert: 1, wantLists: 1, wantVectors: 2, wantTrees: 1},
		{name: "cohesion conversion", queryInit: true, failConvert: 2, wantLists: 1, wantVectors: 2, wantTrees: 1},
		{name: "parent conversion", queryInit: true, failConvert: 3, wantLists: 1, wantVectors: 2, wantTrees: 1},
		{name: "tree inspection", queryInit: true, failTreeInfo: true, wantLists: 1, wantVectors: 2, wantTrees: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initCalls, convertCalls, closedLists, closedVectors, destroyedTrees := 0, 0, 0, 0, 0
			ops := cohesiveBlockOperations{
				newList: func() (*intVectorList, error) {
					initCalls++
					if initCalls == tt.failInit {
						return nil, forced
					}
					return &intVectorList{}, nil
				},
				newVector: func() (*intVector, error) {
					initCalls++
					if initCalls == tt.failInit {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeList:   func(*intVectorList) { closedLists++ },
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVectorList, *intVector, *intVector, *cohesiveBlockTree) (bool, error) {
					if tt.failQuery {
						return tt.queryInit, forced
					}
					return tt.queryInit, nil
				},
				listSlices: func(*intVectorList) ([][]int, error) {
					convertCalls++
					if convertCalls == tt.failConvert {
						return nil, forced
					}
					return [][]int{{}}, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					convertCalls++
					if convertCalls == tt.failConvert {
						return nil, forced
					}
					return []int{0}, nil
				},
				treeInfo: func(*cohesiveBlockTree) (cohesiveBlockTreeInfo, error) {
					if tt.failTreeInfo {
						return cohesiveBlockTreeInfo{}, forced
					}
					return cohesiveBlockTreeInfo{vertexCount: 1, directed: true, edges: []Edge{}}, nil
				},
				destroyTree: func(*cohesiveBlockTree) { destroyedTrees++ },
				adoptTree:   func(*cohesiveBlockTree) *Graph { return &Graph{} },
			}
			_, err := collectCohesiveBlocks(0, ops)
			if !errors.Is(err, forced) {
				t.Errorf("error = %v, want %v", err, forced)
			}
			if closedLists != tt.wantLists || closedVectors != tt.wantVectors || destroyedTrees != tt.wantTrees {
				t.Errorf("cleanup = %d/%d/%d, want %d/%d/%d", closedLists, closedVectors, destroyedTrees, tt.wantLists, tt.wantVectors, tt.wantTrees)
			}
		})
	}
}

func TestValidateCohesiveBlocks(t *testing.T) {
	valid := CohesiveBlocksResult{Blocks: [][]int{{0, 1}, {0}}, Cohesion: []int{1, 2}, Parents: []int{-1, 0}}
	tree := cohesiveBlockTreeInfo{vertexCount: 2, directed: true, edges: []Edge{{0, 1}}}
	if err := validateCohesiveBlocks(valid, 2, tree); err != nil {
		t.Fatal(err)
	}
	invalid := []CohesiveBlocksResult{
		{},
		{Blocks: [][]int{}, Cohesion: []int{}, Parents: []int{}},
		{Blocks: [][]int{{0}}, Cohesion: nil, Parents: []int{-1}},
		{Blocks: [][]int{nil}, Cohesion: []int{0}, Parents: []int{-1}},
		{Blocks: [][]int{{0, 0}}, Cohesion: []int{0}, Parents: []int{-1}},
		{Blocks: [][]int{{1}}, Cohesion: []int{0}, Parents: []int{-1}},
		{Blocks: [][]int{{0}}, Cohesion: []int{-1}, Parents: []int{-1}},
		{Blocks: [][]int{{0}, {0}}, Cohesion: []int{1, 1}, Parents: []int{-1, 0}},
		{Blocks: [][]int{{0}, {0}}, Cohesion: []int{1, 2}, Parents: []int{-1, 1}},
	}
	for _, result := range invalid {
		if err := validateCohesiveBlocks(result, 1, tree); err == nil {
			t.Errorf("valid result: %#v", result)
		}
	}
	badTrees := []cohesiveBlockTreeInfo{
		{vertexCount: 1, directed: true},
		{vertexCount: 2, directed: false, edges: []Edge{{0, 1}}},
		{vertexCount: 2, directed: true},
		{vertexCount: 2, directed: true, edges: []Edge{{1, 0}}},
	}
	for _, bad := range badTrees {
		if err := validateCohesiveBlocks(valid, 2, bad); err == nil {
			t.Errorf("valid tree: %#v", bad)
		}
	}
}
