import api from '../lib/axios';

export interface Incident {
  id: string;
  database_id: string;
  title: string;
  description: string;
  status: 'open' | 'acknowledged' | 'resolved';
  started_at: string;
  acknowledged_at?: string;
  resolved_at?: string;
}

export interface NotificationChannel {
  id: string;
  name: string;
  type: 'slack' | 'webhook' | 'email';
  config: Record<string, any>;
  is_active: boolean;
  created_at: string;
}

export const getIncidents = async (): Promise<Incident[]> => {
  const response = await api.get('/incidents');
  return response.data?.data || [];
};

export const getChannels = async (): Promise<NotificationChannel[]> => {
  const response = await api.get('/notification-channels');
  return response.data?.data || [];
};

export const createChannel = async (data: Partial<NotificationChannel>): Promise<NotificationChannel> => {
  const response = await api.post('/notification-channels', data);
  return response.data?.data;
};

export const deleteChannel = async (id: string): Promise<void> => {
  await api.delete(`/notification-channels/${id}`);
};
