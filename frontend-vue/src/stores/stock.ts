import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Stock, StockFilters, FilterOptions, PaginationMeta } from '@/types'
import { apiService } from '@/services/api'

export const useStockStore = defineStore('stock', () => {
  // State
  const stocks = ref<Stock[]>([])
  const pagination = ref<PaginationMeta | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filterOptions = ref<FilterOptions>({ actions: [], ratings_from: [], ratings_to: [] })
  
  // Initialize filters from localStorage
  const getInitialFilters = (): StockFilters => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('stockFilters')
      if (saved) {
        try {
          const parsed = JSON.parse(saved)
          return {
            search: parsed.search || '',
            action: parsed.action || 'all',
            rating_from: parsed.rating_from || 'all',
            rating_to: parsed.rating_to || 'all',
            target_from_min: parsed.target_from_min || 0,
            target_from_max: parsed.target_from_max || 0,
            target_to_min: parsed.target_to_min || 0,
            target_to_max: parsed.target_to_max || 0
          }
        } catch (e) {
          // Ignore parsing errors
        }
      }
    }
    return {
      search: '',
      action: 'all',
      rating_from: 'all',
      rating_to: 'all',
      target_from_min: 0,
      target_from_max: 0,
      target_to_min: 0,
      target_to_max: 0
    }
  }

  const filters = ref<StockFilters>(getInitialFilters())
  const currentPage = ref(1)
  const pageLength = ref(20)

  // Getters
  const hasActiveFilters = computed(() => {
    return filters.value.search ||
           filters.value.action !== 'all' ||
           filters.value.rating_from !== 'all' ||
           filters.value.rating_to !== 'all' ||
           filters.value.target_from_min > 0 ||
           filters.value.target_from_max > 0 ||
           filters.value.target_to_min > 0 ||
           filters.value.target_to_max > 0
  })

  // Actions
  const loadFilterOptions = async () => {
    try {
      const options = await apiService.getFilterOptions()
      filterOptions.value = options
    } catch (err) {
      console.error('Failed to load filter options:', err)
    }
  }

  const loadStocks = async (page: number = 1) => {
    loading.value = true
    error.value = null
    currentPage.value = page

    try {
      let response
      if (hasActiveFilters.value) {
        response = await apiService.searchStockRatings(filters.value, page, pageLength.value)
      } else {
        response = await apiService.getStockRatings(page, pageLength.value)
      }
      
      stocks.value = response.data
      pagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load stocks'
    } finally {
      loading.value = false
    }
  }

  const applyFilters = () => {
    // Save filters to localStorage
    localStorage.setItem('stockFilters', JSON.stringify(filters.value))
    // Reset to page 1 and load with filters
    loadStocks(1)
  }

  const clearFilters = () => {
    filters.value = {
      search: '',
      action: 'all',
      rating_from: 'all',
      rating_to: 'all',
      target_from_min: 0,
      target_from_max: 0,
      target_to_min: 0,
      target_to_max: 0
    }
    localStorage.removeItem('stockFilters')
    loadStocks(1)
  }

  const setPageLength = (length: number) => {
    pageLength.value = length
    loadStocks(1)
  }

  const goToPage = (page: number) => {
    loadStocks(page)
  }

  return {
    // State
    stocks,
    pagination,
    loading,
    error,
    filters,
    filterOptions,
    currentPage,
    pageLength,
    
    // Getters
    hasActiveFilters,
    
    // Actions
    loadFilterOptions,
    loadStocks,
    applyFilters,
    clearFilters,
    setPageLength,
    goToPage
  }
})