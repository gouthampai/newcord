import { createContext, useContext, useState, useCallback, useEffect } from 'react'
import type { Server, Channel, Member } from '../types'
import { useAuth } from './AuthContext'
import * as api from '../services/api'

interface AppContextType {
  servers: Server[]
  currentServer: Server | null
  channels: Channel[]
  currentChannel: Channel | null
  members: Member[]
  setServers: (servers: Server[]) => void
  selectServer: (server: Server) => Promise<void>
  selectChannel: (channel: Channel) => void
  refreshServers: () => Promise<void>
  refreshChannels: (serverId: string) => Promise<void>
  refreshMembers: (serverId: string) => Promise<void>
  addServer: (server: Server) => void
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
    setCurrentServer(server)
    setCurrentChannel(null)
    await Promise.all([refreshChannels(server.id), refreshMembers(server.id)])
  }, [refreshChannels, refreshMembers])

  const selectChannel = useCallback((channel: Channel) => {
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

  const { user } = useAuth()

  useEffect(() => {
    if (user) {
      refreshServers()
    }
  }, [user, refreshServers])

  return (
    <AppContext.Provider value={{
      servers, currentServer, channels, currentChannel, members,
      setServers, selectServer, selectChannel, refreshServers,
      refreshChannels, refreshMembers, addServer,
    }}>
      {children}
    </AppContext.Provider>
  )
}
