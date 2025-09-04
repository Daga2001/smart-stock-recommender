import { useState } from 'react';
import { Stock, SortField, SortDirection } from '../types/stock';
import { StockBadge } from './StockBadge';
import { ArrowUpDown, ArrowUp, ArrowDown, TrendingUp, TrendingDown } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface StockTableProps {
  stocks: Stock[];
}

export const StockTable = ({ stocks }: StockTableProps) => {
  const [sortField, setSortField] = useState<SortField>('ticker');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

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
    if (sortField !== field) return <ArrowUpDown className="ml-2 h-4 w-4" />;
    return sortDirection === 'asc' ? 
      <ArrowUp className="ml-2 h-4 w-4" /> : 
      <ArrowDown className="ml-2 h-4 w-4" />;
  };

  const getTargetTrend = (from: number, to: number) => {
    if (to > from) return <TrendingUp className="h-4 w-4 text-success" />;
    if (to < from) return <TrendingDown className="h-4 w-4 text-destructive" />;
    return null;
  };

  return (
    <div className="rounded-md border border-border bg-card">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('ticker')}
                  className="h-auto p-0 font-medium hover:bg-transparent hover:text-foreground"
                >
                  Ticker
                  <SortIcon field="ticker" />
                </Button>
              </th>
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('company')}
                  className="h-auto p-0 font-medium hover:bg-transparent hover:text-foreground"
                >
                  Company
                  <SortIcon field="company" />
                </Button>
              </th>
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('brokerage')}
                  className="h-auto p-0 font-medium hover:bg-transparent hover:text-foreground"
                >
                  Brokerage
                  <SortIcon field="brokerage" />
                </Button>
              </th>
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                Action
              </th>
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                Rating
              </th>
              <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
                <Button 
                  variant="ghost" 
                  onClick={() => handleSort('targetFrom')}
                  className="h-auto p-0 font-medium hover:bg-transparent hover:text-foreground"
                >
                  Target Price
                  <SortIcon field="targetFrom" />
                </Button>
              </th>
            </tr>
          </thead>
          <tbody>
            {sortedStocks.map((stock, index) => (
              <tr key={stock.ticker} className="border-b border-border last:border-0 hover:bg-muted/50 transition-colors">
                <td className="h-12 px-4 align-middle">
                  <div className="font-mono font-semibold text-primary">
                    {stock.ticker}
                  </div>
                </td>
                <td className="h-12 px-4 align-middle">
                  <div className="font-medium">{stock.company}</div>
                </td>
                <td className="h-12 px-4 align-middle">
                  <div className="text-sm text-muted-foreground">{stock.brokerage}</div>
                </td>
                <td className="h-12 px-4 align-middle">
                  <div className="text-sm">{stock.action}</div>
                </td>
                <td className="h-12 px-4 align-middle">
                  <div className="flex items-center gap-2">
                    <StockBadge rating={stock.ratingFrom} />
                    {stock.ratingFrom !== stock.ratingTo && (
                      <>
                        <span className="text-muted-foreground">→</span>
                        <StockBadge rating={stock.ratingTo} />
                      </>
                    )}
                  </div>
                </td>
                <td className="h-12 px-4 align-middle">
                  <div className="flex items-center gap-2">
                    <span className="font-mono">${stock.targetFrom.toFixed(2)}</span>
                    {stock.targetFrom !== stock.targetTo && (
                      <>
                        <span className="text-muted-foreground">→</span>
                        <span className="font-mono font-semibold">${stock.targetTo.toFixed(2)}</span>
                        {getTargetTrend(stock.targetFrom, stock.targetTo)}
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};