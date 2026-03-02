---
layout: home

hero:
  name: claude-config-mcp
  text: Configuration & Memory for Claude Code
  tagline: MCP server that manages .claude/ configuration and adds cross-session intelligent memory
  actions:
    - theme: brand
      text: Get Started
      link: /guide/installation
    - theme: alt
      text: API Reference
      link: /reference/tools
    - theme: alt
      text: GitHub
      link: https://github.com/znlnzi/claude-config-studio

features:
  - icon: ⚙️
    title: Configuration Management
    details: Manage Claude Code's .claude/ directory through MCP tools — memory files, CLAUDE.md, settings.json, templates, extensions, and hooks.
  - icon: 🧠
    title: Luoshu Intelligent Memory
    details: Cross-session memory powered by LLM and vector search. Claude remembers your decisions, preferences, and project context across conversations.
  - icon: 📦
    title: Template System
    details: Install pre-built configuration packs for common workflows — hackathon, solo developer, cross-session memory, and more.
  - icon: 🔍
    title: Semantic Search
    details: Find related memories and rules using natural language. Vector similarity search goes beyond keyword matching.
  - icon: 🔌
    title: Multi-Provider Support
    details: Works with OpenAI, DeepSeek, Moonshot, Zhipu, SiliconFlow, Volcengine, or any OpenAI-compatible API.
  - icon: 📡
    title: Dual Transport
    details: Run as stdio for local Claude Code integration or HTTP for Docker and shared deployments.
---

## Architecture

```mermaid
graph TB
    CC[Claude Code CLI] -->|MCP Protocol| MCP[claude-config-mcp Server]

    MCP --> CM[Configuration Management]
    MCP --> LU[Luoshu Memory Engine]
    MCP --> OV[File Semantic Search]

    CM --> Memory[Memory Tools]
    CM --> Config[Config Tools]
    CM --> Templates[Template Engine]
    CM --> Extensions[Extension Manager]
    CM --> Hooks[Hooks Manager]
    CM --> Evolution[Evolution Analyzer]

    LU --> Store[Memory Store — JSONL]
    LU --> VecIdx[Vector Index — Embeddings]
    LU --> Recall[Intelligent Recall — LLM Synthesis]
    LU --> Provider[Multi-Provider — OpenAI Compatible]

    OV --> ClaudeIdx[Claude Index — rules + memory files]

    style MCP fill:#f9f,stroke:#333,stroke-width:2px
    style LU fill:#bbf,stroke:#333
    style CM fill:#bfb,stroke:#333
```
