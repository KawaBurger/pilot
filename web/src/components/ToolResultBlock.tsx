import { useState } from 'react'

interface ToolResultBlockProps {
  content: any
}

export default function ToolResultBlock({ content }: ToolResultBlockProps) {
  const [open, setOpen] = useState(false)

  const text = typeof content === 'string'
    ? content
    : Array.isArray(content)
      ? content.map(c => c.text || JSON.stringify(c)).join('\n')
      : JSON.stringify(content, null, 2)

  if (!text || text.length === 0) return null

  const preview = text.slice(0, 80).replace(/\n/g, ' ')
  const isLong = text.length > 80

  return (
    <div style={{
      borderLeft: '2px solid var(--muted)',
      paddingLeft: '0.5rem',
      marginBottom: '0.25rem',
    }}>
      <div
        onClick={() => isLong && setOpen(!open)}
        style={{
          fontSize: '0.8rem',
          color: 'var(--muted)',
          cursor: isLong ? 'pointer' : 'default',
          fontFamily: 'monospace',
        }}
      >
        {isLong && <span>{open ? '▼ ' : '▶ '}</span>}
        {open ? '' : preview}{isLong && !open ? '...' : ''}
      </div>
      {open && (
        <pre style={{
          fontSize: '0.75rem',
          color: 'var(--fg)',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
          maxHeight: '200px',
          overflowY: 'auto',
          margin: '0.25rem 0 0 0',
          padding: 0,
          background: 'transparent',
        }}>
          {text}
        </pre>
      )}
    </div>
  )
}
