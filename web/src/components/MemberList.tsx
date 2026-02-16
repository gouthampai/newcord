import { useEffect, useState } from 'react'
import { useApp } from '../contexts/AppContext'
import { useAuth } from '../contexts/AuthContext'
import type { User } from '../types'
import * as api from '../services/api'

export default function MemberList() {
  const { members, onlineUsers } = useApp()
  const { user: currentUser } = useAuth()
  const [users, setUsers] = useState<Record<string, User>>({})

  useEffect(() => {
    members.forEach(m => {
      if (!users[m.user_id]) {
        api.getUser(m.user_id).then(u => {
          setUsers(prev => ({ ...prev, [m.user_id]: u }))
        }).catch(() => {})
      }
    })
  }, [members]) // eslint-disable-line react-hooks/exhaustive-deps

  const isOnline = (userId: string) => onlineUsers.has(userId) || userId === currentUser?.id

  const sortOnlineFirst = (a: typeof members[0], b: typeof members[0]) => {
    const aOnline = isOnline(a.user_id) ? 0 : 1
    const bOnline = isOnline(b.user_id) ? 0 : 1
    return aOnline - bOnline
  }

  const roleGroups = {
    owner: members.filter(m => m.role === 'owner').sort(sortOnlineFirst),
    admin: members.filter(m => m.role === 'admin').sort(sortOnlineFirst),
    moderator: members.filter(m => m.role === 'moderator').sort(sortOnlineFirst),
    member: members.filter(m => m.role === 'member').sort(sortOnlineFirst),
  }

  const roleLabels: Record<string, string> = {
    owner: 'Owner',
    admin: 'Admin',
    moderator: 'Moderator',
    member: 'Member',
  }

  return (
    <div className="w-60 bg-dark-secondary shrink-0 overflow-y-auto">
      <div className="px-4 pt-6">
        {Object.entries(roleGroups).map(([role, members]) =>
          members.length > 0 ? (
            <div key={role} className="mb-4">
              <h3 className="text-xs font-bold uppercase text-text-muted mb-2 tracking-wide">
                {roleLabels[role]} &mdash; {members.length}
              </h3>
              {members.map(m => {
                const u = users[m.user_id]
                return (
                  <div
                    key={m.user_id}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-[4px] hover:bg-dark-hover cursor-pointer mb-0.5"
                  >
                    <div className="relative">
                      <div className="w-8 h-8 rounded-full bg-blurple flex items-center justify-center text-white text-xs font-bold">
                        {(u?.display_name || u?.username || '?')[0]?.toUpperCase()}
                      </div>
                      <div className={`absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-dark-secondary
                        ${isOnline(m.user_id) ? 'bg-green' : 'bg-text-muted'}`}
                      />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="text-sm text-text-secondary truncate">
                        {m.nickname || u?.display_name || u?.username || 'Unknown'}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : null
        )}
      </div>
    </div>
  )
}
