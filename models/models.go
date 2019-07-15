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

// Tag represents a tag that is assigned to a document
type Tag struct {
	Title   string  `json:"title"`
	Context *string `json:"context"`
}
