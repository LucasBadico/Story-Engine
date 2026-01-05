/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@story-engine/ui-package', '@story-engine/shared-ts'],
};

module.exports = nextConfig;

