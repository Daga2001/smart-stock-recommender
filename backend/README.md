# Backend - Smart Stock Recommender

Go backend server that fetches stock data from external API and stores it in CockroachDB.

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Set up environment variables in `.env`:
```
DB_HOST=razzed-kelpie-15584.j77.aws-us-east-1.cockroachlabs.cloud
DB_PORT=26257
DB_USER=david
DB_NAME=stock-market-db
DB_SSLMODE=require
API_TOKEN=your_token_here
PORT=8080
```

3. Run the server:
```bash
go run main.go
```

## API Endpoints

- `POST /api/stocks` - Fetch stocks by page number
  - Body: `{"page": 1}`
  - Returns: Stock data from external API and stores in database