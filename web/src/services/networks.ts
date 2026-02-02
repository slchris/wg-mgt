import api from './api'
import type { ApiResponse, Network } from '../types'

export interface CreateNetworkInput {
  name: string
  cidr: string
  gateway?: string
  description?: string
}

export interface UpdateNetworkInput {
  name?: string
  cidr?: string
  gateway?: string
  description?: string
}

export const networkService = {
  async list(): Promise<Network[]> {
    const response = await api.get<ApiResponse<Network[]>>('/networks')
    return response.data.data || []
  },

  async get(id: number): Promise<Network> {
    const response = await api.get<ApiResponse<Network>>(`/networks/${id}`)
    return response.data.data
  },

  async create(input: CreateNetworkInput): Promise<Network> {
    const response = await api.post<ApiResponse<Network>>('/networks', input)
    return response.data.data
  },

  async update(id: number, input: UpdateNetworkInput): Promise<Network> {
    const response = await api.put<ApiResponse<Network>>(`/networks/${id}`, input)
    return response.data.data
  },

  async delete(id: number): Promise<void> {
    await api.delete(`/networks/${id}`)
  },
}
