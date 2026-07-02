import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/lib/query-client'
import { AuthProvider } from '@/app/auth-provider'
import { UserProvider } from '@/app/user-context'
import { router } from '@/app/router'

import { ThemeProvider } from '@/components/theme-provider'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
        <QueryClientProvider client={queryClient}>
          <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <AuthProvider>
              <UserProvider>
                <RouterProvider router={router} context={{ qc: queryClient }} />
              </UserProvider>
            </AuthProvider>
          </ThemeProvider>
        </QueryClientProvider>
  </StrictMode>,
)
