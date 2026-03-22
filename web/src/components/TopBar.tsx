interface TopBarProps {
  title: string
  onBack: () => void
  view: 'conversation' | 'dashboard'
  onToggleView: () => void
}

export default function TopBar({ title, onBack, view, onToggleView }: TopBarProps) {
  const truncated = title.length > 24 ? title.slice(0, 24) + '...' : title

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '0.6rem 0.75rem',
      borderBottom: '1px solid var(--surface)',
      background: 'var(--bg)',
      fontFamily: 'monospace',
      flexShrink: 0,
    }}>
      <button
        onClick={onBack}
        style={{
          background: 'none',
          border: 'none',
          color: 'var(--accent)',
          fontFamily: 'monospace',
          fontSize: '1rem',
          cursor: 'pointer',
          padding: '0.25rem',
        }}
      >
        &larr;
      </button>

      <span style={{
        color: 'var(--fg)',
        fontSize: '0.85rem',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
        flex: 1,
        textAlign: 'center',
        padding: '0 0.5rem',
      }}>
        {truncated}
      </span>

      <button
        onClick={onToggleView}
        style={{
          background: 'none',
          border: 'none',
          fontSize: '1.1rem',
          cursor: 'pointer',
          padding: '0.25rem',
        }}
      >
        {view === 'conversation' ? '\u{1F4CA}' : '\u{1F4AC}'}
      </button>
    </div>
  )
}
