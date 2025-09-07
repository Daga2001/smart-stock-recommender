import { API_CONFIG, StockListResponse, PaginationRequest, SearchRequest, RecommendationsResponse } from '../config/api';

/**
 * Service class to interact with the stock-related API endpoints.
 */

class StockService {
  private baseUrl = API_CONFIG.BASE_URL;

  // Method to fetch paginated stock ratings
  async getStockRatings(pagination: PaginationRequest): Promise<StockListResponse> {
    const response = await fetch(`${this.baseUrl}${API_CONFIG.ENDPOINTS.STOCKS_LIST}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(pagination),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch stock ratings: ${response.statusText}`);
    }

    return response.json();
  }

  // Method to search stock ratings
  async searchStockRatings(searchParams: SearchRequest): Promise<StockListResponse> {
    const response = await fetch(`${this.baseUrl}${API_CONFIG.ENDPOINTS.STOCKS_SEARCH}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(searchParams),
    });

    if (!response.ok) {
      throw new Error(`Failed to search stock ratings: ${response.statusText}`);
    }

    return response.json();
  }

  // Method to fetch available stock actions
  async getStockActions(): Promise<{actions: string[]}> {
    const response = await fetch(`${this.baseUrl}${API_CONFIG.ENDPOINTS.STOCKS_ACTIONS}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch stock actions: ${response.statusText}`);
    }

    return response.json();
  }

  // Method to fetch stock metrics
  async getStockRecommendations(): Promise<RecommendationsResponse> {
    const response = await fetch(`${this.baseUrl}${API_CONFIG.ENDPOINTS.STOCKS_RECOMMENDATIONS}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch stock recommendations: ${response.statusText}`);
    }

    return response.json();
  }

  async getStockMetrics() {
    const response = await fetch(`${this.baseUrl}${API_CONFIG.ENDPOINTS.STOCKS_METRICS}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch stock metrics: ${response.statusText}`);
    }

    return response.json();
  }
}

export const stockService = new StockService();

// Export types for use in components
export type { RecommendationsResponse, StockRecommendation } from '../config/api';