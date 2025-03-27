//go:build darwin

package freeport

import (
	"testing"
)

func TestGetEphemeralPortRange(t *testing.T) {
	minPort, maxPort, err := getEphemeralPortRange()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if minPort <= 0 || maxPort <= 0 || minPort > maxPort {
		t.Fatalf("unexpected values: min=%d, max=%d", minPort, maxPort)
	}
	t.Logf("min=%d, max=%d", minPort, maxPort)
}
