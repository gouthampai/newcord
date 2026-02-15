import { useState } from 'react'
import { X } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import * as api from '../services/api'

export default function CreateServerModal({ onClose }: { onClose: () => void }) {
  const { addServer, selectServer } = useApp()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) { setError('Server name is required'); return }
    setError('')
    setLoading(true)
    try {
      const server = await api.createServer({ name: name.trim(), description: description.trim() })
      addServer(server)
      await selectServer(server)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create server')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-dark-secondary rounded-md w-full max-w-md p-0" onClick={e => e.stopPropagation()}>
        <div className="p-6 pb-0">
          <div className="flex justify-between items-start mb-4">
            <div>
              <h2 className="text-2xl font-bold text-text-primary text-center w-full">Create a server</h2>
              <p className="text-text-muted text-sm text-center mt-1">Give your new server a personality with a name and description.</p>
            </div>
            <X size={24} className="text-text-muted hover:text-text-primary cursor-pointer shrink-0" onClick={onClose} />
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-6 pt-2">
          <div className="mb-4">
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Server Name</label>
            <input
              type="text" value={name} onChange={e => setName(e.target.value)} required autoFocus
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
              placeholder="My Awesome Server"
              maxLength={100}
            />
          </div>
          <div className="mb-4">
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Description <span className="text-text-muted font-normal normal-case">- Optional</span></label>
            <textarea
              value={description} onChange={e => setDescription(e.target.value)}
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base resize-none h-20"
              placeholder="What's your server about?"
              maxLength={1024}
            />
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
              {loading ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
