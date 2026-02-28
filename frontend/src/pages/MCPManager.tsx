import { useEffect, useState, useRef, useCallback } from 'react';
import {
  GetGlobalMCPServers,
  SaveGlobalMCPServers,
  SearchMarketplace,
  InstallFromMarketplace,
} from '../../wailsjs/go/services/MCPService';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import HelpTip from '../components/HelpTip';

interface MCPServer {
  name: string;
  type?: string;
  url?: string;
  headers?: Record<string, string>;
  command?: string;
  args?: string[];
  timeout?: number;
}

interface MarketplaceServer {
  name: string;
  description: string;
  descriptionCN: string;
  repoUrl: string;
  package: string;
  transport: string;
  command: string;
  args: string[];
  version: string;
  publishedAt: string;
}

type EditMode = 'list' | 'add' | 'edit';
type TabMode = 'my' | 'market';
type MarketSource = 'official' | 'smithery' | 'glama';

const MARKET_SOURCES: { id: MarketSource; label: string; url: string }[] = [
  { id: 'official', label: 'Official Registry', url: 'https://registry.modelcontextprotocol.io' },
  { id: 'smithery', label: 'Smithery', url: 'https://smithery.ai' },
  { id: 'glama', label: 'Glama', url: 'https://glama.ai/mcp/servers' },
];

export default function MCPManager() {
  const [servers, setServers] = useState<MCPServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [mode, setMode] = useState<EditMode>('list');
  const [editingServer, setEditingServer] = useState<MCPServer | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  // 市场相关状态
  const [tab, setTab] = useState<TabMode>('my');
  const [marketSource, setMarketSource] = useState<MarketSource>('smithery');
  const [marketQuery, setMarketQuery] = useState('');
  const [marketResults, setMarketResults] = useState<MarketplaceServer[]>([]);
  const [marketLoading, setMarketLoading] = useState(false);
  const [marketLoadingMore, setMarketLoadingMore] = useState(false);
  const [marketError, setMarketError] = useState('');
  const [installingPkg, setInstallingPkg] = useState<string | null>(null);
  const [marketSearched, setMarketSearched] = useState(false);
  const [marketCursor, setMarketCursor] = useState('');
  const [marketTotal, setMarketTotal] = useState(0);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const loadServers = () => {
    setLoading(true);
    GetGlobalMCPServers()
      .then((data: any) => setServers(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadServers(); }, []);

  const doSearch = useCallback((source: MarketSource, query: string, cursor?: string) => {
    const isLoadMore = !!cursor;
    if (isLoadMore) {
      setMarketLoadingMore(true);
    } else {
      setMarketLoading(true);
      setMarketResults([]);
    }
    setMarketError('');
    SearchMarketplace(source, query, cursor || '')
      .then((data: any) => {
        const newServers = data?.servers || [];
        if (isLoadMore) {
          setMarketResults(prev => {
            const existingNames = new Set(prev.map((s: MarketplaceServer) => s.name));
            const unique = newServers.filter((s: MarketplaceServer) => !existingNames.has(s.name));
            return [...prev, ...unique];
          });
        } else {
          setMarketResults(newServers);
        }
        setMarketCursor(data?.nextCursor || '');
        if (data?.total > 0) setMarketTotal(data.total);
        setMarketSearched(true);
      })
      .catch((err: any) => setMarketError(String(err)))
      .finally(() => {
        setMarketLoading(false);
        setMarketLoadingMore(false);
      });
  }, []);

  const handleMarketSearch = (value: string) => {
    setMarketQuery(value);
    setMarketCursor('');
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(marketSource, value), 300);
  };

  const handleSourceChange = (source: MarketSource) => {
    setMarketSource(source);
    setMarketResults([]);
    setMarketCursor('');
    setMarketSearched(false);
    setMarketTotal(0);
    doSearch(source, marketQuery);
  };

  const handleLoadMore = () => {
    if (marketCursor) {
      doSearch(marketSource, marketQuery, marketCursor);
    }
  };

  // 切换到市场 Tab 时自动加载
  const handleTabChange = (t: TabMode) => {
    setTab(t);
    if (t === 'market' && !marketSearched) {
      doSearch(marketSource, '');
    }
  };

  const handleInstall = async (ms: MarketplaceServer) => {
    setInstallingPkg(ms.package);
    try {
      await InstallFromMarketplace(ms.name, ms.package);
      loadServers(); // 刷新本地列表
    } catch (err) {
      console.error(err);
    } finally {
      setInstallingPkg(null);
    }
  };

  const isInstalled = (pkg: string) => {
    return servers.some(s => s.args?.includes(pkg));
  };

  const handleSave = async (server: MCPServer) => {
    setSaving(true);
    try {
      let newServers: MCPServer[];
      if (mode === 'edit') {
        newServers = servers.map(s => s.name === editingServer?.name ? server : s);
      } else {
        newServers = [...servers, server];
      }
      await SaveGlobalMCPServers(newServers as any);
      setServers(newServers);
      setMode('list');
      setEditingServer(null);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (name: string) => {
    const newServers = servers.filter(s => s.name !== name);
    try {
      await SaveGlobalMCPServers(newServers as any);
      setServers(newServers);
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) {
    return <div className="page-container"><div className="loading-state">加载中...</div></div>;
  }

  if (mode !== 'list') {
    return (
      <div className="page-container">
        <MCPServerForm
          server={mode === 'edit' ? editingServer! : undefined}
          existingNames={servers.map(s => s.name)}
          isEdit={mode === 'edit'}
          saving={saving}
          onSave={handleSave}
          onCancel={() => { setMode('list'); setEditingServer(null); }}
        />
      </div>
    );
  }

  return (
    <div className={`page-container ${tab === 'market' ? 'page-full' : ''}`}>
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">MCP 服务管理</h1>
          <p className="page-subtitle">~/.claude/.mcp.json - 管理外部工具和服务集成</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="mcp" />
          {saved && <span className="save-indicator">已保存</span>}
          {tab === 'my' && (
            <button className="btn btn-primary" onClick={() => setMode('add')}>
              添加 MCP 服务
            </button>
          )}
        </div>
      </div>

      <div className="mcp-tabs">
        <button className={`mcp-tab ${tab === 'my' ? 'active' : ''}`}
          onClick={() => handleTabChange('my')}>
          我的服务 {servers.length > 0 && <span className="mcp-tab-count">{servers.length}</span>}
        </button>
        <button className={`mcp-tab ${tab === 'market' ? 'active' : ''}`}
          onClick={() => handleTabChange('market')}>
          在线市场
        </button>
      </div>

      {tab === 'my' ? (
        <>
          {servers.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">🔌</div>
              <h3>暂无 MCP 服务</h3>
              <p>添加 MCP 服务器以扩展 Claude Code 的能力</p>
              <button className="btn btn-primary" onClick={() => setMode('add')}>
                添加第一个 MCP 服务
              </button>
            </div>
          ) : (
            <div className="mcp-grid">
              {servers.map(srv => (
                <MCPServerCard
                  key={srv.name}
                  server={srv}
                  onEdit={() => { setEditingServer(srv); setMode('edit'); }}
                  onDelete={() => handleDelete(srv.name)}
                />
              ))}
            </div>
          )}
        </>
      ) : (
        <div className="marketplace-panel">
          <div className="marketplace-sources">
            {MARKET_SOURCES.map(src => (
              <button
                key={src.id}
                className={`marketplace-source-btn ${marketSource === src.id ? 'active' : ''}`}
                onClick={() => handleSourceChange(src.id)}
              >
                {src.label}
              </button>
            ))}
            {marketTotal > 0 && (
              <span className="marketplace-total">{marketTotal.toLocaleString()} 个服务</span>
            )}
            <a href="#" className="marketplace-source-link" onClick={e => {
              e.preventDefault();
              BrowserOpenURL(MARKET_SOURCES.find(src => src.id === marketSource)!.url);
            }}>
              访问官网 ↗
            </a>
          </div>
          <div className="marketplace-search">
            <input
              className="form-input marketplace-search-input"
              value={marketQuery}
              onChange={e => handleMarketSearch(e.target.value)}
              placeholder="搜索 MCP 服务器... 如 filesystem, github, postgres"
            />
          </div>

          <div className="marketplace-scroll">
            {marketError && <div className="json-error">{marketError}</div>}

            {marketLoading ? (
              <div className="loading-state">搜索中...</div>
            ) : marketResults.length === 0 && marketSearched ? (
              <div className="empty-state">
                <div className="empty-icon">🔍</div>
                <h3>未找到结果</h3>
                <p>换个关键词试试</p>
              </div>
            ) : (
              <>
                <div className="mcp-grid">
                  {marketResults.map(ms => {
                    const installed = isInstalled(ms.package);
                    return (
                      <div key={ms.package || ms.name} className="mcp-card marketplace-card">
                        <div className="mcp-card-header">
                          <span className="mcp-card-icon">📦</span>
                          <span className="mcp-card-name">{ms.name}</span>
                          {ms.transport && (
                            <span className="badge marketplace-transport-badge">{ms.transport}</span>
                          )}
                        </div>
                        {ms.descriptionCN && (
                          <div className="marketplace-card-desc-cn">{ms.descriptionCN}</div>
                        )}
                        <div className="marketplace-card-desc">{ms.description}</div>
                        <div className="marketplace-card-meta">
                          {ms.package && <span className="marketplace-pkg">{ms.package}</span>}
                          {ms.version && <span className="marketplace-ver">v{ms.version}</span>}
                        </div>
                        <div className="mcp-card-actions">
                          {ms.repoUrl && (
                            <button className="btn btn-ghost"
                              onClick={() => BrowserOpenURL(ms.repoUrl)}>
                              仓库
                            </button>
                          )}
                          {installed ? (
                            <span className="badge marketplace-installed-badge">已安装</span>
                          ) : (
                            <button className="btn btn-primary marketplace-install-btn"
                              disabled={installingPkg === ms.package}
                              onClick={() => handleInstall(ms)}>
                              {installingPkg === ms.package ? '安装中...' : '安装'}
                            </button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>

                <div className="marketplace-footer">
                  {marketCursor && (
                    <button className="btn btn-ghost marketplace-load-more"
                      disabled={marketLoadingMore}
                      onClick={handleLoadMore}>
                      {marketLoadingMore ? '加载中...' : '加载更多'}
                    </button>
                  )}
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function MCPServerCard({ server, onEdit, onDelete }: {
  server: MCPServer;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const isHTTP = server.type === 'http';
  return (
    <div className="mcp-card">
      <div className="mcp-card-header">
        <span className="mcp-card-icon">{isHTTP ? '🌐' : '⚡'}</span>
        <span className="mcp-card-name">{server.name}</span>
        <span className="badge">{isHTTP ? 'HTTP' : 'CLI'}</span>
      </div>
      <div className="mcp-card-detail">
        {isHTTP ? server.url : `${server.command} ${(server.args || []).join(' ')}`}
      </div>
      {server.timeout && (
        <div className="mcp-card-meta">超时: {server.timeout}s</div>
      )}
      <div className="mcp-card-actions">
        <button className="btn btn-ghost" onClick={onEdit}>编辑</button>
        <button className="btn btn-danger" onClick={onDelete}>删除</button>
      </div>
    </div>
  );
}

function MCPServerForm({ server, existingNames, isEdit, saving, onSave, onCancel }: {
  server?: MCPServer;
  existingNames: string[];
  isEdit: boolean;
  saving: boolean;
  onSave: (server: MCPServer) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(server?.name || '');
  const [type, setType] = useState<string>(server?.type || 'command');
  const [url, setUrl] = useState(server?.url || '');
  const [headersStr, setHeadersStr] = useState(
    server?.headers ? JSON.stringify(server.headers, null, 2) : '{}'
  );
  const [command, setCommand] = useState(server?.command || '');
  const [argsStr, setArgsStr] = useState(server?.args?.join('\n') || '');
  const [timeout, setTimeout_] = useState(String(server?.timeout || ''));
  const [error, setError] = useState('');

  const isHTTP = type === 'http';

  const validate = (): boolean => {
    if (!name.trim()) { setError('名称不能为空'); return false; }
    if (!isEdit && existingNames.includes(name)) { setError('名称已存在'); return false; }
    if (isHTTP && !url.trim()) { setError('URL 不能为空'); return false; }
    if (!isHTTP && !command.trim()) { setError('命令不能为空'); return false; }
    try {
      if (isHTTP && headersStr.trim()) JSON.parse(headersStr);
    } catch { setError('Headers JSON 格式错误'); return false; }
    setError('');
    return true;
  };

  const handleSubmit = () => {
    if (!validate()) return;
    const srv: MCPServer = { name: name.trim() };
    if (isHTTP) {
      srv.type = 'http';
      srv.url = url.trim();
      try {
        const h = JSON.parse(headersStr);
        if (Object.keys(h).length > 0) srv.headers = h;
      } catch {}
    } else {
      srv.command = command.trim();
      const args = argsStr.split('\n').map(a => a.trim()).filter(Boolean);
      if (args.length > 0) srv.args = args;
    }
    if (timeout) srv.timeout = parseInt(timeout);
    onSave(srv);
  };

  return (
    <>
      <div className="page-header">
        <div className="page-header-left">
          <button className="btn btn-ghost" onClick={onCancel}>← 返回</button>
          <h1 className="page-title">{isEdit ? '编辑 MCP 服务' : '添加 MCP 服务'}</h1>
        </div>
        <div className="page-header-right">
          <button className="btn btn-primary" onClick={handleSubmit} disabled={saving}>
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      {error && <div className="json-error">{error}</div>}

      <div className="form-sections">
        <FormSection title="基本信息">
          <FormRow label="服务名称" desc="唯一标识符，如 context7, github">
            <input
              className="form-input"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="my-mcp-server"
              disabled={isEdit}
            />
          </FormRow>
          <FormRow label="类型">
            <div className="view-toggle">
              <button className={`toggle-btn ${!isHTTP ? 'active' : ''}`}
                onClick={() => setType('command')}>命令行</button>
              <button className={`toggle-btn ${isHTTP ? 'active' : ''}`}
                onClick={() => setType('http')}>HTTP</button>
            </div>
          </FormRow>
        </FormSection>

        {isHTTP ? (
          <FormSection title="HTTP 配置">
            <FormRow label="URL" desc="MCP 服务器地址">
              <input className="form-input" value={url}
                onChange={e => setUrl(e.target.value)}
                placeholder="https://api.example.com/mcp/" />
            </FormRow>
            <FormRow label="请求头" desc="JSON 格式，支持 ${ENV_VAR} 引用环境变量">
              <textarea className="form-textarea" value={headersStr}
                onChange={e => setHeadersStr(e.target.value)}
                rows={4} spellCheck={false}
                placeholder='{"Authorization": "Bearer ${API_KEY}"}' />
            </FormRow>
          </FormSection>
        ) : (
          <FormSection title="命令行配置">
            <FormRow label="命令" desc="要执行的命令">
              <input className="form-input" value={command}
                onChange={e => setCommand(e.target.value)}
                placeholder="npx" />
            </FormRow>
            <FormRow label="参数" desc="每行一个参数">
              <textarea className="form-textarea" value={argsStr}
                onChange={e => setArgsStr(e.target.value)}
                rows={4} spellCheck={false}
                placeholder={"-y\n@upstash/context7-mcp"} />
            </FormRow>
          </FormSection>
        )}

        <FormSection title="高级设置">
          <FormRow label="超时 (秒)" desc="可选，不填则使用默认值">
            <input className="form-input" type="number" value={timeout}
              onChange={e => setTimeout_(e.target.value)}
              placeholder="10" style={{ width: 100 }} />
          </FormRow>
        </FormSection>
      </div>
    </>
  );
}

function FormSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="form-section">
      <div className="form-section-header">
        <h3 className="form-section-title">{title}</h3>
      </div>
      <div className="form-section-body">{children}</div>
    </div>
  );
}

function FormRow({ label, desc, children }: {
  label: string; desc?: string; children: React.ReactNode;
}) {
  return (
    <div className="form-field">
      <div className="form-field-label">
        <span className="field-label">{label}</span>
        {desc && <span className="field-desc">{desc}</span>}
      </div>
      <div className="form-field-control">{children}</div>
    </div>
  );
}
