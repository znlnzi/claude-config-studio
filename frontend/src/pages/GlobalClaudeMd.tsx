import { useEffect, useState, useCallback } from 'react';
import Editor from '@monaco-editor/react';
import {
  GetGlobalConfig,
  SaveGlobalClaudeMd,
} from '../../wailsjs/go/services/ConfigService';
import HelpTip from '../components/HelpTip';

export default function GlobalClaudeMd() {
  const [content, setContent] = useState('');
  const [originalContent, setOriginalContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    GetGlobalConfig()
      .then((config: any) => {
        const md = config.claudeMd || '';
        setContent(md);
        setOriginalContent(md);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleChange = useCallback((value: string | undefined) => {
    const newValue = value ?? '';
    setContent(newValue);
    setHasChanges(newValue !== originalContent);
    setSaved(false);
  }, [originalContent]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await SaveGlobalClaudeMd(content);
      setOriginalContent(content);
      setHasChanges(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('保存失败:', err);
    } finally {
      setSaving(false);
    }
  };

  // Cmd+S 快捷键保存
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        if (hasChanges) handleSave();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [hasChanges, content]);

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading-state">加载中...</div>
      </div>
    );
  }

  return (
    <div className="page-container page-full">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">全局指令文件</h1>
          <p className="page-subtitle">~/.claude/CLAUDE.md - 定义 Claude 的全局行为规范</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="global-claudemd" />
          {saved && <span className="save-indicator">已保存</span>}
          {hasChanges && <span className="unsaved-indicator">未保存</span>}
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={!hasChanges || saving}
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="editor-container">
        <Editor
          height="100%"
          defaultLanguage="markdown"
          value={content}
          onChange={handleChange}
          theme="vs"
          options={{
            fontSize: 13,
            lineHeight: 20,
            fontFamily: "'SF Mono', 'Monaco', 'Menlo', 'Courier New', monospace",
            minimap: { enabled: false },
            wordWrap: 'on',
            lineNumbers: 'on',
            renderLineHighlight: 'line',
            scrollBeyondLastLine: false,
            padding: { top: 12, bottom: 12 },
            automaticLayout: true,
            tabSize: 2,
            smoothScrolling: true,
            cursorSmoothCaretAnimation: 'on',
          }}
        />
      </div>
    </div>
  );
}
