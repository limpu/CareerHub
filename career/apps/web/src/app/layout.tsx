import type { Metadata } from "next";
import { AppHeader } from "../components/shell/AppHeader";
import { AppSidebar } from "../components/shell/AppSidebar";
import "./globals.css";

export const metadata: Metadata = {
  title: "Career & LinkedIn Platform",
  description: "Unified career management, resume tailoring, and authentic LinkedIn networking platform.",
  icons: {
    icon: [
      { url: '/favicon.ico' },
      { url: '/icon.png', type: 'image/png' },
    ],
    shortcut: '/favicon.png',
    apple: '/apple-icon.png',
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="min-h-screen flex flex-col antialiased bg-slate-50 text-slate-900">
        {/* Accessible Skip Link (AT-023) */}
        <a
          href="#main-content"
          className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-indigo-600 focus:text-white focus:font-medium focus:rounded-lg focus:shadow-xl focus:outline-none focus:ring-2 focus:ring-white"
        >
          Skip to main content
        </a>

        {/* Global App Shell Header */}
        <AppHeader />

        {/* App Layout with Left Sidebar */}
        <div className="flex-1 flex">
          <AppSidebar />

          <div className="flex-1 lg:pl-72 flex flex-col min-w-0">
            <main id="main-content" tabIndex={-1} className="flex-1 w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 focus:outline-none">
              {children}
            </main>

            <footer className="border-t border-slate-200 dark:border-slate-800 py-6 text-center text-xs text-slate-500 bg-white/50 dark:bg-slate-900/50">
              <p className="flex items-center justify-center gap-1.5 flex-wrap">
                <span>Develop by</span>
                <a
                  href="https://www.linkedin.com/in/atiqueullahlimon"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-bold text-sky-600 dark:text-sky-400 hover:text-sky-700 dark:hover:text-sky-300 underline underline-offset-4 decoration-sky-400 hover:decoration-sky-600 transition"
                >
                  Atique Ullah
                </a>
                <span className="text-slate-300 dark:text-slate-700 mx-1">•</span>
                <span>Career &amp; LinkedIn Platform</span>
              </p>
            </footer>
          </div>
        </div>
      </body>
    </html>
  );
}