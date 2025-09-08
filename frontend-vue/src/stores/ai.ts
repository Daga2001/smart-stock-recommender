import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChatMessage, ConversationMemory, SummaryResponse, StockRecommendation } from '@/types'
import { apiService } from '@/services/api'

export const useAIStore = defineStore('ai', () => {
  // State
  const summary = ref<SummaryResponse | null>(null)
  const loadingSummary = ref(false)
  const chatMessages = ref<ChatMessage[]>([])
  const sendingMessage = ref(false)
  const conversationMemory = ref<ConversationMemory>({
    summary: '',
    keyTopics: [],
    lastContext: ''
  })
  const recommendations = ref<StockRecommendation[]>([])
  const loadingRecommendations = ref(false)
  const recommendationLimit = ref(10)

  // Actions
  const loadSummary = async () => {
    loadingSummary.value = true
    try {
      const data = await apiService.getSummary()
      summary.value = data
    } catch (error) {
      console.error('Failed to load AI summary:', error)
    } finally {
      loadingSummary.value = false
    }
  }

  const loadRecommendations = async (limit: number = 10) => {
    loadingRecommendations.value = true
    recommendationLimit.value = limit
    try {
      const data = await apiService.getRecommendations(limit)
      recommendations.value = data.recommendations
    } catch (error) {
      console.error('Failed to load recommendations:', error)
    } finally {
      loadingRecommendations.value = false
    }
  }

  const sendMessage = async (message: string) => {
    if (!message.trim() || sendingMessage.value) return

    const userMessage: ChatMessage = {
      role: 'user',
      content: message,
      timestamp: new Date()
    }

    chatMessages.value.push(userMessage)
    sendingMessage.value = true

    try {
      const response = await apiService.sendChatMessage(
        message,
        conversationMemory.value,
        chatMessages.value || []
      )

      const assistantMessage: ChatMessage = {
        role: 'assistant',
        content: response.response || 'No response received',
        timestamp: new Date(),
        context: response.context_used
      }

      chatMessages.value.push(assistantMessage)

      // Update conversation memory
      if (response.updated_memory) {
        conversationMemory.value = response.updated_memory
      }
    } catch (error) {
      console.error('Failed to send chat message:', error)
      const errorMessage: ChatMessage = {
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date()
      }
      chatMessages.value.push(errorMessage)
    } finally {
      sendingMessage.value = false
    }
  }

  const clearChat = () => {
    chatMessages.value = []
    conversationMemory.value = {
      summary: '',
      keyTopics: [],
      lastContext: ''
    }
  }

  return {
    // State
    summary,
    loadingSummary,
    chatMessages,
    sendingMessage,
    conversationMemory,
    recommendations,
    loadingRecommendations,
    recommendationLimit,

    // Actions
    loadSummary,
    loadRecommendations,
    sendMessage,
    clearChat
  }
})