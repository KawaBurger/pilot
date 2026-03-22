import { useRef, useCallback, useEffect } from 'react'
import { createSSE } from '../api/sse'
import { api } from '../api/client'
import type { Message } from '../types'

export function useSSE(
  sessionId: string | null,
  onMessage: (msg: Message) => void,
) {
  const sourceRef = useRef<EventSource | null>(null)
  const lastUUIDRef = useRef<string | null>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    if (!sessionId) return

    sourceRef.current?.close()

    const source = createSSE(
      sessionId,
      (msg) => {
        lastUUIDRef.current = msg.uuid
        onMessageRef.current(msg)
      },
      () => {
        // On disconnect: close, wait 2s, fill gap, reconnect
        sourceRef.current?.close()
        sourceRef.current = null

        setTimeout(async () => {
          if (!sessionId) return

          try {
            // Fetch missed messages — use after if we have a cursor, otherwise full history
            const missed = await api.messages(sessionId, lastUUIDRef.current ?? undefined)
            for (const msg of missed) {
              lastUUIDRef.current = msg.uuid
              onMessageRef.current(msg)
            }
          } catch {
            // Ignore fetch errors, reconnect anyway
          }

          connect()
        }, 2000)
      },
    )

    sourceRef.current = source
  }, [sessionId])

  useEffect(() => {
    connect()

    return () => {
      sourceRef.current?.close()
      sourceRef.current = null
    }
  }, [connect])
}
