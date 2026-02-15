export interface User {
  id: string
  username: string
  email: string
  display_name: string
  avatar_url: string
  status: 'online' | 'offline' | 'away' | 'dnd'
  bio: string
  created_at: string
  updated_at: string
}

export interface Server {
  id: string
  name: string
  description: string
  icon_url: string
  owner_id: string
  created_at: string
  updated_at: string
}

export interface Member {
  server_id: string
  user_id: string
  nickname: string
  role: 'owner' | 'admin' | 'moderator' | 'member'
  joined_at: string
  updated_at: string
}

export interface Channel {
  id: string
  server_id: string
  name: string
  description: string
  type: 'text' | 'voice' | 'dm'
  position: number
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  channel_id: string
  user_id: string
  content: string
  type: 'text' | 'file' | 'audio'
  attachments: string[]
  edited_at?: string
  created_at: string
}

export interface WSMessage {
  type: string
  channel_id?: string
  server_id?: string
  data: Record<string, unknown>
  timestamp: string
}
