import { useState } from 'react'
import { getToken } from './api/client'
import Login from './components/Login'
import Welcome from './components/Welcome'
import NewSession from './components/NewSession'
import SessionList from './components/SessionList'
import Conversation from './components/Conversation'

type View = 'welcome' | 'new' | 'resume' | 'conversation'

function App() {
  const [authed, setAuthed] = useState(() => !!getToken())
  const [view, setView] = useState<View>('welcome')
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [tmuxName, setTmuxName] = useState<string | null>(null)
  const [sessionTitle, setSessionTitle] = useState('')

  if (!authed) {
    return <Login onLogin={() => setAuthed(true)} />
  }

  switch (view) {
    case 'welcome':
      return (
        <Welcome
          onNew={() => setView('new')}
          onResume={() => setView('resume')}
        />
      )

    case 'new':
      return (
        <NewSession
          onCreated={(id, tmux, title) => {
            setSessionId(id)
            setTmuxName(tmux)
            setSessionTitle(title)
            setView('conversation')
          }}
          onBack={() => setView('welcome')}
        />
      )

    case 'resume':
      return (
        <SessionList
          onSelect={(id, tmux, title) => {
            setSessionId(id)
            setTmuxName(tmux)
            setSessionTitle(title)
            setView('conversation')
          }}
          onBack={() => setView('welcome')}
        />
      )

    case 'conversation':
      return (
        <Conversation
          sessionId={sessionId!}
          tmuxName={tmuxName!}
          title={sessionTitle}
          onBack={() => setView('welcome')}
        />
      )
  }
}

export default App
