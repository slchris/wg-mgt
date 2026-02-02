import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  _hasHydrated: boolean
  setAuth: (token: string) => void
  logout: () => void
  isAuthenticated: () => boolean
  setHasHydrated: (state: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      _hasHydrated: false,
      setAuth: (token: string) => set({ token }),
      logout: () => set({ token: null }),
      isAuthenticated: () => {
        const { token } = get()
        return !!token
      },
      setHasHydrated: (state: boolean) => set({ _hasHydrated: state }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token }),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true)
      },
    }
  )
)
