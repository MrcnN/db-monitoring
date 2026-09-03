import { ReactNode } from 'react';

interface EmptyStateProps {
  icon: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
}

export default function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="text-center p-12 bg-gray-800 rounded-lg border border-gray-700">
      <div className="mx-auto h-12 w-12 text-gray-400 flex justify-center items-center">
        {icon}
      </div>
      <h3 className="mt-2 text-sm font-semibold text-gray-200">{title}</h3>
      <p className="mt-1 text-sm text-gray-400 mb-6">{description}</p>
      {action && <div>{action}</div>}
    </div>
  );
}
