import { useState, useRef, useEffect } from 'react';
import { helpData } from '../helpData';

interface HelpTipProps {
  pageId: string;
}

export default function HelpTip({ pageId }: HelpTipProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const data = helpData[pageId];
  if (!data) return null;

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  return (
    <div className="help-tip-wrapper" ref={ref}>
      <button
        className="btn btn-help"
        onClick={() => setOpen(!open)}
        title="操作说明"
      >
        ?
      </button>
      {open && (
        <div className="help-tip-popover">
          <div className="help-tip-header">
            <h3>{data.title}</h3>
            <button className="help-tip-close" onClick={() => setOpen(false)}>×</button>
          </div>
          <p className="help-tip-summary">{data.summary}</p>
          <ul className="help-tip-list">
            {data.tips.map((tip, i) => (
              <li key={i}>{tip}</li>
            ))}
          </ul>
          {data.shortcuts && data.shortcuts.length > 0 && (
            <div className="help-tip-shortcuts">
              <div className="help-tip-shortcuts-title">快捷键</div>
              {data.shortcuts.map((s, i) => (
                <div key={i} className="help-tip-shortcut-row">
                  <kbd>{s.key}</kbd>
                  <span>{s.desc}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
