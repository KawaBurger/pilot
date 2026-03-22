import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Session } from '../types'

function timeAgo(unixSeconds: number): string {
  const diff = Math.floor(Date.now() / 1000 - unixSeconds)
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

interface SessionListProps {
  onSelect: (id: string, tmuxName: string, title: string) => void
  onBack: () => void
}

export default function SessionList({ onSelect, onBack }: SessionListProps) {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [resumingId, setResumingId] = useState<string | null>(null)

  useEffect(() => {
    api.sessions()
      .then(setSessions)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleSelect(session: Session) {
    setResumingId(session.id)
    try {
      const result = await api.resumeSession(session.id)
      onSelect(result.sessionId, result.tmuxSession, session.title || session.id.slice(0, 8))
    } catch (err: any) {
      setError(err.message)
      setResumingId(null)
    }
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
        <span style={{ fontSize: '0.95rem' }}>Sessions</span>
      </div>

      <div style={{ flex: 1, overflowY: 'auto', padding: '0.5rem' }}>
        {loading && (
          <div style={{ color: 'var(--muted)', padding: '1rem', textAlign: 'center' }}>
            Loading...
          </div>
        )}
        {error && (
          <div style={{ color: '#f87171', padding: '1rem', textAlign: 'center' }}>
            {error}
          </div>
        )}
        {!loading && sessions.length === 0 && (
          <div style={{ color: 'var(--muted)', padding: '1rem', textAlign: 'center' }}>
            No sessions found
          </div>
        )}
        {sessions.map((session) => (
          <div
            key={session.id}
            onClick={() => !resumingId && handleSelect(session)}
            style={{
              padding: '0.75rem 1rem',
              borderRadius: '4px',
              background: 'var(--surface)',
              marginBottom: '0.5rem',
              cursor: resumingId ? 'wait' : 'pointer',
              opacity: resumingId && resumingId !== session.id ? 0.5 : 1,
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ fontSize: '0.9rem' }}>
                {session.title || session.id.slice(0, 8)}
              </span>
              {session.status === 'active' && (
                <span style={{
                  display: 'inline-block',
                  width: '8px',
                  height: '8px',
                  borderRadius: '50%',
                  background: '#4ade80',
                  flexShrink: 0,
                }} />
              )}
            </div>
            {session.lastUserMessage && (
              <div style={{
                color: 'var(--fg)',
                fontSize: '0.8rem',
                marginTop: '0.25rem',
                opacity: 0.7,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}>
                {session.lastUserMessage}
              </div>
            )}
            <div style={{
              color: 'var(--muted)',
              fontSize: '0.75rem',
              marginTop: '0.25rem',
              display: 'flex',
              gap: '0.75rem',
            }}>
              <span>{session.project}</span>
              {session.updatedAt > 0 && <span>{timeAgo(session.updatedAt)}</span>}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
