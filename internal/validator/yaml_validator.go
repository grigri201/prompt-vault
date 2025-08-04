package validator

// YAMLValidator defines the interface for YAML validation operations
type YAMLValidator interface {
	// ValidatePromptFile validates a YAML prompt file and returns parsed content
	ValidatePromptFile(content []byte) (*PromptFileContent, error)

	// ValidateRequired validates that required fields are present and valid
	ValidateRequired(prompt *PromptFileContent) error
}

// PromptFileContent represents the structure of a prompt YAML file
type PromptFileContent struct {
	// Metadata contains the YAML front matter metadata
	Metadata PromptMetadata `yaml:",inline"`
	
	// Content contains the actual prompt content after the --- separator
	Content string `yaml:"-"`
}

// PromptMetadata represents the metadata fields in a prompt YAML file
type PromptMetadata struct {
	// Name is the prompt name (required)
	Name string `yaml:"name"`
	
	// Author is the author information (required)
	Author string `yaml:"author"`
	
	// Description is an optional description of the prompt
	Description string `yaml:"description,omitempty"`
	
	// Tags is an optional list of tags for categorization
	Tags []string `yaml:"tags,omitempty"`
	
	// Version is an optional version string
	Version string `yaml:"version,omitempty"`
}