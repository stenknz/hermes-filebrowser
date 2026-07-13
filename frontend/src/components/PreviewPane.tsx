import { useState, useEffect } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
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
      fetch(`/api/files/raw?path=${encodeURIComponent(filePath)}`)
        .then(r => r.text())
        .then(d => setData(d))
        .catch(() => setData(null))
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
