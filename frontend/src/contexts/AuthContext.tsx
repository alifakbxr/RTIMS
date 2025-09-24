'use client'

import React, { createContext, useContext, useEffect, useState } from 'react'
import { User, LoginRequest, RegisterRequest, AuthResponse } from '@/types'
import api from '@/lib/api'

interface AuthContextType {
  user: User | null
  login: (credentials: LoginRequest) => Promise<void>
  register: (userData: RegisterRequest) => Promise<void>
  logout: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check if user is logged in on mount
    const token = localStorage.getItem('access_token')
    if (token) {
      checkAuthStatus()
    } else {
      setLoading(false)
    }
  }, [])

  const checkAuthStatus = async () => {
    try {
      const response = await api.get('/profile')
      setUser(response.data)
    } catch (error) {
      // Token is invalid, remove it
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    } finally {
      setLoading(false)
    }
  }

  const login = async (credentials: LoginRequest) => {
    try {
      const response = await api.post<AuthResponse>('/auth/login', credentials)
      const { user, access_token, token_type } = response.data

      // Store tokens
      localStorage.setItem('access_token', access_token)
      localStorage.setItem('token_type', token_type)

      setUser(user)
    } catch (error) {
      throw new Error('Login failed')
    }
  }

  const register = async (userData: RegisterRequest) => {
    try {
      const response = await api.post<AuthResponse>('/auth/register', userData)
      const { user, access_token, token_type } = response.data

      // Store tokens
      localStorage.setItem('access_token', access_token)
      localStorage.setItem('token_type', token_type)

      setUser(user)
    } catch (error) {
      throw new Error('Registration failed')
    }
  }

  const logout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('token_type')
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, login, register, logout, loading }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}