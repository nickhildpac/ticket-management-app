import { configureStore } from '@reduxjs/toolkit';
import authSlice from './slices/authSlice';
import ticketsSlice from './slices/ticketsSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice,
    tickets: ticketsSlice,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;