import api from './api'
import type { ApiResponse, LoginResponse, SetupCheckResponse, User } from '../types'

export const authService = {
  async checkSetup(): Promise<SetupCheckResponse> {
    const response = await api.get<ApiResponse<SetupCheckResponse>>('/setup/check')
    return response.data.data
  },

  async setup(username: string, password: string): Promise<User> {
    const response = await api.post<ApiResponse<User>>('/setup', {
      username,
      password,
    })
    return response.data.data
  },

  async login(username: string, password: string): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/auth/login', {
      username,
      password,
    })
    return response.data.data
  },

  async me(): Promise<User> {
    const response = await api.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },

  async changePassword(oldPassword: string, newPassword: string): Promise<void> {
    await api.post('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
