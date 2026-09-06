import { Link, useLocation } from 'react-router-dom';
import { 
  HomeIcon, 
  CircleStackIcon, 
  BellAlertIcon, 
  ExclamationTriangleIcon,
  DocumentTextIcon,
  Cog6ToothIcon,
  ArrowLeftOnRectangleIcon,
  CommandLineIcon
} from '@heroicons/react/24/outline';
import { useAuth } from '../../hooks/useAuth';

const navigation = [
  { name: 'Overview', href: '/dashboard', icon: HomeIcon },
  { name: 'Databases', href: '/databases', icon: CircleStackIcon },
  { name: 'Alerts', href: '/alerts', icon: BellAlertIcon },
  { name: 'Incidents', href: '/incidents', icon: ExclamationTriangleIcon },
  { name: 'Audit Logs', href: '/audit-logs', icon: DocumentTextIcon },
  { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
];

export default function Sidebar() {
  const location = useLocation();
  const { user, logout } = useAuth();

  return (
    <div className="flex w-64 flex-col bg-[#0A0A0A] border-r border-zinc-800/50">
      <div className="flex h-16 shrink-0 items-center px-6 border-b border-zinc-800/50">
        <div className="flex items-center space-x-2">
          <div className="w-6 h-6 bg-zinc-100 rounded flex items-center justify-center">
            <CommandLineIcon className="w-4 h-4 text-black" />
          </div>
          <span className="text-lg font-bold text-zinc-100 tracking-tight">DBPlatform</span>
        </div>
      </div>
      <div className="flex flex-1 flex-col overflow-y-auto">
        <nav className="flex-1 space-y-0.5 px-3 py-4">
          <div className="text-[10px] font-semibold text-zinc-500 uppercase tracking-wider mb-2 px-3">Main</div>
          {navigation.map((item) => {
            const isActive = location.pathname.startsWith(item.href);
            return (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors
                  ${isActive 
                    ? 'bg-zinc-800/50 text-zinc-100' 
                    : 'text-zinc-400 hover:bg-zinc-800/30 hover:text-zinc-200'}
                `}
              >
                <item.icon
                  className={`
                    mr-3 h-4 w-4 flex-shrink-0 transition-colors
                    ${isActive ? 'text-zinc-100' : 'text-zinc-500 group-hover:text-zinc-400'}
                  `}
                  aria-hidden="true"
                />
                {item.name}
              </Link>
            );
          })}
        </nav>
      </div>
      <div className="p-4 border-t border-zinc-800/50">
        <div className="flex items-center bg-zinc-900/50 p-2 rounded-xl border border-zinc-800/50">
          <div className="w-8 h-8 rounded-full bg-zinc-800 flex items-center justify-center text-xs font-bold text-zinc-300">
            {user?.full_name.charAt(0).toUpperCase()}
          </div>
          <div className="ml-3 truncate flex-1">
            <p className="text-sm font-medium text-zinc-200 truncate">{user?.full_name}</p>
            <p className="text-[10px] uppercase tracking-wider font-semibold text-zinc-500">{user?.role}</p>
          </div>
          <button
            onClick={logout}
            className="ml-auto flex-shrink-0 p-1.5 text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800 rounded-md transition-colors"
            title="Logout"
          >
            <ArrowLeftOnRectangleIcon className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
}
