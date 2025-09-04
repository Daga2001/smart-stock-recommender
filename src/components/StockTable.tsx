import { useState } from 'react';
import { Stock, SortField, SortDirection } from '../types/stock';
import { StockBadge } from './StockBadge';
import { ArrowUpDown, ArrowUp, ArrowDown, TrendingUp, TrendingDown, Zap, Eye } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface StockTableProps {
  stocks: Stock[];
}

export const StockTable = ({ stocks }: StockTableProps) => {
  const [sortField, setSortField] = useState<SortField>('ticker');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [hoveredRow, setHoveredRow] = useState<string | null>(null);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const sortedStocks = [...stocks].sort((a, b) => {
    let aValue = a[sortField];
    let bValue = b[sortField];

    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase();
      bValue = (bValue as string).toLowerCase();
    }

    if (sortDirection === 'asc') {
      return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
    } else {
      return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
    }
  });

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return <ArrowUpDown className="ml-2 h-4 w-4 opacity-50" />;
    return sortDirection === 'asc' ? 
      <ArrowUp className="ml-2 h-4 w-4 text-primary" /> : 
      <ArrowDown className="ml-2 h-4 w-4 text-primary" />;
  };

  const getTargetTrend = (from: number, to: number) => {
    if (to > from) return <TrendingUp className="h-4 w-4 text-success animate-pulse" />;
    if (to < from) return <TrendingDown className="h-4 w-4 text-destructive animate-pulse" />;
    return <Zap className="h-4 w-4 text-muted-foreground" />;
  };

  const getRowBg = (stock: Stock) => {
    if (stock.targetTo > stock.targetFrom) return 'hover:bg-success/5';
    if (stock.targetTo < stock.targetFrom) return 'hover:bg-destructive/5';
    return 'hover:bg-muted/30';
  };

  return (
    <div className="glass-card border border-border/50 overflow-hidden animate-fade-in">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-gradient-to-r from-muted/30 to-muted/50 backdrop-blur-sm">
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('ticker')}
                  className="h-auto p-0 font-semibold hover:bg-transparent hover:text-primary transition-colors duration-200"
                >
                  Ticker
                  <SortIcon field="ticker" />
                </Button>
              </th>
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('company')}
                  className="h-auto p-0 font-semibold hover:bg-transparent hover:text-primary transition-colors duration-200"
                >
                  Company
                  <SortIcon field="company" />
                </Button>
              </th>
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground hidden md:table-cell">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('brokerage')}
                  className="h-auto p-0 font-semibold hover:bg-transparent hover:text-primary transition-colors duration-200"
                >
                  Brokerage
                  <SortIcon field="brokerage" />
                </Button>
              </th>
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground hidden lg:table-cell">
                Action
              </th>
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground">
                Rating
              </th>
              <th className="h-14 px-6 text-left align-middle font-semibold text-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('targetFrom')}
                  className="h-auto p-0 font-semibold hover:bg-transparent hover:text-primary transition-colors duration-200"
                >
                  Target Price
                  <SortIcon field="targetFrom" />
                </Button>
              </th>
            </tr>
          </thead>
          <tbody>
            {sortedStocks.map((stock, index) => (
              <tr 
                key={stock.ticker} 
                className={`
                  border-b border-border/30 last:border-0 
                  ${getRowBg(stock)} 
                  transition-all duration-300 ease-out
                  ${hoveredRow === stock.ticker ? 'shadow-md shadow-primary/10 scale-[1.02]' : ''}
                  animate-slide-up
                `}
                style={{ 
                  animationDelay: `${index * 0.05}s`,
                  transformOrigin: 'center'
                }}
                onMouseEnter={() => setHoveredRow(stock.ticker)}
                onMouseLeave={() => setHoveredRow(null)}
              >
                <td className="h-16 px-6 align-middle">
                  <div className="flex items-center gap-3">
                    <div className="font-mono font-bold text-lg text-primary relative">
                      {stock.ticker}
                      {hoveredRow === stock.ticker && (
                        <div className="absolute -bottom-1 left-0 right-0 h-0.5 bg-primary rounded-full animate-scale-in" />
                      )}
                    </div>
                    {stock.targetTo > stock.targetFrom && (
                      <div className="w-2 h-2 bg-success rounded-full animate-pulse" />
                    )}
                  </div>
                </td>
                <td className="h-16 px-6 align-middle">
                  <div className="font-medium text-foreground group-hover:text-primary transition-colors">
                    {stock.company}
                  </div>
                </td>
                <td className="h-16 px-6 align-middle hidden md:table-cell">
                  <div className="text-sm text-muted-foreground font-medium">
                    {stock.brokerage}
                  </div>
                </td>
                <td className="h-16 px-6 align-middle hidden lg:table-cell">
                  <div className="text-sm flex items-center gap-2">
                    <Eye className="h-4 w-4 text-muted-foreground" />
                    {stock.action}
                  </div>
                </td>
                <td className="h-16 px-6 align-middle">
                  <div className="flex items-center gap-2">
                    <StockBadge rating={stock.ratingFrom} />
                    {stock.ratingFrom !== stock.ratingTo && (
                      <>
                        <span className="text-muted-foreground animate-pulse">→</span>
                        <StockBadge rating={stock.ratingTo} />
                      </>
                    )}
                  </div>
                </td>
                <td className="h-16 px-6 align-middle">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm text-muted-foreground">
                      ${stock.targetFrom.toFixed(2)}
                    </span>
                    {stock.targetFrom !== stock.targetTo && (
                      <>
                        <span className="text-muted-foreground">→</span>
                        <span className="font-mono font-bold text-lg">
                          ${stock.targetTo.toFixed(2)}
                        </span>
                        {getTargetTrend(stock.targetFrom, stock.targetTo)}
                        <span className={`text-xs font-semibold px-2 py-1 rounded-full ${
                          stock.targetTo > stock.targetFrom 
                            ? 'bg-success/20 text-success' 
                            : 'bg-destructive/20 text-destructive'
                        }`}>
                          {stock.targetTo > stock.targetFrom ? '+' : ''}
                          {(((stock.targetTo - stock.targetFrom) / stock.targetFrom) * 100).toFixed(1)}%
                        </span>
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      
      {stocks.length === 0 && (
        <div className="text-center py-12">
          <div className="text-muted-foreground">No stocks match your current filters.</div>
        </div>
      )}
    </div>
  );
};