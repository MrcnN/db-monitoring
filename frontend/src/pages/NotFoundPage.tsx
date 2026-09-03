import { Link } from 'react-router-dom';

export default function NotFoundPage() {
  return (
    <div className="min-h-screen bg-gray-900 flex flex-col justify-center items-center">
      <h1 className="text-6xl font-bold text-white mb-4">404</h1>
      <p className="text-xl text-gray-400 mb-8">Page not found</p>
      <Link to="/" className="text-primary-500 hover:text-primary-400">
        Go back home
      </Link>
    </div>
  );
}
