package model

type Prompt struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Author      string   `json:"author"`
	GistURL     string   `json:"gist_url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Version     string   `json:"version"`
	Content     string   `json:"content"`
	Parent      *string  `json:"parent,omitempty"`  // 新增：父级 Prompt 的 gist URL
}
