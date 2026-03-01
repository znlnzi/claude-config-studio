package luoshu

import "testing"

func TestAllPresets_ContainsExpectedProviders(t *testing.T) {
	presets := AllPresets()
	expected := []string{"volcengine", "openai", "deepseek", "moonshot", "zhipu", "siliconflow", "custom"}
	for _, name := range expected {
		if _, ok := presets[name]; !ok {
			t.Errorf("expected preset %q not found", name)
		}
	}
}

func TestAllPresets_VolcengineHasDefaults(t *testing.T) {
	presets := AllPresets()
	v := presets["volcengine"]
	if v.LLMEndpoint == "" {
		t.Error("volcengine preset missing LLMEndpoint")
	}
	if v.LLMModel == "" {
		t.Error("volcengine preset missing LLMModel")
	}
	if v.EmbedModel == "" {
		t.Error("volcengine preset missing EmbedModel")
	}
	if v.EmbedDimension <= 0 {
		t.Error("volcengine preset missing EmbedDimension")
	}
}

func TestAllPresets_OpenAIHasDefaults(t *testing.T) {
	presets := AllPresets()
	o := presets["openai"]
	if o.LLMEndpoint != "https://api.openai.com/v1" {
		t.Errorf("unexpected openai LLMEndpoint: %s", o.LLMEndpoint)
	}
	if o.KeyPrefix != "sk-" {
		t.Errorf("expected openai KeyPrefix 'sk-', got %q", o.KeyPrefix)
	}
}

func TestAllPresets_CustomHasNoDefaults(t *testing.T) {
	presets := AllPresets()
	c := presets["custom"]
	if c.LLMEndpoint != "" {
		t.Errorf("custom preset should have empty LLMEndpoint, got %q", c.LLMEndpoint)
	}
}

func TestGetPreset_Exists(t *testing.T) {
	p := GetPreset("openai")
	if p == nil {
		t.Fatal("expected non-nil preset for openai")
	}
	if p.Name != "openai" {
		t.Errorf("expected name 'openai', got %q", p.Name)
	}
}

func TestGetPreset_NotFound(t *testing.T) {
	p := GetPreset("nonexistent")
	if p != nil {
		t.Error("expected nil for nonexistent preset")
	}
}
