/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [],
  theme: {
    extend: {
      colors: {
        // Usar CSS variables dos tokens
        'se-bg': 'var(--se-bg)',
        'se-surface': 'var(--se-surface)',
        'se-text': 'var(--se-text)',
        'se-muted': 'var(--se-muted)',
      },
      spacing: {
        'se-xs': 'var(--se-space-xs)',
        'se-sm': 'var(--se-space-sm)',
        'se-md': 'var(--se-space-md)',
        'se-lg': 'var(--se-space-lg)',
        'se-xl': 'var(--se-space-xl)',
      },
      borderRadius: {
        'se-sm': 'var(--se-radius-sm)',
        'se-md': 'var(--se-radius-md)',
        'se-lg': 'var(--se-radius-lg)',
      },
    },
  },
  plugins: [],
};

