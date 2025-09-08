<template>
  <div class="glass-card p-6 space-y-4 animate-fade-in border border-border/50 rounded-lg">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <div class="p-2 rounded-lg bg-primary/20 animate-glow">
          <Sparkles class="h-5 w-5 text-primary" />
        </div>
        <div>
          <h3 class="font-semibold text-lg">Filters</h3>
          <p class="text-sm text-muted-foreground">Refine your search with precision</p>
        </div>
      </div>
      
      <button
        v-if="stockStore.hasActiveFilters"
        @click="clearFilters"
        class="px-4 py-2 text-sm border border-destructive/50 rounded-lg hover:bg-destructive/10 hover:text-destructive transition-all duration-200 flex items-center gap-2"
      >
        <RefreshCw class="h-4 w-4" />
        Clear All
      </button>
    </div>

    <form @submit.prevent="applyFilters">
      <!-- Search Bar -->
      <div class="mb-4">
        <div class="relative">
          <Search class="absolute left-4 top-1/2 transform -translate-y-1/2 text-muted-foreground h-5 w-5 z-10" />
          <input
            v-model="stockStore.filters.search"
            placeholder="Search stocks, companies, or brokerages..."
            class="w-full pl-12 h-12 bg-background/50 border border-border/50 rounded-lg focus:border-primary/50 focus:bg-background transition-all duration-200 text-base px-4"
            @keydown.enter="applyFilters"
          />
          <div v-if="stockStore.filters.search" class="absolute right-3 top-1/2 transform -translate-y-1/2">
            <div class="text-xs text-primary font-medium bg-primary/20 px-2 py-1 rounded-full">
              Active
            </div>
          </div>
        </div>
      </div>

      <!-- Filter Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
        <!-- Action Filter -->
        <div class="space-y-2">
          <label class="text-sm font-medium flex items-center gap-2">
            <Filter class="h-4 w-4" />
            Action
          </label>
          <select
            v-model="stockStore.filters.action"
            class="w-full h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
          >
            <option value="all">All Actions</option>
            <option v-for="action in stockStore.filterOptions.actions" :key="action" :value="action">
              {{ action.charAt(0).toUpperCase() + action.slice(1) }}
            </option>
          </select>
        </div>

        <!-- Rating From Filter -->
        <div class="space-y-2">
          <label class="text-sm font-medium flex items-center gap-2">
            <Star class="h-4 w-4" />
            Rating From
          </label>
          <select
            v-model="stockStore.filters.rating_from"
            class="w-full h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
          >
            <option value="all">All Ratings</option>
            <option v-for="rating in stockStore.filterOptions.ratings_from" :key="rating" :value="rating">
              {{ rating }}
            </option>
          </select>
        </div>

        <!-- Rating To Filter -->
        <div class="space-y-2">
          <label class="text-sm font-medium flex items-center gap-2">
            <Star class="h-4 w-4" />
            Rating To
          </label>
          <select
            v-model="stockStore.filters.rating_to"
            class="w-full h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
          >
            <option value="all">All Ratings</option>
            <option v-for="rating in stockStore.filterOptions.ratings_to" :key="rating" :value="rating">
              {{ rating }}
            </option>
          </select>
        </div>

        <!-- Apply Button -->
        <div class="space-y-2">
          <label class="text-sm font-medium opacity-0">Action</label>
          <button
            type="submit"
            :disabled="stockStore.loading"
            class="w-full h-10 bg-primary hover:bg-primary/90 text-primary-foreground font-medium rounded-lg transition-all duration-200 hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <RefreshCw v-if="stockStore.loading" class="h-4 w-4 animate-spin" />
            <Search v-else class="h-4 w-4" />
            {{ stockStore.loading ? 'Searching...' : 'Apply Filters' }}
          </button>
        </div>
      </div>

      <!-- Target Price Ranges -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Target From Range -->
        <div class="space-y-2">
          <label class="text-sm font-medium flex items-center gap-2">
            <DollarSign class="h-4 w-4" />
            Target From Range
          </label>
          <div class="flex gap-2">
            <input
              v-model.number="stockStore.filters.target_from_min"
              type="number"
              placeholder="Min"
              class="flex-1 h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
            />
            <input
              v-model.number="stockStore.filters.target_from_max"
              type="number"
              placeholder="Max"
              class="flex-1 h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
            />
          </div>
        </div>

        <!-- Target To Range -->
        <div class="space-y-2">
          <label class="text-sm font-medium flex items-center gap-2">
            <DollarSign class="h-4 w-4" />
            Target To Range
          </label>
          <div class="flex gap-2">
            <input
              v-model.number="stockStore.filters.target_to_min"
              type="number"
              placeholder="Min"
              class="flex-1 h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
            />
            <input
              v-model.number="stockStore.filters.target_to_max"
              type="number"
              placeholder="Max"
              class="flex-1 h-10 bg-background/50 border border-border/50 rounded-lg px-3 focus:border-primary/50 transition-all duration-200"
            />
          </div>
        </div>
      </div>
    </form>

    <!-- Filter Status -->
    <div v-if="stockStore.hasActiveFilters" class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground border-t border-border/30 pt-4">
      <span>Active filters:</span>
      <span v-if="stockStore.filters.search" class="bg-primary/20 text-primary px-2 py-1 rounded-full text-xs font-medium">
        Search: "{{ stockStore.filters.search }}"
      </span>
      <span v-if="stockStore.filters.action !== 'all'" class="bg-success/20 text-success px-2 py-1 rounded-full text-xs font-medium">
        Action: {{ stockStore.filters.action }}
      </span>
      <span v-if="stockStore.filters.rating_from !== 'all'" class="bg-blue-500/20 text-blue-500 px-2 py-1 rounded-full text-xs font-medium">
        From: {{ stockStore.filters.rating_from }}
      </span>
      <span v-if="stockStore.filters.rating_to !== 'all'" class="bg-purple-500/20 text-purple-500 px-2 py-1 rounded-full text-xs font-medium">
        To: {{ stockStore.filters.rating_to }}
      </span>
      <span v-if="stockStore.filters.target_from_min > 0 || stockStore.filters.target_from_max > 0" class="bg-orange-500/20 text-orange-500 px-2 py-1 rounded-full text-xs font-medium">
        Target From: ${{ stockStore.filters.target_from_min || 0 }} - ${{ stockStore.filters.target_from_max || '∞' }}
      </span>
      <span v-if="stockStore.filters.target_to_min > 0 || stockStore.filters.target_to_max > 0" class="bg-green-500/20 text-green-500 px-2 py-1 rounded-full text-xs font-medium">
        Target To: ${{ stockStore.filters.target_to_min || 0 }} - ${{ stockStore.filters.target_to_max || '∞' }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search, Filter, Sparkles, RefreshCw, DollarSign, Star } from 'lucide-vue-next'
import { useStockStore } from '@/stores/stock'

const stockStore = useStockStore()

const applyFilters = () => {
  stockStore.applyFilters()
}

const clearFilters = () => {
  stockStore.clearFilters()
}
</script>