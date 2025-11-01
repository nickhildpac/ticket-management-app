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
import { useAuth } from "./context/useAuth";

function App() {
  const {isAuthenticated,login} = useAuth();
  useEffect(() => {
    const requestOptions = {
      method: "GET",
      headers: {
        "Content-Type": "application/json"
      },
    }
    fetch(`${import.meta.env.VITE_SERVER_URL}/v1/refresh`, {
      ...requestOptions,
      credentials: 'include' as RequestCredentials
    })
    .then((res) => {
      if (!res.ok) {
        throw new Error('Network response was not ok');
      }
      return res.json()
    })
    .then((data) => {
      console.log(data)
      login(data.access_token, data.user.username)
    })
    .catch((err) => {
      console.error(err)
    })
  }, [login])

  return (
    <>
      <Navbar />
      <Routes>
        <Route path="/" element={<HomePage />}></Route>
        <Route path="/about" element={<AboutPage />}></Route>
        <Route path="/create" element={isAuthenticated?<CreateTicket />: <LoginPage />}></Route>
        <Route path="/login" element={isAuthenticated?<HomePage />: <LoginPage />}></Route>
        <Route path="/signup" element={isAuthenticated?<HomePage />: <SignupPage />}></Route>
        <Route path="/tickets" element={isAuthenticated?<TicketList />: <LoginPage />}></Route>
        <Route path="/ticket/:id" element={isAuthenticated?<TicketDetails />: <LoginPage />}></Route>
      </Routes>
    </>
  );
}

export default App;