import { useCallback } from 'react'
import { useApp } from '../contexts/AppContext'
import { useWebSocket } from '../hooks/useWebSocket'
import type { WSMessage } from '../types'
import ServerSidebar from '../components/ServerSidebar'
import ChannelSidebar from '../components/ChannelSidebar'
import MessageArea from '../components/MessageArea'
import MemberList from '../components/MemberList'

export default function MainLayout() {
  const { currentServer, currentChannel, setOnlineUsers, clearCurrentServer, refreshServers, refreshMembers } = useApp()

  const handleWSMessage = useCallback((msg: WSMessage) => {
    if (msg.type === 'presence_list') {
      const userIds = msg.data.user_ids as string[] | undefined
      setOnlineUsers(new Set(userIds || []))
    } else if (msg.type === 'presence_update') {
      const userId = msg.data.user_id as string
      const status = msg.data.status as string
      setOnlineUsers(prev => {
        const next = new Set(prev)
        if (status === 'online') {
          next.add(userId)
        } else {
          next.delete(userId)
        }
        return next
      })
    } else if (msg.type === 'server_delete') {
      clearCurrentServer()
      refreshServers()
    } else if (msg.type === 'member_join') {
      if (currentServer) {
        refreshMembers(currentServer.id)
      }
    }
  }, [setOnlineUsers, clearCurrentServer, refreshServers, refreshMembers, currentServer])

  useWebSocket(currentServer?.id ?? null, handleWSMessage)

  return (
    <div className="flex h-screen w-screen overflow-hidden">
      <ServerSidebar />
      {currentServer ? (
        <>
          <ChannelSidebar />
          {currentChannel ? (
            <>
              <MessageArea />
              <MemberList />
            </>
          ) : (
            <div className="flex-1 bg-dark-tertiary flex items-center justify-center">
              <div className="text-center text-text-muted">
                <p className="text-lg">Select a channel to start chatting</p>
              </div>
            </div>
          )}
        </>
      ) : (
        <div className="flex-1 bg-dark-tertiary flex items-center justify-center">
          <div className="text-center text-text-muted">
            <p className="text-xl font-semibold mb-2">Welcome to Newcord</p>
            <p>Select a server from the sidebar or create a new one</p>
          </div>
        </div>
      )}
    </div>
  )
}
