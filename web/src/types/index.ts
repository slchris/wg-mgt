export interface Node {
  id: number
  name: string
  host: string
  ssh_port: number
  ssh_user: string
  wg_interface: string
  wg_port: number
  wg_address: string
  public_key: string
  endpoint: string
  status: 'online' | 'offline' | 'unknown'
  last_seen: string | null
  network_id: number | null
  network?: Network
  created_at: string
  updated_at: string
  peers?: Peer[]
}

export interface Peer {
  id: number
  name: string
  public_key: string
  address: string
  allowed_ips: string
  dns: string
  node_id: number
  enabled: boolean
  expires_at: string | null
  created_at: string
  updated_at: string
  node?: Node
}

export interface Network {
  id: number
  name: string
  cidr: string
  gateway: string
  description: string
  created_at: string
  updated_at: string
}

export interface User {
  id: number
  username: string
  role: string
  created_at: string
  updated_at: string
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
}

export interface LoginResponse {
  token: string
}

export interface SetupCheckResponse {
  needs_setup: boolean
}
