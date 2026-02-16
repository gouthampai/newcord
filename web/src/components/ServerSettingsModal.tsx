import { useState } from 'react'
import { X, Hash, Trash2, Pencil, Check } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import * as api from '../services/api'

type Section = 'channels' | 'danger'

export default function ServerSettingsModal({ onClose }: { onClose: () => void }) {
  const { currentServer, channels, currentChannel, selectChannel, refreshChannels, refreshServers, clearCurrentServer } = useApp()
  const [section, setSection] = useState<Section>('channels')
  const [editingChannelId, setEditingChannelId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [confirmName, setConfirmName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  if (!currentServer) return null

  const textChannels = channels.filter(c => c.type === 'text' || !c.type)
  const voiceChannels = channels.filter(c => c.type === 'voice')
  const allChannels = [...textChannels, ...voiceChannels]

  const startEditing = (channelId: string, name: string) => {
    setEditingChannelId(channelId)
    setEditName(name)
    setError('')
  }

  const cancelEditing = () => {
    setEditingChannelId(null)
    setEditName('')
  }

  const saveChannelName = async (channelId: string) => {
    const trimmed = editName.trim().toLowerCase().replace(/\s+/g, '-')
    if (!trimmed) return
    setLoading(true)
    setError('')
    try {
      await api.updateChannel(channelId, { name: trimmed })
      await refreshChannels(currentServer.id)
      cancelEditing()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to rename channel')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteChannel = async (channelId: string) => {
    setLoading(true)
    setError('')
    try {
      await api.deleteChannel(channelId)
      if (currentChannel?.id === channelId) {
        const remaining = channels.filter(c => c.id !== channelId && (c.type === 'text' || !c.type))
        if (remaining.length > 0) {
          selectChannel(remaining[0])
        }
      }
      await refreshChannels(currentServer.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete channel')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteServer = async () => {
    if (confirmName !== currentServer.name) return
    setLoading(true)
    setError('')
    try {
      await api.deleteServer(currentServer.id)
      clearCurrentServer()
      await refreshServers()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete server')
      setLoading(false)
    }
  }

  const sidebarItems: { key: Section; label: string }[] = [
    { key: 'channels', label: 'Channels' },
    { key: 'danger', label: 'Danger Zone' },
  ]

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-dark-secondary rounded-md w-full max-w-lg" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="flex justify-between items-center p-4 border-b border-dark-primary">
          <h2 className="text-lg font-bold text-text-primary">Server Settings</h2>
          <X size={24} className="text-text-muted hover:text-text-primary cursor-pointer" onClick={onClose} />
        </div>

        <div className="flex min-h-[300px]">
          {/* Sidebar nav */}
          <div className="w-40 p-2 border-r border-dark-primary shrink-0">
            {sidebarItems.map(item => (
              <div
                key={item.key}
                onClick={() => { setSection(item.key); setError('') }}
                className={`px-3 py-1.5 rounded-[4px] text-sm cursor-pointer mb-0.5
                  ${section === item.key
                    ? 'bg-dark-active text-text-primary'
                    : 'text-text-muted hover:bg-dark-hover hover:text-text-secondary'
                  }
                  ${item.key === 'danger' ? 'text-red' : ''}`}
              >
                {item.label}
              </div>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 p-4 overflow-y-auto">
            {error && <p className="text-red text-sm mb-3">{error}</p>}

            {section === 'channels' && (
              <div>
                <h3 className="text-xs font-bold uppercase text-text-muted mb-3 tracking-wide">Manage Channels</h3>
                {allChannels.length === 0 && (
                  <p className="text-sm text-text-muted">No channels yet.</p>
                )}
                <div className="space-y-1">
                  {allChannels.map(channel => (
                    <div
                      key={channel.id}
                      className="flex items-center gap-2 px-2 py-1.5 rounded-[4px] group hover:bg-dark-hover"
                    >
                      <Hash size={16} className="text-text-muted shrink-0" />
                      {editingChannelId === channel.id ? (
                        <input
                          type="text"
                          value={editName}
                          onChange={e => setEditName(e.target.value)}
                          onKeyDown={e => {
                            if (e.key === 'Enter') saveChannelName(channel.id)
                            if (e.key === 'Escape') cancelEditing()
                          }}
                          autoFocus
                          className="flex-1 bg-dark-primary text-text-primary text-sm px-2 py-0.5 rounded-[3px] border-0"
                        />
                      ) : (
                        <span className="flex-1 text-sm text-text-secondary truncate">{channel.name}</span>
                      )}
                      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        {editingChannelId === channel.id ? (
                          <Check
                            size={14}
                            className="text-green cursor-pointer"
                            onClick={() => saveChannelName(channel.id)}
                          />
                        ) : (
                          <Pencil
                            size={14}
                            className="text-text-muted hover:text-text-primary cursor-pointer"
                            onClick={() => startEditing(channel.id, channel.name)}
                          />
                        )}
                        <Trash2
                          size={14}
                          className="text-text-muted hover:text-red cursor-pointer"
                          onClick={() => handleDeleteChannel(channel.id)}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {section === 'danger' && (
              <div>
                <h3 className="text-xs font-bold uppercase text-red mb-3 tracking-wide">Danger Zone</h3>
                <div className="border border-red/30 rounded-md p-4 bg-red/5">
                  <p className="text-sm text-text-secondary mb-3">
                    Deleting <strong className="text-text-primary">{currentServer.name}</strong> is permanent and cannot be undone. All channels and messages will be lost.
                  </p>
                  <label className="block text-xs font-bold text-text-muted uppercase mb-2">
                    Type the server name to confirm
                  </label>
                  <input
                    type="text"
                    value={confirmName}
                    onChange={e => setConfirmName(e.target.value)}
                    placeholder={currentServer.name}
                    className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-sm mb-3"
                  />
                  <button
                    onClick={handleDeleteServer}
                    disabled={confirmName !== currentServer.name || loading}
                    className="px-4 py-2 bg-red hover:bg-red/80 text-white font-medium rounded-[3px] transition-colors disabled:opacity-50 border-0 cursor-pointer text-sm"
                  >
                    {loading ? 'Deleting...' : 'Delete Server'}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
