import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Modal from '../components/ui/Modal'
import Table from '../components/ui/Table'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import Toast from '../components/ui/Toast'
import { nodeService, CreateNodeInput } from '../services/nodes'
import type { Node } from '../types'

export default function Nodes() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Node | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)
  const [formData, setFormData] = useState<CreateNodeInput>({
    name: '',
    host: '',
    ssh_port: 22,
    ssh_user: 'root',
    ssh_key: '',
    wg_interface: 'wg0',
    wg_port: 51820,
    wg_address: '',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)

  const fetchNodes = async () => {
    try {
      const data = await nodeService.list()
      setNodes(data || [])
    } catch {
      // Handle error
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchNodes()
  }, [])

  const handleCreate = async () => {
    setError('')
    setSaving(true)

    try {
      await nodeService.create(formData)
      setShowModal(false)
      setFormData({
        name: '',
        host: '',
        ssh_port: 22,
        ssh_user: 'root',
        ssh_key: '',
        wg_interface: 'wg0',
        wg_port: 51820,
        wg_address: '',
      })
      setShowAdvanced(false)
      fetchNodes()
    } catch {
      setError('Failed to create node')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: number) => {
    setDeleting(true)
    try {
      await nodeService.delete(id)
      setDeleteTarget(null)
      setToast({ message: 'Node deleted successfully', type: 'success' })
      fetchNodes()
    } catch {
      setToast({ message: 'Failed to delete node', type: 'error' })
    } finally {
      setDeleting(false)
    }
  }

  const [checkingStatus, setCheckingStatus] = useState<number | null>(null)

  const handleCheckStatus = async (id: number) => {
    setCheckingStatus(id)
    try {
      await nodeService.checkStatus(id)
      setToast({ message: 'Status checked successfully', type: 'success' })
      fetchNodes()
    } catch {
      setToast({ message: 'Failed to check status', type: 'error' })
    } finally {
      setCheckingStatus(null)
    }
  }

  const columns = [
    { 
      key: 'name', 
      header: 'Name',
      render: (node: Node) => (
        <Link to={`/nodes/${node.id}`} className="text-apple-blue hover:underline font-medium">
          {node.name}
        </Link>
      ),
    },
    { key: 'host', header: 'Host' },
    { key: 'wg_address', header: 'WG Address' },
    {
      key: 'status',
      header: 'Status',
      render: (node: Node) => (
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
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (node: Node) => (
        <div className="flex items-center space-x-2">
          <Button 
            size="sm" 
            variant="ghost" 
            onClick={() => handleCheckStatus(node.id)}
            loading={checkingStatus === node.id}
          >
            Check
          </Button>
          <Button size="sm" variant="danger" onClick={() => setDeleteTarget(node)}>
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
          <h1 className="text-2xl font-semibold text-apple-gray-500">Nodes</h1>
          <p className="text-apple-gray-300 mt-1">Manage your WireGuard server nodes</p>
        </div>
        <Button onClick={() => setShowModal(true)}>Add Node</Button>
      </div>

      <Card>
        <Table columns={columns} data={nodes} keyExtractor={(node) => node.id} loading={loading} emptyMessage="No nodes configured yet" />
      </Card>

      <Modal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        title="Add Node"
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
            placeholder="Node name"
            required
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Host"
              value={formData.host}
              onChange={(e) => setFormData({ ...formData, host: e.target.value })}
              placeholder="IP or hostname"
              required
            />
            <Input
              label="SSH Port"
              type="number"
              value={formData.ssh_port}
              onChange={(e) => setFormData({ ...formData, ssh_port: parseInt(e.target.value) })}
            />
          </div>
          <Input
            label="SSH User"
            value={formData.ssh_user}
            onChange={(e) => setFormData({ ...formData, ssh_user: e.target.value })}
            required
          />
          <div>
            <label className="label">SSH Private Key</label>
            <textarea
              className="input min-h-[120px] font-mono text-sm"
              value={formData.ssh_key}
              onChange={(e) => setFormData({ ...formData, ssh_key: e.target.value })}
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
              required
            />
          </div>

          <div className="p-3 rounded-apple bg-apple-gray-50 text-apple-gray-400 text-sm">
            WireGuard 配置（网卡、端口、地址）将在添加节点后自动通过 SSH 获取。如需手动指定，请展开高级选项。
          </div>

          <button
            type="button"
            className="text-apple-blue text-sm hover:underline"
            onClick={() => setShowAdvanced(!showAdvanced)}
          >
            {showAdvanced ? '收起高级选项' : '展开高级选项'}
          </button>

          {showAdvanced && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="WG Interface"
                  value={formData.wg_interface}
                  onChange={(e) => setFormData({ ...formData, wg_interface: e.target.value })}
                  placeholder="wg0"
                  helperText="默认: wg0"
                />
                <Input
                  label="WG Port"
                  type="number"
                  value={formData.wg_port}
                  onChange={(e) => setFormData({ ...formData, wg_port: parseInt(e.target.value) || 51820 })}
                  helperText="默认: 51820"
                />
              </div>
              <Input
                label="WG Address"
                value={formData.wg_address}
                onChange={(e) => setFormData({ ...formData, wg_address: e.target.value })}
                placeholder="10.0.0.1/24"
                helperText="留空则自动获取"
              />
            </>
          )}

          {error && <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>}
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && handleDelete(deleteTarget.id)}
        title="Delete Node"
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
