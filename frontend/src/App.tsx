import { useState } from 'react';
import Sidebar from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import GlobalClaudeMd from './pages/GlobalClaudeMd';
import GlobalSettings from './pages/GlobalSettings';
import Projects from './pages/Projects';
import ProjectDetail from './pages/ProjectDetail';
import MCPManager from './pages/MCPManager';
import HooksManager from './pages/HooksManager';
import Templates from './pages/Templates';
import ProjectWizard from './pages/ProjectWizard';
import ExtensionManager from './pages/ExtensionManager';
import PluginManager from './pages/PluginManager';
import ImportExport from './pages/ImportExport';
import SkillsManager from './pages/SkillsManager';
import Help from './pages/Help';

const COMMAND_TEMPLATE = `---
description: 命令的简短描述
allowed-tools: [Read, Glob, Grep, Bash]
---

# 命令名称

在此编写命令的详细说明和执行逻辑。

使用 $ARGUMENTS 获取用户传入的参数。
`;

const AGENT_TEMPLATE = `---
name: agent-name
description: Agent 角色描述
model: sonnet
---

你是一个专业的 [角色名称]。

## 目标
[描述 Agent 的目标和定位]

## 能力
- 能力 1
- 能力 2

## 工作流程
1. 步骤 1
2. 步骤 2
`;

function App() {
  const [activeId, setActiveId] = useState('dashboard');
  const [projectPath, setProjectPath] = useState<string | null>(null);

  const handleNavigate = (id: string, data?: any) => {
    if (id === 'project-detail' && data?.path) {
      setProjectPath(data.path);
    }
    setActiveId(id);
  };

  const renderContent = () => {
    switch (activeId) {
      case 'dashboard':
        return <Dashboard onNavigate={handleNavigate} />;
      case 'global-claudemd':
        return <GlobalClaudeMd />;
      case 'global-settings':
        return <GlobalSettings />;
      case 'projects':
        return <Projects onNavigate={handleNavigate} />;
      case 'project-detail':
        return projectPath ? (
          <ProjectDetail
            projectPath={projectPath}
            onBack={() => setActiveId('projects')}
          />
        ) : null;
      case 'project-wizard':
        return (
          <ProjectWizard
            onComplete={(path) => {
              setProjectPath(path);
              setActiveId('project-detail');
            }}
            onCancel={() => setActiveId('projects')}
          />
        );
      case 'mcp':
        return <MCPManager />;
      case 'hooks':
        return <HooksManager />;
      case 'templates':
        return <Templates />;
      case 'commands':
        return (
          <ExtensionManager
            type="commands"
            title="Commands 管理"
            icon="⌨️"
            description="~/.claude/commands/ - 管理自定义斜杠命令"
            newFileTemplate={COMMAND_TEMPLATE}
          />
        );
      case 'agents':
        return (
          <ExtensionManager
            type="agents"
            title="Agents 管理"
            icon="🤖"
            description="~/.claude/agents/ - 管理 Agent 角色模板"
            newFileTemplate={AGENT_TEMPLATE}
          />
        );
      case 'skills':
        return <SkillsManager />;
      case 'plugins':
        return <PluginManager />;
      case 'import-export':
        return <ImportExport />;
      case 'help':
        return <Help />;
      default:
        return <Dashboard onNavigate={handleNavigate} />;
    }
  };

  return (
    <div className="app-layout">
      <Sidebar activeId={activeId} onNavigate={handleNavigate} />
      <main className="main-content">
        <div className="content-titlebar titlebar-drag" />
        <div className="content-body">
          {renderContent()}
        </div>
      </main>
    </div>
  );
}

export default App;
