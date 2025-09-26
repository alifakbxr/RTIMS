import { SystemStatus, StockAlert } from '@/types'
import { api } from './api'

interface Activity {
  id: string
  type: string
  description: string
  timestamp: string
  user?: string
}

export const dashboardApi = {
  // Get dashboard statistics
  async getStats(): Promise<SystemStatus> {
    const response = await api.get('/dashboard/stats')
    return response.data
  },

  // Get low stock alerts
  async getLowStockAlerts(): Promise<StockAlert[]> {
    const response = await api.get('/dashboard/alerts')
    return response.data
  },

  // Get recent activities
  async getRecentActivities(): Promise<Activity[]> {
    const response = await api.get('/dashboard/activities')
    return response.data
  }
}