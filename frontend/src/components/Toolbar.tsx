import { FiUpload, FiFolderPlus, FiFilePlus, FiEdit3, FiTrash2, FiCopy, FiMove } from 'react-icons/fi'

interface Props {
  selectedFile: string | null
  onUpload: () => void
  onNewFolder: () => void
  onNewFile: () => void
  onRename: () => void
  onDelete: () => void
  onCopy: () => void
  onMove: () => void
  readOnly: boolean
}

export default function Toolbar({ selectedFile, onUpload, onNewFolder, onNewFile, onRename, onDelete, onCopy, onMove, readOnly }: Props) {
  const btn = "flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-surface)] transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b border-[var(--color-border)]">
      <button onClick={onUpload} disabled={readOnly} className={btn}><FiUpload className="w-3.5 h-3.5" /> Upload</button>
      <button onClick={onNewFolder} disabled={readOnly} className={btn}><FiFolderPlus className="w-3.5 h-3.5" /> Folder</button>
      <button onClick={onNewFile} disabled={readOnly} className={btn}><FiFilePlus className="w-3.5 h-3.5" /> File</button>
      <span className="w-px h-5 bg-[var(--color-border)]" />
      <button onClick={onRename} disabled={!selectedFile || readOnly} className={btn}><FiEdit3 className="w-3.5 h-3.5" /> Rename</button>
      <button onClick={onDelete} disabled={!selectedFile || readOnly} className={`${btn} hover:text-[var(--color-danger)]`}><FiTrash2 className="w-3.5 h-3.5" /> Delete</button>
      <button onClick={onCopy} disabled={!selectedFile || readOnly} className={btn}><FiCopy className="w-3.5 h-3.5" /> Copy</button>
      <button onClick={onMove} disabled={!selectedFile || readOnly} className={btn}><FiMove className="w-3.5 h-3.5" /> Move</button>
    </div>
  )
}
