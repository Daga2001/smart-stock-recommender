# Smart Stock Recommender - Vue 3 Frontend

A modern Vue 3 + TypeScript + Pinia + Tailwind CSS frontend for the Smart Stock Recommender application.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## 🛠️ Tech Stack

- **Vue 3** - Progressive JavaScript framework with Composition API
- **TypeScript** - Type safety and better developer experience
- **Pinia** - State management for Vue
- **Tailwind CSS** - Utility-first CSS framework
- **Vite** - Fast build tool and development server
- **Lucide Vue** - Beautiful icons

## 📁 Project Structure

```
src/
├── components/          # Vue components
│   ├── AIAssistant.vue     # AI chat and market summary
│   ├── PaginationControls.vue
│   ├── StatisticsCards.vue
│   ├── StockFilters.vue    # Advanced filtering
│   ├── StockRecommendations.vue
│   └── StockTable.vue
├── stores/              # Pinia stores
│   ├── ai.ts           # AI functionality
│   └── stock.ts        # Stock data management
├── services/           # API services
│   └── api.ts          # Backend communication
├── types/              # TypeScript interfaces
│   └── index.ts
├── views/              # Page components
│   └── Dashboard.vue
├── router/             # Vue Router
│   └── index.ts
├── App.vue             # Root component
├── main.ts             # Application entry point
└── style.css           # Global styles
```

## 🎯 Features

### 📊 Dashboard
- Market analytics overview with statistics cards
- Real-time data visualization
- Professional glass-card design with animations

### 🔍 Advanced Filtering
- Text search across all fields
- Action, rating, and price range filters
- LocalStorage persistence
- Server-side filtering across entire dataset

### 🤖 AI Assistant
- Market summary with GPT-4 analysis
- Interactive chat interface
- Conversation memory system
- Markdown response rendering

### 📈 Stock Recommendations
- Top N recommendations (3, 5, 10, 15, 20)
- Scoring system with visual indicators
- Detailed analysis and reasoning

### 📋 Data Table
- Sortable columns
- Color-coded data (ratings, targets, actions)
- Responsive design
- Loading and error states

### 🔄 Pagination
- Page navigation controls
- Configurable page sizes
- Real-time record counts

## 🎨 Design System

### Colors
- **Primary**: Blue theme for main actions
- **Success**: Green for positive indicators
- **Destructive**: Red for negative indicators
- **Muted**: Gray for secondary information

### Components
- **Glass Cards**: Backdrop blur effects
- **Animations**: Fade-in, slide-up, glow effects
- **Responsive**: Mobile-first design
- **Icons**: Lucide Vue icon library

## 🔧 Configuration

### Environment Variables
Currently no environment variables are required for the frontend.

### API Configuration
The API base URL is configured in `src/services/api.ts`:
```typescript
const BASE_URL = 'http://localhost:8081/api'
```

## 🧪 Development

### Code Style
- TypeScript strict mode enabled
- ESLint for code linting
- Composition API preferred over Options API
- Single File Components (SFC) structure

### State Management
- Pinia stores for reactive state
- Composable pattern for reusable logic
- LocalStorage integration for persistence

### Performance
- Lazy loading for components
- Debounced search inputs
- Optimistic UI updates
- Efficient re-rendering with computed properties

## 🚀 Deployment

```bash
# Build for production
npm run build

# The dist/ folder contains the built application
# Deploy to any static hosting service
```

## 🔗 Backend Integration

This frontend connects to the Go backend API running on `localhost:8081`. Ensure the backend server is running before starting the frontend development server.

Key API endpoints:
- `POST /api/stocks/list` - Get paginated stock data
- `POST /api/stocks/search` - Advanced search with filters
- `GET /api/stocks/recommendations` - Get AI recommendations
- `GET /api/stocks/summary` - Get AI market summary
- `POST /api/stocks/chat` - AI chat interface