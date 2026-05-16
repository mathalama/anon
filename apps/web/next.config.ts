import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  allowedDevOrigins: ["nektokz.org", "lvh.me", "*.loca.lt", "*.ngrok-free.app"],
};

export default nextConfig;
