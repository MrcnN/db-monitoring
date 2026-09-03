import { useQuery } from '@tanstack/react-query';
import { getAuditLogs } from '../api/auditLogs';
import StatusBadge from '../components/common/StatusBadge';
import LoadingSpinner from '../components/common/LoadingSpinner';

export default function AuditLogsPage() {
  const { data: logsResponse, isLoading } = useQuery({
    queryKey: ['auditLogs'],
    queryFn: getAuditLogs,
  });

  const logs = logsResponse?.data || [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold text-white">Audit Logs</h1>
      
      {isLoading ? (
        <LoadingSpinner />
      ) : (
        <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-900/50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Timestamp</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">User</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Action</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Resource</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {logs.map((log) => (
                <tr key={log.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">{new Date(log.created_at).toLocaleString()}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{log.user_email}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-white">{log.action}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">{log.resource_type}: {log.resource_name}</td>
                  <td className="px-6 py-4 whitespace-nowrap"><StatusBadge status={log.status} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
