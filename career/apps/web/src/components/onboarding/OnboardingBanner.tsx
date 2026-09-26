'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Circle, Sparkles, X } from 'lucide-react';

export function OnboardingBanner() {
  const [dismissed, setDismissed] = useState<boolean>(false);

  if (dismissed) return null;

  return (
    <div
      role="region"
      aria-label="Platform guided onboarding"
      className="relative overflow-hidden rounded-2xl border border-indigo-100 dark:border-indigo-950/60 bg-gradient-to-r from-indigo-50/80 via-white to-sky-50/80 dark:from-slate-900 dark:via-indigo-950/30 dark:to-slate-900 p-6 shadow-xs transition"
    >
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
        <div className="space-y-2 max-w-xl">
          <div className="flex items-center gap-2 text-xs font-semibold text-indigo-600 dark:text-indigo-400 tracking-wider uppercase">
            <Sparkles className="w-3.5 h-3.5" aria-hidden="true" />
            <span>Guided Onboarding Progress</span>
          </div>
          <h2 className="text-lg font-bold text-slate-900 dark:text-white">
            Complete your 2-Step Workspace Activation
          </h2>
          <p className="text-xs sm:text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
            Set up your canonical profile facts and master resumes, then verify your authentic LinkedIn networking strategy.
          </p>
        </div>

        {/* Action Controls */}
        <div className="flex items-center gap-2.5">
          <button
            type="button"
            onClick={() => setDismissed(true)}
            aria-label="Dismiss onboarding banner"
            className="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* 2 Step Indicator */}
      <div className="mt-6 pt-5 border-t border-slate-200/80 dark:border-slate-800 grid grid-cols-1 sm:grid-cols-2 gap-4">
        {/* Step 1: Career */}
        <Link
          href="/career"
          className="flex items-start gap-3 p-3.5 rounded-xl border border-blue-200 dark:border-blue-900/50 bg-white dark:bg-slate-900 hover:border-blue-400 transition"
        >
          <div className="mt-0.5">
            <div className="w-4 h-4 rounded-full border-2 border-blue-600 flex items-center justify-center">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-600"></span>
            </div>
          </div>
          <div className="min-w-0">
            <div className="text-xs font-semibold text-slate-900 dark:text-white flex items-center gap-1.5">
              <span>1. Career Profile &amp; Master Resumes</span>
            </div>
            <div className="text-[11px] text-slate-500 dark:text-slate-400 truncate">
              Upload resume (PDF/DOCX) or build profile facts
            </div>
          </div>
        </Link>

        {/* Step 2: LinkedIn */}
        <Link
          href="/linkedin"
          className="flex items-start gap-3 p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-sky-400 transition"
        >
          <Circle className="w-4 h-4 text-slate-400 mt-0.5" />
          <div className="min-w-0">
            <div className="text-xs font-semibold text-slate-900 dark:text-white">
              2. LinkedIn Connect &amp; Networking
            </div>
            <div className="text-[11px] text-slate-500 dark:text-slate-400 truncate">
              Profile synchronization and targeted outreach
            </div>
          </div>
        </Link>
      </div>
    </div>
  );
}
