import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import { AppProvider } from './contexts/AppContext'
import AuthPage from './pages/AuthPage'
import MainLayout from './pages/MainLayout'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="flex items-center justify-center h-screen bg-dark-tertiary text-text-primary">Loading...</div>
  return user ? <>{children}</> : <Navigate to="/login" />
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="flex items-center justify-center h-screen bg-dark-tertiary text-text-primary">Loading...</div>
  return user ? <Navigate to="/channels" /> : <>{children}</>
}

export default function App() {
  return (
    <AuthProvider>
      <AppProvider>
        <Routes>
          <Route path="/login" element={<PublicRoute><AuthPage /></PublicRoute>} />
          <Route path="/channels/*" element={<ProtectedRoute><MainLayout /></ProtectedRoute>} />
          <Route path="*" element={<Navigate to="/channels" />} />
        </Routes>
      </AppProvider>
    </AuthProvider>
  )
}
