import { useState, useEffect } from 'react';
import { stockService } from '../services/stockService';
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

  // Function to fetch stock data from the backend API
  const fetchStockData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await stockService.getStockRatings({
        page_number: pageNumber,
        page_length: pageLength,
      });
      
      setStockData(response.data);
      setPagination(response.pagination);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stock data');
    } finally {
      setLoading(false);
    }
  };

  // Fetch data when pagination parameters change
  useEffect(() => {
    fetchStockData();
  }, [pageNumber, pageLength]);

  const handlePageNumberChange = (newPageNumber: number) => {
    setPageNumber(newPageNumber);
  };

  const handlePageLengthChange = (newPageLength: number) => {
    setPageNumber(1); // Reset to first page when changing page length
    setPageLength(newPageLength);
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
    refetch: fetchStockData,
  };
};