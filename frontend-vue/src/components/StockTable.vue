<template>
  <div class="glass-card border border-border/50 rounded-lg overflow-hidden">
    <div v-if="stockStore.loading" class="p-8 text-center">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
      <p class="mt-2 text-muted-foreground">Loading stock data...</p>
    </div>
    
    <div v-else-if="stockStore.error" class="p-8 text-center text-destructive">
      <p>{{ stockStore.error }}</p>
    </div>
    
    <div v-else-if="stockStore.stocks.length === 0" class="p-8 text-center text-muted-foreground">
      <p>No stocks found matching your criteria.</p>
    </div>
    
    <div v-else class="overflow-x-auto">
      <table class="w-full">
        <thead class="bg-muted/50 border-b border-border/50">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Ticker</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Company</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Action</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Brokerage</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Rating</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Target Price</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Time</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border/50">
          <tr 
            v-for="stock in stockStore.stocks" 
            :key="`${stock.ticker}-${stock.time}`"
            class="hover:bg-muted/30 transition-colors duration-200"
          >
            <td class="px-4 py-3">
              <span class="font-mono font-semibold text-primary">{{ stock.ticker }}</span>
            </td>
            <td class="px-4 py-3">
              <span class="font-medium">{{ stock.company }}</span>
            </td>
            <td class="px-4 py-3">
              <span 
                class="px-2 py-1 rounded-full text-xs font-medium"
                :class="getActionColor(stock.action)"
              >
                {{ stock.action }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-muted-foreground">
              {{ stock.brokerage }}
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2 text-sm">
                <span v-if="stock.rating_from" class="text-muted-foreground">{{ stock.rating_from }}</span>
                <ArrowRight v-if="stock.rating_from && stock.rating_to" class="h-3 w-3 text-muted-foreground" />
                <span v-if="stock.rating_to" class="font-medium" :class="getRatingColor(stock.rating_to)">
                  {{ stock.rating_to }}
                </span>
              </div>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2 text-sm">
                <span v-if="stock.target_from" class="text-muted-foreground font-mono">{{ stock.target_from }}</span>
                <ArrowRight v-if="stock.target_from && stock.target_to" class="h-3 w-3 text-muted-foreground" />
                <span v-if="stock.target_to" class="font-mono font-medium" :class="getTargetColor(stock.target_from, stock.target_to)">
                  {{ stock.target_to }}
                </span>
              </div>
            </td>
            <td class="px-4 py-3 text-sm text-muted-foreground">
              {{ formatTime(stock.time) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'
import { useStockStore } from '@/stores/stock'

const stockStore = useStockStore()

const getActionColor = (action: string) => {
  const lowerAction = action.toLowerCase()
  if (lowerAction.includes('raised') || lowerAction.includes('upgrade') || lowerAction.includes('initiated')) {
    return 'bg-success/20 text-success'
  } else if (lowerAction.includes('lowered') || lowerAction.includes('downgrade')) {
    return 'bg-destructive/20 text-destructive'
  } else {
    return 'bg-muted/50 text-muted-foreground'
  }
}

const getRatingColor = (rating: string) => {
  const lowerRating = rating.toLowerCase()
  if (lowerRating.includes('buy') || lowerRating.includes('outperform')) {
    return 'text-success'
  } else if (lowerRating.includes('sell') || lowerRating.includes('underperform')) {
    return 'text-destructive'
  } else {
    return 'text-foreground'
  }
}

const getTargetColor = (targetFrom: string, targetTo: string) => {
  if (!targetFrom || !targetTo) return 'text-foreground'
  
  const from = parseFloat(targetFrom.replace('$', '').replace(',', ''))
  const to = parseFloat(targetTo.replace('$', '').replace(',', ''))
  
  if (isNaN(from) || isNaN(to)) return 'text-foreground'
  
  if (to > from) return 'text-success'
  if (to < from) return 'text-destructive'
  return 'text-foreground'
}

const formatTime = (timeStr: string) => {
  try {
    const date = new Date(timeStr)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return timeStr
  }
}
</script>