import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/app/auth-provider'
import { UserProvider } from '@/app/user-context'
import { router } from '@/app/router'

import { ThemeProvider } from '@/components/theme-provider'

const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <UserProvider>
          <AuthProvider>
            <RouterProvider router={router} context={{ qc: queryClient }} />
          </AuthProvider>
        </UserProvider>
      </ThemeProvider>
    </QueryClientProvider>
  </StrictMode>,
)
