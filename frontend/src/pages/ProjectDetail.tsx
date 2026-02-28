import { useEffect, useState, useCallback, useRef } from 'react';
import Editor from '@monaco-editor/react';
import {
  GetProjectConfig,
  SaveProjectClaudeMd,
  SaveProjectSettings,
  SaveProjectMcp,
} from '../../wailsjs/go/services/ConfigService';
import { InitProjectConfig } from '../../wailsjs/go/services/ProjectService';
import {
  ListExtensions,
  GetExtension,
  SaveExtension,
  DeleteExtension,
} from '../../wailsjs/go/services/ExtensionService';
import {
  ListUserSkills,
  GetUserSkill,
  SaveUserSkill,
  DeleteUserSkill,
} from '../../wailsjs/go/services/SkillService';

interface ProjectDetailProps {
  projectPath: string;
  onBack: () => void;
}

interface ExtFile {
  name: string;
  fileName: string;
  path: string;
  content: string;
}

interface UserSkillItem {
  name: string;
  description: string;
  scope: string;
  dirName: string;
  isFlat: boolean;
}

type Tab = 'claudemd' | 'settings' | 'mcp' | 'rules' | 'agents' | 'skills';

export default function ProjectDetail({ projectPath, onBack }: ProjectDetailProps) {
  const [activeTab, setActiveTab] = useState<Tab>('claudemd');
  const [claudeMd, setClaudeMd] = useState('');
  const [settingsJson, setSettingsJson] = useState('{}');
  const [mcpJson, setMcpJson] = useState('{}');
  const [originalClaudeMd, setOriginalClaudeMd] = useState('');
  const [originalSettings, setOriginalSettings] = useState('{}');
  const [originalMcp, setOriginalMcp] = useState('{}');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const activeTabRef = useRef<Tab>(activeTab);

  // === 文件列表 tab 状态 (rules / agents) ===
  const [extFiles, setExtFiles] = useState<ExtFile[]>([]);
  const [selectedExtFile, setSelectedExtFile] = useState<ExtFile | null>(null);
  const [extContent, setExtContent] = useState('');
  const [originalExtContent, setOriginalExtContent] = useState('');
  const [extCreating, setExtCreating] = useState(false);
  const [extNewName, setExtNewName] = useState('');

  // === Skills tab 状态 ===
  const [skillFiles, setSkillFiles] = useState<UserSkillItem[]>([]);
  const [selectedSkill, setSelectedSkill] = useState<UserSkillItem | null>(null);
  const [skillContent, setSkillContent] = useState('');
  const [originalSkillContent, setOriginalSkillContent] = useState('');
  const [skillName, setSkillName] = useState('');
  const [skillDesc, setSkillDesc] = useState('');
  const [skillCreating, setSkillCreating] = useState(false);
  const [skillNewDirName, setSkillNewDirName] = useState('');

  const projectName = projectPath.split('/').pop() || projectPath;

  // 加载 rules 或 agents 文件列表
  const loadExtFiles = useCallback(async (type: string) => {
    try {
      const data: any = await ListExtensions(type, projectPath);
      setExtFiles(data || []);
    } catch { setExtFiles([]); }
  }, [projectPath]);

  // 加载 skills 列表
  const loadSkills = useCallback(async () => {
    try {
      const data: any = await ListUserSkills(projectPath);
      setSkillFiles(data || []);
    } catch { setSkillFiles([]); }
  }, [projectPath]);

  useEffect(() => {
    InitProjectConfig(projectPath).catch(() => {});
    GetProjectConfig(projectPath)
      .then((config: any) => {
        const md = config.claudeMd || '';
        // json.RawMessage 从 Go 传来可能是对象，需确保为字符串
        const settingsRaw = config.settings;
        const mcpRaw = config.mcpConfig;
        const settingsStr = typeof settingsRaw === 'object' && settingsRaw !== null
          ? JSON.stringify(settingsRaw, null, 2)
          : (settingsRaw || '{}');
        const mcpStr = typeof mcpRaw === 'object' && mcpRaw !== null
          ? JSON.stringify(mcpRaw, null, 2)
          : (mcpRaw || '{}');
        setClaudeMd(md);
        setOriginalClaudeMd(md);
        setSettingsJson(settingsStr);
        setOriginalSettings(settingsStr);
        setMcpJson(mcpStr);
        setOriginalMcp(mcpStr);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [projectPath]);

  const hasChanges = useCallback(() => {
    switch (activeTab) {
      case 'claudemd': return claudeMd !== originalClaudeMd;
      case 'settings': return settingsJson !== originalSettings;
      case 'mcp': return mcpJson !== originalMcp;
      case 'rules':
      case 'agents': return extContent !== originalExtContent;
      case 'skills': return skillContent !== originalSkillContent;
      default: return false;
    }
  }, [activeTab, claudeMd, originalClaudeMd, settingsJson, originalSettings, mcpJson, originalMcp, extContent, originalExtContent, skillContent, originalSkillContent]);

  const handleSave = async () => {
    setSaving(true);
    try {
      switch (activeTab) {
        case 'claudemd':
          await SaveProjectClaudeMd(projectPath, claudeMd);
          setOriginalClaudeMd(claudeMd);
          break;
        case 'settings':
          await SaveProjectSettings(projectPath, settingsJson);
          setOriginalSettings(settingsJson);
          break;
        case 'mcp':
          await SaveProjectMcp(projectPath, mcpJson);
          setOriginalMcp(mcpJson);
          break;
        case 'rules':
        case 'agents': {
          const extType = activeTab;
          if (extCreating && extNewName.trim()) {
            const fn = extNewName.endsWith('.md') ? extNewName : extNewName + '.md';
            await SaveExtension(extType, projectPath, fn, extContent);
            setExtCreating(false);
            setExtNewName('');
            loadExtFiles(extType);
          } else if (selectedExtFile) {
            await SaveExtension(extType, projectPath, selectedExtFile.fileName, extContent);
          }
          setOriginalExtContent(extContent);
          break;
        }
        case 'skills': {
          if (skillCreating && skillNewDirName.trim()) {
            await SaveUserSkill(projectPath, skillNewDirName.trim(), skillName.trim(), skillDesc.trim(), skillContent, 'false');
            setSkillCreating(false);
            setSkillNewDirName('');
            loadSkills();
          } else if (selectedSkill) {
            await SaveUserSkill(projectPath, selectedSkill.dirName, skillName.trim(), skillDesc.trim(), skillContent, selectedSkill.isFlat ? 'true' : 'false');
          }
          setOriginalSkillContent(skillContent);
          break;
        }
      }
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('保存失败:', err);
    } finally {
      setSaving(false);
    }
  };

  useEffect(() => { activeTabRef.current = activeTab; }, [activeTab]);

  const handleTabChange = (tab: Tab) => {
    setActiveTab(tab);
    if (tab === 'rules' || tab === 'agents') {
      loadExtFiles(tab);
      setSelectedExtFile(null);
      setExtContent('');
      setOriginalExtContent('');
      setExtCreating(false);
    }
    if (tab === 'skills') {
      loadSkills();
      setSelectedSkill(null);
      setSkillContent('');
      setOriginalSkillContent('');
      setSkillCreating(false);
    }
  };

  const handleSelectExtFile = async (file: ExtFile, type: string) => {
    setSelectedExtFile(file);
    setExtCreating(false);
    try {
      const detail: any = await GetExtension(type, projectPath, file.fileName);
      setExtContent(detail.content || '');
      setOriginalExtContent(detail.content || '');
    } catch { setExtContent(''); }
  };

  const handleDeleteExtFile = async (file: ExtFile, type: string) => {
    try {
      await DeleteExtension(type, projectPath, file.fileName);
      if (selectedExtFile?.fileName === file.fileName) {
        setSelectedExtFile(null);
        setExtContent('');
      }
      loadExtFiles(type);
    } catch (err) { console.error(err); }
  };

  const handleSelectSkill = async (skill: UserSkillItem) => {
    setSelectedSkill(skill);
    setSkillCreating(false);
    try {
      const detail: any = await GetUserSkill(projectPath, skill.dirName);
      setSkillName(detail.name || '');
      setSkillDesc(detail.description || '');
      setSkillContent(detail.content || '');
      setOriginalSkillContent(detail.content || '');
    } catch { setSkillContent(''); }
  };

  const handleDeleteSkill = async (skill: UserSkillItem) => {
    try {
      await DeleteUserSkill(projectPath, skill.dirName, skill.isFlat ? 'true' : 'false');
      if (selectedSkill?.dirName === skill.dirName) {
        setSelectedSkill(null);
        setSkillContent('');
      }
      loadSkills();
    } catch (err) { console.error(err); }
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        if (hasChanges()) handleSave();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [hasChanges]);

  const editorValue = activeTab === 'claudemd' ? claudeMd : activeTab === 'settings' ? settingsJson : mcpJson;
  const editorLanguage = activeTab === 'claudemd' ? 'markdown' : 'json';

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
          <button className="btn btn-ghost" onClick={onBack}>← 返回</button>
          <div>
            <h1 className="page-title">{projectName}</h1>
            <p className="page-subtitle">{projectPath}</p>
          </div>
        </div>
        <div className="page-header-right">
          {saved && <span className="save-indicator">已保存</span>}
          {hasChanges() && <span className="unsaved-indicator">未保存</span>}
          {(activeTab === 'claudemd' || activeTab === 'settings' || activeTab === 'mcp') && (
            <button
              className="btn btn-primary"
              onClick={handleSave}
              disabled={!hasChanges() || saving}
            >
              {saving ? '保存中...' : '保存'}
            </button>
          )}
        </div>
      </div>

      <div className="tab-bar">
        <button className={`tab-item ${activeTab === 'claudemd' ? 'active' : ''}`}
          onClick={() => handleTabChange('claudemd')}>CLAUDE.md</button>
        <button className={`tab-item ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => handleTabChange('settings')}>Settings</button>
        <button className={`tab-item ${activeTab === 'mcp' ? 'active' : ''}`}
          onClick={() => handleTabChange('mcp')}>MCP</button>
        <button className={`tab-item ${activeTab === 'rules' ? 'active' : ''}`}
          onClick={() => handleTabChange('rules')}>Rules</button>
        <button className={`tab-item ${activeTab === 'agents' ? 'active' : ''}`}
          onClick={() => handleTabChange('agents')}>Agents</button>
        <button className={`tab-item ${activeTab === 'skills' ? 'active' : ''}`}
          onClick={() => handleTabChange('skills')}>Skills</button>
      </div>

      {/* 原始编辑器 tabs: claudemd / settings / mcp */}
      {(activeTab === 'claudemd' || activeTab === 'settings' || activeTab === 'mcp') && (
        <div className="editor-container">
          <Editor
            height="100%"
            language={editorLanguage}
            value={editorValue}
            theme="vs"
            onChange={v => {
              const tab = activeTabRef.current;
              if (tab === 'claudemd') setClaudeMd(v ?? '');
              else if (tab === 'settings') setSettingsJson(v ?? '{}');
              else if (tab === 'mcp') setMcpJson(v ?? '{}');
            }}
            options={{
              fontSize: 13,
              lineHeight: 20,
              fontFamily: "'SF Mono', 'Monaco', 'Menlo', monospace",
              minimap: { enabled: false },
              wordWrap: activeTab === 'claudemd' ? 'on' : 'off',
              scrollBeyondLastLine: false,
              padding: { top: 12, bottom: 12 },
              automaticLayout: true,
              tabSize: 2,
              smoothScrolling: true,
            }}
          />
        </div>
      )}

      {/* Rules / Agents tab: 文件列表 + 编辑器 */}
      {(activeTab === 'rules' || activeTab === 'agents') && (
        <div className="ext-layout">
          <div className="ext-sidebar">
            <div className="ext-file-list">
              <button className="btn btn-ghost btn-sm" style={{ width: '100%', marginBottom: 8 }}
                onClick={() => { setExtCreating(true); setSelectedExtFile(null); setExtNewName(''); setExtContent(''); setOriginalExtContent(''); }}>
                + 新建
              </button>
              {extFiles.length === 0 && !extCreating && (
                <div className="ext-empty"><span>{activeTab === 'rules' ? '📜' : '🤖'}</span><span>暂无文件</span></div>
              )}
              {extFiles.map(file => (
                <button key={file.fileName}
                  className={`ext-file-item ${selectedExtFile?.fileName === file.fileName && !extCreating ? 'active' : ''}`}
                  onClick={() => handleSelectExtFile(file, activeTab)}>
                  <div className="ext-file-name">{file.name}</div>
                  <button className="ext-file-delete"
                    onClick={e => { e.stopPropagation(); handleDeleteExtFile(file, activeTab); }} title="删除">×</button>
                </button>
              ))}
            </div>
          </div>
          <div className="ext-editor">
            {extCreating && (
              <div className="ext-new-bar">
                <input className="form-input" value={extNewName}
                  onChange={e => setExtNewName(e.target.value)}
                  placeholder="输入文件名（不含 .md 后缀）" autoFocus />
                <button className="btn btn-primary" onClick={handleSave}
                  disabled={!extNewName.trim() || saving}>
                  {saving ? '保存中...' : '创建'}
                </button>
                <button className="btn btn-ghost" onClick={() => setExtCreating(false)}>取消</button>
              </div>
            )}
            {(selectedExtFile || extCreating) ? (
              <div className="ext-editor-content">
                {!extCreating && selectedExtFile && (
                  <div className="ext-editor-header">
                    <span className="ext-editor-filename">{selectedExtFile.fileName}</span>
                    <button className="btn btn-primary" onClick={handleSave}
                      disabled={!hasChanges() || saving}>
                      {saving ? '保存中...' : '保存'}
                    </button>
                  </div>
                )}
                <div className="ext-monaco">
                  <Editor height="100%" defaultLanguage="markdown" value={extContent}
                    onChange={v => setExtContent(v ?? '')} theme="vs"
                    options={{ fontSize: 13, lineHeight: 20, fontFamily: "'SF Mono', 'Monaco', 'Menlo', monospace",
                      minimap: { enabled: false }, wordWrap: 'on', scrollBeyondLastLine: false,
                      padding: { top: 12, bottom: 12 }, automaticLayout: true, tabSize: 2, smoothScrolling: true }} />
                </div>
              </div>
            ) : (
              <div className="ext-empty-editor">
                <span style={{ fontSize: 36 }}>{activeTab === 'rules' ? '📜' : '🤖'}</span>
                <p>选择一个文件编辑，或创建新文件</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Skills tab: 目录列表 + 编辑器 */}
      {activeTab === 'skills' && (
        <div className="ext-layout">
          <div className="ext-sidebar">
            <div className="ext-file-list">
              <button className="btn btn-ghost btn-sm" style={{ width: '100%', marginBottom: 8 }}
                onClick={() => { setSkillCreating(true); setSelectedSkill(null); setSkillNewDirName(''); setSkillName(''); setSkillDesc(''); setSkillContent(''); setOriginalSkillContent(''); }}>
                + 新建
              </button>
              {skillFiles.length === 0 && !skillCreating && (
                <div className="ext-empty"><span>🎯</span><span>暂无 Skill</span></div>
              )}
              {skillFiles.map(skill => (
                <button key={skill.dirName}
                  className={`ext-file-item ${selectedSkill?.dirName === skill.dirName && !skillCreating ? 'active' : ''}`}
                  onClick={() => handleSelectSkill(skill)}>
                  <div className="ext-file-name">🎯 {skill.name || skill.dirName}</div>
                  <button className="ext-file-delete"
                    onClick={e => { e.stopPropagation(); handleDeleteSkill(skill); }} title="删除">×</button>
                </button>
              ))}
            </div>
          </div>
          <div className="ext-editor">
            {skillCreating && (
              <div className="ext-new-bar">
                <input className="form-input" value={skillNewDirName}
                  onChange={e => setSkillNewDirName(e.target.value)}
                  placeholder="Skill 目录名（如 my-rules）" autoFocus />
                <input className="form-input" value={skillName}
                  onChange={e => setSkillName(e.target.value)}
                  placeholder="Skill 名称" />
                <input className="form-input" value={skillDesc}
                  onChange={e => setSkillDesc(e.target.value)}
                  placeholder="描述（何时使用）" />
                <button className="btn btn-primary" onClick={handleSave}
                  disabled={!skillNewDirName.trim() || saving}>
                  {saving ? '保存中...' : '创建'}
                </button>
                <button className="btn btn-ghost" onClick={() => setSkillCreating(false)}>取消</button>
              </div>
            )}
            {(selectedSkill || skillCreating) ? (
              <div className="ext-editor-content">
                {!skillCreating && selectedSkill && (
                  <div className="ext-editor-header">
                    <span className="ext-editor-filename">{selectedSkill.name || selectedSkill.dirName}</span>
                    <span style={{ fontSize: 12, color: 'var(--macos-text-tertiary)' }}>{selectedSkill.description}</span>
                    <button className="btn btn-primary" onClick={handleSave}
                      disabled={!hasChanges() || saving}>
                      {saving ? '保存中...' : '保存'}
                    </button>
                  </div>
                )}
                <div className="ext-monaco">
                  <Editor height="100%" defaultLanguage="markdown" value={skillContent}
                    onChange={v => setSkillContent(v ?? '')} theme="vs"
                    options={{ fontSize: 13, lineHeight: 20, fontFamily: "'SF Mono', 'Monaco', 'Menlo', monospace",
                      minimap: { enabled: false }, wordWrap: 'on', scrollBeyondLastLine: false,
                      padding: { top: 12, bottom: 12 }, automaticLayout: true, tabSize: 2, smoothScrolling: true }} />
                </div>
              </div>
            ) : (
              <div className="ext-empty-editor">
                <span style={{ fontSize: 36 }}>🎯</span>
                <p>选择一个 Skill 编辑，或创建新 Skill</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
