import { ApiResponse } from '../types/stock';

const API_URL = 'https://api.karenai.click/swechallenge/list';
const API_TOKEN = import.meta.env.VITE_STOCK_API_TOKEN;

export const fetchStocks = async (page?: number): Promise<ApiResponse> => {
  const url = page ? `${API_URL}?next_page=${page}` : API_URL;
  const response = await fetch(url, {
    headers: {
      'Authorization': `Token ${API_TOKEN}`,
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch stocks: ${response.statusText}`);
  }

  return response.json();
};