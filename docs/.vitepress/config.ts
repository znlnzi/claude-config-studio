import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(
  defineConfig({
    title: 'claude-config-mcp',
    description: 'MCP server for Claude Code configuration management and cross-session intelligent memory',

    themeConfig: {
      nav: [
        { text: 'Guide', link: '/guide/installation' },
        { text: 'Reference', link: '/reference/tools' },
        { text: 'Changelog', link: '/changelog' },
        {
          text: 'Links',
          items: [
            { text: 'npm', link: 'https://www.npmjs.com/package/claude-config-mcp' },
            { text: 'GitHub', link: 'https://github.com/znlnzi/claude-config-studio' },
            { text: 'Contributing', link: '/contributing' },
          ],
        },
      ],

      sidebar: {
        '/guide/': [
          {
            text: 'Getting Started',
            items: [
              { text: 'Installation', link: '/guide/installation' },
              { text: 'Quick Start', link: '/guide/quickstart' },
            ],
          },
          {
            text: 'Features',
            items: [
              { text: 'Configuration Management', link: '/guide/configuration' },
              { text: 'Luoshu Intelligent Memory', link: '/guide/luoshu' },
            ],
          },
        ],
        '/reference/': [
          {
            text: 'API Reference',
            items: [
              { text: 'MCP Tools', link: '/reference/tools' },
              { text: 'MCP Resources', link: '/reference/resources' },
              { text: 'Templates', link: '/reference/templates' },
              { text: 'Providers', link: '/reference/providers' },
            ],
          },
        ],
      },

      socialLinks: [
        { icon: 'github', link: 'https://github.com/znlnzi/claude-config-studio' },
      ],

      editLink: {
        pattern: 'https://github.com/znlnzi/claude-config-studio/edit/main/docs/:path',
        text: 'Edit this page on GitHub',
      },

      search: {
        provider: 'local',
      },

      footer: {
        message: 'Released under the MIT License.',
        copyright: 'Copyright © 2026-present',
      },
    },
  })
)
