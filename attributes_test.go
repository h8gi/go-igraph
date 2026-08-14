package igraph_test

import (
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestAttributePublicVocabulary(t *testing.T) {
	metadata := []igraph.AttributeMetadata{
		{Name: "dataset", Scope: igraph.AttributeGraph, Type: igraph.AttributeString},
		{Name: "active", Scope: igraph.AttributeVertex, Type: igraph.AttributeBoolean},
		{Name: "weight", Scope: igraph.AttributeEdge, Type: igraph.AttributeNumeric},
	}

	if metadata[0].Scope != igraph.AttributeGraph || metadata[0].Type != igraph.AttributeString {
		t.Fatalf("graph metadata = %#v", metadata[0])
	}
	if metadata[1].Scope != igraph.AttributeVertex || metadata[1].Type != igraph.AttributeBoolean {
		t.Fatalf("vertex metadata = %#v", metadata[1])
	}
	if metadata[2].Scope != igraph.AttributeEdge || metadata[2].Type != igraph.AttributeNumeric {
		t.Fatalf("edge metadata = %#v", metadata[2])
	}
}
