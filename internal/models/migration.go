package models

// LegacyPromptMeta represents the old structure with Category field
type LegacyPromptMeta struct {
	Name        string   `yaml:"name"`
	Author      string   `yaml:"author"`
	Category    string   `yaml:"category,omitempty"`
	Tags        []string `yaml:"tags"`
	Version     string   `yaml:"version,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Parent      string   `yaml:"parent,omitempty"`
	ID          string   `yaml:"id,omitempty"`
}

// MigrateLegacyPrompt converts old format to new format
func MigrateLegacyPrompt(legacy LegacyPromptMeta) PromptMeta {
	// Convert category to tag if present
	tags := legacy.Tags
	if legacy.Category != "" && !contains(tags, legacy.Category) {
		tags = append(tags, legacy.Category)
	}

	return PromptMeta{
		Name:        legacy.Name,
		Author:      legacy.Author,
		Tags:        tags,
		Version:     legacy.Version,
		Description: legacy.Description,
		Parent:      legacy.Parent,
		ID:          legacy.ID,
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
