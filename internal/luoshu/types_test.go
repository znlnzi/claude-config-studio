package luoshu

import (
	"strings"
	"testing"
)

func TestNewMemoryID_Format(t *testing.T) {
	id := NewMemoryID()
	if !strings.HasPrefix(id, "mem-") {
		t.Errorf("expected prefix 'mem-', got %q", id)
	}
	// Format: mem-20260218-150405-a1b2c3d4 (length 28)
	if len(id) != 28 {
		t.Errorf("expected length 28, got %d for %q", len(id), id)
	}
}

func TestNewMemoryID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := NewMemoryID()
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
