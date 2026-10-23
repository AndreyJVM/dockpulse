package monitor

import (
	"testing"
)

func TestGetSwarmSummary(t *testing.T) {
	summary := GetSwarmSummary()
	// Should safely return summary without panic
	if summary.Text == "" {
		t.Errorf("expected non-empty Swarm summary text")
	}
}
