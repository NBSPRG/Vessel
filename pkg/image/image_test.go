package image

import (
	"testing"
)

// TestImageConstants ensures constants are defined correctly
func TestImageConstants(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		isEmpty bool
	}{
		{"RepoFile", RepoFile, RepoFile == ""},
		{"LyrDir", LyrDir, LyrDir == ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isEmpty {
				t.Errorf("%s constant should not be empty", tt.name)
			}
		})
	}
}

// TestImageStructure ensures Image structure is properly defined
func TestImageStructure(t *testing.T) {
	img := &Image{
		ID:         "test-id",
		Registry:   "docker.io",
		Repository: "library",
		Name:       "ubuntu",
		Tag:        "latest",
	}

	tests := []struct {
		field    string
		value    string
		expected string
	}{
		{"ID", img.ID, "test-id"},
		{"Registry", img.Registry, "docker.io"},
		{"Repository", img.Repository, "library"},
		{"Name", img.Name, "ubuntu"},
		{"Tag", img.Tag, "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s should be %s, got %s", tt.field, tt.expected, tt.value)
			}
		})
	}
}

// TestImageDefaults ensures Image fields default to empty strings
func TestImageDefaults(t *testing.T) {
	img := &Image{}

	if img.ID != "" {
		t.Error("Default ID should be empty")
	}

	if img.Registry != "" {
		t.Error("Default Registry should be empty")
	}

	if img.Repository != "" {
		t.Error("Default Repository should be empty")
	}

	if img.Name != "" {
		t.Error("Default Name should be empty")
	}

	if img.Tag != "" {
		t.Error("Default Tag should be empty")
	}
}
