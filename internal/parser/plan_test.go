package parser

import (
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestExtractFrontmatter_BOM(t *testing.T) {
	// BOM prefix: \xEF\xBB\xBF — must be stripped before frontmatter check
	content := "\xEF\xBB\xBF---\nstatus: in_progress\nwave: 1\n---\nsome prose"
	fm, prose := extractFrontmatter(content)
	if fm == "" {
		t.Fatal("expected frontmatter to be extracted, got empty string")
	}
	var pf planFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &pf); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if pf.Status != "in_progress" {
		t.Errorf("Status: got %q, want %q", pf.Status, "in_progress")
	}
	if pf.Wave != 1 {
		t.Errorf("Wave: got %d, want 1", pf.Wave)
	}
	if prose == "" {
		t.Error("expected prose to be non-empty")
	}
}

func TestExtractFrontmatter_LeadingWhitespace(t *testing.T) {
	// Leading spaces and newline before "---" must be stripped
	content := "  \n\t---\nstatus: pending\n---\nbody"
	fm, prose := extractFrontmatter(content)
	if fm == "" {
		t.Fatal("expected frontmatter to be extracted with leading whitespace, got empty string")
	}
	var pf planFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &pf); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if pf.Status != "pending" {
		t.Errorf("Status: got %q, want %q", pf.Status, "pending")
	}
	if prose == "" {
		t.Error("expected prose to be non-empty")
	}
}

func TestExtractFrontmatter_BOMAndWhitespace(t *testing.T) {
	// Both BOM and leading whitespace
	content := "\xEF\xBB\xBF \n---\nstatus: complete\n---\ntext"
	fm, _ := extractFrontmatter(content)
	if fm == "" {
		t.Fatal("expected frontmatter to be extracted with BOM+whitespace, got empty string")
	}
	var pf planFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &pf); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if pf.Status != "complete" {
		t.Errorf("Status: got %q, want %q", pf.Status, "complete")
	}
}

func TestExtractFrontmatter_NoBOM(t *testing.T) {
	// Regression: standard frontmatter without BOM must still work
	content := "---\nstatus: ok\nwave: 3\n---\nbody text"
	fm, prose := extractFrontmatter(content)
	if fm == "" {
		t.Fatal("expected frontmatter with standard --- to be extracted")
	}
	var pf planFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &pf); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if pf.Status != "ok" {
		t.Errorf("Status: got %q, want %q", pf.Status, "ok")
	}
	if pf.Wave != 3 {
		t.Errorf("Wave: got %d, want 3", pf.Wave)
	}
	if prose == "" {
		t.Error("expected prose to be non-empty")
	}
}
