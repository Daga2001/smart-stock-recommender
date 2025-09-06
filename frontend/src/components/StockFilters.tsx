import { StockFilters as FiltersType } from '../types/stock';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Search, Filter, Sparkles, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';

/**
 * Props for the StockFilters component.
 */

interface StockFiltersProps {
  filters: FiltersType;
  onFiltersChange: (filters: FiltersType) => void;
}

/**
 * It's purpose is to provide filtering options for stocks based 
 * on search terms and action types.
 * @param param0 StockFiltersProps
 * @returns 
 */

export const StockFilters = ({ filters, onFiltersChange }: StockFiltersProps) => {
  const clearFilters = () => {
    onFiltersChange({ search: '', action: 'all' });
  };

  return (
    <div className="glass-card p-6 space-y-4 animate-fade-in">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="p-2 rounded-lg bg-primary/20 animate-glow">
            <Sparkles className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h3 className="font-semibold text-lg">Smart Filters</h3>
            <p className="text-sm text-muted-foreground">Refine your search with precision</p>
          </div>
        </div>
        
        {(filters.search || filters.action !== 'all') && (
          <Button 
            onClick={clearFilters}
            variant="outline" 
            size="sm"
            className="hover:bg-destructive/10 hover:border-destructive/50 hover:text-destructive transition-all duration-200"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Clear All
          </Button>
        )}
      </div>

      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-muted-foreground h-5 w-5 z-10" />
          <Input
            placeholder="Search stocks, companies, or brokerages..."
            value={filters.search}
            onChange={(e) => onFiltersChange({ ...filters, search: e.target.value })}
            className="pl-12 h-12 bg-background/50 border border-border/50 focus:border-primary/50 focus:bg-background transition-all duration-200 text-base"
          />
          {filters.search && (
            <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
              <div className="text-xs text-primary font-medium bg-primary/20 px-2 py-1 rounded-full">
                Active
              </div>
            </div>
          )}
        </div>
        
        <div className="flex items-center gap-3 min-w-[250px]">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Filter className="h-5 w-5" />
            <span className="text-sm font-medium hidden sm:block">Action</span>
          </div>
          <Select
            value={filters.action}
            onValueChange={(value) => onFiltersChange({ ...filters, action: value })}
          >
            <SelectTrigger className="h-12 bg-background/50 border border-border/50 focus:border-primary/50 transition-all duration-200">
              <SelectValue placeholder="Filter by action" />
            </SelectTrigger>
            <SelectContent className="bg-popover border border-border/50 shadow-2xl backdrop-blur-xl">
              <SelectItem value="all" className="focus:bg-primary/10 focus:text-primary">
                <span className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-primary rounded-full" />
                  All Actions
                </span>
              </SelectItem>
              <SelectItem value="initiated by" className="focus:bg-success/10 focus:text-success">
                <span className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-success rounded-full" />
                  Initiated
                </span>
              </SelectItem>
              <SelectItem value="target raised by" className="focus:bg-success/10 focus:text-success">
                <span className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-success rounded-full" />
                  Target Raised
                </span>
              </SelectItem>
              <SelectItem value="target lowered by" className="focus:bg-destructive/10 focus:text-destructive">
                <span className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-destructive rounded-full" />
                  Target Lowered
                </span>
              </SelectItem>
              <SelectItem value="reiterated by" className="focus:bg-neutral/10 focus:text-neutral-foreground">
                <span className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-neutral rounded-full" />
                  Reiterated
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Filter Status */}
      {(filters.search || filters.action !== 'all') && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground border-t border-border/30 pt-4">
          <span>Active filters:</span>
          {filters.search && (
            <span className="bg-primary/20 text-primary px-2 py-1 rounded-full text-xs font-medium">
              Search: "{filters.search}"
            </span>
          )}
          {filters.action !== 'all' && (
            <span className="bg-success/20 text-success px-2 py-1 rounded-full text-xs font-medium">
              Action: {filters.action}
            </span>
          )}
        </div>
      )}
    </div>
  );
};