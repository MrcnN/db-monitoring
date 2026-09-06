import { useQuery } from '@tanstack/react-query';
import { getIncidents } from '../api/incidents';
import { ExclamationTriangleIcon, CheckCircleIcon, ClockIcon } from '@heroicons/react/24/outline';
import { Link } from 'react-router-dom';

export default function IncidentsPage() {
  const { data: incidents = [], isLoading } = useQuery({
    queryKey: ['incidents'],
    queryFn: getIncidents,
    refetchInterval: 30000,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-3">
        <ExclamationTriangleIcon className="h-8 w-8 text-red-500" />
        <div>
          <h1 className="text-2xl font-semibold text-white">Incidents</h1>
          <p className="text-sm text-gray-400">Track and manage critical database failures and prolonged alerts.</p>
        </div>
      </div>

      <div className="bg-gray-900 shadow rounded-lg border border-gray-800 overflow-hidden">
        {isLoading ? (
          <div className="flex justify-center p-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
          </div>
        ) : incidents.length === 0 ? (
          <div className="text-center p-8 text-gray-400">
            <CheckCircleIcon className="h-12 w-12 text-green-500 mx-auto mb-3" />
            <p>No incidents found. All systems are operating normally.</p>
          </div>
        ) : (
          <table className="min-w-full divide-y divide-gray-800">
            <thead className="bg-gray-800/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase">Title & Details</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase">Database</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase">Started At</th>
              </tr>
            </thead>
            <tbody className="bg-gray-900 divide-y divide-gray-800">
              {incidents.map((incident) => (
                <tr key={incident.id} className="hover:bg-gray-800/50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    {incident.status === 'open' && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-900/50 text-red-400 border border-red-800">
                        <ExclamationTriangleIcon className="w-3 h-3 mr-1" /> Open
                      </span>
                    )}
                    {incident.status === 'acknowledged' && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-900/50 text-yellow-400 border border-yellow-800">
                        <ClockIcon className="w-3 h-3 mr-1" /> Acknowledged
                      </span>
                    )}
                    {incident.status === 'resolved' && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-900/50 text-green-400 border border-green-800">
                        <CheckCircleIcon className="w-3 h-3 mr-1" /> Resolved
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm font-medium text-gray-200">{incident.title}</div>
                    <div className="text-xs text-gray-500 mt-1 max-w-md break-words">{incident.description}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link to={`/databases/${incident.database_id}`} className="text-sm text-indigo-400 hover:text-indigo-300">
                      View Database
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                    {new Date(incident.started_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
