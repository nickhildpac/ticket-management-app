import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';

interface UserInfo {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
}

interface Comment {
  id: string;
  ticket_id: string;
  created_by: string;
  creator: UserInfo;
  description: string;
  created_at: string;
  updated_at: string | null;
}

interface CommentsState {
  comments: Comment[];
  loading: boolean;
  error: string | null;
}

const initialState: CommentsState = {
  comments: [],
  loading: false,
  error: null,
};

export const fetchCommentsByTicketId = createAsyncThunk(
  'comments/fetchCommentsByTicketId',
  async (ticketId: string, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${import.meta.env.VITE_SERVER_URL}/ticket/${ticketId}/comments`, {
        method: 'GET',
        headers: {
          'Authorization': state.auth.token,
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch comments');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch comments');
    }
  }
);

export const createComment = createAsyncThunk(
  'comments/createComment',
  async (comment: { ticket_id: string; description: string }, { rejectWithValue, getState }) => {
    try {
      const state = getState() as { auth: { token: string } };
      const response = await fetch(`${import.meta.env.VITE_SERVER_URL}/comment`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': state.auth.token,
        },
        credentials: 'include',
        body: JSON.stringify(comment),
      });

      if (!response.ok) {
        throw new Error('Failed to create comment');
      }

      const data = await response.json();
      return data;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Failed to create comment');
    }
  }
);

const commentsSlice = createSlice({
  name: 'comments',
  initialState,
  reducers: {
    clearComments: (state) => {
      state.comments = [];
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch comments
      .addCase(fetchCommentsByTicketId.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchCommentsByTicketId.fulfilled, (state, action) => {
        state.loading = false;
        state.comments = action.payload;
      })
      .addCase(fetchCommentsByTicketId.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Create comment
      .addCase(createComment.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createComment.fulfilled, (state, action) => {
        state.loading = false;
        state.comments.push(action.payload);
      })
      .addCase(createComment.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearComments, clearError } = commentsSlice.actions;
export default commentsSlice.reducer;