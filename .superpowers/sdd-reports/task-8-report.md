# Task 8 Report — API Client + Auth Context

**Status:** DONE

**Commits created:**
- `ff08639` feat: add API client and auth context

**Build result:** ✅ Success — `npx tsc --noEmit` passed, `npm run build` produced `dist/` with `index.html`, `assets/` (CSS 4.39 kB, JS 233.36 kB).

**Files:**
- `frontend/src/api/client.ts` — HTTP client with `get`, `post`, `put`, `delete`; auto-attaches Bearer token from localStorage and CSRF token from cookie.
- `frontend/src/context/AuthContext.tsx` — React context with `AuthProvider` component and `useAuth()` hook exposing `user`, `token`, `login`, `logout`, `isAuthenticated`.
