import { User, UserRole } from '@/types'

export interface UsersFilter {
  search?: string
  role?: UserRole | ''
}

export interface CreateUserRequest {
  name: string
  email: string
  password: string
  role: UserRole
}

export interface UpdateUserRequest {
  name?: string
  email?: string
  role?: UserRole
  is_active?: boolean
}

import { api } from './api'

export const usersApi = {
  async getUsers(filters?: UsersFilter): Promise<User[]> {
    try {
      const params = new URLSearchParams()
      if (filters?.search) params.append('search', filters.search)
      if (filters?.role) params.append('role', filters.role)

      const response = await api.get(`/admin/users?${params.toString()}`)
      return response.data.users || response.data || []
    } catch (error) {
      console.error('Failed to fetch users:', error)
      throw error
    }
  },

  async createUser(userData: CreateUserRequest): Promise<User> {
    try {
      const response = await api.post('/admin/users', userData)
      return response.data
    } catch (error) {
      console.error('Failed to create user:', error)
      throw error
    }
  },

  async updateUser(userId: string, userData: UpdateUserRequest): Promise<User> {
    try {
      const response = await api.put(`/admin/users/${userId}`, userData)
      return response.data
    } catch (error) {
      console.error('Failed to update user:', error)
      throw error
    }
  },

  async deleteUser(userId: string): Promise<void> {
    try {
      await api.delete(`/admin/users/${userId}`)
    } catch (error) {
      console.error('Failed to delete user:', error)
      throw error
    }
  },

  async toggleUserStatus(userId: string, isActive: boolean): Promise<User> {
    return this.updateUser(userId, { is_active: isActive })
  }
}