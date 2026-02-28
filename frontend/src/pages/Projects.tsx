import { useEffect, useState } from 'react';
import { ScanProjects } from '../../wailsjs/go/services/ProjectService';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import type { ProjectInfo } from '../types';
import HelpTip from '../components/HelpTip';

interface ProjectsProps {
  onNavigate: (id: string, data?: any) => void;
}

export default function Projects({ onNavigate }: ProjectsProps) {
  const [projects, setProjects] = useState<ProjectInfo[]>([]);
  const [loading, setLoading] = useState(true);

  const loadProjects = () => {
    setLoading(true);
    ScanProjects()
      .then((data: any) => setProjects(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const handleAddProject = async () => {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        onNavigate('project-detail', { path: dir });
      }
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading-state">扫描项目中...</div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">项目管理</h1>
          <p className="page-subtitle">管理已有项目的 Claude Code 配置</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="projects" />
          <button className="btn btn-secondary" onClick={loadProjects}>
            刷新
          </button>
          <button className="btn btn-primary" onClick={handleAddProject}>
            添加项目
          </button>
        </div>
      </div>

      {projects.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📁</div>
          <h3>暂无项目</h3>
          <p>未检测到已配置 Claude Code 的项目</p>
          <button className="btn btn-primary" onClick={handleAddProject}>
            添加项目
          </button>
        </div>
      ) : (
        <div className="project-grid">
          {projects.map(project => (
            <ProjectCard
              key={project.path}
              project={project}
              onClick={() => onNavigate('project-detail', { path: project.path })}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function ProjectCard({ project, onClick }: { project: ProjectInfo; onClick: () => void }) {
  const badges = [];
  if (project.hasClaudeMd) badges.push('CLAUDE.md');
  if (project.hasSettings) badges.push('Settings');
  if (project.hasMcp) badges.push('MCP');
  if (project.hasHooks) badges.push('Hooks');
  if (project.hasCommands) badges.push('Commands');
  if (project.hasAgents) badges.push('Agents');
  if (project.hasSkills) badges.push('Skills');

  return (
    <button className="project-card" onClick={onClick}>
      <div className="project-card-header">
        <span className="project-icon">📁</span>
        <span className="project-name">{project.name}</span>
      </div>
      <div className="project-path">{project.path}</div>
      <div className="project-badges">
        {badges.map(b => (
          <span key={b} className="badge">{b}</span>
        ))}
        {badges.length === 0 && (
          <span className="badge badge-empty">无配置</span>
        )}
      </div>
      <div className="project-meta">
        <span>{project.configCount} 项配置</span>
      </div>
    </button>
  );
}
