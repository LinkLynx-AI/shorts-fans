import type { MetadataRoute } from "next";

import { appMetadata } from "@/shared/config";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: appMetadata.name,
    short_name: appMetadata.shortName,
    description: appMetadata.description,
    start_url: "/shorts",
    display: "standalone",
    background_color: appMetadata.backgroundColor,
    theme_color: appMetadata.themeColor,
    icons: [
      {
        src: "/icon-192x192.png",
        sizes: "192x192",
        type: "image/png",
        purpose: "maskable",
      },
      {
        src: "/icon-512x512.png",
        sizes: "512x512",
        type: "image/png",
        purpose: "maskable",
      },
    ],
  };
}
