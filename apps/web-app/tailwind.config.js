/** @type {import('tailwindcss').Config} */
const baseConfig = require('../../tailwind.config.js');

module.exports = {
  ...baseConfig,
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './src/**/*.{js,ts,jsx,tsx,mdx}',
    '../../packages/ui-package/src/**/*.{js,ts,jsx,tsx}',
  ],
};

