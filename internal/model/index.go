package model

import "time"

// Index represents a prompt index entry that links prompt files to GitHub gist URLs
type IndexedPrompt struct {
	GistURL     string    `json:"gist_url"`
	FilePath    string    `json:"file_path"`
	Author      string    `json:"author"`      // 存储 YAML 中的 author
	Name        string    `json:"name"`        // 存储 prompt 名称
	LastUpdated time.Time `json:"last_updated"`
}

type Index struct {
	Prompts     []IndexedPrompt `json:"prompts"`
	LastUpdated time.Time       `json:"last_updated"`
}
