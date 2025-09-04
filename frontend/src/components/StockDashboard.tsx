import { useState, useMemo } from 'react';
import { Stock, StockFilters } from '../types/stock';
import { StockTable } from './StockTable';
import { StockFilters as FiltersComponent } from './StockFilters';
import { StockRecommendations } from './StockRecommendations';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TrendingUp, TrendingDown, BarChart3, Users, Activity, DollarSign, Target, Star } from 'lucide-react';

interface StockDashboardProps {
  stocks: Stock[];
  currentPage: number;
  onPageChange: (page: number) => void;
  loading: boolean;
}

export const StockDashboard = ({ stocks, currentPage, onPageChange, loading }: StockDashboardProps) => {
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
    const targetsRaised = stocks.filter(s => {
      const targetTo = parseFloat(s.target_to.replace('$', ''));
      const targetFrom = parseFloat(s.target_from.replace('$', ''));
      return targetTo > targetFrom;
    }).length;
    const targetsLowered = stocks.filter(s => {
      const targetTo = parseFloat(s.target_to.replace('$', ''));
      const targetFrom = parseFloat(s.target_from.replace('$', ''));
      return targetTo < targetFrom;
    }).length;
    const buyRatings = stocks.filter(s => 
      s.rating_to.toLowerCase().includes('buy') || 
      s.rating_to.toLowerCase().includes('outperform')
    ).length;
    
    const avgPriceChange = stocks.reduce((sum, stock) => {
      const targetTo = parseFloat(stock.target_to.replace('$', ''));
      const targetFrom = parseFloat(stock.target_from.replace('$', ''));
      return sum + ((targetTo - targetFrom) / targetFrom) * 100;
    }, 0) / stocks.length;
    
    return {
      totalStocks,
      targetsRaised,
      targetsLowered,
      buyRatings,
      avgPriceChange
    };
  }, [stocks]);

  return (
    <div className="min-h-screen gradient-hero">
      {/* Professional Header with Glass Effect */}
      <div className="sticky top-0 z-50 glass-card border-b border-border/50 backdrop-blur-xl">
        <div className="container mx-auto px-6 py-8">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-xl bg-primary/20 animate-glow">
                  <Activity className="h-8 w-8 text-primary" />
                </div>
                <div>
                  <h1 className="text-4xl font-bold tracking-tight animate-fade-in">
                    Stock Market <span className="gradient-primary bg-clip-text text-transparent">Intelligence</span>
                  </h1>
                  <p className="text-lg text-muted-foreground animate-slide-up">
                    Advanced analytics • Real-time insights • Professional recommendations
                  </p>
                </div>
              </div>
            </div>
            <div className="text-right animate-scale-in">
              <div className="text-3xl font-bold text-primary font-mono">{stats.totalStocks}</div>
              <div className="text-sm text-muted-foreground font-medium">Active Positions</div>
              <div className={`text-sm font-semibold mt-1 ${stats.avgPriceChange >= 0 ? 'text-success' : 'text-destructive'}`}>
                {stats.avgPriceChange >= 0 ? '+' : ''}{stats.avgPriceChange.toFixed(2)}% Avg Change
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Enhanced Statistics Dashboard */}
      <div className="container mx-auto px-6 py-8">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4 mb-12">
          <Card className="glass-card hover:shadow-premium transition-all duration-300 hover:scale-105 animate-fade-in group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Targets Raised</CardTitle>
              <div className="p-2 rounded-lg bg-success/20 group-hover:animate-pulse-green">
                <TrendingUp className="h-5 w-5 text-success" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-success mb-1">{stats.targetsRaised}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <Target className="h-3 w-3" />
                Bullish analyst revisions
              </p>
              <div className="mt-2 h-1 bg-muted rounded-full overflow-hidden">
                <div 
                  className="h-full bg-success transition-all duration-1000 ease-out"
                  style={{ width: `${(stats.targetsRaised / stats.totalStocks) * 100}%` }}
                />
              </div>
            </CardContent>
          </Card>
          
          <Card className="glass-card hover:shadow-premium transition-all duration-300 hover:scale-105 animate-fade-in group" style={{ animationDelay: '0.1s' }}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Targets Lowered</CardTitle>
              <div className="p-2 rounded-lg bg-destructive/20">
                <TrendingDown className="h-5 w-5 text-destructive" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-destructive mb-1">{stats.targetsLowered}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <Target className="h-3 w-3" />
                Bearish analyst revisions
              </p>
              <div className="mt-2 h-1 bg-muted rounded-full overflow-hidden">
                <div 
                  className="h-full bg-destructive transition-all duration-1000 ease-out"
                  style={{ width: `${(stats.targetsLowered / stats.totalStocks) * 100}%` }}
                />
              </div>
            </CardContent>
          </Card>
          
          <Card className="glass-card hover:shadow-premium transition-all duration-300 hover:scale-105 animate-fade-in group" style={{ animationDelay: '0.2s' }}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Buy Ratings</CardTitle>
              <div className="p-2 rounded-lg bg-primary/20 group-hover:animate-glow">
                <Star className="h-5 w-5 text-primary" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-primary mb-1">{stats.buyRatings}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <BarChart3 className="h-3 w-3" />
                Strong buy signals
              </p>
              <div className="mt-2 h-1 bg-muted rounded-full overflow-hidden">
                <div 
                  className="h-full bg-primary transition-all duration-1000 ease-out animate-glow"
                  style={{ width: `${(stats.buyRatings / stats.totalStocks) * 100}%` }}
                />
              </div>
            </CardContent>
          </Card>
          
          <Card className="glass-card hover:shadow-premium transition-all duration-300 hover:scale-105 animate-fade-in group" style={{ animationDelay: '0.3s' }}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Active Filter</CardTitle>
              <div className="p-2 rounded-lg bg-accent/20">
                <Users className="h-5 w-5 text-accent-foreground" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold mb-1">{filteredStocks.length}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <DollarSign className="h-3 w-3" />
                Matching your criteria
              </p>
              <div className="mt-2 h-1 bg-muted rounded-full overflow-hidden">
                <div 
                  className="h-full bg-accent-foreground transition-all duration-1000 ease-out"
                  style={{ width: `${(filteredStocks.length / stats.totalStocks) * 100}%` }}
                />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Premium Recommendations Section */}
        <div className="mb-12 animate-slide-up" style={{ animationDelay: '0.4s' }}>
          <StockRecommendations stocks={stocks} />
        </div>

        {/* Interactive Filters */}
        <div className="mb-8 animate-slide-up" style={{ animationDelay: '0.5s' }}>
          <FiltersComponent filters={filters} onFiltersChange={setFilters} />
        </div>

        {/* Professional Stock Analysis Table */}
        <div className="space-y-6 animate-slide-up" style={{ animationDelay: '0.6s' }}>
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-semibold flex items-center gap-3">
                <div className="h-8 w-1 bg-primary rounded-full animate-pulse" />
                Market Analysis
              </h2>
              <p className="text-muted-foreground mt-1">
                Professional-grade stock analysis and recommendations
              </p>
            </div>
            <div className="text-right">
              <div className="text-sm text-muted-foreground">
                Displaying <span className="font-semibold text-primary">{filteredStocks.length}</span> of <span className="font-semibold">{stats.totalStocks}</span> positions
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                Last updated: {new Date().toLocaleTimeString()}
              </div>
            </div>
          </div>
          
          <StockTable 
            stocks={filteredStocks} 
            currentPage={currentPage}
            onPageChange={onPageChange}
            loading={loading}
          />
        </div>
      </div>
    </div>
  );
};