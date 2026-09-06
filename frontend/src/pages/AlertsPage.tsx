import { useQuery } from '@tanstack/react-query';
import { getAlerts } from '../api/alerts';
import { AlertsTable } from '../components/database/AlertsTable';
import { BellAlertIcon } from '@heroicons/react/24/outline';

export default function AlertsPage() {
  const { data: alerts = [], isLoading } = useQuery({
    queryKey: ['globalAlerts'],
    queryFn: () => getAlerts(),
    refetchInterval: 30000, // refresh every 30 seconds
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-3">
        <BellAlertIcon className="h-8 w-8 text-indigo-400" />
        <div>
          <h1 className="text-2xl font-semibold text-white">System Alerts</h1>
          <p className="text-sm text-gray-400">Global view of all triggered database alerts.</p>
        </div>
      </div>
      
      <AlertsTable alerts={alerts} isLoading={isLoading} />
    </div>
  );
}
