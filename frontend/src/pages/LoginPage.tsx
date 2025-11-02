import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { useAuth } from "../context/useAuth";

const LoginPage = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const { login } = useAuth()
  const navigate = useNavigate();

  useEffect(() => {
    const requestOptions = {
      method: "GET",
      headers: {
        "Content-Type": "application/json"
      },
    }
    fetch(`${import.meta.env.VITE_SERVER_URL}/refresh`, {
      ...requestOptions,
      credentials: 'include' as RequestCredentials
    })
      .then((res) => res.json())
      .then((data) => {
        console.log(data)
        login(data.access_token, data.user.username)
      })
  }, [login])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log(username)
    console.log(password)
    const reqBody = {
      username,
      password
    }
    if (username && password) {
      fetch(`${import.meta.env.VITE_SERVER_URL}/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        credentials: "include",
        body: JSON.stringify(reqBody)
      })
        .then((res) => res.json())
        .then(data => {
          console.log(data)
          if (data && data.access_token) {
            login(data.access_token, data.user.username)
            navigate('/');
          }
        })
    }
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Login</h1>
        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <div className="rounded-md shadow-sm space-y-4">
            <Input
              type={"text"}
              name={"username"}
              label={"Username"}
              value={username}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
            />
            <Input
              type={"password"}
              name={"password"}
              label={"Password"}
              value={password}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
            />
          </div>
          <Button label="Login" onClick={() => { }} />
          <div className="text-sm text-center mt-4">
            <p>
              Don't have an account?{" "}
              <Link
                to="/signup"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Signup
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
};

export default LoginPage;
