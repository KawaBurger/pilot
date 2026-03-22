import { useState, type FormEvent } from 'react'
import { api, setToken } from '../api/client'

interface LoginProps {
  onLogin: () => void
}

export default function Login({ onLogin }: LoginProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const { token } = await api.login(username, password)
      setToken(token)
      onLogin()
    } catch (err: any) {
      setError(err.message || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100vh',
      padding: '1rem',
      background: 'var(--bg)',
      color: 'var(--fg)',
    }}>
      <h1 style={{ fontFamily: 'monospace', marginBottom: '2rem', color: 'var(--accent)' }}>
        Pilot
      </h1>
      <form
        onSubmit={handleSubmit}
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: '0.75rem',
          width: '100%',
          maxWidth: '320px',
        }}
      >
        <input
          type="text"
          placeholder="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
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
        <input
          type="password"
          placeholder="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
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
        {error && (
          <div style={{
            color: '#f87171',
            fontFamily: 'monospace',
            fontSize: '0.8rem',
          }}>
            {error}
          </div>
        )}
        <button
          type="submit"
          disabled={loading}
          style={{
            padding: '0.6rem',
            background: 'var(--accent)',
            color: 'var(--bg)',
            border: 'none',
            borderRadius: '4px',
            fontFamily: 'monospace',
            fontSize: '0.9rem',
            cursor: loading ? 'wait' : 'pointer',
            opacity: loading ? 0.6 : 1,
          }}
        >
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  )
}
