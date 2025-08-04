package model

import "time"

// Index represents a prompt index entry that links prompt files to GitHub gist URLs
type IndexedPrompt struct {
	GistURL     string    `json:"gist_url"`
	FilePath    string    `json:"file_path"`
	LastUpdated time.Time `json:"last_updated"`
}

type Index struct {
	Prompts     []IndexedPrompt `json:"prompts"`
	LastUpdated time.Time       `json:"last_updated"`
}
