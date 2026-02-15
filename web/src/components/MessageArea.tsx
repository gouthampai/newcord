import { useState, useEffect, useRef, useCallback } from 'react'
import { Hash, Send } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import { useAuth } from '../contexts/AuthContext'
import { useWebSocket } from '../hooks/useWebSocket'
import type { Message, WSMessage } from '../types'
import * as api from '../services/api'

function formatTime(dateStr: string) {
  const d = new Date(dateStr)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (isToday) return `Today at ${time}`
  return `${d.toLocaleDateString()} ${time}`
}

export default function MessageArea() {
  const { currentServer, currentChannel } = useApp()
  const { user } = useAuth()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [userCache, setUserCache] = useState<Record<string, { display_name: string; username: string }>>({})

  const handleWSMessage = useCallback((msg: WSMessage) => {
    if (msg.type === 'message' && msg.channel_id === currentChannel?.id) {
      const data = msg.data as Record<string, string>
      const newMsg: Message = {
        id: crypto.randomUUID(),
        channel_id: msg.channel_id!,
        user_id: data.user_id || '',
        content: data.content || '',
        type: (data.type as Message['type']) || 'text',
        attachments: [],
        created_at: msg.timestamp,
      }
      setMessages(prev => [...prev, newMsg])
    }
  }, [currentChannel?.id])

  useWebSocket(currentServer?.id ?? null, handleWSMessage)

  useEffect(() => {
    if (!currentChannel) return
    setLoading(true)
    api.getMessages(currentChannel.id).then(msgs => {
      setMessages((msgs || []).reverse())
      // Fetch display names for unique user IDs
      const userIds = [...new Set((msgs || []).map(m => m.user_id))]
      userIds.forEach(uid => {
        if (!userCache[uid]) {
          api.getUser(uid).then(u => {
            setUserCache(prev => ({ ...prev, [uid]: { display_name: u.display_name, username: u.username } }))
          }).catch(() => {})
        }
      })
    }).catch(() => setMessages([])).finally(() => setLoading(false))
  }, [currentChannel?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || !currentChannel) return
    const content = input.trim()
    setInput('')
    try {
      const msg = await api.createMessage(currentChannel.id, { content })
      setMessages(prev => [...prev, msg])
    } catch { /* message failed */ }
  }

  const getDisplayName = (userId: string) => {
    if (userId === user?.id) return user.display_name || user.username
    const cached = userCache[userId]
    return cached?.display_name || cached?.username || 'Unknown User'
  }

  if (!currentChannel) return null

  return (
    <div className="flex-1 bg-dark-tertiary flex flex-col min-w-0">
      {/* Channel header */}
      <div className="h-12 px-4 flex items-center border-b border-dark-primary shadow-sm shrink-0">
        <Hash size={20} className="text-text-muted mr-2" />
        <span className="font-semibold text-text-primary">{currentChannel.name}</span>
        {currentChannel.description && (
          <>
            <div className="w-px h-5 bg-dark-active mx-3" />
            <span className="text-sm text-text-muted truncate">{currentChannel.description}</span>
          </>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-4">
        {loading ? (
          <div className="text-text-muted text-center py-8">Loading messages...</div>
        ) : messages.length === 0 ? (
          <div className="text-center py-8">
            <Hash size={40} className="text-text-muted mx-auto mb-2" />
            <h3 className="text-xl font-bold text-text-primary mb-1">Welcome to #{currentChannel.name}!</h3>
            <p className="text-text-muted text-sm">This is the start of the channel.</p>
          </div>
        ) : (
          messages.map((msg, i) => {
            const prevMsg = messages[i - 1]
            const showHeader = !prevMsg || prevMsg.user_id !== msg.user_id ||
              new Date(msg.created_at).getTime() - new Date(prevMsg.created_at).getTime() > 5 * 60 * 1000

            return (
              <div key={msg.id} className={`group hover:bg-dark-hover/30 px-2 rounded ${showHeader ? 'mt-4 pt-1' : 'py-0'}`}>
                {showHeader ? (
                  <div className="flex items-start gap-3">
                    <div className="w-10 h-10 rounded-full bg-blurple flex items-center justify-center text-white text-sm font-bold shrink-0 mt-0.5">
                      {getDisplayName(msg.user_id)[0]?.toUpperCase() || '?'}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-baseline gap-2">
                        <span className="font-medium text-text-primary hover:underline cursor-pointer">
                          {getDisplayName(msg.user_id)}
                        </span>
                        <span className="text-xs text-text-muted">{formatTime(msg.created_at)}</span>
                        {msg.edited_at && <span className="text-xs text-text-muted">(edited)</span>}
                      </div>
                      <p className="text-text-secondary text-[15px] leading-relaxed break-words">{msg.content}</p>
                    </div>
                  </div>
                ) : (
                  <div className="pl-[52px]">
                    <p className="text-text-secondary text-[15px] leading-relaxed break-words">{msg.content}</p>
                  </div>
                )}
              </div>
            )
          })
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Message input */}
      <div className="px-4 pb-6 shrink-0">
        <form onSubmit={handleSend} className="flex items-center bg-dark-active rounded-lg px-4">
          <input
            type="text"
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder={`Message #${currentChannel.name}`}
            className="flex-1 bg-transparent text-text-primary py-3 text-[15px] placeholder:text-text-muted border-0"
            maxLength={2000}
          />
          <button
            type="submit"
            disabled={!input.trim()}
            className="text-text-muted hover:text-text-primary disabled:opacity-30 bg-transparent border-0 cursor-pointer p-1"
          >
            <Send size={20} />
          </button>
        </form>
      </div>
    </div>
  )
}
