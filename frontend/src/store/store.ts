import { configureStore } from '@reduxjs/toolkit';
import authSlice from './slices/authSlice';
import ticketsSlice from './slices/ticketsSlice';
import commentsSlice from './slices/commentsSlice';
import usersSlice from './slices/usersSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice,
    tickets: ticketsSlice,
    comments: commentsSlice,
    users: usersSlice,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;