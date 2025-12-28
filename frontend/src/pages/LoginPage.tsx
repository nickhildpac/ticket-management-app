import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { useAppDispatch } from "../hooks/redux";
import { login } from "../store/slices/authSlice";

const LoginPage = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const dispatch = useAppDispatch();
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
        dispatch(login({ token: data.access_token, user: data.user.username }))
      })
  }, [dispatch])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log(email)
    console.log(password)
    const reqBody = {
      email,
      password
    }
    if (email && password) {
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
            dispatch(login({ token: data.access_token, user: data.user.username }))
            navigate('/');
          }
        })
    }
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md dark:bg-gray-800">
        <h1 className="text-2xl font-bold text-center dark:text-white">Login</h1>
        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <div className="rounded-md shadow-sm space-y-4">
            <Input
              type={"email"}
              name={"email"}
              label={"Email"}
              value={email}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
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
          <div className="text-sm text-center mt-4 dark:text-gray-300">
            <p>
              Don't have an account?{" "}
              <Link
                to="/signup"
                className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
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
