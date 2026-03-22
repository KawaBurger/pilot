import { useEffect, useState, useRef } from 'react'
import { api } from '../api/client'

interface TerminalPromptProps {
  sessionId: string
  isThinking: boolean
}

export default function TerminalPrompt({ sessionId, isThinking }: TerminalPromptProps) {
  const [content, setContent] = useState('')
  const [hasPrompt, setHasPrompt] = useState(false)
  const [sending, setSending] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!isThinking) {
      setContent('')
      setHasPrompt(false)
      if (pollRef.current) clearInterval(pollRef.current)
      return
    }

    const poll = async () => {
      try {
        const result = await api.terminal(sessionId)
        setContent(result.content)
        setHasPrompt(result.hasPrompt)
      } catch {
        // ignore - session might not have tmux mapping yet
      }
    }

    poll()
    pollRef.current = setInterval(poll, 2000)

    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [sessionId, isThinking])

  async function sendKey(keys: string[]) {
    setSending(true)
    try {
      await api.sendKeys(sessionId, keys)
      // Re-poll immediately
      const result = await api.terminal(sessionId)
      setContent(result.content)
      setHasPrompt(result.hasPrompt)
    } catch {
      // ignore
    } finally {
      setSending(false)
    }
  }

  if (!isThinking || !hasPrompt) return null

  return (
    <div style={{
      margin: '0 0.5rem',
      padding: '0.5rem',
      background: '#1e1e2e',
      border: '1px solid var(--accent)',
      borderRadius: '4px',
      fontSize: '0.75rem',
      fontFamily: 'monospace',
    }}>
      <pre style={{
        margin: 0,
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-word',
        color: 'var(--fg)',
        maxHeight: '200px',
        overflow: 'auto',
      }}>
        {content}
      </pre>
      <div style={{
        display: 'flex',
        gap: '0.5rem',
        marginTop: '0.5rem',
        justifyContent: 'flex-end',
      }}>
        <button
          onClick={() => sendKey(['Escape'])}
          disabled={sending}
          style={{
            background: 'var(--surface)',
            color: '#f87171',
            border: '1px solid #f87171',
            borderRadius: '4px',
            padding: '0.25rem 0.75rem',
            fontFamily: 'monospace',
            fontSize: '0.75rem',
            cursor: 'pointer',
          }}
        >
          Reject (Esc)
        </button>
        <button
          onClick={() => sendKey(['Enter'])}
          disabled={sending}
          style={{
            background: 'var(--accent)',
            color: '#1a1b26',
            border: 'none',
            borderRadius: '4px',
            padding: '0.25rem 0.75rem',
            fontFamily: 'monospace',
            fontSize: '0.75rem',
            cursor: 'pointer',
          }}
        >
          Approve (Enter)
        </button>
      </div>
    </div>
  )
}
