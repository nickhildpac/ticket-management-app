import animate from "tailwindcss-animate"

/** MD3-style palette: HSL components in index.css (:root / .dark), supports opacity modifiers. */
const mds = {
  "inverse-primary": "hsl(var(--mds-inverse-primary) / <alpha-value>)",
  "surface-container": "hsl(var(--mds-surface-container) / <alpha-value>)",
  "primary-fixed-dim": "hsl(var(--mds-primary-fixed-dim) / <alpha-value>)",
  "on-primary-fixed-variant": "hsl(var(--mds-on-primary-fixed-variant) / <alpha-value>)",
  "surface-tint": "hsl(var(--mds-surface-tint) / <alpha-value>)",
  "primary-fixed": "hsl(var(--mds-primary-fixed) / <alpha-value>)",
  "on-background": "hsl(var(--mds-on-background) / <alpha-value>)",
  "surface-container-highest": "hsl(var(--mds-surface-container-highest) / <alpha-value>)",
  "secondary-fixed": "hsl(var(--mds-secondary-fixed) / <alpha-value>)",
  "surface-container-low": "hsl(var(--mds-surface-container-low) / <alpha-value>)",
  "on-primary-container": "hsl(var(--mds-on-primary-container) / <alpha-value>)",
  error: "hsl(var(--mds-error) / <alpha-value>)",
  "on-tertiary-container": "hsl(var(--mds-on-tertiary-container) / <alpha-value>)",
  "outline-variant": "hsl(var(--mds-outline-variant) / <alpha-value>)",
  "on-secondary": "hsl(var(--mds-on-secondary) / <alpha-value>)",
  "surface-container-lowest": "hsl(var(--mds-surface-container-lowest) / <alpha-value>)",
  "on-secondary-container": "hsl(var(--mds-on-secondary-container) / <alpha-value>)",
  "on-tertiary-fixed-variant": "hsl(var(--mds-on-tertiary-fixed-variant) / <alpha-value>)",
  "secondary-container": "hsl(var(--mds-secondary-container) / <alpha-value>)",
  "on-secondary-fixed": "hsl(var(--mds-on-secondary-fixed) / <alpha-value>)",
  "on-surface-variant": "hsl(var(--mds-on-surface-variant) / <alpha-value>)",
  "surface-variant": "hsl(var(--mds-surface-variant) / <alpha-value>)",
  "on-primary": "hsl(var(--mds-on-primary) / <alpha-value>)",
  "mds-secondary": "hsl(var(--mds-mds-secondary) / <alpha-value>)",
  "inverse-on-surface": "hsl(var(--mds-inverse-on-surface) / <alpha-value>)",
  "on-error-container": "hsl(var(--mds-on-error-container) / <alpha-value>)",
  "on-secondary-fixed-variant": "hsl(var(--mds-on-secondary-fixed-variant) / <alpha-value>)",
  outline: "hsl(var(--mds-outline) / <alpha-value>)",
  "secondary-fixed-dim": "hsl(var(--mds-secondary-fixed-dim) / <alpha-value>)",
  "tertiary-fixed-dim": "hsl(var(--mds-tertiary-fixed-dim) / <alpha-value>)",
  surface: "hsl(var(--mds-surface) / <alpha-value>)",
  "inverse-surface": "hsl(var(--mds-inverse-surface) / <alpha-value>)",
  "tertiary-fixed": "hsl(var(--mds-tertiary-fixed) / <alpha-value>)",
  "on-tertiary-fixed": "hsl(var(--mds-on-tertiary-fixed) / <alpha-value>)",
  "surface-bright": "hsl(var(--mds-surface-bright) / <alpha-value>)",
  "on-surface": "hsl(var(--mds-on-surface) / <alpha-value>)",
  "error-container": "hsl(var(--mds-error-container) / <alpha-value>)",
  "on-primary-fixed": "hsl(var(--mds-on-primary-fixed) / <alpha-value>)",
  "surface-container-high": "hsl(var(--mds-surface-container-high) / <alpha-value>)",
  "on-error": "hsl(var(--mds-on-error) / <alpha-value>)",
  "primary-container": "hsl(var(--mds-primary-container) / <alpha-value>)",
  "surface-dim": "hsl(var(--mds-surface-dim) / <alpha-value>)",
  "tertiary-container": "hsl(var(--mds-tertiary-container) / <alpha-value>)",
  "on-tertiary": "hsl(var(--mds-on-tertiary) / <alpha-value>)",
  "tertiary": "hsl(var(--mds-tertiary) / <alpha-value>)",
} as const

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./app/**/*.{ts,tsx}",
    "./src/**/*.{ts,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      fontFamily: {
        headline: ["Manrope", "system-ui", "sans-serif"],
        body: ["Inter", "system-ui", "sans-serif"],
        label: ["Inter", "system-ui", "sans-serif"],
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
        xl: "0.75rem",
      },
      boxShadow: {
        "primary-soft": "0 10px 15px -3px rgb(129 140 248 / 0.2), 0 4px 6px -4px rgb(129 140 248 / 0.15)",
      },
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        ...mds,
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [animate],
}
