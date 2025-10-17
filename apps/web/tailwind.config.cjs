const config = {
  darkMode: 'class',
  content: ['./src/**/*.{svelte,ts}'],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#f2f9ff',
          100: '#e6f2ff',
          200: '#bfdeff',
          300: '#99caff',
          400: '#4d9eff',
          500: '#0072ff',
          600: '#005bd6',
          700: '#00459f',
          800: '#003068',
          900: '#001b31'
        }
      },
      fontFamily: {
        sans: ['"Inter var"', 'Inter', 'system-ui', 'sans-serif']
      }
    }
  },
  plugins: [require('@tailwindcss/forms')]
};

module.exports = config;
