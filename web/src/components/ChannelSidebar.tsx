import { useState } from 'react'
import { Hash, Volume2, Plus, ChevronDown, Settings, UserPlus } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import { useAuth } from '../contexts/AuthContext'
import CreateChannelModal from './CreateChannelModal'
import UserSettingsPanel from './UserSettingsPanel'
import InviteModal from './InviteModal'
import ServerSettingsModal from './ServerSettingsModal'

export default function ChannelSidebar() {
  const { currentServer, channels, currentChannel, selectChannel } = useApp()
  const { user } = useAuth()
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [showInvite, setShowInvite] = useState(false)
  const [showServerSettings, setShowServerSettings] = useState(false)
  const isOwner = currentServer?.owner_id === user?.id

  if (!currentServer) return null

  const textChannels = channels.filter(c => c.type === 'text' || !c.type)
  const voiceChannels = channels.filter(c => c.type === 'voice')

  return (
    <>
      <div className="w-60 bg-dark-secondary flex flex-col shrink-0">
        {/* Server header */}
        <div
          className="h-12 px-4 flex items-center border-b border-dark-primary shadow-sm font-semibold text-text-primary cursor-pointer hover:bg-dark-hover"
          onClick={isOwner ? () => setShowServerSettings(true) : undefined}
        >
          <span className="truncate flex-1">{currentServer.name}</span>
          {isOwner
            ? <Settings size={16} className="text-text-muted" />
            : <ChevronDown size={16} className="text-text-muted" />}
        </div>

        {/* Invite button */}
        <div className="px-2 pt-3 pb-1">
          <button
            onClick={() => setShowInvite(true)}
            className="w-full flex items-center gap-2 px-2 py-1.5 rounded-[4px] text-text-muted hover:bg-dark-hover hover:text-text-secondary text-sm border-0 bg-transparent cursor-pointer"
          >
            <UserPlus size={16} />
            <span>Invite People</span>
          </button>
        </div>

        {/* Channels list */}
        <div className="flex-1 overflow-y-auto px-2 pt-2">
          {/* Text channels */}
          <div className="mb-4">
            <div className="flex items-center justify-between px-1 mb-1">
              <span className="text-xs font-bold uppercase text-text-muted tracking-wide">Text Channels</span>
              <Plus
                size={16}
                className="text-text-muted hover:text-text-primary cursor-pointer"
                onClick={() => setShowCreateChannel(true)}
              />
            </div>
            {textChannels.map(channel => (
              <div
                key={channel.id}
                onClick={() => selectChannel(channel)}
                className={`flex items-center gap-1.5 px-2 py-1.5 rounded-[4px] cursor-pointer group mb-0.5
                  ${currentChannel?.id === channel.id
                    ? 'bg-dark-active text-text-primary'
                    : 'text-text-muted hover:bg-dark-hover hover:text-text-secondary'
                  }`}
              >
                <Hash size={18} className="shrink-0 opacity-70" />
                <span className="truncate text-sm">{channel.name}</span>
              </div>
            ))}
          </div>

          {/* Voice channels */}
          {voiceChannels.length > 0 && (
            <div className="mb-4">
              <div className="flex items-center justify-between px-1 mb-1">
                <span className="text-xs font-bold uppercase text-text-muted tracking-wide">Voice Channels</span>
              </div>
              {voiceChannels.map(channel => (
                <div
                  key={channel.id}
                  className="flex items-center gap-1.5 px-2 py-1.5 rounded-[4px] cursor-pointer text-text-muted hover:bg-dark-hover hover:text-text-secondary mb-0.5"
                >
                  <Volume2 size={18} className="shrink-0 opacity-70" />
                  <span className="truncate text-sm">{channel.name}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* User panel */}
        <div className="h-[52px] bg-dark-primary/50 px-2 flex items-center gap-2">
          <div className="relative">
            <div className="w-8 h-8 rounded-full bg-blurple flex items-center justify-center text-white text-xs font-bold">
              {user?.display_name?.[0]?.toUpperCase() || user?.username?.[0]?.toUpperCase() || '?'}
            </div>
            <div className="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-dark-primary bg-green" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-text-primary truncate">{user?.display_name || user?.username}</div>
            <div className="text-xs text-text-muted truncate">Online</div>
          </div>
          <Settings
            size={18}
            className="text-text-muted hover:text-text-primary cursor-pointer shrink-0"
            onClick={() => setShowSettings(true)}
          />
        </div>
      </div>

      {showCreateChannel && <CreateChannelModal onClose={() => setShowCreateChannel(false)} />}
      {showSettings && <UserSettingsPanel onClose={() => setShowSettings(false)} />}
      {showInvite && <InviteModal serverId={currentServer.id} onClose={() => setShowInvite(false)} />}
      {showServerSettings && <ServerSettingsModal onClose={() => setShowServerSettings(false)} />}
    </>
  )
}
