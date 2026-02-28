interface PlaceholderProps {
  title: string;
  icon: string;
  description: string;
}

export default function Placeholder({ title, icon, description }: PlaceholderProps) {
  return (
    <div className="page-container">
      <div className="empty-state">
        <div className="empty-icon" style={{ fontSize: 48 }}>{icon}</div>
        <h2>{title}</h2>
        <p>{description}</p>
        <span className="badge">即将推出</span>
      </div>
    </div>
  );
}
