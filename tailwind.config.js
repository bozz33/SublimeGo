/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    // C'est ICI qu'on dit à Tailwind de scanner nos fichiers .templ
    "./pkg/**/*.templ",
    "./views/**/*.templ",
    "./cmd/**/*.templ",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}