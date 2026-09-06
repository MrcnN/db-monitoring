export interface User {
  id: string;
  email: string;
  full_name: string;
  role: 'admin' | 'operator' | 'viewer';
  is_active: boolean;
  created_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  full_name: string;
}

export interface MonitoredDatabase {
  id: string;
  name: string;
  description: string;
  type: 'postgresql' | 'mysql';
  host: string;
  port: number;
  database_name: string;
  username: string;
  ssl_mode: string;
  monitoring_interval: number;
  status: 'active' | 'inactive' | 'error';
  is_monitoring_enabled: boolean;
  last_checked_at: string | null;
  created_at: string;
}

export interface CreateDatabaseRequest {
  name: string;
  description?: string;
  type: 'postgresql' | 'mysql';
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
  ssl_mode?: string;
  monitoring_interval?: number;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}

export interface AuditLog {
  id: string;
  user_email: string;
  action: string;
  resource_type: string;
  resource_name: string;
  status: string;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface Metric {
  id: string;
  database_id: string;
  cpu_usage: number;
  memory_usage: number;
  connections_total: number;
  connections_active: number;
  connections_idle: number;
  connection_usage_pct: number;
  query_rate: number;
  p95_latency_ms: number;
  cache_hit_ratio: number;
  database_size_bytes: number;
  dead_tuples_count: number;
  active_locks_count: number;
  replication_lag_seconds: number;
  created_at: string;
}

export interface HealthIssue {
  severity: 'info' | 'warning' | 'critical';
  title: string;
  description: string;
  current_value: string;
  threshold: string;
}

export interface HealthResult {
  status: 'healthy' | 'warning' | 'critical' | 'unknown';
  score: number;
  issues: HealthIssue[];
  summary: string;
}

export interface SlowQuery {
  query: string;
  calls: number;
  total_time_ms: number;
  mean_time_ms: number;
  max_time_ms: number;
  rows: number;
}
