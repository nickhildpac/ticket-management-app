import { Route, Routes } from "react-router-dom";

import Navbar from "./components/Navbar";
import CreateTicket from "./pages/CreateTicket";
import TicketDetails from "./pages/TicketDetails";
import SignupPage from "./pages/SignupPage";
import LoginPage from "./pages/LoginPage";
import TicketList from "./pages/TicketList";
import HomePage from "./pages/HomePage";
import AboutPage from "./pages/AboutPage";
import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "./hooks/redux";
import { refreshToken } from "./store/slices/authSlice";

function App() {
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const dispatch = useAppDispatch();

  useEffect(() => {
    dispatch(refreshToken());
  }, [dispatch]);

  return (
    <>
      <Navbar />
      <Routes>
        <Route path="/" element={<HomePage />}></Route>
        <Route path="/about" element={<AboutPage />}></Route>
        <Route path="/create" element={isAuthenticated ? <CreateTicket /> : <LoginPage />}></Route>
        <Route path="/login" element={isAuthenticated ? <HomePage /> : <LoginPage />}></Route>
        <Route path="/signup" element={isAuthenticated ? <HomePage /> : <SignupPage />}></Route>
        <Route path="/tickets" element={isAuthenticated ? <TicketList /> : <LoginPage />}></Route>
        <Route path="/ticket/:id" element={isAuthenticated ? <TicketDetails /> : <LoginPage />}></Route>
      </Routes>
    </>
  );
}

export default App;
