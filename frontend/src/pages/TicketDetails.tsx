import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { Input } from '../components/Input';
import { useAuth } from '../context/useAuth';

interface ticket {
  id: number;
  title: string;
  description: string;
  created_by: string;
  assigned_to: string;
  priority: string;
  status: string;
  comments: [comment: {
    id: number;
    ticket_id: number;
    description: string;
    created_by: string;
    created_at: string;
  }]
}

const TicketDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const {token,user} = useAuth();
  
  // Mock data for demonstration
  const [ticket, setTicket] = useState<ticket>({
    id: 0,
    title: '',
    description: '',
    created_by: '',
    assigned_to: '',
    priority: '',
    status: '',
    comments: [{
      id: 0,
      ticket_id: 0,
      description: '',
      created_by: '',
      created_at: '',
    }]
  });

  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(true);

  // Fetch ticket data based on ID
  useEffect(() => {
    // Simulate API call with setTimeout
    const requestOptions = {
      method: "GET",
      headers: {
        "Authorization": token
      }
    }
    setTimeout(() => {
      fetch(`${import.meta.env.VITE_SERVER_URL}/v1/ticket/${id}`,requestOptions)
        .then((response) => response.json())
        .then((data) => {
          console.log(data)
          setTicket(prev => ({...prev, ...data}));
          fetch(`${import.meta.env.VITE_SERVER_URL}/v1/ticket/${id}/comments`,requestOptions)
            .then((response) => response.json())
            .then((data) => {
              console.log(data)
              setTicket(prev => ({...prev, comments: data}));
              setLoading(false);
            });
        })
    }, 500);
  }, [id, token]);

  const handleAddComment = () => {
    if (newComment.trim() === '') return;
    
    const comment = {
      id: -1,
      description: newComment,
      ticket_id: Number(id),
      created_by: user, // In a real app, this would be the logged-in user
      created_at: new Date().toISOString(),
    };
    console.log(comment)
    fetch(`${import.meta.env.VITE_SERVER_URL}/v1/comment`, {
      method: "POST",
      headers: {
        "Authorization": token,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(comment)
    }).then((response) => response.json())
      .then((data) => {
        console.log(data)
        const newc = data
        const comments = [...ticket.comments, newc] as unknown as typeof ticket.comments
        setTicket({...ticket, comments});
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
      
      <div className="bg-white shadow-md rounded-lg p-6">
        <h1 className="text-2xl font-bold mb-4">{ticket.title}</h1>
        
        <div className="grid grid-cols-2 gap-4 mb-6">
          <div>
            <p className="text-sm text-gray-500">Priority</p>
            <p className="font-medium">{ticket.priority}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Assigned To</p>
            <p className="font-medium">{ticket.assigned_to}</p>
          </div>
        </div>
        
        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2">Description</p>
          <p className="text-gray-700">{ticket.description}</p>
        </div>
        
        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2">Skills</p>
          <div className="flex flex-wrap gap-2">
            {/* {ticket.skills.map((skill, index) => (
              <span key={index} className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded">{skill}</span>
            ))} */}
          </div>
        </div>
        
        <div className="border-t pt-6">
          <h2 className="text-lg font-semibold mb-4">Comments</h2>
          
          <div className="space-y-4 mb-6">
            {ticket.comments && ticket.comments.map((comment) => (
              <div key={comment.id} className="bg-gray-50 p-4 rounded">
                <div className="flex justify-between mb-2">
                  <p className="font-medium">{comment.created_by}</p>
                  <p className="text-sm text-gray-500">{comment.created_at}</p>
                </div>
                 <p>{comment.description}</p>
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
            <div className="mt-2">
              <Button label="Add Comment" onClick={handleAddComment} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
export default TicketDetails;
