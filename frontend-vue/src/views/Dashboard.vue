<template>
  <div class="min-h-screen gradient-hero">
    <!-- Header -->
    <header class="sticky top-0 z-50 glass-card border-b border-border/50 backdrop-blur-xl">
      <div class="container mx-auto px-6 py-8">
        <div class="flex items-center justify-between w-full">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl bg-primary/20 animate-glow">
              <Activity class="h-8 w-8 text-primary" />
            </div>
            <div>
              <h1 class="text-4xl font-bold tracking-tight animate-fade-in">
                Stock Market <span class="gradient-primary bg-clip-text text-transparent">Intelligence</span>
              </h1>
              <p class="text-lg text-muted-foreground animate-slide-up">
                Advanced analytics • Valuable insights • AI assisted recommendations
              </p>
            </div>
          </div>
          

        </div>
      </div>
    </header>

    <div class="container mx-auto px-6 py-8">
      <!-- Statistics Cards -->
      <StatisticsCards />

      <!-- AI Assistant -->
      <div class="mb-12 animate-slide-up" style="animation-delay: 0.4s">
        <AIAssistant />
      </div>

      <!-- Recommendations -->
      <div class="mb-12 animate-slide-up" style="animation-delay: 0.5s">
        <StockRecommendations />
      </div>

      <!-- Filters -->
      <div class="mb-8 animate-slide-up" style="animation-delay: 0.6s">
        <StockFilters />
      </div>

      <!-- Stock Table -->
      <div class="space-y-6 animate-slide-up" style="animation-delay: 0.7s">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-2xl font-semibold flex items-center gap-3">
              <div class="h-8 w-1 bg-primary rounded-full animate-pulse" />
              Market Analysis
            </h2>
            <p class="text-muted-foreground mt-1">
              Professional Stock Analysis Table
            </p>
          </div>
          <div class="text-right">
            <div class="text-sm text-muted-foreground">
              Displaying <span class="font-semibold text-primary">{{ stockStore.stocks.length }}</span> positions
            </div>
            <div class="text-xs text-muted-foreground mt-1">
              Last updated: {{ new Date().toLocaleTimeString() }}
            </div>
          </div>
        </div>

        <!-- Pagination Controls -->
        <PaginationControls />

        <!-- Stock Table -->
        <StockTable />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { Activity } from 'lucide-vue-next'
import { useStockStore } from '@/stores/stock'
import { useAIStore } from '@/stores/ai'
import StatisticsCards from '@/components/StatisticsCards.vue'
import AIAssistant from '@/components/AIAssistant.vue'
import StockRecommendations from '@/components/StockRecommendations.vue'
import StockFilters from '@/components/StockFilters.vue'
import StockTable from '@/components/StockTable.vue'
import PaginationControls from '@/components/PaginationControls.vue'

const stockStore = useStockStore()
const aiStore = useAIStore()

onMounted(async () => {
  await stockStore.loadFilterOptions()
  await stockStore.loadStocks()
  await aiStore.loadSummary()
  await aiStore.loadRecommendations()
})
</script>