import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { Toast } from "../components/Toast";
import { useAppDispatch, useAppSelector } from "../hooks/redux";
import { clearError, createTicket } from "../store/slices/ticketsSlice";

const CreateTicket = () => {
  const [description, setDescription] = useState("");
  const [title, setTitle] = useState("");
  const [toast, setToast] = useState<{ message: string; variant: "success" | "error" } | null>(null);
  const dispatch = useAppDispatch();
  const { loading, error } = useAppSelector((state) => state.tickets);
  const navigate = useNavigate();

  useEffect(() => {
    if (error) {
      setToast({ message: error, variant: "error" });
    }
  }, [error]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const result = await dispatch(createTicket({ title, description }));

    if (createTicket.fulfilled.match(result)) {
      setToast({ message: "Ticket created successfully", variant: "success" });
      setTimeout(() => navigate("/tickets"), 1200);
    }
  };
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
      <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md dark:bg-gray-800">
        <h1 className="text-2xl font-bold text-center dark:text-white">Raise a ticket</h1>
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
      {toast && (
        <Toast
          message={toast.message}
          variant={toast.variant}
          onClose={() => {
            setToast(null);
            dispatch(clearError());
          }}
        />
      )}
    </div>
  );
};
export default CreateTicket;
