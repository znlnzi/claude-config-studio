import { useEffect, useState } from 'react';
import {
  GetEnabledPlugins,
  TogglePlugin,
  GetInstalledPlugins,
} from '../../wailsjs/go/services/PluginService';
import HelpTip from '../components/HelpTip';

interface PluginInfo {
  name: string;
  source: string;
  enabled: boolean;
  description?: string;
  version?: string;
}

export default function PluginManager() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [toggling, setToggling] = useState<string | null>(null);

  const loadPlugins = async () => {
    setLoading(true);
    try {
      // 获取已启用插件
      const enabled: any = await GetEnabledPlugins();
      // 获取已安装插件
      const installed: any = await GetInstalledPlugins();

      // 合并列表（去重）
      const map = new Map<string, PluginInfo>();
      for (const p of (installed || [])) {
        const key = p.source ? `${p.name}@${p.source}` : p.name;
        map.set(key, p);
      }
      for (const p of (enabled || [])) {
        const key = p.source ? `${p.name}@${p.source}` : p.name;
        if (!map.has(key)) {
          map.set(key, p);
        } else {
          map.get(key)!.enabled = true;
        }
      }

      setPlugins(Array.from(map.values()).sort((a, b) => {
        if (a.enabled !== b.enabled) return a.enabled ? -1 : 1;
        return a.name.localeCompare(b.name);
      }));
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadPlugins(); }, []);

  const handleToggle = async (plugin: PluginInfo) => {
    const key = plugin.source ? `${plugin.name}@${plugin.source}` : plugin.name;
    setToggling(key);
    try {
      await TogglePlugin(key, !plugin.enabled);
      setPlugins(prev => prev.map(p => {
        const pKey = p.source ? `${p.name}@${p.source}` : p.name;
        return pKey === key ? { ...p, enabled: !p.enabled } : p;
      }));
    } catch (err) {
      console.error(err);
    } finally {
      setToggling(null);
    }
  };

  const enabledCount = plugins.filter(p => p.enabled).length;

  if (loading) {
    return <div className="page-container"><div className="loading-state">加载中...</div></div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">插件管理</h1>
          <p className="page-subtitle">
            已启用 {enabledCount} / {plugins.length} 个插件
          </p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="plugins" />
          <button className="btn btn-secondary" onClick={loadPlugins}>刷新</button>
        </div>
      </div>

      {plugins.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📦</div>
          <h3>暂无插件</h3>
          <p>在 Claude Code 中使用 /plugins 命令安装插件</p>
        </div>
      ) : (
        <div className="plugin-list">
          {plugins.map(plugin => {
            const key = plugin.source ? `${plugin.name}@${plugin.source}` : plugin.name;
            return (
              <div key={key} className="plugin-card">
                <div className="plugin-info">
                  <div className="plugin-name">{plugin.name}</div>
                  {plugin.source && (
                    <div className="plugin-source">{plugin.source}</div>
                  )}
                  {plugin.description && (
                    <div className="plugin-desc">{plugin.description}</div>
                  )}
                </div>
                <label className="toggle-switch">
                  <input
                    type="checkbox"
                    checked={plugin.enabled}
                    onChange={() => handleToggle(plugin)}
                    disabled={toggling === key}
                  />
                  <span className="toggle-slider" />
                </label>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
