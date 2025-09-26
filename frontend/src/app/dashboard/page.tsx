'use client'

import { useEffect, useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Navigation } from '@/components/ui/navigation'
import { StockAlert, SystemStatus } from '@/types'
import { Package, AlertTriangle, Activity, Users, Clock } from 'lucide-react'
import { dashboardApi } from '@/lib/dashboard'

export default function DashboardPage() {
  const { user, logout } = useAuth()
  const [stats, setStats] = useState<SystemStatus | null>(null)
  const [alerts, setAlerts] = useState<StockAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setLoading(true)
        setError(null)

        // Fetch dashboard statistics
        const dashboardData = await dashboardApi.getStats()
        setStats(dashboardData)

        // Fetch low stock alerts
        const alertsData = await dashboardApi.getLowStockAlerts()
        setAlerts(alertsData)
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err)
        setError('Failed to load dashboard data. Please try again.')
      } finally {
        setLoading(false)
      }
    }

    fetchDashboardData()
  }, [])

  if (!user) {
    return <div>Please log in to access the dashboard.</div>
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <Navigation />
        <div className="lg:pl-64">
          <div className="flex items-center justify-center min-h-[60vh]">
            <div className="text-center">
              <div className="w-12 h-12 mx-auto mb-4 border-b-2 border-blue-600 rounded-full animate-spin"></div>
              <p className="text-lg text-gray-600">Loading dashboard...</p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background">
        <Navigation />
        <div className="lg:pl-64">
          <div className="flex items-center justify-center min-h-[60vh]">
            <div className="text-center">
              <div className="p-4 mb-4 border border-red-200 rounded-lg bg-red-50">
                <AlertTriangle className="w-8 h-8 mx-auto mb-2 text-red-600" />
                <p className="font-medium text-red-800">Error</p>
                <p className="text-sm text-red-600">{error}</p>
              </div>
              <Button
                onClick={() => window.location.reload()}
                className="bg-blue-600 hover:bg-blue-700"
              >
                Retry
              </Button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Navigation */}
      <Navigation />

      {/* Main Content */}
      <div className="lg:pl-64">
        {/* Top Header */}
        <header className="sticky top-0 z-10 border-b border-gray-100 shadow-sm bg-white/95 backdrop-blur-sm">
          <div className="px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between py-6">
              <div className="space-y-1">
                <h1 className="text-3xl font-bold tracking-tight text-gray-900">Dashboard</h1>
                <p className="text-lg text-gray-600">Welcome back, <span className="font-medium text-gray-900">{user.name}</span></p>
              </div>
              <div className="flex items-center space-x-3">
                <Badge
                  variant="secondary"
                  className="px-4 py-2 text-sm font-medium text-blue-700 border-blue-200 rounded-full bg-blue-50"
                >
                  {user.role}
                </Badge>
              </div>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="px-4 py-8 sm:px-6 lg:px-8">
          {/* Stats Cards */}
          <div className="grid grid-cols-1 gap-6 mb-8 sm:grid-cols-2 lg:grid-cols-4">
            <Card className="animate-fade-in">
              <CardHeader className="flex flex-row items-center justify-between pb-3 space-y-0">
                <CardTitle className="text-sm font-medium text-gray-600">Total Products</CardTitle>
                <div className="p-2 rounded-lg bg-blue-50">
                  <Package className="w-4 h-4 text-blue-600" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-gray-900">{stats?.total_products || 0}</div>
                <p className="mt-2 text-sm text-gray-500">Active inventory items</p>
              </CardContent>
            </Card>

            <Card className="animate-fade-in">
              <CardHeader className="flex flex-row items-center justify-between pb-3 space-y-0">
                <CardTitle className="text-sm font-medium text-gray-600">Low Stock Items</CardTitle>
                <div className="p-2 rounded-lg bg-red-50">
                  <AlertTriangle className="w-4 h-4 text-red-600" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-red-600">{stats?.low_stock_count || 0}</div>
                <p className="mt-2 text-sm text-gray-500">Require attention</p>
              </CardContent>
            </Card>

            <Card className="animate-fade-in">
              <CardHeader className="flex flex-row items-center justify-between pb-3 space-y-0">
                <CardTitle className="text-sm font-medium text-gray-600">Total Users</CardTitle>
                <div className="p-2 rounded-lg bg-purple-50">
                  <Users className="w-4 h-4 text-purple-600" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-gray-900">{stats?.total_users || 0}</div>
                <p className="mt-2 text-sm text-gray-500">System users</p>
              </CardContent>
            </Card>

            <Card className="animate-fade-in">
              <CardHeader className="flex flex-row items-center justify-between pb-3 space-y-0">
                <CardTitle className="text-sm font-medium text-gray-600">Server Time</CardTitle>
                <div className="p-2 rounded-lg bg-green-50">
                  <Clock className="w-4 h-4 text-green-600" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-xl font-bold text-gray-900">
                  {stats?.server_time ? new Date(stats.server_time).toLocaleTimeString() : 'Loading...'}
                </div>
                <p className="mt-2 text-sm text-gray-500">Current system time</p>
              </CardContent>
            </Card>
          </div>

          {/* Alerts Section */}
          <div className="grid grid-cols-1 gap-4 xl:grid-cols-2 sm:gap-6">
            <Card className="transition-all duration-300 bg-white border border-gray-200 shadow-sm hover:shadow-md animate-fade-in">
              <CardHeader>
                <CardTitle className="text-lg font-semibold text-gray-900 sm:text-xl">Low Stock Alerts</CardTitle>
                <CardDescription className="text-gray-600">Products that need restocking</CardDescription>
              </CardHeader>
              <CardContent>
                {alerts.length === 0 ? (
                  <div className="py-6 text-center sm:py-8">
                    <p className="text-base text-gray-600 sm:text-lg">No low stock alerts</p>
                    <p className="mt-1 text-sm text-gray-500">All products are well-stocked</p>
                  </div>
                ) : (
                  <div className="space-y-3 sm:space-y-4">
                    {alerts.map((alert) => (
                      <div key={alert.id} className="flex flex-col p-3 transition-colors border border-red-200 rounded-lg sm:flex-row sm:items-center sm:justify-between sm:p-4 bg-red-50 hover:bg-red-100/50">
                        <div className="mb-2 space-y-1 sm:mb-0">
                          <p className="font-semibold text-gray-900">{alert.name}</p>
                          <p className="text-sm text-gray-600">SKU: {alert.sku}</p>
                        </div>
                        <div className="space-y-1 text-left sm:text-right">
                          <p className="text-base font-bold text-red-600 sm:text-lg">
                            Stock: {alert.stock}
                          </p>
                          <p className="text-sm text-gray-600">
                            Min: {alert.minimum_threshold}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="transition-all duration-300 bg-white border border-gray-200 shadow-sm hover:shadow-md animate-fade-in">
              <CardHeader>
                <CardTitle className="text-lg font-semibold text-gray-900 sm:text-xl">Quick Actions</CardTitle>
                <CardDescription className="text-gray-600">Common tasks and shortcuts</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-1 sm:gap-3">
                  <Button
                    className="justify-start w-full h-10 font-medium text-left text-blue-700 transition-all duration-300 border-blue-200 sm:h-12 hover:shadow-md bg-blue-50 hover:bg-blue-100"
                    variant="outline"
                  >
                    <span className="text-sm sm:text-base">Add New Product</span>
                  </Button>
                  <Button
                    className="justify-start w-full h-10 font-medium text-left text-blue-700 transition-all duration-300 border-blue-200 sm:h-12 hover:shadow-md bg-blue-50 hover:bg-blue-100"
                    variant="outline"
                  >
                    <span className="text-sm sm:text-base">Update Stock Levels</span>
                  </Button>
                  <Button
                    className="justify-start w-full h-10 font-medium text-left text-blue-700 transition-all duration-300 border-blue-200 sm:h-12 hover:shadow-md bg-blue-50 hover:bg-blue-100"
                    variant="outline"
                  >
                    <span className="text-sm sm:text-base">Generate Report</span>
                  </Button>
                  <Button
                    className="justify-start w-full h-10 font-medium text-left text-blue-700 transition-all duration-300 border-blue-200 sm:h-12 hover:shadow-md bg-blue-50 hover:bg-blue-100"
                    variant="outline"
                  >
                    <span className="text-sm sm:text-base">Manage Users</span>
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </main>
      </div>
    </div>
  )
}