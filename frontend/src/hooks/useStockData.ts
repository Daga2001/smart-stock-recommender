import { useState, useEffect } from 'react';
import { stockService } from '../services/stockService';
import { API_CONFIG } from '../config/api';
import { StockRating, PaginationMeta } from '../config/api';

/**
 * Custom hook to manage stock data fetching with pagination.
 * @param initialPageNumber 
 * @param initialPageLength 
 * @returns 
 */

export const useStockData = (initialPageNumber = 1, initialPageLength = 20) => {
  const [stockData, setStockData] = useState<StockRating[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pageNumber, setPageNumber] = useState(initialPageNumber);
  const [pageLength, setPageLength] = useState(initialPageLength);
  const [searchFilter, setSearchFilter] = useState<string>('');
  const [isSearchMode, setIsSearchMode] = useState<boolean>(false);

  // Function to fetch stock data from the backend API
  const fetchStockData = async (forceSearchTerm?: string, forceSearchMode?: boolean) => {
    setLoading(true);
    setError(null);
    
    try {
      let response;
      const currentSearchTerm = forceSearchTerm !== undefined ? forceSearchTerm : searchFilter;
      const currentSearchMode = forceSearchMode !== undefined ? forceSearchMode : isSearchMode;
      
      if (currentSearchMode && currentSearchTerm) {
        // Use search endpoint
        const searchResponse = await fetch(`${API_CONFIG.BASE_URL}/api/stocks/search`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            page_number: pageNumber,
            page_length: pageLength,
            search_term: currentSearchTerm,
          }),
        });
        
        if (!searchResponse.ok) {
          throw new Error('Search failed');
        }
        
        response = await searchResponse.json();
      } else {
        // Use regular list endpoint
        response = await stockService.getStockRatings({
          page_number: pageNumber,
          page_length: pageLength,
        });
      }
      
      setStockData(response.data || []);
      setPagination(response.pagination);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stock data');
    } finally {
      setLoading(false);
    }
  };

  // Fetch data on initial load
  useEffect(() => {
    fetchStockData();
  }, []);

  // Auto-refresh when page length changes
  useEffect(() => {
    // Skip the initial render, but refresh on any subsequent pageLength change
    if (pagination !== null) {
      fetchStockData();
    }
  }, [pageLength]);

  const handlePageNumberChange = (newPageNumber: number) => {
    setPageNumber(newPageNumber);
  };

  const handlePageLengthChange = (newPageLength: number) => {
    setPageNumber(1); // Reset to first page when changing page length
    setPageLength(newPageLength);
  };

  const handleRefresh = (search?: string) => {
    if (search !== undefined) {
      const newSearchMode = search.length > 0;
      setSearchFilter(search);
      setIsSearchMode(newSearchMode);
      setPageNumber(1);
      // Immediately fetch with the new search parameters
      fetchStockData(search, newSearchMode);
    } else {
      fetchStockData();
    }
  };

  const handleClearSearch = () => {
    setSearchFilter('');
    setIsSearchMode(false);
    setPageNumber(1);
    fetchStockData();
  };

  return {
    stockData,
    pagination,
    loading,
    error,
    pageNumber,
    pageLength,
    handlePageNumberChange,
    handlePageLengthChange,
    handleRefresh,
    handleClearSearch,
    isSearchMode,
    searchFilter,
    refetch: fetchStockData,
  };
};