import type { Message } from '../types'

interface DashboardProps {
  messages: Message[]
}

export default function Dashboard({ messages }: DashboardProps) {
  const lastMessage = messages[messages.length - 1]
  const gitBranch = lastMessage?.gitBranch || '-'
  const cwd = lastMessage?.cwd || '-'

  const messageCount = messages.length
  let toolCallCount = 0
  const recentTools: { name: string; id: string }[] = []
  const changedFiles = new Set<string>()

  for (const msg of messages) {
    if (!msg.message) continue
    for (const block of msg.message.content) {
      if (block.type === 'tool_use') {
        toolCallCount++
        recentTools.push({ name: block.name || 'tool', id: block.id || String(toolCallCount) })

        if (block.input) {
          const input = block.input
          if (typeof input.file_path === 'string') {
            changedFiles.add(input.file_path)
          }
          if (typeof input.path === 'string') {
            changedFiles.add(input.path)
          }
        }
      }
    }
  }

  const last10Tools = recentTools.slice(-10).reverse()

  const statStyle: React.CSSProperties = {
    padding: '0.6rem 0.75rem',
    background: 'var(--surface)',
    borderRadius: '4px',
    fontSize: '0.8rem',
  }

  const labelStyle: React.CSSProperties = {
    color: 'var(--muted)',
    fontSize: '0.7rem',
    marginBottom: '0.2rem',
  }

  return (
    <div style={{
      flex: 1,
      overflowY: 'auto',
      padding: '0.75rem',
      fontFamily: 'monospace',
      color: 'var(--fg)',
    }}>
      <div style={statStyle}>
        <div style={labelStyle}>Git Branch</div>
        <div style={{ color: 'var(--accent)' }}>{gitBranch}</div>
      </div>

      <div style={{ ...statStyle, marginTop: '0.5rem' }}>
        <div style={labelStyle}>Working Directory</div>
        <div>{cwd}</div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr 1fr',
        gap: '0.5rem',
        marginTop: '0.5rem',
      }}>
        <div style={statStyle}>
          <div style={labelStyle}>Messages</div>
          <div style={{ fontSize: '1.1rem', color: 'var(--accent)' }}>{messageCount}</div>
        </div>
        <div style={statStyle}>
          <div style={labelStyle}>Tool Calls</div>
          <div style={{ fontSize: '1.1rem', color: 'var(--accent)' }}>{toolCallCount}</div>
        </div>
        <div style={statStyle}>
          <div style={labelStyle}>Files</div>
          <div style={{ fontSize: '1.1rem', color: 'var(--accent)' }}>{changedFiles.size}</div>
        </div>
      </div>

      <div style={{ marginTop: '0.75rem' }}>
        <div style={{ color: 'var(--muted)', fontSize: '0.75rem', marginBottom: '0.35rem' }}>
          Recent Tools
        </div>
        {last10Tools.length === 0 && (
          <div style={{ color: 'var(--muted)', fontSize: '0.8rem' }}>None yet</div>
        )}
        {last10Tools.map((tool, i) => (
          <div
            key={`${tool.id}-${i}`}
            style={{
              padding: '0.35rem 0.6rem',
              background: 'var(--surface)',
              borderRadius: '3px',
              marginBottom: '0.25rem',
              fontSize: '0.8rem',
              color: 'var(--fg)',
            }}
          >
            {'\u{1F527}'} {tool.name}
          </div>
        ))}
      </div>
    </div>
  )
}
