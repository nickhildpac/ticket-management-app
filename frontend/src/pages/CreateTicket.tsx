import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { useAuth } from "../context/useAuth";

const CreateTicket = () => {
  const [description, setDescription] = useState("");
  const [title, setTitle] = useState("");
  const { token } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const ticket = {
      title: title,
      description: description,
      created_by: "dpac"
    };
    fetch(`${import.meta.env.VITE_SERVER_URL}/ticket`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token
      },
      body: JSON.stringify(ticket),
    })
      .then((response) => {
        if (response.ok) {
          alert("Ticket created successfully");
          navigate("/tickets");
        }
      })
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Raise a ticket</h1>
        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <div className="rounded-md shadow-sm space-y-4">
            <Input
              type={"text"}
              name={"title"}
              label={"Title"}
              value={title}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTitle(e.target.value)}
            />
            <Input
              type={"textarea"}
              name={"description"}
              label={"Description"}
              value={description}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDescription(e.target.value)}
            />
          </div>
          <Button label="Create ticket" onClick={() => { }} />
        </form>
      </div>
    </div>
  );
};
export default CreateTicket;
