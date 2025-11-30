import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { Input } from '../components/Input';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { fetchTicketById } from '../store/slices/ticketsSlice';
import { store } from '../store/store';

const TicketDetails = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { currentTicket, loading, error } = useAppSelector((state) => state.tickets);

  const [newComment, setNewComment] = useState('');
  const [comments, setComments] = useState<any[]>([]);

  // Fetch ticket data based on ID
  useEffect(() => {
    if (id) {
      dispatch(fetchTicketById(Number(id)));

      // Fetch comments separately
      const token = store.getState().auth.token;
      fetch(`${import.meta.env.VITE_SERVER_URL}/ticket/${id}/comments`, {
        method: "GET",
        headers: {
          "Authorization": token
        }
      })
        .then((response) => response.json())
        .then((data) => {
          setComments(data);
        });
    }
  }, [id, dispatch]);

  const handleAddComment = () => {
    if (newComment.trim() === '') return;

    const token = store.getState().auth.token;
    const user = store.getState().auth.user;
    const comment = {
      id: -1,
      description: newComment,
      ticket_id: Number(id),
      created_by: user, // In a real app, this would be the logged-in user
      created_at: new Date().toISOString(),
    };

    fetch(`${import.meta.env.VITE_SERVER_URL}/comment`, {
      method: "POST",
      headers: {
        "Authorization": token,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(comment)
    }).then((response) => response.json())
      .then((data) => {
        setComments(prev => [...prev, data]);
        setNewComment("");
      })
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8 flex justify-center items-center h-64">
        <p className="text-lg">Loading ticket details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8 text-red-600">
        Error: {error}
      </div>
    );
  }

  if (!currentTicket) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p className="text-lg">Ticket not found</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <button
          onClick={() => navigate('/tickets')}
          className="text-blue-600 hover:text-blue-800 flex items-center"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M9.707 14.707a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 1.414L7.414 9H15a1 1 0 110 2H7.414l2.293 2.293a1 1 0 010 1.414z" clipRule="evenodd" />
          </svg>
          Back to Tickets
        </button>
      </div>

      <div className="bg-white shadow-md rounded-lg p-6 dark:bg-gray-800 transition-colors duration-200">
        <h1 className="text-2xl font-bold mb-4 dark:text-white">{currentTicket.title}</h1>

        <div className="grid grid-cols-2 gap-4 mb-6">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">Created By</p>
            <p className="font-medium dark:text-white">{currentTicket.created_by}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">Assigned To</p>
            <p className="font-medium dark:text-white">{currentTicket.assigned_to || 'Unassigned'}</p>
          </div>
        </div>

        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2 dark:text-gray-400">Description</p>
          <p className="text-gray-700 dark:text-gray-300">{currentTicket.description}</p>
        </div>

        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2 dark:text-gray-400">Skills</p>
          <div className="flex flex-wrap gap-2">
            {/* {ticket.skills.map((skill, index) => (
              <span key={index} className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded">{skill}</span>
            ))} */}
          </div>
        </div>

        <div className="border-t pt-6 dark:border-gray-700">
          <h2 className="text-lg font-semibold mb-4 dark:text-white">Comments</h2>

          <div className="space-y-4 mb-6">
            {comments && comments.map((comment) => (
              <div key={comment.id} className="bg-gray-50 p-4 rounded dark:bg-gray-700">
                <div className="flex justify-between mb-2">
                  <p className="font-medium dark:text-white">{comment.created_by}</p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{comment.created_at}</p>
                </div>
                <p className="dark:text-gray-300">{comment.description}</p>
              </div>
            ))}
          </div>

          <div className="mt-4">
            <Input
              type="text"
              label="Add a comment"
              name="comment"
              value={newComment}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewComment(e.target.value)}
            />
            <div className="mt-2 flex justify-start">
              <Button label="Add Comment" onClick={handleAddComment} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
export default TicketDetails;
