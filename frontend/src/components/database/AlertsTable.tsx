import React from 'react';
import { Alert } from '../../api/alerts';
import { ExclamationTriangleIcon, XCircleIcon, CheckCircleIcon } from '@heroicons/react/24/outline';

interface AlertsTableProps {
  alerts: Alert[];
  isLoading: boolean;
}

export const AlertsTable: React.FC<AlertsTableProps> = ({ alerts, isLoading }) => {
  if (isLoading) {
    return (
      <div className="flex justify-center p-8 bg-gray-900 rounded-lg border border-gray-800">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  if (alerts.length === 0) {
    return (
      <div className="text-center p-8 text-gray-400 bg-gray-900 rounded-lg border border-gray-800 flex flex-col items-center">
        <CheckCircleIcon className="h-12 w-12 text-green-500 mb-3" />
        <p>No active or historical alerts for this database. Everything is healthy!</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto bg-gray-900 rounded-lg shadow border border-gray-800">
      <table className="min-w-full divide-y divide-gray-800">
        <thead className="bg-gray-800/50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Status & Severity
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Issue Title
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Value
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Time
            </th>
          </tr>
        </thead>
        <tbody className="bg-gray-900 divide-y divide-gray-800">
          {alerts.map((alert) => (
            <tr key={alert.id} className="hover:bg-gray-800/50 transition-colors">
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="flex items-center space-x-2">
                  {alert.status === 'active' ? (
                    <span className="flex h-2 w-2 relative">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
                    </span>
                  ) : (
                    <span className="h-2 w-2 rounded-full bg-gray-500"></span>
                  )}
                  {alert.severity === 'critical' ? (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-900/60 text-red-300 border border-red-800">
                      <XCircleIcon className="w-3 h-3 mr-1" />
                      Critical
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-900/60 text-yellow-300 border border-yellow-800">
                      <ExclamationTriangleIcon className="w-3 h-3 mr-1" />
                      Warning
                    </span>
                  )}
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="text-sm font-medium text-gray-200">{alert.title}</div>
                <div className="text-xs text-gray-500 mt-1 max-w-md break-words">{alert.description}</div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <span className="text-sm font-mono text-gray-300 bg-gray-800 px-2 py-1 rounded">
                  {alert.current_value || 'N/A'}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-400">
                <div>Created: {new Date(alert.created_at).toLocaleString()}</div>
                {alert.status === 'resolved' && alert.resolved_at && (
                  <div className="text-green-400 mt-1">
                    Resolved: {new Date(alert.resolved_at).toLocaleString()}
                  </div>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
