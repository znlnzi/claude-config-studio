import { useEffect, useState } from 'react';
import {
  GetGlobalHooks,
  SaveGlobalHooks,
} from '../../wailsjs/go/services/HooksService';
import HelpTip from '../components/HelpTip';

interface HookCommand {
  type: string;
  command: string;
  timeout?: number;
}

interface HookEntry {
  matcher?: string;
  hooks: HookCommand[];
}

interface HooksConfig {
  event: string;
  entries: HookEntry[];
}

const EVENT_TYPES = [
  { id: 'PreToolUse', label: 'PreToolUse', desc: '工具调用前触发', icon: '⏮' },
  { id: 'PostToolUse', label: 'PostToolUse', desc: '工具调用后触发', icon: '⏭' },
  { id: 'SessionStart', label: 'SessionStart', desc: '会话开始时触发', icon: '▶️' },
  { id: 'Stop', label: 'Stop', desc: 'Claude 准备停止时触发', icon: '⏹' },
  { id: 'UserPromptSubmit', label: 'UserPromptSubmit', desc: '用户提交提示时触发', icon: '💬' },
];

export default function HooksManager() {
  const [hooks, setHooks] = useState<HooksConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [expandedEvent, setExpandedEvent] = useState<string | null>(null);

  useEffect(() => {
    GetGlobalHooks()
      .then((data: any) => setHooks(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await SaveGlobalHooks(hooks as any);
      setHasChanges(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const getHooksByEvent = (event: string): HooksConfig | undefined => {
    return hooks.find(h => h.event === event);
  };

  const updateHooks = (newHooks: HooksConfig[]) => {
    setHooks(newHooks);
    setHasChanges(true);
    setSaved(false);
  };

  const addEntry = (event: string) => {
    const existing = hooks.find(h => h.event === event);
    const newEntry: HookEntry = {
      hooks: [{ type: 'command', command: '' }],
    };
    if (event === 'PreToolUse' || event === 'PostToolUse') {
      newEntry.matcher = '';
    }

    if (existing) {
      updateHooks(hooks.map(h =>
        h.event === event ? { ...h, entries: [...h.entries, newEntry] } : h
      ));
    } else {
      updateHooks([...hooks, { event, entries: [newEntry] }]);
    }
    setExpandedEvent(event);
  };

  const updateEntry = (event: string, entryIdx: number, entry: HookEntry) => {
    updateHooks(hooks.map(h =>
      h.event === event ? {
        ...h,
        entries: h.entries.map((e, i) => i === entryIdx ? entry : e),
      } : h
    ));
  };

  const removeEntry = (event: string, entryIdx: number) => {
    updateHooks(hooks.map(h =>
      h.event === event ? {
        ...h,
        entries: h.entries.filter((_, i) => i !== entryIdx),
      } : h
    ).filter(h => h.entries.length > 0));
  };

  const addHookCommand = (event: string, entryIdx: number) => {
    const config = hooks.find(h => h.event === event);
    if (!config) return;
    const entry = config.entries[entryIdx];
    const newHooks = [...entry.hooks, { type: 'command', command: '' }];
    updateEntry(event, entryIdx, { ...entry, hooks: newHooks });
  };

  const updateHookCommand = (event: string, entryIdx: number, hookIdx: number, cmd: HookCommand) => {
    const config = hooks.find(h => h.event === event);
    if (!config) return;
    const entry = config.entries[entryIdx];
    const newHooks = entry.hooks.map((h, i) => i === hookIdx ? cmd : h);
    updateEntry(event, entryIdx, { ...entry, hooks: newHooks });
  };

  const removeHookCommand = (event: string, entryIdx: number, hookIdx: number) => {
    const config = hooks.find(h => h.event === event);
    if (!config) return;
    const entry = config.entries[entryIdx];
    const newHooks = entry.hooks.filter((_, i) => i !== hookIdx);
    if (newHooks.length === 0) {
      removeEntry(event, entryIdx);
    } else {
      updateEntry(event, entryIdx, { ...entry, hooks: newHooks });
    }
  };

  if (loading) {
    return <div className="page-container"><div className="loading-state">加载中...</div></div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">Hooks 管理</h1>
          <p className="page-subtitle">~/.claude/settings.json → hooks - 配置生命周期钩子</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="hooks" />
          {saved && <span className="save-indicator">已保存</span>}
          {hasChanges && <span className="unsaved-indicator">未保存</span>}
          <button className="btn btn-primary" onClick={handleSave}
            disabled={!hasChanges || saving}>
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="hooks-list">
        {EVENT_TYPES.map(evt => {
          const config = getHooksByEvent(evt.id);
          const isExpanded = expandedEvent === evt.id;
          const entryCount = config?.entries.length || 0;

          return (
            <div key={evt.id} className="hook-event-card">
              <button
                className="hook-event-header"
                onClick={() => setExpandedEvent(isExpanded ? null : evt.id)}
              >
                <span className="hook-event-icon">{evt.icon}</span>
                <div className="hook-event-info">
                  <span className="hook-event-name">{evt.label}</span>
                  <span className="hook-event-desc">{evt.desc}</span>
                </div>
                <span className="hook-event-count">
                  {entryCount > 0 ? `${entryCount} 条规则` : '无规则'}
                </span>
                <span className="hook-event-arrow">{isExpanded ? '▼' : '▶'}</span>
              </button>

              {isExpanded && (
                <div className="hook-event-body">
                  {config?.entries.map((entry, eIdx) => (
                    <div key={eIdx} className="hook-entry">
                      {(evt.id === 'PreToolUse' || evt.id === 'PostToolUse') && (
                        <div className="hook-matcher">
                          <label className="hook-label">匹配器 (matcher)</label>
                          <input
                            className="form-input"
                            value={entry.matcher || ''}
                            onChange={e => updateEntry(evt.id, eIdx, { ...entry, matcher: e.target.value })}
                            placeholder="工具名称，如 Bash, Edit, Write"
                          />
                        </div>
                      )}
                      {entry.hooks.map((hook, hIdx) => (
                        <div key={hIdx} className="hook-command-row">
                          <input
                            className="form-input hook-command-input"
                            value={hook.command}
                            onChange={e => updateHookCommand(evt.id, eIdx, hIdx, { ...hook, command: e.target.value })}
                            placeholder="要执行的命令..."
                          />
                          <input
                            className="form-input hook-timeout-input"
                            type="number"
                            value={hook.timeout || ''}
                            onChange={e => updateHookCommand(evt.id, eIdx, hIdx, {
                              ...hook,
                              timeout: e.target.value ? parseInt(e.target.value) : undefined,
                            })}
                            placeholder="超时(s)"
                          />
                          <button className="btn btn-icon btn-danger"
                            onClick={() => removeHookCommand(evt.id, eIdx, hIdx)}>×</button>
                        </div>
                      ))}
                      <div className="hook-entry-actions">
                        <button className="btn btn-ghost" onClick={() => addHookCommand(evt.id, eIdx)}>
                          + 添加命令
                        </button>
                        <button className="btn btn-danger" onClick={() => removeEntry(evt.id, eIdx)}>
                          删除规则
                        </button>
                      </div>
                    </div>
                  ))}
                  <button className="btn btn-secondary hook-add-btn" onClick={() => addEntry(evt.id)}>
                    + 添加规则
                  </button>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
