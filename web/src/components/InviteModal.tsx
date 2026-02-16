import { useState, useEffect } from 'react'
import { X, Copy, Check } from 'lucide-react'
import * as api from '../services/api'

interface Props {
  serverId: string
  onClose: () => void
}

export default function InviteModal({ serverId, onClose }: Props) {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    const generate = async () => {
      setLoading(true)
      try {
        const invite = await api.createInvite(serverId)
        setCode(invite.code)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to create invite')
      } finally {
        setLoading(false)
      }
    }
    generate()
  }, [serverId])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-dark-secondary rounded-md w-full max-w-md p-0" onClick={e => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex justify-between items-start mb-4">
            <h2 className="text-xl font-bold text-text-primary">Invite People</h2>
            <X size={24} className="text-text-muted hover:text-text-primary cursor-pointer shrink-0" onClick={onClose} />
          </div>

          <p className="text-text-secondary text-sm mb-4">Share this invite code with others to let them join your server.</p>

          {loading && <p className="text-text-muted text-sm">Generating invite code...</p>}
          {error && <p className="text-red text-sm">{error}</p>}

          {code && (
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={code}
                className="flex-1 px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base font-mono tracking-wider"
              />
              <button
                onClick={handleCopy}
                className="px-4 py-2 bg-blurple hover:bg-blurple-hover text-white font-medium rounded-[3px] transition-colors border-0 cursor-pointer flex items-center gap-1.5 text-sm"
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
