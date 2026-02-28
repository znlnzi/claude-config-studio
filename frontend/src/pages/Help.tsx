import { useState } from 'react';
import { guideSections } from '../helpData';

export default function Help() {
  const [activeSection, setActiveSection] = useState('overview');

  const current = guideSections.find(s => s.id === activeSection) || guideSections[0];

  return (
    <div className="page-container page-full">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">帮助文档</h1>
          <p className="page-subtitle">ClaudeCode Config Studio 操作指南</p>
        </div>
      </div>

      <div className="help-layout">
        {/* 左侧目录 */}
        <div className="help-toc">
          {guideSections.map(section => (
            <button
              key={section.id}
              className={`help-toc-item ${activeSection === section.id ? 'active' : ''}`}
              onClick={() => setActiveSection(section.id)}
            >
              <span className="help-toc-icon">{section.icon}</span>
              <span className="help-toc-label">{section.title}</span>
            </button>
          ))}
        </div>

        {/* 右侧内容 */}
        <div className="help-content">
          <div className="help-content-header">
            <span className="help-content-icon">{current.icon}</span>
            <h2>{current.title}</h2>
          </div>
          <div className="help-content-body">
            {current.content.split('\n').map((line, i) => {
              const trimmed = line.trimStart();
              if (trimmed.startsWith('•')) {
                return (
                  <div key={i} className="help-bullet">
                    <span className="help-bullet-dot">•</span>
                    <span>{trimmed.slice(1).trim()}</span>
                  </div>
                );
              }
              if (/^\d+\./.test(trimmed)) {
                return (
                  <div key={i} className="help-numbered">
                    {trimmed}
                  </div>
                );
              }
              if (trimmed === '') {
                return <div key={i} className="help-spacer" />;
              }
              if (trimmed.startsWith('---')) {
                return <div key={i} className="help-code-line">{trimmed}</div>;
              }
              // 检测标题样式的行
              if (trimmed.endsWith('：') || trimmed.endsWith(':')) {
                return <div key={i} className="help-section-label">{trimmed}</div>;
              }
              return <div key={i} className="help-text">{trimmed}</div>;
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
