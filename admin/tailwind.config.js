/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        navy: {
          900: '#0d1b2e',
          800: '#122035',
          700: '#1a2f4a',
          600: '#1e3a5f',
        }
      }
    }
  },
  plugins: [],
}
