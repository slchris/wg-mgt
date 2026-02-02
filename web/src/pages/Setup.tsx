import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { authService } from '../services/auth'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'

export default function Setup() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    authService
      .checkSetup()
      .then((data) => {
        if (!data.needs_setup) {
          navigate('/login')
        }
      })
      .catch(() => {
        navigate('/login')
      })
      .finally(() => {
        setChecking(false)
      })
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    setLoading(true)

    try {
      await authService.setup(username, password)
      navigate('/login')
    } catch {
      setError('Failed to create account')
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
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
            <h1 className="text-2xl font-semibold text-apple-gray-500">Welcome to WG MGT</h1>
            <p className="text-apple-gray-300 mt-2">Create your admin account to get started</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Choose a username"
              required
            />

            <Input
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Choose a strong password"
              helperText="At least 8 characters"
              required
            />

            <Input
              label="Confirm Password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your password"
              required
            />

            {error && (
              <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>
            )}

            <Button type="submit" className="w-full" loading={loading}>
              Create Account
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
