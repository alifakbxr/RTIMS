'use client'

import { useEffect, useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Navigation } from '@/components/ui/navigation'
import { User, UserRole, CreateUserRequest, UpdateUserRequest } from '@/types'
import { Search, Plus, Edit, Trash2, Users, Shield, UserCheck, AlertTriangle, X } from 'lucide-react'
import { usersApi } from '@/lib/users'

export default function UsersPage() {
  const { user } = useAuth()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedRole, setSelectedRole] = useState<UserRole | ''>('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [newUser, setNewUser] = useState<CreateUserRequest>({
    name: '',
    email: '',
    password: '',
    role: 'staff'
  })

  useEffect(() => {
    fetchUsers()
  }, [searchTerm, selectedRole])

  const fetchUsers = async () => {
    try {
      setLoading(true)
      setError(null)

      const filters = {
        search: searchTerm,
        role: selectedRole
      }

      const data = await usersApi.getUsers(filters)
      setUsers(data)
    } catch (error) {
      console.error('Failed to fetch users:', error)
      setError('Failed to load users. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const getRoleIcon = (role: UserRole) => {
    return role === 'admin' ? <Shield className="w-4 h-4" /> : <UserCheck className="w-4 h-4" />
  }

  const getRoleBadgeVariant = (role: UserRole) => {
    return role === 'admin' ? 'default' : 'secondary'
  }

  const handleCreateUser = () => {
    setEditingUser(null)
    setNewUser({
      name: '',
      email: '',
      password: '',
      role: 'staff'
    })
    setShowCreateModal(true)
  }

  const handleEditUser = (user: User) => {
    setEditingUser(user)
    setNewUser({
      name: user.name,
      email: user.email,
      password: '',
      role: user.role
    })
    setShowCreateModal(true)
  }

  const handleDeleteUser = async (userId: string) => {
    if (!confirm('Are you sure you want to delete this user?')) return

    try {
      await usersApi.deleteUser(userId)
      fetchUsers()
    } catch (error) {
      console.error('Failed to delete user:', error)
      setError('Failed to delete user. Please try again.')
    }
  }

  const handleSaveUser = async () => {
    try {
      setError(null)
      if (editingUser) {
        const updateData: UpdateUserRequest = {
          name: newUser.name,
          email: newUser.email,
          role: newUser.role
        }
        await usersApi.updateUser(editingUser.id, updateData)
      } else {
        await usersApi.createUser(newUser)
      }
      setShowCreateModal(false)
      fetchUsers()
    } catch (error) {
      console.error('Failed to save user:', error)
      setError('Failed to save user. Please try again.')
    }
  }

  const handleToggleUserStatus = async (userId: string, isActive: boolean) => {
    try {
      await usersApi.toggleUserStatus(userId, isActive)
      fetchUsers()
    } catch (error) {
      console.error('Failed to toggle user status:', error)
      setError('Failed to update user status. Please try again.')
    }
  }

  if (!user) {
    return <div>Please log in to access users.</div>
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
                <h1 className="text-3xl font-bold text-gray-900">Users</h1>
                <p className="mt-1 text-sm text-gray-600">Manage system users</p>
              </div>
              <Button
                onClick={handleCreateUser}
                className="flex items-center gap-2 px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                Add User
              </Button>
            </div>
          </div>
        </header>

        <main className="px-4 py-8 sm:px-6 lg:px-8">
          {/* Filters */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-lg">Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="relative">
                  <Search className="absolute w-4 h-4 text-gray-400 transform -translate-y-1/2 left-3 top-1/2" />
                  <Input
                    placeholder="Search users..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                <select
                  value={selectedRole}
                  onChange={(e) => setSelectedRole(e.target.value as UserRole | '')}
                  className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">All Roles</option>
                  <option value="admin">Admin</option>
                  <option value="staff">Staff</option>
                </select>
              </div>
            </CardContent>
          </Card>

          {/* Users Grid */}
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-b-2 border-blue-600 rounded-full animate-spin"></div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
              {users.map((user) => (
                <Card key={user.id} className="transition-shadow duration-200 hover:shadow-lg">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="flex items-center justify-center w-10 h-10 bg-blue-100 rounded-full">
                          <span className="text-sm font-medium text-blue-600">
                            {user.name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div>
                          <CardTitle className="text-lg font-semibold text-gray-900">
                            {user.name}
                          </CardTitle>
                          <CardDescription className="text-sm text-gray-600">
                            {user.email}
                          </CardDescription>
                        </div>
                      </div>
                      <Badge
                        variant={getRoleBadgeVariant(user.role)}
                        className="flex items-center gap-1"
                      >
                        {getRoleIcon(user.role)}
                        {user.role}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Status:</span>
                        <Badge variant={user.is_active ? 'default' : 'destructive'}>
                          {user.is_active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Created:</span>
                        <span className="text-sm text-gray-900">
                          {new Date(user.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      <div className="flex gap-2 pt-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleEditUser(user)}
                          className="flex-1"
                        >
                          <Edit className="w-4 h-4 mr-1" />
                          Edit
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDeleteUser(user.id)}
                          className="flex-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="w-4 h-4 mr-1" />
                          Delete
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}

          {!loading && users.length === 0 && (
            <div className="py-12 text-center">
              <Users className="w-12 h-12 mx-auto text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No users found</h3>
              <p className="mt-1 text-sm text-gray-500">No users match your current filters.</p>
            </div>
          )}
        </main>
      </div>

      {/* User Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">
                {editingUser ? 'Edit User' : 'Add New User'}
              </h2>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowCreateModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-5 h-5" />
              </Button>
            </div>

            <form onSubmit={(e) => { e.preventDefault(); handleSaveUser(); }} className="p-6 space-y-4">
              <div>
                <Label htmlFor="name">Name *</Label>
                <Input
                  id="name"
                  value={newUser.name}
                  onChange={(e) => setNewUser({...newUser, name: e.target.value})}
                  placeholder="Enter full name"
                  required
                />
              </div>

              <div>
                <Label htmlFor="email">Email *</Label>
                <Input
                  id="email"
                  type="email"
                  value={newUser.email}
                  onChange={(e) => setNewUser({...newUser, email: e.target.value})}
                  placeholder="Enter email address"
                  required
                />
              </div>

              <div>
                <Label htmlFor="password">Password *</Label>
                <Input
                  id="password"
                  type="password"
                  value={newUser.password}
                  onChange={(e) => setNewUser({...newUser, password: e.target.value})}
                  placeholder="Enter password"
                  required={!editingUser}
                />
                {editingUser && (
                  <p className="mt-1 text-sm text-gray-500">Leave blank to keep current password</p>
                )}
              </div>

              <div>
                <Label htmlFor="role">Role *</Label>
                <select
                  id="role"
                  value={newUser.role}
                  onChange={(e) => setNewUser({...newUser, role: e.target.value as UserRole})}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                >
                  <option value="staff">Staff</option>
                  <option value="admin">Admin</option>
                </select>
              </div>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  className="flex-1 text-white bg-blue-600 hover:bg-blue-700"
                >
                  {editingUser ? 'Update User' : 'Create User'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}