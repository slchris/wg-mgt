import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Toast from '../components/ui/Toast'
import { nodeService, WireGuardStatus, SystemInfo, InitializeWireGuardResult } from '../services/nodes'
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
  const [loadingInit, setLoadingInit] = useState(false)
  const [loadingSave, setLoadingSave] = useState(false)
  const [loadingRestart, setLoadingRestart] = useState(false)
  const [showInitModal, setShowInitModal] = useState(false)
  const [initAddress, setInitAddress] = useState('10.0.0.1/24')
  const [initPort, setInitPort] = useState(51820)
  const [initResult, setInitResult] = useState<InitializeWireGuardResult | null>(null)
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

  const handleInitialize = async () => {
    if (!id) return
    setLoadingInit(true)
    try {
      const result = await nodeService.initializeWireGuard(parseInt(id), {
        address: initAddress,
        port: initPort,
      })
      setInitResult(result)
      setShowInitModal(false)
      setToast({ message: result.message, type: 'success' })
      // Refresh node data
      fetchNode()
    } catch {
      setToast({ message: 'Failed to initialize WireGuard', type: 'error' })
    } finally {
      setLoadingInit(false)
    }
  }

  const handleSaveConfig = async () => {
    if (!id) return
    setLoadingSave(true)
    try {
      await nodeService.saveWireGuardConfig(parseInt(id))
      setToast({ message: 'Config saved to /etc/wireguard/', type: 'success' })
    } catch {
      setToast({ message: 'Failed to save config', type: 'error' })
    } finally {
      setLoadingSave(false)
    }
  }

  const handleRestart = async () => {
    if (!id) return
    setLoadingRestart(true)
    try {
      await nodeService.restartWireGuard(parseInt(id))
      setToast({ message: 'WireGuard restarted', type: 'success' })
      // Refresh status
      setTimeout(fetchWgStatus, 1000)
    } catch {
      setToast({ message: 'Failed to restart WireGuard', type: 'error' })
    } finally {
      setLoadingRestart(false)
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
          {(!node.wg_address || !node.public_key) && (
            <Button variant="primary" onClick={() => setShowInitModal(true)}>
              Initialize WG
            </Button>
          )}
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
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-medium text-apple-gray-500">WireGuard Status</h2>
            <div className="flex space-x-2">
              <Button size="sm" variant="secondary" onClick={handleSaveConfig} loading={loadingSave}>
                Save Config
              </Button>
              <Button size="sm" variant="secondary" onClick={handleRestart} loading={loadingRestart}>
                Restart
              </Button>
            </div>
          </div>
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

      {/* Initialize Result */}
      {initResult && (
        <Card>
          <h2 className="text-lg font-medium text-apple-gray-500 mb-4">Initialization Result</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-sm text-apple-gray-300">Installed</p>
              <p className="font-medium text-apple-gray-500">
                {initResult.was_installed ? (
                  <span className="text-apple-green">Just Installed</span>
                ) : initResult.installed ? (
                  <span className="text-apple-green">Yes</span>
                ) : (
                  <span className="text-apple-red">No</span>
                )}
              </p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">Configured</p>
              <p className="font-medium text-apple-gray-500">
                {initResult.was_configured ? (
                  <span className="text-apple-green">Just Configured</span>
                ) : initResult.configured ? (
                  <span className="text-apple-green">Yes</span>
                ) : (
                  <span className="text-apple-red">No</span>
                )}
              </p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">Interface</p>
              <p className="font-medium text-apple-gray-500">{initResult.interface}</p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">Address</p>
              <p className="font-medium text-apple-gray-500">{initResult.address}</p>
            </div>
            <div>
              <p className="text-sm text-apple-gray-300">Port</p>
              <p className="font-medium text-apple-gray-500">{initResult.port}</p>
            </div>
            <div className="col-span-2 md:col-span-3">
              <p className="text-sm text-apple-gray-300">Public Key</p>
              <p className="font-mono text-xs text-apple-gray-500 break-all">{initResult.public_key}</p>
            </div>
          </div>
          <p className="mt-4 text-sm text-apple-gray-400">{initResult.message}</p>
        </Card>
      )}

      {/* Initialize Modal */}
      {showInitModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-apple-lg p-6 w-full max-w-md shadow-apple">
            <h2 className="text-lg font-semibold text-apple-gray-500 mb-4">Initialize WireGuard</h2>
            <p className="text-sm text-apple-gray-400 mb-4">
              This will install WireGuard (if needed), generate keys, and create the config file.
            </p>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-apple-gray-500 mb-1">
                  WireGuard Address (CIDR)
                </label>
                <input
                  type="text"
                  value={initAddress}
                  onChange={(e) => setInitAddress(e.target.value)}
                  className="w-full px-3 py-2 border border-apple-gray-200 rounded-apple focus:outline-none focus:ring-2 focus:ring-apple-blue"
                  placeholder="10.0.0.1/24"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-apple-gray-500 mb-1">
                  Listen Port
                </label>
                <input
                  type="number"
                  value={initPort}
                  onChange={(e) => setInitPort(parseInt(e.target.value))}
                  className="w-full px-3 py-2 border border-apple-gray-200 rounded-apple focus:outline-none focus:ring-2 focus:ring-apple-blue"
                  placeholder="51820"
                />
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <Button variant="secondary" onClick={() => setShowInitModal(false)}>
                Cancel
              </Button>
              <Button onClick={handleInitialize} loading={loadingInit}>
                Initialize
              </Button>
            </div>
          </div>
        </div>
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
