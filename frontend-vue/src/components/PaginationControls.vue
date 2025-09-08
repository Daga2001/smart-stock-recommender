<template>
  <div class="glass-card border border-border/50 p-6 animate-fade-in rounded-lg">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-6">
        <div class="flex items-center gap-3">
          <label class="text-sm font-semibold text-foreground">Page:</label>
          <input
            v-model.number="currentPageInput"
            type="number"
            :min="1"
            :max="stockStore.pagination?.total_pages || 1"
            @keydown.enter="goToPage"
            class="w-20 px-3 py-2 glass-card border border-border/50 rounded-lg text-sm font-mono font-semibold text-center focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-lg"
          />
          <span class="text-sm text-muted-foreground font-medium">
            of {{ stockStore.pagination?.total_pages || 1 }}
          </span>
        </div>
        
        <div class="flex items-center gap-3">
          <label class="text-sm font-semibold text-foreground">Show:</label>
          <select
            v-model="stockStore.pageLength"
            @change="stockStore.setPageLength(stockStore.pageLength)"
            class="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-lg cursor-pointer"
          >
            <option :value="10">10 per page</option>
            <option :value="20">20 per page</option>
            <option :value="50">50 per page</option>
            <option :value="100">100 per page</option>
          </select>
        </div>
        
        <div class="flex items-center gap-3">
          <button
            @click="refreshData"
            :disabled="stockStore.loading"
            class="px-4 py-2 glass-card border border-border/50 rounded-lg text-sm font-medium bg-primary/10 hover:bg-primary/20 text-primary focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all duration-200 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <RefreshCw :class="{ 'animate-spin': stockStore.loading }" class="h-4 w-4" />
            {{ stockStore.loading ? 'Loading...' : 'Refresh' }}
          </button>
          <div class="text-xs text-muted-foreground">
            💡 Click refresh after changing page number to navigate through the table
          </div>
        </div>
      </div>
      
      <div class="text-sm text-muted-foreground font-medium">
        <span class="text-primary font-semibold">
          {{ stockStore.pagination?.total_records?.toLocaleString() || 0 }}
        </span> 
        total records
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useStockStore } from '@/stores/stock'

const stockStore = useStockStore()
const currentPageInput = ref(stockStore.currentPage)

// Watch for changes in store's current page
watch(() => stockStore.currentPage, (newPage) => {
  currentPageInput.value = newPage
})

const goToPage = () => {
  const page = Math.max(1, Math.min(stockStore.pagination?.total_pages || 1, currentPageInput.value))
  stockStore.goToPage(page)
}

const refreshData = () => {
  if (currentPageInput.value !== stockStore.currentPage) {
    goToPage()
  } else {
    stockStore.loadStocks(stockStore.currentPage)
  }
}
</script>