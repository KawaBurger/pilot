export interface Session {
  id: string
  title: string
  project: string
  cwd: string
  gitBranch: string
  status: 'active' | 'ended'
  jsonlPath: string
  updatedAt: number
  lastUserMessage: string
}

export interface ContentBlock {
  type: 'text' | 'tool_use' | 'tool_result' | 'thinking'
  text?: string
  name?: string
  id?: string
  input?: any
  content?: any
}

export interface Message {
  type: 'user' | 'assistant' | 'progress'
  uuid: string
  parentUuid: string
  timestamp: string
  sessionId: string
  cwd: string
  gitBranch: string
  message?: {
    role: string
    content: ContentBlock[]
    model?: string
    stop_reason?: string
  }
}

export interface Command {
  name: string
  label: string
  template: string
}
