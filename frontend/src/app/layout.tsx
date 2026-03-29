import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "shorts-fans frontend",
  description: "shorts-fans frontend development foundation",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
