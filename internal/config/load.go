package config

import (
	"errors"
	"io/fs"

	"github.com/BurntSushi/toml"
)

// ThemeColors holds optional hex color overrides for the 5 status-tree colors.
// Pointer fields: nil = not set by user; non-nil = user provided a value.
type ThemeColors struct {
	Complete  *string `toml:"complete"`
	Active    *string `toml:"active"`
	Pending   *string `toml:"pending"`
	Failed    *string `toml:"failed"`
	NowMarker *string `toml:"now_marker"`
}

// Config holds user configuration loaded from TOML.
type Config struct {
	Emoji  bool        `toml:"emoji"`
	Preset string      `toml:"preset"`
	Colors ThemeColors `toml:"theme"`
}

// ConfigPath is the XDG-relative config file path (joined with os.UserHomeDir()).
const ConfigPath = ".config/gsd-watch/config.toml"

// UnknownKeysError is returned when the config file contains keys not mapped to Config fields.
type UnknownKeysError struct {
	Keys []string
}

func (e *UnknownKeysError) Error() string {
	return "unknown config keys"
}

// Defaults returns a Config with all fields at their documented defaults.
// Emoji is true (show emoji by default). Preset is "" (Phase 14 interprets as "default").
func Defaults() Config {
	return Config{Emoji: true, Preset: ""}
}

// Load reads the config file at path and returns the decoded Config.
// Returns:
//   - (Defaults(), nil) when file is missing — CFG-01
//   - (Defaults(), error) when file has malformed TOML — CFG-02
//   - (cfg, *UnknownKeysError) when file has unrecognised keys — CFG-03
//   - (cfg, nil) on success
func Load(path string) (Config, error) {
	cfg := Defaults()
	md, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Defaults(), nil
		}
		return Defaults(), err
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, k := range undecoded {
			keys[i] = k.String()
		}
		return cfg, &UnknownKeysError{Keys: keys}
	}
	return cfg, nil
}
