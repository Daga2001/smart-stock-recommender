import { StockFilters as FiltersType } from '../types/stock';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Search, Filter, Sparkles, RefreshCw, DollarSign, Star } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useState, useEffect } from 'react';
import { stockService } from '../services/stockService';

/**
 * Props for the StockFilters component.
 */

interface StockFiltersProps {
  filters: FiltersType;
  onFiltersChange: (filters: FiltersType) => void;
  onApplyFilter: () => void;
  onClearAll?: () => void;
  loading?: boolean;
}

/**
 * It's purpose is to provide filtering options for stocks based 
 * on search terms and action types.
 * @param param0 StockFiltersProps
 * @returns 
 */

export const StockFilters = ({ filters, onFiltersChange, onApplyFilter, onClearAll, loading = false }: StockFiltersProps) => {
  const [availableActions, setAvailableActions] = useState<string[]>([]);
  const [availableRatingsFrom, setAvailableRatingsFrom] = useState<string[]>([]);
  const [availableRatingsTo, setAvailableRatingsTo] = useState<string[]>([]);

  useEffect(() => {
    const loadFilterOptions = async () => {
      try {
        // Load actions
        const actionsResponse = await stockService.getStockActions();
        setAvailableActions(actionsResponse.actions);
        
        // Load all filter options including ratings
        const filterOptionsResponse = await fetch('http://localhost:8081/api/stocks/filter-options');
        if (filterOptionsResponse.ok) {
          const data = await filterOptionsResponse.json();
          console.log('Filter options loaded:', data);
          setAvailableRatingsFrom(data.ratings_from || []);
          setAvailableRatingsTo(data.ratings_to || []);
        } else {
          console.error('Failed to fetch filter options:', filterOptionsResponse.status);
        }
      } catch (error) {
        console.error('Failed to load filter options:', error);
      }
    };
    loadFilterOptions();
  }, []);

  const clearFilters = () => {
    onFiltersChange({ 
      search: '', 
      action: 'all',
      rating_from: 'all',
      rating_to: 'all',
      target_from_min: 0,
      target_from_max: 0,
      target_to_min: 0,
      target_to_max: 0
    });
  };

  const hasActiveFilters = () => {
    return filters.search || 
           filters.action !== 'all' || 
           filters.rating_from !== 'all' || 
           filters.rating_to !== 'all' ||
           filters.target_from_min > 0 ||
           filters.target_from_max > 0 ||
           filters.target_to_min > 0 ||
           filters.target_to_max > 0;
  };

  return (
    <div className="glass-card p-6 space-y-4 animate-fade-in">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="p-2 rounded-lg bg-primary/20 animate-glow">
            <Sparkles className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h3 className="font-semibold text-lg">Filters</h3>
            <p className="text-sm text-muted-foreground">Refine your search with precision</p>
          </div>
        </div>
        
        {hasActiveFilters() && (
          <Button 
            type="button"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              clearFilters();
              onClearAll?.();
            }}
            variant="outline" 
            size="sm"
            className="hover:bg-destructive/10 hover:border-destructive/50 hover:text-destructive transition-all duration-200"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Clear All
          </Button>
        )}
      </div>

      <form onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        onApplyFilter();
      }}>
        {/* Search Bar */}
        <div className="mb-4">
          <div className="relative">
            <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-muted-foreground h-5 w-5 z-10" />
            <Input
              placeholder="Search stocks, companies, or brokerages..."
              value={filters.search}
              onChange={(e) => onFiltersChange({ ...filters, search: e.target.value })}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  onApplyFilter();
                }
              }}
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
        </div>

        {/* Filter Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          {/* Action Filter */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              <Filter className="h-4 w-4" />
              Action
            </Label>
            <Select
              value={filters.action}
              onValueChange={(value) => onFiltersChange({ ...filters, action: value })}
            >
              <SelectTrigger className="h-10 bg-background/50 border border-border/50">
                <SelectValue placeholder="All Actions" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Actions</SelectItem>
                {availableActions.map((action) => (
                  <SelectItem key={action} value={action}>
                    {action.charAt(0).toUpperCase() + action.slice(1)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Rating From Filter */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              <Star className="h-4 w-4" />
              Rating From
            </Label>
            <Select
              value={filters.rating_from}
              onValueChange={(value) => onFiltersChange({ ...filters, rating_from: value })}
            >
              <SelectTrigger className="h-10 bg-background/50 border border-border/50">
                <SelectValue placeholder="All Ratings" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Ratings</SelectItem>
                {availableRatingsFrom.length > 0 ? availableRatingsFrom.map((rating) => (
                  <SelectItem key={rating} value={rating}>
                    {rating}
                  </SelectItem>
                )) : (
                  <SelectItem value="loading" disabled>Loading...</SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Rating To Filter */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              <Star className="h-4 w-4" />
              Rating To
            </Label>
            <Select
              value={filters.rating_to}
              onValueChange={(value) => onFiltersChange({ ...filters, rating_to: value })}
            >
              <SelectTrigger className="h-10 bg-background/50 border border-border/50">
                <SelectValue placeholder="All Ratings" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Ratings</SelectItem>
                {availableRatingsTo.length > 0 ? availableRatingsTo.map((rating) => (
                  <SelectItem key={rating} value={rating}>
                    {rating}
                  </SelectItem>
                )) : (
                  <SelectItem value="loading" disabled>Loading...</SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Apply Button */}
          <div className="space-y-2">
            <Label className="text-sm font-medium opacity-0">Action</Label>
            <Button 
              type="button"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onApplyFilter();
              }}
              disabled={loading}
              className="w-full h-10 bg-primary hover:bg-primary/90 text-primary-foreground font-medium transition-all duration-200 hover:shadow-premium disabled:opacity-50"
            >
              {loading ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  Searching...
                </>
              ) : (
                <>
                  <Search className="h-4 w-4 mr-2" />
                  Apply Filters
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Target Price Ranges */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Target From Range */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              Target From Range
            </Label>
            <div className="flex gap-2">
              <Input
                type="number"
                placeholder="Min"
                value={filters.target_from_min || ''}
                onChange={(e) => onFiltersChange({ ...filters, target_from_min: parseFloat(e.target.value) || 0 })}
                className="h-10 bg-background/50 border border-border/50"
              />
              <Input
                type="number"
                placeholder="Max"
                value={filters.target_from_max || ''}
                onChange={(e) => onFiltersChange({ ...filters, target_from_max: parseFloat(e.target.value) || 0 })}
                className="h-10 bg-background/50 border border-border/50"
              />
            </div>
          </div>

          {/* Target To Range */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              Target To Range
            </Label>
            <div className="flex gap-2">
              <Input
                type="number"
                placeholder="Min"
                value={filters.target_to_min || ''}
                onChange={(e) => onFiltersChange({ ...filters, target_to_min: parseFloat(e.target.value) || 0 })}
                className="h-10 bg-background/50 border border-border/50"
              />
              <Input
                type="number"
                placeholder="Max"
                value={filters.target_to_max || ''}
                onChange={(e) => onFiltersChange({ ...filters, target_to_max: parseFloat(e.target.value) || 0 })}
                className="h-10 bg-background/50 border border-border/50"
              />
            </div>
          </div>
        </div>
      </form>

      {/* Filter Status */}
      {hasActiveFilters() && (
        <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground border-t border-border/30 pt-4">
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
          {filters.rating_from !== 'all' && (
            <span className="bg-blue/20 text-blue px-2 py-1 rounded-full text-xs font-medium">
              From: {filters.rating_from}
            </span>
          )}
          {filters.rating_to !== 'all' && (
            <span className="bg-purple/20 text-purple px-2 py-1 rounded-full text-xs font-medium">
              To: {filters.rating_to}
            </span>
          )}
          {(filters.target_from_min > 0 || filters.target_from_max > 0) && (
            <span className="bg-orange/20 text-orange px-2 py-1 rounded-full text-xs font-medium">
              Target From: ${filters.target_from_min || 0} - ${filters.target_from_max || '∞'}
            </span>
          )}
          {(filters.target_to_min > 0 || filters.target_to_max > 0) && (
            <span className="bg-green/20 text-green px-2 py-1 rounded-full text-xs font-medium">
              Target To: ${filters.target_to_min || 0} - ${filters.target_to_max || '∞'}
            </span>
          )}
        </div>
      )}
    </div>
  );
};