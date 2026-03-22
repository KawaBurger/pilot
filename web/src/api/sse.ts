import type { Message } from '../types'
import { getToken } from './client'

export function createSSE(
  sessionId: string,
  onMessage: (msg: Message) => void,
  onError: (err: Event) => void,
): EventSource {
  const token = getToken()
  const url = `/api/sessions/${sessionId}/stream?token=${encodeURIComponent(token ?? '')}`
  const source = new EventSource(url)

  source.addEventListener('message', (event) => {
    const msg: Message = JSON.parse(event.data)
    onMessage(msg)
  })

  source.onerror = onError

  return source
}
