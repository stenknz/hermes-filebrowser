### Task 7: Frontend Scaffolding

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/postcss.config.js`
- Create: `frontend/tsconfig.json`
- Create: `frontend/tsconfig.app.json`
- Create: `frontend/tsconfig.node.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/index.css`
- Create: `frontend/src/vite-env.d.ts`

- [ ] **Step 1: Scaffold project with Vite**

Run from repo root: `npm create vite@latest frontend -- --template react-ts`

Then:
```bash
cd frontend
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install react-router-dom react-icons react-pdf
```

- [ ] **Step 2: Configure Vite with Tailwind and API proxy**

File: `frontend/vite.config.ts`
```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false
  }
})
```

- [ ] **Step 3: Set up CSS**

File: `frontend/src/index.css`
```css
@import "tailwindcss";

:root {
  --color-bg: #0f0f0f;
  --color-surface: #1a1a1a;
  --color-border: #2a2a2a;
  --color-text: #e0e0e0;
  --color-text-muted: #888;
  --color-accent: #6c5ce7;
  --color-accent-hover: #5a4bd1;
  --color-danger: #e74c3c;
  --color-success: #2ecc71;
}

body {
  background-color: var(--color-bg);
  color: var(--color-text);
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}
```

- [ ] **Step 4: Create App.tsx with router**

File: `frontend/src/App.tsx`
```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import LoginPage from './pages/LoginPage'
import BrowserPage from './pages/BrowserPage'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<BrowserPage />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
```

- [ ] **Step 5: Verify frontend builds**

```bash
cd frontend && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/
git commit -m "feat: scaffold React frontend with Vite + Tailwind"
```
