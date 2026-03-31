package filesystem

import (
	"testing"
)

// TestMountOption ensures MountOption structure is properly defined
func TestMountOption(t *testing.T) {
	opt := MountOption{
		Source: "/source",
		Target: "/target",
		Type:   "tmpfs",
		Flag:   0,
		Option: "size=512m",
	}

	if opt.Source != "/source" {
		t.Errorf("Source should be /source, got %s", opt.Source)
	}

	if opt.Target != "/target" {
		t.Errorf("Target should be /target, got %s", opt.Target)
	}

	if opt.Type != "tmpfs" {
		t.Errorf("Type should be tmpfs, got %s", opt.Type)
	}

	if opt.Option != "size=512m" {
		t.Errorf("Option should be size=512m, got %s", opt.Option)
	}
}

// TestMountOptionDefaults ensures default values work
func TestMountOptionDefaults(t *testing.T) {
	opt := MountOption{}

	if opt.Source != "" {
		t.Error("Default Source should be empty")
	}

	if opt.Target != "" {
		t.Error("Default Target should be empty")
	}

	if opt.Type != "" {
		t.Error("Default Type should be empty")
	}

	if opt.Flag != 0 {
		t.Errorf("Default Flag should be 0, got %d", opt.Flag)
	}

	if opt.Option != "" {
		t.Error("Default Option should be empty")
	}
}

// TestUnmounter ensures Unmounter function type is callable
func TestUnmounter(t *testing.T) {
	var unmounter Unmounter
	unmounter = func() error {
		return nil
	}

	if unmounter == nil {
		t.Error("Unmounter should be assignable")
	}

	err := unmounter()
	if err != nil {
		t.Errorf("Unmounter should succeed, got error: %v", err)
	}
}
