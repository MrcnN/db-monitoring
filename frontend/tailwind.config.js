/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          600: '#2563eb',
          500: '#3b82f6',
        },
        background: {
          dark: '#111827',
        },
        card: {
          dark: '#1f2937',
        },
        border: {
          dark: '#374151',
        },
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
