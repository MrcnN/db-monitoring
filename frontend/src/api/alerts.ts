import api from '../lib/axios';

export interface Alert {
  id: string;
  database_id: string;
  title: string;
  description: string;
  severity: 'warning' | 'critical';
  status: 'active' | 'resolved';
  metric_name?: string;
  current_value?: string;
  resolved_at?: string;
  created_at: string;
}

export const getAlerts = async (dbId?: string): Promise<Alert[]> => {
  const url = dbId ? `/databases/${dbId}/alerts` : '/alerts';
  const response = await api.get(url);
  return response.data?.data || [];
};
