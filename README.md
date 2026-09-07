# Database Health & Performance Platform

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Docker Ready](https://img.shields.io/badge/docker-ready-blue.svg)](https://www.docker.com/)

## Overview
The Database Health & Performance Platform is a unified monitoring and management solution for databases. It provides deep visibility into database health, tracks slow queries, generates metrics, and acts on alerts to ensure your data infrastructure is performant and secure.

## Features
-  **DB Management**: Connect and monitor PostgreSQL and MySQL targets
-  **Live Telemetry**: Track Connections, Cache Hits, Query Rates, and Latency
-  **Interactive SQL Console**: Execute raw queries safely on your databases directly from the dashboard
-  **Live Lock & Blocking Query Monitor**: Instantly visualize queries that are blocking each other (Deadlock/Contention analysis)
-  **Storage & Bloat Analyzer**: Identify wasted space (dead tuples) and optimize table/index footprints
-  **Smart Query Advisor**: Heuristics-based and AI-powered recommendations for optimizing slow queries
-  **Visual EXPLAIN Plan**: Turn complex database execution plans into clear, visual trees
-  **Alert Engine & Incidents**: Get notified via custom channels when thresholds are breached
-  **RBAC Auth & Audit Logging**: Enterprise-grade security and activity tracking
-  **Docker Ready**: Fully containerized with a streamlined `docker-compose` setup

## Architecture
```text
               +-----------+
               | Frontend  |
               +-----+-----+
                     |
               +-----v-----+
               |   Go API  |
               +--+--+--+--+
                  |  |  |
      +-----------+  |  +------------+
      |              |               |
+-----v----+   +-----v-----+  +------v------+
| Postgres |   |   Redis   |  | Alert Engine|
| (Config) |   | (Cache)   |  | (Future)    |
+----------+   +-----------+  +-------------+
                     |
               +-----v-----+
               | Collectors|
               +--+--+--+--+
                  |  |  |
         +--------+  |  +--------+
         |           |           |
    +----v---+  +----v---+  +----v---+
    | Target |  | Target |  | Target |
    |   DB   |  |   DB   |  |   DB   |
    +--------+  +--------+  +--------+
```

## Tech Stack
| Backend         | Frontend        | Infrastructure    |
| --------------- | --------------- | ----------------- |
| Go 1.23         | React           | Docker / Compose  |
| PostgreSQL      | TypeScript      | Prometheus        |
| Redis           | Tailwind CSS    | Grafana           |

## Quick Start

### Windows Kullanıcıları İçin (Tek Tıkla Başlatma)
Projeyi indirdikten sonra ana dizinde bulunan `baslat.bat` dosyasına çift tıklayarak sistemi tek seferde ayağa kaldırabilirsiniz. Bu dosya, projeyi kendi ihtiyaçlarınıza göre düzenlediğinizde de (yeni özellik eklediğinizde vb.) hızlıca derleyip açmanız için tasarlanmıştır. `baslat.bat` içeriğini sağ tıklayıp düzenleyerek dilediğiniz gibi değiştirebilirsiniz.

### Terminal/CLI Üzerinden
```bash
git clone https://github.com/MrcnN/db-monitoring.git
cd db-monitoring
cp .env.example .env

# Edit .env with your secrets
make generate-secrets  # generate JWT_SECRET and ENCRYPTION_KEY

docker compose up --build -d
# Visit http://localhost:3000
```

## Manual Installation
Requirements: Go 1.23+, PostgreSQL, Redis
1. Start local Postgres and Redis
2. Execute `make generate-secrets`
3. Configure `.env`
4. `make dev`

## Configuration Reference
| Env Var | Description | Default |
| ------- | ----------- | ------- |
| `APP_ENV` | Application environment (development/production) | `development` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `REDIS_URL` | Redis connection string | - |
| `JWT_SECRET` | Secret key for JWT signing | - |
| `ENCRYPTION_KEY` | 32-byte hex string for AES | - |
| `LOG_LEVEL` | Logging level | `info` |

## API Documentation
Core API paths:
- `POST /api/v1/auth/login` - Authenticate
- `GET /api/v1/databases` - List databases
- `POST /api/v1/databases` - Register database

## Development
- `make dev`: Start the application with hot-reload or basic run
- `make test`: Run all unit tests
- `make test-integration`: Run integration tests (requires Docker)

## Project Structure
```
dbplatform/
├── cmd/           # Application entrypoints
├── deploy/        # Demo configurations, Grafana/Prometheus
├── docker/        # Dockerfiles
├── docs/          # Documentation
├── frontend/      # React frontend application
├── internal/      # Go application code
├── migrations/    # SQL database migrations
└── tests/         # Integration and E2E tests
```


## License
MIT License.
