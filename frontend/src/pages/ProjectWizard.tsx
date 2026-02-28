import { useEffect, useState } from 'react';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import { GetTemplates, ApplyTemplate } from '../../wailsjs/go/services/TemplateService';
import { InitProjectConfig } from '../../wailsjs/go/services/ProjectService';

interface Template {
  id: string;
  name: string;
  category: string;
  description: string;
  tags: string[];
  claudeMd?: string;
}

interface TemplateCategory {
  id: string;
  name: string;
  icon: string;
  templates: Template[];
}

interface WizardProps {
  onComplete: (path: string) => void;
  onCancel: () => void;
}

type Step = 'directory' | 'template' | 'confirm';

export default function ProjectWizard({ onComplete, onCancel }: WizardProps) {
  const [step, setStep] = useState<Step>('directory');
  const [projectPath, setProjectPath] = useState('');
  const [categories, setCategories] = useState<TemplateCategory[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | null>(null);
  const [applying, setApplying] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    GetTemplates()
      .then((data: any) => setCategories(data || []))
      .catch(console.error);
  }, []);

  const handleSelectDir = async () => {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        setProjectPath(dir);
        setStep('template');
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleApply = async () => {
    setApplying(true);
    setError('');
    try {
      await InitProjectConfig(projectPath);
      if (selectedTemplateId) {
        await ApplyTemplate(projectPath, selectedTemplateId, false);
      }
      onComplete(projectPath);
    } catch (err: any) {
      setError(err?.message || '应用失败');
    } finally {
      setApplying(false);
    }
  };

  const selectedTemplate = selectedTemplateId
    ? categories.flatMap(c => c.templates).find(t => t.id === selectedTemplateId)
    : null;

  const projectName = projectPath.split('/').pop() || '';

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <button className="btn btn-ghost" onClick={onCancel}>← 返回</button>
          <h1 className="page-title">新建项目配置</h1>
        </div>
      </div>

      {/* 步骤指示器 */}
      <div className="wizard-steps">
        <WizardStep num={1} label="选择目录" active={step === 'directory'} done={step !== 'directory'} />
        <div className="wizard-step-line" />
        <WizardStep num={2} label="选择模板" active={step === 'template'} done={step === 'confirm'} />
        <div className="wizard-step-line" />
        <WizardStep num={3} label="确认创建" active={step === 'confirm'} done={false} />
      </div>

      {/* 步骤内容 */}
      {step === 'directory' && (
        <div className="wizard-content">
          <div className="empty-state">
            <div className="empty-icon">📂</div>
            <h3>选择项目目录</h3>
            <p>选择需要添加 Claude Code 配置的项目根目录</p>
            <button className="btn btn-primary" onClick={handleSelectDir}>
              选择目录
            </button>
          </div>
        </div>
      )}

      {step === 'template' && (
        <div className="wizard-content">
          <div className="wizard-selected-dir">
            <span className="wizard-dir-icon">📂</span>
            <span className="wizard-dir-path">{projectPath}</span>
            <button className="btn btn-ghost" onClick={() => setStep('directory')}>更换</button>
          </div>

          <h3 className="wizard-section-title">选择配置模板（可选）</h3>

          <button
            className={`template-card wizard-template-option ${!selectedTemplateId ? 'selected' : ''}`}
            onClick={() => setSelectedTemplateId(null)}
          >
            <div className="template-card-name">空白配置</div>
            <div className="template-card-desc">只创建 .claude/ 目录结构，不添加预设内容</div>
          </button>

          {categories.map(cat => (
            <div key={cat.id}>
              <h4 className="wizard-category-title">{cat.icon} {cat.name}</h4>
              <div className="template-grid">
                {cat.templates.map(tmpl => (
                  <button
                    key={tmpl.id}
                    className={`template-card wizard-template-option ${selectedTemplateId === tmpl.id ? 'selected' : ''}`}
                    onClick={() => setSelectedTemplateId(tmpl.id)}
                  >
                    <div className="template-card-name">{tmpl.name}</div>
                    <div className="template-card-desc">{tmpl.description}</div>
                    <div className="template-card-tags">
                      {tmpl.tags.map(tag => <span key={tag} className="badge">{tag}</span>)}
                    </div>
                  </button>
                ))}
              </div>
            </div>
          ))}

          <div className="wizard-actions">
            <button className="btn btn-secondary" onClick={() => setStep('directory')}>上一步</button>
            <button className="btn btn-primary" onClick={() => setStep('confirm')}>下一步</button>
          </div>
        </div>
      )}

      {step === 'confirm' && (
        <div className="wizard-content">
          {error && <div className="json-error">{error}</div>}

          <div className="wizard-confirm-card">
            <h3>确认配置</h3>
            <div className="wizard-confirm-row">
              <span className="wizard-confirm-label">项目目录</span>
              <span className="wizard-confirm-value">{projectPath}</span>
            </div>
            <div className="wizard-confirm-row">
              <span className="wizard-confirm-label">项目名称</span>
              <span className="wizard-confirm-value">{projectName}</span>
            </div>
            <div className="wizard-confirm-row">
              <span className="wizard-confirm-label">配置模板</span>
              <span className="wizard-confirm-value">{selectedTemplate?.name || '空白配置'}</span>
            </div>
            <div className="wizard-confirm-row">
              <span className="wizard-confirm-label">将创建</span>
              <span className="wizard-confirm-value">
                {projectPath}/.claude/ 目录
                {selectedTemplate?.claudeMd && '、CLAUDE.md'}
              </span>
            </div>
          </div>

          {selectedTemplate?.claudeMd && (
            <div className="template-content-section">
              <h4 className="template-content-title">CLAUDE.md 预览</h4>
              <pre className="template-code">{selectedTemplate.claudeMd}</pre>
            </div>
          )}

          <div className="wizard-actions">
            <button className="btn btn-secondary" onClick={() => setStep('template')}>上一步</button>
            <button className="btn btn-primary" onClick={handleApply} disabled={applying}>
              {applying ? '创建中...' : '确认创建'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function WizardStep({ num, label, active, done }: {
  num: number; label: string; active: boolean; done: boolean;
}) {
  return (
    <div className={`wizard-step ${active ? 'active' : ''} ${done ? 'done' : ''}`}>
      <div className="wizard-step-num">{done ? '✓' : num}</div>
      <span className="wizard-step-label">{label}</span>
    </div>
  );
}
