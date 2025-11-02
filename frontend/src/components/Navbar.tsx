import { Link } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../hooks/redux";
import { logoutAsync } from "../store/slices/authSlice";

const Navbar = () => {
  const dispatch = useAppDispatch();
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  return (
    <div className="flex justify-between items-center bg-gray-800 px-6 py-4">
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
