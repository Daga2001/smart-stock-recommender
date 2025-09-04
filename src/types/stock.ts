export interface Stock {
  ticker: string;
  company: string;
  brokerage: string;
  action: string;
  rating_from: string;
  rating_to: string;
  target_from: string;
  target_to: string;
  time: string;
}

export interface ApiResponse {
  items: Stock[];
  next_page: string;
}

export type SortField = 'ticker' | 'company' | 'brokerage' | 'target_from' | 'target_to' | 'action';
export type SortDirection = 'asc' | 'desc';

export interface StockFilters {
  search: string;
  action: string;
}