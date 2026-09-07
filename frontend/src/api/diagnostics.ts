import api from './index';

export interface LockInfo {
  pid: number;
  query: string;
  state: string;
  blocked_by_pid?: number;
  blocking_query?: string;
  wait_event?: string;
  wait_event_type?: string;
}

export interface TableStorageInfo {
  table_name: string;
  total_bytes: number;
  index_bytes: number;
  live_tuples: number;
  dead_tuples: number;
  bloat_ratio: number;
}

export const getDatabaseLocks = async (id: string) => {
  const response = await api.get<{ data: LockInfo[] }>(`/databases/${id}/diagnostics/locks`);
  return response.data;
};

export const getDatabaseStorage = async (id: string) => {
  const response = await api.get<{ data: TableStorageInfo[] }>(`/databases/${id}/diagnostics/storage`);
  return response.data;
};
