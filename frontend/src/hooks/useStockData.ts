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
  const getInitialPageNumber = () => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('currentPageNumber');
      if (saved) {
        return parseInt(saved) || initialPageNumber;
      }
    }
    return initialPageNumber;
  };

  const [pageNumber, setPageNumber] = useState(getInitialPageNumber);

  // Update page number and save to localStorage
  const updatePageNumber = (newPage: number) => {
    setPageNumber(newPage);
    localStorage.setItem('currentPageNumber', newPage.toString());
  };
  const [pageLength, setPageLength] = useState(initialPageLength);
  const [searchFilter, setSearchFilter] = useState<string>('');
  const [isSearchMode, setIsSearchMode] = useState<boolean>(false);

  // Function to fetch stock data from the backend API
  const fetchStockData = async (forceSearchTerm?: string, forceSearchMode?: boolean, forcePageNumber?: number) => {
    setLoading(true);
    setError(null);
    
    try {
      let response;
      const currentSearchTerm = forceSearchTerm !== undefined ? forceSearchTerm : searchFilter;
      const currentSearchMode = forceSearchMode !== undefined ? forceSearchMode : isSearchMode;
      const currentPageNumber = forcePageNumber !== undefined ? forcePageNumber : pageNumber;
      
      console.log('🔄 Fetching page', currentPageNumber, 'with search:', currentSearchTerm, 'mode:', currentSearchMode);
      
      if (currentSearchMode && currentSearchTerm) {
        // Use search endpoint
        const searchResponse = await fetch(`${API_CONFIG.BASE_URL}/api/stocks/search`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            page_number: currentPageNumber,
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
          page_number: currentPageNumber,
          page_length: pageLength,
        });
      }
      
      setStockData(response.data || []);
      setPagination(response.pagination);
      // Update page number to match API response
      if (response.pagination && response.pagination.page_number) {
        updatePageNumber(response.pagination.page_number);
      }
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
    // Don't auto-fetch, let user click refresh to navigate
  };

  const handlePageInputChange = (newPageNumber: number) => {
    updatePageNumber(newPageNumber);
  };

  const handlePageLengthChange = (newPageLength: number) => {
    setPageNumber(1); // Reset to first page when changing page length
    setPageLength(newPageLength);
  };

  const handleRefresh = (search?: string, resetToPageOne?: boolean) => {
    if (search !== undefined) {
      // Check if search is a JSON string (advanced search) or regular search
      let isAdvancedSearch = false;
      let searchRequest = null;
      
      try {
        searchRequest = JSON.parse(search);
        isAdvancedSearch = true;
      } catch {
        // Not JSON, treat as regular search
        isAdvancedSearch = false;
      }
      
      if (isAdvancedSearch && searchRequest) {
        // Advanced search with filters
        setIsSearchMode(true);
        setSearchFilter(search); // Store the full JSON request
        if (resetToPageOne) {
          updatePageNumber(1);
          fetchAdvancedStockData(searchRequest, 1);
        } else {
          fetchAdvancedStockData(searchRequest, pageNumber);
        }
      } else {
        // Regular search
        const newSearchMode = search.length > 0;
        setSearchFilter(search);
        setIsSearchMode(newSearchMode);
        if (resetToPageOne || search === '') {
          updatePageNumber(1);
          fetchStockData(search, newSearchMode, 1);
        } else {
          fetchStockData(search, newSearchMode, pageNumber);
        }
      }
    } else {
      fetchStockData(undefined, undefined, pageNumber);
    }
  };

  // Function to fetch stock data with advanced filters
  const fetchAdvancedStockData = async (searchRequest: any, forcePageNumber?: number) => {
    setLoading(true);
    setError(null);
    
    try {
      const currentPageNumber = forcePageNumber !== undefined ? forcePageNumber : pageNumber;
      
      // Update page number in request
      const requestWithPage = {
        ...searchRequest,
        page_number: currentPageNumber,
        page_length: pageLength
      };
      
      console.log('🔍 Advanced search request:', requestWithPage);
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/stocks/search`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestWithPage),
      });
      
      if (!response.ok) {
        throw new Error('Advanced search failed');
      }
      
      const data = await response.json();
      setStockData(data.data || []);
      setPagination(data.pagination);
      
      if (data.pagination && data.pagination.page_number) {
        updatePageNumber(data.pagination.page_number);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stock data');
    } finally {
      setLoading(false);
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
    handlePageInputChange,
    handlePageLengthChange,
    handleRefresh,
    currentPageNumber: pageNumber,
    handleClearSearch,
    isSearchMode,
    searchFilter,
    refetch: fetchStockData,
    fetchAdvancedStockData,
  };
};