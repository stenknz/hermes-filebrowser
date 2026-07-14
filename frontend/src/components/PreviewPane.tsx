import { useState, useEffect } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import { FileIcon } from './FileIcon'
import { FiEdit3, FiSave, FiX, FiDownload } from 'react-icons/fi'

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

const CODE_EXTS = ['py','js','ts','tsx','jsx','go','yaml','yml','json','xml','toml','sh','bash','css','html','rs','rb','php','java','c','cpp','h','sql','r','lua','pl']
const TEXT_EXTS = ['txt','log','env','cfg','ini','conf','csv','tsv']
const MD_EXTS = ['md','markdown']
const IMG_EXTS = ['jpg','jpeg','png','heic','webp','gif','svg','bmp','ico']
const AUDIO_EXTS = ['mp3','wav','flac','m4a','ogg','aac','wma']
const VIDEO_EXTS = ['mp4','mov','avi','mkv','webm','flv']
const ARCHIVE_EXTS = ['zip','rar','7z','tar','gz','bz2','xz']
function renderMarkdown(text: string): string {
  let html = text
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  // Headings
  html = html.replace(/^### (.+)$/gm, '<h3 class="text-sm font-bold mt-3 mb-1">$1</h3>')
  html = html.replace(/^## (.+)$/gm, '<h2 class="text-base font-bold mt-4 mb-1">$1</h2>')
  html = html.replace(/^# (.+)$/gm, '<h1 class="text-lg font-bold mt-4 mb-2">$1</h1>')
  // Bold & italic
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')
  // Links
  html = html.replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" class="text-[var(--color-accent)] underline">$1</a>')
  // Inline code
  html = html.replace(/`(.+?)`/g, '<code class="bg-[var(--color-bg)] px-1 rounded text-xs">$1</code>')
  // Unordered lists
  html = html.replace(/^- (.+)$/gm, '<li class="ml-4">$1</li>')
  html = html.replace(/(<li.*>[\s\S]*?)(<\/li>)/g, (m) => m)
  // Line breaks
  html = html.replace(/\n\n/g, '</p><p class="mb-2">')
  html = html.replace(/\n/g, '<br>')
  return '<p class="mb-2">' + html + '</p>'
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}

function renderCode(text: string): string {
  const lines = text.split('\n')
  return lines.map((line) =>
    `<span class="line-number"></span><span>${escapeHtml(line) || ' '}</span>`
  ).join('\n')
}

interface Props {
  filePath: string | null
  onRefresh?: () => void
}

export default function PreviewPane({ filePath, onRefresh }: Props) {
  const [data, setData] = useState<string | null>(null)
  const [editing, setEditing] = useState(false)
  const [editContent, setEditContent] = useState('')
  const [saving, setSaving] = useState(false)
  const [lightbox, setLightbox] = useState(false)
  const [imgError, setImgError] = useState(false)
  const ext = filePath?.split('.').pop()?.toLowerCase()
  const isCode = CODE_EXTS.includes(ext || '')
  const isText = TEXT_EXTS.includes(ext || '')
  const isMd = MD_EXTS.includes(ext || '')
  const isImage = IMG_EXTS.includes(ext || '')
  const isAudio = AUDIO_EXTS.includes(ext || '')
  const isVideo = VIDEO_EXTS.includes(ext || '')
  const isArchive = ARCHIVE_EXTS.includes(ext || '')
  const isPdf = ext === 'pdf'
  const canEdit = isCode || isText || isMd
  const rawUrl = filePath ? `/api/files/raw?path=${encodeURIComponent(filePath)}` : ''

  useEffect(() => {
    setEditing(false)
    setLightbox(false)
    setImgError(false)
    if (!filePath) { setData(null); return }
    if (isImage || isAudio || isVideo) {
      setData(rawUrl)
    } else {
      fetch(rawUrl)
        .then(r => r.text())
        .then(d => { setData(d); setEditContent(d) })
        .catch(() => setData(null))
    }
  }, [filePath])

  useEffect(() => {
    if (!lightbox) return
    function handleKey(e: KeyboardEvent) { if (e.key === 'Escape') setLightbox(false) }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [lightbox])

  async function handleSave() {
    if (!filePath) return
    setSaving(true)
    try {
      await fetch(`/api/files/file`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': document.cookie.match(/(?:^| )csrf_token=([^;]+)/)?.[1] || '' },
        body: JSON.stringify({ path: filePath, content: editContent })
      })
      setEditing(false)
      onRefresh?.()
    } catch { alert('Save failed') }
    finally { setSaving(false) }
  }

  if (!filePath) return null

  const fileName = filePath.split('/').pop() || ''

  return (
    <div className="border-t border-[var(--color-border)] bg-[var(--color-surface)] flex flex-col max-h-96">
      <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]">
        <div className="flex items-center gap-2 min-w-0">
          <FileIcon name={fileName} isDir={false} className="w-4 h-4 shrink-0" />
          <span className="text-sm font-medium truncate">{fileName}</span>
        </div>
        <div className="flex items-center gap-1 shrink-0">
          {canEdit && !editing && <button onClick={() => { setEditContent(data || ''); setEditing(true) }} className="flex items-center gap-1 text-xs px-2 py-1 rounded border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors"><FiEdit3 className="w-3 h-3" /> Edit</button>}
          {canEdit && editing && <><button onClick={handleSave} disabled={saving} className="flex items-center gap-1 text-xs px-2 py-1 rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors"><FiSave className="w-3 h-3" /> Save</button><button onClick={() => setEditing(false)} className="flex items-center gap-1 text-xs px-2 py-1 rounded border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors"><FiX className="w-3 h-3" /></button></>}
          {!canEdit && <a href={rawUrl} download={fileName} className="flex items-center gap-1 text-xs px-2 py-1 rounded border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors"><FiDownload className="w-3 h-3" /> Download</a>}
        </div>
      </div>
      <div className={`p-4 ${lightbox ? 'hidden' : ''}`}>
        {isImage && data && !imgError && <img src={data} alt={fileName} className="max-h-48 rounded cursor-pointer hover:opacity-90" onClick={() => setLightbox(true)} onError={() => setImgError(true)} />}
        {isImage && imgError && <div className="text-sm text-[var(--color-text-muted)] p-2">Format not supported by browser — <a href={rawUrl} download={fileName} className="text-[var(--color-accent)] underline">Download</a></div>}

        {isPdf && <Document file={rawUrl}><Page pageNumber={1} width={400} /></Document>}
        {isAudio && data && <audio controls src={data} className="w-full max-w-md" />}
        {isVideo && data && <video controls src={data} className="max-h-48 rounded" />}
        {isArchive && <div className="text-sm text-[var(--color-text-muted)] p-4 text-center">Archive: <a href={rawUrl} download={fileName} className="text-[var(--color-accent)] underline">Download {fileName}</a></div>}
        {isMd && !editing && data && <div className="text-sm leading-relaxed prose prose-invert max-w-none" dangerouslySetInnerHTML={{ __html: renderMarkdown(data) }} />}
        {isCode && !editing && data && <div className="text-xs leading-relaxed whitespace-pre-wrap font-mono" dangerouslySetInnerHTML={{ __html: renderCode(data) }} />}
        {isText && !editing && data && <div className="text-xs leading-relaxed whitespace-pre-wrap font-mono" dangerouslySetInnerHTML={{ __html: renderCode(data) }} />}
        {canEdit && editing && <textarea className="w-full min-h-[200px] max-h-[400px] bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg p-3 text-xs font-mono leading-relaxed resize-y focus:outline-none focus:border-[var(--color-accent)]" value={editContent} onChange={e => setEditContent(e.target.value)} />}
      </div>
      {lightbox && isImage && data && !imgError && (
        <div className="fixed inset-0 z-50 bg-black/95 flex items-center justify-center" onClick={() => setLightbox(false)}>
          <button onClick={() => setLightbox(false)} className="absolute top-4 right-4 text-white/60 hover:text-white z-10"><FiX className="w-6 h-6" /></button>
          <img src={data} alt={fileName} className="max-w-[95vw] max-h-[95vh] object-contain" onClick={e => e.stopPropagation()} onError={() => setLightbox(false)} />
        </div>
      )}
    </div>
  )
}
