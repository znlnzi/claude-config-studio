import { useEffect, useState } from 'react';
import { GetGlobalStats } from '../../wailsjs/go/services/ProjectService';
import HelpTip from '../components/HelpTip';

interface Stats {
  hasGlobalClaudeMd: boolean;
  hasGlobalSettings: boolean;
  hasLspConfig: boolean;
  globalAgentCount: number;
  globalCommandCount: number;
  projectCount: number;
  enabledPluginCount: number;
}

interface DashboardProps {
  onNavigate: (id: string) => void;
}

export default function Dashboard({ onNavigate }: DashboardProps) {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    GetGlobalStats()
      .then((data: any) => setStats(data as Stats))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading-state">加载中...</div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">仪表盘</h1>
          <p className="page-subtitle">Claude Code 配置概览</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="dashboard" />
        </div>
      </div>

      {/* 状态卡片 */}
      <div className="stats-grid">
        <StatCard
          icon="📝"
          title="全局指令"
          value={stats?.hasGlobalClaudeMd ? '已配置' : '未配置'}
          status={stats?.hasGlobalClaudeMd ? 'success' : 'warning'}
          onClick={() => onNavigate('global-claudemd')}
        />
        <StatCard
          icon="⚙️"
          title="全局设置"
          value={stats?.hasGlobalSettings ? '已配置' : '未配置'}
          status={stats?.hasGlobalSettings ? 'success' : 'warning'}
          onClick={() => onNavigate('global-settings')}
        />
        <StatCard
          icon="📁"
          title="项目数量"
          value={String(stats?.projectCount ?? 0)}
          status="info"
          onClick={() => onNavigate('projects')}
        />
        <StatCard
          icon="🔌"
          title="启用插件"
          value={String(stats?.enabledPluginCount ?? 0)}
          status="info"
          onClick={() => onNavigate('mcp')}
        />
        <StatCard
          icon="🤖"
          title="Agent 模板"
          value={String(stats?.globalAgentCount ?? 0)}
          status="info"
          onClick={() => onNavigate('agents')}
        />
        <StatCard
          icon="⌨️"
          title="自定义命令"
          value={String(stats?.globalCommandCount ?? 0)}
          status="info"
          onClick={() => onNavigate('commands')}
        />
      </div>

      {/* 快速操作 */}
      <div className="section">
        <h2 className="section-title">快速操作</h2>
        <div className="quick-actions">
          <QuickAction
            icon="✏️"
            title="编辑全局指令"
            desc="修改 ~/.claude/CLAUDE.md"
            onClick={() => onNavigate('global-claudemd')}
          />
          <QuickAction
            icon="📂"
            title="配置新项目"
            desc="为项目创建 Claude Code 配置"
            onClick={() => onNavigate('projects')}
          />
          <QuickAction
            icon="📋"
            title="使用模板"
            desc="从最佳实践模板快速配置"
            onClick={() => onNavigate('templates')}
          />
        </div>
      </div>
    </div>
  );
}

function StatCard({ icon, title, value, status, onClick }: {
  icon: string;
  title: string;
  value: string;
  status: 'success' | 'warning' | 'info';
  onClick: () => void;
}) {
  const statusColor = {
    success: 'var(--macos-success)',
    warning: 'var(--macos-warning)',
    info: 'var(--macos-accent)',
  }[status];

  return (
    <button className="stat-card" onClick={onClick}>
      <div className="stat-icon">{icon}</div>
      <div className="stat-info">
        <div className="stat-title">{title}</div>
        <div className="stat-value" style={{ color: statusColor }}>{value}</div>
      </div>
    </button>
  );
}

function QuickAction({ icon, title, desc, onClick }: {
  icon: string;
  title: string;
  desc: string;
  onClick: () => void;
}) {
  return (
    <button className="quick-action-card" onClick={onClick}>
      <div className="quick-action-icon">{icon}</div>
      <div className="quick-action-info">
        <div className="quick-action-title">{title}</div>
        <div className="quick-action-desc">{desc}</div>
      </div>
      <div className="quick-action-arrow">→</div>
    </button>
  );
}
