import { useEffect, useState, useCallback, useRef } from 'react';
import Editor from '@monaco-editor/react';
import {
  ListExtensions,
  GetExtension,
  SaveExtension,
  DeleteExtension,
  SearchOnlineExtensions,
  InstallOnlineExtension,
} from '../../wailsjs/go/services/ExtensionService';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import HelpTip from '../components/HelpTip';

interface ExtensionFile {
  name: string;
  fileName: string;
  path: string;
  content: string;
  lastModified: string;
  size: number;
}

interface OnlineExtension {
  name: string;
  description: string;
  category: string;
  source: string;
  repoUrl: string;
  downloadUrl: string;
  extType: string;
}

interface ExtensionManagerProps {
  type: 'commands' | 'agents';
  title: string;
  icon: string;
  description: string;
  newFileTemplate: string;
}

type TabMode = 'my' | 'market';
type SourceFilter = 'all' | 'builtin' | 'github';

export default function ExtensionManager({ type, title, icon, description, newFileTemplate }: ExtensionManagerProps) {
  const [files, setFiles] = useState<ExtensionFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<ExtensionFile | null>(null);
  const [editContent, setEditContent] = useState('');
  const [originalContent, setOriginalContent] = useState('');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [newFileName, setNewFileName] = useState('');

  // 在线市场状态（仅 agents 使用）
  const [tab, setTab] = useState<TabMode>('my');
  const [sourceFilter, setSourceFilter] = useState<SourceFilter>('all');
  const [marketQuery, setMarketQuery] = useState('');
  const [marketResults, setMarketResults] = useState<OnlineExtension[]>([]);
  const [marketLoading, setMarketLoading] = useState(false);
  const [marketError, setMarketError] = useState('');
  const [marketSearched, setMarketSearched] = useState(false);
  const [installingName, setInstallingName] = useState<string | null>(null);
  const [categoryFilter, setCategoryFilter] = useState('');
  const [marketVisibleCount, setMarketVisibleCount] = useState(24);
  const [marketInstallScope, setMarketInstallScope] = useState<string>('global');
  const [marketProjectPath, setMarketProjectPath] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const hasMarket = type === 'agents';

  const loadFiles = useCallback(() => {
    setLoading(true);
    ListExtensions(type, 'global')
      .then((data: any) => setFiles(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [type]);

  useEffect(() => { loadFiles(); }, [loadFiles]);

  const handleSelect = async (file: ExtensionFile) => {
    try {
      const detail: any = await GetExtension(type, 'global', file.fileName);
      setSelectedFile(detail);
      setEditContent(detail.content || '');
      setOriginalContent(detail.content || '');
      setIsCreating(false);
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreate = () => {
    setIsCreating(true);
    setNewFileName('');
    setSelectedFile(null);
    setEditContent(newFileTemplate);
    setOriginalContent('');
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      let fileName: string;
      if (isCreating) {
        fileName = newFileName.endsWith('.md') ? newFileName : newFileName + '.md';
      } else {
        fileName = selectedFile!.fileName;
      }
      await SaveExtension(type, 'global', fileName, editContent);
      setSaved(true);
      setOriginalContent(editContent);
      setTimeout(() => setSaved(false), 2000);
      if (isCreating) {
        setIsCreating(false);
        loadFiles();
      }
    } catch (err) {
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (file: ExtensionFile) => {
    try {
      await DeleteExtension(type, 'global', file.fileName);
      if (selectedFile?.fileName === file.fileName) {
        setSelectedFile(null);
        setEditContent('');
      }
      loadFiles();
    } catch (err) {
      console.error(err);
    }
  };

  const hasChanges = editContent !== originalContent;

  // 快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        if (hasChanges || isCreating) handleSave();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [hasChanges, isCreating, editContent]);

  // === 在线市场逻辑 ===
  const doSearch = useCallback((source: string, query: string) => {
    setMarketLoading(true);
    setMarketResults([]);
    setMarketError('');
    setMarketVisibleCount(24);
    SearchOnlineExtensions('agents', source, query)
      .then((data: any) => {
        setMarketResults(data?.extensions || []);
        setMarketSearched(true);
      })
      .catch((err: any) => setMarketError(String(err)))
      .finally(() => setMarketLoading(false));
  }, []);

  const handleMarketSearch = (value: string) => {
    setMarketQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(sourceFilter === 'all' ? '' : sourceFilter, value), 300);
  };

  const handleSourceFilterChange = (sf: SourceFilter) => {
    setSourceFilter(sf);
    setCategoryFilter('');
    setMarketVisibleCount(24);
    doSearch(sf === 'all' ? '' : sf, marketQuery);
  };

  const handleTabChange = (t: TabMode) => {
    setTab(t);
    if (t === 'market' && !marketSearched) {
      doSearch('', '');
    }
  };

  const handleInstall = async (ext: OnlineExtension, installScope: string = 'global') => {
    setInstallingName(ext.name);
    try {
      await InstallOnlineExtension('agents', ext as any, installScope);
      loadFiles();
    } catch (err) {
      console.error(err);
    } finally {
      setInstallingName(null);
    }
  };

  const isInstalled = (name: string) => {
    const normalizedName = name.toLowerCase().replace(/[^a-z0-9]/g, '');
    return files.some(f => {
      const fn = f.fileName.replace('.md', '').toLowerCase().replace(/[^a-z0-9]/g, '');
      return fn === normalizedName;
    });
  };

  // 获取所有分类
  const categories = [...new Set(marketResults.map(e => e.category).filter(Boolean))];

  // 过滤结果
  const filteredMarketResults = categoryFilter
    ? marketResults.filter(e => e.category === categoryFilter)
    : marketResults;

  if (loading) {
    return <div className="page-container"><div className="loading-state">加载中...</div></div>;
  }

  return (
    <div className={`page-container ${tab === 'my' ? 'page-full' : ''}`}>
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">{title}</h1>
          <p className="page-subtitle">{description}</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId={type} />
          {tab === 'my' && saved && <span className="save-indicator">已保存</span>}
          {tab === 'my' && hasChanges && <span className="unsaved-indicator">未保存</span>}
          {tab === 'my' && (
            <button className="btn btn-primary" onClick={handleCreate}>
              新建
            </button>
          )}
        </div>
      </div>

      {hasMarket && (
        <div className="mcp-tabs">
          <button className={`mcp-tab ${tab === 'my' ? 'active' : ''}`}
            onClick={() => handleTabChange('my')}>
            我的 Agents {files.length > 0 && <span className="mcp-tab-count">{files.length}</span>}
          </button>
          <button className={`mcp-tab ${tab === 'market' ? 'active' : ''}`}
            onClick={() => handleTabChange('market')}>
            在线市场
          </button>
        </div>
      )}

      {tab === 'my' ? (
        <div className="ext-layout">
          {/* 文件列表侧边栏 */}
          <div className="ext-sidebar">
            <div className="ext-file-list">
              {files.length === 0 && !isCreating && (
                <div className="ext-empty">
                  <span>{icon}</span>
                  <span>暂无文件</span>
                </div>
              )}
              {files.map(file => (
                <button
                  key={file.fileName}
                  className={`ext-file-item ${selectedFile?.fileName === file.fileName && !isCreating ? 'active' : ''}`}
                  onClick={() => handleSelect(file)}
                >
                  <div className="ext-file-name">{file.name}</div>
                  <button
                    className="ext-file-delete"
                    onClick={e => { e.stopPropagation(); handleDelete(file); }}
                    title="删除"
                  >
                    ×
                  </button>
                </button>
              ))}
            </div>
          </div>

          {/* 编辑区 */}
          <div className="ext-editor">
            {isCreating && (
              <div className="ext-new-bar">
                <input
                  className="form-input"
                  value={newFileName}
                  onChange={e => setNewFileName(e.target.value)}
                  placeholder="输入文件名（不含 .md 后缀）"
                  autoFocus
                />
                <button className="btn btn-primary" onClick={handleSave}
                  disabled={!newFileName.trim() || saving}>
                  {saving ? '保存中...' : '创建'}
                </button>
                <button className="btn btn-ghost" onClick={() => setIsCreating(false)}>
                  取消
                </button>
              </div>
            )}

            {(selectedFile || isCreating) ? (
              <div className="ext-editor-content">
                {!isCreating && (
                  <div className="ext-editor-header">
                    <span className="ext-editor-filename">{selectedFile!.fileName}</span>
                    <button className="btn btn-primary" onClick={handleSave}
                      disabled={!hasChanges || saving}>
                      {saving ? '保存中...' : '保存'}
                    </button>
                  </div>
                )}
                <div className="ext-monaco">
                  <Editor
                    height="100%"
                    defaultLanguage="markdown"
                    value={editContent}
                    onChange={v => setEditContent(v ?? '')}
                    theme="vs"
                    options={{
                      fontSize: 13,
                      lineHeight: 20,
                      fontFamily: "'SF Mono', 'Monaco', 'Menlo', monospace",
                      minimap: { enabled: false },
                      wordWrap: 'on',
                      scrollBeyondLastLine: false,
                      padding: { top: 12, bottom: 12 },
                      automaticLayout: true,
                      tabSize: 2,
                      smoothScrolling: true,
                    }}
                  />
                </div>
              </div>
            ) : (
              <div className="ext-empty-editor">
                <span style={{ fontSize: 36 }}>{icon}</span>
                <p>选择一个文件进行编辑，或创建新文件</p>
              </div>
            )}
          </div>
        </div>
      ) : (
        /* 在线市场面板 */
        <div className="marketplace-panel">
          <div className="marketplace-sources">
            {(['all', 'builtin', 'github'] as SourceFilter[]).map(sf => (
              <button
                key={sf}
                className={`marketplace-source-btn ${sourceFilter === sf ? 'active' : ''}`}
                onClick={() => handleSourceFilterChange(sf)}
              >
                {sf === 'all' ? '全部' : sf === 'builtin' ? '内置精选' : 'GitHub'}
              </button>
            ))}
            <a href="#" className="marketplace-source-link" onClick={e => {
              e.preventDefault();
              BrowserOpenURL('https://github.com/anthropics/claude-code');
            }}>
              Claude Code ↗
            </a>
          </div>

          <div className="marketplace-search">
            <input
              className="form-input marketplace-search-input"
              value={marketQuery}
              onChange={e => handleMarketSearch(e.target.value)}
              placeholder="搜索 Agent... 如 frontend, python, devops"
            />
          </div>

          <div className="skill-scope-bar">
            <span className="skill-scope-label">安装到：</span>
            <div className="marketplace-sources">
              <button
                className={`marketplace-source-btn ${marketInstallScope === 'global' ? 'active' : ''}`}
                onClick={() => setMarketInstallScope('global')}
              >
                全局
              </button>
              <button
                className={`marketplace-source-btn ${marketInstallScope === 'project' ? 'active' : ''}`}
                onClick={async () => {
                  if (!marketProjectPath) {
                    try {
                      const dir = await SelectDirectory();
                      if (dir) {
                        setMarketProjectPath(dir);
                        setMarketInstallScope('project');
                      }
                    } catch {}
                  } else {
                    setMarketInstallScope('project');
                  }
                }}
              >
                项目级
              </button>
            </div>
            {marketInstallScope === 'project' && (
              <button className="btn btn-ghost skill-scope-path" onClick={async () => {
                try {
                  const dir = await SelectDirectory();
                  if (dir) setMarketProjectPath(dir);
                } catch {}
              }}>
                {marketProjectPath || '选择项目目录'}
              </button>
            )}
          </div>

          {/* 分类筛选 */}
          {categories.length > 1 && (
            <div className="marketplace-categories">
              <button
                className={`marketplace-category-btn ${!categoryFilter ? 'active' : ''}`}
                onClick={() => setCategoryFilter('')}
              >
                全部 ({marketResults.length})
              </button>
              {categories.map(cat => {
                const count = marketResults.filter(e => e.category === cat).length;
                return (
                  <button
                    key={cat}
                    className={`marketplace-category-btn ${categoryFilter === cat ? 'active' : ''}`}
                    onClick={() => setCategoryFilter(cat)}
                  >
                    {cat} ({count})
                  </button>
                );
              })}
            </div>
          )}

          <div className="marketplace-scroll">
            {marketError && <div className="json-error">{marketError}</div>}

            {marketLoading ? (
              <div className="loading-state">搜索中...</div>
            ) : filteredMarketResults.length === 0 && marketSearched ? (
              <div className="empty-state">
                <div className="empty-icon">🔍</div>
                <h3>未找到结果</h3>
                <p>换个关键词试试</p>
              </div>
            ) : (
              <>
                <div className="marketplace-result-info">
                  共 {filteredMarketResults.length} 个 Agent
                  {filteredMarketResults.length > marketVisibleCount && `，当前显示 ${marketVisibleCount} 个`}
                </div>
                <div className="mcp-grid">
                  {filteredMarketResults.slice(0, marketVisibleCount).map(ext => {
                    const installed = isInstalled(ext.name);
                    return (
                      <div key={`${ext.source}-${ext.name}`} className="mcp-card marketplace-card">
                        <div className="mcp-card-header">
                          <span className="mcp-card-icon">🤖</span>
                          <span className="mcp-card-name">{ext.name}</span>
                          <span className={`badge ${ext.source === 'builtin' ? 'badge-builtin' : 'badge-github'}`}>
                            {ext.source === 'builtin' ? '内置' : 'GitHub'}
                          </span>
                        </div>
                        <div className="marketplace-card-desc-cn">{ext.description}</div>
                        {ext.category && (
                          <div className="marketplace-card-meta">
                            <span className="marketplace-pkg">{ext.category}</span>
                          </div>
                        )}
                        <div className="mcp-card-actions">
                          {ext.repoUrl && (
                            <button className="btn btn-ghost"
                              onClick={() => BrowserOpenURL(ext.repoUrl)}>
                              仓库
                            </button>
                          )}
                          {installed ? (
                            <span className="badge marketplace-installed-badge">已安装</span>
                          ) : (
                            <button className="btn btn-primary marketplace-install-btn"
                              disabled={installingName === ext.name || (marketInstallScope === 'project' && !marketProjectPath)}
                              onClick={() => handleInstall(ext, marketInstallScope === 'project' ? marketProjectPath : 'global')}>
                              {installingName === ext.name ? '安装中...' : '安装'}
                            </button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
                {filteredMarketResults.length > marketVisibleCount && (
                  <div className="marketplace-footer">
                    <button
                      className="btn btn-ghost marketplace-load-more"
                      onClick={() => setMarketVisibleCount(prev => prev + 24)}
                    >
                      加载更多（还有 {filteredMarketResults.length - marketVisibleCount} 个）
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
