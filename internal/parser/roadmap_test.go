package parser

import (
	"path/filepath"
	"testing"
)

func TestParseRoadmap_Valid(t *testing.T) {
	result := parseRoadmap(filepath.Join("testdata", "roadmap.md"))

	expected := map[int]string{
		1: "Core TUI Scaffold",
		2: "Live Data Layer",
		3: "File Watching",
		4: "Plugin & Delivery",
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d phases, got %d: %v", len(expected), len(result), result)
	}

	for num, name := range expected {
		if result[num] != name {
			t.Errorf("phase %d: expected %q, got %q", num, name, result[num])
		}
	}
}

func TestParseRoadmap_MissingFile(t *testing.T) {
	result := parseRoadmap(filepath.Join("testdata", "nonexistent-roadmap.md"))

	if result == nil {
		t.Fatal("expected non-nil map for missing file, got nil")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map for missing file, got %v", result)
	}
}

func TestParseRoadmap_NoHeadings(t *testing.T) {
	result := parseRoadmap(filepath.Join("testdata", "roadmap-no-headings.md"))

	if result == nil {
		t.Fatal("expected non-nil map for file with no headings, got nil")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map for file with no headings, got %v", result)
	}
}

func TestParseRoadmap_H2H4Headings(t *testing.T) {
	result := parseRoadmap(filepath.Join("testdata", "roadmap-h2-h4.md"))

	expected := map[int]string{
		5: "H2 Phase",
		6: "H3 Phase",
		7: "H4 Phase",
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d phases, got %d: %v", len(expected), len(result), result)
	}

	for num, name := range expected {
		if result[num] != name {
			t.Errorf("phase %d: expected %q, got %q", num, name, result[num])
		}
	}
}
