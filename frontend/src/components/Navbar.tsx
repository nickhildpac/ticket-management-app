import { Link } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../hooks/redux";
import { logoutAsync } from "../store/slices/authSlice";
import { useDarkMode } from "../hooks/useDarkMode";

const Navbar = () => {
  const dispatch = useAppDispatch();
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const { isDarkMode, toggleDarkMode } = useDarkMode();

  return (
    <div className="flex justify-between items-center bg-gray-800 px-6 py-4 transition-colors duration-200 dark:bg-gray-900">
      <div className="flex items-center">
        <Link to="/" className="flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-indigo-500" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-6-3a2 2 0 11-4 0 2 2 0 014 0zm-2 4a5 5 0 00-4.546 2.916A5.986 5.986 0 005 10a6 6 0 0012 0c0-.35-.035-.691-.1-1.02A4.96 4.96 0 0010 11z" clipRule="evenodd" />
          </svg>
          <span className="ml-2 text-xl font-bold text-white">Ticket Management</span>
        </Link>
      </div>

      <nav>
        <ul className="flex gap-6 items-center">
          <li>
            <Link to="/" className="text-white hover:text-indigo-300 transition-colors">Home</Link>
          </li>
          <li>
            <Link to="/create" className="text-white hover:text-indigo-300 transition-colors">Create Ticket</Link>
          </li>
          <li>
            <Link to="/tickets" className="text-white hover:text-indigo-300 transition-colors">Manage Tickets</Link>
          </li>
          <li>
            <Link to="/about" className="text-white hover:text-indigo-300 transition-colors">About</Link>
          </li>
          <li>
            <button
              onClick={toggleDarkMode}
              className="p-2 rounded-full text-white hover:bg-gray-700 transition-colors focus:outline-none"
              aria-label="Toggle dark mode"
            >
              {isDarkMode ? (
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              ) : (
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                </svg>
              )}
            </button>
          </li>
          {isAuthenticated ? (
            <li>
              <button
                onClick={() => dispatch(logoutAsync())}
                className="ml-4 bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md transition-colors"
              >
                Logout
              </button>
            </li>
          ) : (
            <li>
              <Link
                to="/login"
                className="ml-4 bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md transition-colors"
              >
                Login
              </Link>
            </li>
          )}
        </ul>
      </nav>
    </div>
  );
};
export default Navbar;
