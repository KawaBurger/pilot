import { useState, useEffect, useRef, type KeyboardEvent } from 'react'
import { api } from '../api/client'
import type { Command } from '../types'
import SlashCommands from './SlashCommands'

interface InputBarProps {
  onSend: (text: string) => void
  onInterrupt: () => void
  isThinking?: boolean
}

export default function InputBar({ onSend, onInterrupt, isThinking }: InputBarProps) {
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const [commands, setCommands] = useState<Command[]>([])
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    api.commands().then(setCommands).catch(() => {})
  }, [])

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  async function handleSend() {
    const trimmed = text.trim()
    if (!trimmed) return
    try {
      await onSend(trimmed)
      setText('')
    } catch (err: any) {
      setError(err.message || 'Failed to send')
    }
  }

  function handleSlashSelect(template: string) {
    setText(template)
    textareaRef.current?.focus()
  }

  const showSlash = text.startsWith('/')

  return (
    <div style={{
      position: 'relative',
      borderTop: '1px solid var(--surface)',
      padding: '0.5rem',
      background: 'var(--bg)',
      flexShrink: 0,
    }}>
      {error && (
        <div style={{
          padding: '0.25rem 0.5rem',
          marginBottom: '0.4rem',
          color: '#f87171',
          fontSize: '0.8rem',
          cursor: 'pointer',
        }} onClick={() => setError('')}>
          {error}
        </div>
      )}

      {isThinking && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.4rem',
          padding: '0.25rem 0.5rem',
          marginBottom: '0.4rem',
          color: 'var(--accent)',
          fontSize: '0.8rem',
        }}>
          <span className="thinking-dots">Thinking</span>
        </div>
      )}

      {showSlash && (
        <SlashCommands
          commands={commands}
          onSelect={handleSlashSelect}
          filter={text}
        />
      )}

      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '0.4rem',
      }}>
        <button
          onClick={onInterrupt}
          style={{
            background: 'none',
            border: 'none',
            fontSize: '1.1rem',
            cursor: 'pointer',
            padding: '0.3rem',
            flexShrink: 0,
          }}
          title="Interrupt"
        >
          {'\u23F9'}
        </button>

        <textarea
          ref={textareaRef}
          value={text}
          onChange={(e) => { setText(e.target.value); setError('') }}
          onKeyDown={handleKeyDown}
          rows={1}
          placeholder="Send a message..."
          style={{
            flex: 1,
            padding: '0.5rem 0.6rem',
            background: 'var(--surface)',
            color: 'var(--fg)',
            border: '1px solid var(--muted)',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.85rem',
            resize: 'none',
            outline: 'none',
          }}
        />

        <button
          onClick={handleSend}
          style={{
            background: 'none',
            border: 'none',
            fontSize: '1.1rem',
            cursor: 'pointer',
            padding: '0.3rem',
            flexShrink: 0,
          }}
          title="Send"
        >
          {'\u27A4'}
        </button>
      </div>
    </div>
  )
}
