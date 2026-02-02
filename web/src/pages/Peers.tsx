import { useState, useEffect, useCallback } from 'react'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Modal from '../components/ui/Modal'
import Table from '../components/ui/Table'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import Toast from '../components/ui/Toast'
import { peerService, CreatePeerInput } from '../services/peers'
import { nodeService } from '../services/nodes'
import type { Peer, Node } from '../types'
import { useAuthStore } from '../stores/auth'

export default function Peers() {
  const [peers, setPeers] = useState<Peer[]>([])
  const [nodes, setNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [showQRModal, setShowQRModal] = useState(false)
  const [selectedPeer, setSelectedPeer] = useState<Peer | null>(null)
  const [qrCodeUrl, setQrCodeUrl] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Peer | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)
  const [formData, setFormData] = useState<CreatePeerInput>({
    name: '',
    node_id: 0,
    address: '',
    allowed_ips: '0.0.0.0/0',
    dns: '1.1.1.1',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const { token } = useAuthStore()

  const fetchData = useCallback(async () => {
    try {
      const [peersData, nodesData] = await Promise.all([peerService.list(), nodeService.list()])
      setPeers(peersData || [])
      setNodes(nodesData || [])
      if (nodesData && nodesData.length > 0) {
        setFormData((prev) => prev.node_id === 0 ? { ...prev, node_id: nodesData[0].id } : prev)
      }
    } catch {
      // Handle error
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleCreate = async () => {
    setError('')
    setSaving(true)

    try {
      await peerService.create(formData)
      setShowModal(false)
      setFormData({
        name: '',
        node_id: nodes.length > 0 ? nodes[0].id : 0,
        address: '',
        allowed_ips: '0.0.0.0/0',
        dns: '1.1.1.1',
      })
      fetchData()
    } catch {
      setError('Failed to create peer')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: number) => {
    setDeleting(true)
    try {
      await peerService.delete(id)
      setDeleteTarget(null)
      setToast({ message: 'Peer deleted successfully', type: 'success' })
      fetchData()
    } catch {
      setToast({ message: 'Failed to delete peer', type: 'error' })
    } finally {
      setDeleting(false)
    }
  }

  const handleToggleEnabled = async (peer: Peer) => {
    try {
      await peerService.update(peer.id, { enabled: !peer.enabled })
      fetchData()
    } catch {
      // Handle error
    }
  }

  const handleDownloadConfig = (peer: Peer) => {
    const url = `/api/v1/peers/${peer.id}/config`
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `${peer.name}.conf`)
    // Add authorization header via fetch
    fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((res) => res.blob())
      .then((blob) => {
        const url = window.URL.createObjectURL(blob)
        link.href = url
        document.body.appendChild(link)
        link.click()
        link.remove()
        window.URL.revokeObjectURL(url)
      })
  }

  const handleShowQR = async (peer: Peer) => {
    setSelectedPeer(peer)
    setQrCodeUrl(null)
    setShowQRModal(true)
    
    try {
      const res = await fetch(`/api/v1/peers/${peer.id}/qrcode`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.ok) {
        const blob = await res.blob()
        setQrCodeUrl(URL.createObjectURL(blob))
      }
    } catch {
      // Handle error silently
    }
  }

  const columns = [
    { key: 'name', header: 'Name' },
    {
      key: 'node',
      header: 'Node',
      render: (peer: Peer) => peer.node?.name || '-',
    },
    { key: 'address', header: 'Address' },
    {
      key: 'enabled',
      header: 'Status',
      render: (peer: Peer) => (
        <button
          onClick={() => handleToggleEnabled(peer)}
          className={`px-2 py-1 text-xs font-medium rounded-full ${
            peer.enabled ? 'bg-green-100 text-apple-green' : 'bg-gray-100 text-apple-gray-400'
          }`}
        >
          {peer.enabled ? 'Enabled' : 'Disabled'}
        </button>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (peer: Peer) => (
        <div className="flex items-center space-x-2">
          <Button size="sm" variant="ghost" onClick={() => handleShowQR(peer)}>
            QR
          </Button>
          <Button size="sm" variant="ghost" onClick={() => handleDownloadConfig(peer)}>
            Download
          </Button>
          <Button size="sm" variant="danger" onClick={() => setDeleteTarget(peer)}>
            Delete
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-apple-gray-500">Peers</h1>
          <p className="text-apple-gray-300 mt-1">Manage your WireGuard client peers</p>
        </div>
        <Button onClick={async () => {
          setShowModal(true)
          // Auto-fetch next available IP for the first node
          if (nodes.length > 0) {
            const firstNodeId = nodes[0].id
            try {
              const nextIP = await peerService.getNextIP(firstNodeId)
              setFormData((prev) => ({ ...prev, node_id: firstNodeId, address: nextIP }))
            } catch {
              // Ignore error
            }
          }
        }} disabled={nodes.length === 0}>
          Add Peer
        </Button>
      </div>

      {nodes.length === 0 && (
        <div className="p-4 rounded-apple bg-yellow-50 text-yellow-800 text-sm">
          Please add a node first before creating peers.
        </div>
      )}

      <Card>
        <Table columns={columns} data={peers} keyExtractor={(peer) => peer.id} loading={loading} emptyMessage="No peers configured yet" />
      </Card>

      <Modal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        title="Add Peer"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} loading={saving}>
              Create
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Peer name"
            required
          />
          <div>
            <label className="label">Node</label>
            <select
              className="input"
              value={formData.node_id}
              onChange={async (e) => {
                const nodeId = parseInt(e.target.value)
                setFormData({ ...formData, node_id: nodeId })
                // Auto-fetch next available IP when node changes
                try {
                  const nextIP = await peerService.getNextIP(nodeId)
                  setFormData((prev) => ({ ...prev, node_id: nodeId, address: nextIP }))
                } catch {
                  // Ignore error, user can still input manually
                }
              }}
            >
              {nodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.name}
                </option>
              ))}
            </select>
          </div>
          <Input
            label="Address"
            value={formData.address}
            onChange={(e) => setFormData({ ...formData, address: e.target.value })}
            placeholder="Auto-generated from node, or enter manually"
            required
          />
          <Input
            label="Allowed IPs"
            value={formData.allowed_ips}
            onChange={(e) => setFormData({ ...formData, allowed_ips: e.target.value })}
            placeholder="0.0.0.0/0"
          />
          <Input
            label="DNS"
            value={formData.dns}
            onChange={(e) => setFormData({ ...formData, dns: e.target.value })}
            placeholder="1.1.1.1"
          />
          {error && <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>}
        </div>
      </Modal>

      <Modal
        isOpen={showQRModal}
        onClose={() => {
          setShowQRModal(false)
          setQrCodeUrl(null)
        }}
        title={`QR Code - ${selectedPeer?.name}`}
      >
        {selectedPeer && (
          <div className="flex flex-col items-center">
            {qrCodeUrl ? (
              <img src={qrCodeUrl} alt="QR Code" className="w-64 h-64" />
            ) : (
              <div className="w-64 h-64 flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-apple-blue"></div>
              </div>
            )}
            <p className="mt-4 text-sm text-apple-gray-300">
              Scan this QR code with your WireGuard app
            </p>
          </div>
        )}
      </Modal>

      <ConfirmDialog
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && handleDelete(deleteTarget.id)}
        title="Delete Peer"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        loading={deleting}
      />

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
