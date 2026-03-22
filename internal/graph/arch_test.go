package graph

import (
	"os"
	"testing"
)

func TestVerifyAction(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fenrir-test-*")
	defer os.RemoveAll(tmpDir)

	g, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	g.Init()

	g.SaveDecision("Global State", "We should avoid global variables", "critical", "global")

	// Test conflict
	status, msg, _ := g.VerifyAction("Add a Global State variable", "")
	if status != "warning" {
		t.Errorf("Expected warning, got %s", status)
	}
	if msg == "" {
		t.Error("Expected conflict message")
	}

	// Test allowed
	status, msg, _ = g.VerifyAction("Add a refactor to the internal/graph module", "")
	if status != "success" {
		t.Errorf("Expected success, got %s", status)
	}
}
