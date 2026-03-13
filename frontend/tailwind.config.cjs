/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,svelte}", "./src/app.html"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light"],
  },
};
