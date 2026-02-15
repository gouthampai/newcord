import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login, register } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (isLogin) {
        await login(email, password)
      } else {
        if (username.length < 3) { setError('Username must be at least 3 characters'); setLoading(false); return }
        if (password.length < 8) { setError('Password must be at least 8 characters'); setLoading(false); return }
        await register(username, email, password, displayName || username)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-dark-primary flex items-center justify-center p-4">
      <div className="bg-dark-tertiary rounded-md shadow-lg p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-text-primary text-center mb-2">
          {isLogin ? 'Welcome back!' : 'Create an account'}
        </h1>
        <p className="text-text-muted text-center mb-6 text-sm">
          {isLogin ? "We're so excited to see you again!" : 'We look forward to seeing you!'}
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          {!isLogin && (
            <>
              <div>
                <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Username</label>
                <input
                  type="text" value={username} onChange={e => setUsername(e.target.value)} required
                  className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
                />
              </div>
              <div>
                <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Display Name</label>
                <input
                  type="text" value={displayName} onChange={e => setDisplayName(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
                />
              </div>
            </>
          )}
          <div>
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Email</label>
            <input
              type="email" value={email} onChange={e => setEmail(e.target.value)} required
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
            />
          </div>
          <div>
            <label className="block text-xs font-bold text-text-secondary uppercase mb-2">Password</label>
            <input
              type="password" value={password} onChange={e => setPassword(e.target.value)} required
              className="w-full px-3 py-2 bg-dark-primary text-text-primary rounded-[3px] border-0 text-base"
            />
          </div>

          {error && <p className="text-red text-sm">{error}</p>}

          <button
            type="submit" disabled={loading}
            className="w-full py-2.5 bg-blurple hover:bg-blurple-hover text-white font-medium rounded-[3px] transition-colors disabled:opacity-50 text-base"
          >
            {loading ? 'Please wait...' : isLogin ? 'Log In' : 'Register'}
          </button>
        </form>

        <p className="text-sm text-text-muted mt-4">
          {isLogin ? "Need an account? " : "Already have an account? "}
          <button
            onClick={() => { setIsLogin(!isLogin); setError('') }}
            className="text-blurple hover:underline bg-transparent border-0 cursor-pointer p-0 text-sm"
          >
            {isLogin ? 'Register' : 'Log In'}
          </button>
        </p>
      </div>
    </div>
  )
}
