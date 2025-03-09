// Frontend model interfaces that match backend types

// UUID helper type
export type UUID = string

// User model
export interface User {
  id: UUID
  username: string
  email?: string
  role: 'admin' | 'operator' | 'viewer'
  created_at?: string
  updated_at?: string
}

// Fleet model
export interface Fleet {
  id: UUID
  name: string
  description?: string
  devices?: Device[]
  created_at?: string
  updated_at?: string
}

// Device model
export interface Device {
  id: UUID
  device_id: string
  name: string
  fleet_id?: UUID
  status: 'pending' | 'online' | 'offline' | 'updating' | 'error'
  last_seen?: string
  ip_address?: string
  os_version?: string
  hardware_info?: string
  ssh_port?: number
  ssh_public_key?: string
  subdomain?: string
  subdomain_enabled?: boolean
  created_at?: string
  updated_at?: string
}

// Software model
export interface Software {
  id: UUID
  name: string
  source: 'github' | 'manual'
  repo_url?: string
  current_version?: string
  versions?: string
  docker_compose_yaml?: string
  default_env_vars?: string
  created_at?: string
  updated_at?: string
}

// Deployment model
export interface Deployment {
  id: UUID
  software_id: UUID
  fleet_id?: UUID
  device_id?: UUID
  version: string
  pinned: boolean
  status: 'pending' | 'deployed' | 'failed'
  env_vars?: string
  created_at?: string
  updated_at?: string
}

// Auth request/response interfaces
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  user: User
  token: string
}

// API response wrappers
export interface ApiResponse<T> {
  data?: T
  error?: string
}
