import { FileIcon } from './FileIcon'

interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: string
}

interface Props {
  file: FileInfo
  onNavigate: (path: string) => void
  onSelect?: (path: string) => void
  selected?: boolean
}

export default function FileRow({ file, onNavigate, onSelect, selected }: Props) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-2 rounded-lg cursor-pointer transition-colors ${
        selected ? 'bg-[var(--color-accent)]/20' : 'hover:bg-[var(--color-surface)]'
      }`}
      onClick={() => file.isDir ? onNavigate(file.path) : onSelect?.(file.path)}
    >
      <FileIcon name={file.name} isDir={file.isDir} className="w-5 h-5 shrink-0" />
      <span className="flex-1 truncate text-sm">{file.name}</span>
      {!file.isDir && (
        <span className="text-xs text-[var(--color-text-muted)] w-20 text-right">
          {file.size > 1024 * 1024
            ? (file.size / 1024 / 1024).toFixed(1) + ' MB'
            : file.size > 1024
            ? (file.size / 1024).toFixed(1) + ' KB'
            : file.size + ' B'}
        </span>
      )}
      <span className="text-xs text-[var(--color-text-muted)] w-24 text-right">
        {new Date(file.modTime).toLocaleDateString()}
      </span>
    </div>
  )
}
