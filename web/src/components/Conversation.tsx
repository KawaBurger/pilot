import { useEffect, useState, useRef, useMemo } from 'react'
import { useSession } from '../hooks/useSession'
import { api } from '../api/client'
import TopBar from './TopBar'
import MessageList from './MessageList'
import Dashboard from './Dashboard'
import InputBar from './InputBar'
import TerminalPrompt from './TerminalPrompt'

interface ConversationProps {
  sessionId: string
  tmuxName: string
  title: string
  onBack: () => void
}

export default function Conversation({ sessionId, title: initialTitle, onBack }: ConversationProps) {
  const [view, setView] = useState<'conversation' | 'dashboard'>('conversation')
  const [title, setTitle] = useState(initialTitle)
  const { messages, loading, loadHistory, send, interrupt, interrupted } = useSession(sessionId)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    loadHistory()
  }, [loadHistory])

  // Poll for title if not set yet
  useEffect(() => {
    if (title) return

    pollRef.current = setInterval(async () => {
      try {
        const session = await api.session(sessionId)
        if (session.title) {
          setTitle(session.title)
          if (pollRef.current) clearInterval(pollRef.current)
        }
      } catch {
        // ignore
      }
    }, 5000)

    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [sessionId, title])

  const isThinking = useMemo(() => {
    if (interrupted) return false
    if (messages.length === 0) return false
    const last = messages[messages.length - 1]
    if (last.type === 'user') return true
    if (last.type === 'assistant' && last.message?.stop_reason === 'tool_use') return true
    return false
  }, [messages])

  return (
    <div className="flex flex-col h-screen">
      <TopBar
        title={title || sessionId.slice(0, 8)}
        onBack={onBack}
        view={view}
        onToggleView={() => setView((v) => (v === 'conversation' ? 'dashboard' : 'conversation'))}
      />

      {loading ? (
        <div className="flex-1 flex items-center justify-center text-[var(--muted)]">
          Loading conversation...
        </div>
      ) : view === 'conversation' ? (
        <MessageList messages={messages} />
      ) : (
        <Dashboard messages={messages} />
      )}

      <TerminalPrompt sessionId={sessionId} isThinking={isThinking} />
      <InputBar onSend={send} onInterrupt={interrupt} isThinking={isThinking} />
    </div>
  )
}
