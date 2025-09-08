import type { 
  Stock, 
  StockFilters, 
  FilterOptions, 
  StockRecommendation, 
  ApiResponse, 
  SummaryResponse, 
  ChatResponse,
  ConversationMemory,
  ChatMessage
} from '@/types'

const BASE_URL = 'http://localhost:8081/api'

class ApiService {
  async getStockRatings(pageNumber: number, pageLength: number): Promise<ApiResponse<Stock>> {
    const response = await fetch(`${BASE_URL}/stocks/list`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ page_number: pageNumber, page_length: pageLength })
    })
    if (!response.ok) throw new Error('Failed to fetch stock ratings')
    return response.json()
  }

  async searchStockRatings(filters: StockFilters, pageNumber: number, pageLength: number): Promise<ApiResponse<Stock>> {
    const searchRequest = {
      page_number: pageNumber,
      page_length: pageLength,
      search_term: filters.search.trim(),
      action: filters.action !== 'all' ? filters.action : '',
      rating_from: filters.rating_from !== 'all' ? filters.rating_from : '',
      rating_to: filters.rating_to !== 'all' ? filters.rating_to : '',
      target_from_min: filters.target_from_min || 0,
      target_from_max: filters.target_from_max || 0,
      target_to_min: filters.target_to_min || 0,
      target_to_max: filters.target_to_max || 0
    }

    const response = await fetch(`${BASE_URL}/stocks/search`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(searchRequest)
    })
    if (!response.ok) throw new Error('Failed to search stock ratings')
    return response.json()
  }

  async getFilterOptions(): Promise<FilterOptions> {
    const response = await fetch(`${BASE_URL}/stocks/filter-options`)
    if (!response.ok) throw new Error('Failed to fetch filter options')
    return response.json()
  }

  async getStockActions(): Promise<{ actions: string[] }> {
    const response = await fetch(`${BASE_URL}/stocks/actions`)
    if (!response.ok) throw new Error('Failed to fetch stock actions')
    return response.json()
  }

  async getRecommendations(limit: number = 10): Promise<{ recommendations: StockRecommendation[] }> {
    const response = await fetch(`${BASE_URL}/stocks/recommendations?limit=${limit}`)
    if (!response.ok) throw new Error('Failed to fetch recommendations')
    return response.json()
  }

  async getSummary(): Promise<SummaryResponse> {
    const response = await fetch(`${BASE_URL}/stocks/summary`)
    if (!response.ok) throw new Error('Failed to fetch summary')
    return response.json()
  }

  async sendChatMessage(
    message: string, 
    conversationMemory?: ConversationMemory, 
    recentMessages?: ChatMessage[]
  ): Promise<ChatResponse> {
    const response = await fetch(`${BASE_URL}/stocks/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        conversation_memory: conversationMemory || { summary: '', keyTopics: [], lastContext: '' },
        recent_messages: (recentMessages || []).slice(-4).map(msg => ({
          role: msg.role,
          content: msg.content
        }))
      })
    })
    if (!response.ok) throw new Error('Failed to send chat message')
    return response.json()
  }

  async getMetrics(): Promise<any> {
    const response = await fetch(`${BASE_URL}/stocks/metrics`)
    if (!response.ok) throw new Error('Failed to fetch metrics')
    return response.json()
  }
}

export const apiService = new ApiService()