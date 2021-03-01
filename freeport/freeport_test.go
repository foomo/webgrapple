package freeport

import (
	"fmt"
	"sync"
	"testing"
)

// reset will reverse the setup from initialize() and then redo it (for tests)
func reset() {
	mu.Lock()
	defer mu.Unlock()

	logf("INFO", "resetting the freeport package state")

	effectiveMaxBlocks = 0
	firstPort = 0
	if lockLn != nil {
		lockLn.Close()
		lockLn = nil
	}

	once = sync.Once{}

	freePorts = nil
	pendingPorts = nil
	total = 0
}

// peekFree returns the next port that will be returned by Take to aid in testing.
func peekFree() int {
	mu.Lock()
	defer mu.Unlock()
	return freePorts.Front().Value.(int)
}

// peekAllFree returns all free ports that could be returned by Take to aid in testing.
func peekAllFree() []int {
	mu.Lock()
	defer mu.Unlock()

	var out []int
	for elem := freePorts.Front(); elem != nil; elem = elem.Next() {
		port := elem.Value.(int)
		out = append(out, port)
	}

	return out
}

// stats returns diagnostic data to aid in testing
func stats() (numTotal, numPending, numFree int) {
	mu.Lock()
	defer mu.Unlock()
	return total, pendingPorts.Len(), freePorts.Len()
}

func TestIntervalOverlap(t *testing.T) {
	cases := []struct {
		min1, max1, min2, max2 int
		overlap                bool
	}{
		{0, 0, 0, 0, true},
		{1, 1, 1, 1, true},
		{1, 3, 1, 3, true},  // same
		{1, 3, 4, 6, false}, // serial
		{1, 4, 3, 6, true},  // inner overlap
		{1, 6, 3, 4, true},  // nest
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d:%d vs %d:%d", tc.min1, tc.max1, tc.min2, tc.max2), func(t *testing.T) {
			if tc.overlap != intervalOverlap(tc.min1, tc.max1, tc.min2, tc.max2) { // 1 vs 2
				t.Fatalf("expected %v but got %v", tc.overlap, !tc.overlap)
			}
			if tc.overlap != intervalOverlap(tc.min2, tc.max2, tc.min1, tc.max1) { // 2 vs 1
				t.Fatalf("expected %v but got %v", tc.overlap, !tc.overlap)
			}
		})
	}
}
