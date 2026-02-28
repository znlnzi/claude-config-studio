import { useEffect, useState, useCallback } from 'react';
import {
  GetTemplates,
  InstallTemplateRules,
  UninstallTemplateRules,
  GetInstalledTemplates,
} from '../../wailsjs/go/services/TemplateService';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import HelpTip from '../components/HelpTip';

interface Template {
  id: string;
  name: string;
  category: string;
  description: string;
  tags: string[];
  claudeMd?: string;
  settings?: any;
  mcpServers?: any;
  hooks?: any;
  agents?: Record<string, string>;
  commands?: Record<string, string>;
  skills?: Record<string, string>;
  rules?: Record<string, string>;
}

interface TemplateCategory {
  id: string;
  name: string;
  icon: string;
  templates: Template[];
}

interface InstalledInfo {
  templateId: string;
  scope: string;
  filePath: string;
}

type InstallScope = 'global' | 'project';

export default function Templates() {
  const [categories, setCategories] = useState<TemplateCategory[]>([]);
  const [loading, setLoading] = useState(true);

  // 安装范围
  const [installScope, setInstallScope] = useState<InstallScope>('global');
  const [projectPath, setProjectPath] = useState('');

  // 已安装状态
  const [installedGlobal, setInstalledGlobal] = useState<string[]>([]);
  const [installedProject, setInstalledProject] = useState<string[]>([]);

  // 操作状态
  const [installingIds, setInstallingIds] = useState<Set<string>>(new Set());
  const [updatingIds, setUpdatingIds] = useState<Set<string>>(new Set());
  const [uninstallingIds, setUninstallingIds] = useState<Set<string>>(new Set());

  // 操作提示
  const [statusMsg, setStatusMsg] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const showStatus = (type: 'success' | 'error', message: string) => {
    setStatusMsg({ type, message });
    setTimeout(() => setStatusMsg(null), 4000);
  };

  // 预览
  const [previewTemplate, setPreviewTemplate] = useState<Template | null>(null);
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({});

  // 加载模板列表
  useEffect(() => {
    GetTemplates()
      .then((data: any) => setCategories(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  // 刷新已安装状态
  const refreshInstalled = useCallback(() => {
    GetInstalledTemplates('global', '')
      .then((data: any) => {
        setInstalledGlobal((data || []).map((i: InstalledInfo) => i.templateId));
      })
      .catch(console.error);

    if (projectPath) {
      GetInstalledTemplates('project', projectPath)
        .then((data: any) => {
          setInstalledProject((data || []).map((i: InstalledInfo) => i.templateId));
        })
        .catch(console.error);
    } else {
      setInstalledProject([]);
    }
  }, [projectPath]);

  useEffect(() => { refreshInstalled(); }, [refreshInstalled]);

  // 选择项目目录
  const handleSelectProject = async () => {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        setProjectPath(dir);
        setInstallScope('project');
      }
    } catch (err) {
      console.error(err);
    }
  };

  // 切换范围
  const handleScopeChange = (scope: InstallScope) => {
    if (scope === 'project' && !projectPath) {
      handleSelectProject();
      return;
    }
    setInstallScope(scope);
  };

  // 安装
  const handleInstall = async (templateId: string) => {
    if (installScope === 'project' && !projectPath) {
      handleSelectProject();
      return;
    }

    setInstallingIds(prev => new Set(prev).add(templateId));
    try {
      await InstallTemplateRules(installScope, installScope === 'global' ? '' : projectPath, [templateId], false);
      refreshInstalled();
      showStatus('success', '安装成功');
    } catch (err: any) {
      showStatus('error', err?.message || '安装失败');
    } finally {
      setInstallingIds(prev => {
        const next = new Set(prev);
        next.delete(templateId);
        return next;
      });
    }
  };

  // 更新（强制覆盖）
  const handleUpdate = async (templateId: string, scope: InstallScope) => {
    setUpdatingIds(prev => new Set(prev).add(templateId));
    try {
      await InstallTemplateRules(scope, scope === 'global' ? '' : projectPath, [templateId], true);
      refreshInstalled();
      showStatus('success', '更新成功');
    } catch (err: any) {
      showStatus('error', err?.message || '更新失败');
    } finally {
      setUpdatingIds(prev => {
        const next = new Set(prev);
        next.delete(templateId);
        return next;
      });
    }
  };

  // 卸载
  const handleUninstall = async (templateId: string, scope: InstallScope) => {
    setUninstallingIds(prev => new Set(prev).add(templateId));
    try {
      await UninstallTemplateRules(scope, scope === 'global' ? '' : projectPath, [templateId]);
      refreshInstalled();
      showStatus('success', '卸载成功');
    } catch (err: any) {
      showStatus('error', err?.message || '卸载失败');
    } finally {
      setUninstallingIds(prev => {
        const next = new Set(prev);
        next.delete(templateId);
        return next;
      });
    }
  };

  // 获取安装状态
  const getInstallStatus = (id: string): { installed: boolean; scope?: string } => {
    if (installedGlobal.includes(id)) return { installed: true, scope: 'global' };
    if (installedProject.includes(id)) return { installed: true, scope: 'project' };
    return { installed: false };
  };

  // 统计组件数量
  const countExtensions = (tmpl: Template) => {
    const counts: { label: string; count: number }[] = [];
    if (tmpl.agents && Object.keys(tmpl.agents).length > 0)
      counts.push({ label: 'Agents', count: Object.keys(tmpl.agents).length });
    if (tmpl.commands && Object.keys(tmpl.commands).length > 0)
      counts.push({ label: 'Commands', count: Object.keys(tmpl.commands).length });
    if (tmpl.skills && Object.keys(tmpl.skills).length > 0)
      counts.push({ label: 'Skills', count: Object.keys(tmpl.skills).length });
    if (tmpl.rules && Object.keys(tmpl.rules).length > 0)
      counts.push({ label: 'Rules', count: Object.keys(tmpl.rules).length });
    return counts;
  };

  const toggleSection = (key: string) => {
    setExpandedSections(prev => ({ ...prev, [key]: !prev[key] }));
  };

  if (loading) {
    return <div className="page-container"><div className="loading-state">加载中...</div></div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">模板库</h1>
          <p className="page-subtitle">
            基于最佳实践的配置模板，安装为 rules 文件，多个模板可共存
          </p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="templates" />
        </div>
      </div>

      {/* 安装范围选择 */}
      <div className="template-scope-bar">
        <span className="template-scope-label">安装到：</span>
        <div className="marketplace-sources">
          <button
            className={`marketplace-source-btn ${installScope === 'global' ? 'active' : ''}`}
            onClick={() => handleScopeChange('global')}
          >
            全局 (~/.claude/rules/)
          </button>
          <button
            className={`marketplace-source-btn ${installScope === 'project' ? 'active' : ''}`}
            onClick={() => handleScopeChange('project')}
          >
            项目级
          </button>
        </div>
        {installScope === 'project' && (
          <button className="btn btn-ghost template-scope-path" onClick={handleSelectProject}>
            {projectPath ? projectPath : '选择项目目录'}
          </button>
        )}
      </div>

      {/* 操作提示 */}
      {statusMsg && (
        <div className={statusMsg.type === 'success' ? 'status-success' : 'json-error'}>
          {statusMsg.message}
        </div>
      )}

      {/* 模板分类列表 */}
      {categories.map(cat => (
        <div key={cat.id} className="template-category">
          <h2 className="template-category-title">
            <span>{cat.icon}</span> {cat.name}
          </h2>
          <div className="template-grid">
            {cat.templates.map(tmpl => {
              const extCounts = countExtensions(tmpl);
              const status = getInstallStatus(tmpl.id);
              const isInstalling = installingIds.has(tmpl.id);
              const isUninstalling = uninstallingIds.has(tmpl.id);

              return (
                <div
                  key={tmpl.id}
                  className={`template-card ${status.installed ? 'installed' : ''}`}
                >
                  <div className="template-card-name">{tmpl.name}</div>
                  <div className="template-card-desc">{tmpl.description}</div>
                  <div className="template-card-tags">
                    {tmpl.tags.map(tag => (
                      <span key={tag} className="badge">{tag}</span>
                    ))}
                  </div>
                  <div className="template-card-includes">
                    {tmpl.claudeMd && <span className="template-includes-item">Rules</span>}
                    {tmpl.settings && <span className="template-includes-item">Settings</span>}
                    {extCounts.map(({ label, count }) => (
                      <span key={label} className="template-includes-item template-includes-ext">
                        {count} {label}
                      </span>
                    ))}
                  </div>

                  {/* 状态 + 操作 */}
                  <div className="template-card-actions">
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => { setPreviewTemplate(tmpl); setExpandedSections({}); }}
                    >
                      预览
                    </button>
                    {status.installed ? (
                      <>
                        <span className="badge badge-builtin template-status-badge">
                          已安装({status.scope === 'global' ? '全局' : '项目'})
                        </span>
                        <button
                          className="btn btn-secondary btn-sm"
                          disabled={updatingIds.has(tmpl.id)}
                          onClick={() => handleUpdate(tmpl.id, status.scope as InstallScope)}
                        >
                          {updatingIds.has(tmpl.id) ? '更新中...' : '更新'}
                        </button>
                        <button
                          className="btn btn-danger btn-sm"
                          disabled={isUninstalling}
                          onClick={() => handleUninstall(tmpl.id, status.scope as InstallScope)}
                        >
                          {isUninstalling ? '卸载中...' : '卸载'}
                        </button>
                      </>
                    ) : (
                      <button
                        className="btn btn-primary btn-sm"
                        disabled={isInstalling}
                        onClick={() => handleInstall(tmpl.id)}
                      >
                        {isInstalling ? '安装中...' : '安装'}
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ))}

      {/* 预览弹窗 */}
      {previewTemplate && (
        <div className="template-modal-overlay" onClick={() => setPreviewTemplate(null)}>
          <div className="template-modal" onClick={e => e.stopPropagation()}>
            <div className="template-modal-header">
              <div>
                <h2 className="template-modal-title">{previewTemplate.name}</h2>
                <p className="template-modal-desc">{previewTemplate.description}</p>
              </div>
              <button className="btn btn-ghost" onClick={() => setPreviewTemplate(null)}>
                关闭
              </button>
            </div>

            <div className="template-modal-body">
              <div className="template-badges">
                {previewTemplate.tags.map(tag => (
                  <span key={tag} className="badge">{tag}</span>
                ))}
              </div>

              <div className="template-content-sections">
                {previewTemplate.claudeMd && (
                  <div className="template-content-section">
                    <h3 className="template-content-title">Rules 内容 (tpl-{previewTemplate.id}.md)</h3>
                    <pre className="template-code">{previewTemplate.claudeMd}</pre>
                  </div>
                )}
                {previewTemplate.settings && (
                  <div className="template-content-section">
                    <h3 className="template-content-title">settings.json</h3>
                    <pre className="template-code">{JSON.stringify(previewTemplate.settings, null, 2)}</pre>
                  </div>
                )}

                {/* Agents */}
                {previewTemplate.agents && Object.keys(previewTemplate.agents).length > 0 && (
                  <div className="template-content-section">
                    <button className="template-ext-header" onClick={() => toggleSection('agents')}>
                      <h3 className="template-content-title">
                        Agents ({Object.keys(previewTemplate.agents).length})
                      </h3>
                      <span className="template-ext-toggle">
                        {expandedSections['agents'] ? '收起' : '展开'}
                      </span>
                    </button>
                    {!expandedSections['agents'] ? (
                      <div className="template-ext-badges">
                        {Object.keys(previewTemplate.agents).map(name => (
                          <span key={name} className="badge">{name}</span>
                        ))}
                      </div>
                    ) : (
                      <div className="template-ext-files">
                        {Object.entries(previewTemplate.agents).map(([name, content]) => (
                          <div key={name} className="template-ext-file">
                            <div className="template-ext-file-name">{name}.md</div>
                            <pre className="template-code">{content}</pre>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Commands */}
                {previewTemplate.commands && Object.keys(previewTemplate.commands).length > 0 && (
                  <div className="template-content-section">
                    <button className="template-ext-header" onClick={() => toggleSection('commands')}>
                      <h3 className="template-content-title">
                        Commands ({Object.keys(previewTemplate.commands).length})
                      </h3>
                      <span className="template-ext-toggle">
                        {expandedSections['commands'] ? '收起' : '展开'}
                      </span>
                    </button>
                    {!expandedSections['commands'] ? (
                      <div className="template-ext-badges">
                        {Object.keys(previewTemplate.commands).map(name => (
                          <span key={name} className="badge">{name}</span>
                        ))}
                      </div>
                    ) : (
                      <div className="template-ext-files">
                        {Object.entries(previewTemplate.commands).map(([name, content]) => (
                          <div key={name} className="template-ext-file">
                            <div className="template-ext-file-name">{name}.md</div>
                            <pre className="template-code">{content}</pre>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Skills */}
                {previewTemplate.skills && Object.keys(previewTemplate.skills).length > 0 && (
                  <div className="template-content-section">
                    <button className="template-ext-header" onClick={() => toggleSection('skills')}>
                      <h3 className="template-content-title">
                        Skills ({Object.keys(previewTemplate.skills).length})
                      </h3>
                      <span className="template-ext-toggle">
                        {expandedSections['skills'] ? '收起' : '展开'}
                      </span>
                    </button>
                    {!expandedSections['skills'] ? (
                      <div className="template-ext-badges">
                        {Object.keys(previewTemplate.skills).map(name => (
                          <span key={name} className="badge">{name}</span>
                        ))}
                      </div>
                    ) : (
                      <div className="template-ext-files">
                        {Object.entries(previewTemplate.skills).map(([name, content]) => (
                          <div key={name} className="template-ext-file">
                            <div className="template-ext-file-name">{name}.md</div>
                            <pre className="template-code">{content}</pre>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Rules */}
                {previewTemplate.rules && Object.keys(previewTemplate.rules).length > 0 && (
                  <div className="template-content-section">
                    <button className="template-ext-header" onClick={() => toggleSection('rules')}>
                      <h3 className="template-content-title">
                        Rules ({Object.keys(previewTemplate.rules).length})
                      </h3>
                      <span className="template-ext-toggle">
                        {expandedSections['rules'] ? '收起' : '展开'}
                      </span>
                    </button>
                    {!expandedSections['rules'] ? (
                      <div className="template-ext-badges">
                        {Object.keys(previewTemplate.rules).map(name => (
                          <span key={name} className="badge">{name}</span>
                        ))}
                      </div>
                    ) : (
                      <div className="template-ext-files">
                        {Object.entries(previewTemplate.rules).map(([name, content]) => (
                          <div key={name} className="template-ext-file">
                            <div className="template-ext-file-name">{name}.md</div>
                            <pre className="template-code">{content}</pre>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>

            <div className="template-modal-footer">
              {(() => {
                const st = getInstallStatus(previewTemplate.id);
                if (st.installed) {
                  return (
                    <>
                      <span className="badge badge-builtin">已安装({st.scope === 'global' ? '全局' : '项目'})</span>
                      <button
                        className="btn btn-secondary"
                        disabled={updatingIds.has(previewTemplate.id)}
                        onClick={() => handleUpdate(previewTemplate.id, st.scope as InstallScope)}
                      >
                        {updatingIds.has(previewTemplate.id) ? '更新中...' : '更新'}
                      </button>
                      <button
                        className="btn btn-danger"
                        disabled={uninstallingIds.has(previewTemplate.id)}
                        onClick={() => handleUninstall(previewTemplate.id, st.scope as InstallScope)}
                      >
                        {uninstallingIds.has(previewTemplate.id) ? '卸载中...' : '卸载'}
                      </button>
                    </>
                  );
                }
                return (
                  <button
                    className="btn btn-primary"
                    disabled={installingIds.has(previewTemplate.id)}
                    onClick={() => handleInstall(previewTemplate.id)}
                  >
                    {installingIds.has(previewTemplate.id) ? '安装中...' : `安装到${installScope === 'global' ? '全局' : '项目'}`}
                  </button>
                );
              })()}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
