package parser

import (
	"encoding/json"
	"os"
)

// configData holds the fields we need from config.json.
type configData struct {
	ModelProfile string `json:"model_profile"`
	Mode         string `json:"mode"`
}

// parseConfig reads config.json and extracts model_profile and mode.
// Returns zero-value configData on any error.
func parseConfig(path string) (configData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return configData{}, err
	}
	var cfg configData
	if err := json.Unmarshal(content, &cfg); err != nil {
		return configData{}, err
	}
	return cfg, nil
}
