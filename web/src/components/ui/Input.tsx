import { forwardRef } from 'react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  helperText?: string
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, className = '', ...props }, ref) => {
    return (
      <div className="w-full">
        {label && (
          <label className="block text-sm font-medium text-apple-gray-400 mb-1">{label}</label>
        )}
        <input
          ref={ref}
          className={`w-full px-4 py-2 text-apple-gray-500 bg-white border rounded-apple transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-opacity-20 ${
            error
              ? 'border-apple-red focus:border-apple-red focus:ring-apple-red'
              : 'border-apple-gray-200 focus:border-apple-blue focus:ring-apple-blue'
          } ${className}`}
          {...props}
        />
        {error && <p className="mt-1 text-sm text-apple-red">{error}</p>}
        {helperText && !error && <p className="mt-1 text-sm text-apple-gray-300">{helperText}</p>}
      </div>
    )
  }
)

Input.displayName = 'Input'

export default Input
