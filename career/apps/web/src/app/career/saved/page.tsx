'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  SavedJob,
  SavedJobStatus,
  JobExclusion,
  ExclusionType,
  ApplicationRecord,
  JobDedupeResult,
  DedupeMatch,
} from '@social-platform/contracts';

const INITIAL_SAVED_JOBS: SavedJob[] = [
  {
    id: 'saved-1',
    user_id: 'current-user',
    job_id: 'job-gh-101',
    canonical_url: 'https://boards.greenhouse.io/stripe/jobs/5412987',
    fingerprint: '3b890aef12d098e',
    title: 'Senior Distributed Systems Engineer (Go)',
    company: 'Stripe',
    location: {
      raw_location: 'San Francisco, CA / Remote',
      is_remote: true,
      is_hybrid: false,
      city: 'San Francisco',
      country: 'USA',
    },
    job_type: 'full_time',
    compensation: {
      min_amount: 195000,
      max_amount: 245000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    required_skills: ['Go', 'Distributed Systems', 'Kafka', 'PostgreSQL', 'Kubernetes'],
    direct_apply_url: 'https://boards.greenhouse.io/stripe/jobs/5412987#apply',
    status: 'shortlisted',
    notes: 'Prioritize for next week. Matches my 6 years Go distributed systems background.',
    priority: 5,
    tags: ['golang', 'distributed', 'high-priority'],
    saved_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
  },
  {
    id: 'saved-2',
    user_id: 'current-user',
    job_id: 'job-lev-202',
    canonical_url: 'https://jobs.lever.co/monzo/8912401',
    fingerprint: '77890bfa3298c11',
    title: 'Backend Platform Engineer (Microservices)',
    company: 'Monzo Bank',
    location: {
      raw_location: 'London, UK / Hybrid',
      is_remote: false,
      is_hybrid: true,
      city: 'London',
      country: 'UK',
    },
    job_type: 'full_time',
    compensation: {
      min_amount: 90000,
      max_amount: 120000,
      currency: 'GBP',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    required_skills: ['Go', 'Microservices', 'Cassandra', 'Envoy', 'gRPC'],
    direct_apply_url: 'https://jobs.lever.co/monzo/8912401/apply',
    status: 'ready_to_apply',
    notes: 'Tailored resume generated. Waiting for human approval in Step 11.',
    priority: 4,
    tags: ['fintech', 'microservices'],
    saved_at: new Date(Date.now() - 5 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'saved-3',
    user_id: 'current-user',
    job_id: 'job-ash-303',
    canonical_url: 'https://jobs.ashbyhq.com/linear/2198031',
    fingerprint: '9901ac88b20e402',
    title: 'Staff Full-Stack & Infrastructure Engineer',
    company: 'Linear',
    location: {
      raw_location: 'San Francisco, CA / Remote',
      is_remote: true,
      is_hybrid: false,
      city: 'San Francisco',
      country: 'USA',
    },
    job_type: 'full_time',
    compensation: {
      min_amount: 210000,
      max_amount: 260000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    required_skills: ['TypeScript', 'Node.js', 'React', 'GraphQL', 'PostgreSQL'],
    direct_apply_url: 'https://jobs.ashbyhq.com/linear/2198031/apply',
    status: 'applied',
    notes: 'Applied via company direct ATS on 2026-09-15. Confirmation email stored.',
    priority: 4,
    tags: ['remote', 'staff'],
    saved_at: new Date(Date.now() - 10 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    applied_at: new Date(Date.now() - 2 * 86400000).toISOString(),
  },
];

const INITIAL_APPLICATIONS: ApplicationRecord[] = [
  {
    id: 'app-rec-1',
    user_id: 'current-user',
    job_id: 'job-ash-303',
    canonical_url: 'https://jobs.ashbyhq.com/linear/2198031',
    fingerprint: '9901ac88b20e402',
    title: 'Staff Full-Stack & Infrastructure Engineer',
    company: 'Linear',
    status: 'submitted',
    submission_mode: 'manual_browser',
    applied_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    notes: 'Submitted via Linear Ashby portal. Automated receipt received #LIN-99412.',
  },
  {
    id: 'app-rec-2',
    user_id: 'current-user',
    job_id: 'job-prev-88',
    canonical_url: 'https://databricks.com/careers/lead-sre-88',
    fingerprint: '110298aefb904c7',
    title: 'Lead Cloud Infrastructure SRE',
    company: 'Databricks',
    status: 'in_review',
    submission_mode: 'direct_ats',
    applied_at: new Date(Date.now() - 12 * 86400000).toISOString(),
    notes: 'Screening interview completed on 2026-09-12.',
  },
];

const INITIAL_EXCLUSIONS: JobExclusion[] = [
  {
    id: 'excl-1',
    user_id: 'current-user',
    type: 'company_name',
    value: 'BadCulture Corp',
    reason: 'Poor work-life balance and uncompetitive compensation',
    company_name: 'BadCulture Corp',
    created_at: new Date(Date.now() - 14 * 86400000).toISOString(),
  },
  {
    id: 'excl-2',
    user_id: 'current-user',
    type: 'job_canonical_url',
    value: 'https://spammyrecruiter.com/job/fake-lead-role',
    reason: 'Scraped ghost listing / third-party aggregator spam',
    job_title: 'Lead Software Architect',
    company_name: 'Unknown Agency',
    created_at: new Date(Date.now() - 8 * 86400000).toISOString(),
  },
];

export default function SavedJobsPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<'saved' | 'applications' | 'exclusions' | 'dedupe'>('saved');

  // Saved Jobs state
  const [savedJobs, setSavedJobs] = useState<SavedJob[]>(INITIAL_SAVED_JOBS);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [tagFilter, setTagFilter] = useState<string>('');

  // Applications state
  const [applications, setApplications] = useState<ApplicationRecord[]>(INITIAL_APPLICATIONS);

  // Exclusions state
  const [exclusions, setExclusions] = useState<JobExclusion[]>(INITIAL_EXCLUSIONS);
  const [showAddExclusionModal, setShowAddExclusionModal] = useState(false);
  const [newExclType, setNewExclType] = useState<ExclusionType>('company_name');
  const [newExclValue, setNewExclValue] = useState('');
  const [newExclReason, setNewExclReason] = useState('');

  // Dedupe & Ambiguity Inspector state (AT-004)
  const [testJobTitle, setTestJobTitle] = useState('Senior Distributed Systems Software Engineer');
  const [testJobCompany, setTestJobCompany] = useState('Stripe');
  const [testJobURL, setTestJobURL] = useState('https://otherjobboard.com/stripe/dist-eng');
  const [dedupeInspectResult, setDedupeInspectResult] = useState<JobDedupeResult | null>(null);
  const [isInspecting, setIsInspecting] = useState(false);

  // Inline editing notes
  const [editingNotesId, setEditingNotesId] = useState<string | null>(null);
  const [tempNotes, setTempNotes] = useState<string>('');

  // Toast feedback
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  // Update status of a saved job
  const handleUpdateStatus = (jobId: string, newStatus: SavedJobStatus) => {
    setSavedJobs((prev) =>
      prev.map((j) => (j.id === jobId ? { ...j, status: newStatus, updated_at: new Date().toISOString() } : j))
    );
    showToast(`Status updated to "${newStatus}"`);
  };

  // Update priority
  const handleUpdatePriority = (jobId: string, priority: number) => {
    setSavedJobs((prev) =>
      prev.map((j) => (j.id === jobId ? { ...j, priority, updated_at: new Date().toISOString() } : j))
    );
    showToast(`Priority updated to ${priority} stars`);
  };

  // Save notes
  const handleSaveNotes = (jobId: string) => {
    setSavedJobs((prev) =>
      prev.map((j) => (j.id === jobId ? { ...j, notes: tempNotes, updated_at: new Date().toISOString() } : j))
    );
    setEditingNotesId(null);
    showToast('Notes saved successfully');
  };

  // Delete saved job
  const handleDeleteSavedJob = (jobId: string) => {
    if (confirm('Are you sure you want to remove this job from your saved list?')) {
      setSavedJobs((prev) => prev.filter((j) => j.id !== jobId));
      showToast('Job removed from saved list');
    }
  };

  // Add Exclusion
  const handleAddExclusion = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newExclValue.trim()) return;

    const newExcl: JobExclusion = {
      id: `excl-${Date.now()}`,
      user_id: 'current-user',
      type: newExclType,
      value: newExclValue.trim(),
      reason: newExclReason.trim() || 'Candidate specified exclusion rule',
      company_name: newExclType === 'company_name' ? newExclValue.trim() : undefined,
      created_at: new Date().toISOString(),
    };

    setExclusions((prev) => [newExcl, ...prev]);
    setShowAddExclusionModal(false);
    setNewExclValue('');
    setNewExclReason('');
    showToast(`Added persistent exclusion rule for "${newExcl.value}"`);
  };

  // Remove Exclusion
  const handleDeleteExclusion = (id: string, value: string) => {
    if (confirm(`Unblock "${value}" and allow jobs from this entity again?`)) {
      setExclusions((prev) => prev.filter((ex) => ex.id !== id));
      showToast(`Removed exclusion rule for "${value}"`);
    }
  };

  // Run Dedupe Inspection
  const handleRunDedupeInspection = () => {
    setIsInspecting(true);

    setTimeout(() => {
      // Check against exclusions
      const excluded = exclusions.find(
        (ex) =>
          (ex.type === 'company_name' && ex.value.toLowerCase() === testJobCompany.toLowerCase()) ||
          (ex.type === 'job_canonical_url' && ex.value.toLowerCase() === testJobURL.toLowerCase())
      );

      if (excluded) {
        setDedupeInspectResult({
          job_id: 'test-job-inspection',
          canonical_url: testJobURL,
          fingerprint: 'test-fingerprint-excl',
          matches: [
            {
              match_level: 'exact',
              target_ledger: 'exclusions',
              existing_id: excluded.id,
              existing_title: excluded.job_title || '',
              existing_company: excluded.company_name || excluded.value,
              similarity_score: 1.0,
              reason: `Excluded by persistent blocklist: ${excluded.reason}`,
              is_reviewable: false,
            },
          ],
          is_duplicate: true,
          is_ambiguous: false,
          is_excluded: true,
          action_taken: 'filtered_out',
        });
        setIsInspecting(false);
        return;
      }

      // Check against applications
      const appMatch = applications.find(
        (app) =>
          app.canonical_url.toLowerCase() === testJobURL.toLowerCase() ||
          (app.company.toLowerCase() === testJobCompany.toLowerCase() &&
            app.title.toLowerCase() === testJobTitle.toLowerCase())
      );

      if (appMatch) {
        setDedupeInspectResult({
          job_id: 'test-job-inspection',
          canonical_url: testJobURL,
          fingerprint: 'test-fingerprint-app',
          matches: [
            {
              match_level: 'exact',
              target_ledger: 'applications',
              existing_id: appMatch.id,
              existing_title: appMatch.title,
              existing_company: appMatch.company,
              similarity_score: 1.0,
              reason: `Already Applied on ${new Date(appMatch.applied_at).toLocaleDateString()} via ${appMatch.submission_mode}. Status: ${appMatch.status}.`,
              is_reviewable: false,
            },
          ],
          is_duplicate: true,
          is_ambiguous: false,
          is_excluded: false,
          action_taken: 'linked',
        });
        setIsInspecting(false);
        return;
      }

      // Check saved jobs: exact or ambiguous similarity
      const savedSameCompany = savedJobs.filter(
        (s) => s.company.toLowerCase() === testJobCompany.toLowerCase()
      );

      const matches: DedupeMatch[] = [];
      let isAmbiguous = false;
      let isDuplicate = false;

      for (const s of savedSameCompany) {
        if (s.canonical_url.toLowerCase() === testJobURL.toLowerCase() || s.title.toLowerCase() === testJobTitle.toLowerCase()) {
          isDuplicate = true;
          matches.push({
            match_level: 'exact',
            target_ledger: 'saved_jobs',
            existing_id: s.id,
            existing_title: s.title,
            existing_company: s.company,
            similarity_score: 1.0,
            reason: `Exact match with saved job in '${s.status}' state.`,
            is_reviewable: false,
          });
          break;
        }

        // Token Jaccard similarity test
        const wordsA = new Set(s.title.toLowerCase().split(/\W+/).filter(Boolean));
        const wordsB = new Set(testJobTitle.toLowerCase().split(/\W+/).filter(Boolean));
        let intersect = 0;
        wordsA.forEach((w) => {
          if (wordsB.has(w)) intersect++;
        });
        const union = new Set([...Array.from(wordsA), ...Array.from(wordsB)]).size;
        const sim = union > 0 ? intersect / union : 0;

        if (sim >= 0.75) {
          isAmbiguous = true;
          matches.push({
            match_level: 'ambiguous',
            target_ledger: 'saved_jobs',
            existing_id: s.id,
            existing_title: s.title,
            existing_company: s.company,
            similarity_score: Math.round(sim * 100) / 100,
            reason: `High textual similarity (${Math.round(sim * 100)}% token overlap) with saved job '${s.title}'. Requires human review (AT-004).`,
            is_reviewable: true,
          });
        }
      }

      setDedupeInspectResult({
        job_id: 'test-job-inspection',
        canonical_url: testJobURL,
        fingerprint: 'test-fingerprint-evaluated',
        matches,
        is_duplicate: isDuplicate,
        is_ambiguous: isAmbiguous,
        is_excluded: false,
        action_taken: isDuplicate ? 'linked' : isAmbiguous ? 'flagged_for_review' : 'retained',
      });
      setIsInspecting(false);
    }, 400);
  };

  // Filter saved jobs
  const filteredSavedJobs = savedJobs.filter((job) => {
    if (statusFilter !== 'all' && job.status !== statusFilter) return false;
    if (tagFilter && !job.tags.some((t) => t.toLowerCase().includes(tagFilter.toLowerCase()))) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchTitle = job.title.toLowerCase().includes(q);
      const matchCompany = job.company.toLowerCase().includes(q);
      const matchSkills = job.required_skills.some((s) => s.toLowerCase().includes(q));
      if (!matchTitle && !matchCompany && !matchSkills) return false;
    }
    return true;
  });

  const getStatusBadge = (status: SavedJobStatus) => {
    switch (status) {
      case 'saved':
        return <span className="bg-blue-100 text-blue-800 border border-blue-200 text-xs px-2.5 py-0.5 rounded-full font-bold uppercase tracking-wider">Saved</span>;
      case 'shortlisted':
        return <span className="bg-purple-100 text-purple-800 border border-purple-200 text-xs px-2.5 py-0.5 rounded-full font-bold uppercase tracking-wider">Shortlisted</span>;
      case 'ready_to_apply':
        return <span className="bg-amber-100 text-amber-900 border border-amber-300 text-xs px-2.5 py-0.5 rounded-full font-bold uppercase tracking-wider">Ready to Apply</span>;
      case 'applied':
        return <span className="bg-emerald-100 text-emerald-800 border border-emerald-300 text-xs px-2.5 py-0.5 rounded-full font-bold uppercase tracking-wider">Applied</span>;
      case 'archived':
        return <span className="bg-gray-100 text-gray-700 border border-gray-300 text-xs px-2.5 py-0.5 rounded-full font-bold uppercase tracking-wider">Archived</span>;
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Toast */}
      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 bg-gray-900 text-white px-4 py-2.5 rounded-xl shadow-lg text-sm font-medium flex items-center gap-2 animate-bounce">
          <span>✓</span> {toastMessage}
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-gray-200 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Link href="/career" className="text-sm font-medium text-blue-600 hover:underline">
              ← Career Hub
            </Link>
            <span className="text-gray-400">/</span>
            <span className="text-sm text-gray-500 font-medium">Saved & Application Ledgers</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">
            Saved Opportunities, Deduplication & Exclusion History
          </h1>
          <p className="text-sm text-gray-600 mt-1">
            Maintain independent ledgers for discoveries and applications (CAR-10), persistent exclusion blocklists (C8), and reviewable similarity deduplication (AT-004).
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/career/discovery"
            className="px-4 py-2 bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 font-medium text-sm rounded-lg shadow-sm transition"
          >
            ← Explore Job Discovery
          </Link>
          <Link
            href="/career/tailor"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-lg shadow-sm transition"
          >
            Tailor Application Artifacts
          </Link>
        </div>
      </div>

      {/* Primary Navigation Tabs */}
      <div className="flex border-b border-gray-200 space-x-8 text-sm font-semibold">
        <button
          onClick={() => setActiveTab('saved')}
          className={`pb-3 border-b-2 transition flex items-center gap-2 ${
            activeTab === 'saved'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          }`}
        >
          <span>⭐</span> Saved / Shortlist ({savedJobs.length})
        </button>
        <button
          onClick={() => setActiveTab('applications')}
          className={`pb-3 border-b-2 transition flex items-center gap-2 ${
            activeTab === 'applications'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          }`}
        >
          <span>📝</span> Applications Ledger ({applications.length})
        </button>
        <button
          onClick={() => setActiveTab('exclusions')}
          className={`pb-3 border-b-2 transition flex items-center gap-2 ${
            activeTab === 'exclusions'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          }`}
        >
          <span>🚫</span> Exclusion Blocklist ({exclusions.length})
        </button>
        <button
          onClick={() => setActiveTab('dedupe')}
          className={`pb-3 border-b-2 transition flex items-center gap-2 ${
            activeTab === 'dedupe'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          }`}
        >
          <span>🔍</span> Dedupe & Ambiguity Inspector
        </button>
      </div>

      {/* TAB 1: SAVED JOBS & SHORTLIST */}
      {activeTab === 'saved' && (
        <div className="space-y-6">
          {/* Controls Bar */}
          <div className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm flex flex-col md:flex-row items-center justify-between gap-4">
            {/* Status Pills */}
            <div className="flex flex-wrap items-center gap-1.5 w-full md:w-auto">
              {(['all', 'saved', 'shortlisted', 'ready_to_apply', 'applied', 'archived'] as const).map((st) => (
                <button
                  key={st}
                  onClick={() => setStatusFilter(st)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-semibold capitalize transition ${
                    statusFilter === st
                      ? 'bg-blue-600 text-white shadow-sm'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  }`}
                >
                  {st.replace('_', ' ')}
                </button>
              ))}
            </div>

            {/* Search and Tag filter */}
            <div className="flex items-center gap-3 w-full md:w-auto">
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Filter title, company, skill..."
                className="text-xs border border-gray-300 rounded-lg px-3 py-2 w-full md:w-56 focus:ring-2 focus:ring-blue-500 focus:outline-none"
              />
              <input
                type="text"
                value={tagFilter}
                onChange={(e) => setTagFilter(e.target.value)}
                placeholder="Tag filter..."
                className="text-xs border border-gray-300 rounded-lg px-3 py-2 w-full md:w-36 focus:ring-2 focus:ring-blue-500 focus:outline-none"
              />
            </div>
          </div>

          {/* Jobs List */}
          {filteredSavedJobs.length === 0 ? (
            <div className="p-12 text-center bg-white border border-gray-200 rounded-xl space-y-3">
              <p className="text-gray-500 text-sm">No saved jobs found matching your filters.</p>
              <Link
                href="/career/discovery"
                className="inline-block px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-lg shadow-sm"
              >
                Browse Job Discovery
              </Link>
            </div>
          ) : (
            <div className="space-y-4">
              {filteredSavedJobs.map((job) => (
                <div
                  key={job.id}
                  className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm space-y-4 hover:border-blue-300 transition"
                >
                  <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="text-base font-bold text-gray-900">{job.title}</h3>
                        <span className="text-xs font-semibold text-gray-700 bg-gray-100 px-2.5 py-0.5 rounded">
                          {job.company}
                        </span>
                        {getStatusBadge(job.status)}
                        <span className="text-xs text-amber-500 font-semibold flex items-center">
                          {'★'.repeat(job.priority)}
                          <span className="text-gray-300">{'★'.repeat(5 - job.priority)}</span>
                        </span>
                      </div>

                      <div className="flex items-center gap-4 text-xs text-gray-500 flex-wrap">
                        <span>📍 {job.location.raw_location}</span>
                        {job.compensation.is_disclosed && (
                          <span className="font-semibold text-gray-700">
                            💰 ${(job.compensation.min_amount || 0).toLocaleString()} - ${(job.compensation.max_amount || 0).toLocaleString()} {job.compensation.currency}
                          </span>
                        )}
                        <span>Saved: {new Date(job.saved_at).toLocaleDateString()}</span>
                        {job.applied_at && (
                          <span className="text-emerald-700 font-semibold">
                            ✓ Applied: {new Date(job.applied_at).toLocaleDateString()}
                          </span>
                        )}
                      </div>
                    </div>

                    {/* Quick Status Selector */}
                    <div className="flex items-center gap-2 shrink-0">
                      <label className="text-xs font-medium text-gray-600">Stage:</label>
                      <select
                        value={job.status}
                        onChange={(e) => handleUpdateStatus(job.id, e.target.value as SavedJobStatus)}
                        className="text-xs border border-gray-300 rounded-lg px-2.5 py-1.5 font-semibold bg-white focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="saved">Saved</option>
                        <option value="shortlisted">Shortlisted</option>
                        <option value="ready_to_apply">Ready to Apply</option>
                        <option value="applied">Applied</option>
                        <option value="archived">Archived</option>
                      </select>
                    </div>
                  </div>

                  {/* Skills & Tags */}
                  <div className="flex flex-wrap items-center gap-2 text-xs">
                    <span className="text-gray-400 font-medium">Skills:</span>
                    {job.required_skills.map((s, idx) => (
                      <span key={idx} className="bg-gray-100 text-gray-700 px-2 py-0.5 rounded font-medium text-[11px]">
                        {s}
                      </span>
                    ))}
                    {job.tags.length > 0 && (
                      <>
                        <span className="text-gray-400 font-medium ml-2">Tags:</span>
                        {job.tags.map((t, idx) => (
                          <span key={idx} className="bg-purple-50 text-purple-700 border border-purple-200 px-2 py-0.5 rounded font-semibold text-[11px]">
                            #{t}
                          </span>
                        ))}
                      </>
                    )}
                  </div>

                  {/* Candidate Notes */}
                  <div className="bg-slate-50 border border-slate-200 rounded-lg p-3 text-xs space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-gray-700">Personal Notes & Strategy:</span>
                      {editingNotesId === job.id ? (
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleSaveNotes(job.id)}
                            className="text-blue-600 hover:text-blue-800 font-bold"
                          >
                            Save
                          </button>
                          <button
                            onClick={() => setEditingNotesId(null)}
                            className="text-gray-500 hover:text-gray-700"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => {
                            setEditingNotesId(job.id);
                            setTempNotes(job.notes);
                          }}
                          className="text-blue-600 hover:text-blue-800 font-medium"
                        >
                          ✎ Edit Notes
                        </button>
                      )}
                    </div>

                    {editingNotesId === job.id ? (
                      <textarea
                        value={tempNotes}
                        onChange={(e) => setTempNotes(e.target.value)}
                        className="w-full text-xs p-2 border border-blue-400 rounded-md bg-white focus:outline-none"
                        rows={2}
                      />
                    ) : (
                      <p className="text-gray-600 italic">
                        {job.notes || 'No notes added yet. Click edit to record research or interview notes.'}
                      </p>
                    )}
                  </div>

                  {/* Card Actions Footer */}
                  <div className="border-t border-gray-100 pt-3 flex flex-wrap items-center justify-between gap-3 text-xs">
                    <div className="flex items-center gap-3">
                      <a
                        href={job.canonical_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:underline font-medium"
                      >
                        Canonical Job Posting ↗
                      </a>
                      {job.direct_apply_url && (
                        <a
                          href={job.direct_apply_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-emerald-600 hover:underline font-medium"
                        >
                          Direct Apply URL ↗
                        </a>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <Link
                        href={`/career/apply?jobId=${encodeURIComponent(job.job_id || job.id)}`}
                        className="px-3 py-1.5 bg-purple-50 hover:bg-purple-100 text-purple-700 font-semibold rounded-lg transition"
                      >
                        ✍️ Review & Apply →
                      </Link>
                      <Link
                        href={`/career/tailor?title=${encodeURIComponent(job.title)}&company=${encodeURIComponent(job.company)}&url=${encodeURIComponent(job.canonical_url)}`}
                        className="px-3 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-700 font-semibold rounded-lg transition"
                      >
                        Tailor Resume →
                      </Link>
                      <button
                        onClick={() => handleDeleteSavedJob(job.id)}
                        className="px-2.5 py-1.5 text-gray-400 hover:text-rose-600 font-medium transition"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* TAB 2: APPLICATIONS LEDGER (CAR-10, AT-004, AT-006) */}
      {activeTab === 'applications' && (
        <div className="space-y-6">
          <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-4 text-xs text-emerald-900 flex items-start gap-3">
            <span className="text-base">🛡️</span>
            <div>
              <span className="font-bold">Separated Application Submission Ledger: </span>
              In accordance with CAR-10, discovered opportunities and actual application submissions are tracked in distinct immutable ledgers. Once recorded here, deduplication rules permanently prevent duplicate submissions across runs and channels.
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-xs">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Company & Role</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Submission Mode</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Applied Date</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Status</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Audit / Confirmation Notes</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {applications.map((app) => (
                  <tr key={app.id} className="hover:bg-gray-50 transition">
                    <td className="px-4 py-3.5">
                      <div className="font-bold text-gray-900">{app.title}</div>
                      <div className="text-gray-500">{app.company}</div>
                      <a
                        href={app.canonical_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[11px] text-blue-600 hover:underline mt-0.5 block"
                      >
                        {app.canonical_url}
                      </a>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="bg-slate-100 text-slate-800 font-semibold px-2 py-1 rounded text-[11px] uppercase">
                        {app.submission_mode}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-gray-600 whitespace-nowrap">
                      {new Date(app.applied_at).toLocaleDateString()} {new Date(app.applied_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="bg-emerald-100 text-emerald-800 font-bold px-2.5 py-0.5 rounded-full uppercase text-[10px]">
                        {app.status}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-gray-600 italic">
                      {app.notes || '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: PERSISTENT EXCLUSION BLOCKLIST (CAR-10, C8) */}
      {activeTab === 'exclusions' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              Permanently block employers, specific job URLs, or fingerprints from ever appearing in discovery searches (C8 & JobFunnel pattern).
            </div>
            <button
              onClick={() => setShowAddExclusionModal(true)}
              className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white font-medium text-xs rounded-lg shadow-sm transition flex items-center gap-1.5"
            >
              <span>+</span> Add Exclusion Rule
            </button>
          </div>

          <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-xs">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Type</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Blocked Value</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Reason / Justification</th>
                  <th className="px-4 py-3 text-left font-bold text-gray-700 uppercase tracking-wider">Added Date</th>
                  <th className="px-4 py-3 text-right font-bold text-gray-700 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {exclusions.map((ex) => (
                  <tr key={ex.id} className="hover:bg-gray-50 transition">
                    <td className="px-4 py-3.5 font-bold uppercase text-[11px] text-rose-700">
                      {ex.type.replace('_', ' ')}
                    </td>
                    <td className="px-4 py-3.5 font-semibold text-gray-900">
                      {ex.value}
                    </td>
                    <td className="px-4 py-3.5 text-gray-600 italic">
                      {ex.reason}
                    </td>
                    <td className="px-4 py-3.5 text-gray-500 whitespace-nowrap">
                      {new Date(ex.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <button
                        onClick={() => handleDeleteExclusion(ex.id, ex.value)}
                        className="text-rose-600 hover:text-rose-800 font-bold hover:underline"
                      >
                        Unblock
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 4: DEDUPE & AMBIGUITY INSPECTOR (AT-004) */}
      {activeTab === 'dedupe' && (
        <div className="space-y-6">
          <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 text-xs text-blue-900 space-y-1">
            <div className="font-bold flex items-center gap-1.5">
              <span>🔬</span> Multi-Ledger Deduplication & Ambiguous Match Resolution (AT-004):
            </div>
            <div>
              Evaluate a potential job against discovery ledgers, past application submissions, and the persistent exclusion blocklist. Matches with &ge; 75% token similarity from the same company are flagged as <strong>Reviewable Ambiguous Matches</strong> rather than silently discarded.
            </div>
          </div>

          {/* Inspection Form */}
          <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm space-y-4">
            <h3 className="text-sm font-bold text-gray-900">Test Candidate Job for Deduplication</h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
              <div>
                <label className="block font-semibold text-gray-700 mb-1">Job Title</label>
                <input
                  type="text"
                  value={testJobTitle}
                  onChange={(e) => setTestJobTitle(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg p-2.5 focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block font-semibold text-gray-700 mb-1">Company</label>
                <input
                  type="text"
                  value={testJobCompany}
                  onChange={(e) => setTestJobCompany(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg p-2.5 focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block font-semibold text-gray-700 mb-1">Canonical URL</label>
                <input
                  type="text"
                  value={testJobURL}
                  onChange={(e) => setTestJobURL(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg p-2.5 focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <button
              onClick={handleRunDedupeInspection}
              disabled={isInspecting}
              className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs rounded-lg shadow-sm transition"
            >
              {isInspecting ? 'Evaluating Ledgers...' : 'Run Deduplication Audit'}
            </button>
          </div>

          {/* Inspection Result Card */}
          {dedupeInspectResult && (
            <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm space-y-4 text-xs">
              <div className="flex items-center justify-between border-b border-gray-200 pb-3">
                <div className="flex items-center gap-3">
                  <span className="text-sm font-bold text-gray-900">Audit Outcome:</span>
                  {dedupeInspectResult.action_taken === 'filtered_out' ? (
                    <span className="bg-rose-100 text-rose-800 px-3 py-1 rounded-full font-bold uppercase border border-rose-200">
                      🚫 Filtered Out (Persistent Exclusion)
                    </span>
                  ) : dedupeInspectResult.action_taken === 'linked' ? (
                    <span className="bg-purple-100 text-purple-800 px-3 py-1 rounded-full font-bold uppercase border border-purple-200">
                      🔗 Linked (Exact Duplicate Detected)
                    </span>
                  ) : dedupeInspectResult.action_taken === 'flagged_for_review' ? (
                    <span className="bg-amber-100 text-amber-900 px-3 py-1 rounded-full font-bold uppercase border border-amber-300">
                      ⚠️ Ambiguous Match (Flagged for Human Review - AT-004)
                    </span>
                  ) : (
                    <span className="bg-emerald-100 text-emerald-800 px-3 py-1 rounded-full font-bold uppercase border border-emerald-200">
                      🟢 Eligible (Retained for Candidate)
                    </span>
                  )}
                </div>

                <div className="text-gray-500 font-mono text-[11px]">
                  Fingerprint: {dedupeInspectResult.fingerprint}
                </div>
              </div>

              {/* Match Details */}
              {dedupeInspectResult.matches.length > 0 ? (
                <div className="space-y-3">
                  <h4 className="font-bold text-gray-800">Identified Ledger Matches:</h4>
                  {dedupeInspectResult.matches.map((m, idx) => (
                    <div
                      key={idx}
                      className={`p-3 rounded-lg border text-xs space-y-1 ${
                        m.match_level === 'ambiguous'
                          ? 'bg-amber-50/50 border-amber-200 text-amber-900'
                          : 'bg-slate-50 border-slate-200 text-slate-800'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="font-bold flex items-center gap-2">
                          <span className="uppercase text-[10px] bg-white px-2 py-0.5 rounded border">
                            Ledger: {m.target_ledger}
                          </span>
                          <span>Level: {m.match_level}</span>
                          {m.similarity_score < 1.0 && (
                            <span className="text-amber-700">({Math.round(m.similarity_score * 100)}% similarity)</span>
                          )}
                        </div>
                        {m.is_reviewable && (
                          <span className="text-xs font-bold text-amber-800 bg-amber-200/60 px-2 py-0.5 rounded">
                            Review Required
                          </span>
                        )}
                      </div>

                      <p className="text-gray-700">{m.reason}</p>

                      {m.is_reviewable && (
                        <div className="pt-2 flex items-center gap-2">
                          <button
                            onClick={() => showToast('Match marked as duplicate and linked')}
                            className="px-3 py-1 bg-amber-600 hover:bg-amber-700 text-white rounded font-medium text-[11px]"
                          >
                            Confirm Duplicate (Link)
                          </button>
                          <button
                            onClick={() => showToast('Match confirmed as separate opportunity')}
                            className="px-3 py-1 bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 rounded font-medium text-[11px]"
                          >
                            Mark as Different Job
                          </button>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-emerald-700 font-medium py-2">
                  ✓ No conflicts found in saved jobs, application submissions, or exclusion blocklists.
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Add Exclusion Modal */}
      {showAddExclusionModal && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl max-w-md w-full p-6 space-y-4 shadow-xl">
            <h3 className="text-base font-bold text-gray-900">Add Persistent Exclusion Rule</h3>
            <p className="text-xs text-gray-600">
              Never see jobs from this entity again. Excluded jobs are automatically dropped from searches and deduplication.
            </p>

            <form onSubmit={handleAddExclusion} className="space-y-3 text-xs">
              <div>
                <label className="block font-semibold text-gray-700 mb-1">Exclusion Type</label>
                <select
                  value={newExclType}
                  onChange={(e) => setNewExclType(e.target.value as ExclusionType)}
                  className="w-full border border-gray-300 rounded-lg p-2"
                >
                  <option value="company_name">Company Name (Block whole company)</option>
                  <option value="job_canonical_url">Canonical Job URL (Block specific link)</option>
                  <option value="job_fingerprint">Job Fingerprint</option>
                </select>
              </div>

              <div>
                <label className="block font-semibold text-gray-700 mb-1">Target Value</label>
                <input
                  type="text"
                  required
                  placeholder={newExclType === 'company_name' ? 'e.g. Acme Corp' : 'https://...'}
                  value={newExclValue}
                  onChange={(e) => setNewExclValue(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg p-2 focus:ring-2 focus:ring-rose-500"
                />
              </div>

              <div>
                <label className="block font-semibold text-gray-700 mb-1">Reason / Note</label>
                <input
                  type="text"
                  placeholder="e.g. Negative glassdoor reviews, third-party spam"
                  value={newExclReason}
                  onChange={(e) => setNewExclReason(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg p-2"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAddExclusionModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 font-semibold"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded-lg font-semibold"
                >
                  Add Exclusion
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
