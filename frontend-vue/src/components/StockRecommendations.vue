<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-semibold flex items-center gap-3">
          <div class="h-8 w-1 bg-primary rounded-full animate-pulse" />
          Top Stock Recommendations
        </h2>
        <p class="text-muted-foreground mt-1">
          AI-powered analysis of the best investment opportunities
        </p>
      </div>
      <div class="flex items-center gap-3">
        <select
          v-model="aiStore.recommendationLimit"
          @change="loadRecommendations"
          class="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-lg cursor-pointer"
        >
          <option :value="3">Top 3</option>
          <option :value="5">Top 5</option>
          <option :value="10">Top 10</option>
          <option :value="15">Top 15</option>
          <option :value="20">Top 20</option>
        </select>
        <button
          @click="loadRecommendations"
          :disabled="aiStore.loadingRecommendations"
          class="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium bg-primary/10 hover:bg-primary/20 text-primary focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-lg disabled:opacity-50 flex items-center gap-2"
        >
          <RefreshCw :class="{ 'animate-spin': aiStore.loadingRecommendations }" class="h-4 w-4" />
          Refresh
        </button>
      </div>
    </div>

    <div v-if="aiStore.loadingRecommendations" class="text-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
      <p class="mt-2 text-muted-foreground">Analyzing market data...</p>
    </div>

    <div v-else-if="aiStore.recommendations.length === 0" class="text-center py-8 text-muted-foreground">
      <p>No recommendations available at this time.</p>
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="(rec, index) in aiStore.recommendations"
        :key="rec.ticker"
        class="glass-card border border-border/50 p-6 rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 animate-fade-in group"
        :style="{ animationDelay: `${index * 0.1}s` }"
      >
        <div class="flex items-start justify-between mb-4">
          <div>
            <h3 class="font-mono font-bold text-lg text-primary">{{ rec.ticker }}</h3>
            <p class="text-sm text-muted-foreground line-clamp-2">{{ rec.company }}</p>
          </div>
          <div class="text-right">
            <div class="text-2xl font-bold" :class="getScoreColor(rec.score)">
              {{ rec.score.toFixed(1) }}
            </div>
            <div class="text-xs text-muted-foreground">Score</div>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">Recommendation:</span>
            <span 
              class="px-2 py-1 rounded-full text-xs font-medium"
              :class="getRecommendationColor(rec.recommendation)"
            >
              {{ rec.recommendation }}
            </span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">Current Rating:</span>
            <span class="text-sm font-medium" :class="getRatingColor(rec.current_rating)">
              {{ rec.current_rating }}
            </span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">Target Price:</span>
            <span class="text-sm font-mono font-bold text-success">{{ rec.target_price }}</span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">Brokerage:</span>
            <span class="text-sm font-medium">{{ rec.brokerage }}</span>
          </div>

          <div class="pt-2 border-t border-border/50">
            <p class="text-xs text-muted-foreground">{{ rec.reason }}</p>
          </div>

          <!-- Score Progress Bar -->
          <div class="mt-3">
            <div class="flex items-center justify-between text-xs text-muted-foreground mb-1">
              <span>Investment Score</span>
              <span>{{ rec.score.toFixed(1) }}/10</span>
            </div>
            <div class="h-2 bg-muted rounded-full overflow-hidden">
              <div 
                class="h-full transition-all duration-1000 ease-out"
                :class="getScoreBarColor(rec.score)"
                :style="{ width: `${(rec.score / 10) * 100}%` }"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { useAIStore } from '@/stores/ai'

const aiStore = useAIStore()

const loadRecommendations = () => {
  aiStore.loadRecommendations(aiStore.recommendationLimit)
}

const getScoreColor = (score: number) => {
  if (score >= 8.5) return 'text-success'
  if (score >= 7.0) return 'text-primary'
  if (score >= 6.0) return 'text-yellow-500'
  return 'text-muted-foreground'
}

const getScoreBarColor = (score: number) => {
  if (score >= 8.5) return 'bg-success'
  if (score >= 7.0) return 'bg-primary'
  if (score >= 6.0) return 'bg-yellow-500'
  return 'bg-muted-foreground'
}

const getRecommendationColor = (recommendation: string) => {
  const lower = recommendation.toLowerCase()
  if (lower.includes('strong buy')) return 'bg-success/20 text-success'
  if (lower.includes('buy')) return 'bg-primary/20 text-primary'
  if (lower.includes('moderate')) return 'bg-yellow-500/20 text-yellow-500'
  return 'bg-muted/50 text-muted-foreground'
}

const getRatingColor = (rating: string) => {
  const lower = rating.toLowerCase()
  if (lower.includes('buy') || lower.includes('outperform')) return 'text-success'
  if (lower.includes('sell') || lower.includes('underperform')) return 'text-destructive'
  return 'text-foreground'
}
</script>