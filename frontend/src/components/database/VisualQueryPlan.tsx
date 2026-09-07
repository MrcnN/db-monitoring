import React from 'react';
import { GitCommit, AlignLeft, Layers, Zap } from 'lucide-react';

interface Props {
  planData: any; // Raw EXPLAIN JSON
}

export const VisualQueryPlan: React.FC<Props> = ({ planData }) => {
  if (!planData || planData.length === 0) {
    return (
      <div className="text-zinc-500 text-sm italic">
        Visual plan data not available.
      </div>
    );
  }

  // Postgres EXPLAIN format usually is an array with a "Plan" object inside
  const plan = Array.isArray(planData) && planData[0] && planData[0].Plan ? planData[0].Plan : planData;

  const renderNode = (node: any, depth: number = 0) => {
    if (!node) return null;

    const isSeqScan = node['Node Type'] === 'Seq Scan';
    const isIndexScan = node['Node Type'] === 'Index Scan' || node['Node Type'] === 'Index Only Scan';

    return (
      <div key={Math.random().toString()} className="flex flex-col relative" style={{ marginLeft: depth > 0 ? '24px' : '0px' }}>
        {depth > 0 && (
          <div className="absolute -left-6 top-6 w-6 border-b-2 border-l-2 border-zinc-800 rounded-bl h-full" style={{ height: 'calc(100% - 24px)' }} />
        )}
        
        <div className={`mt-3 relative z-10 flex flex-col p-4 rounded-lg border ${
          isSeqScan ? 'border-yellow-500/30 bg-yellow-500/5' :
          isIndexScan ? 'border-emerald-500/30 bg-emerald-500/5' :
          'border-zinc-800 bg-zinc-900/50'
        } min-w-[280px] max-w-lg`}>
          
          <div className="flex items-center gap-2 mb-2 border-b border-zinc-800/50 pb-2">
            {isSeqScan ? <Layers className="w-4 h-4 text-yellow-500" /> : 
             isIndexScan ? <Zap className="w-4 h-4 text-emerald-500" /> :
             <GitCommit className="w-4 h-4 text-zinc-400" />}
            <span className={`font-semibold text-sm ${
              isSeqScan ? 'text-yellow-500' : isIndexScan ? 'text-emerald-500' : 'text-white'
            }`}>
              {node['Node Type']}
            </span>
          </div>
          
          <div className="space-y-1.5 text-xs">
            {node['Relation Name'] && (
              <div className="flex justify-between">
                <span className="text-zinc-500">Relation:</span>
                <span className="text-zinc-300 font-mono">{node['Relation Name']}</span>
              </div>
            )}
            {node['Index Name'] && (
              <div className="flex justify-between">
                <span className="text-zinc-500">Index:</span>
                <span className="text-zinc-300 font-mono">{node['Index Name']}</span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-zinc-500">Cost:</span>
              <span className="text-zinc-400">{node['Startup Cost']}..{node['Total Cost']}</span>
            </div>
            {node['Plan Rows'] !== undefined && (
              <div className="flex justify-between">
                <span className="text-zinc-500">Rows:</span>
                <span className="text-zinc-400">{node['Plan Rows']}</span>
              </div>
            )}
            {node['Filter'] && (
              <div className="mt-2 pt-2 border-t border-zinc-800/50 text-zinc-500 line-clamp-2" title={node['Filter']}>
                <span className="mr-1">Filter:</span>
                <span className="text-zinc-400 font-mono">{node['Filter']}</span>
              </div>
            )}
          </div>
        </div>

        {node.Plans && node.Plans.length > 0 && (
          <div className="flex flex-col relative">
            <div className="absolute left-6 top-0 bottom-0 w-px bg-zinc-800" />
            {node.Plans.map((child: any, idx: number) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="p-4 bg-black border border-zinc-800 rounded-lg overflow-x-auto min-h-[300px]">
      <div className="flex items-center gap-2 mb-4">
        <AlignLeft className="w-5 h-5 text-zinc-400" />
        <h3 className="text-white font-medium">Query Execution Plan (Visualized)</h3>
      </div>
      <div className="pl-2">
        {renderNode(plan)}
      </div>
    </div>
  );
};
