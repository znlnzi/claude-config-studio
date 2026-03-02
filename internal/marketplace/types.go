package marketplace

// IndexEntry represents a single template entry in the registry index.
type IndexEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	URL         string   `json:"url"`
	Downloads   int      `json:"downloads"`
	Stars       int      `json:"stars"`
}

// RegistryIndex represents the top-level structure of the remote index.json.
type RegistryIndex struct {
	Version   int          `json:"version"`
	UpdatedAt string       `json:"updated_at"`
	Templates []IndexEntry `json:"templates"`
}
