import { useCallback } from 'react'
import { useApp } from '../contexts/AppContext'
import { useWebSocket } from '../hooks/useWebSocket'
import type { WSMessage, Message } from '../types'
import ServerSidebar from '../components/ServerSidebar'
import ChannelSidebar from '../components/ChannelSidebar'
import MessageArea from '../components/MessageArea'
import MemberList from '../components/MemberList'

export default function MainLayout() {
  const { currentServer, currentChannel } = useApp()

  const handleWSMessage = useCallback((_msg: WSMessage) => {
    // WebSocket messages are handled by MessageArea via a shared callback
  }, [])

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
