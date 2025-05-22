/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/views/**/*.html"],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          DEFAULT: '#1a1a1a',
          lighter: '#2d2d2d',
          darker: '#0f0f0f',
        }
      }
    }
  },
  plugins: [],
}