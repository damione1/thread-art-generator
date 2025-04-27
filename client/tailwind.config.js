/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/**/*.templ",
    "./internal/components/**/*.templ",
    "./internal/layouts/**/*.templ",
  ],
  darkMode: "class",
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
          purple: "#9333ea",
        },
        dark: {
          100: "#111827",
          200: "#1f2937",
          300: "#374151",
          400: "#4b5563",
          500: "#6b7280",
        },
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
      animation: {
        "slow-pulse": "slow-pulse 8s ease-in-out infinite",
        "spin-slow": "spin-slow 12s linear infinite",
      },
      keyframes: {
        "slow-pulse": {
          "0%, 100%": { opacity: "0.1", transform: "scale(1)" },
          "50%": { opacity: "0.2", transform: "scale(1.05)" },
        },
        "spin-slow": {
          from: { transform: "rotate(0deg)" },
          to: { transform: "rotate(360deg)" },
        },
      },
    },
  },
  plugins: [],
};
