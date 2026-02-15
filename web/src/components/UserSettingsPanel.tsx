import { useState } from 'react'
import { X } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export default function UserSettingsPanel({ onClose }: { onClose: () => void }) {
  const { user, updateUser, logout } = useAuth()
  const [displayName, setDisplayName] = useState(user?.display_name || '')
  const [bio, setBio] = useState(user?.bio || '')
  const [status, setStatus] = useState(user?.status || 'online')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  const handleSave = async () => {
    setError('')
    setSaving(true)
    setSaved(false)
    try {
      await updateUser({ display_name: displayName, bio, status: status as 'online' | 'offline' | 'away' | 'dnd' })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-dark-tertiary rounded-md w-full max-w-lg" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-6 pb-0">
          <h2 className="text-xl font-bold text-text-primary">User Settings</h2>
          <X size={24} className="text-text-muted hover:text-text-primary cursor-pointer" onClick={onClose} />
        </div>

        <div className="p-6 space-y-4">
          {/* Profile card preview */}
          <div className="bg-dark-secondary rounded-lg overflow-hidden">
            <div className="h-16 bg-blurple" />
            <div className="px-4 pb-4 -mt-8">
              <div className="w-16 h-16 rounded-full bg-blurple border-4 border-dark-secondary flex items-center justify-center text-white text-xl font-bold">
                {(displayName || user?.username || '?')[0]?.toUpperCase()}
              </div>
              <div className="mt-2 font-bold text-text-primary">{displayName || user?.username}</div>
              <div className="text-sm text-text-muted">{user?.username}</div>
            </div>
          </div>

          <div>
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Display Name</label>
            <input
              type="text" value={displayName} onChange={e => setDisplayName(e.target.value)}
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
              maxLength={64}
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">About Me</label>
            <textarea
              value={bio} onChange={e => setBio(e.target.value)}
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base resize-none h-20"
              maxLength={256}
              placeholder="Tell the world about yourself"
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Status</label>
            <div className="grid grid-cols-2 gap-2">
              {(['online', 'away', 'dnd', 'offline'] as const).map(s => (
                <label
                  key={s}
                  className={`flex items-center gap-2 p-2.5 rounded-[4px] cursor-pointer transition-colors
                    ${status === s ? 'bg-dark-active' : 'bg-dark-primary hover:bg-dark-hover'}`}
                >
                  <input type="radio" name="status" checked={status === s} onChange={() => setStatus(s)} className="hidden" />
                  <div className={`w-3 h-3 rounded-full ${s === 'online' ? 'bg-green' : s === 'away' ? 'bg-yellow' : s === 'dnd' ? 'bg-red' : 'bg-text-muted'}`} />
                  <span className="text-sm text-text-primary capitalize">{s === 'dnd' ? 'Do Not Disturb' : s}</span>
                </label>
              ))}
            </div>
          </div>

          {error && <p className="text-red text-sm">{error}</p>}
          {saved && <p className="text-green text-sm">Settings saved!</p>}

          <div className="flex justify-between pt-2">
            <button
              onClick={logout}
              className="px-4 py-2 bg-red/20 text-red hover:bg-red/30 rounded-[3px] transition-colors border-0 cursor-pointer text-sm font-medium"
            >
              Log Out
            </button>
            <div className="flex gap-3">
              <button type="button" onClick={onClose} className="px-4 py-2 text-text-primary bg-transparent border-0 cursor-pointer hover:underline text-sm">
                Cancel
              </button>
              <button
                onClick={handleSave} disabled={saving}
                className="px-6 py-2 bg-blurple hover:bg-blurple-hover text-white font-medium rounded-[3px] transition-colors disabled:opacity-50 border-0 cursor-pointer text-sm"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
