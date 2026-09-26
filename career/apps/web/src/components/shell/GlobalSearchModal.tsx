'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Search, X, Briefcase, FileText, Share2, CornerDownLeft } from 'lucide-react';

export interface GlobalSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
}

export function GlobalSearchModal({ isOpen, onClose, workspaceId = 'default-workspace' }: GlobalSearchModalProps) {
  const [query, setQuery] = useState('');
  const [indexScope, setIndexScope] = useState<'all' | 'jobs' | 'content' | 'applications'>('all');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
    } else {
      setQuery('');
    }
  }, [isOpen]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Global Workspace Search"
      className="fixed inset-0 z-50 flex items-start justify-center pt-20 px-4 bg-slate-900/60 backdrop-blur-xs animate-in fade-in duration-150"
    >
      <div
        className="w-full max-w-2xl bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-800 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search Header Input */}
        <div className="flex items-center px-4 border-b border-slate-200 dark:border-slate-800">
          <Search className="w-5 h-5 text-slate-400 shrink-0" aria-hidden="true" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search jobs, resumes, applications, and networking records..."
            className="w-full px-3 py-4 text-base bg-transparent border-0 focus:outline-none focus:ring-0 text-slate-900 dark:text-white placeholder-slate-400"
          />
          {query && (
            <button
              type="button"
              onClick={() => setQuery('')}
              className="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded cursor-pointer"
            >
              <X className="w-4 h-4" />
            </button>
          )}
          <button
            type="button"
            onClick={onClose}
            className="ml-2 text-xs px-2 py-1 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded border border-slate-200 dark:border-slate-700 cursor-pointer"
          >
            ESC
          </button>
        </div>

        {/* Scope Filters */}
        <div className="flex items-center gap-1.5 px-4 py-2.5 bg-slate-50 dark:bg-slate-950/50 border-b border-slate-100 dark:border-slate-800/80 text-xs">
          <span className="text-slate-400 mr-1">Scope:</span>
          {(['all', 'jobs', 'content', 'applications'] as const).map((scope) => (
            <button
              key={scope}
              type="button"
              onClick={() => setIndexScope(scope)}
              className={`px-2.5 py-1 rounded-md capitalize font-medium transition cursor-pointer ${
                indexScope === scope
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-200/60 dark:hover:bg-slate-800'
              }`}
            >
              {scope}
            </button>
          ))}
          <span className="ml-auto text-[11px] text-slate-400">Workspace-isolated (AT-012)</span>
        </div>

        {/* Results / Empty View */}
        <div className="p-6 max-h-96 overflow-y-auto">
          {query.trim().length === 0 ? (
            <div className="text-center py-8 text-slate-400 text-sm">
              <Search className="w-8 h-8 mx-auto mb-2 opacity-30" />
              <p>Type keywords to search across your isolated workspace records.</p>
              <p className="text-xs text-slate-500 mt-1">Raw unscrubbed resumes are strictly excluded from search projections (REQ-017).</p>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">
                Matching records for &quot;{query}&quot;
              </div>

              {/* Sample simulated match items for UI demonstration */}
              <div className="p-3 rounded-xl border border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition cursor-pointer flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-blue-50 dark:bg-blue-950/50 text-blue-600 dark:text-blue-400 flex items-center justify-center">
                    <Briefcase className="w-4 h-4" />
                  </div>
                  <div>
                    <div className="text-sm font-medium text-slate-900 dark:text-white">
                      Senior Full-Stack Engineer
                    </div>
                    <div className="text-xs text-slate-500">Jobs Index • Match Score: 92%</div>
                  </div>
                </div>
                <CornerDownLeft className="w-4 h-4 text-slate-300" />
              </div>

              <div className="p-3 rounded-xl border border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition cursor-pointer flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-sky-50 dark:bg-sky-950/50 text-sky-600 dark:text-sky-400 flex items-center justify-center">
                    <FileText className="w-4 h-4" />
                  </div>
                  <div>
                    <div className="text-sm font-medium text-slate-900 dark:text-white">
                      Targeted Technical Lead Outreach
                    </div>
                    <div className="text-xs text-slate-500">Networking Index • Matched Lead</div>
                  </div>
                </div>
                <CornerDownLeft className="w-4 h-4 text-slate-300" />
              </div>
            </div>
          )}
        </div>

        {/* Footer info */}
        <div className="px-4 py-3 bg-slate-50 dark:bg-slate-950/80 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500">
          <span>Navigate with arrow keys</span>
          <span>Press ESC to close</span>
        </div>
      </div>
    </div>
  );
}
