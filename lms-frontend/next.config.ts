import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  poweredByHeader: false,
  devIndicators: false,
  turbopack: { root: __dirname },
  async rewrites() {
    const auth = process.env.AUTH_SERVICE_URL || "http://127.0.0.1:3010";
    const core = process.env.CORE_SERVICE_URL || "http://127.0.0.1:3011";
    return [
      { source: "/api/auth/:path*", destination: `${auth}/v1/auth/:path*` },
      { source: "/api/core/:path*", destination: `${core}/api/v1/:path*` },
      { source: "/uploads/:path*", destination: `${core}/uploads/:path*` },
    ];
  },
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "X-Frame-Options", value: "DENY" },
        ],
      },
    ];
  },
};
export default nextConfig;
