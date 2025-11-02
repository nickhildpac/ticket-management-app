import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { useAppDispatch, useAppSelector } from "../hooks/redux";
import { createTicket } from "../store/slices/ticketsSlice";

const CreateTicket = () => {
  const [description, setDescription] = useState("");
  const [title, setTitle] = useState("");
  const dispatch = useAppDispatch();
  const { loading, error } = useAppSelector((state) => state.tickets);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const result = await dispatch(createTicket({ title, description }));

    if (createTicket.fulfilled.match(result)) {
      alert("Ticket created successfully");
      navigate("/tickets");
    }
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Raise a ticket</h1>
        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}
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
          <Button label="Create ticket" onClick={() => { }} disabled={loading} />
        </form>
      </div>
    </div>
  );
};
export default CreateTicket;
