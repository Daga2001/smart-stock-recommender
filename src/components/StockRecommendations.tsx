import { Stock } from '../types/stock';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { StockBadge } from './StockBadge';
import { TrendingUp, Award, Target } from 'lucide-react';

interface StockRecommendationsProps {
  stocks: Stock[];
}

export const StockRecommendations = ({ stocks }: StockRecommendationsProps) => {
  // Calculate recommendations based on target price increases and buy ratings
  const recommendations = stocks
    .filter(stock => {
      const hasTargetIncrease = stock.targetTo > stock.targetFrom;
      const hasBuyRating = stock.ratingTo.toLowerCase().includes('buy') || 
                          stock.ratingTo.toLowerCase().includes('outperform');
      return hasTargetIncrease || hasBuyRating;
    })
    .sort((a, b) => {
      // Sort by target price increase percentage
      const aIncrease = ((a.targetTo - a.targetFrom) / a.targetFrom) * 100;
      const bIncrease = ((b.targetTo - b.targetFrom) / b.targetFrom) * 100;
      return bIncrease - aIncrease;
    })
    .slice(0, 3);

  const getRecommendationReason = (stock: Stock) => {
    const targetIncrease = ((stock.targetTo - stock.targetFrom) / stock.targetFrom) * 100;
    const hasBuyRating = stock.ratingTo.toLowerCase().includes('buy');
    const hasOutperform = stock.ratingTo.toLowerCase().includes('outperform');
    
    if (targetIncrease > 0) {
      return `Target price increased by ${targetIncrease.toFixed(1)}%`;
    } else if (hasBuyRating) {
      return 'Strong Buy rating maintained';
    } else if (hasOutperform) {
      return 'Outperform rating suggests upside potential';
    }
    return 'Analyst confidence remains high';
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Award className="h-5 w-5 text-primary" />
        <h2 className="text-xl font-semibold">Top Stock Recommendations</h2>
      </div>
      
      <div className="grid gap-4 md:grid-cols-3">
        {recommendations.map((stock, index) => {
          const targetIncrease = ((stock.targetTo - stock.targetFrom) / stock.targetFrom) * 100;
          
          return (
            <Card key={stock.ticker} className="relative overflow-hidden">
              <div className={`absolute top-0 right-0 w-8 h-8 transform rotate-45 translate-x-4 -translate-y-4 ${
                index === 0 ? 'bg-primary' : index === 1 ? 'bg-success' : 'bg-outperform'
              }`} />
              
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg font-mono">{stock.ticker}</CardTitle>
                  <div className="text-right">
                    <div className="text-sm text-muted-foreground">#{index + 1}</div>
                  </div>
                </div>
                <CardDescription className="text-sm font-medium">
                  {stock.company}
                </CardDescription>
              </CardHeader>
              
              <CardContent className="space-y-3">
                <div className="flex items-center justify-between">
                  <StockBadge rating={stock.ratingTo} size="lg" />
                  <div className="text-right">
                    <div className="text-sm text-muted-foreground">by {stock.brokerage}</div>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <Target className="h-4 w-4 text-muted-foreground" />
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm">${stock.targetFrom.toFixed(2)}</span>
                    <span className="text-muted-foreground">→</span>
                    <span className="font-mono font-semibold">${stock.targetTo.toFixed(2)}</span>
                    {targetIncrease > 0 && (
                      <span className="text-success text-sm font-medium">
                        +{targetIncrease.toFixed(1)}%
                      </span>
                    )}
                  </div>
                </div>
                
                <div className="text-xs text-muted-foreground border-t pt-2">
                  {getRecommendationReason(stock)}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {recommendations.length === 0 && (
        <Card>
          <CardContent className="flex items-center justify-center py-8">
            <div className="text-center">
              <TrendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-2" />
              <p className="text-muted-foreground">No strong recommendations available with current data.</p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};