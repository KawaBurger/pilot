import { useState, useCallback, useRef } from 'react'
import { api } from '../api/client'
import { useSSE } from './useSSE'
import type { Message } from '../types'

export function useSession(sessionId: string | null) {
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(true)
  const [interrupted, setInterrupted] = useState(false)
  const seenUUIDs = useRef(new Set<string>())

  const addMessage = useCallback((msg: Message) => {
    if (msg.type !== 'user' && msg.type !== 'assistant') return
    if (seenUUIDs.current.has(msg.uuid)) return
    seenUUIDs.current.add(msg.uuid)
    setMessages((prev) => [...prev, msg])
  }, [])

  useSSE(sessionId, addMessage)

  const loadHistory = useCallback(async () => {
    if (!sessionId) return
    setLoading(true)
    try {
      const history = await api.messages(sessionId)
      // Deduplicate: merge history with any SSE messages already received
      const allUUIDs = new Set<string>()
      const merged: Message[] = []
      for (const msg of history) {
        if (!allUUIDs.has(msg.uuid)) {
          allUUIDs.add(msg.uuid)
          merged.push(msg)
        }
      }
      setMessages((prev) => {
        for (const msg of prev) {
          if (!allUUIDs.has(msg.uuid)) {
            allUUIDs.add(msg.uuid)
            merged.push(msg)
          }
        }
        return merged
      })
      seenUUIDs.current = allUUIDs
    } catch {
      // JSONL may not exist yet for new sessions — keep existing messages
    } finally {
      setLoading(false)
    }
  }, [sessionId])

  const send = useCallback(
    async (message: string) => {
      if (!sessionId) return
      setInterrupted(false)
      await api.send(sessionId, message)
    },
    [sessionId],
  )

  const interrupt = useCallback(async () => {
    if (!sessionId) return
    await api.interrupt(sessionId)
    setInterrupted(true)
  }, [sessionId])

  return { messages, loading, loadHistory, send, interrupt, interrupted }
}
