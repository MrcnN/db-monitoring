import api from '../lib/axios';
import { Metric, HealthResult, SlowQuery } from '../types';

export const getMetrics = async (dbId: string, timeRange: string = '1h'): Promise<Metric[]> => {
  const response = await api.get(`/databases/${dbId}/metrics?range=${timeRange}`);
  return response.data?.data || [];
};

export const getLatestMetric = async (dbId: string): Promise<Metric | null> => {
  const response = await api.get(`/databases/${dbId}/metrics/latest`);
  return response.data?.data || null;
};

export const getDatabaseHealth = async (dbId: string): Promise<HealthResult> => {
  const response = await api.get(`/databases/${dbId}/health`);
  return response.data?.data || {
    status: 'unknown',
    score: 50,
    issues: [],
    summary: 'Health data not available',
  };
};

export const getSlowQueries = async (dbId: string): Promise<SlowQuery[]> => {
  const response = await api.get(`/databases/${dbId}/slow-queries`);
  return response.data?.data || [];
};
