package luoshu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PreValidateKey performs format pre-validation (no network requests)
// Returns the detected provider name, or an error
func PreValidateKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if len(key) < 20 {
		return "", fmt.Errorf("API Key is too short (minimum 20 characters), possibly incomplete")
	}

	switch {
	case strings.HasPrefix(key, "sk-proj-"):
		return "openai", nil
	case strings.HasPrefix(key, "sk-ant-"):
		return "anthropic", nil
	case strings.HasPrefix(key, "AKIA"):
		return "aws", nil
	case strings.HasPrefix(key, "ghp_"):
		return "github", nil
	}

	return "unknown", nil
}

// MaskKey masks an API Key: keeps the first 3 and last 4 characters, replaces the middle with ****
func MaskKey(key string) string {
	if len(key) <= 7 {
		return "****"
	}
	return key[:3] + "****" + key[len(key)-4:]
}

// connectionTestTimeout is the timeout duration for connection tests
const connectionTestTimeout = 10 * time.Second

// TestConnection performs a connection test by sending a minimal request to verify the API Key and Endpoint
// Returns: connected, status, error
// status: "ok" | "auth_failed" | "quota_exceeded" | "network_error" | "timeout"
func TestConnection(cfg *Config) (bool, string, error) {
	if cfg.LLM.APIKey == "" {
		return false, "auth_failed", fmt.Errorf("LLM API Key is not set")
	}
	if cfg.LLM.Endpoint == "" {
		return false, "network_error", fmt.Errorf("LLM Endpoint is not set")
	}

	url := strings.TrimRight(cfg.LLM.Endpoint, "/") + "/chat/completions"

	// Construct a minimal test request
	body := map[string]interface{}{
		"model": cfg.LLM.Model,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 1,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return false, "network_error", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return false, "network_error", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.LLM.APIKey)

	client := &http.Client{Timeout: connectionTestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return false, "timeout", err
		}
		return false, "network_error", err
	}
	defer resp.Body.Close()

	return classifyResponse(resp.StatusCode)
}

// classifyResponse classifies the connection test result based on HTTP status code
func classifyResponse(statusCode int) (bool, string, error) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return true, "ok", nil
	case statusCode == 401 || statusCode == 403:
		return false, "auth_failed", fmt.Errorf("authentication failed (HTTP %d)", statusCode)
	case statusCode == 429:
		return false, "quota_exceeded", fmt.Errorf("quota exceeded (HTTP 429)")
	default:
		return false, "network_error", fmt.Errorf("request failed (HTTP %d)", statusCode)
	}
}

// isTimeoutError checks whether the error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded")
}
