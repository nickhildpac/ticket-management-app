import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../hooks/redux";
import { logoutAsync, logoutSync } from "../store/slices/authSlice";
import { useDarkMode } from "../hooks/useDarkMode";

const Navbar = () => {
  const dispatch = useAppDispatch();
  const { isAuthenticated, user } = useAppSelector((state) => state.auth);
  const { isDarkMode, toggleDarkMode } = useDarkMode();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const profileMenuRef = useRef<HTMLLIElement | null>(null);
  const userInitial = (userName: string | null) => {
    if (!userName) return "U";
    const trimmed = userName.trim();
    if (!trimmed) return "U";
    return trimmed[0].toUpperCase();
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (!profileMenuRef.current) return;
      if (!profileMenuRef.current.contains(event.target as Node)) {
        setIsProfileOpen(false);
      }
    };

    if (isProfileOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isProfileOpen]);

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
            <Link to="/tickets" className="text-white hover:text-indigo-300 transition-colors">My Tickets</Link>
          </li>
          <li>
            <Link to="/assignments" className="text-white hover:text-indigo-300 transition-colors">My Assignments</Link>
          </li>
          {/* <li>
            <Link to="/about" className="text-white hover:text-indigo-300 transition-colors">About</Link>
          </li> */}
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
            <li className="relative" ref={profileMenuRef}>
              <button
                type="button"
                onClick={() => setIsProfileOpen((prev) => !prev)}
                className="ml-4 flex items-center rounded-full p-1 text-white transition-colors hover:bg-gray-700"
                aria-haspopup="menu"
                aria-expanded={isProfileOpen}
              >
                <span
                  className={`inline-flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold uppercase ${
                    isDarkMode ? "bg-gray-700 text-white" : "bg-gray-200 text-gray-900"
                  }`}
                >
                  {userInitial(user)}
                </span>
              </button>
              {isProfileOpen ? (
                <div
                  className="absolute right-0 z-20 mt-2 w-44 overflow-hidden rounded-md border border-gray-700 bg-gray-800 shadow-lg"
                  role="menu"
                >
                  <Link
                    to="/profile"
                    className="block px-4 py-2 text-sm text-white transition-colors hover:bg-gray-700"
                    role="menuitem"
                    onClick={() => setIsProfileOpen(false)}
                  >
                    My Profile
                  </Link>
                  <button
                    type="button"
                    onClick={() => {
                      setIsProfileOpen(false);
                      dispatch(logoutSync());
                      dispatch(logoutAsync());
                    }}
                    className="block w-full px-4 py-2 text-left text-sm text-white transition-colors hover:bg-gray-700"
                    role="menuitem"
                  >
                    Logout
                  </button>
                </div>
              ) : null}
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
