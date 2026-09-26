'use client';

import React from 'react';
import { AlertCircle, HelpCircle } from 'lucide-react';

export interface GuidedEmptyStateProps {
  id: string;
  title: string;
  description: string;
  icon?: React.ReactNode;
  primaryAction?: {
    label: string;
    onClick: () => void;
    icon?: React.ReactNode;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  tip?: string;
  policyNote?: string;
}

export function GuidedEmptyState({
  id,
  title,
  description,
  icon,
  primaryAction,
  secondaryAction,
  tip,
  policyNote,
}: GuidedEmptyStateProps) {
  return (
    <section
      role="region"
      aria-labelledby={`${id}-heading`}
      className="border border-dashed border-slate-300 dark:border-slate-700 rounded-2xl p-8 sm:p-12 text-center bg-white/60 dark:bg-slate-900/60 backdrop-blur-sm transition hover:border-slate-400 dark:hover:border-slate-600"
    >
      <div className="max-w-xl mx-auto flex flex-col items-center">
        {icon && (
          <div className="w-14 h-14 rounded-2xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 flex items-center justify-center mb-5 shadow-inner">
            {icon}
          </div>
        )}

        <h2
          id={`${id}-heading`}
          className="text-xl font-bold tracking-tight text-slate-900 dark:text-white"
        >
          {title}
        </h2>

        <p className="mt-2 text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
          {description}
        </p>

        {(primaryAction || secondaryAction) && (
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            {primaryAction && (
              <button
                type="button"
                onClick={primaryAction.onClick}
                className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 active:bg-indigo-800 text-white text-sm font-semibold rounded-lg shadow-sm hover:shadow transition focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-slate-900 cursor-pointer"
              >
                {primaryAction.icon}
                <span>{primaryAction.label}</span>
              </button>
            )}

            {secondaryAction && (
              <button
                type="button"
                onClick={secondaryAction.onClick}
                className="inline-flex items-center px-4 py-2.5 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-sm font-medium rounded-lg transition focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2 dark:focus:ring-offset-slate-900 cursor-pointer"
              >
                {secondaryAction.label}
              </button>
            )}
          </div>
        )}

        {tip && (
          <div className="mt-6 inline-flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800/60 px-3 py-1.5 rounded-full">
            <HelpCircle className="w-3.5 h-3.5 text-indigo-500 shrink-0" aria-hidden="true" />
            <span>{tip}</span>
          </div>
        )}

        {policyNote && (
          <div className="mt-4 flex items-start gap-2 text-left text-xs text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-950/30 border border-amber-200/60 dark:border-amber-800/40 p-3 rounded-lg w-full">
            <AlertCircle className="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" aria-hidden="true" />
            <span>{policyNote}</span>
          </div>
        )}
      </div>
    </section>
  );
}
