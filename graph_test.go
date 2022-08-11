package igraph

import (
	"os"
	"testing"
)

func TestGraphWriteGraphML(t *testing.T) {
	dimvector := NewVector(2)
	dimvector.Set(0, 30)
	dimvector.Set(1, 30)
	g := NewLattice(*dimvector, 0, false, false, true)

	f, err := os.Create("./testdata/test.graphml")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer f.Close()
	if err := g.WriteGraphML(f, false); err != nil {
		t.Log(err)
		t.Fail()
	}
}
