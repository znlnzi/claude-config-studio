import { useState } from 'react';

interface SidebarProps {
  activeId: string;
  onNavigate: (id: string) => void;
}

interface NavSection {
  title?: string;
  items: NavItemDef[];
}

interface NavItemDef {
  id: string;
  label: string;
  icon: string;
  children?: NavItemDef[];
}

const navSections: NavSection[] = [
  {
    items: [
      { id: 'dashboard', label: '仪表盘', icon: '⌘' },
    ],
  },
  {
    title: '全局配置',
    items: [
      { id: 'global-claudemd', label: '指令文件', icon: '📝' },
      { id: 'global-settings', label: '全局设置', icon: '⚙️' },
    ],
  },
  {
    title: '项目管理',
    items: [
      { id: 'projects', label: '项目列表', icon: '📁' },
      { id: 'project-wizard', label: '新建配置', icon: '✨' },
    ],
  },
  {
    title: '扩展管理',
    items: [
      { id: 'mcp', label: 'MCP 服务', icon: '🔌' },
      { id: 'hooks', label: 'Hooks', icon: '🪝' },
      { id: 'commands', label: 'Commands', icon: '⌨️' },
      { id: 'agents', label: 'Agents', icon: '🤖' },
      { id: 'skills', label: 'Skills', icon: '🎯' },
      { id: 'plugins', label: '插件', icon: '📦' },
    ],
  },
  {
    title: '工具',
    items: [
      { id: 'templates', label: '模板库', icon: '📋' },
      { id: 'import-export', label: '导入/导出', icon: '💾' },
      { id: 'help', label: '帮助文档', icon: '❓' },
    ],
  },
];

export default function Sidebar({ activeId, onNavigate }: SidebarProps) {
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggleSection = (title: string) => {
    setCollapsed(prev => ({ ...prev, [title]: !prev[title] }));
  };

  return (
    <aside className="sidebar">
      {/* macOS 窗口拖拽区域 */}
      <div className="sidebar-titlebar titlebar-drag" />

      <nav className="sidebar-nav">
        {navSections.map((section, sIdx) => (
          <div key={sIdx} className="nav-section">
            {section.title && (
              <div
                className="nav-section-title"
                onClick={() => toggleSection(section.title!)}
              >
                {section.title}
              </div>
            )}
            {!collapsed[section.title || ''] && (
              <ul className="nav-list">
                {section.items.map(item => (
                  <li key={item.id}>
                    <button
                      className={`nav-item ${activeId === item.id ? 'active' : ''}`}
                      onClick={() => onNavigate(item.id)}
                    >
                      <span className="nav-icon">{item.icon}</span>
                      <span className="nav-label">{item.label}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        ))}
      </nav>
    </aside>
  );
}
