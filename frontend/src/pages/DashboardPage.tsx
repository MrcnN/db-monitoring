import { useQuery } from '@tanstack/react-query';
import { getDatabases } from '../api/databases';
import StatusBadge from '../components/common/StatusBadge';
import LoadingSpinner from '../components/common/LoadingSpinner';

export default function DashboardPage() {
  const { data: dbsResponse, isLoading } = useQuery({
    queryKey: ['databases'],
    queryFn: getDatabases,
  });

  const databases = dbsResponse?.data || [];
  const healthyCount = databases.filter(db => db.status === 'active').length;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold text-white">Dashboard</h1>
      
      {isLoading ? (
        <LoadingSpinner />
      ) : (
        <>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            <div className="bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-700">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-400 truncate">Total Databases</p>
                    <p className="mt-1 text-3xl font-semibold text-white">{databases.length}</p>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-700">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-400 truncate">Healthy Databases</p>
                    <p className="mt-1 text-3xl font-semibold text-green-400">{healthyCount}</p>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-700">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-400 truncate">Active Alerts</p>
                    <p className="mt-1 text-3xl font-semibold text-yellow-400">0</p>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-700">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-400 truncate">Active Incidents</p>
                    <p className="mt-1 text-3xl font-semibold text-red-400">0</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <h2 className="text-lg font-medium text-white mt-8 mb-4">Recent Databases</h2>
          <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
            <ul className="divide-y divide-gray-700">
              {databases.map((db) => (
                <li key={db.id} className="p-4 hover:bg-gray-750">
                  <div className="flex items-center justify-between">
                    <div className="flex flex-col">
                      <span className="text-sm font-medium text-white">{db.name}</span>
                      <span className="text-xs text-gray-400">{db.host}:{db.port}</span>
                    </div>
                    <div>
                      <StatusBadge status={db.status} />
                    </div>
                  </div>
                </li>
              ))}
              {databases.length === 0 && (
                <li className="p-4 text-sm text-gray-400 text-center">No databases configured yet.</li>
              )}
            </ul>
          </div>
        </>
      )}
    </div>
  );
}
