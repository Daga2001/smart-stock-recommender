<template>
  <div class="space-y-6">
    <!-- AI Market Summary -->
    <div class="glass-card border border-border/50 animate-fade-in rounded-lg">
      <div class="p-6 border-b border-border/50">
        <div class="flex items-center gap-3 mb-2">
          <div class="p-2 rounded-lg bg-primary/20 animate-glow">
            <TrendingUp class="h-5 w-5 text-primary" />
          </div>
          <h3 class="text-lg font-semibold">AI Market Summary</h3>
        </div>
        <p class="text-sm text-muted-foreground">
          AI analysis of the 50 most recent analyst ratings.
        </p>
      </div>
      
      <div class="p-6">
        <div v-if="aiStore.loadingSummary" class="flex items-center gap-2 text-muted-foreground">
          <Bot class="h-4 w-4 animate-spin" />
          Generating AI analysis...
        </div>
        
        <div v-else-if="aiStore.summary" class="space-y-4">
          <p class="text-sm leading-relaxed">{{ aiStore.summary.summary }}</p>
          <div class="flex items-center justify-between text-xs text-muted-foreground">
            <span>Generated: {{ formatDate(aiStore.summary.generated_at) }}</span>
            <span>Tokens used: {{ aiStore.summary.tokens_used }}</span>
          </div>
          <button 
            @click="aiStore.loadSummary()" 
            :disabled="aiStore.loadingSummary"
            class="px-4 py-2 border border-border/50 rounded-lg text-sm font-medium hover:bg-muted/50 transition-all duration-200 flex items-center gap-2"
          >
            <Sparkles class="h-4 w-4" />
            Refresh Analysis
          </button>
        </div>
        
        <p v-else class="text-muted-foreground">Failed to load AI summary</p>
      </div>
    </div>

    <!-- AI Chat Assistant -->
    <div class="glass-card border border-border/50 animate-fade-in rounded-lg">
      <div class="p-6 border-b border-border/50">
        <div class="flex items-center gap-3 mb-2">
          <div class="p-2 rounded-lg bg-success/20">
            <Bot class="h-5 w-5 text-success" />
          </div>
          <h3 class="text-lg font-semibold">Need Help?</h3>
        </div>
        <p class="text-sm text-muted-foreground">
          Ask our AI agent for personalized stock analysis and investment insights
        </p>
      </div>
      
      <div class="p-6">
        <button
          @click="showChatModal = true"
          class="w-full px-4 py-3 bg-primary hover:bg-primary/90 text-primary-foreground font-medium rounded-lg transition-all duration-200 hover:shadow-lg flex items-center justify-center gap-2"
        >
          <MessageCircle class="h-4 w-4" />
          Chat with AI Assistant
        </button>
      </div>
    </div>

    <!-- Chat Modal -->
    <div v-if="showChatModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="glass-card border border-border/50 rounded-lg w-full max-w-2xl max-h-[80vh] flex flex-col">
        <!-- Modal Header -->
        <div class="p-6 border-b border-border/50 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Bot class="h-5 w-5 text-primary" />
            <h3 class="text-lg font-semibold">AI Stock Assistant</h3>
          </div>
          <button
            @click="showChatModal = false"
            class="p-2 hover:bg-muted/50 rounded-lg transition-colors"
          >
            <X class="h-4 w-4" />
          </button>
        </div>
        
        <div class="p-4 text-sm text-muted-foreground border-b border-border/50">
          Ask questions about stocks, market trends, or get investment advice
        </div>
        
        <!-- Chat Messages -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4 min-h-96">
          <div v-if="aiStore.chatMessages.length === 0" class="text-center space-y-6">
            <div class="flex items-center justify-center">
              <div class="p-4 rounded-xl bg-primary/20 animate-glow">
                <Bot class="h-12 w-12 text-primary" />
              </div>
            </div>
            <div class="space-y-4">
              <div>
                <h4 class="text-lg font-semibold mb-2">AI Stock Assistant</h4>
                <p class="text-muted-foreground">Ask me anything about stocks, market trends, or get investment advice</p>
              </div>
              
              <div class="glass-card border border-border/50 p-4 text-left space-y-3 rounded-lg">
                <div class="flex items-center gap-2 mb-3">
                  <Sparkles class="h-4 w-4 text-primary" />
                  <span class="font-semibold text-primary">Pro Tips for Better Results</span>
                </div>
                
                <div class="space-y-3 text-sm">
                  <div class="space-y-2">
                    <div class="flex items-start gap-2">
                      <span class="text-destructive font-medium">❌</span>
                      <span class="text-muted-foreground">"What stocks are good?"</span>
                    </div>
                    <div class="flex items-start gap-2">
                      <span class="text-success font-medium">✅</span>
                      <span class="font-medium">"Which biotech stocks have recent buy ratings from Goldman Sachs?"</span>
                    </div>
                  </div>
                  
                  <div class="h-px bg-border/50"></div>
                  
                  <div class="space-y-2">
                    <div class="flex items-start gap-2">
                      <span class="text-destructive font-medium">❌</span>
                      <span class="text-muted-foreground">"Tell me about AAPL"</span>
                    </div>
                    <div class="flex items-start gap-2">
                      <span class="text-success font-medium">✅</span>
                      <span class="font-medium">"What are AAPL's recent target price changes and analyst ratings?"</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div v-else>
            <div
              v-for="(message, index) in aiStore.chatMessages"
              :key="index"
              class="flex"
              :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-[80%] p-3 rounded-lg"
                :class="message.role === 'user' 
                  ? 'bg-primary text-primary-foreground' 
                  : 'bg-background border glass-card'"
              >
                <div v-if="message.role === 'assistant'" class="text-sm prose prose-sm max-w-none" v-html="formatMarkdown(message.content)"></div>
                <p v-else class="text-sm">{{ message.content }}</p>
                <p class="text-xs opacity-70 mt-1">
                  {{ message.timestamp.toLocaleTimeString() }}
                </p>
              </div>
            </div>
          </div>
          
          <div v-if="aiStore.sendingMessage" class="flex justify-start">
            <div class="glass-card border border-border/50 p-3 rounded-lg">
              <div class="flex items-center gap-2 text-muted-foreground">
                <Bot class="h-4 w-4 animate-spin" />
                <span class="text-sm">Analyzing market data...</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Active Context -->
        <div v-if="aiStore.conversationMemory.keyTopics && aiStore.conversationMemory.keyTopics.length > 0" class="px-4 py-2 border-t border-border/50">
          <div class="glass-card border border-border/50 p-2 rounded-lg">
            <div class="flex items-center gap-2 text-xs">
              <div class="p-1 rounded bg-primary/20">
                <Activity class="h-3 w-3 text-primary" />
              </div>
              <span class="font-medium text-muted-foreground">Active Context:</span>
              <span class="text-primary font-medium">{{ (aiStore.conversationMemory.keyTopics || []).join(', ') }}</span>
            </div>
          </div>
        </div>

        <!-- Chat Input -->
        <div class="p-4 border-t border-border/50">
          <div class="flex gap-2">
            <input
              v-model="currentMessage"
              placeholder="Ask about stocks, trends, or get investment advice..."
              @keydown.enter="sendMessage"
              :disabled="aiStore.sendingMessage"
              class="flex-1 px-4 py-2 bg-background/50 border border-border/50 rounded-lg focus:border-primary/50 focus:bg-background transition-all duration-200"
            />
            <button 
              @click="sendMessage" 
              :disabled="!currentMessage.trim() || aiStore.sendingMessage"
              class="px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Send class="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { TrendingUp, Bot, MessageCircle, Sparkles, Send, X, Activity } from 'lucide-vue-next'
import { useAIStore } from '@/stores/ai'
import MarkdownIt from 'markdown-it'

const aiStore = useAIStore()
const showChatModal = ref(false)
const currentMessage = ref('')

const md = new MarkdownIt()

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

const formatMarkdown = (content: string) => {
  return md.render(content)
}

const sendMessage = async () => {
  if (!currentMessage.value.trim() || aiStore.sendingMessage) return
  
  const message = currentMessage.value
  currentMessage.value = ''
  
  await aiStore.sendMessage(message)
}
</script>