import { useState, useEffect } from 'react'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Modal from '../components/ui/Modal'
import Table from '../components/ui/Table'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import Toast from '../components/ui/Toast'
import { networkService, CreateNetworkInput } from '../services/networks'
import type { Network } from '../types'

export default function Networks() {
  const [networks, setNetworks] = useState<Network[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Network | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)
  const [formData, setFormData] = useState<CreateNetworkInput>({
    name: '',
    description: '',
    cidr: '',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const fetchNetworks = async () => {
    try {
      const data = await networkService.list()
      setNetworks(data || [])
    } catch {
      // Handle error
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchNetworks()
  }, [])

  const handleCreate = async () => {
    setError('')
    setSaving(true)

    try {
      await networkService.create(formData)
      setShowModal(false)
      setFormData({ name: '', description: '', cidr: '' })
      fetchNetworks()
    } catch {
      setError('Failed to create network')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: number) => {
    setDeleting(true)
    try {
      await networkService.delete(id)
      setDeleteTarget(null)
      setToast({ message: 'Network deleted successfully', type: 'success' })
      fetchNetworks()
    } catch {
      setToast({ message: 'Failed to delete network', type: 'error' })
    } finally {
      setDeleting(false)
    }
  }

  const columns = [
    { key: 'name', header: 'Name' },
    { key: 'cidr', header: 'CIDR' },
    { key: 'description', header: 'Description' },
    {
      key: 'actions',
      header: 'Actions',
      render: (network: Network) => (
        <Button size="sm" variant="danger" onClick={() => setDeleteTarget(network)}>
          Delete
        </Button>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-apple-gray-500">Networks</h1>
          <p className="text-apple-gray-300 mt-1">Manage your virtual networks</p>
        </div>
        <Button onClick={() => setShowModal(true)}>Add Network</Button>
      </div>

      <Card>
        <Table
          columns={columns}
          data={networks}
          keyExtractor={(network) => network.id}
          loading={loading}
          emptyMessage="No networks configured yet"
        />
      </Card>

      <Modal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        title="Add Network"
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
            placeholder="Network name"
            required
          />
          <Input
            label="CIDR"
            value={formData.cidr}
            onChange={(e) => setFormData({ ...formData, cidr: e.target.value })}
            placeholder="10.0.0.0/24"
            required
          />
          <div>
            <label className="label">Description</label>
            <textarea
              className="input min-h-[80px]"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Optional description"
            />
          </div>
          {error && <div className="p-3 rounded-apple bg-red-50 text-apple-red text-sm">{error}</div>}
        </div>
      </Modal>

      <ConfirmDialog
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => deleteTarget && handleDelete(deleteTarget.id)}
        title="Delete Network"
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
