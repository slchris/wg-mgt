import api from './api'
import type { ApiResponse, Node, Peer } from '../types'

export interface CreateNodeInput {
  name: string
  host: string
  ssh_port?: number
  ssh_user: string
  ssh_key: string
  wg_interface?: string
  wg_port?: number
  wg_address: string
  endpoint?: string
  network_id?: number | null
}

export interface UpdateNodeInput {
  name?: string
  host?: string
  ssh_port?: number
  ssh_user?: string
  ssh_key?: string
  wg_interface?: string
  wg_port?: number
  wg_address?: string
  endpoint?: string
  network_id?: number | null
}

export interface WireGuardPeerStatus {
  public_key: string
  endpoint: string
  allowed_ips: string[]
  latest_handshake: string
  transfer_rx: number
  transfer_tx: number
}

export interface WireGuardStatus {
  interface: string
  address: string
  public_key: string
  listen_port: number
  peers: WireGuardPeerStatus[]
  is_running: boolean
  total_rx: number
  total_tx: number
}

export interface SystemInfo {
  hostname: string
  kernel: string
  os: string
  wireguard_installed: string
  wireguard_version?: string
}

export const nodeService = {
  async list(): Promise<Node[]> {
    const response = await api.get<ApiResponse<Node[]>>('/nodes')
    return response.data.data || []
  },

  async get(id: number): Promise<Node> {
    const response = await api.get<ApiResponse<Node>>(`/nodes/${id}`)
    return response.data.data
  },

  async create(input: CreateNodeInput): Promise<Node> {
    const response = await api.post<ApiResponse<Node>>('/nodes', input)
    return response.data.data
  },

  async update(id: number, input: UpdateNodeInput): Promise<Node> {
    const response = await api.put<ApiResponse<Node>>(`/nodes/${id}`, input)
    return response.data.data
  },

  async delete(id: number): Promise<void> {
    await api.delete(`/nodes/${id}`)
  },

  async checkStatus(id: number): Promise<Node> {
    const response = await api.post<ApiResponse<Node>>(`/nodes/${id}/check`)
    return response.data.data
  },

  async getWireGuardStatus(id: number): Promise<WireGuardStatus> {
    const response = await api.get<ApiResponse<WireGuardStatus>>(`/nodes/${id}/wg-status`)
    return response.data.data
  },

  async getSystemInfo(id: number): Promise<SystemInfo> {
    const response = await api.get<ApiResponse<SystemInfo>>(`/nodes/${id}/system-info`)
    return response.data.data
  },

  async getPeers(nodeId: number): Promise<Peer[]> {
    const response = await api.get<ApiResponse<Peer[]>>(`/nodes/${nodeId}/peers`)
    return response.data.data || []
  },
}
