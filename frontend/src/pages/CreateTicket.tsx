import { useState } from "react";
import { Link } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";

const CreateTicket = () => {
  const [description, setDescription] = useState("");
  const [title, setTitle] = useState("");

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
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
            />
            <Input
              type={"textarea"}
              name={"description"}
              label={"Description"}
              value={description}
            />
          </div>
          <Button label="Create ticket" onClick={""} />
        </form>
      </div>
    </div>
  );
};
export default CreateTicket;
