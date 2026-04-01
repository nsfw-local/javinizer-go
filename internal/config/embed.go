package config

import (
	_ "embed"
)

//go:embed config.yaml.example
var embeddedConfig []byte

// GetEmbeddedConfig returns the embedded config.yaml.example content as a string.
// This is used when generating new config files to ensure Docker and non-Docker
// deployments produce identical commented configurations.
func GetEmbeddedConfig() string {
	return string(embeddedConfig)
}

// EmbeddedConfigBytes returns the raw embedded config bytes.
// Use this when you need the byte slice directly (e.g., for YAML parsing).
func EmbeddedConfigBytes() []byte {
	// Return a copy to prevent mutation of the embedded data
	result := make([]byte, len(embeddedConfig))
	copy(result, embeddedConfig)
	return result
}
