import api from '../lib/axios';
import { AuditLog, PaginatedResponse } from '../types';

export const getAuditLogs = async (): Promise<PaginatedResponse<AuditLog>> => {
  try {
    const response = await api.get('/audit-logs');
    const items: AuditLog[] = response.data?.data || [];
    const meta = response.data?.meta?.page || {
      total: items.length,
      limit: 50,
      offset: 0,
    };
    return {
      data: items,
      meta: {
        total: meta.total,
        limit: meta.limit,
        offset: meta.offset,
      },
    };
  } catch (err) {
    return {
      data: [],
      meta: { total: 0, limit: 10, offset: 0 },
    };
  }
};
