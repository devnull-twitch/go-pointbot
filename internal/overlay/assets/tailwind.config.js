module.exports = {
  content: [
    './../tpl/*.html.tmpl',
    './src/edit.js',
    './src/view.js'
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
