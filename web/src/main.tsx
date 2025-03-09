import './index.css'
import { router } from './router'
import { RouterProvider } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

// Use standard promise pattern instead of top-level await
router.load().then(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <RouterProvider
        router={router}
        context={{
          auth: undefined, // This will be provided by AuthProvider
        }}
        defaultPreload="intent"
        defaultPreloadDelay={500}
      />
    </StrictMode>,
  )
}).catch(error => {
  console.error('Failed to load router:', error)
})
