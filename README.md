# 📈 Smart Stock Recommender Project

A full-stack application that retrieves stock information from an external API, stores it in a scalable database, and presents insights through a beautiful and interactive UI.  
The system also provides intelligent stock recommendations to help users identify the best investment opportunities today.  

---

## ✨ Features

- 🔗 **External API Integration** – Securely connects to the stock data API with proper authentication and error handling.  
- 🗄️ **Reliable Data Storage** – Persists data in CockroachDB for scalability and fault tolerance.  
- ⚡ **Backend API in Go** – Exposes stock data and recommendations via REST endpoints.  
- 🎨 **Interactive UI** – Built with Vue 3, TypeScript, Pinia, and styled with Tailwind CSS.  
- 🔍 **Search & Filter** – Quickly find stocks by ticker, company, or brokerage.  
- 📊 **Sorting Options** – Sort by rating, target price, or analyst action.  
- 🤖 **Recommendation Engine** – Analyzes stock performance trends and suggests top picks.  
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
- RESTful API design  

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

### `POST /api/stocks`
Fetch stock data by page number from external API and store in database.

**Request:**
- **URL:** `http://localhost:8081/api/stocks`
- **Method:** `POST`
- **Headers:** `Content-Type: application/json`
- **Body:** `{"page": 1}`

**Response:**
```json
{
  "items": [
    {
      "ticker": "CECO",
      "target_from": "$44.00",
      "target_to": "$52.00",
      "company": "CECO Environmental",
      "action": "target raised by",
      "brokerage": "Needham & Company LLC",
      "rating_from": "Buy",
      "rating_to": "Buy",
      "time": "2025-08-22T00:30:05.141533767Z"
    }
  ],
  "next_page": "CECO"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8081/api/stocks \
  -H "Content-Type: application/json" \
  -d '{"page": 1}'
```

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
| `API_TOKEN` | External stock API authentication token | `eyJhbGciOiJIUzI1NiIs...` |
| `PORT` | Backend server port | `8081` |

### Frontend Environment Variables (`frontend/.env`)

| Variable | Description | Example |
|----------|-------------|----------|
| `VITE_STOCK_API_TOKEN` | External stock API token (Vite prefix required) | `eyJhbGciOiJIUzI1NiIs...` |

---

## 🎯 Roadmap

- [ ] Implement caching for API requests  
- [ ] Add authentication for UI users  
- [ ] Enhance recommendation algorithm with external data sources  
- [ ] Deploy to cloud (Render / Vercel / Fly.io)  

---

## 🧑‍💻 Developer Notes

- Keep API keys and credentials **out of version control**.  
- Follow clean coding practices (linting, formatting, modular code).  
- Write unit tests for critical logic (backend + frontend).  

---

## 📜 License

This project is open-source and available under the [MIT License](LICENSE).  

---

## 🙌 Acknowledgments

Special thanks to the reviewers and interviewers for this challenge.  
This project was built as part of a technical assessment and continues to evolve with improvements.  

