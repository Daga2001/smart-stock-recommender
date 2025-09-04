# 📈 Smart Stock Recommender Project

A full-stack application that retrieves stock information from an external API, stores it in a scalable database, and presents insights through a clean, modern UI.  
The system also provides intelligent stock recommendations to help users identify the best investment opportunities today.  

---

## ✨ Features

- 🔗 **External API Integration** – Securely connects to the stock data API with proper authentication and error handling.  
- 🗄️ **Reliable Data Storage** – Persists data in CockroachDB for scalability and fault tolerance.  
- ⚡ **Backend API in Go** – Exposes stock data and recommendations via REST endpoints.  
- 🎨 **Interactive UI** – Built with React, Vite, TypeScript, and styled with Tailwind CSS + shadcn-ui.  
- 🔍 **Search & Filter** – Quickly find stocks by ticker, company, or brokerage.  
- 📊 **Sorting Options** – Sort by rating, target price, or analyst action.  
- 🤖 **Recommendation Engine** – Analyzes stock performance trends and suggests top picks.  
- 🧪 **Unit Testing** – Ensures stability and reliability of backend and UI logic.  

---

## 🛠️ Tech Stack

**Frontend**
- [Vite](https://vitejs.dev/) – blazing fast development build tool  
- [React](https://react.dev/) – component-based UI library  
- [TypeScript](https://www.typescriptlang.org/) – static typing for safer code  
- [shadcn-ui](https://ui.shadcn.com/) – accessible, headless components  
- [Tailwind CSS](https://tailwindcss.com/) – utility-first styling  

**Backend**
- [Golang](https://go.dev/) – high-performance, statically typed backend  
- RESTful API design (with potential GraphQL support)  

**Database**
- [CockroachDB](https://www.cockroachlabs.com/product/cockroachdb/) – distributed, resilient SQL database  

**Testing**
- Go testing framework for backend logic  
- React Testing Library / Vitest for frontend components  

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)  
- [Node.js 18+](https://nodejs.org/) & npm/yarn/pnpm  
- [CockroachDB](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html)  

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/smart-stock-recommender.git
   cd smart-stock-recommender
