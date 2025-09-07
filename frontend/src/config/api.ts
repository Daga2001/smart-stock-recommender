// API Configuration

/**
 * The goal of this file is to centralize API configuration details,
 * such as base URLs and endpoint paths, to facilitate easy updates
 * and maintenance.
 */

export const API_CONFIG = {
  BASE_URL: 'http://localhost:8081',
  ENDPOINTS: {
    STOCKS_LIST: '/api/stocks/list',
    STOCKS_SEARCH: '/api/stocks/search',
    STOCKS_ACTIONS: '/api/stocks/actions',
    STOCKS_RECOMMENDATIONS: '/api/stocks/recommendations',
    STOCKS_METRICS: '/api/stocks/metrics',
    STOCKS_BULK: '/api/stocks/bulk',
    STOCKS_SINGLE: '/api/stocks'
  }
} as const;

// API Response Types
export interface StockRating {
  id: number;
  ticker: string;
  target_from: string;
  target_to: string;
  company: string;
  action: string;
  brokerage: string;
  rating_from: string;
  rating_to: string;
  time: string;
  created_at: string;
}

// Pagination Metadata
export interface PaginationMeta {
  page_number: number;
  page_length: number;
  total_records: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
}

// Stock List Response
export interface StockListResponse {
  data: StockRating[];
  pagination: PaginationMeta;
}

export interface PaginationRequest {
  page_number: number;
  page_length: number;
}

export interface SearchRequest {
  page_number: number;
  page_length: number;
  search_term: string;
}

export interface StockRecommendation {
  ticker: string;
  company: string;
  current_rating: string;
  target_price: string;
  score: number;
  recommendation: string;
  reason: string;
  brokerage: string;
  price_change: number;
  rating_improvement: boolean;
}

export interface RecommendationsResponse {
  recommendations: StockRecommendation[];
  generated_at: string;
  total_analyzed: number;
}