import { Link } from 'react-router-dom';
import { useAppSelector } from '../hooks/redux';

const HomePage = () => {
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  return (
    <div className="bg-gray-50 min-h-screen dark:bg-gray-900 transition-colors duration-200">
      {/* Hero Section */}
      <div className="bg-indigo-700 text-white py-16 dark:bg-indigo-900">
        <div className="container mx-auto px-4 text-center">
          <h1 className="text-4xl font-bold mb-4">Ticket Management System</h1>
          <p className="text-xl mb-8">Streamline your support process with our easy-to-use ticket system</p>

          {isAuthenticated ? (
            <div className="flex justify-center gap-4">
              <Link
                to="/create"
                className="bg-white text-indigo-700 px-6 py-3 rounded-md font-medium hover:bg-gray-100 transition-colors dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
              >
                Create New Ticket
              </Link>
              <Link
                to="/tickets"
                className="bg-indigo-600 text-white px-6 py-3 rounded-md font-medium hover:bg-indigo-500 transition-colors dark:bg-indigo-700 dark:hover:bg-indigo-600"
              >
                View Tickets
              </Link>
            </div>
          ) : (
            <div className="flex justify-center gap-4">
              <Link
                to="/login"
                className="bg-white text-indigo-700 px-6 py-3 rounded-md font-medium hover:bg-gray-100 transition-colors dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
              >
                Login
              </Link>
              <Link
                to="/signup"
                className="bg-indigo-600 text-white px-6 py-3 rounded-md font-medium hover:bg-indigo-500 transition-colors dark:bg-indigo-700 dark:hover:bg-indigo-600"
              >
                Sign Up
              </Link>
            </div>
          )}
        </div>
      </div>

      {/* Features Section */}
      <div className="py-16">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-12 dark:text-white">Key Features</h2>

          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-white p-6 rounded-lg shadow-md dark:bg-gray-800 transition-colors duration-200">
              <div className="text-indigo-600 mb-4 dark:text-indigo-400">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2 dark:text-white">Ticket Management</h3>
              <p className="text-gray-600 dark:text-gray-300">Create, track, and manage support tickets in one centralized location.</p>
            </div>

            <div className="bg-white p-6 rounded-lg shadow-md dark:bg-gray-800 transition-colors duration-200">
              <div className="text-indigo-600 mb-4 dark:text-indigo-400">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2 dark:text-white">Communication</h3>
              <p className="text-gray-600 dark:text-gray-300">Seamless communication between support staff and users through ticket comments.</p>
            </div>

            <div className="bg-white p-6 rounded-lg shadow-md dark:bg-gray-800 transition-colors duration-200">
              <div className="text-indigo-600 mb-4 dark:text-indigo-400">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold mb-2 dark:text-white">Priority Management</h3>
              <p className="text-gray-600 dark:text-gray-300">Assign priority levels to tickets to ensure critical issues are addressed first.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage;
