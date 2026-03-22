import type { Session, Message, Command } from '../types'

const TOKEN_KEY = 'pilot_token'

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(path, { ...options, headers })

  if (res.status === 401) {
    clearToken()
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(`${res.status}: ${body}`)
  }

  return res.json()
}

export const api = {
  login(username: string, password: string) {
    return request<{ token: string }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  },

  sessions() {
    return request<Session[]>('/api/sessions')
  },

  session(id: string) {
    return request<Session>(`/api/sessions/${id}`)
  },

  messages(sessionId: string, after?: string) {
    const params = after ? `?after=${encodeURIComponent(after)}` : ''
    return request<Message[]>(`/api/sessions/${sessionId}/messages${params}`)
  },

  newSession(cwd: string, prompt?: string) {
    return request<{ sessionId: string; tmuxSession: string }>('/api/sessions', {
      method: 'POST',
      body: JSON.stringify({ cwd, prompt }),
    })
  },

  resumeSession(id: string) {
    return request<{ sessionId: string; tmuxSession: string }>(`/api/sessions/${id}/resume`, {
      method: 'POST',
    })
  },

  send(id: string, message: string) {
    return request<{ ok: boolean }>(`/api/sessions/${id}/message`, {
      method: 'POST',
      body: JSON.stringify({ message }),
    })
  },

  interrupt(id: string) {
    return request<{ ok: boolean }>(`/api/sessions/${id}/interrupt`, {
      method: 'POST',
    })
  },

  terminal(id: string) {
    return request<{ content: string; hasPrompt: boolean }>(`/api/sessions/${id}/terminal`)
  },

  sendKeys(id: string, keys: string[]) {
    return request<{ ok: boolean }>(`/api/sessions/${id}/keys`, {
      method: 'POST',
      body: JSON.stringify({ keys }),
    })
  },

  commands() {
    return request<Command[]>('/api/commands')
  },
}
