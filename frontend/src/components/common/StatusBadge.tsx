interface StatusBadgeProps {
  status: 'active' | 'inactive' | 'error' | 'healthy' | 'warning' | 'critical' | 'unknown' | string;
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const normalizedStatus = status.toLowerCase();
  
  let colorClass = 'bg-gray-500/10 text-gray-400 border-gray-500/20';
  
  if (['active', 'healthy', 'success'].includes(normalizedStatus)) {
    colorClass = 'bg-green-500/10 text-green-400 border-green-500/20';
  } else if (['error', 'critical', 'failed'].includes(normalizedStatus)) {
    colorClass = 'bg-red-500/10 text-red-400 border-red-500/20';
  } else if (['warning'].includes(normalizedStatus)) {
    colorClass = 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20';
  }

  return (
    <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium border ${colorClass}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}
