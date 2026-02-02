import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'
import { authService } from '../services/auth'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'

export default function Login() {
  const navigate = useNavigate()
  const { token, setAuth } = useAuthStore()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [checkingSetup, setCheckingSetup] = useState(true)

  // Redirect when token is set
  useEffect(() => {
    if (token) {
      navigate('/', { replace: true })
    }
  }, [token, navigate])

  useEffect(() => {
    authService
      .checkSetup()
      .then((data) => {
        if (data.needs_setup) {
          navigate('/setup')
        }
      })
      .catch(() => {
        // Ignore errors
      })
      .finally(() => {
        setCheckingSetup(false)
      })
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const data = await authService.login(username, password)
      setAuth(data.token)
      // Navigation will happen via useEffect when token changes
    } catch {
      setError('Invalid username or password')
    } finally {
      setLoading(false)
    }
  }

  if (checkingSetup) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-apple-gray-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-apple-blue"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-apple-gray-50 p-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-apple-lg shadow-apple p-8">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-semibold text-apple-gray-500">WG MGT</h1>
            <p className="text-apple-gray-300 mt-2">Sign in to your account</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              required
            />

            <Input
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
            />

            {error && (
              <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>
            )}

            <Button type="submit" className="w-full" loading={loading}>
              Sign In
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
