package config

// Store defines the interface for configuration storage
type Store interface {
	// SaveToken saves the GitHub authentication token
	SaveToken(token string) error

	// GetToken retrieves the stored GitHub authentication token
	GetToken() (string, error)

	// DeleteToken removes the stored token
	DeleteToken() error

	// GetConfigPath returns the configuration file path
	GetConfigPath() string
}

// Config represents the application configuration
type Config struct {
	// GitHubToken is the encrypted/obfuscated GitHub personal access token
	GitHubToken string `json:"github_token,omitempty"`
}
