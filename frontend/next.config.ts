import type { NextConfig } from "next";

const adminUiEnabled = process.env.ADMIN_UI_ENABLED === "1";
const isDevelopment = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  distDir: adminUiEnabled && isDevelopment ? ".next-admin" : ".next",
  turbopack: {
    root: process.cwd(),
  },
};

export default nextConfig;
