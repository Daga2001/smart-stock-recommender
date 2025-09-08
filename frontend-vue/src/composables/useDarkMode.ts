import { ref, watch } from 'vue'

const isDark = ref(false)

// Initialize dark mode from localStorage or default to dark
const initializeDarkMode = () => {
  const stored = localStorage.getItem('darkMode')
  if (stored !== null) {
    isDark.value = JSON.parse(stored)
  } else {
    isDark.value = true // Default to dark mode
  }
  updateDarkMode()
}

// Update DOM and localStorage
const updateDarkMode = () => {
  if (isDark.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  localStorage.setItem('darkMode', JSON.stringify(isDark.value))
}

// Watch for changes
watch(isDark, updateDarkMode)

// Toggle function
const toggleDarkMode = () => {
  isDark.value = !isDark.value
}

export const useDarkMode = () => {
  return {
    isDark,
    toggleDarkMode,
    initializeDarkMode
  }
}