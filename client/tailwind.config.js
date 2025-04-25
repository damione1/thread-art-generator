/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/**/*.templ",
    "./internal/components/**/*.templ",
    "./internal/layouts/**/*.templ",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#f0f7ff",
          100: "#e0f0ff",
          200: "#c7e2ff",
          300: "#a0cfff",
          400: "#73b5fe",
          500: "#5a7fff",
          600: "#3e5df7",
          700: "#3346e8",
          800: "#2b3bcc",
          900: "#283aa1",
          950: "#1c2566",
        },
        accent: {
          purple: "#6d5dde",
        },
        dark: {
          100: "#121826",
          200: "#0f141f",
          300: "#0b101a",
        },
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};
