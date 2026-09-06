import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { executeQuery, QueryResult } from '../../api/databases';
import { PlayIcon, CommandLineIcon, ExclamationCircleIcon, BoltIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

interface SqlConsoleProps {
  databaseId: string;
}

export function SqlConsole({ databaseId }: SqlConsoleProps) {
  const [query, setQuery] = useState('');

  const mutation = useMutation({
    mutationFn: () => executeQuery(databaseId, query),
    onSuccess: () => toast.success('Query executed successfully'),
    onError: (err: any) => toast.error(err?.response?.data?.error?.message || 'Query failed')
  });

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      if (query.trim()) mutation.mutate();
    }
  };

  const res = mutation.data as QueryResult | undefined;

  return (
    <div className="flex flex-col h-[700px] border border-zinc-800 rounded-xl overflow-hidden bg-[#0A0A0A] shadow-2xl">
      <div className="flex-1 flex flex-col border-b border-zinc-800/50">
        <div className="flex items-center justify-between px-4 py-2 bg-zinc-900/40 border-b border-zinc-800/50">
          <div className="flex items-center text-zinc-400 text-xs font-medium">
            <CommandLineIcon className="w-4 h-4 mr-2" />
            SQL Console
          </div>
          <div className="text-xs text-zinc-500">Press <kbd className="font-mono bg-zinc-800 px-1 rounded text-zinc-300">Cmd/Ctrl + Enter</kbd> to run</div>
        </div>
        <textarea
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="SELECT * FROM users LIMIT 10;"
          className="flex-1 w-full p-4 bg-transparent text-zinc-200 font-mono text-sm resize-none focus:outline-none focus:ring-0 selection:bg-indigo-500/30"
          spellCheck={false}
        />
      </div>

      <div className="h-64 flex flex-col bg-zinc-950/50">
        <div className="flex items-center justify-between px-4 py-2 border-b border-zinc-800/50 bg-zinc-900/40">
          <div className="flex items-center gap-4">
            <button
              onClick={() => mutation.mutate()}
              disabled={mutation.isPending || !query.trim()}
              className="inline-flex items-center px-3 py-1.5 text-xs font-medium bg-zinc-100 text-black hover:bg-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {mutation.isPending ? (
                <div className="w-3 h-3 border-2 border-black border-t-transparent rounded-full animate-spin mr-1.5" />
              ) : (
                <PlayIcon className="w-3.5 h-3.5 mr-1.5 fill-current" />
              )}
              Run Query
            </button>
            {res && (
              <div className="flex items-center text-xs text-zinc-500 space-x-3">
                <span className="flex items-center"><BoltIcon className="w-3 h-3 mr-1" /> {res.time_ms} ms</span>
                <span>{res.rows?.length || 0} rows</span>
              </div>
            )}
          </div>
        </div>
        
        <div className="flex-1 overflow-auto p-0">
          {mutation.isPending ? (
            <div className="flex items-center justify-center h-full text-zinc-500 text-sm font-medium">
              <div className="w-4 h-4 border-2 border-zinc-500 border-t-transparent rounded-full animate-spin mr-2" />
              Executing...
            </div>
          ) : mutation.isError ? (
            <div className="p-4 text-red-400 text-sm font-mono flex items-start">
              <ExclamationCircleIcon className="w-5 h-5 mr-2 flex-shrink-0" />
              <div className="whitespace-pre-wrap break-all">
                {(mutation.error as any)?.response?.data?.error?.message || mutation.error?.message}
              </div>
            </div>
          ) : !res ? (
            <div className="flex items-center justify-center h-full text-zinc-600 text-sm">
              No results. Run a query to see output here.
            </div>
          ) : res.rows?.length === 0 ? (
            <div className="p-4 text-zinc-500 text-sm font-mono">
              Query returned 0 rows.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-zinc-800 text-sm">
              <thead className="bg-zinc-900/80 sticky top-0">
                <tr>
                  {res.columns.map((col, i) => (
                    <th key={i} className="px-4 py-2 text-left font-semibold text-zinc-400 whitespace-nowrap border-r border-zinc-800 last:border-0">
                      {col}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/50 bg-transparent font-mono text-xs">
                {res.rows.map((row, i) => (
                  <tr key={i} className="hover:bg-zinc-900/50 transition-colors">
                    {row.map((val, j) => (
                      <td key={j} className="px-4 py-2 text-zinc-300 whitespace-nowrap border-r border-zinc-800/50 last:border-0 max-w-sm truncate">
                        {val === null ? <span className="text-zinc-600 italic">null</span> : String(val)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
