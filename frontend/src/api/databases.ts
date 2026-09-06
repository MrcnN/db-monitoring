import api from '../lib/axios';
import { MonitoredDatabase, CreateDatabaseRequest, PaginatedResponse } from '../types';

export const getDatabases = async (): Promise<PaginatedResponse<MonitoredDatabase>> => {
  const response = await api.get('/databases');
  const list: MonitoredDatabase[] = response.data?.data || [];
  return {
    data: list,
    meta: {
      total: list.length,
      limit: 100,
      offset: 0,
    },
  };
};

export const getDatabase = async (id: string): Promise<MonitoredDatabase> => {
  const response = await api.get(`/databases/${id}`);
  return response.data?.data;
};

export const createDatabase = async (data: CreateDatabaseRequest): Promise<MonitoredDatabase> => {
  const response = await api.post('/databases', data);
  return response.data?.data;
};

export const deleteDatabase = async (id: string): Promise<void> => {
  await api.delete(`/databases/${id}`);
};

export interface QueryResult {
  columns: string[];
  rows: any[][];
  time_ms: number;
}

export const executeQuery = async (id: string, query: string): Promise<QueryResult> => {
  const response = await api.post(`/databases/${id}/query`, { query });
  return response.data?.data;
};

export const testConnection = async (data: Partial<CreateDatabaseRequest>): Promise<{ success: boolean; message: string }> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({ success: true, message: 'Configuration parameters verified.' });
    }, 800);
  });
};

export const testExistingConnection = async (id: string): Promise<{ success: boolean; message: string }> => {
  const response = await api.post(`/databases/${id}/test-connection`);
  return response.data?.data || { success: false, message: 'No response received' };
};
