interface Props {
  path: string
  onNavigate: (path: string) => void
}

export default function Breadcrumb({ path, onNavigate }: Props) {
  const parts = path.split('/').filter(Boolean)
  return (
    <nav className="flex items-center gap-1 text-sm text-[var(--color-text-muted)] px-4 py-2">
      <button onClick={() => onNavigate('')} className="hover:text-[var(--color-text)] transition-colors">Root</button>
      {parts.map((part, i) => {
        const p = parts.slice(0, i + 1).join('/')
        return (
          <span key={p} className="flex items-center gap-1">
            <span className="text-[var(--color-border)]">/</span>
            <button onClick={() => onNavigate(p)} className="hover:text-[var(--color-text)] transition-colors">{part}</button>
          </span>
        )
      })}
    </nav>
  )
}
