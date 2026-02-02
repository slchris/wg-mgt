import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Toast from '../components/ui/Toast'
import { nodeService, WireGuardStatus, SystemInfo } from '../services/nodes'
import type { Node } from '../types'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatHandshake(timestamp: string): string {
  if (!timestamp || timestamp === '0' || timestamp === '') return 'Never'
  const date = new Date(timestamp)
  if (isNaN(date.getTime())) return timestamp
  const now = new Date()
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000)
  if (diff < 60) return `${diff} seconds ago`
  if (diff < 3600) return `${Math.floor(diff / 60)} minutes ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)} hours ago`
  return `${Math.floor(diff / 86400)} days ago`
}

export default function NodeDetail() {
  const { id } = useParams<{ id: string }>()
  const [node, setNode] = useState<Node | null>(null)
  const [wgStatus, setWgStatus] = useState<WireGuardStatus | null>(null)
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingWg, setLoadingWg] = useState(false)
  const [loadingSys, setLoadingSys] = useState(false)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)

  const fetchNode = useCallback(async () => {
    if (!id) return
    try {
      const data = await nodeService.get(parseInt(id))
      setNode(data)
    } catch {
      setToast({ message: 'Failed to fetch node', type: 'error' })
    } finally {
      setLoading(false)
    }
  }, [id])

  const fetchWgStatus = async () => {
    if (!id) return
    setLoadingWg(true)
    try {
      const data = await nodeService.getWireGuardStatus(parseInt(id))
      setWgStatus(data)
      setToast({ message: 'WireGuard status loaded', type: 'success' })
    } catch {
      setToast({ message: 'Failed to fetch WireGuard status', type: 'error' })
    } finally {
      setLoadingWg(false)
    }
  }

  const fetchSystemInfo = async () => {
    if (!id) return
    setLoadingSys(true)
    try {
      const data = await nodeService.getSystemInfo(parseInt(id))
      setSystemInfo(data)
      setToast({ message: 'System info loaded', type: 'success' })
    } catch {
      setToast({ message: 'Failed to fetch system info', type: 'error' })
    } finally {
      setLoadingSys(false)
    }
  }

  useEffect(() => {
    fetchNode()
  }, [fetchNode])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-apple-blue"></div>
      </div>
    )
  }

  if (!node) {
    return (
      <div className="text-center py-12">
        <p className="text-apple-gray-400">Node not found</p>
        <Link to="/nodes" className="text-apple-blue hover:underline mt-2 inline-block">
          Back to Nodes
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center space-x-4">
            <Link to="/nodes" className="text-apple-gray-400 hover:text-apple-gray-500">
              ← Back
            </Link>
            <h1 className="text-2xl font-semibold text-apple-gray-500">{node.name}</h1>
            <span
              className={`px-2 py-1 text-xs font-medium rounded-full ${
                node.status === 'online'
                  ? 'bg-green-100 text-apple-green'
                  : node.status === 'offline'
                  ? 'bg-red-100 text-apple-red'
                  : 'bg-gray-100 text-apple-gray-400'
              }`}
            >
              {node.status || 'unknown'}
            </span>
          </div>
          <p className="text-apple-gray-300 mt-1">{node.host}</p>
        </div>
        <div className="flex space-x-3">
          <Button variant="secondary" onClick={fetchSystemInfo} loading={loadingSys}>
            System Info
          </Button>
          <Button onClick={fetchWgStatus} loading={loadingWg}>
            WireGuard Status
          </Button>
        </div>
      </div>

      {/* Node Info */}
      <Card>
        <h2 className="text-lg font-medium text-apple-gray-500 mb-4">Node Information</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <p className="text-sm text-apple-gray-300">Host</p>
            <p className="font-medium text-apple-gray-500">{node.host}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">SSH Port</p>
            <p className="font-medium text-apple-gray-500">{node.ssh_port}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">SSH User</p>
            <p className="font-medium text-apple-gray-500">{node.ssh_user}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">WireGuard Interface</p>
            <p className="font-medium text-apple-gray-500">{node.wg_interface}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">WG Port</p>
            <p className="font-medium text-apple-gray-500">{node.wg_port}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">WG Address</p>
            <p className="font-medium text-apple-gray-500">{node.wg_address}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">Endpoint</p>
            <p className="font-medium text-apple-gray-500">{node.endpoint || '-'}</p>
          </div>
          <div>
            <p className="text-sm text-apple-gray-300">Last Seen</p>
            <p className="font-medium text-apple-gray-500">
              {node.last_seen ? new Date(node.last_seen).toLocaleString() : 'Never'}
            </p>
          </div>
        </div>
      </Card>

      {/* System Info */}
      {systemInfo && (
        <Card>
          <h2 className="text-lg font-medium text-apple-gray-500 mb-4">System Information</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-apple-gray-300">Hostname</p>
              <p className="font-medium text-apple-gray-500">{systemInfo.hostname}</p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">OS</p>
              <p className="font-medium text-apple-gray-500">{systemInfo.os}</p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">Kernel</p>
              <p className="font-medium text-apple-gray-500">{systemInfo.kernel}</p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">WireGuard Installed</p>
              <p className="font-medium text-apple-gray-500">
                {systemInfo.wireguard_installed === 'true' ? (
                  <span className="text-apple-green">Yes</span>
                ) : (
                  <span className="text-apple-red">No</span>
                )}
              </p>
            </div>
            {systemInfo.wireguard_version && (
              <div>
                <p className="text-sm text-apple-gray-300">WireGuard Version</p>
                <p className="font-medium text-apple-gray-500">{systemInfo.wireguard_version}</p>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* WireGuard Status */}
      {wgStatus && (
        <Card>
          <h2 className="text-lg font-medium text-apple-gray-500 mb-4">WireGuard Status</h2>
          <div className="space-y-6">
            {/* Interface Info */}
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4 pb-4 border-b border-apple-gray-100">
              <div>
                <p className="text-sm text-apple-gray-300">Interface</p>
                <p className="font-medium text-apple-gray-500">{wgStatus.interface}</p>
              </div>
              <div>
                <p className="text-sm text-apple-gray-300">Address</p>
                <p className="font-medium text-apple-gray-500">{wgStatus.address || '-'}</p>
              </div>
              <div>
                <p className="text-sm text-apple-gray-300">Status</p>
                <p className={`font-medium ${wgStatus.is_running ? 'text-apple-green' : 'text-apple-red'}`}>
                  {wgStatus.is_running ? 'Running' : 'Stopped'}
                </p>
              </div>
              <div>
                <p className="text-sm text-apple-gray-300">Listen Port</p>
                <p className="font-medium text-apple-gray-500">{wgStatus.listen_port}</p>
              </div>
              <div>
                <p className="text-sm text-apple-gray-300">Public Key</p>
                <p className="font-mono text-xs text-apple-gray-500 truncate">{wgStatus.public_key}</p>
              </div>
            </div>

            {/* Transfer Stats */}
            <div className="grid grid-cols-2 gap-4 pb-4 border-b border-apple-gray-100">
              <div>
                <p className="text-sm text-apple-gray-300">Total Received</p>
                <p className="font-medium text-apple-gray-500">{formatBytes(wgStatus.total_rx)}</p>
              </div>
              <div>
                <p className="text-sm text-apple-gray-300">Total Sent</p>
                <p className="font-medium text-apple-gray-500">{formatBytes(wgStatus.total_tx)}</p>
              </div>
            </div>

            {/* Peers */}
            {wgStatus.peers && wgStatus.peers.length > 0 && (
              <div>
                <h3 className="text-md font-medium text-apple-gray-500 mb-3">Connected Peers ({wgStatus.peers.length})</h3>
                <div className="space-y-4">
                  {wgStatus.peers.map((peer, index) => (
                    <div key={index} className="bg-apple-gray-50 rounded-apple p-4">
                      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                        <div className="col-span-2 md:col-span-3">
                          <p className="text-sm text-apple-gray-300">Public Key</p>
                          <p className="font-mono text-xs text-apple-gray-500">{peer.public_key}</p>
                        </div>
                        <div>
                          <p className="text-sm text-apple-gray-300">Endpoint</p>
                          <p className="font-medium text-apple-gray-500">{peer.endpoint || '-'}</p>
                        </div>
                        <div>
                          <p className="text-sm text-apple-gray-300">Allowed IPs</p>
                          <p className="font-medium text-apple-gray-500">{peer.allowed_ips?.join(', ') || '-'}</p>
                        </div>
                        <div>
                          <p className="text-sm text-apple-gray-300">Latest Handshake</p>
                          <p className="font-medium text-apple-gray-500">{formatHandshake(peer.latest_handshake)}</p>
                        </div>
                        <div>
                          <p className="text-sm text-apple-gray-300">Transfer ↓</p>
                          <p className="font-medium text-apple-gray-500">{formatBytes(peer.transfer_rx)}</p>
                        </div>
                        <div>
                          <p className="text-sm text-apple-gray-300">Transfer ↑</p>
                          <p className="font-medium text-apple-gray-500">{formatBytes(peer.transfer_tx)}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {(!wgStatus.peers || wgStatus.peers.length === 0) && (
              <p className="text-apple-gray-300 text-center py-4">No connected peers</p>
            )}
          </div>
        </Card>
      )}

      {/* No status yet */}
      {!wgStatus && !systemInfo && (
        <Card>
          <div className="text-center py-8">
            <p className="text-apple-gray-400 mb-4">Click "WireGuard Status" or "System Info" to fetch node details via SSH</p>
          </div>
        </Card>
      )}

      {toast && (
        <Toast
          isOpen={!!toast}
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  )
}
