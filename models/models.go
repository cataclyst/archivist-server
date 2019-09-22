package models

// Document represents a single document in the data store
type Document struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Date        string   `json:"date"`
	TagIDs      []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"`
	ModifiedAt  string   `json:"modifiedAt"`
}

// DocumentInput represents the data that must be specified to create a new Document
type DocumentInput struct {
	Title       string      `json:"title"`
	Description *string     `json:"description"`
	Date        string      `json:"date"`
	Tags        []*TagInput `json:"tags"`
	BinaryData  *string     `json:"binaryData"`
}

// Tag represents a tag that is assigned to a document
type Tag struct {
	Title   string  `json:"title"`
	Context *string `json:"context"`
}

// TagInput represents the data that must be specified to create a new Tag
// (or associate an existing one with a Document).
type TagInput struct {
	Title   string  `json:"title"`
	Context *string `json:"context"`
}
