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
        ground: "#0A0A0B",
        surface: {
          DEFAULT: "#111113",
          raised: "#17171A",
          high: "#1D1D21",
          sunken: "#0C0C0E",
        },
        line: {
          DEFAULT: "#232327",
          strong: "#33333A",
        },
        ink: {
          DEFAULT: "#F4F4F5",
          muted: "#A2A2AA",
          faint: "#6B6B74",
        },
        brass: {
          DEFAULT: "#C79A4B",
          hover: "#E0B463",
          ink: "#17130A",
        },
        thread: "#EDEDEE",
        ok: "#7FB98A",
        danger: "#D4756E",
      },
      fontFamily: {
        sans: ['"Instrument Sans"', '"Helvetica Neue"', "Arial", "sans-serif"],
        serif: ['"Instrument Serif"', '"Iowan Old Style"', '"Times New Roman"', "serif"],
        mono: ['"JetBrains Mono"', "ui-monospace", '"SF Mono"', "Menlo", "monospace"],
      },
      borderRadius: {
        card: "20px",
        field: "10px",
      },
      keyframes: {
        "fade-in": {
          from: { opacity: "0", transform: "translateY(6px)" },
          to: { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        "fade-in": "fade-in .25s ease-out both",
      },
    },
  },
  plugins: [],
};
