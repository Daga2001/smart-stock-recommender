/**
 * Frontend Test Setup Configuration
 * 
 * PURPOSE:
 * Configures the testing environment for React components using Vitest and React Testing Library.
 * Provides necessary mocks and polyfills for browser APIs that aren't available in Node.js.
 * 
 * WHAT THIS PROVIDES:
 * - Browser API mocks (matchMedia, ResizeObserver, localStorage)
 * - Jest-DOM matchers for enhanced assertions
 * - Consistent testing environment across all component tests
 * 
 * WHY THIS IS NEEDED:
 * - React components often use browser APIs not available in test environment
 * - Mocking prevents errors and allows focus on component logic
 * - Provides consistent, predictable test environment
 */

import { vi } from 'vitest'
import '@testing-library/jest-dom'

/**
 * Mock window.matchMedia API
 * 
 * PURPOSE: Provides mock implementation of CSS media query matching
 * USAGE: Components that use responsive design need this mock
 */
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,              // Consistent false for predictable tests
    media: query,                // Returns the query string
    onchange: null,              // Event handler placeholder
    addListener: vi.fn(),        // Legacy event listener
    removeListener: vi.fn(),     // Legacy event listener
    addEventListener: vi.fn(),    // Modern event listener
    removeEventListener: vi.fn(), // Modern event listener
    dispatchEvent: vi.fn(),      // Event dispatching
  })),
})

/**
 * Mock ResizeObserver API
 * 
 * PURPOSE: Provides mock implementation of element resize observation
 * USAGE: Components that respond to element size changes
 */
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),      // Mock element observation
  unobserve: vi.fn(),    // Mock element unobservation
  disconnect: vi.fn(),   // Mock observer disconnection
}))

/**
 * Mock localStorage API
 * 
 * PURPOSE: Provides mock implementation of browser local storage
 * USAGE: Components that persist data locally
 */
const localStorageMock = {
  getItem: vi.fn(),      // Mock data retrieval
  setItem: vi.fn(),      // Mock data storage
  removeItem: vi.fn(),   // Mock data removal
  clear: vi.fn(),        // Mock storage clearing
}
global.localStorage = localStorageMock as any