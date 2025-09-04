import { StockDashboard } from "@/components/StockDashboard";
import { useState, useEffect } from "react";
import { Stock } from "@/types/stock";
import { fetchStocks } from "@/services/stockService";

const Index = () => {
  const [allStocks, setAllStocks] = useState<Stock[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [pageLoading, setPageLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stocks, setStocks] = useState<Stock[]>([]);

  const loadStocks = async (page: number, isInitial = false) => {
    try {
      if (isInitial) {
        setLoading(true);
      } else {
        setPageLoading(true);
      }
      const data = await fetchStocks(page);
      setStocks(data.items);
      setCurrentPage(page);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load stocks');
    } finally {
      if (isInitial) {
        setLoading(false);
      } else {
        setPageLoading(false);
      }
    }
  };

  useEffect(() => {
    loadStocks(1, true);
  }, []);

  const handlePageChange = (page: number) => {
    if (page >= 1) {
      loadStocks(page, false);
    }
  };

  if (loading) return <div className="flex justify-center items-center h-64">Loading...</div>;
  if (error) return <div className="text-red-500 text-center">Error: {error}</div>;

  return (
    <StockDashboard 
      stocks={stocks} 
      currentPage={currentPage}
      onPageChange={handlePageChange}
      loading={pageLoading}
    />
  );
};

export default Index;
