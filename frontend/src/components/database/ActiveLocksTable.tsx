import React from 'react';
import { LockInfo } from '../../api/diagnostics';
import { ShieldAlert, CheckCircle, Clock } from 'lucide-react';

interface Props {
  locks: LockInfo[];
}

export const ActiveLocksTable: React.FC<Props> = ({ locks }) => {
  if (!locks || locks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-12 bg-[#0A0A0A] border border-zinc-800 rounded-lg">
        <CheckCircle className="w-16 h-16 text-emerald-500 mb-4" />
        <h3 className="text-xl font-medium text-white mb-2">No Active Locks</h3>
        <p className="text-zinc-400 text-center max-w-md">
          Your database is currently running smoothly. No queries are blocking other transactions.
        </p>
      </div>
    );
  }

  // Count blocked queries
  const blockedCount = locks.filter((l) => l.blocked_by_pid).length;

  return (
    <div className="space-y-6">
      {blockedCount > 0 && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 flex items-start gap-4">
          <ShieldAlert className="w-6 h-6 text-red-500 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="text-red-500 font-medium text-lg">Lock Contention Detected</h3>
            <p className="text-red-400 text-sm mt-1">
              There are {blockedCount} queries currently waiting on other transactions to finish. This can severely impact performance.
            </p>
          </div>
        </div>
      )}

      <div className="bg-[#0A0A0A] border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-left text-sm text-zinc-300">
          <thead className="bg-zinc-900 text-zinc-400 border-b border-zinc-800">
            <tr>
              <th className="px-6 py-4 font-medium">PID</th>
              <th className="px-6 py-4 font-medium">State</th>
              <th className="px-6 py-4 font-medium">Blocked By</th>
              <th className="px-6 py-4 font-medium">Query</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800">
            {locks.map((lock) => (
              <tr key={lock.pid} className="hover:bg-zinc-900/50 transition-colors">
                <td className="px-6 py-4 font-mono text-zinc-400">{lock.pid}</td>
                <td className="px-6 py-4">
                  <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${
                    lock.state === 'active' 
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                      : 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20'
                  }`}>
                    {lock.state}
                  </span>
                  {lock.wait_event && (
                    <div className="flex items-center gap-1.5 mt-2 text-xs text-zinc-500">
                      <Clock className="w-3.5 h-3.5" />
                      {lock.wait_event_type}: {lock.wait_event}
                    </div>
                  )}
                </td>
                <td className="px-6 py-4">
                  {lock.blocked_by_pid ? (
                    <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      PID: {lock.blocked_by_pid}
                    </span>
                  ) : (
                    <span className="text-zinc-600">-</span>
                  )}
                </td>
                <td className="px-6 py-4">
                  <div className="font-mono text-xs bg-black p-3 rounded border border-zinc-800 overflow-x-auto whitespace-pre-wrap max-w-2xl">
                    {lock.query || '<insufficient privileges or empty>'}
                  </div>
                  {lock.blocking_query && (
                    <div className="mt-3 pl-4 border-l-2 border-red-500/30">
                      <p className="text-xs text-red-400 mb-1 font-medium">Waiting for query:</p>
                      <div className="font-mono text-xs text-zinc-500 bg-black p-2 rounded border border-zinc-800/50 overflow-x-auto whitespace-pre-wrap max-w-2xl">
                        {lock.blocking_query}
                      </div>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
