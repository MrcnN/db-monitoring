import api from '../lib/axios';

export interface AdvisorResult {
  plan: any;
  recommendations: string[];
}

export const explainQuery = async (dbId: string, query: string): Promise<AdvisorResult> => {
  const response = await api.post(`/databases/${dbId}/advisor/explain`, { query });
  return response.data?.data;
};
