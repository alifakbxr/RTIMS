export interface Settings {
  lowStockThreshold: number
  notifications: {
    email: boolean
    lowStock: boolean
    systemAlerts: boolean
  }
  system: {
    maintenanceMode: boolean
    autoBackup: boolean
    backupFrequency: 'hourly' | 'daily' | 'weekly' | 'monthly'
  }
  appearance: {
    theme: 'light' | 'dark' | 'system'
    compactMode: boolean
  }
}

import { api } from './api'

export const settingsApi = {
  async getSettings(): Promise<Settings> {
    try {
      const response = await api.get('/admin/settings')
      return response.data
    } catch (error) {
      console.error('Failed to fetch settings:', error)
      throw error
    }
  },

  async updateSettings(settings: Settings): Promise<Settings> {
    try {
      const response = await api.put('/admin/settings', settings)
      return response.data
    } catch (error) {
      console.error('Failed to update settings:', error)
      throw error
    }
  },

  async getSystemStatus(): Promise<{
    database: 'healthy' | 'warning' | 'error'
    cache: 'healthy' | 'warning' | 'error'
    storage: 'healthy' | 'warning' | 'error'
    lastBackup: string | null
  }> {
    try {
      const response = await api.get('/admin/settings/status')
      return response.data
    } catch (error) {
      console.error('Failed to fetch system status:', error)
      throw error
    }
  },

  async triggerBackup(): Promise<{ success: boolean; message: string }> {
    try {
      const response = await api.post('/admin/settings/backup')
      return response.data
    } catch (error) {
      console.error('Failed to trigger backup:', error)
      throw error
    }
  }
}