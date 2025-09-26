import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  serverExternalPackages: ['@tailwindcss/postcss'],
};

export default nextConfig;
