import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { CircleStackIcon, CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';
import { getDatabases } from '../api/databases';
import LoadingSpinner from '../components/common/LoadingSpinner';

export default function DashboardPage() {
  const { data: dbsResponse, isLoading } = useQuery({
    queryKey: ['databases'],
    queryFn: getDatabases,
  });

  const databases = dbsResponse?.data || [];
  const activeDBs = databases.filter((db) => db.status === 'active');
  const errorDBs = databases.filter((db) => db.status === 'error');

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-zinc-100 tracking-tight">Overview</h1>
        <Link
          to="/databases"
          className="inline-flex items-center rounded-lg bg-white px-3 py-2 text-sm font-semibold text-black hover:bg-zinc-200 transition-colors"
        >
          <CircleStackIcon className="h-4 w-4 mr-2" />
          View All Databases
        </Link>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div className="bg-[#0A0A0A] overflow-hidden rounded-xl border border-zinc-800/50 hover:border-zinc-700/50 transition-colors">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <CircleStackIcon className="h-5 w-5 text-zinc-400" aria-hidden="true" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="truncate text-xs font-medium text-zinc-500 uppercase tracking-wider">
                    Total Databases
                  </dt>
                  <dd>
                    <div className="text-2xl font-bold text-zinc-100">{databases.length}</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-[#0A0A0A] overflow-hidden rounded-xl border border-zinc-800/50 hover:border-green-900/50 transition-colors">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <CheckCircleIcon className="h-5 w-5 text-green-500" aria-hidden="true" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="truncate text-xs font-medium text-zinc-500 uppercase tracking-wider">
                    Healthy
                  </dt>
                  <dd>
                    <div className="text-2xl font-bold text-zinc-100">{activeDBs.length}</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-[#0A0A0A] overflow-hidden rounded-xl border border-zinc-800/50 hover:border-red-900/50 transition-colors">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <ExclamationCircleIcon className="h-5 w-5 text-red-500" aria-hidden="true" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="truncate text-xs font-medium text-zinc-500 uppercase tracking-wider">
                    Issues
                  </dt>
                  <dd>
                    <div className="text-2xl font-bold text-zinc-100">{errorDBs.length}</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-8 bg-[#0A0A0A] shadow rounded-xl border border-zinc-800/50">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-base font-semibold leading-6 text-zinc-100">System Status</h3>
          <div className="mt-2 max-w-xl text-sm text-zinc-400">
            <p>
              The platform is monitoring {databases.length} database instances across your infrastructure.
              All core services are operating normally.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
