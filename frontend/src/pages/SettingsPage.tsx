import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getChannels, createChannel, deleteChannel, NotificationChannel } from '../api/incidents';
import { TrashIcon, PlusIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [type, setType] = useState<'slack' | 'webhook'>('slack');
  const [url, setUrl] = useState('');

  const { data: channels = [], isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: getChannels,
  });

  const createMutation = useMutation({
    mutationFn: createChannel,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      toast.success('Notification channel created');
      setName('');
      setUrl('');
    },
    onError: () => toast.error('Failed to create channel'),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteChannel,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      toast.success('Channel deleted');
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate({
      name,
      type,
      config: { url },
      is_active: true,
    });
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <h1 className="text-2xl font-semibold text-white">Settings</h1>

      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 p-6">
        <h3 className="text-lg font-medium text-white mb-4">Notification Channels</h3>
        <p className="text-gray-400 text-sm mb-6">Configure Webhooks and Slack integration for database incident alerts.</p>
        
        <form onSubmit={handleCreate} className="flex gap-4 items-end mb-8 bg-gray-900 p-4 rounded border border-gray-700">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-400 mb-1">Name</label>
            <input required value={name} onChange={e => setName(e.target.value)} type="text" className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white text-sm" placeholder="Ops Team Slack" />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 mb-1">Type</label>
            <select value={type} onChange={e => setType(e.target.value as any)} className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white text-sm">
              <option value="slack">Slack</option>
              <option value="webhook">Webhook</option>
            </select>
          </div>
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-400 mb-1">Webhook URL</label>
            <input required value={url} onChange={e => setUrl(e.target.value)} type="url" className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-white text-sm" placeholder="https://hooks.slack.com/services/..." />
          </div>
          <button type="submit" disabled={createMutation.isPending} className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded text-sm font-medium flex items-center">
            <PlusIcon className="w-4 h-4 mr-1" /> Add
          </button>
        </form>

        {isLoading ? (
          <div className="animate-pulse flex space-x-4">
            <div className="flex-1 space-y-4 py-1">
              <div className="h-4 bg-gray-700 rounded w-3/4"></div>
              <div className="space-y-2">
                <div className="h-4 bg-gray-700 rounded"></div>
                <div className="h-4 bg-gray-700 rounded w-5/6"></div>
              </div>
            </div>
          </div>
        ) : channels.length === 0 ? (
          <div className="text-center text-sm text-gray-400 py-4">No channels configured.</div>
        ) : (
          <div className="space-y-3">
            {channels.map(ch => (
              <div key={ch.id} className="flex items-center justify-between bg-gray-900 border border-gray-700 p-4 rounded">
                <div>
                  <div className="text-sm font-medium text-gray-200">{ch.name}</div>
                  <div className="text-xs text-gray-500 mt-1 uppercase">{ch.type}</div>
                </div>
                <button
                  onClick={() => deleteMutation.mutate(ch.id)}
                  className="text-red-400 hover:text-red-300 p-2"
                  title="Delete Channel"
                >
                  <TrashIcon className="w-5 h-5" />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
