package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/znlnzi/claude-config-studio/internal/templatedata"
)

const (
	// DefaultRegistryURL points to the community template registry hosted on GitHub.
	DefaultRegistryURL = "https://raw.githubusercontent.com/znlnzi/claude-config-studio/main/registry/index.json"

	maxIndexSize    = 2 * 1024 * 1024 // 2 MB
	maxTemplateSize = 1 * 1024 * 1024 // 1 MB
	httpTimeout     = 15 * time.Second
)

// Client fetches and searches the remote template registry.
type Client struct {
	httpClient  *http.Client
	registryURL string
}

// NewClient creates a Client. If registryURL is empty, DefaultRegistryURL is used.
func NewClient(registryURL string) *Client {
	if registryURL == "" {
		registryURL = DefaultRegistryURL
	}
	return &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
		registryURL: registryURL,
	}
}

// FetchIndex downloads and parses the registry index.json.
func (c *Client) FetchIndex(ctx context.Context) (*RegistryIndex, error) {
	body, err := c.fetchURL(ctx, c.registryURL, maxIndexSize)
	if err != nil {
		return nil, fmt.Errorf("fetch index: %w", err)
	}

	var idx RegistryIndex
	if err := json.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	return &idx, nil
}

// FetchTemplate downloads and parses a single template.json from the given URL.
func (c *Client) FetchTemplate(ctx context.Context, url string) (*templatedata.Template, error) {
	body, err := c.fetchURL(ctx, url, maxTemplateSize)
	if err != nil {
		return nil, fmt.Errorf("fetch template: %w", err)
	}

	var tmpl templatedata.Template
	if err := json.Unmarshal(body, &tmpl); err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &tmpl, nil
}

// fetchURL performs a GET request and reads the response body up to maxBytes.
func (c *Client) fetchURL(ctx context.Context, url string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxBytes))
}

// Search filters entries whose name, description, tags or author contain the query string (case-insensitive).
func Search(entries []IndexEntry, query string) []IndexEntry {
	if query == "" {
		return entries
	}
	q := strings.ToLower(query)
	var results []IndexEntry
	for _, e := range entries {
		if matchesQuery(e, q) {
			results = append(results, e)
		}
	}
	return results
}

// FilterByCategory returns entries that belong to the given category (case-insensitive).
func FilterByCategory(entries []IndexEntry, category string) []IndexEntry {
	if category == "" {
		return entries
	}
	cat := strings.ToLower(category)
	var results []IndexEntry
	for _, e := range entries {
		if strings.ToLower(e.Category) == cat {
			results = append(results, e)
		}
	}
	return results
}

// matchesQuery checks if an entry matches the search query.
func matchesQuery(e IndexEntry, q string) bool {
	if strings.Contains(strings.ToLower(e.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Description), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Author), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.ID), q) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}
