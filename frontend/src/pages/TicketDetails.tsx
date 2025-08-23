import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { Input } from '../components/Input';

const TicketDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  
  // Mock data for demonstration
  const [ticket, setTicket] = useState({
    id: '1',
    title: 'Fix login page',
    priority: 'High',
    assignedTo: 'John Doe',
    description: 'The login page is not working properly. Users are unable to log in with correct credentials.',
    skills: ['React', 'TypeScript', 'Tailwind CSS'],
    comments: [
      { id: '1', author: 'Jane Smith', text: 'I will look into this issue.', date: '2023-06-15' },
      { id: '2', author: 'John Doe', text: 'I found the issue. Working on a fix.', date: '2023-06-16' }
    ]
  });

  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(true);

  // Fetch ticket data based on ID
  useEffect(() => {
    // Simulate API call with setTimeout
    setTimeout(() => {
      // In a real app, this would fetch the ticket with the given ID from an API
      setLoading(false);
    }, 500);
  }, [id]);

  const handleAddComment = () => {
    if (newComment.trim() === '') return;
    
    const comment = {
      id: ticket.comments.length + 1,
      text: newComment,
      author: 'Current User', // In a real app, this would be the logged-in user
      timestamp: new Date().toLocaleString()
    };
    
    setTicket({
      ...ticket,
      comments: [...ticket.comments, comment]
    });
    
    setNewComment('');
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
            <p className="font-medium">{ticket.assignedTo}</p>
          </div>
        </div>
        
        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2">Description</p>
          <p className="text-gray-700">{ticket.description}</p>
        </div>
        
        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2">Skills</p>
          <div className="flex flex-wrap gap-2">
            {ticket.skills.map((skill, index) => (
              <span key={index} className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded">{skill}</span>
            ))}
          </div>
        </div>
        
        <div className="border-t pt-6">
          <h2 className="text-lg font-semibold mb-4">Comments</h2>
          
          <div className="space-y-4 mb-6">
            {ticket.comments.map((comment) => (
              <div key={comment.id} className="bg-gray-50 p-4 rounded">
                <div className="flex justify-between mb-2">
                  <p className="font-medium">{comment.author}</p>
                  <p className="text-sm text-gray-500">{comment.timestamp}</p>
                </div>
                <p>{comment.text}</p>
              </div>
            ))}
          </div>
          
          <div className="mt-4">
            <Input
              label="Add a comment"
              name="comment"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
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
