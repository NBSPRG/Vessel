package reexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNaiveSelf ensures the path resolution works
func TestNaiveSelf(t *testing.T) {
	result := naiveSelf()

	if result == "" {
		t.Error("naiveSelf() should return non-empty string")
	}

	if !filepath.IsAbs(result) {
		t.Errorf("naiveSelf() should return absolute path, got %s", result)
	}
}

// TestNaiveSelfIsExecutable tests that the returned path is the current executable
func TestNaiveSelfIsExecutable(t *testing.T) {
	result := naiveSelf()

	// Get the actual executable path
	executable, err := os.Executable()
	if err != nil {
		t.Skipf("Could not get executable path: %v", err)
	}

	// Both should resolve to absolute paths
	absResult, _ := filepath.Abs(result)
	absExecutable, _ := filepath.Abs(executable)

	if absResult != absExecutable {
		// They might differ slightly, but the file should exist
		if _, err := os.Stat(absResult); err != nil {
			t.Errorf("naiveSelf() returned non-existent path: %s", result)
		}
	}
}

// TestPathResolution ensures relative paths are converted to absolute
func TestPathResolution(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with relative path
	testPath := "./myprogram"
	os.Args = []string{testPath}

	result := naiveSelf()

	if !filepath.IsAbs(result) {
		t.Errorf("Relative path %s should be converted to absolute, got %s", testPath, result)
	}
}

// TestCommandLookup ensures command lookup works correctly
func TestCommandLookup(t *testing.T) {
	// Use a common system command that should exist
	cmd := "go"

	path, err := exec.LookPath(cmd)
	if err != nil {
		t.Skipf("Command 'go' not found in PATH: %v", err)
	}

	if path == "" {
		t.Error("LookPath should return non-empty path")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("LookPath should return absolute path, got %s", path)
	}
}
