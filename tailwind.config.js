/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.tmpl", "./static/*.html"],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Roboto', 'sans-serif'],
        mono: ['Roboto Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}

