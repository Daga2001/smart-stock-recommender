export interface Stock {
  ticker: string
  company: string
  brokerage: string
  action: string
  rating_from: string
  rating_to: string
  target_from: string
  target_to: string
  time: string
}

export interface StockFilters {
  search: string
  action: string
  rating_from: string
  rating_to: string
  target_from_min: number
  target_from_max: number
  target_to_min: number
  target_to_max: number
}

export interface FilterOptions {
  actions: string[]
  ratings_from: string[]
  ratings_to: string[]
}

export interface StockRecommendation {
  ticker: string
  company: string
  current_rating: string
  target_price: string
  score: number
  recommendation: string
  reason: string
  brokerage: string
  price_change: number
  rating_improvement: boolean
}

export interface ConversationMemory {
  summary: string
  keyTopics: string[]
  lastContext: string
}

export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  context?: string
}

export interface PaginationMeta {
  page_number: number
  page_length: number
  total_records: number
  total_pages: number
  has_next: boolean
  has_previous: boolean
}

export interface ApiResponse<T> {
  data: T[]
  pagination: PaginationMeta
}

export interface SummaryResponse {
  summary: string
  generated_at: string
  tokens_used: number
}

export interface ChatResponse {
  response: string
  tokens_used: number
  generated_at: string
  context_used?: string
  updated_memory?: ConversationMemory
}