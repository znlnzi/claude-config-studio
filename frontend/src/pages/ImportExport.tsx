import { useState } from 'react';
import {
  ExportGlobalConfig,
  ImportGlobalConfig,
  ExportProjectConfig,
  ImportProjectConfig,
} from '../../wailsjs/go/services/ExportService';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import HelpTip from '../components/HelpTip';

export default function ImportExport() {
  const [status, setStatus] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [processing, setProcessing] = useState(false);

  const showStatus = (type: 'success' | 'error', message: string) => {
    setStatus({ type, message });
    setTimeout(() => setStatus(null), 4000);
  };

  const handleExportGlobal = async () => {
    setProcessing(true);
    try {
      const path = await ExportGlobalConfig();
      if (path) {
        showStatus('success', `全局配置已导出到: ${path}`);
      }
    } catch (err: any) {
      showStatus('error', err?.message || '导出失败');
    } finally {
      setProcessing(false);
    }
  };

  const handleImportGlobal = async () => {
    setProcessing(true);
    try {
      const count = await ImportGlobalConfig();
      if (count > 0) {
        showStatus('success', `成功导入 ${count} 个文件到全局配置`);
      }
    } catch (err: any) {
      showStatus('error', err?.message || '导入失败');
    } finally {
      setProcessing(false);
    }
  };

  const handleExportProject = async () => {
    setProcessing(true);
    try {
      const dir = await SelectDirectory();
      if (!dir) { setProcessing(false); return; }
      const path = await ExportProjectConfig(dir);
      if (path) {
        showStatus('success', `项目配置已导出到: ${path}`);
      }
    } catch (err: any) {
      showStatus('error', err?.message || '导出失败');
    } finally {
      setProcessing(false);
    }
  };

  const handleImportProject = async () => {
    setProcessing(true);
    try {
      const dir = await SelectDirectory();
      if (!dir) { setProcessing(false); return; }
      const count = await ImportProjectConfig(dir);
      if (count > 0) {
        showStatus('success', `成功导入 ${count} 个文件到项目 ${dir}`);
      }
    } catch (err: any) {
      showStatus('error', err?.message || '导入失败');
    } finally {
      setProcessing(false);
    }
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">导入 / 导出</h1>
          <p className="page-subtitle">备份和迁移 Claude Code 配置</p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="import-export" />
        </div>
      </div>

      {status && (
        <div className={status.type === 'success' ? 'status-success' : 'json-error'}>
          {status.message}
        </div>
      )}

      <div className="ie-sections">
        {/* 全局配置 */}
        <div className="ie-section">
          <div className="ie-section-header">
            <div className="ie-section-icon">🌍</div>
            <div className="ie-section-info">
              <h3>全局配置</h3>
              <p>导出/导入 ~/.claude/ 下的全局配置文件</p>
            </div>
          </div>
          <div className="ie-section-detail">
            <div className="ie-includes">
              <span className="ie-include-label">包含：</span>
              <span className="badge">CLAUDE.md</span>
              <span className="badge">settings.json</span>
              <span className="badge">cclsp.json</span>
              <span className="badge">.mcp.json</span>
              <span className="badge">commands/</span>
              <span className="badge">agents/</span>
            </div>
          </div>
          <div className="ie-section-actions">
            <button className="btn btn-primary" onClick={handleExportGlobal} disabled={processing}>
              导出为 ZIP
            </button>
            <button className="btn btn-secondary" onClick={handleImportGlobal} disabled={processing}>
              从 ZIP 导入
            </button>
          </div>
        </div>

        {/* 项目配置 */}
        <div className="ie-section">
          <div className="ie-section-header">
            <div className="ie-section-icon">📁</div>
            <div className="ie-section-info">
              <h3>项目配置</h3>
              <p>导出/导入指定项目的 .claude/ 配置目录</p>
            </div>
          </div>
          <div className="ie-section-detail">
            <div className="ie-includes">
              <span className="ie-include-label">包含：</span>
              <span className="badge">CLAUDE.md</span>
              <span className="badge">.claude/ 全部内容</span>
            </div>
          </div>
          <div className="ie-section-actions">
            <button className="btn btn-primary" onClick={handleExportProject} disabled={processing}>
              选择项目并导出
            </button>
            <button className="btn btn-secondary" onClick={handleImportProject} disabled={processing}>
              选择项目并导入
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
