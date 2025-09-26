import { Category, CreateCategoryRequest, UpdateCategoryRequest } from '@/types'
import { api } from './api'

export const categoriesApi = {
  async getCategories(): Promise<Category[]> {
    try {
      const response = await api.get('/categories')
      return response.data
    } catch (error) {
      console.error('Failed to fetch categories:', error)
      throw error
    }
  },

  async createCategory(category: CreateCategoryRequest): Promise<Category> {
    try {
      const response = await api.post('/categories', category)
      return response.data
    } catch (error) {
      console.error('Failed to create category:', error)
      throw error
    }
  },

  async updateCategory(id: string, category: UpdateCategoryRequest): Promise<Category> {
    try {
      const response = await api.put(`/categories/${id}`, category)
      return response.data
    } catch (error) {
      console.error('Failed to update category:', error)
      throw error
    }
  },

  async deleteCategory(id: string): Promise<void> {
    try {
      await api.delete(`/categories/${id}`)
    } catch (error) {
      console.error('Failed to delete category:', error)
      throw error
    }
  }
}