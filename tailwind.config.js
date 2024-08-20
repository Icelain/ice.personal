/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./templates/*.{html,js,gohtml}"],
	theme: {
		extend: {

			screens: {

				'mobile': {'raw': '(max-width: 800px)'}

			}
		},
		fontFamily: {

			'mono': ["vga8"],

		},
	},
	plugins: [require('@tailwindcss/typography')],
}

