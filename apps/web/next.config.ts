import type { NextConfig } from "next";
import withPWAInit from "@ducanh2912/next-pwa";

const withPWA = withPWAInit({
  dest: "public",
  disable: process.env.NODE_ENV === "development",
});

const nextConfig: NextConfig = {
  /* config options here */
  output: "standalone",
  // Ignore typescript error if turbopack is not typed properly in some older versions
  // @ts-ignore
  turbopack: {},
};

export default withPWA(nextConfig);
