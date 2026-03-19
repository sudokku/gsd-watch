package parser

import (
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		wantModelProfile string
		wantMode         string
		wantErr          bool
	}{
		{
			name:             "Valid",
			path:             filepath.Join("testdata", "config.json"),
			wantModelProfile: "balanced",
			wantMode:         "yolo",
			wantErr:          false,
		},
		{
			name:    "MissingFile",
			path:    filepath.Join("testdata", "nonexistent-config.json"),
			wantErr: true,
		},
		{
			name:    "BadJSON",
			path:    filepath.Join("testdata", "bad-config.json"),
			wantErr: true,
		},
		{
			name:             "MissingFields",
			path:             filepath.Join("testdata", "config-missing-fields.json"),
			wantModelProfile: "",
			wantMode:         "",
			wantErr:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := parseConfig(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseConfig(%q) expected error, got nil", tc.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseConfig(%q) unexpected error: %v", tc.path, err)
			}
			if cfg.ModelProfile != tc.wantModelProfile {
				t.Errorf("ModelProfile: got %q, want %q", cfg.ModelProfile, tc.wantModelProfile)
			}
			if cfg.Mode != tc.wantMode {
				t.Errorf("Mode: got %q, want %q", cfg.Mode, tc.wantMode)
			}
		})
	}
}
