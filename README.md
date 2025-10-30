# Attendance Backend API

RESTful API backend untuk sistem absensi berbasis GPS, dibangun dengan Golang dan Gin framework.

## 🛠️ Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **Database**: PostgreSQL + PostGIS
- **ORM**: GORM
- **Authentication**: JWT
- **Validation**: go-playground/validator

## 📁 Project Structure

```
attendance-backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── controller/              # HTTP handlers
│   ├── middleware/              # Middleware (auth, cors, etc)
│   ├── model/                   # Data models
│   ├── repository/              # Database operations
│   ├── service/                 # Business logic
│   └── utils/                   # Helper functions
├── pkg/
│   ├── database/                # Database connection
│   ├── jwt/                     # JWT utilities
│   └── validator/               # Custom validators
├── migrations/                  # SQL migrations
├── .env.example                 # Environment variables template
├── go.mod                       # Go modules
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL with PostGIS extension
- Docker (optional, for database)

### Installation

1. Clone the repository and navigate to backend directory:
```bash
cd attendance-backend
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Update `.env` with your configuration:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=attendance_db
JWT_SECRET=your-secret-key
```

4. Install dependencies:
```bash
go mod download
```

5. Run the application:
```bash
go run cmd/api/main.go
```

Server will start on `http://localhost:8000`

## 📝 API Documentation

### Health Check
```
GET /health
```

### Authentication
```
POST   /api/v1/auth/register          # Register new user
POST   /api/v1/auth/login             # Login user
POST   /api/v1/auth/refresh-token     # Refresh JWT token
POST   /api/v1/auth/logout            # Logout user
GET    /api/v1/auth/me                # Get current user info
```

### Attendance (User)
```
GET    /api/v1/attendance/locations              # Get nearby locations
POST   /api/v1/attendance/check-in                # Check-in
POST   /api/v1/attendance/check-out               # Check-out
GET    /api/v1/attendance/history                 # Get history
GET    /api/v1/attendance/today                   # Get today's attendance
GET    /api/v1/attendance/status                  # Check current status
POST   /api/v1/attendance/validate-location      # Validate location
```

### Admin - Users
```
GET    /api/v1/admin/users                # Get all users
GET    /api/v1/admin/users/:id            # Get user detail
POST   /api/v1/admin/users                # Create user
PUT    /api/v1/admin/users/:id            # Update user
DELETE /api/v1/admin/users/:id            # Delete user
PATCH  /api/v1/admin/users/:id/status     # Change status
```

### Admin - Locations
```
GET    /api/v1/admin/locations            # Get all locations
GET    /api/v1/admin/locations/:id        # Get location detail
POST   /api/v1/admin/locations            # Create location
PUT    /api/v1/admin/locations/:id        # Update location
DELETE /api/v1/admin/locations/:id        # Delete location
```

### Admin - Reports
```
GET    /api/v1/admin/attendances                 # Get all attendances
GET    /api/v1/admin/attendances/:id             # Get attendance detail
GET    /api/v1/admin/reports/daily               # Daily report
GET    /api/v1/admin/reports/monthly             # Monthly report
GET    /api/v1/admin/reports/export              # Export CSV/Excel
```

## 🧮 GPS Validation

Backend menggunakan Haversine Formula untuk menghitung jarak antara koordinat user dengan lokasi absen:

```go
distance := CalculateDistance(
    userLat, userLon,
    locationLat, locationLon
)

if distance <= radius {
    // User dalam radius, boleh absen
}
```

## 🗄️ Database

### Run Migrations

Migrations akan otomatis dijalankan saat container PostgreSQL pertama kali distart via docker-compose.

Manual migration:
```bash
psql -U postgres -d attendance_db -f migrations/001_init_schema.sql
```

### Default Admin User

- Email: `admin@attendance.com`
- Password: `admin123`

**⚠️ Ubah password di production!**

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/service/...
```

## 📦 Build

```bash
# Build for current platform
go build -o bin/api cmd/api/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/api-linux cmd/api/main.go

# Run binary
./bin/api
```

## 🔐 Security

- JWT authentication
- Password hashing with bcrypt
- CORS configuration
- Input validation
- SQL injection prevention (GORM)
- Rate limiting

## 🐳 Docker

```bash
# Build image
docker build -t attendance-backend .

# Run container
docker run -p 8000:8000 --env-file .env attendance-backend
```

## 📊 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8000 |
| `GIN_MODE` | Gin mode (debug/release) | debug |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | postgres |
| `DB_NAME` | Database name | attendance_db |
| `JWT_SECRET` | JWT secret key | required |
| `JWT_EXPIRATION` | Token expiration | 24h |

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Write tests
4. Submit pull request

## 📝 License

Proprietary

---

**Version:** 1.0.0
**Status:** In Development
