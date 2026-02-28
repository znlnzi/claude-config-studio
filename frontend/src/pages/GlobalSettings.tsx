import { useEffect, useState } from 'react';
import {
  GetGlobalConfig,
  SaveGlobalSettings,
} from '../../wailsjs/go/services/ConfigService';
import HelpTip from '../components/HelpTip';

interface SettingsData {
  env?: Record<string, string>;
  language?: string;
  alwaysThinkingEnabled?: boolean;
  hooks?: Record<string, unknown>;
  statusLine?: unknown;
  enabledPlugins?: Record<string, unknown>;
  [key: string]: unknown;
}

export default function GlobalSettings() {
  const [settings, setSettings] = useState<SettingsData>({});
  const [rawJson, setRawJson] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [viewMode, setViewMode] = useState<'form' | 'json'>('form');
  const [jsonError, setJsonError] = useState('');

  // 环境变量编辑状态
  const [newEnvKey, setNewEnvKey] = useState('');
  const [newEnvValue, setNewEnvValue] = useState('');

  useEffect(() => {
    GetGlobalConfig()
      .then((config: any) => {
        if (config.settings) {
          try {
            const parsed = JSON.parse(config.settings);
            setSettings(parsed);
            setRawJson(JSON.stringify(parsed, null, 2));
          } catch {
            setSettings({});
            setRawJson('{}');
          }
        } else {
          setSettings({});
          setRawJson('{}');
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const updateSettings = (newSettings: SettingsData) => {
    setSettings(newSettings);
    setRawJson(JSON.stringify(newSettings, null, 2));
    setHasChanges(true);
    setSaved(false);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const dataToSave = viewMode === 'json' ? rawJson : JSON.stringify(settings, null, 2);
      await SaveGlobalSettings(dataToSave);
      setHasChanges(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err: any) {
      console.error('保存失败:', err);
      setJsonError(err?.message || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const handleJsonChange = (value: string) => {
    setRawJson(value);
    setHasChanges(true);
    setSaved(false);
    try {
      const parsed = JSON.parse(value);
      setSettings(parsed);
      setJsonError('');
    } catch {
      setJsonError('JSON 格式错误');
    }
  };

  // 环境变量操作
  const addEnvVar = () => {
    if (!newEnvKey.trim()) return;
    const env = { ...(settings.env || {}), [newEnvKey]: newEnvValue };
    updateSettings({ ...settings, env });
    setNewEnvKey('');
    setNewEnvValue('');
  };

  const removeEnvVar = (key: string) => {
    const env = { ...(settings.env || {}) };
    delete env[key];
    updateSettings({ ...settings, env });
  };

  const updateEnvVar = (key: string, value: string) => {
    const env = { ...(settings.env || {}), [key]: value };
    updateSettings({ ...settings, env });
  };

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
          <h1 className="page-title">全局设置</h1>
          <p className="page-subtitle">~/.claude/settings.json</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="global-settings" />
          <div className="view-toggle">
            <button
              className={`toggle-btn ${viewMode === 'form' ? 'active' : ''}`}
              onClick={() => setViewMode('form')}
            >
              表单
            </button>
            <button
              className={`toggle-btn ${viewMode === 'json' ? 'active' : ''}`}
              onClick={() => setViewMode('json')}
            >
              JSON
            </button>
          </div>
          {saved && <span className="save-indicator">已保存</span>}
          {hasChanges && <span className="unsaved-indicator">未保存</span>}
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={!hasChanges || saving || !!jsonError}
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="settings-content">
        {viewMode === 'form' ? (
          <div className="form-sections">
            {/* 基本设置 */}
            <FormSection title="基本设置">
              <FormField label="界面语言" desc="设置 Claude Code 的回复语言">
                <select
                  className="form-select"
                  value={settings.language || ''}
                  onChange={e => updateSettings({ ...settings, language: e.target.value || undefined })}
                >
                  <option value="">默认 (English)</option>
                  <option value="Chinese">中文</option>
                  <option value="Japanese">日本語</option>
                  <option value="Korean">한국어</option>
                  <option value="Spanish">Español</option>
                  <option value="French">Français</option>
                  <option value="German">Deutsch</option>
                </select>
              </FormField>
              <FormField label="思考模式" desc="启用后 Claude 会展示思考过程">
                <label className="toggle-switch">
                  <input
                    type="checkbox"
                    checked={settings.alwaysThinkingEnabled ?? false}
                    onChange={e => updateSettings({ ...settings, alwaysThinkingEnabled: e.target.checked })}
                  />
                  <span className="toggle-slider" />
                </label>
              </FormField>
            </FormSection>

            {/* 环境变量 */}
            <FormSection title="环境变量" desc="配置 Claude Code 运行时的环境变量">
              <div className="env-list">
                {Object.entries(settings.env || {}).map(([key, value]) => (
                  <div key={key} className="env-item">
                    <span className="env-key">{key}</span>
                    <input
                      className="form-input env-value"
                      value={value}
                      onChange={e => updateEnvVar(key, e.target.value)}
                    />
                    <button
                      className="btn btn-icon btn-danger"
                      onClick={() => removeEnvVar(key)}
                      title="删除"
                    >
                      ×
                    </button>
                  </div>
                ))}
                <div className="env-add">
                  <input
                    className="form-input"
                    placeholder="变量名"
                    value={newEnvKey}
                    onChange={e => setNewEnvKey(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && addEnvVar()}
                  />
                  <input
                    className="form-input"
                    placeholder="变量值"
                    value={newEnvValue}
                    onChange={e => setNewEnvValue(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && addEnvVar()}
                  />
                  <button className="btn btn-secondary" onClick={addEnvVar}>
                    添加
                  </button>
                </div>
              </div>
            </FormSection>
          </div>
        ) : (
          <div className="json-editor-container">
            {jsonError && <div className="json-error">{jsonError}</div>}
            <textarea
              className="json-editor"
              value={rawJson}
              onChange={e => handleJsonChange(e.target.value)}
              spellCheck={false}
            />
          </div>
        )}
      </div>
    </div>
  );
}

function FormSection({ title, desc, children }: {
  title: string;
  desc?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="form-section">
      <div className="form-section-header">
        <h3 className="form-section-title">{title}</h3>
        {desc && <p className="form-section-desc">{desc}</p>}
      </div>
      <div className="form-section-body">{children}</div>
    </div>
  );
}

function FormField({ label, desc, children }: {
  label: string;
  desc?: string;
  children: React.ReactNode;
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
