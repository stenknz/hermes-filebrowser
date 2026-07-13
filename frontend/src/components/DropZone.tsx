import { useState, DragEvent } from 'react'
import { api } from '../api/client'

interface Props {
  path: string
  onUploadComplete: () => void
  children: React.ReactNode
}

export default function DropZone({ path, onUploadComplete, children }: Props) {
  const [dragging, setDragging] = useState(false)

  async function handleDrop(e: DragEvent) {
    e.preventDefault()
    setDragging(false)
    const files = Array.from(e.dataTransfer.files)
    for (const file of files) {
      const fd = new FormData()
      fd.append('file', file)
      await api.post(`/api/files/upload?path=${encodeURIComponent(path)}`, fd)
    }
    onUploadComplete()
  }

  return (
    <div
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
      className={`flex-1 flex flex-col ${dragging ? 'ring-2 ring-[var(--color-accent)] bg-[var(--color-accent)]/5' : ''}`}
    >
      {children}
    </div>
  )
}
