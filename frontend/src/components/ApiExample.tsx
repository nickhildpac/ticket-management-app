import { useState } from 'react';
import { apiClient, ApiError } from '../lib/api';
import { useAuth } from '../context/AuthContext';

interface Ticket {
  id: number;
  title: string;
  description: string;
}

const ApiExample = () => {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { isAuthenticated } = useAuth();

  const fetchTickets = async () => {
    if (!isAuthenticated) {
      setError('Please log in first');
      return;
    }

    setLoading(true);
    setError('');
    
    try {
      // This will automatically handle token refresh if the access token expires
      const response = await apiClient.get('/ticket/all');
      setTickets(response.data as Ticket[]);
      console.log('Tickets fetched successfully:', response.data);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`Error ${err.status}: ${err.message}`);
      } else {
        setError('An unexpected error occurred');
      }
      console.error('Error fetching tickets:', err);
    } finally {
      setLoading(false);
    }
  };

  const createTicket = async () => {
    if (!isAuthenticated) {
      setError('Please log in first');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const newTicket = {
        title: 'Test Ticket',
        description: 'This is a test ticket created via API client',
        priority: 'medium'
      };

      const response = await apiClient.post('/ticket/', newTicket);
      console.log('Ticket created successfully:', response.data);
      
      // Refresh the tickets list
      await fetchTickets();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`Error ${err.status}: ${err.message}`);
      } else {
        setError('An unexpected error occurred');
      }
      console.error('Error creating ticket:', err);
    } finally {
      setLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="p-4 bg-yellow-100 rounded">
        <p>Please log in to test API functionality</p>
      </div>
    );
  }

  return (
    <div className="p-4 space-y-4">
      <h2 className="text-xl font-bold">API Client Example</h2>
      <p className="text-sm text-gray-600">
        This component demonstrates automatic token refresh when making API calls.
      </p>
      
      <div className="space-x-2">
        <button
          onClick={fetchTickets}
          disabled={loading}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {loading ? 'Loading...' : 'Fetch Tickets'}
        </button>
        
        <button
          onClick={createTicket}
          disabled={loading}
          className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
        >
          {loading ? 'Loading...' : 'Create Test Ticket'}
        </button>
      </div>

      {error && (
        <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {tickets.length > 0 && (
        <div className="space-y-2">
          <h3 className="font-semibold">Tickets:</h3>
          <div className="max-h-40 overflow-y-auto">
            {tickets.map((ticket) => (
              <div key={ticket.id} className="p-2 bg-gray-100 rounded text-sm">
                <strong>{ticket.title}</strong> - {ticket.description}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default ApiExample;
