import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  // @ts-ignore - allowedDevOrigins is a valid but sometimes untyped property in dev
  allowedDevOrigins: ["nektokz.org", "lvh.me", "*.loca.lt", "*.ngrok-free.app"],
};

export default nextConfig;
