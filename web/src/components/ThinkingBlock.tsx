import { useState } from 'react'

interface ThinkingBlockProps {
  text: string
}

export default function ThinkingBlock({ text }: ThinkingBlockProps) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div
      onClick={() => setExpanded(!expanded)}
      style={{
        margin: '0.25rem 0',
        cursor: 'pointer',
        fontFamily: 'monospace',
        fontSize: '0.8rem',
      }}
    >
      <span style={{ color: 'var(--muted)' }}>
        {expanded ? '\u25BC' : '\u25B6'} thinking...
      </span>
      {expanded && (
        <div style={{
          color: 'var(--muted)',
          fontStyle: 'italic',
          padding: '0.4rem 0 0.4rem 1rem',
          fontSize: '0.75rem',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
        }}>
          {text}
        </div>
      )}
    </div>
  )
}
