import type { User, Server, Member, Channel, Message } from '../types'

const API_BASE = '/api/v1'

function getToken(): string | null {
  return localStorage.getItem('token')
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

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `Request failed: ${res.status}`)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

// Auth
export async function register(data: { username: string; email: string; password: string; display_name: string }) {
  return request<{ token: string; user: User }>('/auth/register', { method: 'POST', body: JSON.stringify(data) })
}

export async function login(data: { email: string; password: string }) {
  return request<{ token: string; user: User }>('/auth/login', { method: 'POST', body: JSON.stringify(data) })
}

// Users
export async function getUser(id: string) {
  return request<User>(`/users/${id}`)
}

export async function updateUser(id: string, data: Partial<Pick<User, 'display_name' | 'avatar_url' | 'status' | 'bio'>>) {
  return request<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

// Servers
export async function createServer(data: { name: string; description?: string; icon_url?: string }) {
  return request<Server>('/servers', { method: 'POST', body: JSON.stringify(data) })
}

export async function getServer(id: string) {
  return request<Server>(`/servers/${id}`)
}

export async function updateServer(id: string, data: Partial<Pick<Server, 'name' | 'description' | 'icon_url'>>) {
  return request<Server>(`/servers/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

export async function deleteServer(id: string) {
  return request<void>(`/servers/${id}`, { method: 'DELETE' })
}

export async function getMembers(serverId: string) {
  return request<Member[]>(`/servers/${serverId}/members`)
}

export async function addMember(serverId: string, data: { user_id: string; nickname?: string; role?: string }) {
  return request<Member>(`/servers/${serverId}/members`, { method: 'POST', body: JSON.stringify(data) })
}

// My Servers
export async function getMyServers() {
  return request<Server[]>('/users/@me/servers')
}

// Invites
export async function createInvite(serverId: string, data?: { max_uses?: number; expires_in?: number }) {
  return request<{ id: string; server_id: string; code: string; max_uses: number; uses: number; expires_at: string; created_at: string }>(
    `/servers/${serverId}/invites`, { method: 'POST', body: JSON.stringify(data || {}) }
  )
}

export async function joinInvite(code: string) {
  return request<Server>(`/invites/${code}/join`, { method: 'POST' })
}

export async function getServerInvites(serverId: string) {
  return request<{ id: string; server_id: string; code: string; max_uses: number; uses: number; expires_at: string; created_at: string }[]>(
    `/servers/${serverId}/invites`
  )
}

// Channels
export async function createChannel(data: { server_id: string; name: string; description?: string; type?: string; position?: number }) {
  return request<Channel>('/channels', { method: 'POST', body: JSON.stringify(data) })
}

export async function getChannel(id: string) {
  return request<Channel>(`/channels/${id}`)
}

export async function getServerChannels(serverId: string) {
  return request<Channel[]>(`/servers/${serverId}/channels`)
}

export async function updateChannel(id: string, data: Partial<Pick<Channel, 'name' | 'description' | 'position'>>) {
  return request<Channel>(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

export async function deleteChannel(id: string) {
  return request<void>(`/channels/${id}`, { method: 'DELETE' })
}

// Messages
export async function createMessage(channelId: string, data: { content: string; type?: string; attachments?: string[] }) {
  return request<Message>(`/channels/${channelId}/messages`, { method: 'POST', body: JSON.stringify({ type: 'text', ...data }) })
}

export async function getMessages(channelId: string, limit = 50) {
  return request<Message[]>(`/channels/${channelId}/messages?limit=${limit}`)
}

export async function updateMessage(channelId: string, messageId: string, data: { content: string }) {
  return request<Message>(`/channels/${channelId}/messages/${messageId}`, { method: 'PUT', body: JSON.stringify(data) })
}

export async function deleteMessage(channelId: string, messageId: string) {
  return request<void>(`/channels/${channelId}/messages/${messageId}`, { method: 'DELETE' })
}
