'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { usePathname } from 'next/navigation';
import { Search, Bell, Shield, Menu, X, ChevronDown, CheckCircle2, User } from 'lucide-react';
import { GlobalSearchModal } from './GlobalSearchModal';
import { ReviewInboxModal } from './ReviewInboxModal';

export function AppHeader() {
  const pathname = usePathname();
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isReviewOpen, setIsReviewOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  // Global Ctrl+K / Cmd+K listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        setIsSearchOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const navItems = [
    { href: '/career', label: 'Career Hub', sectionId: 'career' },
    { href: '/linkedin', label: 'LinkedIn Hub', sectionId: 'linkedin' },
  ];

  return (
    <>
      <header className="border-b border-slate-200 dark:border-slate-800 bg-white/85 dark:bg-slate-900/85 backdrop-blur-md sticky top-0 z-40 transition-colors">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
          {/* Brand & Section Navigation Tabs (REQ-001) */}
          <div className="flex items-center space-x-6 md:space-x-8">
            <Link
              href="/"
              className="text-lg font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-2.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded-lg p-1"
            >
              <div className="w-8 h-8 rounded-xl overflow-hidden border border-slate-200 dark:border-slate-700 bg-white flex items-center justify-center shadow-xs">
                <Image
                  src="/logo.png"
                  alt="Career Platform Logo"
                  width={32}
                  height={32}
                  className="w-full h-full object-contain"
                  priority
                />
              </div>
              <span className="hidden sm:inline font-bold tracking-tight text-slate-900 dark:text-white">
                Career<span className="text-indigo-600 dark:text-indigo-400">Hub</span>
              </span>
            </Link>

            {/* Desktop Navigation Tabs (REQ-001, AT-023) */}
            <nav
              role="navigation"
              aria-label="Main platform sections"
              className="hidden md:flex items-center space-x-1"
            >
              {navItems.map((item) => {
                const isActive = pathname.startsWith(item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    aria-current={isActive ? 'page' : undefined}
                    className={`px-3.5 py-2 rounded-lg text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-indigo-500 ${
                      isActive
                        ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 shadow-xs'
                        : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800/60'
                    }`}
                  >
                    {item.label}
                  </Link>
                );
              })}
            </nav>
          </div>

          {/* Right Action Bar (Search, Review Inbox, Safe Mode, User) */}
          <div className="flex items-center space-x-3">
            {/* Global Search Button (AT-012) */}
            <button
              type="button"
              onClick={() => setIsSearchOpen(true)}
              aria-label="Open global workspace search"
              className="hidden sm:inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/40 text-slate-500 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700 text-xs font-medium transition focus:outline-none focus:ring-2 focus:ring-indigo-500 cursor-pointer"
            >
              <Search className="w-3.5 h-3.5" aria-hidden="true" />
              <span>Search...</span>
              <kbd className="hidden lg:inline-block px-1.5 py-0.5 text-[10px] bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded font-mono text-slate-400">
                Ctrl+K
              </kbd>
            </button>

            {/* Mobile Search Icon Button */}
            <button
              type="button"
              onClick={() => setIsSearchOpen(true)}
              aria-label="Open search"
              className="sm:hidden p-2 text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
            >
              <Search className="w-4 h-4" />
            </button>

            {/* Review Inbox & Action Center (REQ-015, REQ-019) */}
            <button
              type="button"
              onClick={() => setIsReviewOpen(true)}
              aria-label="Review inbox and notifications (1 pending approval)"
              className="relative p-2 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition focus:outline-none focus:ring-2 focus:ring-indigo-500 cursor-pointer"
            >
              <Bell className="w-4 h-4" />
              <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-indigo-600 ring-2 ring-white dark:ring-slate-900" />
            </button>

            {/* Circuit-breaker / Safe Mode Pill */}
            <div
              className="hidden lg:flex items-center gap-1.5 text-[11px] px-2.5 py-1 rounded-full border border-emerald-500/30 bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-400 font-medium"
              title="Platform is operating under strict human-approval and fail-closed quotas (REQ-009, REQ-015)"
            >
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              <span>Safe Mode (Deterministic)</span>
            </div>

            {/* Workspace & User Avatar */}
            <div className="flex items-center gap-2 pl-1 border-l border-slate-200 dark:border-slate-800">
              <div
                className="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/60 text-indigo-700 dark:text-indigo-300 flex items-center justify-center text-xs font-bold ring-1 ring-indigo-500/20"
                title="Personal Workspace (Owner)"
              >
                <User className="w-4 h-4" />
              </div>
            </div>

            {/* Mobile Menu Toggle Button */}
            <button
              type="button"
              onClick={() => setIsMobileMenuOpen((prev) => !prev)}
              aria-label="Toggle navigation menu"
              className="md:hidden p-2 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition"
            >
              {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile Navigation Dropdown */}
        {isMobileMenuOpen && (
          <div className="md:hidden border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-4 py-3 space-y-1">
            {navItems.map((item) => {
              const isActive = pathname.startsWith(item.href);
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`block px-3 py-2 rounded-lg text-sm font-semibold transition ${
                    isActive
                      ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400'
                      : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
                  }`}
                >
                  {item.label}
                </Link>
              );
            })}
          </div>
        )}
      </header>

      {/* Global Modals */}
      <GlobalSearchModal isOpen={isSearchOpen} onClose={() => setIsSearchOpen(false)} />
      <ReviewInboxModal isOpen={isReviewOpen} onClose={() => setIsReviewOpen(false)} />
    </>
  );
}
