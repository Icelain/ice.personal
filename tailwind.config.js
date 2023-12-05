/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/*.{html,js,gohtml}"],
  theme: {
    extend: {},
  },
  plugins: [require('@tailwindcss/typography')],
}

