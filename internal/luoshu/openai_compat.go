package luoshu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatProvider is an OpenAI-compatible API provider.
// Works with OpenAI, DeepSeek, Moonshot, Zhipu, Volcengine, SiliconFlow, and any custom endpoint.
type OpenAICompatProvider struct {
	providerName string
	apiKey       string
	endpoint     string
	llmModel     string
	embedModel   string
	dimensions   int
	httpClient   *http.Client
}

// ─── Interface implementation ─────────────────────────────────────

// Name returns the provider name
func (p *OpenAICompatProvider) Name() string { return p.providerName }

// Dimensions returns the vector dimensions
func (p *OpenAICompatProvider) Dimensions() int { return p.dimensions }

// Chat sends a chat request via OpenAI-compatible API
func (p *OpenAICompatProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	messages := make([]chatMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = chatMessage(m)
	}

	apiReq := chatCompletionRequest{
		Model:       p.llmModel,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	respBody, err := p.doRequest(ctx, "/chat/completions", apiReq)
	if err != nil {
		return nil, fmt.Errorf("Chat request failed: %w", err)
	}

	var apiResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Chat response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, errors.New("empty Chat response: no results returned")
	}

	return &ChatResponse{
		Content:    apiResp.Choices[0].Message.Content,
		TokensUsed: apiResp.Usage.TotalTokens,
	}, nil
}

// Embed converts text into vectors.
// Supports two API formats:
//   - Standard OpenAI format: /embeddings (input is a string array, data is an array)
//   - Multimodal format: /embeddings/multimodal (input is an object array, data is a single object)
//
// Auto-detection: uses multimodal format when embedModel contains "vision" or "multimodal"
func (p *OpenAICompatProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	if p.isMultimodalEmbed() {
		return p.embedMultimodal(ctx, texts)
	}
	return p.embedStandard(ctx, texts)
}

// isMultimodalEmbed determines whether to use the multimodal embedding API
func (p *OpenAICompatProvider) isMultimodalEmbed() bool {
	m := strings.ToLower(p.embedModel)
	return strings.Contains(m, "vision") || strings.Contains(m, "multimodal")
}

// embedStandard performs standard OpenAI-compatible embedding
func (p *OpenAICompatProvider) embedStandard(ctx context.Context, texts []string) ([][]float32, error) {
	apiReq := embeddingRequest{
		Model:          p.embedModel,
		Input:          texts,
		EncodingFormat: "float",
	}

	respBody, err := p.doRequest(ctx, "/embeddings", apiReq)
	if err != nil {
		return nil, fmt.Errorf("Embedding request failed: %w", err)
	}

	var apiResp embeddingResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Embedding response: %w", err)
	}

	vectors := make([][]float32, len(apiResp.Data))
	for i, d := range apiResp.Data {
		vectors[i] = d.Embedding
	}

	return vectors, nil
}

// embedMultimodal performs multimodal embedding (one request per text, each returning one vector)
func (p *OpenAICompatProvider) embedMultimodal(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))

	for i, text := range texts {
		apiReq := multimodalEmbeddingRequest{
			Model: p.embedModel,
			Input: []multimodalInput{{Type: "text", Text: text}},
		}

		respBody, err := p.doRequest(ctx, "/embeddings/multimodal", apiReq)
		if err != nil {
			return nil, fmt.Errorf("Embedding[%d] request failed: %w", i, err)
		}

		var apiResp multimodalEmbeddingResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse Embedding[%d] response: %w", i, err)
		}

		vectors[i] = apiResp.Data.Embedding
	}

	return vectors, nil
}

// ─── HTTP request layer ──────────────────────────────────

// doRequest sends an HTTP request with automatic retry on 5xx (2s backoff)
func (p *OpenAICompatProvider) doRequest(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	url := strings.TrimRight(p.endpoint, "/") + path

	// First attempt
	respBody, statusCode, err := p.sendHTTP(ctx, url, body)
	if err != nil {
		return nil, classifyNetError(err)
	}

	// Retry once on 5xx
	if statusCode >= 500 {
		time.Sleep(2 * time.Second)
		respBody, statusCode, err = p.sendHTTP(ctx, url, body)
		if err != nil {
			return nil, classifyNetError(err)
		}
	}

	return checkStatus(respBody, statusCode)
}

// sendHTTP sends a single HTTP POST request
func (p *OpenAICompatProvider) sendHTTP(ctx context.Context, url string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// ─── Error handling ─────────────────────────────────────

// classifyNetError classifies network errors into user-friendly messages
func classifyNetError(err error) error {
	if isTimeout(err) {
		return fmt.Errorf("network timeout: %w", err)
	}
	return fmt.Errorf("network request failed: %w", err)
}

// checkStatus returns the response body or an error based on the HTTP status code
func checkStatus(body []byte, statusCode int) ([]byte, error) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return body, nil
	case statusCode == 401 || statusCode == 403:
		return nil, errors.New("API key is invalid or expired")
	case statusCode == 429:
		return nil, errors.New("API quota exceeded")
	case statusCode >= 500:
		return nil, fmt.Errorf("server error: HTTP %d", statusCode)
	default:
		return nil, fmt.Errorf("request failed: HTTP %d, %s", statusCode, string(body))
	}
}

// isTimeout checks whether the error is a timeout error
func isTimeout(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

// ─── API request/response structures (internal use) ──────────────────

// chatCompletionRequest is an OpenAI-compatible chat request
type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// chatMessage represents a chat message
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionResponse represents a chat completion response
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// embeddingRequest represents an embedding request
type embeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format"`
}

// embeddingResponse represents an embedding response (standard OpenAI format, data is an array)
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// multimodalEmbeddingRequest represents a multimodal embedding request
type multimodalEmbeddingRequest struct {
	Model string            `json:"model"`
	Input []multimodalInput `json:"input"`
}

// multimodalInput represents a multimodal input item
type multimodalInput struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// multimodalEmbeddingResponse represents a multimodal embedding response (data is a single object)
type multimodalEmbeddingResponse struct {
	Data struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}
