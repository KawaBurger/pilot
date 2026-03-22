import { useState, type FormEvent } from 'react'
import { api } from '../api/client'

interface NewSessionProps {
  onCreated: (sessionId: string, tmuxName: string, title: string) => void
  onBack: () => void
}

export default function NewSession({ onCreated, onBack }: NewSessionProps) {
  const [cwd, setCwd] = useState('~')
  const [prompt, setPrompt] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await api.newSession(cwd, prompt.trim() || undefined)
      onCreated(result.sessionId, result.tmuxSession, prompt.trim().slice(0, 40) || 'New Session')
    } catch (err: any) {
      setError(err.message)
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: 'var(--bg)',
        color: 'var(--accent)',
        fontFamily: 'monospace',
      }}>
        <div style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>
          Starting Claude...
        </div>
        <div style={{ color: 'var(--muted)', fontSize: '0.8rem' }}>
          This may take up to 30 seconds
        </div>
      </div>
    )
  }

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      background: 'var(--bg)',
      color: 'var(--fg)',
      fontFamily: 'monospace',
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        padding: '0.75rem 1rem',
        borderBottom: '1px solid var(--surface)',
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
            padding: 0,
            marginRight: '1rem',
          }}
        >
          &larr; Back
        </button>
        <span style={{ fontSize: '0.95rem' }}>New Session</span>
      </div>

      <form
        onSubmit={handleSubmit}
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: '0.75rem',
          padding: '1rem',
          flex: 1,
        }}
      >
        <label style={{ color: 'var(--muted)', fontSize: '0.8rem' }}>
          Working directory
        </label>
        <input
          type="text"
          value={cwd}
          onChange={(e) => setCwd(e.target.value)}
          style={{
            padding: '0.6rem 0.8rem',
            background: 'var(--surface)',
            color: 'var(--fg)',
            border: '1px solid var(--muted)',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.9rem',
          }}
        />

        <label style={{ color: 'var(--muted)', fontSize: '0.8rem' }}>
          Prompt (optional)
        </label>
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={4}
          placeholder="What should Claude do?"
          style={{
            padding: '0.6rem 0.8rem',
            background: 'var(--surface)',
            color: 'var(--fg)',
            border: '1px solid var(--muted)',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.9rem',
            resize: 'vertical',
          }}
        />

        {error && (
          <div style={{ color: '#f87171', fontSize: '0.8rem' }}>
            {error}
          </div>
        )}

        <button
          type="submit"
          style={{
            padding: '0.6rem',
            background: 'var(--accent)',
            color: 'var(--bg)',
            border: 'none',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.9rem',
            cursor: 'pointer',
            marginTop: '0.5rem',
          }}
        >
          Start Session
        </button>
      </form>
    </div>
  )
}
