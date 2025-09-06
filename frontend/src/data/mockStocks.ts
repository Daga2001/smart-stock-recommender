import { Stock } from '../types/stock';

/**
 * mockStocks provides a list of mock stock data for testing and development purposes.
 */

export const mockStocks: Stock[] = [
  {
    ticker: "AKBA",
    company: "Akebia Therapeutics",
    brokerage: "HC Wainwright",
    action: "initiated by",
    ratingFrom: "Buy",
    ratingTo: "Buy",
    targetFrom: 8.00,
    targetTo: 8.00
  },
  {
    ticker: "CECO",
    company: "CECO Environmental",
    brokerage: "Needham & Company LLC",
    action: "target raised by",
    ratingFrom: "Buy",
    ratingTo: "Buy",
    targetFrom: 44.00,
    targetTo: 52.00
  },
  {
    ticker: "BLND",
    company: "Blend Labs",
    brokerage: "Canaccord Genuity Group",
    action: "reiterated by",
    ratingFrom: "Buy",
    ratingTo: "Buy",
    targetFrom: 5.25,
    targetTo: 5.25
  },
  {
    ticker: "FLOC",
    company: "Flowco",
    brokerage: "Evercore ISI",
    action: "target lowered by",
    ratingFrom: "Outperform",
    ratingTo: "Outperform",
    targetFrom: 28.00,
    targetTo: 26.00
  },
  {
    ticker: "VYGR",
    company: "Voyager Therapeutics",
    brokerage: "Wedbush",
    action: "target lowered by",
    ratingFrom: "Outperform",
    ratingTo: "Outperform",
    targetFrom: 9.00,
    targetTo: 8.00
  },
  {
    ticker: "BCBP",
    company: "BCB Bancorp, Inc. (NJ)",
    brokerage: "Piper Sandler",
    action: "target raised by",
    ratingFrom: "Neutral",
    ratingTo: "Neutral",
    targetFrom: 9.00,
    targetTo: 9.50
  },
  {
    ticker: "DEFT",
    company: "DeFi Technologies",
    brokerage: "HC Wainwright",
    action: "reiterated by",
    ratingFrom: "Buy",
    ratingTo: "Buy",
    targetFrom: 5.50,
    targetTo: 5.50
  },
  {
    ticker: "LAMR",
    company: "Lamar Advertising",
    brokerage: "Wells Fargo & Company",
    action: "target lowered by",
    ratingFrom: "Equal Weight",
    ratingTo: "Equal Weight",
    targetFrom: 122.00,
    targetTo: 119.00
  },
  {
    ticker: "TRIN",
    company: "Trinity Capital",
    brokerage: "Wells Fargo & Company",
    action: "reiterated by",
    ratingFrom: "Underweight",
    ratingTo: "Underweight",
    targetFrom: 13.00,
    targetTo: 13.50
  },
  {
    ticker: "AMP",
    company: "Ameriprise Financial",
    brokerage: "Royal Bank Of Canada",
    action: "target raised by",
    ratingFrom: "Outperform",
    ratingTo: "Outperform",
    targetFrom: 595.00,
    targetTo: 601.00
  }
];