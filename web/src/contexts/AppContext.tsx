import { createContext, useContext, useState, useCallback, useEffect, useRef } from 'react'
import type { Server, Channel, Member } from '../types'
import { useAuth } from './AuthContext'
import * as api from '../services/api'

interface AppContextType {
  servers: Server[]
  currentServer: Server | null
  channels: Channel[]
  currentChannel: Channel | null
  members: Member[]
  onlineUsers: Set<string>
  setServers: (servers: Server[]) => void
  selectServer: (server: Server) => Promise<void>
  selectChannel: (channel: Channel) => void
  refreshServers: () => Promise<void>
  refreshChannels: (serverId: string) => Promise<void>
  refreshMembers: (serverId: string) => Promise<void>
  addServer: (server: Server) => void
  clearCurrentServer: () => void
  setOnlineUsers: React.Dispatch<React.SetStateAction<Set<string>>>
}

const AppContext = createContext<AppContextType>(null!)

export function useApp() {
  return useContext(AppContext)
}

export function AppProvider({ children }: { children: React.ReactNode }) {
  const [servers, setServers] = useState<Server[]>([])
  const [currentServer, setCurrentServer] = useState<Server | null>(null)
  const [channels, setChannels] = useState<Channel[]>([])
  const [currentChannel, setCurrentChannel] = useState<Channel | null>(null)
  const [members, setMembers] = useState<Member[]>([])
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())
  const restoredRef = useRef(false)

  const refreshChannels = useCallback(async (serverId: string) => {
    try {
      const chs = await api.getServerChannels(serverId)
      setChannels(chs || [])
    } catch {
      setChannels([])
    }
  }, [])

  const refreshMembers = useCallback(async (serverId: string) => {
    try {
      const mems = await api.getMembers(serverId)
      setMembers(mems || [])
    } catch {
      setMembers([])
    }
  }, [])

  const selectServer = useCallback(async (server: Server) => {
    localStorage.setItem('lastServerId', server.id)
    setCurrentServer(server)
    setCurrentChannel(null)
    setOnlineUsers(new Set())

    const [chs] = await Promise.all([
      api.getServerChannels(server.id).catch(() => [] as Channel[]),
      refreshMembers(server.id),
    ])
    const fetchedChannels = chs || []
    setChannels(fetchedChannels)

    // Auto-select: saved channel > first text channel
    const lastChannelId = localStorage.getItem('lastChannelId')
    const textChannels = fetchedChannels.filter(c => c.type === 'text' || !c.type)
    const saved = lastChannelId ? textChannels.find(c => c.id === lastChannelId) : null
    const toSelect = saved || textChannels[0] || null
    if (toSelect) {
      setCurrentChannel(toSelect)
      localStorage.setItem('lastChannelId', toSelect.id)
    }
  }, [refreshMembers])

  const selectChannel = useCallback((channel: Channel) => {
    localStorage.setItem('lastChannelId', channel.id)
    setCurrentChannel(channel)
  }, [])

  const refreshServers = useCallback(async () => {
    try {
      const srvs = await api.getMyServers()
      setServers(srvs || [])
    } catch {
      setServers([])
    }
  }, [])

  const addServer = useCallback((server: Server) => {
    setServers(prev => [...prev, server])
  }, [])

  const clearCurrentServer = useCallback(() => {
    setCurrentServer(null)
    setChannels([])
    setCurrentChannel(null)
    setMembers([])
    setOnlineUsers(new Set())
    localStorage.removeItem('lastServerId')
    localStorage.removeItem('lastChannelId')
  }, [])

  const { user } = useAuth()

  useEffect(() => {
    if (user) {
      refreshServers()
    }
  }, [user, refreshServers])

  // Restore last server after server list loads
  useEffect(() => {
    if (servers.length === 0 || currentServer || restoredRef.current) return
    restoredRef.current = true

    const lastServerId = localStorage.getItem('lastServerId')
    const saved = lastServerId ? servers.find(s => s.id === lastServerId) : null
    const toSelect = saved || servers[0]
    if (toSelect) {
      selectServer(toSelect)
    }
  }, [servers, currentServer, selectServer])

  return (
    <AppContext.Provider value={{
      servers, currentServer, channels, currentChannel, members, onlineUsers,
      setServers, selectServer, selectChannel, refreshServers,
      refreshChannels, refreshMembers, addServer, clearCurrentServer, setOnlineUsers,
    }}>
      {children}
    </AppContext.Provider>
  )
}
