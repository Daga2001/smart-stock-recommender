import { Badge } from "@/components/ui/badge";

interface StockBadgeProps {
  rating: string;
  size?: "sm" | "lg";
}

export const StockBadge = ({ rating, size = "sm" }: StockBadgeProps) => {
  const getRatingVariant = (rating: string) => {
    const lowerRating = rating.toLowerCase();
    if (lowerRating.includes('buy')) return 'buy';
    if (lowerRating.includes('outperform')) return 'outperform';
    if (lowerRating.includes('neutral') || lowerRating.includes('equal')) return 'neutral';
    if (lowerRating.includes('underweight')) return 'underweight';
    return 'secondary';
  };

  const variant = getRatingVariant(rating);

  return (
    <Badge 
      variant={variant as any}
      className={`${size === 'lg' ? 'px-3 py-1 text-sm' : 'px-2 py-0.5 text-xs'} font-medium`}
    >
      {rating}
    </Badge>
  );
};