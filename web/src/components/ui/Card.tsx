interface CardProps {
  children: React.ReactNode
  className?: string
  title?: string
  description?: string
  actions?: React.ReactNode
}

export default function Card({ children, className = '', title, description, actions }: CardProps) {
  return (
    <div className={`bg-white rounded-apple-lg shadow-apple border border-apple-gray-100 ${className}`}>
      {(title || description || actions) && (
        <div className="px-6 py-4 border-b border-apple-gray-100 flex items-center justify-between">
          <div>
            {title && <h3 className="text-lg font-semibold text-apple-gray-500">{title}</h3>}
            {description && <p className="text-sm text-apple-gray-300 mt-1">{description}</p>}
          </div>
          {actions && <div className="flex items-center space-x-2">{actions}</div>}
        </div>
      )}
      <div className="p-6">{children}</div>
    </div>
  )
}
