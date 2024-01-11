/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./templates/*.{html,js,gohtml}"],
	theme: {
		extend: {

			screens: {

				'mobile': {'raw': '(max-width: 800px)'}

			}
		},
	},
	plugins: [require('@tailwindcss/typography')],
}

