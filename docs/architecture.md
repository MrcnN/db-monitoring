# Architecture

## System Overview
The Database Health & Performance Platform uses a modular, API-first architecture designed to decouple configuration management, metric collection, and frontend presentation.

## Component Descriptions
- **Frontend**: A React single-page application (SPA).
- **Go API**: Core service providing REST endpoints for UI and integration.
- **Config Database (PostgreSQL)**: Stores user accounts, RBAC, monitored database configurations, and audit logs.
- **Cache (Redis)**: Session handling, rate limiting, and ephemeral task queues.
- **Collectors**: Pluggable modules inside the Go API that periodically fetch metrics from target databases.
- **Prometheus/Grafana**: Ingest and visualize platform health and generic system metrics.

## Architecture Diagram
```text
+-------------------+
|      Client       |
+---------+---------+
          |
+---------v---------+
|     Frontend      |
+---------+---------+
          | (REST)
+---------v---------+
|      Go API       |
+---+---+---+---+---+
    |   |   |   |
    |   |   |   +------> [ Collectors ] ----> [ Target DBs ]
    |   |   |
    |   |   +----------> [ PostgreSQL (Config) ]
    |   |
    |   +--------------> [ Redis (Cache & Limits) ]
    |
    +------------------> [ Prometheus ]
```

## Request Flow Walkthrough
1. **Login**: User submits credentials. API checks PostgreSQL, sets a session in Redis, and returns a JWT.
2. **Register DB**: Authenticated user provides DB credentials. API encrypts the password with AES-256-GCM and stores the config in PostgreSQL.
3. **Monitor**: The collector subsystem wakes up on an interval, decrypts the credentials in memory, connects to the target DB, executes statistical queries, and records the results.

## Security Architecture
- Credentials at rest are encrypted.
- Stateless-like JWT augmented with stateful Redis invalidation for rapid revocation.
- Fine-grained RBAC on API endpoints.

## Scalability Considerations
- **Stateless API**: Can be scaled horizontally.
- **Redis Queue**: Background tasks and collector scheduling can be distributed.

## Future Extensibility
- **Machine Learning**: Anomaly detection module can subscribe to collector streams.
- **New Databases**: New collector implementations easily added by fulfilling the `Collector` interface.
