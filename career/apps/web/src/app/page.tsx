import Link from 'next/link';
import { OnboardingBanner } from '../components/onboarding/OnboardingBanner';
import { Briefcase, Share2, ArrowRight, ShieldCheck, FileText, CheckCircle2, Clock, Users } from 'lucide-react';

export default function HomePage() {
  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white">
          Unified Workspace Dashboard
        </h1>
        <p className="mt-2 text-sm sm:text-base text-slate-600 dark:text-slate-400 max-w-3xl leading-relaxed">
          Manage your verified professional career, resume tailoring, and authentic LinkedIn networking from one governed, privacy-preserving workspace.
        </p>
      </div>

      {/* Guided Onboarding */}
      <OnboardingBanner />

      {/* Two Primary Sections Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 1. Career Hub Card */}
        <div className="p-6 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/90 shadow-xs hover:shadow-md transition flex flex-col justify-between group">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-xl bg-blue-50 dark:bg-blue-950/60 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold shadow-inner">
                <Briefcase className="w-6 h-6" />
              </div>
              <span className="text-[11px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300">
                Hub 1
              </span>
            </div>

            <h2 className="text-xl font-bold text-slate-900 dark:text-white group-hover:text-blue-600 transition">
              Career Management
            </h2>
            <p className="mt-2 text-xs sm:text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              Upload existing resumes (PDF/DOCX) or build a profile manually. Generate ATS-friendly master resumes, discover matched jobs, and track verified applications.
            </p>

            <div className="mt-5 space-y-2 text-xs text-slate-500">
              <div className="flex items-center justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                <span>Profile Status:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">Draft (Not Verified)</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                <span>Tailored Resumes:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">0 Versions</span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span>Applications Logged:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">0 Verified</span>
              </div>
            </div>
          </div>

          <div className="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800">
            <Link
              href="/career"
              className="inline-flex items-center gap-2 text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 group-hover:translate-x-0.5 transition"
            >
              <span>Open Career Hub</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>

        {/* 2. LinkedIn Networking Card */}
        <div className="p-6 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/90 shadow-xs hover:shadow-md transition flex flex-col justify-between group">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-xl bg-sky-50 dark:bg-sky-950/60 text-sky-600 dark:text-sky-400 flex items-center justify-center font-bold shadow-inner">
                <Users className="w-6 h-6" />
              </div>
              <span className="text-[11px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full bg-sky-50 dark:bg-sky-950 text-sky-700 dark:text-sky-300">
                Hub 2
              </span>
            </div>

            <h2 className="text-xl font-bold text-slate-900 dark:text-white group-hover:text-sky-600 transition">
              LinkedIn Networking & Strategy
            </h2>
            <p className="mt-2 text-xs sm:text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              Maintain profile consistency, review network recommendations, and draft genuine relationship messages via official OIDC with zero unauthorized scraping.
            </p>

            <div className="mt-5 space-y-2 text-xs text-slate-500">
              <div className="flex items-center justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                <span>Account Status:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">Ready</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                <span>Daily Message Quota:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">0 / 10 Used</span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span>Consistency Diff:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">Synchronized</span>
              </div>
            </div>
          </div>

          <div className="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800">
            <Link
              href="/linkedin"
              className="inline-flex items-center gap-2 text-sm font-semibold text-sky-600 dark:text-sky-400 hover:text-sky-700 dark:hover:text-sky-300 group-hover:translate-x-0.5 transition"
            >
              <span>Open LinkedIn Hub</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </div>

      {/* Shared Platform Governance & Safety Note */}
      <div className="p-5 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-start gap-3">
          <ShieldCheck className="w-5 h-5 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" />
          <div className="text-xs sm:text-sm">
            <span className="font-bold text-slate-900 dark:text-white">Strict Privacy & Human Review Invariant: </span>
            <span className="text-slate-600 dark:text-slate-400">
              Personal resumes are owner-isolated (`AT-011`). Outbound side effects are blocked until cryptographic single-use review (`REQ-015`).
            </span>
          </div>
        </div>
        <div className="text-xs text-slate-400 shrink-0">
          Meilisearch Projections: Isolated (AT-012)
        </div>
      </div>
    </div>
  );
}