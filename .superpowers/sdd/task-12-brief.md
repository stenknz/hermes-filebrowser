### Task 12: Preview Pane + Search

**Files:**
- Create: `frontend/src/components/PreviewPane.tsx`
- Create: `frontend/src/components/SearchBar.tsx`
- Modify: `frontend/src/pages/BrowserPage.tsx` — add search bar (above Toolbar) and preview pane (below file list)

- [ ] **Step 1: Create PreviewPane**

File: `frontend/src/components/PreviewPane.tsx`
```tsx
import { useState, useEffect } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import { api } from '../api/client'
import { FileIcon } from './FileIcon'

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

interface Props {
  filePath: string | null
}

export default function PreviewPane({ filePath }: Props) {
  const [data, setData] = useState<string | null>(null)
  const ext = filePath?.split('.').pop()?.toLowerCase()

  useEffect(() => {
    if (!filePath) { setData(null); return }
    if (['jpg','jpeg','png','gif','webp','svg'].includes(ext || '')) {
      setData(`/api/files/raw?path=${encodeURIComponent(filePath)}`)
    } else {
      api.get(`/api/files/raw?path=${encodeURIComponent(filePath)}`).then(d => setData(d)).catch(() => setData(null))
    }
  }, [filePath])

  if (!filePath) return null

  const fileName = filePath.split('/').pop() || ''

  return (
    <div className="border-t border-[var(--color-border)] bg-[var(--color-surface)] p-4 max-h-64 overflow-auto">
      <div className="flex items-center gap-2 mb-3">
        <FileIcon name={fileName} isDir={false} className="w-4 h-4" />
        <span className="text-sm font-medium">{fileName}</span>
      </div>
      {['jpg','jpeg','png','gif','webp','svg'].includes(ext || '') && data && (
        <img src={data} alt={fileName} className="max-h-48 rounded" />
      )}
      {ext === 'pdf' && (
        <Document file={`/api/files/raw?path=${encodeURIComponent(filePath)}`}>
          <Page pageNumber={1} width={400} />
        </Document>
      )}
      {['txt','md','json','xml','yml','yaml','js','ts','jsx','tsx','css','html','go','py','sh','env','cfg','ini','log'].includes(ext || '') && data && (
        <pre className="text-xs leading-relaxed overflow-x-auto whitespace-pre-wrap">{data}</pre>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Create SearchBar**

File: `frontend/src/components/SearchBar.tsx`
```tsx
import { useState, useEffect, useRef } from 'react'
import { FiSearch } from 'react-icons/fi'
import { api } from '../api/client'

interface Props {
  path: string
  onResults: (files: any[]) => void
  onClear: () => void
}

export default function SearchBar({ path, onResults, onClear }: Props) {
  const [query, setQuery] = useState('')
  const timer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    if (!query.trim()) { onClear(); return }
    clearTimeout(timer.current)
    timer.current = setTimeout(async () => {
      const res = await api.get(`/api/search?q=${encodeURIComponent(query)}&path=${encodeURIComponent(path)}`)
      onResults(res.data)
    }, 300)
    return () => clearTimeout(timer.current)
  }, [query, path])

  return (
    <div className="relative px-4 py-2">
      <FiSearch className="absolute left-6 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] w-4 h-4" />
      <input
        className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg pl-9 pr-3 py-1.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
        placeholder="Search files..."
        value={query}
        onChange={e => setQuery(e.target.value)}
      />
    </div>
  )
}
```

- [ ] **Step 3: Wire into BrowserPage**

In BrowserPage.tsx, add:
1. Import `SearchBar` and `PreviewPane`
2. Add state: `const [searchResults, setSearchResults] = useState<any[] | null>(null)`
3. Add `PreviewPane` at the bottom (below DropZone/FileList), passing the selected file path
4. Add `SearchBar` above or inside the toolbar area, passing `path`, `onResults` (sets searchResults), and `onClear` (sets searchResults to null)
5. When searchResults is not null, show search results instead of the normal file list

Add between the Toolbar and Breadcrumb:
```tsx
<SearchBar
  path={path}
  onResults={(results) => setSearchResults(results)}
  onClear={() => setSearchResults(null)}
/>
```

And add PreviewPane at the bottom, replacing the current closing `</div>` structure:
```tsx
</DropZone>
<PreviewPane filePath={selectedFile} />
```

- [ ] **Step 4: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/PreviewPane.tsx frontend/src/components/SearchBar.tsx
git commit -m "feat: add preview pane and search bar"
```
