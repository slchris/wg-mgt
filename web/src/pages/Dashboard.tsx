import { useState, useEffect } from 'react'
import Card from '../components/ui/Card'
import { nodeService } from '../services/nodes'
import { peerService } from '../services/peers'
import { networkService } from '../services/networks'
import type { Node, Peer, Network } from '../types'

export default function Dashboard() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [peers, setPeers] = useState<Peer[]>([])
  const [networks, setNetworks] = useState<Network[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([nodeService.list(), peerService.list(), networkService.list()])
      .then(([nodesData, peersData, networksData]) => {
        setNodes(nodesData || [])
        setPeers(peersData || [])
        setNetworks(networksData || [])
      })
      .catch(() => {
        // Handle error silently
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  const onlineNodes = nodes.filter((n) => n.status === 'online').length
  const enabledPeers = peers.filter((p) => p.enabled).length

  const stats = [
    { label: 'Total Nodes', value: nodes.length, color: 'bg-apple-blue' },
    { label: 'Online Nodes', value: onlineNodes, color: 'bg-apple-green' },
    { label: 'Total Peers', value: peers.length, color: 'bg-apple-orange' },
    { label: 'Active Peers', value: enabledPeers, color: 'bg-purple-500' },
    { label: 'Networks', value: networks.length, color: 'bg-apple-teal' },
  ]

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-apple-blue"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-apple-gray-500">Dashboard</h1>
        <p className="text-apple-gray-300 mt-1">Overview of your WireGuard network</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <Card key={stat.label} className="!p-4">
            <div className="flex items-center">
              <div className={`w-12 h-12 ${stat.color} rounded-apple flex items-center justify-center`}>
                <span className="text-white text-lg font-semibold">{stat.value}</span>
              </div>
              <div className="ml-4">
                <p className="text-sm text-apple-gray-300">{stat.label}</p>
                <p className="text-xl font-semibold text-apple-gray-500">{stat.value}</p>
              </div>
            </div>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card title="Recent Nodes">
          {nodes.length === 0 ? (
            <p className="text-apple-gray-300 text-center py-4">No nodes configured yet</p>
          ) : (
            <div className="space-y-3">
              {nodes.slice(0, 5).map((node) => (
                <div
                  key={node.id}
                  className="flex items-center justify-between p-3 bg-apple-gray-50 rounded-apple"
                >
                  <div>
                    <p className="font-medium text-apple-gray-500">{node.name}</p>
                    <p className="text-sm text-apple-gray-300">{node.host}</p>
                  </div>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded-full ${
                      node.status === 'online'
                        ? 'bg-green-100 text-apple-green'
                        : node.status === 'offline'
                        ? 'bg-red-100 text-apple-red'
                        : 'bg-gray-100 text-apple-gray-400'
                    }`}
                  >
                    {node.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card title="Recent Peers">
          {peers.length === 0 ? (
            <p className="text-apple-gray-300 text-center py-4">No peers configured yet</p>
          ) : (
            <div className="space-y-3">
              {peers.slice(0, 5).map((peer) => (
                <div
                  key={peer.id}
                  className="flex items-center justify-between p-3 bg-apple-gray-50 rounded-apple"
                >
                  <div>
                    <p className="font-medium text-apple-gray-500">{peer.name}</p>
                    <p className="text-sm text-apple-gray-300">{peer.address}</p>
                  </div>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded-full ${
                      peer.enabled
                        ? 'bg-green-100 text-apple-green'
                        : 'bg-gray-100 text-apple-gray-400'
                    }`}
                  >
                    {peer.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  )
}
