
import { sora, spaceGrotesk, inter } from "../core/utils/_fonts";
import "./globals.css";


import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "MindMenu",
  description: "AI-powered restaurant chatbot landing page.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en"
      className={`${sora.variable} ${spaceGrotesk.variable} ${inter.variable}`}
      
    >
      <body className="antialiased">
        {children}
      </body>
    </html>
  );
}
