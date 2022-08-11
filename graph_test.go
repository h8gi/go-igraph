package igraph

import (
	"os"
	"testing"
)

func TestGraphWriteGraphML(t *testing.T) {
	dimvector := NewVector(2)
	dimvector.Set(0, 30)
	dimvector.Set(1, 30)
	g := NewLattice(dimvector, 0, false, false, true)

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

func TestVectorFromSlice(t *testing.T) {
	v := NewVectorFromSlice([]float64{1, 2, 3, 4})
	v.Get(1)
}

func TestVectorView(t *testing.T) {
	a := []float64{30, 30}
	v := VectorView(a)
	NewLattice(v, 0, true, false, true)
}
