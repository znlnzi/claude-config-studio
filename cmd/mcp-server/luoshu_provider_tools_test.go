package main

import (
	"context"
	"testing"
)

func TestHandleLuoshuProviderList(t *testing.T) {
	ctx := context.Background()

	t.Run("returns provider list", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleLuoshuProviderList(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		providers, ok := m["providers"].([]any)
		if !ok || len(providers) == 0 {
			t.Error("expected at least one provider")
		}

		// Verify known providers exist
		providerNames := make(map[string]bool)
		for _, p := range providers {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if name, ok := pm["name"].(string); ok {
				providerNames[name] = true
			}
		}

		for _, expected := range []string{"openai", "deepseek", "volcengine"} {
			if !providerNames[expected] {
				t.Errorf("expected provider %q not found", expected)
			}
		}

		// Verify usage hint exists
		if _, ok := m["usage"].(string); !ok {
			t.Error("expected usage string")
		}
	})
}
