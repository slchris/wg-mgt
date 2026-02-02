import { useState } from 'react'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import { authService } from '../services/auth'

export default function Settings() {
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setMessage('')

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match')
      return
    }

    if (newPassword.length < 8) {
      setError('New password must be at least 8 characters')
      return
    }

    setLoading(true)

    try {
      await authService.changePassword(oldPassword, newPassword)
      setMessage('Password changed successfully')
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch {
      setError('Failed to change password. Please check your old password.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-apple-gray-500">Settings</h1>
        <p className="text-apple-gray-300 mt-1">Manage your account settings</p>
      </div>

      <Card title="Change Password" description="Update your account password">
        <form onSubmit={handleChangePassword} className="space-y-4 max-w-md">
          <Input
            label="Current Password"
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            placeholder="Enter current password"
            required
          />
          <Input
            label="New Password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Enter new password"
            helperText="At least 8 characters"
            required
          />
          <Input
            label="Confirm New Password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Confirm new password"
            required
          />

          {error && <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>}
          {message && (
            <div className="p-3 rounded-apple bg-green-50 text-apple-green text-sm">{message}</div>
          )}

          <Button type="submit" loading={loading}>
            Change Password
          </Button>
        </form>
      </Card>

      <Card title="About" description="Application information">
        <div className="space-y-2 text-sm">
          <p className="text-apple-gray-400">
            <span className="font-medium">Version:</span> 1.0.0
          </p>
          <p className="text-apple-gray-400">
            <span className="font-medium">WG MGT</span> - WireGuard Management Tool
          </p>
          <p className="text-apple-gray-300 mt-4">
            A simple and efficient tool for managing WireGuard VPN networks across multiple servers.
          </p>
        </div>
      </Card>
    </div>
  )
}
