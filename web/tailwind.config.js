/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'apple-blue': '#007AFF',
        'apple-gray': {
          50: '#F5F5F7',
          100: '#E8E8ED',
          200: '#D2D2D7',
          300: '#86868B',
          400: '#6E6E73',
          500: '#1D1D1F',
        },
        'apple-green': '#34C759',
        'apple-red': '#FF3B30',
        'apple-orange': '#FF9500',
      },
      fontFamily: {
        'sf': ['-apple-system', 'BlinkMacSystemFont', 'SF Pro Display', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      borderRadius: {
        'apple': '12px',
        'apple-lg': '18px',
      },
      boxShadow: {
        'apple': '0 4px 24px rgba(0, 0, 0, 0.08)',
        'apple-lg': '0 8px 32px rgba(0, 0, 0, 0.12)',
      },
    },
  },
  plugins: [],
}
