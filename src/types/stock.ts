export interface Stock {
  ticker: string;
  company: string;
  brokerage: string;
  action: string;
  ratingFrom: string;
  ratingTo: string;
  targetFrom: number;
  targetTo: number;
}

export type SortField = 'ticker' | 'company' | 'brokerage' | 'targetFrom' | 'targetTo' | 'action';
export type SortDirection = 'asc' | 'desc';

export interface StockFilters {
  search: string;
  action: string;
}