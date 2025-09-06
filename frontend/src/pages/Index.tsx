import { StockDashboard } from "@/components/StockDashboard";
import { PaginationControls } from "../components/PaginationControls";
import { useStockData } from "../hooks/useStockData";

/**
 * Index Page Component
 * @returns The main index page component displaying stock data and pagination controls.
 */

const Index = () => {
  const {
    stockData,
    pagination,
    loading,
    error,
    pageNumber,
    pageLength,
    handlePageNumberChange,
    handlePageLengthChange,
  } = useStockData();

  if (loading) return <div className="flex justify-center items-center h-64">Loading...</div>;
  if (error) return <div className="text-red-500 text-center">Error: {error}</div>;

  return (
    <StockDashboard 
      stocks={stockData.map(stock => ({
        ticker: stock.ticker || '',
        target_from: stock.target_from || '$0.00',
        target_to: stock.target_to || '$0.00', 
        company: stock.company || '',
        action: stock.action || '',
        brokerage: stock.brokerage || '',
        rating_from: stock.rating_from || '',
        rating_to: stock.rating_to || '',
        time: stock.time || '',
      }))}
      currentPage={pageNumber}
      onPageChange={handlePageNumberChange}
      loading={loading}
      pageLength={pageLength}
      onPageLengthChange={handlePageLengthChange}
      totalPages={pagination?.total_pages || 1}
      totalRecords={pagination?.total_records || 0}
    />
  );
};

export default Index;
