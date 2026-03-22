import type { Command } from '../types'

interface SlashCommandsProps {
  commands: Command[]
  onSelect: (template: string) => void
  filter: string
}

export default function SlashCommands({ commands, onSelect, filter }: SlashCommandsProps) {
  const query = filter.slice(1).toLowerCase()
  const filtered = commands.filter((cmd) =>
    cmd.name.toLowerCase().includes(query)
  )

  if (filtered.length === 0) return null

  return (
    <div style={{
      position: 'absolute',
      bottom: '100%',
      left: 0,
      right: 0,
      maxHeight: '200px',
      overflowY: 'auto',
      background: 'var(--surface)',
      border: '1px solid var(--muted)',
      borderRadius: '4px',
      marginBottom: '4px',
      fontFamily: 'monospace',
      zIndex: 10,
    }}>
      {filtered.map((cmd) => (
        <div
          key={cmd.name}
          onClick={() => onSelect(cmd.template)}
          style={{
            padding: '0.5rem 0.75rem',
            cursor: 'pointer',
            fontSize: '0.8rem',
            borderBottom: '1px solid var(--bg)',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'var(--bg)'
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'transparent'
          }}
        >
          <div style={{ color: 'var(--accent)' }}>/{cmd.name}</div>
          <div style={{ color: 'var(--muted)', fontSize: '0.7rem' }}>{cmd.label}</div>
        </div>
      ))}
    </div>
  )
}
