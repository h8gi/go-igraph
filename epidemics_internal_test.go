package igraph

import (
	"errors"
	"sync"
	"testing"
)

func TestSIRFailureCleanup(t *testing.T) {
	forced := errors.New("forced")
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	base := defaultSIRAdapters()
	failedInit := base
	failedInit.newResults = func() (*sirResults, error) { return nil, forced }
	if _, err := g.sir(1, 1, 1, SIROptions{}, &failedInit); !errors.Is(err, forced) {
		t.Fatalf("init error = %v", err)
	}
	closed := 0
	upstream := base
	upstream.close = func(*sirResults) { closed++ }
	upstream.run = func(*Graph, float64, float64, int, *sirResults) int { return 4 }
	if _, err := g.sir(1, 1, 1, SIROptions{}, &upstream); err == nil || closed != 1 {
		t.Fatalf("upstream = %v, closes=%d", err, closed)
	}
	closed = 0
	conversion := base
	conversion.close = func(*sirResults) { closed++ }
	conversion.run = func(*Graph, float64, float64, int, *sirResults) int { return 0 }
	conversion.extract = func(*sirResults) ([]SIRTrajectory, error) { return nil, forced }
	if _, err := g.sir(1, 1, 1, SIROptions{}, &conversion); !errors.Is(err, forced) || closed != 1 {
		t.Fatalf("conversion = %v, closes=%d", err, closed)
	}
}

func TestSIRCloseRaceWaitsForReadLock(t *testing.T) {
	g, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	base := defaultSIRAdapters()
	started, release := make(chan struct{}), make(chan struct{})
	base.run = func(*Graph, float64, float64, int, *sirResults) int { close(started); <-release; return 4 }
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); _, _ = g.sir(1, 1, 1, SIROptions{}, &base) }()
	<-started
	closed := make(chan struct{})
	go func() { _ = g.Close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("Close completed while SIR held the read lock")
	default:
	}
	close(release)
	wg.Wait()
	<-closed
}
