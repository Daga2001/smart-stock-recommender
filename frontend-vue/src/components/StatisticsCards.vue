<template>
  <div class="space-y-4 mb-12">
    <div>
      <h2 class="text-xl font-semibold mb-2">Market Analytics Overview</h2>
      <p class="text-muted-foreground text-sm">
        Statistical analysis of brokerage actions from the current page, based on target price changes and rating adjustments
      </p>
    </div>
    
    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
      <!-- Targets Raised -->
      <div class="glass-card hover:shadow-lg transition-all duration-300 hover:scale-105 animate-fade-in group p-6 rounded-lg border border-border/50">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-sm font-medium text-muted-foreground">Targets Raised</h3>
            <p class="text-xs mt-1 text-muted-foreground">Analysts increased price expectations</p>
          </div>
          <div class="p-2 rounded-lg bg-success/20 group-hover:animate-pulse">
            <TrendingUp class="h-5 w-5 text-success" />
          </div>
        </div>
        <div class="text-3xl font-bold text-success mb-1">{{ stats.targetsRaised }}</div>
        <p class="text-xs text-muted-foreground flex items-center gap-1">
          <Target class="h-3 w-3" />
          Bullish market signals
        </p>
        <div class="mt-2 h-1 bg-muted rounded-full overflow-hidden">
          <div 
            class="h-full bg-success transition-all duration-1000 ease-out"
            :style="{ width: `${(stats.targetsRaised / stats.totalStocks) * 100}%` }"
          />
        </div>
      </div>

      <!-- Targets Lowered -->
      <div class="glass-card hover:shadow-lg transition-all duration-300 hover:scale-105 animate-fade-in group p-6 rounded-lg border border-border/50" style="animation-delay: 0.1s">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-sm font-medium text-muted-foreground">Targets Lowered</h3>
            <p class="text-xs mt-1 text-muted-foreground">Analysts reduced price expectations</p>
          </div>
          <div class="p-2 rounded-lg bg-destructive/20">
            <TrendingDown class="h-5 w-5 text-destructive" />
          </div>
        </div>
        <div class="text-3xl font-bold text-destructive mb-1">{{ stats.targetsLowered }}</div>
        <p class="text-xs text-muted-foreground flex items-center gap-1">
          <Target class="h-3 w-3" />
          Bearish market signals
        </p>
        <div class="mt-2 h-1 bg-muted rounded-full overflow-hidden">
          <div 
            class="h-full bg-destructive transition-all duration-1000 ease-out"
            :style="{ width: `${(stats.targetsLowered / stats.totalStocks) * 100}%` }"
          />
        </div>
      </div>

      <!-- Buy Ratings -->
      <div class="glass-card hover:shadow-lg transition-all duration-300 hover:scale-105 animate-fade-in group p-6 rounded-lg border border-border/50" style="animation-delay: 0.2s">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-sm font-medium text-muted-foreground">Buy Ratings</h3>
            <p class="text-xs mt-1 text-muted-foreground">Stocks with Buy/Outperform ratings</p>
          </div>
          <div class="p-2 rounded-lg bg-primary/20 group-hover:animate-glow">
            <Star class="h-5 w-5 text-primary" />
          </div>
        </div>
        <div class="text-3xl font-bold text-primary mb-1">{{ stats.buyRatings }}</div>
        <p class="text-xs text-muted-foreground flex items-center gap-1">
          <BarChart3 class="h-3 w-3" />
          Investment opportunities
        </p>
        <div class="mt-2 h-1 bg-muted rounded-full overflow-hidden">
          <div 
            class="h-full bg-primary transition-all duration-1000 ease-out animate-glow"
            :style="{ width: `${(stats.buyRatings / stats.totalStocks) * 100}%` }"
          />
        </div>
      </div>

      <!-- Unique Tickers -->
      <div class="glass-card hover:shadow-lg transition-all duration-300 hover:scale-105 animate-fade-in group p-6 rounded-lg border border-border/50" style="animation-delay: 0.3s">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-sm font-medium text-muted-foreground">Unique Tickers</h3>
            <p class="text-xs mt-1 text-muted-foreground">Different companies being analyzed</p>
          </div>
          <div class="p-2 rounded-lg bg-accent/20">
            <BarChart3 class="h-5 w-5 text-accent-foreground" />
          </div>
        </div>
        <div class="text-3xl font-bold mb-1">{{ stats.uniqueTickers }}</div>
        <p class="text-xs text-muted-foreground flex items-center gap-1">
          <DollarSign class="h-3 w-3" />
          Market coverage
        </p>
        <div class="mt-2 h-1 bg-muted rounded-full overflow-hidden">
          <div 
            class="h-full bg-accent-foreground transition-all duration-1000 ease-out"
            :style="{ width: `${Math.min((stats.uniqueTickers / 100) * 100, 100)}%` }"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TrendingUp, TrendingDown, BarChart3, Star, Target, DollarSign } from 'lucide-vue-next'
import { useStockStore } from '@/stores/stock'

const stockStore = useStockStore()

const stats = computed(() => {
  const stocks = stockStore.stocks
  const totalStocks = stocks.length

  const targetsRaised = stocks.filter(s => {
    if (!s.target_to || !s.target_from) return false
    const targetTo = parseFloat(s.target_to.replace('$', ''))
    const targetFrom = parseFloat(s.target_from.replace('$', ''))
    return !isNaN(targetTo) && !isNaN(targetFrom) && targetTo > targetFrom
  }).length

  const targetsLowered = stocks.filter(s => {
    if (!s.target_to || !s.target_from) return false
    const targetTo = parseFloat(s.target_to.replace('$', ''))
    const targetFrom = parseFloat(s.target_from.replace('$', ''))
    return !isNaN(targetTo) && !isNaN(targetFrom) && targetTo < targetFrom
  }).length

  const buyRatings = stocks.filter(s => 
    s.rating_to && (
      s.rating_to.toLowerCase().includes('buy') || 
      s.rating_to.toLowerCase().includes('outperform')
    )
  ).length

  const uniqueTickers = new Set(stocks.map(s => s.ticker)).size

  return {
    totalStocks,
    targetsRaised,
    targetsLowered,
    buyRatings,
    uniqueTickers
  }
})
</script>