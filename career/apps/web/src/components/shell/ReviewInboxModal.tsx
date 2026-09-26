'use client';

import React, { useState } from 'react';
import { Bell, Check, X, ShieldAlert, Clock, CheckCircle2, AlertTriangle } from 'lucide-react';

export interface ReviewInboxModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ReviewInboxModal({ isOpen, onClose }: ReviewInboxModalProps) {
  const [activeTab, setActiveTab] = useState<'approvals' | 'notifications'>('approvals');
  const [actionDone, setActionDone] = useState<string | null>(null);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Shared Review Inbox & Notifications"
      className="fixed inset-0 z-50 flex items-start justify-end p-4 sm:p-6 bg-slate-900/40 backdrop-blur-xs animate-in fade-in duration-150"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-800 overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div className="p-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-bold">
              <Bell className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900 dark:text-white">Review & Action Center</h2>
              <p className="text-[11px] text-slate-500">Human-in-the-loop validation (REQ-015)</p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tab Selector */}
        <div className="flex border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/50">
          <button
            type="button"
            onClick={() => setActiveTab('approvals')}
            className={`flex-1 py-2.5 text-xs font-semibold text-center border-b-2 transition cursor-pointer flex items-center justify-center gap-1.5 ${
              activeTab === 'approvals'
                ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
                : 'border-transparent text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
            }`}
          >
            <span>Pending Approvals</span>
            <span className="px-1.5 py-0.5 rounded-full text-[10px] bg-indigo-100 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300">
              1
            </span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('notifications')}
            className={`flex-1 py-2.5 text-xs font-semibold text-center border-b-2 transition cursor-pointer flex items-center justify-center gap-1.5 ${
              activeTab === 'notifications'
                ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
                : 'border-transparent text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
            }`}
          >
            <span>Notifications</span>
            <span className="px-1.5 py-0.5 rounded-full text-[10px] bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              2
            </span>
          </button>
        </div>

        {/* Tab Content */}
        <div className="p-4 overflow-y-auto flex-1 space-y-3">
          {actionDone && (
            <div className="p-3 bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/40 rounded-xl text-xs text-emerald-800 dark:text-emerald-300 flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
              <span>{actionDone}</span>
            </div>
          )}

          {activeTab === 'approvals' ? (
            <div className="space-y-3">
              {/* Sample Pending Approval Card */}
              <div className="p-4 rounded-xl border border-indigo-100 dark:border-indigo-950 bg-indigo-50/40 dark:bg-indigo-950/20 space-y-3">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-blue-100 dark:bg-blue-950 text-blue-700 dark:text-blue-300">
                      LinkedIn DM Outreach
                    </span>
                    <h3 className="text-sm font-semibold text-slate-900 dark:text-white mt-1.5">
                      Recruiter Connection Follow-up
                    </h3>
                    <p className="text-xs text-slate-600 dark:text-slate-400 mt-1">
                      Recipient: Tech Lead @ Innovatech • Action consumes 1 slot from Recruiter Outreach budget.
                    </p>
                  </div>
                  <Clock className="w-4 h-4 text-slate-400 shrink-0" />
                </div>

                <div className="text-[11px] font-mono text-slate-500 bg-white/80 dark:bg-slate-900/80 p-2 rounded-lg border border-slate-200 dark:border-slate-800">
                  Payload SHA-256: 8f9b2d...4a1 (AT-007 Immutable)
                </div>

                <div className="flex items-center gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => setActionDone('Approval gate consumed & signed with single-use token (AT-007).')}
                    className="flex-1 py-1.5 px-3 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold rounded-lg shadow-xs transition flex items-center justify-center gap-1.5 cursor-pointer"
                  >
                    <Check className="w-3.5 h-3.5" />
                    <span>Approve & Sign</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setActionDone('Approval request rejected. No side effects dispatched.')}
                    className="py-1.5 px-3 border border-slate-300 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-medium rounded-lg transition cursor-pointer"
                  >
                    Reject
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex items-start gap-3">
                <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
                <div>
                  <div className="text-xs font-semibold text-slate-900 dark:text-white">
                    Quota Warning: LinkedIn DMs (80% reached)
                  </div>
                  <div className="text-[11px] text-slate-500 mt-0.5">
                    Account-wide rolling limit: 8 of 10 used today (AT-008, AT-018).
                  </div>
                </div>
              </div>

              <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex items-start gap-3">
                <ShieldAlert className="w-4 h-4 text-blue-500 shrink-0 mt-0.5" />
                <div>
                  <div className="text-xs font-semibold text-slate-900 dark:text-white">
                    Audit Log: Storage upload scrubbed
                  </div>
                  <div className="text-[11px] text-slate-500 mt-0.5">
                    Resume file uploaded and verified against quarantine policy (REQ-023).
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="p-3 bg-slate-50 dark:bg-slate-950/80 border-t border-slate-200 dark:border-slate-800 text-center text-xs text-slate-500">
          Immutable approval lifecycle with replay prevention (REQ-015)
        </div>
      </div>
    </div>
  );
}
