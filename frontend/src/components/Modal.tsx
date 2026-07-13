import { useState, useEffect, useRef } from 'react'

interface ModalProps {
  open: boolean
  title: string
  label: string
  initialValue?: string
  confirmText?: string
  onConfirm: (value: string) => void
  onCancel: () => void
}

export function PromptModal({ open, title, label, initialValue = '', confirmText = 'OK', onConfirm, onCancel }: ModalProps) {
  const [value, setValue] = useState(initialValue)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (open) {
      setValue(initialValue)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open, initialValue])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onCancel}>
      <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-sm space-y-4" onClick={e => e.stopPropagation()}>
        <h2 className="text-sm font-medium">{title}</h2>
        <label className="text-xs text-[var(--color-text-muted)]">{label}</label>
        <input
          ref={inputRef}
          className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]"
          value={value}
          onChange={e => setValue(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') onConfirm(value); if (e.key === 'Escape') onCancel() }}
        />
        <div className="flex justify-end gap-2">
          <button onClick={onCancel} className="text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors">Cancel</button>
          <button onClick={() => onConfirm(value)} className="text-xs px-3 py-1.5 rounded-md bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors">{confirmText}</button>
        </div>
      </div>
    </div>
  )
}

interface ConfirmProps {
  open: boolean
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmModal({ open, title, message, confirmText = 'Delete', danger, onConfirm, onCancel }: ConfirmProps) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onCancel}>
      <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-sm space-y-4" onClick={e => e.stopPropagation()}>
        <h2 className="text-sm font-medium">{title}</h2>
        <p className="text-sm text-[var(--color-text-muted)]">{message}</p>
        <div className="flex justify-end gap-2">
          <button onClick={onCancel} className="text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors">Cancel</button>
          <button onClick={onConfirm} className={`text-xs px-3 py-1.5 rounded-md text-white transition-colors ${danger ? 'bg-[var(--color-danger)] hover:bg-red-600' : 'bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)]'}`}>{confirmText}</button>
        </div>
      </div>
    </div>
  )
}
