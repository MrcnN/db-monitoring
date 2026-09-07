import React from 'react';
import { TableStorageInfo } from '../../api/diagnostics';
import { Database, TrendingUp, AlertTriangle } from 'lucide-react';

interface Props {
  storageStats: TableStorageInfo[];
}

export const StorageAnalyzer: React.FC<Props> = ({ storageStats }) => {
  if (!storageStats || storageStats.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-12 bg-[#0A0A0A] border border-zinc-800 rounded-lg">
        <Database className="w-16 h-16 text-zinc-600 mb-4" />
        <h3 className="text-xl font-medium text-white mb-2">No Storage Data</h3>
        <p className="text-zinc-400 text-center">
          We couldn't retrieve storage information for this database.
        </p>
      </div>
    );
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const totalBytesAll = storageStats.reduce((sum, stat) => sum + stat.total_bytes, 0);

  return (
    <div className="space-y-6">
      <div className="bg-[#0A0A0A] border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-left text-sm text-zinc-300">
          <thead className="bg-zinc-900 text-zinc-400 border-b border-zinc-800">
            <tr>
              <th className="px-6 py-4 font-medium">Table Name</th>
              <th className="px-6 py-4 font-medium">Total Size</th>
              <th className="px-6 py-4 font-medium">Index Size</th>
              <th className="px-6 py-4 font-medium">Live Tuples</th>
              <th className="px-6 py-4 font-medium">Dead Tuples (Bloat)</th>
              <th className="px-6 py-4 font-medium">% Bloat</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800">
            {storageStats.map((stat) => {
              const bloatPct = (stat.bloat_ratio * 100).toFixed(1);
              const isHighBloat = stat.bloat_ratio > 0.2 && stat.dead_tuples > 1000;
              const isWarningBloat = stat.bloat_ratio > 0.1 && stat.dead_tuples > 500;
              
              return (
                <tr key={stat.table_name} className="hover:bg-zinc-900/50 transition-colors">
                  <td className="px-6 py-4 font-medium text-white">{stat.table_name}</td>
                  <td className="px-6 py-4 text-zinc-400">{formatBytes(stat.total_bytes)}</td>
                  <td className="px-6 py-4 text-zinc-400">{formatBytes(stat.index_bytes)}</td>
                  <td className="px-6 py-4 text-zinc-400">{stat.live_tuples.toLocaleString()}</td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center gap-1.5 ${
                      isHighBloat ? 'text-red-400' : isWarningBloat ? 'text-yellow-400' : 'text-zinc-400'
                    }`}>
                      {stat.dead_tuples.toLocaleString()}
                      {isHighBloat && <AlertTriangle className="w-4 h-4" />}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-full bg-zinc-800 rounded-full h-1.5 max-w-[100px]">
                        <div 
                          className={`h-1.5 rounded-full ${
                            isHighBloat ? 'bg-red-500' : isWarningBloat ? 'bg-yellow-500' : 'bg-emerald-500'
                          }`}
                          style={{ width: `${Math.min(Number(bloatPct), 100)}%` }}
                        ></div>
                      </div>
                      <span className={`text-xs ${
                        isHighBloat ? 'text-red-400 font-medium' : isWarningBloat ? 'text-yellow-400 font-medium' : 'text-zinc-500'
                      }`}>
                        {bloatPct}%
                      </span>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
