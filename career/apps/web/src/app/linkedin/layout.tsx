'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Users,
  Briefcase,
  Shield,
  FileEdit,
  BarChart3,
  Building2,
  Calendar,
  Layers,
  ArrowLeft,
  Lock,
  Menu,
  X,
  Compass,
  Database,
  Sparkles,
  Send,
  UserCheck,
  Megaphone,
} from 'lucide-react';

const NAV_ITEMS = [
  {
    name: 'Overview & Matrix',
    href: '/linkedin',
    icon: Shield,
    description: 'Scopes, capabilities & connect status',
    badge: 'Core',
  },
  {
    name: 'Profile & Vault',
    href: '/linkedin/profile',
    icon: Sparkles,
    description: 'Scan, optimize & normalized records',
    badge: 'Studio',
  },
  {
    name: 'Network & Outreach',
    href: '/linkedin/network',
    icon: Compass,
    description: 'Discovery, recruiter leads & queue',
    badge: 'Active',
  },
  {
    name: 'Content & Voice',
    href: '/linkedin/content',
    icon: FileEdit,
    description: 'Posts, comments, voice & story bank',
    badge: 'Creator',
  },
  {
    name: 'Growth & Calendar',
    href: '/linkedin/growth',
    icon: Calendar,
    description: 'Content planner, analytics & advocacy',
    badge: 'Metrics',
  },
  {
    name: 'Safety & Tools',
    href: '/linkedin/safety',
    icon: Lock,
    description: 'Rate limits, diagnostics & exports',
    badge: 'Guard',
  },
];

export default function LinkedInLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 flex flex-col">
      {/* Top Universal Navbar */}
      <header className="sticky top-0 z-30 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-900/95 backdrop-blur-sm shadow-xs">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Link
              href="/career"
              className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-sky-600 transition"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              <span>Career Hub</span>
            </Link>
            <span className="text-slate-300 dark:text-slate-700">/</span>
            <div className="flex items-center gap-2">
              <div className="w-5 h-5 rounded bg-sky-600 text-white flex items-center justify-center font-bold text-xs">
                in
              </div>
              <span className="font-bold text-sm tracking-tight text-slate-900 dark:text-white">
                LinkedIn Strategy & Operation Hub
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <span className="hidden sm:inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              Vault Grounded (AT-010)
            </span>

            {/* Mobile menu trigger */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden p-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
              aria-label="Toggle navigation"
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>
      </header>

      {/* Main Body with Persistent Navigation */}
      <div className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex flex-col md:flex-row gap-6">
        {/* Desktop Sidebar */}
        <aside className="hidden md:block w-64 shrink-0">
          <div className="sticky top-20 space-y-4">
            <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-3 shadow-xs">
              <div className="text-[10px] font-bold uppercase tracking-wider text-slate-400 px-3 py-1.5">
                LinkedIn Studios
              </div>
              <nav className="space-y-1">
                {NAV_ITEMS.map((item) => {
                  const Icon = item.icon;
                  const isActive =
                    item.href === '/linkedin'
                      ? pathname === '/linkedin'
                      : pathname?.startsWith(item.href);

                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`group flex items-start gap-3 px-3 py-2.5 rounded-xl text-xs font-semibold transition-all ${
                        isActive
                          ? 'bg-sky-50 dark:bg-sky-950/60 text-sky-700 dark:text-sky-300 ring-1 ring-sky-200 dark:ring-sky-800'
                          : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100/80 dark:hover:bg-slate-800/80'
                      }`}
                    >
                      <Icon
                        className={`w-4 h-4 shrink-0 mt-0.5 transition ${
                          isActive
                            ? 'text-sky-600 dark:text-sky-400'
                            : 'text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-200'
                        }`}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <span className="truncate">{item.name}</span>
                          {item.badge && (
                            <span
                              className={`text-[9px] font-bold px-1.5 py-0.2 rounded ${
                                isActive
                                  ? 'bg-sky-200/60 dark:bg-sky-900/60 text-sky-800 dark:text-sky-200'
                                  : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                              }`}
                            >
                              {item.badge}
                            </span>
                          )}
                        </div>
                        <p className="text-[10px] font-normal text-slate-400 dark:text-slate-500 truncate mt-0.5">
                          {item.description}
                        </p>
                      </div>
                    </Link>
                  );
                })}
              </nav>
            </div>

            {/* Platform Quick Specs Card */}
            <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/70 dark:bg-slate-900/70 p-4 text-[11px] text-slate-500 dark:text-slate-400 space-y-2">
              <div className="flex items-center gap-1.5 font-bold text-slate-700 dark:text-slate-300">
                <Shield className="w-3.5 h-3.5 text-sky-600" />
                <span>Zero-Ban Safety Guard</span>
              </div>
              <p className="text-[10px] leading-relaxed">
                All interactions operate strictly within LinkedIn Consumer OAuth limits and assisted-manual protocols (REQ-009, AT-010).
              </p>
            </div>
          </div>
        </aside>

        {/* Mobile Navigation Drawer */}
        {mobileMenuOpen && (
          <div className="md:hidden mb-4 p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm space-y-1">
            <div className="text-[10px] font-bold uppercase tracking-wider text-slate-400 px-3 py-1">
              Select LinkedIn Module
            </div>
            {NAV_ITEMS.map((item) => {
              const Icon = item.icon;
              const isActive =
                item.href === '/linkedin'
                  ? pathname === '/linkedin'
                  : pathname?.startsWith(item.href);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => setMobileMenuOpen(false)}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-semibold ${
                    isActive
                      ? 'bg-sky-50 dark:bg-sky-950/60 text-sky-700 dark:text-sky-300 font-bold'
                      : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100'
                  }`}
                >
                  <Icon className="w-4 h-4 text-sky-600" />
                  <span>{item.name}</span>
                </Link>
              );
            })}
          </div>
        )}

        {/* Content Viewport */}
        <main className="flex-1 min-w-0">{children}</main>
      </div>
    </div>
  );
}
