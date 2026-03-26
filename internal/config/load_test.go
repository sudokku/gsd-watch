package config

import (
	"errors"
	"testing"
)

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
			wantConfig: Config{Emoji: true, Theme: ""},
			wantErr:    false,
		},
		{
			name:       "valid_config",
			path:       "testdata/valid.toml",
			wantConfig: Config{Emoji: false, Theme: "minimal"},
			wantErr:    false,
		},
		{
			name:       "malformed_toml",
			path:       "testdata/malformed.toml",
			wantConfig: Config{Emoji: true, Theme: ""},
			wantErr:    true,
		},
		{
			name:        "unknown_keys",
			path:        "testdata/unknown-keys.toml",
			wantConfig:  Config{Emoji: false, Theme: "default"},
			wantErr:     true,
			wantUnknown: true,
			unknownKeys: []string{"color"},
		},
		{
			name:       "empty_file",
			path:       "testdata/empty.toml",
			wantConfig: Config{Emoji: true, Theme: ""},
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
			if got.Theme != tt.wantConfig.Theme {
				t.Errorf("Load(%q) Config.Theme = %q, want %q", tt.path, got.Theme, tt.wantConfig.Theme)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	d := Defaults()
	if !d.Emoji {
		t.Errorf("Defaults().Emoji = false, want true")
	}
	if d.Theme != "" {
		t.Errorf("Defaults().Theme = %q, want empty string", d.Theme)
	}
}
