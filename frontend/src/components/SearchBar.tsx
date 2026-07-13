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
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

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
