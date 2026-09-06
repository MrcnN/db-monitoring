import { useEffect, useState, useRef } from 'react';
import { Metric, HealthResult } from '../types';

interface LiveData {
  latestMetric: Metric | null;
  health: HealthResult | null;
  isConnected: boolean;
}

export function useLiveMetrics(dbId: string | undefined): LiveData {
  const [latestMetric, setLatestMetric] = useState<Metric | null>(null);
  const [health, setHealth] = useState<HealthResult | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!dbId) return;

    let reconnectTimeout: ReturnType<typeof setTimeout>;
    let ws: WebSocket;

    const connect = () => {
      // Use wss:// if https:// is used
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.hostname;
      // In development, the backend runs on 8080, in production it might be different
      // Since frontend is typically hitting /api/v1 via proxy or direct, we'll construct the URL
      const port = import.meta.env.VITE_API_URL ? new URL(import.meta.env.VITE_API_URL).port : '8080';
      const wsUrl = `${protocol}//${host}:${port}/api/v1/databases/${dbId}/live`;

      ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          if (message.type === 'metric_update') {
            setLatestMetric(message.data);
          } else if (message.type === 'health_update') {
            setHealth(message.data);
          }
        } catch (e) {
          console.error("Failed to parse websocket message", e);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Attempt to reconnect after 3 seconds
        reconnectTimeout = setTimeout(connect, 3000);
      };

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        ws.close();
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimeout);
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [dbId]);

  return { latestMetric, health, isConnected };
}
