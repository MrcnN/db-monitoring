import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';

export default function MainLayout() {
  return (
    <div className="flex h-screen bg-[#000000] text-zinc-200 selection:bg-zinc-800">
      <Sidebar />
      <div className="flex-1 overflow-auto">
        <main className="h-full p-8 max-w-7xl mx-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
