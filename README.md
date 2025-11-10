# 🛍️ Kaleb E-Commerce Backend

A scalable and secure **e-commerce backend** built with **Go**, **Gin**, **GORM**, and **PostgreSQL**.  
Includes authentication, product management, order processing, caching, and rate-limiting middleware.

---

## 🚀 Tech Stack

| Layer | Technology |
|--------|-------------|
| Language | Go (1.20+) |
| Framework | Gin Web Framework |
| ORM | GORM |
| Database | PostgreSQL |
| Auth | JWT (golang-jwt/jwt) |
| Cache | go-cache (in-memory) |
| Rate Limiting | Custom middleware using go-cache |
| Containerization | Docker & Docker Compose |

---

## 📂 Project Structure

```
kalebecommerce/
├── cache/                 # In-memory cache logic
│   └── cache_service.go
├── cmd/                   # Application entry point
│   └── main.go
├── config/                # Configuration and DB setup
│   └── config.go
├── controllers/           # Business logic for routes
│   ├── auth_controller.go
│   ├── product_controller.go
│   ├── order_controller.go
│   └── *_test.go
├── db/
│   └── migrations/        # SQL migrations for schema setup
├── middleware/            # Auth and rate-limiting middleware
│   ├── auth_middleware.go
│   └── rate_limiter.go
├── routes/                # All route definitions
│   └── routes.go
├── utils/                 # Helper utilities
│   ├── password.go
│   ├── response.go
│   └── upload_image.go
├── .env                   # Environment variables (not committed)
├── docker-compose.yml     # Local container orchestration
├── go.mod / go.sum        # Go module definitions
└── README.md              # Project documentation
```

---

## ⚙️ Environment Variables

Create a `.env` file in the project root:

```env
DATABASE_URL=host=localhost user=postgres password=postgres dbname=ecom port=5432 sslmode=disable TimeZone=UTC
JWT_SECRET=replace-with-strong-secret
PORT=8080
```

---

## 🐳 Run with Docker

```bash
docker-compose up --build
```

This will spin up both the **PostgreSQL** database and the **API server**.

---

## 🧩 Run Locally (without Docker)

1. **Install dependencies**
   ```bash
   go mod tidy
   ```

2. **Run the server**
   ```bash
   go run ./cmd
   ```

3. **API will be available at:**  
   👉 `http://localhost:8080/api`

---

## 🔐 Authentication Endpoints

| Method | Endpoint | Description |
|--------|-----------|--------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Log in and receive a JWT token |

---

## 🛒 Product Endpoints

| Method | Endpoint | Access | Description |
|--------|-----------|---------|--------------|
| GET | `/api/products` | Public | List/search products (cached) |
| GET | `/api/products/:id` | Public | View single product |
| POST | `/api/products` | Admin | Create new product |
| PUT | `/api/products/:id` | Admin | Update product |
| DELETE | `/api/products/:id` | Admin | Delete product |

---

## 📦 Order Endpoints

| Method | Endpoint | Access | Description |
|--------|-----------|---------|--------------|
| POST | `/api/orders` | Authenticated | Place a new order |
| GET | `/api/orders` | Authenticated | List user orders |

---

## ⚡ Middleware

- **AuthRequired** → Validates JWT token for protected routes.  
- **AdminOnly** → Restricts access to admin-only endpoints.  
- **RateLimitMiddleware** → Limits clients to `5 requests / 10 seconds` by IP.  
- **Cache Service** → Used for caching frequently accessed product data.

---

## 🧠 Features

- 🔐 Secure JWT Authentication  
- 🧮 Product management (CRUD)  
- 💰 Order placement with transaction safety  
- ⚡ In-memory caching for performance  
- 🚦 Rate limiting to prevent abuse  
- 🐳 Docker support for easy deployment  
- ✅ Unit test structure ready

---

## 🧪 Run Tests

```bash
cd ./controllers
go test ./...
```

---

## 📄 License

MIT © 2025 [Kaleb Tilahun]
