import { StockDashboard } from "@/components/StockDashboard";
import { mockStocks } from "@/data/mockStocks";

const Index = () => {
  return <StockDashboard stocks={mockStocks} />;
};

export default Index;
