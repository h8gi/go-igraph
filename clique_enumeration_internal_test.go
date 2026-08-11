package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestCollectCliqueEnumerationFailureCleanup(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name       string
		newError   bool
		queryError bool
		sliceError bool
	}{
		{name: "initialization", newError: true},
		{name: "upstream query", queryError: true},
		{name: "partial nested conversion", sliceError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closed := 0
			result, err := collectCliqueEnumeration(
				VertexSetEnumerationOptions{MaxResults: 1},
				cliqueEnumerationOperations{
					newList: func() (*intVectorList, error) {
						if tt.newError {
							return nil, forced
						}
						return &intVectorList{}, nil
					},
					closeList: func(*intVectorList) { closed++ },
					query: func(*intVectorList) error {
						if tt.queryError {
							return forced
						}
						return nil
					},
					listSlices: func(*intVectorList) ([][]int, error) {
						if tt.sliceError {
							return nil, forced
						}
						return [][]int{{1, 0}, {2}}, nil
					},
				},
			)
			if !errors.Is(err, forced) || result.Sets == nil || len(result.Sets) != 0 {
				t.Errorf("result = %#v, %v", result, err)
			}
			wantClosed := 1
			if tt.newError {
				wantClosed = 0
			}
			if closed != wantClosed {
				t.Errorf("closed = %d; want %d", closed, wantClosed)
			}
		})
	}
}

func TestCollectCliqueEnumerationCanonicalizationAndExactTruncation(t *testing.T) {
	closed := 0
	result, err := collectCliqueEnumeration(
		VertexSetEnumerationOptions{MaxResults: 1},
		cliqueEnumerationOperations{
			newList:   func() (*intVectorList, error) { return &intVectorList{}, nil },
			closeList: func(*intVectorList) { closed++ },
			query:     func(*intVectorList) error { return nil },
			listSlices: func(*intVectorList) ([][]int, error) {
				return [][]int{{2, 0, 1}, {4, 3}}, nil
			},
		},
	)
	if err != nil || !result.Truncated || !reflect.DeepEqual(result.Sets, [][]int{{0, 1, 2}}) || closed != 1 {
		t.Errorf("result = %#v, err = %v, closed = %d", result, err, closed)
	}
}
