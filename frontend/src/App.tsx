import { Route, Routes } from "react-router-dom";

import Navbar from "./components/Navbar";
import CreateTicket from "./pages/CreateTicket";
import TicketDetails from "./pages/TicketDetails";
import SignupPage from "./pages/SignupPage";
import LoginPage from "./pages/LoginPage";
import TicketList from "./pages/TicketList";
import HomePage from "./pages/HomePage";
import AboutPage from "./pages/AboutPage";
import { AuthProvider } from "./context/AuthContext";

function App() {
  return (
    <AuthProvider>
      <Navbar />
      <Routes>
        <Route path="/" element={<HomePage />}></Route>
        <Route path="/about" element={<AboutPage />}></Route>
        <Route path="/create" element={<CreateTicket />}></Route>
        <Route path="/login" element={<LoginPage />}></Route>
        <Route path="/signup" element={<SignupPage />}></Route>
        <Route path="/tickets" element={<TicketList />}></Route>
        <Route path="/ticket/:id" element={<TicketDetails />}></Route>
      </Routes>
    </AuthProvider>
  );
}

export default App;