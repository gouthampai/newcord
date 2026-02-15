import { useState } from 'react'
import { X, Hash, Volume2 } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import * as api from '../services/api'

export default function CreateChannelModal({ onClose }: { onClose: () => void }) {
  const { currentServer, refreshChannels } = useApp()
  const [name, setName] = useState('')
  const [type, setType] = useState<'text' | 'voice'>('text')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !currentServer) { setError('Channel name is required'); return }
    setError('')
    setLoading(true)
    try {
      await api.createChannel({ server_id: currentServer.id, name: name.trim().toLowerCase().replace(/\s+/g, '-'), type })
      await refreshChannels(currentServer.id)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create channel')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-dark-secondary rounded-md w-full max-w-md" onClick={e => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold text-text-primary">Create Channel</h2>
            <X size={24} className="text-text-muted hover:text-text-primary cursor-pointer" onClick={onClose} />
          </div>

          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Channel Type</label>
              <div className="space-y-2">
                <label
                  className={`flex items-center gap-3 p-3 rounded-[4px] cursor-pointer transition-colors
                    ${type === 'text' ? 'bg-dark-active' : 'bg-dark-primary hover:bg-dark-hover'}`}
                >
                  <input type="radio" name="type" checked={type === 'text'} onChange={() => setType('text')} className="hidden" />
                  <Hash size={20} className="text-text-muted" />
                  <div>
                    <div className="text-sm font-medium text-text-primary">Text</div>
                    <div className="text-xs text-text-muted">Send messages, images, and more</div>
                  </div>
                </label>
                <label
                  className={`flex items-center gap-3 p-3 rounded-[4px] cursor-pointer transition-colors
                    ${type === 'voice' ? 'bg-dark-active' : 'bg-dark-primary hover:bg-dark-hover'}`}
                >
                  <input type="radio" name="type" checked={type === 'voice'} onChange={() => setType('voice')} className="hidden" />
                  <Volume2 size={20} className="text-text-muted" />
                  <div>
                    <div className="text-sm font-medium text-text-primary">Voice</div>
                    <div className="text-xs text-text-muted">Hang out together with voice and video</div>
                  </div>
                </label>
              </div>
            </div>

            <div className="mb-4">
              <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Channel Name</label>
              <div className="flex items-center bg-dark-primary rounded-[3px]">
                {type === 'text' ? <Hash size={16} className="text-text-muted ml-3" /> : <Volume2 size={16} className="text-text-muted ml-3" />}
                <input
                  type="text" value={name} onChange={e => setName(e.target.value)} required autoFocus
                  className="flex-1 px-2 py-2 bg-transparent text-text-primary border-0 text-base"
                  placeholder="new-channel"
                  maxLength={100}
                />
              </div>
            </div>

            {error && <p className="text-red text-sm mb-3">{error}</p>}

            <div className="flex justify-end gap-3">
              <button type="button" onClick={onClose} className="px-4 py-2 text-text-primary bg-transparent border-0 cursor-pointer hover:underline text-sm">
                Cancel
              </button>
              <button
                type="submit" disabled={loading}
                className="px-6 py-2.5 bg-blurple hover:bg-blurple-hover text-white font-medium rounded-[3px] transition-colors disabled:opacity-50 border-0 cursor-pointer text-sm"
              >
                {loading ? 'Creating...' : 'Create Channel'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
