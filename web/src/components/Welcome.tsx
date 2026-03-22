interface WelcomeProps {
  onNew: () => void
  onResume: () => void
}

const ASCII_LOGO = `
 ____  _ _       _
|  _ \\(_) | ___ | |_
| |_) | | |/ _ \\| __|
|  __/| | | (_) | |_
|_|   |_|_|\\___/ \\__|
`

export default function Welcome({ onNew, onResume }: WelcomeProps) {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100vh',
      padding: '1rem',
      background: 'var(--bg)',
      color: 'var(--fg)',
    }}>
      <pre style={{
        color: 'var(--accent)',
        fontFamily: 'monospace',
        fontSize: '1rem',
        lineHeight: 1.2,
        textAlign: 'center',
      }}>
        {ASCII_LOGO}
      </pre>
      <p style={{
        color: 'var(--muted)',
        fontFamily: 'monospace',
        fontSize: '0.85rem',
        marginBottom: '2rem',
      }}>
        Remote control for Claude Code
      </p>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '0.75rem',
        width: '100%',
        maxWidth: '280px',
      }}>
        <div
          onClick={onNew}
          style={{
            padding: '0.75rem 1rem',
            fontFamily: 'monospace',
            fontSize: '1rem',
            color: 'var(--accent)',
            cursor: 'pointer',
            borderRadius: '4px',
            background: 'var(--surface)',
          }}
        >
          &gt; New Session
        </div>
        <div
          onClick={onResume}
          style={{
            padding: '0.75rem 1rem',
            fontFamily: 'monospace',
            fontSize: '1rem',
            color: 'var(--accent)',
            cursor: 'pointer',
            borderRadius: '4px',
            background: 'var(--surface)',
          }}
        >
          &gt; Resume Session
        </div>
      </div>
    </div>
  )
}
