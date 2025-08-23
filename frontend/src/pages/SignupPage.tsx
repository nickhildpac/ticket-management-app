import { useState } from "react";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { Link } from "react-router";

export default function SignupPage() {
  const [email, setEmail] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Signup</h1>
        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <div className="rounded-md shadow-sm space-y-4">
            <Input
              type={"text"}
              name={"first_name"}
              label={"First Name"}
              value={firstName}
            />
            <Input
              type={"text"}
              name={"last_name"}
              label={"Last Name"}
              value={lastName}
            />
            <Input
              type={"email"}
              name={"email"}
              label={"Email"}
              value={email}
            />
            <Input
              type={"password"}
              name={"password"}
              label={"Password"}
              value={password}
            />
            <Input
              type={"password"}
              name={"confirm_password"}
              label={"Confirm Password"}
              value={confirmPassword}
            />
          </div>
          <Button label="Signup" onClick={""} />
          <div className="text-sm text-center mt-4">
            <p>
              Already have an account?{" "}
              <Link
                to="/login"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Login
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
}
