import { StockFilters as FiltersType } from '../types/stock';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Search, Filter } from 'lucide-react';

interface StockFiltersProps {
  filters: FiltersType;
  onFiltersChange: (filters: FiltersType) => void;
}

export const StockFilters = ({ filters, onFiltersChange }: StockFiltersProps) => {
  return (
    <div className="flex flex-col sm:flex-row gap-4 p-4 bg-card border border-border rounded-lg">
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          placeholder="Search by ticker, company, or brokerage..."
          value={filters.search}
          onChange={(e) => onFiltersChange({ ...filters, search: e.target.value })}
          className="pl-10"
        />
      </div>
      
      <div className="flex items-center gap-2 min-w-[200px]">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <Select
          value={filters.action}
          onValueChange={(value) => onFiltersChange({ ...filters, action: value })}
        >
          <SelectTrigger>
            <SelectValue placeholder="Filter by action" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Actions</SelectItem>
            <SelectItem value="initiated by">Initiated</SelectItem>
            <SelectItem value="target raised by">Target Raised</SelectItem>
            <SelectItem value="target lowered by">Target Lowered</SelectItem>
            <SelectItem value="reiterated by">Reiterated</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
};