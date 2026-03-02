// Package luoshu implements the cross-session intelligent memory system for Claude Code.
//
// It provides vector-indexed semantic search, LLM-powered recall, memory extraction,
// and multi-provider OpenAI-compatible API support. Core components include:
//
//   - Config: local configuration management (~/.luoshu/config.json)
//   - Store: JSONL-based memory persistence with tagging and metadata
//   - Searcher: keyword and semantic search over memory entries
//   - Recaller: LLM-synthesized intelligent recall combining multiple sources
//   - VectorIndex: embedding-based vector similarity search
//   - ClaudeIndex: file-level semantic search over .claude/ directory files
//   - OpenAICompatProvider: multi-provider LLM/embedding client (OpenAI, DeepSeek, etc.)
package luoshu
