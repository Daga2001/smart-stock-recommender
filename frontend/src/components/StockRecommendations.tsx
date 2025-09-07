import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { StockBadge } from './StockBadge';
import { TrendingUp, Award, Target, Crown, Zap, Star, Medal, Trophy, Bot, RefreshCw } from 'lucide-react';
import { stockService, type StockRecommendation } from '../services/stockService';
import { Button } from '@/components/ui/button';

// Props for the StockRecommendations component.
interface StockRecommendationsProps {
  // No props needed - fetches data from API
}

/**
 * It's purpose is to analyze stock data and highlight top recommendations based on target price increases and buy ratings.
 * @param param0 StockRecommendationsProps
 * @returns 
 */

export const StockRecommendations = ({}: StockRecommendationsProps) => {
  const [recommendations, setRecommendations] = useState<StockRecommendation[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadRecommendations = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await stockService.getStockRecommendations();
      setRecommendations(response.recommendations.slice(0, 3)); // Top 3 only
    } catch (err) {
      setError('Failed to load recommendations');
      console.error('Failed to load recommendations:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRecommendations();
  }, []);

  const getRecommendationReason = (rec: StockRecommendation) => {
    return rec.reason || 'AI-powered analysis indicates strong potential';
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
              Top 3 AI Recommendations
              <Bot className="h-5 w-5 text-primary animate-pulse" />
            </h2>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Algorithm considers: <span className="font-semibold">Target price changes</span>, <span className="font-semibold">Rating improvements</span> (Buy, Outperform), and <span className="font-semibold">Analyst sentiment</span>
            </p>
          </div>
        </div>
        <div className="text-right">
          <Button 
            onClick={loadRecommendations} 
            variant="outline" 
            size="sm"
            disabled={loading}
            className="mb-2"
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            {loading ? 'Loading...' : 'Refresh'}
          </Button>
          <div className="text-xs text-muted-foreground">
            AI-powered analysis
          </div>
        </div>
      </div>
      
      {(() => {
        if (loading) {
          return (
            <div className="flex items-center justify-center py-12">
              <RefreshCw className="h-8 w-8 animate-spin text-primary" />
              <span className="ml-2 text-muted-foreground">Loading AI recommendations...</span>
            </div>
          );
        }
        
        if (error) {
          return (
            <div className="text-center py-12">
              <p className="text-destructive mb-4">{error}</p>
              <Button onClick={loadRecommendations} variant="outline">
                <RefreshCw className="h-4 w-4 mr-2" />
                Try Again
              </Button>
            </div>
          );
        }
        
        if (recommendations.length === 0) {
          return (
            <div className="text-center py-12 text-muted-foreground">
              <Bot className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No recommendations available at this time.</p>
            </div>
          );
        }
        
        return (
          <div className="grid gap-6 md:grid-cols-3">
          {recommendations.map((rec, index) => {
            return (
            <Card 
              key={rec.ticker} 
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
                      {rec.ticker}
                    </CardTitle>
                    <div className="flex items-center gap-2">
                      <div className="text-xs bg-primary/20 text-primary px-2 py-1 rounded-full font-semibold">
                        #{index + 1} PICK
                      </div>
                      <div className="text-xs bg-success/20 text-success px-2 py-1 rounded-full font-semibold">
                        {rec.recommendation}
                      </div>
                      {rec.price_change > 10 && (
                        <div className="text-xs bg-destructive/20 text-destructive px-2 py-1 rounded-full font-semibold animate-pulse">
                          HOT 🔥
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                <CardDescription className="text-sm font-medium text-foreground/80 group-hover:text-foreground transition-colors">
                  {rec.company}
                </CardDescription>
              </CardHeader>
              
              <CardContent className="space-y-4 relative z-10">
                <div className="flex items-center justify-between">
                  <StockBadge rating={rec.current_rating} size="lg" />
                  <div className="text-right">
                    <div className="text-xs text-muted-foreground font-medium">{rec.brokerage}</div>
                    <div className="text-xs font-semibold text-primary">Score: {rec.score.toFixed(1)}/10</div>
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Target className="h-4 w-4 text-primary" />
                    <span className="text-sm font-medium">Price Target</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="font-mono font-bold text-xl text-primary">
                        {rec.target_price}
                      </span>
                    </div>
                    {rec.price_change > 0 && (
                      <div className="flex items-center gap-1">
                        <TrendingUp className="h-4 w-4 text-success" />
                        <span className="text-success font-bold text-lg">
                          +{rec.price_change.toFixed(1)}%
                        </span>
                      </div>
                    )}
                  </div>
                  
                  {/* Progress Bar */}
                  <div className="w-full bg-muted/30 rounded-full h-2 overflow-hidden">
                    <div 
                      className="h-full bg-gradient-to-r from-primary to-success transition-all duration-1000 ease-out"
                      style={{ width: `${Math.min(rec.score * 10, 100)}%` }}
                    />
                  </div>
                </div>
                
                <div className="pt-2 border-t border-border/30">
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    <span className="font-semibold text-foreground">Why recommended:</span> {getRecommendationReason(rec)}
                  </p>
                </div>
              </CardContent>
            </Card>
            );
          })}
          </div>
        );
      })()}
    </div>
  );
};