import { Stock } from '../types/stock';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { StockBadge } from './StockBadge';
import { TrendingUp, Award, Target, Crown, Zap, Star, Medal, Trophy } from 'lucide-react';

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
      return `Price target boosted ${targetIncrease.toFixed(1)}% - Strong upside potential`;
    } else if (hasBuyRating) {
      return 'Analyst confidence with sustained BUY rating';
    } else if (hasOutperform) {
      return 'Outperform rating indicates above-market returns';
    }
    return 'Maintained analyst confidence signals stability';
  };

  const getRankIcon = (index: number) => {
    switch (index) {
      case 0: return <Crown className="h-6 w-6 text-yellow-400" />;
      case 1: return <Medal className="h-6 w-6 text-gray-400" />;
      case 2: return <Trophy className="h-6 w-6 text-amber-600" />;
      default: return <Star className="h-6 w-6 text-primary" />;
    }
  };

  const getRankBg = (index: number) => {
    switch (index) {
      case 0: return 'from-yellow-500/20 to-amber-500/20 border-yellow-500/30';
      case 1: return 'from-slate-500/20 to-gray-500/20 border-slate-500/30';
      case 2: return 'from-amber-500/20 to-orange-500/20 border-amber-500/30';
      default: return 'from-primary/20 to-primary/30 border-primary/30';
    }
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-xl bg-gradient-to-br from-primary/20 to-success/20 animate-glow">
            <Award className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h2 className="text-2xl font-bold flex items-center gap-2">
              Premium Recommendations
              <Zap className="h-5 w-5 text-yellow-400 animate-pulse" />
            </h2>
            <p className="text-muted-foreground">AI-powered analysis of top-performing opportunities</p>
          </div>
        </div>
        <div className="text-right">
          <div className="text-sm text-muted-foreground">
            Confidence Score: <span className="text-primary font-semibold">94.2%</span>
          </div>
          <div className="text-xs text-muted-foreground">
            Updated {new Date().toLocaleTimeString()}
          </div>
        </div>
      </div>
      
      <div className="grid gap-6 md:grid-cols-3">
        {recommendations.map((stock, index) => {
          const targetIncrease = ((stock.targetTo - stock.targetFrom) / stock.targetFrom) * 100;
          
          return (
            <Card 
              key={stock.ticker} 
              className={`
                relative overflow-hidden glass-card border-2 
                bg-gradient-to-br ${getRankBg(index)}
                hover:shadow-premium transition-all duration-500 
                hover:scale-105 hover:-rotate-1 animate-fade-in
                group cursor-pointer
              `}
              style={{ animationDelay: `${index * 0.2}s` }}
            >
              {/* Rank Badge */}
              <div className="absolute -top-3 -right-3 w-12 h-12 bg-gradient-to-br from-background to-muted rounded-full flex items-center justify-center border-2 border-border shadow-lg group-hover:animate-float">
                {getRankIcon(index)}
              </div>
              
              {/* Glow Effect */}
              <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              
              <CardHeader className="pb-4 relative z-10">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <CardTitle className="text-2xl font-mono font-bold text-primary group-hover:text-primary/80 transition-colors">
                      {stock.ticker}
                    </CardTitle>
                    <div className="flex items-center gap-2">
                      <div className="text-xs bg-primary/20 text-primary px-2 py-1 rounded-full font-semibold">
                        #{index + 1} PICK
                      </div>
                      {targetIncrease > 10 && (
                        <div className="text-xs bg-success/20 text-success px-2 py-1 rounded-full font-semibold animate-pulse">
                          HOT 🔥
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                <CardDescription className="text-sm font-medium text-foreground/80 group-hover:text-foreground transition-colors">
                  {stock.company}
                </CardDescription>
              </CardHeader>
              
              <CardContent className="space-y-4 relative z-10">
                <div className="flex items-center justify-between">
                  <StockBadge rating={stock.ratingTo} size="lg" />
                  <div className="text-right">
                    <div className="text-xs text-muted-foreground font-medium">{stock.brokerage}</div>
                    <div className="text-xs text-muted-foreground">{stock.action}</div>
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Target className="h-4 w-4 text-primary" />
                    <span className="text-sm font-medium">Price Target</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="font-mono text-sm text-muted-foreground line-through">
                        ${stock.targetFrom.toFixed(2)}
                      </span>
                      <span className="text-muted-foreground">→</span>
                      <span className="font-mono font-bold text-xl text-primary">
                        ${stock.targetTo.toFixed(2)}
                      </span>
                    </div>
                    {targetIncrease > 0 && (
                      <div className="flex items-center gap-1">
                        <TrendingUp className="h-4 w-4 text-success" />
                        <span className="text-success font-bold text-lg">
                          +{targetIncrease.toFixed(1)}%
                        </span>
                      </div>
                    )}
                  </div>
                  
                  {/* Progress Bar */}
                  <div className="w-full bg-muted/30 rounded-full h-2 overflow-hidden">
                    <div 
                      className="h-full bg-gradient-to-r from-primary to-success rounded-full transition-all duration-1000 ease-out animate-glow"
                      style={{ width: `${Math.min(100, Math.max(10, targetIncrease * 2))}%` }}
                    />
                  </div>
                </div>
                
                <div className="text-xs text-muted-foreground border-t border-border/30 pt-3 italic">
                  💡 {getRecommendationReason(stock)}
                </div>
                
                {/* Action Button */}
                <div className="pt-2">
                  <div className="w-full bg-primary/10 hover:bg-primary/20 text-primary text-center py-2 rounded-lg font-semibold text-sm transition-all duration-200 cursor-pointer group-hover:bg-primary group-hover:text-primary-foreground">
                    View Analysis →
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {recommendations.length === 0 && (
        <Card className="glass-card animate-fade-in">
          <CardContent className="flex items-center justify-center py-12">
            <div className="text-center space-y-4">
              <div className="p-4 rounded-full bg-muted/20 w-fit mx-auto">
                <TrendingUp className="h-12 w-12 text-muted-foreground" />
              </div>
              <div>
                <h3 className="font-semibold text-lg mb-2">No Strong Signals Available</h3>
                <p className="text-muted-foreground">
                  Our AI is analyzing market conditions. Premium recommendations will appear when optimal opportunities are identified.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};