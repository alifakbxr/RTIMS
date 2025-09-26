export interface ReportStats {
  total_reports: number
  this_month: number
  data_points: number
}

export interface ReportType {
  id: string
  name: string
  description: string
  available: boolean
  icon: string
  color: string
}

export interface RecentReport {
  id: string
  name: string
  type: string
  format: string
  generated_at: string
  size: number
}

import { api } from './api'

export const reportsApi = {
  async getReportStats(): Promise<ReportStats> {
    try {
      const response = await api.get('/admin/reports/stats')
      return response.data
    } catch (error) {
      console.error('Failed to fetch report stats:', error)
      throw error
    }
  },

  async getReportTypes(): Promise<ReportType[]> {
    try {
      const response = await api.get('/admin/reports/types')
      return response.data
    } catch (error) {
      console.error('Failed to fetch report types:', error)
      throw error
    }
  },

  async getRecentReports(): Promise<RecentReport[]> {
    try {
      const response = await api.get('/admin/reports/recent')
      return response.data
    } catch (error) {
      console.error('Failed to fetch recent reports:', error)
      throw error
    }
  },

  async generateReport(reportType: string, format: 'json' | 'csv' | 'pdf'): Promise<Blob> {
    try {
      const response = await api.get(`/admin/reports/${reportType}`, {
        params: { format },
        responseType: 'blob'
      })
      return response.data
    } catch (error) {
      console.error('Failed to generate report:', error)
      throw error
    }
  },

  downloadReport(blob: Blob, filename: string): void {
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  }
}