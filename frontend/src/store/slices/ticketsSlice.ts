import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { API_BASE_URL } from '../../config/api';

interface UserInfo {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
}

interface Ticket {
  id: string;
  created_by: string;
  creator: UserInfo;
  assigned_to: string[] | null;
  title: string;
  description: string;
  state: number;
  priority: number;
  created_at: string;
  updated_at: string | null;
}

interface TicketsState {
  tickets: Ticket[];
  assignedTickets: Ticket[];
  currentTicket: Ticket | null;
  loading: boolean;
  assignedLoading: boolean;
  error: string | null;
}

const initialState: TicketsState = {
  tickets: [],
  assignedTickets: [],
  currentTicket: null,
  loading: false,
  assignedLoading: false,
  error: null,
};

export const fetchTickets = createAsyncThunk(
  'tickets/fetchTickets',
  async (_, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket/all`, {
        method: 'GET',
        headers: {
          'Authorization': state.auth.token,
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch tickets');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch tickets');
    }
  }
);

export const fetchAssignedTickets = createAsyncThunk(
  'tickets/fetchAssignedTickets',
  async (_, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket/assigned`, {
        method: 'GET',
        headers: {
          'Authorization': state.auth.token,
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch assigned tickets');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch assigned tickets');
    }
  }
);

export const fetchTicketById = createAsyncThunk(
  'tickets/fetchTicketById',
  async (id: string, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket/${id}`, {
        method: 'GET',
        headers: {
          'Authorization': state.auth.token,
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch ticket');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch ticket');
    }
  }
);

export const createTicket = createAsyncThunk(
  'tickets/createTicket',
  async (ticket: { title: string; description: string }, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': state.auth.token,
        },
        credentials: 'include',
        body: JSON.stringify(ticket),
      });

      if (!response.ok) {
        throw new Error('Failed to create ticket');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to create ticket');
    }
  }
);

export const updateTicketState = createAsyncThunk(
  'tickets/updateTicketState',
  async ({ id, state: newState }: { id: string; state: string }, { rejectWithValue, getState }) => {
    try {
      const authState = getState() as { auth: { token: string } };
      const response = await fetch(`http://localhost:8080/api/v1/ticket/${id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': authState.auth.token,
        },
        credentials: 'include',
        body: JSON.stringify({ state: newState }),
      });

      if (!response.ok) {
        throw new Error('Failed to update ticket');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to update ticket');
    }
  }
);

export const updateTicketAssignment = createAsyncThunk(
  'tickets/updateTicketAssignment',
  async ({ id, assignedTo }: { id: string; assignedTo: string[] }, { rejectWithValue, getState }) => {
    try {
      const authState = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket/${id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': authState.auth.token,
        },
        credentials: 'include',
        body: JSON.stringify({ assigned_to: assignedTo }),
      });

      if (!response.ok) {
        throw new Error('Failed to update ticket assignment');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to update ticket assignment');
    }
  }
);

export const updateTicket = createAsyncThunk(
  'tickets/updateTicket',
  async ({ id, assignedTo, priority, state, description }: { 
    id: string; 
    assignedTo: string[]; 
    priority: string; 
    state: string;
    description: string 
  }, { rejectWithValue, getState }) => {
    try {
      const authState = getState() as { auth: { token: string } };
      const response = await fetch(`${API_BASE_URL}/ticket/${id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': authState.auth.token,
        },
        credentials: 'include',
        body: JSON.stringify({ 
          assigned_to: assignedTo,
          priority: priority,
          state: state,
          description: description
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to update ticket');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to update ticket');
    }
  }
);

const ticketsSlice = createSlice({
  name: 'tickets',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    clearCurrentTicket: (state) => {
      state.currentTicket = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch tickets
      .addCase(fetchTickets.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTickets.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets = action.payload;
      })
      .addCase(fetchTickets.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Fetch assigned tickets
      .addCase(fetchAssignedTickets.pending, (state) => {
        state.assignedLoading = true;
        state.error = null;
      })
      .addCase(fetchAssignedTickets.fulfilled, (state, action) => {
        state.assignedLoading = false;
        state.assignedTickets = action.payload;
      })
      .addCase(fetchAssignedTickets.rejected, (state, action) => {
        state.assignedLoading = false;
        state.error = action.payload as string;
      })
      // Fetch ticket by ID
      .addCase(fetchTicketById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTicketById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentTicket = action.payload;
      })
      .addCase(fetchTicketById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Create ticket
      .addCase(createTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createTicket.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets.push(action.payload);
      })
      .addCase(createTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Update ticket state
      .addCase(updateTicketState.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateTicketState.fulfilled, (state, action) => {
        state.loading = false;
        state.currentTicket = action.payload;
        state.tickets = state.tickets.map((ticket) =>
          ticket.id === action.payload.id ? action.payload : ticket
        );
      })
      .addCase(updateTicketState.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Update ticket assignment
      .addCase(updateTicketAssignment.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateTicketAssignment.fulfilled, (state, action) => {
        state.loading = false;
        state.currentTicket = action.payload;
        state.tickets = state.tickets.map((ticket) =>
          ticket.id === action.payload.id ? action.payload : ticket
        );
      })
      .addCase(updateTicketAssignment.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Update ticket
      .addCase(updateTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateTicket.fulfilled, (state, action) => {
        state.loading = false;
        state.currentTicket = action.payload;
        state.tickets = state.tickets.map((ticket) =>
          ticket.id === action.payload.id ? action.payload : ticket
        );
      })
      .addCase(updateTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearError, clearCurrentTicket } = ticketsSlice.actions;
export default ticketsSlice.reducer;
