import { FiFolder, FiFileText, FiImage, FiCode, FiArchive, FiFile, FiVideo, FiMusic } from 'react-icons/fi'

const iconMap: Record<string, any> = {
  dir: FiFolder,
  txt: FiFileText, md: FiFileText, json: FiCode, yml: FiCode, yaml: FiCode, xml: FiCode,
  js: FiCode, ts: FiCode, tsx: FiCode, jsx: FiCode, css: FiCode, html: FiCode, go: FiCode, py: FiCode,
  jpg: FiImage, jpeg: FiImage, png: FiImage, gif: FiImage, webp: FiImage, svg: FiImage,
  zip: FiArchive, tar: FiArchive, gz: FiArchive, rar: FiArchive, '7z': FiArchive,
  mp4: FiVideo, avi: FiVideo, mov: FiVideo, mkv: FiVideo,
  mp3: FiMusic, wav: FiMusic, flac: FiMusic, ogg: FiMusic,
}

export function FileIcon({ name, isDir, className = '' }: { name: string; isDir: boolean; className?: string }) {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const Icon = isDir ? FiFolder : iconMap[ext] || FiFile
  return <Icon className={`${className} ${isDir ? 'text-amber-400' : 'text-[var(--color-text-muted)]'}`} />
}
