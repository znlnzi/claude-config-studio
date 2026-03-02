package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/znlnzi/claude-config-studio/internal/templatedata"
)

func sampleIndex() RegistryIndex {
	return RegistryIndex{
		Version:   1,
		UpdatedAt: "2026-03-02T10:00:00Z",
		Templates: []IndexEntry{
			{
				ID:          "community/react-testing",
				Name:        "React Testing Best Practices",
				Description: "Comprehensive React testing patterns",
				Author:      "johndoe",
				Version:     "1.0.0",
				Category:    "Frontend",
				Tags:        []string{"react", "testing"},
				URL:         "", // filled in per-test
				Downloads:   100,
				Stars:       10,
			},
			{
				ID:          "community/go-microservice",
				Name:        "Go Microservice Starter",
				Description: "Production-ready Go microservice template",
				Author:      "janedoe",
				Version:     "2.0.0",
				Category:    "Backend",
				Tags:        []string{"go", "microservice"},
				URL:         "",
				Downloads:   200,
				Stars:       25,
			},
		},
	}
}

func TestFetchIndex(t *testing.T) {
	idx := sampleIndex()
	data, _ := json.Marshal(idx)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	got, err := client.FetchIndex(context.Background())
	if err != nil {
		t.Fatalf("FetchIndex failed: %v", err)
	}
	if got.Version != 1 {
		t.Errorf("version = %d, want 1", got.Version)
	}
	if len(got.Templates) != 2 {
		t.Errorf("templates count = %d, want 2", len(got.Templates))
	}
}

func TestFetchIndex_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	_, err := client.FetchIndex(context.Background())
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestFetchIndex_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	_, err := client.FetchIndex(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchTemplate(t *testing.T) {
	tmpl := templatedata.Template{
		ID:       "community/react-testing",
		Name:     "React Testing Best Practices",
		Category: "Frontend",
		ClaudeMd: "# React Testing\nTest your components.",
	}
	data, _ := json.Marshal(tmpl)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	got, err := client.FetchTemplate(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("FetchTemplate failed: %v", err)
	}
	if got.ID != "community/react-testing" {
		t.Errorf("ID = %q, want community/react-testing", got.ID)
	}
	if got.ClaudeMd == "" {
		t.Error("expected non-empty ClaudeMd")
	}
}

func TestFetchTemplate_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	_, err := client.FetchTemplate(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
}

func TestSearch(t *testing.T) {
	entries := sampleIndex().Templates

	t.Run("empty query returns all", func(t *testing.T) {
		got := Search(entries, "")
		if len(got) != len(entries) {
			t.Errorf("got %d, want %d", len(got), len(entries))
		}
	})

	t.Run("match by name", func(t *testing.T) {
		got := Search(entries, "react")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
		if got[0].ID != "community/react-testing" {
			t.Errorf("got %q", got[0].ID)
		}
	})

	t.Run("match by tag", func(t *testing.T) {
		got := Search(entries, "microservice")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("match by author", func(t *testing.T) {
		got := Search(entries, "janedoe")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got := Search(entries, "REACT")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := Search(entries, "nonexistent")
		if len(got) != 0 {
			t.Errorf("got %d, want 0", len(got))
		}
	})
}

func TestFilterByCategory(t *testing.T) {
	entries := sampleIndex().Templates

	t.Run("empty category returns all", func(t *testing.T) {
		got := FilterByCategory(entries, "")
		if len(got) != len(entries) {
			t.Errorf("got %d, want %d", len(got), len(entries))
		}
	})

	t.Run("filter frontend", func(t *testing.T) {
		got := FilterByCategory(entries, "Frontend")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got := FilterByCategory(entries, "backend")
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := FilterByCategory(entries, "Mobile")
		if len(got) != 0 {
			t.Errorf("got %d, want 0", len(got))
		}
	})
}

func TestNewClient_DefaultURL(t *testing.T) {
	client := NewClient("")
	if client.registryURL != DefaultRegistryURL {
		t.Errorf("expected default URL, got %q", client.registryURL)
	}
}

func TestNewClient_CustomURL(t *testing.T) {
	custom := "https://example.com/index.json"
	client := NewClient(custom)
	if client.registryURL != custom {
		t.Errorf("expected %q, got %q", custom, client.registryURL)
	}
}
