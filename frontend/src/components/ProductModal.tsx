'use client'

import { useState, useEffect } from 'react'
import { Product, CreateProductRequest, UpdateProductRequest, Category } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { X } from 'lucide-react'
import { productsApi } from '@/lib/products'
import { categoriesApi } from '@/lib/categories'

interface ProductModalProps {
  isOpen: boolean
  onClose: () => void
  product?: Product | null
}

export function ProductModal({ isOpen, onClose, product }: ProductModalProps) {
  const [formData, setFormData] = useState<CreateProductRequest>({
    name: '',
    sku: '',
    stock: 0,
    price: 0,
    category: '',
    minimum_threshold: 0,
    supplier_info: '',
  })
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (product) {
      setFormData({
        name: product.name,
        sku: product.sku,
        stock: product.stock,
        price: product.price,
        category: product.category,
        minimum_threshold: product.minimum_threshold,
        supplier_info: product.supplier_info || '',
      })
    } else {
      setFormData({
        name: '',
        sku: '',
        stock: 0,
        price: 0,
        category: '',
        minimum_threshold: 0,
        supplier_info: '',
      })
    }
  }, [product, isOpen])

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const categoriesData = await categoriesApi.getCategories()
        setCategories(categoriesData)
      } catch (error) {
        console.error('Failed to fetch categories:', error)
      }
    }

    if (isOpen) {
      fetchCategories()
    }
  }, [isOpen])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      if (product) {
        await productsApi.updateProduct(product.id, formData)
      } else {
        await productsApi.createProduct(formData)
      }
      onClose()
    } catch (error) {
      console.error('Error saving product:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (field: keyof CreateProductRequest, value: string | number) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }))
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">
            {product ? 'Edit Product' : 'Add New Product'}
          </h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="w-5 h-5" />
          </Button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <Label htmlFor="name">Product Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleChange('name', e.target.value)}
              placeholder="Enter product name"
              required
            />
          </div>

          <div>
            <Label htmlFor="sku">SKU *</Label>
            <Input
              id="sku"
              value={formData.sku}
              onChange={(e) => handleChange('sku', e.target.value)}
              placeholder="Enter SKU"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="stock">Stock Quantity *</Label>
              <Input
                id="stock"
                type="number"
                min="0"
                value={formData.stock}
                onChange={(e) => handleChange('stock', parseInt(e.target.value) || 0)}
                required
              />
            </div>
            <div>
              <Label htmlFor="price">Price *</Label>
              <Input
                id="price"
                type="number"
                min="0"
                step="0.01"
                value={formData.price}
                onChange={(e) => handleChange('price', parseFloat(e.target.value) || 0)}
                required
              />
            </div>
          </div>

          <div>
            <Label htmlFor="category">Category *</Label>
            <select
              id="category"
              value={formData.category}
              onChange={(e) => handleChange('category', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
            >
              <option value="">Select a category</option>
              {categories.map((category) => (
                <option key={category.id} value={category.name}>{category.name}</option>
              ))}
            </select>
          </div>

          <div>
            <Label htmlFor="minimum_threshold">Minimum Threshold *</Label>
            <Input
              id="minimum_threshold"
              type="number"
              min="0"
              value={formData.minimum_threshold}
              onChange={(e) => handleChange('minimum_threshold', parseInt(e.target.value) || 0)}
              required
            />
          </div>

          <div>
            <Label htmlFor="supplier_info">Supplier Info</Label>
            <Input
              id="supplier_info"
              value={formData.supplier_info}
              onChange={(e) => handleChange('supplier_info', e.target.value)}
              placeholder="Enter supplier information"
            />
          </div>

          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="flex-1 text-white bg-blue-600 hover:bg-blue-700"
            >
              {loading ? 'Saving...' : (product ? 'Update Product' : 'Create Product')}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}