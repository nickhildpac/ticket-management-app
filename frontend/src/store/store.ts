import { configureStore } from '@reduxjs/toolkit';
import authSlice from './slices/authSlice';
import ticketsSlice from './slices/ticketsSlice';
import commentsSlice from './slices/commentsSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice,
    tickets: ticketsSlice,
    comments: commentsSlice,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;