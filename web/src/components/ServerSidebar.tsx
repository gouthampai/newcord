import { useState } from 'react'
import { Plus, LogOut } from 'lucide-react'
import { useApp } from '../contexts/AppContext'
import { useAuth } from '../contexts/AuthContext'
import CreateServerModal from './CreateServerModal'

export default function ServerSidebar() {
  const { servers, currentServer, selectServer } = useApp()
  const { logout } = useAuth()
  const [showCreateModal, setShowCreateModal] = useState(false)

  return (
    <>
      <div className="w-[72px] bg-dark-primary flex flex-col items-center py-3 gap-2 shrink-0 overflow-y-auto">
        {/* Home button */}
        <div className="w-12 h-12 bg-dark-tertiary rounded-full flex items-center justify-center text-text-primary font-bold hover:bg-blurple hover:rounded-2xl transition-all cursor-pointer mb-2 text-lg">
          N
        </div>

        <div className="w-8 h-0.5 bg-dark-active rounded-full mb-1" />

        {/* Server list */}
        {servers.map(server => (
          <div
            key={server.id}
            onClick={() => selectServer(server)}
            className={`w-12 h-12 flex items-center justify-center cursor-pointer transition-all text-sm font-medium
              ${currentServer?.id === server.id
                ? 'bg-blurple rounded-2xl text-white'
                : 'bg-dark-tertiary rounded-full text-text-primary hover:bg-blurple hover:rounded-2xl hover:text-white'
              }`}
            title={server.name}
          >
            {server.icon_url ? (
              <img src={server.icon_url} alt="" className="w-full h-full rounded-inherit object-cover" />
            ) : (
              server.name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
            )}
          </div>
        ))}

        {/* Add server button */}
        <div
          onClick={() => setShowCreateModal(true)}
          className="w-12 h-12 bg-dark-tertiary rounded-full flex items-center justify-center text-green hover:bg-green hover:text-white hover:rounded-2xl transition-all cursor-pointer"
        >
          <Plus size={20} />
        </div>

        <div className="flex-1" />

        {/* Logout button */}
        <div
          onClick={logout}
          className="w-12 h-12 bg-dark-tertiary rounded-full flex items-center justify-center text-red hover:bg-red hover:text-white hover:rounded-2xl transition-all cursor-pointer"
          title="Log out"
        >
          <LogOut size={18} />
        </div>
      </div>

      {showCreateModal && <CreateServerModal onClose={() => setShowCreateModal(false)} />}
    </>
  )
}
