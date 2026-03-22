import { useState } from 'react'

interface ToolUseBlockProps {
  name: string
  input: any
}

function formatToolSummary(name: string, input: any): string {
  if (!input) return name
  switch (name) {
    case 'Bash':
      return `$ ${input.command || ''}`.slice(0, 60)
    case 'Read':
      return `Read ${input.file_path || ''}`.replace(/.*\//, 'Read .../')
    case 'Write':
      return `Write ${input.file_path || ''}`.replace(/.*\//, 'Write .../')
    case 'Edit':
      return `Edit ${input.file_path || ''}`.replace(/.*\//, 'Edit .../')
    case 'Glob':
      return `Glob ${input.pattern || ''}`
    case 'Grep':
      return `Grep "${input.pattern || ''}"`
    case 'Agent':
      return `Agent: ${input.description || ''}`
    default:
      return name
  }
}

export default function ToolUseBlock({ name, input }: ToolUseBlockProps) {
  const [expanded, setExpanded] = useState(false)
  const summary = formatToolSummary(name, input)

  return (
    <div style={{
      border: '1px solid var(--surface)',
      borderRadius: '4px',
      margin: '0.25rem 0',
      fontFamily: 'monospace',
    }}>
      <div
        onClick={() => setExpanded(!expanded)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          padding: '0.4rem 0.6rem',
          cursor: 'pointer',
          background: 'var(--surface)',
          fontSize: '0.8rem',
          color: 'var(--fg)',
        }}
      >
        <span style={{ fontSize: '0.7rem' }}>{expanded ? '▼' : '▶'}</span>
        <span style={{ color: 'var(--accent)' }}>{summary}</span>
      </div>
      {expanded && (
        <pre style={{
          padding: '0.5rem 0.6rem',
          margin: 0,
          fontSize: '0.75rem',
          color: 'var(--muted)',
          overflowX: 'auto',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          maxHeight: '300px',
          overflowY: 'auto',
        }}>
          {JSON.stringify(input, null, 2)}
        </pre>
      )}
    </div>
  )
}
