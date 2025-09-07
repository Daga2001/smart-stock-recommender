import { useState, useMemo } from 'react';
import * as React from 'react';
import { Stock, StockFilters } from '../types/stock';
import { StockTable } from './StockTable';
import { StockFilters as FiltersComponent } from './StockFilters';
import { StockRecommendations } from './StockRecommendations';
import { AIAssistant } from './AIAssistant';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TrendingUp, TrendingDown, BarChart3, Users, Activity, DollarSign, Target, Star } from 'lucide-react';

/**
 * Props for the StockDashboard component.
 */

interface StockDashboardProps {
  stocks: Stock[];
  currentPage: number;
  onPageChange: (page: number) => void;
  loading: boolean;
  pageLength?: number;
  onPageLengthChange?: (length: number) => void;
  onRefresh?: (search?: string, resetToPageOne?: boolean) => void;
  onPageInputChange?: (page: number) => void;
  currentPageNumber?: number;
  totalPages?: number;
  totalRecords?: number;
}

/**
 * StockDashboard Component, a comprehensive dashboard for 
 * displaying stock data with filtering, statistics, and recommendations.
 * @param param0 StockDashboardProps
 * @returns 
 */

export const StockDashboard = ({ 
  stocks, 
  currentPage, 
  onPageChange, 
  loading, 
  pageLength = 20,
  onPageLengthChange,
  onRefresh,
  onPageInputChange,
  currentPageNumber = 1,
  totalPages = 1,
  totalRecords = 0
}: StockDashboardProps) => {
  // Initialize filters from URL params or localStorage
  const getInitialFilters = (): StockFilters => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('stockFilters');
      if (saved) {
        try {
          return JSON.parse(saved);
        } catch (e) {
          // Ignore parsing errors
        }
      }
    }
    return { search: '', action: 'all' };
  };

  const [filters, setFilters] = useState<StockFilters>(getInitialFilters);
  const [appliedSearch, setAppliedSearch] = useState<string>(getInitialFilters().search);

  const filteredStocks = useMemo(() => {
    return stocks.filter(stock => {
      const matchesAction = filters.action === 'all' || !filters.action || stock.action === filters.action;
      return matchesAction;
    });
  }, [stocks, filters.action]);

  // Handle action filter changes
  const handleActionFilterChange = (newFilters: StockFilters) => {
    setFilters(newFilters);
    // Save to localStorage
    localStorage.setItem('stockFilters', JSON.stringify(newFilters));
    // Action filter is applied client-side, no need to refresh data
  };



  // Calculate statistics
  const stats = useMemo(() => {
    const totalStocks = stocks.length;
    const targetsRaised = stocks.filter(s => {
      if (!s.target_to || !s.target_from) return false;
      const targetTo = parseFloat(s.target_to.replace('$', ''));
      const targetFrom = parseFloat(s.target_from.replace('$', ''));
      return !isNaN(targetTo) && !isNaN(targetFrom) && targetTo > targetFrom;
    }).length;
    const targetsLowered = stocks.filter(s => {
      if (!s.target_to || !s.target_from) return false;
      const targetTo = parseFloat(s.target_to.replace('$', ''));
      const targetFrom = parseFloat(s.target_from.replace('$', ''));
      return !isNaN(targetTo) && !isNaN(targetFrom) && targetTo < targetFrom;
    }).length;
    const buyRatings = stocks.filter(s => 
      s.rating_to && (
        s.rating_to.toLowerCase().includes('buy') || 
        s.rating_to.toLowerCase().includes('outperform')
      )
    ).length;
    const uniqueTickers = new Set(stocks.map(s => s.ticker)).size;
    
    const validStocks = stocks.filter(s => s.target_to && s.target_from);
    const avgPriceChange = validStocks.length > 0 ? validStocks.reduce((sum, stock) => {
      const targetTo = parseFloat(stock.target_to.replace('$', ''));
      const targetFrom = parseFloat(stock.target_from.replace('$', ''));
      if (isNaN(targetTo) || isNaN(targetFrom) || targetFrom === 0) return sum;
      return sum + ((targetTo - targetFrom) / targetFrom) * 100;
    }, 0) / validStocks.length : 0;
    
    return {
      totalStocks,
      targetsRaised,
      targetsLowered,
      buyRatings,
      uniqueTickers,
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
                    Advanced analytics • Valuable insights • AI assited recommendations
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
        <div className="space-y-4 mb-12">
          <div>
            <h2 className="text-xl font-semibold mb-2">Market Analytics Overview</h2>
            <p className="text-muted-foreground text-sm">
              Real-time analysis of analyst actions and market sentiment based on target price changes and rating adjustments
            </p>
          </div>
          
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <Card className="glass-card hover:shadow-premium transition-all duration-300 hover:scale-105 animate-fade-in group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <div>
                <CardTitle className="text-sm font-medium text-muted-foreground">Targets Raised</CardTitle>
                <CardDescription className="text-xs mt-1">Analysts increased price expectations</CardDescription>
              </div>
              <div className="p-2 rounded-lg bg-success/20 group-hover:animate-pulse-green">
                <TrendingUp className="h-5 w-5 text-success" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-success mb-1">{stats.targetsRaised}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <Target className="h-3 w-3" />
                Bullish market signals
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
              <div>
                <CardTitle className="text-sm font-medium text-muted-foreground">Targets Lowered</CardTitle>
                <CardDescription className="text-xs mt-1">Analysts reduced price expectations</CardDescription>
              </div>
              <div className="p-2 rounded-lg bg-destructive/20">
                <TrendingDown className="h-5 w-5 text-destructive" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-destructive mb-1">{stats.targetsLowered}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <Target className="h-3 w-3" />
                Bearish market signals
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
              <div>
                <CardTitle className="text-sm font-medium text-muted-foreground">Buy Ratings</CardTitle>
                <CardDescription className="text-xs mt-1">Stocks with Buy/Outperform ratings</CardDescription>
              </div>
              <div className="p-2 rounded-lg bg-primary/20 group-hover:animate-glow">
                <Star className="h-5 w-5 text-primary" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-primary mb-1">{stats.buyRatings}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <BarChart3 className="h-3 w-3" />
                Investment opportunities
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
              <div>
                <CardTitle className="text-sm font-medium text-muted-foreground">Unique Tickers</CardTitle>
                <CardDescription className="text-xs mt-1">Different companies being analyzed</CardDescription>
              </div>
              <div className="p-2 rounded-lg bg-accent/20">
                <BarChart3 className="h-5 w-5 text-accent-foreground" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold mb-1">{stats.uniqueTickers}</div>
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <DollarSign className="h-3 w-3" />
                Market coverage
              </p>
              <div className="mt-2 h-1 bg-muted rounded-full overflow-hidden">
                <div 
                  className="h-full bg-accent-foreground transition-all duration-1000 ease-out"
                  style={{ width: `${Math.min((stats.uniqueTickers / 100) * 100, 100)}%` }}
                />
              </div>
            </CardContent>
          </Card>
          </div>
        </div>

        {/* AI Assistant Section */}
        {/* <div className="mb-12 animate-slide-up" style={{ animationDelay: '0.4s' }}>
          <AIAssistant />
        </div> */}

        {/* Top 3 AI Recommendations Section */}
        <div className="mb-12 animate-slide-up" style={{ animationDelay: '0.5s' }}>
          <StockRecommendations />
        </div>

        {/* Interactive Filters */}
        <div className="mb-8 animate-slide-up" style={{ animationDelay: '0.6s' }}>
          <FiltersComponent 
            filters={filters} 
            onFiltersChange={handleActionFilterChange}
            onApplyFilter={() => {
              const searchTerm = filters.search.trim();
              setAppliedSearch(searchTerm);
              // Save current filters to localStorage
              localStorage.setItem('stockFilters', JSON.stringify(filters));
              // Immediately trigger the refresh with the search term and reset to page 1
              if (searchTerm) {
                onRefresh?.(searchTerm, true); // true = reset to page 1
              } else {
                onRefresh?.('', true); // true = reset to page 1
              }
            }}
            onClearAll={() => {
              setAppliedSearch('');
              // Clear localStorage
              localStorage.removeItem('stockFilters');
              // Refresh with empty search to show all data and reset to page 1
              onRefresh?.('', true); // true = reset to page 1
            }}
            loading={loading}
          />
        </div>

        {/* Professional Stock Analysis Table */}
        <div className="space-y-6 animate-slide-up" style={{ animationDelay: '0.7s' }}>
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
          
          {/* Pagination Controls */}
          <div className="glass-card border border-border/50 p-6 animate-fade-in">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-6">
                <div className="flex items-center gap-3">
                  <label className="text-sm font-semibold text-foreground">Page:</label>
                  <input
                    type="number"
                    min="1"
                    max={totalPages}
                    value={currentPage}
                    onChange={(e) => onPageInputChange?.(Math.max(1, Math.min(totalPages, parseInt(e.target.value) || 1)))}
                    className="w-20 px-3 py-2 glass-card border border-border/50 rounded-lg text-sm font-mono font-semibold text-center focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-premium"
                  />
                  <span className="text-sm text-muted-foreground font-medium">of {totalPages}</span>
                </div>
                
                <div className="flex items-center gap-3">
                  <label className="text-sm font-semibold text-foreground">Show:</label>
                  <select
                    value={pageLength}
                    onChange={(e) => onPageLengthChange?.(parseInt(e.target.value))}
                    className="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-premium cursor-pointer"
                  >
                    <option value={10}>10 per page</option>
                    <option value={20}>20 per page</option>
                    <option value={50}>50 per page</option>
                    <option value={100}>100 per page</option>
                  </select>
                </div>
                
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => {
                      if (!appliedSearch) {
                        onRefresh?.();
                      } else {
                        onRefresh?.(appliedSearch);
                      }
                    }}
                    disabled={loading}
                    className="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium bg-primary/10 hover:bg-primary/20 text-primary focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-premium disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {loading ? 'Loading...' : 'Refresh'}
                  </button>
                  <div className="text-xs text-muted-foreground">
                    💡 Click refresh after changing page number to navigate through the table
                  </div>
                </div>
              </div>
              
              <div className="text-sm text-muted-foreground font-medium">
                <span className="text-primary font-semibold">{totalRecords.toLocaleString()}</span> total records
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