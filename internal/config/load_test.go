package config

import (
	"errors"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantConfig  Config
		wantErr     bool
		wantUnknown bool
		unknownKeys []string
	}{
		{
			name:       "missing_file",
			path:       "testdata/does-not-exist.toml",
			wantConfig: Config{Emoji: true, Preset: ""},
			wantErr:    false,
		},
		{
			name:       "valid_config",
			path:       "testdata/valid.toml",
			wantConfig: Config{Emoji: false, Preset: "minimal"},
			wantErr:    false,
		},
		{
			name:       "malformed_toml",
			path:       "testdata/malformed.toml",
			wantConfig: Config{Emoji: true, Preset: ""},
			wantErr:    true,
		},
		{
			name:        "unknown_keys",
			path:        "testdata/unknown-keys.toml",
			wantConfig:  Config{Emoji: false, Preset: "default"},
			wantErr:     true,
			wantUnknown: true,
			unknownKeys: []string{"color"},
		},
		{
			name:       "empty_file",
			path:       "testdata/empty.toml",
			wantConfig: Config{Emoji: true, Preset: ""},
			wantErr:    false,
		},
		{
			name: "theme_colors",
			path: "testdata/theme-colors.toml",
			wantConfig: Config{Emoji: true, Preset: "default", Colors: ThemeColors{
				Complete: strPtr("#00ff00"),
				Failed:   strPtr("#ff0000"),
			}},
			wantErr: false,
		},
		{
			name: "theme_colors_invalid_still_decodes",
			path: "testdata/theme-colors-invalid.toml",
			wantConfig: Config{Emoji: true, Preset: "default", Colors: ThemeColors{
				Complete: strPtr("not-a-hex"),
				Failed:   strPtr("#ff"),
			}},
			wantErr: false,
		},
		{
			// theme = "string" now maps to Config.Colors (ThemeColors table) — TOML
			// decoder returns a type mismatch error (string vs table), not an
			// UnknownKeysError. The old key still produces a fatal load error,
			// which is the intended behaviour: users are warned via cfg-02 stderr path.
			name:       "old_theme_key",
			path:       "testdata/old-theme-key.toml",
			wantConfig: Config{Emoji: true, Preset: ""},
			wantErr:    true,
		},
		{
			name:       "empty_theme_section",
			path:       "testdata/empty-theme-section.toml",
			wantConfig: Config{Emoji: true, Preset: "default"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}

			// For the malformed case: error must NOT be *UnknownKeysError
			if tt.wantErr && !tt.wantUnknown {
				var ukErr *UnknownKeysError
				if errors.As(err, &ukErr) {
					t.Errorf("Load(%q) got UnknownKeysError but expected a parse error", tt.path)
				}
			}

			// For the unknown_keys case: error must be *UnknownKeysError with expected keys
			if tt.wantUnknown {
				var ukErr *UnknownKeysError
				if !errors.As(err, &ukErr) {
					t.Errorf("Load(%q) expected *UnknownKeysError, got %T: %v", tt.path, err, err)
				} else {
					for _, key := range tt.unknownKeys {
						found := false
						for _, k := range ukErr.Keys {
							if k == key {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Load(%q) UnknownKeysError.Keys = %v, want key %q", tt.path, ukErr.Keys, key)
						}
					}
				}
			}

			if got.Emoji != tt.wantConfig.Emoji {
				t.Errorf("Load(%q) Config.Emoji = %v, want %v", tt.path, got.Emoji, tt.wantConfig.Emoji)
			}
			if got.Preset != tt.wantConfig.Preset {
				t.Errorf("Load(%q) Config.Preset = %q, want %q", tt.path, got.Preset, tt.wantConfig.Preset)
			}

			// Compare ThemeColors pointer fields
			checkStringPtr(t, tt.path, "Colors.Complete", got.Colors.Complete, tt.wantConfig.Colors.Complete)
			checkStringPtr(t, tt.path, "Colors.Active", got.Colors.Active, tt.wantConfig.Colors.Active)
			checkStringPtr(t, tt.path, "Colors.Pending", got.Colors.Pending, tt.wantConfig.Colors.Pending)
			checkStringPtr(t, tt.path, "Colors.Failed", got.Colors.Failed, tt.wantConfig.Colors.Failed)
			checkStringPtr(t, tt.path, "Colors.NowMarker", got.Colors.NowMarker, tt.wantConfig.Colors.NowMarker)
		})
	}
}

// checkStringPtr compares two *string values by field name, reporting mismatches.
func checkStringPtr(t *testing.T, path, field string, got, want *string) {
	t.Helper()
	if got != nil && want != nil {
		if *got != *want {
			t.Errorf("Load(%q) Config.%s = %q, want %q", path, field, *got, *want)
		}
	} else if (got == nil) != (want == nil) {
		t.Errorf("Load(%q) Config.%s nil=%v, want nil=%v", path, field, got == nil, want == nil)
	}
}

func TestDefaults(t *testing.T) {
	d := Defaults()
	if !d.Emoji {
		t.Errorf("Defaults().Emoji = false, want true")
	}
	if d.Preset != "" {
		t.Errorf("Defaults().Preset = %q, want empty string", d.Preset)
	}
}
