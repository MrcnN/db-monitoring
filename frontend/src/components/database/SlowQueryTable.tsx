import React from 'react';
import { useState } from 'react';
import { SlowQuery } from '../../types';
import QueryAdvisorModal from './QueryAdvisorModal';
import { SparklesIcon } from '@heroicons/react/24/outline';

interface SlowQueryTableProps {
  queries: SlowQuery[];
  isLoading: boolean;
  databaseId?: string;
}

export const SlowQueryTable: React.FC<SlowQueryTableProps> = ({ queries, isLoading, databaseId }) => {
  const [selectedQuery, setSelectedQuery] = useState<string | null>(null);

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
            <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">
              Action
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
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                {databaseId && (
                  <button
                    onClick={() => setSelectedQuery(q.query)}
                    className="inline-flex items-center text-indigo-400 hover:text-indigo-300 bg-indigo-900/20 px-3 py-1 rounded border border-indigo-800/50"
                  >
                    <SparklesIcon className="w-4 h-4 mr-1" />
                    Analyze
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {databaseId && selectedQuery && (
        <QueryAdvisorModal
          isOpen={!!selectedQuery}
          onClose={() => setSelectedQuery(null)}
          databaseId={databaseId}
          query={selectedQuery}
        />
      )}
    </div>
  );
};
