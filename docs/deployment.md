# Deployment Guide

## Prerequisites
- Docker & Docker Compose V2.
- Dedicated VM or cluster (e.g., AWS EC2, DigitalOcean Droplet).

## Deployment Steps
1. Clone the repository on the target server.
2. Run `make generate-secrets` to create robust keys.
3. Copy `.env.example` to `.env` and insert the generated secrets.
4. Modify `APP_ENV=production`.
5. Run `docker compose build`.
6. Run `docker compose up -d`.

## Environment Variables
Ensure `JWT_SECRET` and `ENCRYPTION_KEY` are backed up securely. Losing `ENCRYPTION_KEY` means all saved target database passwords are lost.

## Health Check Verification
- API: `http://localhost:8080/health`
- Compose handles automatic restarts on failure.

## Log Monitoring
- `docker compose logs -f api`
- Use Grafana (`http://localhost:3001`) to view runtime metrics.

## Backup Strategy
- Periodically dump the `postgres_data` volume.
- Use `pg_dump` targeting the `postgres` service.

## Updating the Platform
```bash
git pull origin main
docker compose build
docker compose up -d
```
