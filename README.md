# Database Health & Performance Platform

```text
    ____  ____    ____  __      __  ____                       
   / __ \/ __ )  / __ \/ /___ _/ /_/ __/___  _________ ___  
  / / / / __  | / /_/ / / __ `/ __/ /_/ __ \/ ___/ __ `__ \ 
 / /_/ / /_/ / / ____/ / /_/ / /_/ __/ /_/ / /  / / / / / / 
/_____/_____/ /_/   /_/\__,_/\__/_/  \____/_/  /_/ /_/ /_/  
```

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Docker Ready](https://img.shields.io/badge/docker-ready-blue.svg)](https://www.docker.com/)

## Overview
The Database Health & Performance Platform is a unified monitoring and management solution for databases. It provides deep visibility into database health, tracks slow queries, generates metrics, and acts on alerts to ensure your data infrastructure is performant and secure.

## Features
- ✅ **DB Management**: Connect and monitor PostgreSQL and MySQL targets
- ✅ **RBAC Auth**: JWT-based authentication with role-based access control
- ✅ **Audit Logging**: Comprehensive activity tracking and security logging
- ✅ **Docker Ready**: Fully containerized with a streamlined `docker-compose` setup
- ✅ **Prometheus Metrics**: Export performance data for rich visualizations
- 🔄 **Metrics & Collectors**: Modular architecture for database statistics
- 📋 **Query Analysis**: Slow query logging and optimization insights (Phase 3)
- 📋 **Alert Engine**: Incident management and smart alerts (Phase 4)

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
```bash
git clone https://github.com/example/dbplatform.git
cd dbplatform
cp .env.example .env

# Edit .env with your secrets
make generate-secrets  # generate JWT_SECRET and ENCRYPTION_KEY

docker compose up -d
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

## Roadmap
| Phase | Status | Focus |
| ----- | ------ | ----- |
| Phase 1 | ✅ | Foundation (auth, DB CRUD, Docker) |
| Phase 2 | 🔄 | Metrics & Collectors |
| Phase 3 | 📋 | Query Analysis |
| Phase 4 | 📋 | Alerts & Incidents |
| Phase 5 | 📋 | Schema & Index Analysis |
| Phase 6 | 📋 | Advanced Observability |
| Phase 7 | 📋 | Production Hardening |

## Contributing
Please see `CONTRIBUTING.md` (coming soon) for details on code of conduct and the submission process.

## License
MIT License.
