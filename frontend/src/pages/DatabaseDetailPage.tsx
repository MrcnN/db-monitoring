import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeftIcon,
  ArrowPathIcon,
  CircleStackIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  BoltIcon,
  CpuChipIcon,
  ServerIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

import { getDatabase, testExistingConnection } from '../api/databases';
import { getSlowQueries, getMetrics, getLatestMetric, getDatabaseHealth } from '../api/metrics';
import { TimeSeriesChart } from '../components/charts/TimeSeriesChart';
import LoadingSpinner from '../components/common/LoadingSpinner';
import { SlowQueryTable } from '../components/database/SlowQueryTable';
import { AlertsTable } from '../components/database/AlertsTable';
import { SqlConsole } from '../components/database/SqlConsole';
import { useLiveMetrics } from '../hooks/useLiveMetrics';
import { Metric } from '../types';
import { getAlerts } from '../api/alerts';

export default function DatabaseDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [timeRange, setTimeRange] = useState<string>('1h');
  const [activeTab, setActiveTab] = useState<'metrics' | 'slow_queries' | 'alerts' | 'sql_console'>('metrics');

  const { data: db, isLoading: dbLoading } = useQuery({
    queryKey: ['database', id],
    queryFn: () => getDatabase(id!),
    enabled: !!id,
  });

  // Fetch initial data once, without polling
  const { data: initialLatestMetric, isLoading: metricLoading } = useQuery({
    queryKey: ['latestMetric', id],
    queryFn: () => getLatestMetric(id!),
    enabled: !!id,
  });

  const { data: timeSeries = [] } = useQuery({
    queryKey: ['timeSeries', id, timeRange],
    queryFn: () => getMetrics(id!, timeRange),
    enabled: !!id,
  });

  const { data: initialHealthData } = useQuery({
    queryKey: ['databaseHealth', id],
    queryFn: () => getDatabaseHealth(id!),
    enabled: !!id,
  });

  const { data: alerts = [], isLoading: alertsLoading } = useQuery({
    queryKey: ['databaseAlerts', id],
    queryFn: () => getAlerts(id!),
    enabled: !!id && activeTab === 'alerts',
    refetchInterval: 30000,
  });

  // Connect to WebSocket for live updates
  const { latestMetric: liveMetric, health: liveHealth, isConnected } = useLiveMetrics(id);

  // Combine initial state with live state
  const latestMetric = liveMetric || initialLatestMetric;
  const healthData = liveHealth || initialHealthData;

  // Append live metrics to timeSeries dynamically without refetching all historical data
  useEffect(() => {
    if (liveMetric && id) {
      queryClient.setQueryData<Metric[]>(['timeSeries', id, timeRange], (oldData) => {
        if (!oldData) return [liveMetric];
        // Keep the array size reasonable (e.g. last 100 points or so depending on time range)
        // Here we just append. In a real app we might shift old ones.
        return [...oldData, liveMetric];
      });
    }
  }, [liveMetric, id, timeRange, queryClient]);

  const { data: slowQueries = [], isLoading: slowQueriesLoading } = useQuery({
    queryKey: ['slowQueries', id],
    queryFn: () => getSlowQueries(id!),
    enabled: !!id && activeTab === 'slow_queries',
    refetchInterval: 30000,
  });

  const testMutation = useMutation({
    mutationFn: () => testExistingConnection(id!),
    onSuccess: (data) => {
      if (data.success) {
        toast.success(data.message || 'Connection test successful!');
      } else {
        toast.error(data.message || 'Connection test failed');
      }
    },
    onError: () => {
      toast.error('Connection test failed');
    },
  });

  if (dbLoading || metricLoading) {
    return <LoadingSpinner />;
  }

  if (!db) {
    return (
      <div className="text-center py-12">
        <h3 className="text-lg font-medium text-white">Database not found</h3>
        <button
          onClick={() => navigate('/databases')}
          className="mt-4 inline-flex items-center text-sm text-primary-400 hover:text-primary-300"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" /> Back to Databases
        </button>
      </div>
    );
  }

  const connData = timeSeries.map((m) => ({
    time: new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: m.connections_active,
  }));

  const cacheHitData = timeSeries.map((m) => ({
    time: new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: m.cache_hit_ratio,
  }));

  const queryRateData = timeSeries.map((m) => ({
    time: new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: m.query_rate,
  }));

  const latencyData = timeSeries.map((m) => ({
    time: new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: m.p95_latency_ms,
  }));

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 MB';
    const mb = bytes / (1024 * 1024);
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(2)} GB`;
    }
    return `${mb.toFixed(1)} MB`;
  };

  const getHealthBadge = () => {
    const status = healthData?.status || 'unknown';
    switch (status) {
      case 'healthy':
        return (
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-green-900/60 text-green-300 border border-green-700">
            <CheckCircleIcon className="h-4 w-4 mr-1.5 text-green-400" />
            HEALTHY ({healthData?.score ?? 100}/100)
          </span>
        );
      case 'warning':
        return (
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-yellow-900/60 text-yellow-300 border border-yellow-700">
            <ExclamationTriangleIcon className="h-4 w-4 mr-1.5 text-yellow-400" />
            WARNING ({healthData?.score ?? 70}/100)
          </span>
        );
      case 'critical':
        return (
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-red-900/60 text-red-300 border border-red-700">
            <XCircleIcon className="h-4 w-4 mr-1.5 text-red-400" />
            CRITICAL ({healthData?.score ?? 0}/100)
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-gray-800 text-gray-400 border border-gray-700">
            UNKNOWN
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-3">
          <button
            onClick={() => navigate('/databases')}
            className="p-1.5 rounded-lg bg-gray-800 text-gray-400 hover:text-white border border-gray-700"
          >
            <ArrowLeftIcon className="h-5 w-5" />
          </button>
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-2xl font-bold text-white tracking-tight">{db.name}</h1>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded text-xs font-medium bg-blue-900/50 text-blue-300 border border-blue-800 uppercase">
                {db.type}
              </span>
              {getHealthBadge()}
            </div>
            <p className="text-xs text-gray-400 mt-1">
              Host: <span className="text-gray-300 font-mono">{db.host}:{db.port}</span> | Database: <span className="text-gray-300">{db.database_name}</span> | Interval: {db.monitoring_interval}s
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={() => testMutation.mutate()}
            disabled={testMutation.isPending}
            className="inline-flex items-center px-3.5 py-2 border border-gray-700 text-xs font-medium rounded-md text-gray-200 bg-gray-800 hover:bg-gray-700 focus:outline-none transition-colors"
          >
            <BoltIcon className={`h-4 w-4 mr-1.5 text-yellow-400 ${testMutation.isPending ? 'animate-spin' : ''}`} />
            {testMutation.isPending ? 'Testing...' : 'Test Connection'}
          </button>
        </div>
      </div>

      {healthData?.issues && healthData.issues.length > 0 && (
        <div className="bg-yellow-950/40 border border-yellow-700/60 rounded-lg p-4">
          <div className="flex items-start space-x-3">
            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-400 mt-0.5 flex-shrink-0" />
            <div className="flex-1">
              <h3 className="text-sm font-semibold text-yellow-200">
                Active Health Findings ({healthData.issues.length})
              </h3>
              <p className="text-xs text-yellow-300/80 mt-0.5 mb-2">{healthData.summary}</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-2">
                {healthData.issues.map((issue, idx) => (
                  <div key={idx} className="bg-gray-900/60 rounded p-2.5 border border-yellow-800/40 text-xs">
                    <div className="flex justify-between font-medium text-yellow-100 mb-1">
                      <span>{issue.title}</span>
                      <span className="text-gray-400">Current: {issue.current_value}</span>
                    </div>
                    <p className="text-gray-300 text-[11px]">{issue.description}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">Connections</span>
            <ServerIcon className="h-4 w-4 text-blue-400" />
          </div>
          <div className="text-xl font-bold text-white">
            {latestMetric?.connections_active ?? 0}
            <span className="text-xs font-normal text-gray-400 ml-1">
              / {latestMetric?.connections_total ?? 100}
            </span>
          </div>
          <p className="text-[11px] text-gray-400 mt-1">
            {(latestMetric?.connection_usage_pct ?? 0).toFixed(1)}% pool used
          </p>
        </div>

        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">Cache Hit Ratio</span>
            <CheckCircleIcon className="h-4 w-4 text-green-400" />
          </div>
          <div className="text-xl font-bold text-green-400">
            {(latestMetric?.cache_hit_ratio ?? 99.0).toFixed(1)}%
          </div>
          <p className="text-[11px] text-gray-400 mt-1">Target: &gt; 95%</p>
        </div>

        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">Query Rate</span>
            <BoltIcon className="h-4 w-4 text-yellow-400" />
          </div>
          <div className="text-xl font-bold text-white">
            {(latestMetric?.query_rate ?? 0).toFixed(0)}
            <span className="text-xs font-normal text-gray-400 ml-1">TPS</span>
          </div>
          <p className="text-[11px] text-gray-400 mt-1">Transactions / sec</p>
        </div>

        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">P95 Latency</span>
            <ClockIcon className="h-4 w-4 text-purple-400" />
          </div>
          <div className="text-xl font-bold text-white">
            {(latestMetric?.p95_latency_ms ?? 0.5).toFixed(1)}
            <span className="text-xs font-normal text-gray-400 ml-1">ms</span>
          </div>
          <p className="text-[11px] text-gray-400 mt-1">Response time</p>
        </div>

        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">Database Size</span>
            <CircleStackIcon className="h-4 w-4 text-indigo-400" />
          </div>
          <div className="text-xl font-bold text-white">
            {formatBytes(latestMetric?.database_size_bytes ?? 0)}
          </div>
          <p className="text-[11px] text-gray-400 mt-1">Disk footprint</p>
        </div>

        <div className="bg-gray-800 p-3.5 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between text-gray-400 mb-1">
            <span className="text-xs">Est. CPU Load</span>
            <CpuChipIcon className="h-4 w-4 text-cyan-400" />
          </div>
          <div className="text-xl font-bold text-white">
            {(latestMetric?.cpu_usage ?? 5.0).toFixed(0)}%
          </div>
          <p className="text-[11px] text-gray-400 mt-1">Active worker load</p>
        </div>
      </div>

      <div className="border-b border-gray-700">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('metrics')}
            className={`${
              activeTab === 'metrics'
                ? 'border-indigo-500 text-indigo-400'
                : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors`}
          >
            Metrics Dashboard
          </button>
          <button
            onClick={() => setActiveTab('slow_queries')}
            className={`${
              activeTab === 'slow_queries'
                ? 'border-indigo-500 text-indigo-400'
                : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors`}
          >
            Slow Query Analyzer
          </button>
          <button
            onClick={() => setActiveTab('alerts')}
            className={`${
              activeTab === 'alerts'
                ? 'border-indigo-500 text-indigo-400'
                : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors`}
          >
            Alerts History
          </button>
          <button
            onClick={() => setActiveTab('sql_console')}
            className={`${
              activeTab === 'sql_console'
                ? 'border-zinc-100 text-zinc-100'
                : 'border-transparent text-zinc-500 hover:text-zinc-300 hover:border-zinc-700'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors`}
          >
            SQL Console
          </button>
        </nav>
      </div>

      {activeTab === 'metrics' ? (
        <>
          <div className="flex items-center justify-between pt-2">
            <div className="flex items-center space-x-2">
              <span className={`relative flex h-2 w-2 ${!isConnected ? 'opacity-50' : ''}`}>
                {isConnected && (
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                )}
                <span className={`relative inline-flex rounded-full h-2 w-2 ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></span>
              </span>
              <span className="text-xs text-gray-400">
                {isConnected ? 'Live Telemetry Active' : 'Live Telemetry Disconnected'}
              </span>
            </div>

            <div className="inline-flex rounded-md shadow-sm bg-gray-800 p-0.5 border border-gray-700">
              {(['15m', '1h', '6h', '24h'] as const).map((r) => (
                <button
                  key={r}
                  onClick={() => setTimeRange(r)}
                  className={`px-3 py-1 text-xs font-medium rounded ${
                    timeRange === r
                      ? 'bg-blue-600 text-white shadow-sm'
                      : 'text-gray-400 hover:text-white hover:bg-gray-700'
                  }`}
                >
                  {r.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <TimeSeriesChart
              title="Active Connections"
              data={connData}
              unit="conn"
              color="#3b82f6"
              threshold={latestMetric?.connections_total ? latestMetric.connections_total * 0.8 : 80}
            />

            <TimeSeriesChart
              title="Cache Hit Ratio (%)"
              data={cacheHitData}
              unit="%"
              color="#10b981"
              threshold={95}
            />

            <TimeSeriesChart
              title="Transactions / Query Rate"
              data={queryRateData}
              unit="TPS"
              color="#f59e0b"
            />

            <TimeSeriesChart
              title="Query Latency (P95)"
              data={latencyData}
              unit="ms"
              color="#8b5cf6"
              threshold={200}
            />
          </div>
        </>
      ) : activeTab === 'slow_queries' ? (
        <div className="pt-2">
          <SlowQueryTable queries={slowQueries} isLoading={slowQueriesLoading} databaseId={id} />
        </div>
      ) : activeTab === 'alerts' ? (
        <div className="pt-2">
          <AlertsTable alerts={alerts} isLoading={alertsLoading} />
        </div>
      ) : (
        <div className="pt-6">
          <SqlConsole databaseId={id!} />
        </div>
      )}
    </div>
  );
}
