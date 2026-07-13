import { useState, useEffect } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import { FileIcon } from './FileIcon'
import { FiEdit3, FiSave, FiX } from 'react-icons/fi'

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

const TEXT_EXTS = ['txt','md','json','xml','yml','yaml','js','ts','jsx','tsx','css','html','go','py','sh','env','cfg','ini','log']
const IMG_EXTS = ['jpg','jpeg','png','gif','webp','svg']

interface Props {
  filePath: string | null
  onRefresh?: () => void
}

export default function PreviewPane({ filePath, onRefresh }: Props) {
  const [data, setData] = useState<string | null>(null)
  const [editing, setEditing] = useState(false)
  const [editContent, setEditContent] = useState('')
  const [saving, setSaving] = useState(false)
  const ext = filePath?.split('.').pop()?.toLowerCase()
  const isText = TEXT_EXTS.includes(ext || '')
  const isImage = IMG_EXTS.includes(ext || '')

  useEffect(() => {
    setEditing(false)
    if (!filePath) { setData(null); return }
    if (isImage) {
      setData(`/api/files/raw?path=${encodeURIComponent(filePath)}`)
    } else {
      fetch(`/api/files/raw?path=${encodeURIComponent(filePath)}`)
        .then(r => r.text())
        .then(d => { setData(d); setEditContent(d) })
        .catch(() => setData(null))
    }
  }, [filePath])

  async function handleSave() {
    if (!filePath) return
    setSaving(true)
    try {
      await fetch(`/api/files/file`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': document.cookie.match(/(?:^| )csrf_token=([^;]+)/)?.[1] || '', 'Authorization': `Bearer ${localStorage.getItem('token') || ''}` },
        body: JSON.stringify({ path: filePath, content: editContent })
      })
      setEditing(false)
      onRefresh?.()
    } catch { alert('Save failed') }
    finally { setSaving(false) }
  }

  function startEdit() {
    setEditContent(data || '')
    setEditing(true)
  }

  if (!filePath) return null

  const fileName = filePath.split('/').pop() || ''

  return (
    <div className="border-t border-[var(--color-border)] bg-[var(--color-surface)] flex flex-col max-h-96">
      <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]">
        <div className="flex items-center gap-2">
          <FileIcon name={fileName} isDir={false} className="w-4 h-4" />
          <span className="text-sm font-medium">{fileName}</span>
        </div>
        {isText && editing ? (
          <div className="flex gap-1">
            <button onClick={handleSave} disabled={saving} className="flex items-center gap-1 text-xs px-2 py-1 rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors"><FiSave className="w-3 h-3" /> Save</button>
            <button onClick={() => setEditing(false)} className="flex items-center gap-1 text-xs px-2 py-1 rounded border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors"><FiX className="w-3 h-3" /></button>
          </div>
        ) : isText ? (
          <button onClick={startEdit} className="flex items-center gap-1 text-xs px-2 py-1 rounded border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors"><FiEdit3 className="w-3 h-3" /> Edit</button>
        ) : null}
      </div>
      <div className="p-4 overflow-auto">
        {isImage && data && <img src={data} alt={fileName} className="max-h-48 rounded" />}
        {ext === 'pdf' && <Document file={`/api/files/raw?path=${encodeURIComponent(filePath)}`}><Page pageNumber={1} width={400} /></Document>}
        {isText && !editing && data && <pre className="text-xs leading-relaxed whitespace-pre-wrap">{data}</pre>}
        {isText && editing && <textarea className="w-full h-48 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg p-3 text-xs font-mono leading-relaxed resize-y focus:outline-none focus:border-[var(--color-accent)]" value={editContent} onChange={e => setEditContent(e.target.value)} />}
      </div>
    </div>
  )
}
