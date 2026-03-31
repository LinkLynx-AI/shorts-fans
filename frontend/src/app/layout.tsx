import type { Metadata, Viewport } from "next";

import { appMetadata } from "@/shared/config";
import "./globals.css";

export const metadata: Metadata = {
  applicationName: appMetadata.name,
  title: {
    default: appMetadata.name,
    template: `%s | ${appMetadata.name}`,
  },
  description: appMetadata.description,
  icons: {
    apple: "/apple-touch-icon.png",
  },
  manifest: "/manifest.webmanifest",
};

export const viewport: Viewport = {
  themeColor: appMetadata.themeColor,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja" data-scroll-behavior="smooth">
      <body>{children}</body>
    </html>
  );
}
