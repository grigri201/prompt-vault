package interfaces

// InputType represents the type of input for the add command
type InputType int

const (
	InputTypeFile InputType = iota
	InputTypeGistURL
	InputTypeUnknown
)

// AddService defines the interface for unified add service operations
type AddService interface {
	AddFromFile(filePath string, force bool) error
	AddFromGistURL(gistURL string, force bool) error
	DetectInputType(input string) InputType
}

// UploadService defines the interface for upload operations
type UploadService interface {
	UploadPrompt(filePath string, force bool) error
}

// ImportService defines the interface for import operations
type ImportService interface {
	ImportFromGist(gistURL string, force bool) error
}

// DeleteService defines the interface for delete operations
type DeleteService interface {
	DeletePrompt(promptID string) error
}

// SearchService defines the interface for search operations
type SearchService interface {
	SearchPrompts(query string) ([]string, error)
}

// ShareService defines the interface for share operations
type ShareService interface {
	SharePrompt(promptID string) (string, error)
}
