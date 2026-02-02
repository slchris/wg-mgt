import api from './api'
import type { ApiResponse, Peer } from '../types'

export interface CreatePeerInput {
  name: string
  node_id: number
  address: string
  allowed_ips: string
  dns?: string
  enabled?: boolean
}

export interface UpdatePeerInput {
  name?: string
  address?: string
  allowed_ips?: string
  dns?: string
  enabled?: boolean
}

export const peerService = {
  async list(): Promise<Peer[]> {
    const response = await api.get<ApiResponse<Peer[]>>('/peers')
    return response.data.data || []
  },

  async listByNode(nodeId: number): Promise<Peer[]> {
    const response = await api.get<ApiResponse<Peer[]>>(`/nodes/${nodeId}/peers`)
    return response.data.data || []
  },

  async get(id: number): Promise<Peer> {
    const response = await api.get<ApiResponse<Peer>>(`/peers/${id}`)
    return response.data.data
  },

  async create(input: CreatePeerInput): Promise<Peer> {
    const response = await api.post<ApiResponse<Peer>>('/peers', input)
    return response.data.data
  },

  async update(id: number, input: UpdatePeerInput): Promise<Peer> {
    const response = await api.put<ApiResponse<Peer>>(`/peers/${id}`, input)
    return response.data.data
  },

  async delete(id: number): Promise<void> {
    await api.delete(`/peers/${id}`)
  },

  async getConfig(id: number): Promise<string> {
    const response = await api.get<ApiResponse<{ config: string }>>(`/peers/${id}/config`)
    return response.data.data.config
  },

  async getNextIP(nodeId: number): Promise<string> {
    const response = await api.get<ApiResponse<{ next_ip: string }>>(`/peers/next-ip/${nodeId}`)
    return response.data.data.next_ip
  },
}
