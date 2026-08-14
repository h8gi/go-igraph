package igraph_test

import (
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestSpatialPublicVocabulary(t *testing.T) {
	maximum := 3
	cutoff := 2.5
	options := igraph.NearestNeighborOptions{
		Metric:       igraph.SpatialManhattan,
		MaxNeighbors: &maximum,
		Cutoff:       &cutoff,
		Directed:     true,
	}
	if options.Metric != igraph.SpatialManhattan || options.MaxNeighbors == nil || options.Cutoff == nil || !options.Directed {
		t.Fatalf("spatial options changed unexpectedly: %#v", options)
	}
	if igraph.SpatialEuclidean != 0 {
		t.Fatalf("zero spatial metric = %d, want SpatialEuclidean", igraph.SpatialEuclidean)
	}
}
