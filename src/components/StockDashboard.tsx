import { useState, useMemo } from 'react';
import { Stock, StockFilters } from '../types/stock';
import { StockTable } from './StockTable';
import { StockFilters as FiltersComponent } from './StockFilters';
import { StockRecommendations } from './StockRecommendations';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TrendingUp, TrendingDown, BarChart3, Users } from 'lucide-react';

interface StockDashboardProps {
  stocks: Stock[];
}

export const StockDashboard = ({ stocks }: StockDashboardProps) => {
  const [filters, setFilters] = useState<StockFilters>({
    search: '',
    action: 'all'
  });

  const filteredStocks = useMemo(() => {
    return stocks.filter(stock => {
      const searchLower = filters.search.toLowerCase();
      const matchesSearch = !filters.search || 
        stock.ticker.toLowerCase().includes(searchLower) ||
        stock.company.toLowerCase().includes(searchLower) ||
        stock.brokerage.toLowerCase().includes(searchLower);
      
      const matchesAction = filters.action === 'all' || !filters.action || stock.action === filters.action;
      
      return matchesSearch && matchesAction;
    });
  }, [stocks, filters]);

  // Calculate statistics
  const stats = useMemo(() => {
    const totalStocks = stocks.length;
    const targetsRaised = stocks.filter(s => s.targetTo > s.targetFrom).length;
    const targetsLowered = stocks.filter(s => s.targetTo < s.targetFrom).length;
    const buyRatings = stocks.filter(s => 
      s.ratingTo.toLowerCase().includes('buy') || 
      s.ratingTo.toLowerCase().includes('outperform')
    ).length;
    
    return {
      totalStocks,
      targetsRaised,
      targetsLowered,
      buyRatings
    };
  }, [stocks]);

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Stock Analysis Dashboard</h1>
              <p className="text-muted-foreground mt-1">
                Real-time stock recommendations and analyst insights
              </p>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold text-primary">{stats.totalStocks}</div>
              <div className="text-sm text-muted-foreground">Total Stocks</div>
            </div>
          </div>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="container mx-auto px-4 py-6">
        <div className="grid gap-4 md:grid-cols-4 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Targets Raised</CardTitle>
              <TrendingUp className="h-4 w-4 text-success" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-success">{stats.targetsRaised}</div>
              <p className="text-xs text-muted-foreground">
                Price targets increased
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Targets Lowered</CardTitle>
              <TrendingDown className="h-4 w-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">{stats.targetsLowered}</div>
              <p className="text-xs text-muted-foreground">
                Price targets decreased
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Buy Ratings</CardTitle>
              <BarChart3 className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.buyRatings}</div>
              <p className="text-xs text-muted-foreground">
                Buy/Outperform ratings
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Filtered Results</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{filteredStocks.length}</div>
              <p className="text-xs text-muted-foreground">
                Matching your filters
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Recommendations */}
        <div className="mb-8">
          <StockRecommendations stocks={stocks} />
        </div>

        {/* Filters */}
        <div className="mb-6">
          <FiltersComponent filters={filters} onFiltersChange={setFilters} />
        </div>

        {/* Stock Table */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Stock Analysis</h2>
            <div className="text-sm text-muted-foreground">
              Showing {filteredStocks.length} of {stats.totalStocks} stocks
            </div>
          </div>
          
          <StockTable stocks={filteredStocks} />
        </div>
      </div>
    </div>
  );
};