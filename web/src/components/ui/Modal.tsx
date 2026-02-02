import { useEffect } from 'react'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
  footer?: React.ReactNode
}

export default function Modal({ isOpen, onClose, title, children, footer }: ModalProps) {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }

    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = 'unset'
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="fixed inset-0 bg-black bg-opacity-25 transition-opacity" onClick={onClose} />
        <div className="relative w-full max-w-lg transform rounded-apple-lg bg-white shadow-apple-lg transition-all">
          <div className="flex items-center justify-between border-b border-apple-gray-100 px-6 py-4">
            <h3 className="text-lg font-semibold text-apple-gray-500">{title}</h3>
            <button
              onClick={onClose}
              className="rounded-lg p-1 text-apple-gray-300 hover:bg-apple-gray-50 hover:text-apple-gray-400"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="px-6 py-4">{children}</div>
          {footer && (
            <div className="flex justify-end space-x-3 border-t border-apple-gray-100 px-6 py-4">
              {footer}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
