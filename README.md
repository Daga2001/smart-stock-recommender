# 📈 Smart Stock Recommender Project

A full-stack application that retrieves stock information from an external API, stores it in a scalable database, and presents insights through a beautiful and interactive UI.  
The system also provides intelligent stock recommendations to help users identify the best investment opportunities today.  

---

## ✨ Features

- 🔗 **External API Integration** – Securely connects to the stock data API with proper authentication and error handling.  
- 🗄️ **Reliable Data Storage** – Persists data in CockroachDB for scalability and fault tolerance.  
- 🚀 **Parallel Data Fetching** – Concurrent API calls with rate limiting and retry logic for maximum efficiency.  
- 📝 **Interactive API Documentation** – Auto-generated Swagger docs with try-it-out functionality.  
- 🎨 **Interactive UI** – Built with Vue 3, TypeScript, Pinia, and styled with Tailwind CSS.  
- 🔍 **Advanced Search & Filter** – RegEx-powered search across all fields in a stock information dataset (ticker, company, brokerage, action, ratings).  
- 📊 **Comprehensive Dashboard** – Market analytics overview with statistics cards and professional insights.  
- 🤖 **Statistical Recommendations** – Configurable weighted scoring system analyzing target price changes (40%), rating analysis (30%), action analysis (20%), and timing factors (10%).  
- 🎯 **Dynamic Top N Recommendations** – Flexible recommendation display (3, 5, 10, 15, 20 picks) with responsive grid layout.  
- 🧠 **AI Market Analysis** – GPT-4.1-nano powered market summaries with Wall Street analyst-level insights and interactive chat functionality.  
- ⚖️ **Weight Validation** – Ensures recommendation algorithm weights sum to 100% for accurate scoring.  
- 🔄 **Filtering** – Case-insensitive search with instant results and pagination persistence.  
- 🔒 **SQL Injection Protection** – Parameterized queries and input validation for security.  
- 🧪 **Unit Testing** – Ensures stability and reliability of backend and UI logic.  

---

## 🛠️ Tech Stack

**Frontend**
- [Vue 3](https://vuejs.org/) – progressive JavaScript framework  
- [TypeScript](https://www.typescriptlang.org/) – static typing for safer code  
- [Pinia](https://pinia.vuejs.org/) – state management for Vue  
- [Tailwind CSS](https://tailwindcss.com/) – utility-first styling  

**Backend**
- [Golang](https://go.dev/) – high-performance, statically typed backend  
- [Gin](https://gin-gonic.com/) – HTTP web framework  
- [Swagger](https://swagger.io/) – API documentation and testing  
- RESTful API design with parallel processing  

**Database**
- [CockroachDB](https://www.cockroachlabs.com/product/cockroachdb/) – distributed, resilient SQL database  

**Testing**
- Go testing framework for backend logic  
- React Testing Library / Vitest for frontend components  

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.25+](https://go.dev/dl/)  
- [Node.js 20+](https://nodejs.org/) & npm/yarn/pnpm  
- [CockroachDB](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html)  

### Installation

1. **📂 Clone the repository**
   ```bash
   git clone https://github.com/your-username/smart-stock-recommender.git
   cd smart-stock-recommender
   cd smart-stock-recommender
   ```

2. **🖥️ Setup backend (Go)**
  Navigate to backend repository and install dependencies:
   ```bash
   cd backend
   go mod tidy
   ```
   
   Create `backend/.env` file:
   ```env
   DB_HOST=your-cockroachdb-host
   DB_PORT=26257
   DB_USER=your-username
   DB_PASSWORD=your-database-password
   DB_NAME=stock-market-db
   DB_SSLMODE=require
   API_TOKEN=your-external-api-token
   PORT=8081
   ```
   
   Install hot reload tool (optional):
   ```bash
   go install github.com/air-verse/air@latest
   ```
   
   Run the server:
   ```bash
   # With hot reload (recommended for development)
   air
   
   # Or without hot reload
   go run main.go
   ```

3. **🌐 Setup frontend (Vue 3)**
  Navigate to frontend repository and install dependencies:
   ```bash
   cd frontend
   npm install
   ```
   
   Create `frontend/.env` file:
   ```env
   VITE_STOCK_API_TOKEN=your-external-api-token
   ```
   
   Run the server:
   ```bash
   npm run dev
   ```
   
4. **🗄️ Database (CockroachDB)**
   - Database automatically creates tables on first run
   - Configure connection parameters in `backend/.env`  

---

## 📡 API Endpoints (Backend)

**Base URL:** `http://localhost:8081`

### 📚 **Interactive API Documentation**
Visit **http://localhost:8081/swagger/index.html** for complete interactive API documentation with:
- All available endpoints
- Request/response examples
- Model schemas
- Try-it-out functionality

### **Key Endpoints:**

#### `POST /api/stocks`
Fetch stock data by page number from external API and store in database.
- **Body:** `{"page": 1}`
- **Features:** Single page fetch with retry logic

#### `POST /api/stocks/bulk` 🚀
Fetch stock data for multiple pages with **parallel processing**.
- **Body:** `{"start_page": 1, "end_page": 22}`
- **Features:** 
  - **Parallel API calls** (up to 20 concurrent requests)
  - **Automatic retry logic** for empty pages
  - **Batch database inserts** for optimal performance
  - **Rate limiting** to prevent API overload
  - **Database clearing** before bulk insert

#### `POST /api/stocks/list` 📋
Retrieve paginated stock ratings from database.
- **Body:** `{"page_number": 1, "page_length": 20}`
- **Features:** 
  - **Paginated results** with metadata
  - **Sorting** by creation date (newest first)
  - **Flexible page sizes** (1-1000 records)

#### `POST /api/stocks/search` 🔍
Search stock ratings using **regular expressions** across all dataset fields.
- **Body:** `{"page_number": 1, "page_length": 20, "search_term": "AAPL"}`
- **Features:** 
  - **RegEx-powered search** across ticker, company, brokerage, action, and ratings
  - **Case-insensitive matching** for flexible queries
  - **Paginated search results** with accurate totals
  - **Multi-field search** - one term searches all columns

#### `GET /api/stocks/metrics` 📊
Get comprehensive market analytics and insights.
- **Features:** 
  - **Parallel processing** for fast metrics calculation
  - **Target price analysis** (raised/lowered/maintained)
  - **Rating distribution** and sentiment analysis
  - **Top brokerages** by activity
  - **Market trends** and statistics

**Quick Test:**
```bash
# Search for stocks containing "zillow"
curl -X POST http://localhost:8081/api/stocks/search \
  -H "Content-Type: application/json" \
  -d '{"page_number": 1, "page_length": 20, "search_term": "zillow"}'
```

---

---

## 📝 Swagger Documentation

### **Accessing API Documentation**
Visit **http://localhost:8081/swagger/index.html** for interactive API documentation.

### **Adding Documentation for New Endpoints**

1. **Add Swagger annotations** to your handler functions:
```go
// @Summary Your endpoint description
// @Description Detailed description of what the endpoint does
// @Tags your-tag
// @Accept json
// @Produce json
// @Param request body YourModel true "Request description"
// @Success 200 {object} YourResponse
// @Failure 400 {object} map[string]string
// @Router /your-endpoint [post]
func (h *YourHandler) YourEndpoint(c *gin.Context) {
    // Your implementation
}
```

2. **Generate documentation**:
```bash
cd backend
swag init
```

3. **Restart server** - documentation updates automatically!

### **Swagger Dependencies**
- `github.com/swaggo/gin-swagger` - Gin integration
- `github.com/swaggo/files` - Static files
- `github.com/swaggo/swag` - Documentation generator

---

## 🧠 **GPT: Conversation Memory System**

### **💾 Memory Storage & Retrieval**

- Memory stored in frontend React component state
- No server-side session storage or database persistence
- Memory lost when user refreshes page or closes browser
- Lightweight and fast access for real-time conversations

**Memory Structure**:
```typescript
ConversationMemory {
  summary: string;        // Compressed conversation history (max 150 chars)
  keyTopics: string[];    // Extracted topics ["AAPL", "ratings", "target_prices"]
  lastContext: string;    // Cached database context for reuse
}
```

**Storage Flow**:
1. **User sends message** → Frontend sends to backend with current memory
2. **Backend processes** → Updates memory with new topics and context
3. **Response returned** → Frontend updates React state with new memory
4. **Next message** → Reuses cached context if topics match

**Cost Efficiency**:
- **Traditional**: Send full conversation history (1000+ tokens)
- **Memory System**: Send only summary + recent messages (200-300 tokens)
- **Savings**: 80-90% reduction in API costs for follow-up questions

**Example Memory Evolution**:
```
Initial: {summary: "", topics: [], context: ""}

After "AAPL ratings":
{summary: "User asked about AAPL ratings", topics: ["AAPL", "ratings"], context: "AAPL data..."}

Follow-up "What about target prices?":
→ Detects AAPL topic match → Reuses cached context → No new database query!
```

---

## 🤖 **AI Assistant Best Practices**

### **💡 Writing Effective Prompts**

The AI assistant works best with **specific, detailed questions**. Here's how to get accurate responses:

**❌ Avoid Vague Prompts:**
- "What stocks are good?"
- "Tell me about AAPL"
- "Any recommendations?"
- "What's happening in the market?"

**✅ Use Specific Prompts:**
- "Which biotech stocks have recent buy ratings from Goldman Sachs?"
- "What are AAPL's recent target price changes and analyst ratings?"
- "Show me stocks with target price increases above 15% this week"
- "Which companies had downgrades from JPMorgan recently?"

**🎯 Prompt Structure Tips:**
1. **Specify the stock/sector**: "AAPL", "biotech companies", "tech sector"
2. **Define the data type**: "target prices", "ratings", "analyst actions"
3. **Add time context**: "recent", "this week", "latest"
4. **Include criteria**: "above $100", "buy ratings", "from Goldman Sachs"

**🔍 Example Query Types:**
- **Stock-specific**: "MSFT target price changes by Morgan Stanley"
- **Sector analysis**: "Recent upgrades in pharmaceutical companies"
- **Comparative**: "AAPL vs GOOGL analyst sentiment comparison"
- **Filtered searches**: "Stocks with Strong Buy ratings and 20%+ target increases"

**⚡ Why Specificity Matters:**
- **Accurate database queries**: Specific prompts generate precise SQL
- **Relevant results**: Targeted searches return focused data
- **Better AI responses**: Clear context leads to accurate analysis
- **Faster responses**: Specific queries use cached context efficiently

**🚀 Pro Tips:**
- Mention specific **ticker symbols** for targeted analysis
- Include **brokerage names** for analyst-specific insights  
- Use **action words** like "raised", "lowered", "upgraded", "downgraded"
- Add **numerical criteria** for precise filtering

---

## ⚙️ Environment Configuration

### Backend Environment Variables (`backend/.env`)

| Variable | Description | Example |
|----------|-------------|----------|
| `DB_HOST` | CockroachDB cluster hostname | `cluster-name.aws-region.cockroachlabs.cloud` |
| `DB_PORT` | Database port (default: 26257) | `26257` |
| `DB_USER` | Database username | `your-username` |
| `DB_PASSWORD` | Database password | `your-database-password` |
| `DB_NAME` | Database name | `stock-market-db` |
| `DB_SSLMODE` | SSL connection mode | `require` |
| `API_TOKEN` | External stock API authentication token (assigned for this challenge) | `eyJhbGciOiJIUzI1NiIs...` |
| `OPENAI_API_KEY` | OpenAI API key for AI market analysis and chat | `sk-proj-...` |
| `PORT` | Backend server port | `8081` |

### Frontend Environment Variables (`frontend/.env`)

**NOTE:** by the moment there're no environment variables required for the frontend server.

---

## 🎯 Roadmap

- [x] **Parallel API Processing** – Implemented concurrent requests with rate limiting
- [x] **Swagger Documentation** – Auto-generated interactive API docs
- [x] **SQL Injection Protection** – Parameterized queries and input validation
- [x] **Hot Reload Development** – Air tool for automatic server restarts
- [x] **Statistical Recommendations** – Configurable weighted scoring algorithm with validation
- [x] **Market Analytics Dashboard** – Valuable statistics overview with AI assisted recommendations.
- [x] **AI Market Analysis** – GPT-4.1-nano integration with Wall Street analyst-level insights
- [x] **Dynamic Recommendations** – Flexible Top N display with pagination persistence
- [x] **Interactive AI Chat** – Real-time market discussion and analysis capabilities
- [x] **Unit Testing** – tests for backend and frontend components
- [ ] Enhance recommendation algorithm with external data sources  
- [ ] Deploy to cloud (Render / Vercel / Fly.io)  

---

## 🧑‍💻 Developer Notes

- Keep API keys and credentials **out of version control**.  
- Follow clean coding practices (linting, formatting, modular code).  
- Write unit tests for critical logic (backend + frontend).  

---

## 🧪 Testing

### **Backend Tests**
```bash
cd backend
go test ./handlers -v          # API handler tests
go test ./models -v            # Data model tests  
go test ./... -cover           # All tests with coverage
```

### **Frontend Tests**
```bash
cd frontend
npm test                       # Run all tests
npm run test:ui               # Interactive test UI
npm run test:coverage         # Coverage report
```

---

## 📜 License

This project is open-source and available under the [MIT License](LICENSE).  

---

## 🙌 Acknowledgments

Special thanks to the reviewers and interviewers for this challenge.  
This project was built as part of a technical assessment and continues to evolve with improvements.  