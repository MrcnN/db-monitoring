import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getDatabases, createDatabase, deleteDatabase } from '../api/databases';
import { CircleStackIcon, PlusIcon, ArrowTopRightOnSquareIcon, TrashIcon } from '@heroicons/react/24/outline';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import StatusBadge from '../components/common/StatusBadge';
import LoadingSpinner from '../components/common/LoadingSpinner';
import EmptyState from '../components/common/EmptyState';
import Modal from '../components/common/Modal';
import { CreateDatabaseRequest } from '../types';

export default function DatabasesPage() {
  const queryClient = useQueryClient();
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  const [formData, setFormData] = useState<CreateDatabaseRequest>({
    name: '',
    description: '',
    type: 'postgresql',
    host: 'postgres-target',
    port: 5432,
    database_name: 'demo_db',
    username: 'demo',
    password: 'demo_password',
    ssl_mode: 'disable',
    monitoring_interval: 15,
  });

  const { data: dbsResponse, isLoading } = useQuery({
    queryKey: ['databases'],
    queryFn: getDatabases,
    refetchInterval: 10000,
  });

  const createMutation = useMutation({
    mutationFn: createDatabase,
    onSuccess: () => {
      toast.success('Database added successfully');
      queryClient.invalidateQueries({ queryKey: ['databases'] });
      setIsAddModalOpen(false);
      setFormData({
        name: '',
        description: '',
        type: 'postgresql',
        host: 'postgres-target',
        port: 5432,
        database_name: 'demo_db',
        username: 'demo',
        password: '',
        ssl_mode: 'disable',
        monitoring_interval: 15,
      });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error?.message || 'Failed to add database';
      toast.error(msg);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteDatabase,
    onSuccess: () => {
      toast.success('Database removed');
      queryClient.invalidateQueries({ queryKey: ['databases'] });
    },
    onError: () => toast.error('Failed to delete database'),
  });

  const databases = dbsResponse?.data || [];

  const handleTypeChange = (type: 'postgresql' | 'mysql') => {
    setFormData((prev) => ({
      ...prev,
      type,
      port: type === 'postgresql' ? 5432 : 3306,
      host: type === 'postgresql' ? 'postgres-target' : 'mysql-target',
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name || !formData.host || !formData.username || !formData.password) {
      toast.error('Please fill in all required fields');
      return;
    }
    createMutation.mutate(formData);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Monitored Databases</h1>
          <p className="text-xs text-gray-400 mt-0.5">Manage and monitor your PostgreSQL and MySQL instances</p>
        </div>
        <button
          onClick={() => setIsAddModalOpen(true)}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-blue-600 hover:bg-blue-500 transition-colors"
        >
          <PlusIcon className="-ml-1 mr-2 h-4 w-4" />
          Add Database
        </button>
      </div>

      {isLoading ? (
        <LoadingSpinner />
      ) : databases.length === 0 ? (
        <EmptyState
          icon={<CircleStackIcon className="w-8 h-8" />}
          title="No databases registered"
          description="Get started by configuring a new PostgreSQL or MySQL instance to monitor."
          action={
            <button
              onClick={() => setIsAddModalOpen(true)}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-blue-600 hover:bg-blue-500"
            >
              <PlusIcon className="-ml-1 mr-2 h-4 w-4" />
              Add Database
            </button>
          }
        />
      ) : (
        <div className="bg-[#0A0A0A] border border-zinc-800/50 shadow rounded-xl overflow-hidden">
          <table className="min-w-full divide-y divide-zinc-800/50">
            <thead className="bg-zinc-900/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Instance</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Engine</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Endpoint</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Interval</th>
                <th className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800/50 bg-transparent">
              {databases.map((db) => (
                <tr key={db.id} className="hover:bg-zinc-900/30 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link to={`/databases/${db.id}`} className="font-semibold text-zinc-200 hover:text-white flex items-center space-x-1.5">
                      <span>{db.name}</span>
                      <ArrowTopRightOnSquareIcon className="h-3.5 w-3.5 text-zinc-500" />
                    </Link>
                    <span className="text-xs text-zinc-500">{db.database_name}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs">
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-zinc-800 text-zinc-300 uppercase border border-zinc-700">
                      {db.type}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs font-mono text-zinc-400">
                    {db.host}:{db.port}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <StatusBadge status={db.status} />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs text-zinc-500">
                    {db.monitoring_interval}s
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-xs font-medium space-x-3">
                    <Link
                      to={`/databases/${db.id}`}
                      className="inline-flex items-center px-2.5 py-1 rounded bg-zinc-100 text-black hover:bg-white transition-colors"
                    >
                      Inspect &amp; Telemetry
                    </Link>
                    <button
                      onClick={() => {
                        if (confirm(`Are you sure you want to remove ${db.name}?`)) {
                          deleteMutation.mutate(db.id);
                        }
                      }}
                      className="text-zinc-500 hover:text-red-400 p-1 transition-colors"
                      title="Delete"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={isAddModalOpen} onClose={() => setIsAddModalOpen(false)} title="Register Monitored Database">
        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Instance Name *</label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g. Production PostgreSQL"
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Engine *</label>
              <select
                value={formData.type}
                onChange={(e) => handleTypeChange(e.target.value as 'postgresql' | 'mysql')}
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              >
                <option value="postgresql">PostgreSQL</option>
                <option value="mysql">MySQL</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="block text-xs font-medium text-gray-300 mb-1">Host / Address *</label>
              <input
                type="text"
                required
                value={formData.host}
                onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                placeholder="postgres-target or 127.0.0.1"
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Port *</label>
              <input
                type="number"
                required
                value={formData.port}
                onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) || 5432 })}
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Database Name *</label>
              <input
                type="text"
                required
                value={formData.database_name}
                onChange={(e) => setFormData({ ...formData, database_name: e.target.value })}
                placeholder="demo_db"
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Username *</label>
              <input
                type="text"
                required
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                placeholder="demo"
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Password *</label>
              <input
                type="password"
                required
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                placeholder="••••••••"
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">SSL Mode</label>
              <select
                value={formData.ssl_mode}
                onChange={(e) => setFormData({ ...formData, ssl_mode: e.target.value })}
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              >
                <option value="disable">disable</option>
                <option value="require">require</option>
                <option value="verify-ca">verify-ca</option>
                <option value="verify-full">verify-full</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-300 mb-1">Polling Interval (seconds)</label>
              <input
                type="number"
                min={5}
                max={3600}
                value={formData.monitoring_interval}
                onChange={(e) => setFormData({ ...formData, monitoring_interval: parseInt(e.target.value) || 15 })}
                className="w-full px-3 py-1.5 text-xs bg-gray-700 border border-gray-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-3 border-t border-gray-700">
            <button
              type="button"
              onClick={() => setIsAddModalOpen(false)}
              className="px-4 py-2 border border-gray-600 rounded-md text-xs text-gray-300 hover:bg-gray-750"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-xs font-medium transition-colors"
            >
              {createMutation.isPending ? 'Saving...' : 'Register Database'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
