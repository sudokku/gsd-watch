package parser

import (
	"path/filepath"
	"testing"
)

func TestParsePlan(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantStatus  string
		wantWave    int
		wantTitle   string
		wantErr     bool
	}{
		{
			name:       "ValidFrontmatter",
			path:       filepath.Join("testdata", "valid-plan.md"),
			wantStatus: "in_progress",
			wantWave:   2,
			wantTitle:  "Build the widget parser",
			wantErr:    false,
		},
		{
			name:       "MinimalFrontmatter",
			path:       filepath.Join("testdata", "minimal-plan.md"),
			wantStatus: "",
			wantWave:   0,
			wantTitle:  "",
			wantErr:    false,
		},
		{
			name:    "MalformedYAML",
			path:    filepath.Join("testdata", "malformed-plan.md"),
			wantErr: true,
		},
		{
			name:    "MissingFile",
			path:    filepath.Join("testdata", "nonexistent-plan.md"),
			wantErr: true,
		},
		{
			name:       "TitleFromObjective",
			path:       filepath.Join("testdata", "valid-plan.md"),
			wantStatus: "in_progress",
			wantWave:   2,
			wantTitle:  "Build the widget parser",
			wantErr:    false,
		},
		{
			name:      "TitleFallback",
			path:      filepath.Join("testdata", "minimal-plan.md"),
			wantTitle: "",
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := parsePlan(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parsePlan(%q) expected error, got nil", tc.path)
				}
				// Ensure no panic — returning zero Plan is fine
				return
			}
			if err != nil {
				t.Fatalf("parsePlan(%q) unexpected error: %v", tc.path, err)
			}
			if plan.Status != tc.wantStatus {
				t.Errorf("Status: got %q, want %q", plan.Status, tc.wantStatus)
			}
			if plan.Wave != tc.wantWave {
				t.Errorf("Wave: got %d, want %d", plan.Wave, tc.wantWave)
			}
			if plan.Title != tc.wantTitle {
				t.Errorf("Title: got %q, want %q", plan.Title, tc.wantTitle)
			}
		})
	}
}
