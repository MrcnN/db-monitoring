import React from 'react';
import { SlowQuery } from '../../types';

interface SlowQueryTableProps {
  queries: SlowQuery[];
  isLoading: boolean;
}

export const SlowQueryTable: React.FC<SlowQueryTableProps> = ({ queries, isLoading }) => {
  if (isLoading) {
    return (
      <div className="flex justify-center p-8 bg-gray-900 rounded-lg border border-gray-800">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  if (queries.length === 0) {
    return (
      <div className="text-center p-8 text-gray-400 bg-gray-900 rounded-lg border border-gray-800">
        No slow queries detected (or pg_stat_statements is not enabled).
      </div>
    );
  }

  return (
    <div className="overflow-x-auto bg-gray-900 rounded-lg shadow border border-gray-800">
      <table className="min-w-full divide-y divide-gray-800">
        <thead className="bg-gray-800/50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Query
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Calls
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Total Time (ms)
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Mean Time (ms)
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
              Rows
            </th>
          </tr>
        </thead>
        <tbody className="bg-gray-900 divide-y divide-gray-800">
          {queries.map((q, idx) => (
            <tr key={idx} className="hover:bg-gray-800/50 transition-colors">
              <td className="px-6 py-4 whitespace-normal text-sm font-mono text-gray-300 max-w-lg truncate">
                {q.query}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                {q.calls.toLocaleString()}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                {q.total_time_ms.toFixed(2)}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                {q.mean_time_ms.toFixed(2)}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                {q.rows.toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
