import type { Message, ContentBlock } from '../types'
import TextBlock from './TextBlock'
import ToolUseBlock from './ToolUseBlock'
import ThinkingBlock from './ThinkingBlock'
import ToolResultBlock from './ToolResultBlock'

interface MessageBubbleProps {
  message: Message
}

export default function MessageBubble({ message }: MessageBubbleProps) {
  const msg = message.message
  if (!msg || !msg.content || msg.content.length === 0) return null

  const blocks = msg.content

  // User messages that only contain tool_result blocks are tool responses,
  // not actual user input — render them differently
  const isToolResponse = msg.role === 'user' && blocks.every(b => b.type === 'tool_result')
  const isUser = msg.role === 'user' && !isToolResponse

  // Skip pure tool_result messages — they'll be shown inline with the tool_use
  if (isToolResponse) {
    return (
      <div style={{ padding: '0 0.75rem', marginBottom: '0.25rem' }}>
        {blocks.map((block, i) => (
          <ToolResultBlock key={i} content={block.content} />
        ))}
      </div>
    )
  }

  return (
    <div style={{
      padding: '0.6rem 0.75rem',
      marginBottom: '0.5rem',
      borderRadius: '6px',
      background: isUser ? 'var(--surface)' : 'transparent',
      fontFamily: 'monospace',
    }}>
      <div style={{
        fontSize: '0.75rem',
        color: isUser ? 'var(--accent)' : 'var(--muted)',
        marginBottom: '0.35rem',
      }}>
        {isUser ? '> You' : '◆ Claude'}
      </div>

      {blocks.map((block, i) => renderBlock(block, i))}
    </div>
  )
}

function renderBlock(block: ContentBlock, i: number) {
  switch (block.type) {
    case 'text':
      return block.text ? <TextBlock key={i} text={block.text} /> : null
    case 'tool_use':
      return <ToolUseBlock key={i} name={block.name || 'tool'} input={block.input} />
    case 'tool_result':
      return <ToolResultBlock key={i} content={block.content} />
    case 'thinking':
      // thinking blocks may have no text (just a marker)
      return block.text ? <ThinkingBlock key={i} text={block.text} /> : null
    default:
      return null
  }
}
