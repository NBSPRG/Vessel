package container

import (
	"errors"
	"syscall"
	"testing"
)

// TestNewContainer ensures a new container is initialized properly
func TestNewContainer(t *testing.T) {
	ctr := NewContainer()

	if ctr == nil {
		t.Fatal("NewContainer returned nil")
	}

	if ctr.Config == nil {
		t.Error("Config should be initialized")
	}

	if ctr.Digest == "" {
		t.Error("Digest should be generated")
	}

	if len(ctr.Digest) != DigestStdLen {
		t.Errorf("Digest length should be %d, got %d", DigestStdLen, len(ctr.Digest))
	}

	if ctr.tier != -1 {
		t.Errorf("Initial tier should be -1, got %d", ctr.tier)
	}
}

// TestSetHostname ensures hostname is set correctly
func TestSetHostname(t *testing.T) {
	ctr := NewContainer()
	if err := ctr.SetHostname(); err != nil && !errors.Is(err, syscall.EPERM) {
		t.Fatalf("SetHostname returned error: %v", err)
	}

	if ctr.Config.Hostname == "" {
		t.Error("Hostname should not be empty after SetHostname")
	}

	// Hostname should be digest[:12] when initially empty
	expectedHostname := ctr.Digest[:12]
	if ctr.Config.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s, got %s", expectedHostname, ctr.Config.Hostname)
	}
}

// TestContainerFields ensures Container fields are properly initialized
func TestContainerFields(t *testing.T) {
	ctr := NewContainer()

	// Verify RootFS is empty initially
	if ctr.RootFS != "" {
		t.Errorf("RootFS should be empty initially, got %s", ctr.RootFS)
	}

	// Verify Pids is initialized as a slice
	if ctr.Pids != nil && len(ctr.Pids) != 0 {
		t.Errorf("Pids should be nil or empty, got %v", ctr.Pids)
	}

	// Verify memory tier values
	if ctr.mem != 0 {
		t.Errorf("mem should be 0 initially, got %d", ctr.mem)
	}

	if ctr.swap != 0 {
		t.Errorf("swap should be 0 initially, got %d", ctr.swap)
	}

	if ctr.cpus != 0 {
		t.Errorf("cpus should be 0 initially, got %f", ctr.cpus)
	}
}
