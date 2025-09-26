'use client'

import { useEffect, useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Navigation } from '@/components/ui/navigation'
import { BarChart3, Download, FileText, TrendingUp, Package, Users, Activity, AlertTriangle } from 'lucide-react'
import { reportsApi, ReportStats, ReportType, RecentReport } from '@/lib/reports'

export default function ReportsPage() {
  const { user } = useAuth()
  const [loading, setLoading] = useState(false)
  const [stats, setStats] = useState<ReportStats | null>(null)
  const [reportTypes, setReportTypes] = useState<ReportType[]>([])
  const [recentReports, setRecentReports] = useState<RecentReport[]>([])
  const [error, setError] = useState<string | null>(null)

  const [defaultReportTypes, setDefaultReportTypes] = useState<ReportType[]>([])

  useEffect(() => {
    const fetchReportsData = async () => {
      try {
        setError(null)

        // Fetch report statistics
        const statsData = await reportsApi.getReportStats()
        setStats(statsData)

        // Fetch report types
        const typesData = await reportsApi.getReportTypes()
        setReportTypes(typesData)

        // Fetch recent reports
        const recentData = await reportsApi.getRecentReports()
        setRecentReports(recentData)
      } catch (error) {
        console.error('Failed to fetch reports data:', error)
        setError('Failed to load reports data. Please try again.')
      }
    }

    fetchReportsData()
  }, [])

  const iconMap: Record<string, typeof Package> = {
    Package,
    Activity,
    Users,
    TrendingUp,
    BarChart3,
    FileText,
  }

  const handleGenerateReport = async (reportType: string, format: 'json' | 'csv' | 'pdf') => {
    setLoading(true)
    try {
      const blob = await reportsApi.generateReport(reportType, format)
      reportsApi.downloadReport(blob, `${reportType}-report.${format}`)
    } catch (error) {
      console.error('Failed to generate report:', error)
      setError('Failed to generate report. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  if (!user) {
    return <div>Please log in to access reports.</div>
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50">
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
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="lg:pl-64">
        <header className="bg-white border-b border-gray-200 shadow-sm">
          <div className="px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between py-6">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Reports</h1>
                <p className="mt-1 text-sm text-gray-600">Generate and download system reports</p>
              </div>
              <Badge variant="secondary" className="px-3 py-1">
                <BarChart3 className="w-4 h-4 mr-1" />
                Analytics
              </Badge>
            </div>
          </div>
        </header>

        <main className="px-4 py-8 sm:px-6 lg:px-8">
          {/* Quick Stats */}
          <div className="grid grid-cols-1 gap-6 mb-8 md:grid-cols-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">Total Reports</CardTitle>
                <FileText className="w-4 h-4 text-gray-400" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-gray-900">{stats?.total_reports || 0}</div>
                <p className="text-xs text-gray-500">Available report types</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">This Month</CardTitle>
                <Activity className="w-4 h-4 text-gray-400" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-gray-900">{stats?.this_month || 0}</div>
                <p className="text-xs text-gray-500">Reports generated</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">Data Points</CardTitle>
                <BarChart3 className="w-4 h-4 text-gray-400" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-gray-900">{stats?.data_points?.toLocaleString() || '0'}</div>
                <p className="text-xs text-gray-500">Records analyzed</p>
              </CardContent>
            </Card>
          </div>

          {/* Report Types */}
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            {(reportTypes.length > 0 ? reportTypes : defaultReportTypes).map((report) => {
              const IconComponent = iconMap[report.icon as keyof typeof iconMap] || Package
              const colorClasses = {
                blue: 'bg-blue-50 border-blue-200 text-blue-700',
                green: 'bg-green-50 border-green-200 text-green-700',
                purple: 'bg-purple-50 border-purple-200 text-purple-700',
                orange: 'bg-orange-50 border-orange-200 text-orange-700',
              }

              return (
                <Card key={report.id} className={`transition-all duration-200 ${report.available ? 'hover:shadow-lg' : 'opacity-60'}`}>
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${colorClasses[report.color as keyof typeof colorClasses] || colorClasses.blue}`}>
                          <IconComponent className="w-5 h-5" />
                        </div>
                        <div>
                          <CardTitle className="text-lg font-semibold text-gray-900">
                            {report.name}
                          </CardTitle>
                          <CardDescription className="text-gray-600">
                            {report.description}
                          </CardDescription>
                        </div>
                      </div>
                      <Badge variant={report.available ? 'default' : 'secondary'}>
                        {report.available ? 'Available' : 'Coming Soon'}
                      </Badge>
                    </div>
                  </CardHeader>

                  {report.available && (
                    <CardContent>
                      <div className="flex flex-col gap-3 sm:flex-row">
                        <Button
                          onClick={() => handleGenerateReport(report.id, 'json')}
                          disabled={loading}
                          className="flex-1 text-white bg-blue-600 hover:bg-blue-700"
                        >
                          <Download className="w-4 h-4 mr-2" />
                          {loading ? 'Generating...' : 'Download JSON'}
                        </Button>
                        <Button
                          onClick={() => handleGenerateReport(report.id, 'csv')}
                          disabled={loading}
                          variant="outline"
                          className="flex-1"
                        >
                          <Download className="w-4 h-4 mr-2" />
                          {loading ? 'Generating...' : 'Download CSV'}
                        </Button>
                      </div>
                    </CardContent>
                  )}
                </Card>
              )
            })}
          </div>

          {/* Recent Reports */}
          <Card className="mt-8">
            <CardHeader>
              <CardTitle className="text-lg font-semibold text-gray-900">Recent Reports</CardTitle>
              <CardDescription className="text-gray-600">Your recently generated reports</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="py-8 text-center">
                <FileText className="w-12 h-12 mx-auto text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No recent reports</h3>
                <p className="mt-1 text-sm text-gray-500">Generate your first report to see it here.</p>
              </div>
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  )
}