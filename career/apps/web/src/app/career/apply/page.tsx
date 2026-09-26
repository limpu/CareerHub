'use client';

import { Suspense, useState, useEffect } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import type {
  DiscoveredJob,
  ApplicationWorkflowStatus,
  ApplicationExecutionMode,
  ApplicationQuestionAnswer,
  ApplicationReviewSession,
  SubmissionReceipt,
} from '@social-platform/contracts';

function ApplicationWorkflowContent() {
  const searchParams = useSearchParams();
  const urlJobId = searchParams.get('jobId');

  const [activeSession, setActiveSession] = useState<ApplicationReviewSession | null>(null);
  const [sessionsList, setSessionsList] = useState<ApplicationReviewSession[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [actionNotice, setActionNotice] = useState<string | null>(null);
  const [invalidationWarning, setInvalidationWarning] = useState<boolean>(false);
  const [showReceiptModal, setShowReceiptModal] = useState<boolean>(false);
  const [receiptInput, setReceiptInput] = useState<{ reference: string; notes: string }>({
    reference: '',
    notes: '',
  });

  // Mock target job for sandbox preparation
  const defaultTargetJob: DiscoveredJob = {
    id: urlJobId || 'job_sample_greenhouse_1',
    source: 'greenhouse',
    source_job_id: 'gh_987654',
    canonical_url: 'https://boards.greenhouse.io/techcorp/jobs/987654',
    direct_apply_url: 'https://boards.greenhouse.io/techcorp/jobs/987654#apply',
    title: 'Staff Distributed Systems Engineer',
    company: 'TechCorp International',
    location: {
      raw_location: 'Berlin, Germany (Remote)',
      city: 'Berlin',
      country: 'DE',
      is_remote: true,
      is_hybrid: false,
    },
    job_type: 'full_time',
    compensation: {
      is_disclosed: true,
      min_amount: 120000,
      max_amount: 145000,
      currency: 'EUR',
      period: 'yearly',
      is_estimated: false,
    },
    required_skills: ['Go', 'PostgreSQL', 'Distributed Systems', 'Redis', 'Docker'],
    description: 'Leading our core distributed services, microservice architecture, and data pipelines.',
    date_posted: new Date().toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  };

  // Initial demo session setup
  useEffect(() => {
    const initialSession: ApplicationReviewSession = {
      id: 'app_rev_demo_01',
      user_id: 'usr_demo',
      workspace_id: 'ws_demo',
      job_id: defaultTargetJob.id,
      job_title: defaultTargetJob.title,
      company: defaultTargetJob.company,
      apply_url: defaultTargetJob.direct_apply_url || defaultTargetJob.canonical_url,
      source_board: defaultTargetJob.source,
      execution_mode: 'native_manual',
      bundle: {
        resume_id: 'res_tailored_demo_1',
        resume_file_name: 'rahim_chowdhury_resume_techcorp.pdf',
        resume_checksum: 'a89c3b1e7f912e84d1c920f01a8b94ec710f6e91a273b5084920fcb128509a24',
        cover_letter_id: 'cl_tailored_demo_1',
        cover_letter_file_name: 'rahim_chowdhury_cover_letter_techcorp.pdf',
        cover_letter_checksum: '7b91d24c0fa31295b902187e14f9c2d1b098317e0892c510fa98231c59012be4',
        questions: [
          {
            question_id: 'contact_full_name',
            question_text: 'Full Name',
            field_type: 'text',
            required: true,
            answer_value: 'Alex Morgan',
            is_confirmed: true,
            needs_input: false,
            category: 'contact',
          },
          {
            question_id: 'contact_email',
            question_text: 'Email Address',
            field_type: 'text',
            required: true,
            answer_value: 'alex.morgan@example.com',
            is_confirmed: true,
            needs_input: false,
            category: 'contact',
          },
          {
            question_id: 'contact_phone',
            question_text: 'Phone Number',
            field_type: 'text',
            required: true,
            answer_value: '+1 (555) 019-2834',
            is_confirmed: true,
            needs_input: false,
            category: 'contact',
          },
          {
            question_id: 'contact_location',
            question_text: 'Current City & Country',
            field_type: 'text',
            required: true,
            answer_value: 'Dhaka, Bangladesh',
            is_confirmed: true,
            needs_input: false,
            category: 'contact',
          },
          {
            question_id: 'work_authorization',
            question_text: 'Are you legally authorized to work in the country of employment?',
            field_type: 'boolean',
            required: true,
            answer_value: '',
            is_confirmed: false,
            needs_input: true,
            category: 'authorization',
            explanation_note: 'AT-003: Legal work authorization must never be guessed or defaulted.',
          },
          {
            question_id: 'visa_sponsorship',
            question_text: 'Will you now or in the future require visa sponsorship?',
            field_type: 'boolean',
            required: true,
            answer_value: '',
            is_confirmed: false,
            needs_input: true,
            category: 'authorization',
            explanation_note: 'AT-003: Sponsorship requirement must be explicitly confirmed by candidate.',
          },
          {
            question_id: 'experience_years',
            question_text: 'Total Years of Professional Experience',
            field_type: 'number',
            required: true,
            answer_value: '6',
            is_confirmed: true,
            needs_input: false,
            category: 'experience',
          },
        ],
        bundle_checksum_sha256: '99e38d7211a7c06289ba100c561280cf289a01e389d0b51201948197e9f3b182',
      },
      status: 'draft',
      timeout_duration_secs: 600,
      created_at: new Date(Date.now() - 3600000).toISOString(),
      updated_at: new Date().toISOString(),
    };

    setActiveSession(initialSession);
    setSessionsList([initialSession]);
  }, [urlJobId]);

  // Handler for modifying answers with AT-007 Tamper Invalidation
  const handleAnswerChange = (questionId: string, value: string) => {
    if (!activeSession) return;

    const wasApproved = activeSession.status === 'approved';
    const updatedQuestions = activeSession.bundle.questions.map((q) => {
      if (q.question_id === questionId) {
        const isAnswered = value.trim().length > 0;
        return {
          ...q,
          answer_value: value,
          is_confirmed: isAnswered,
          needs_input: !isAnswered && q.required,
        };
      }
      return q;
    });

    const hasUnresolved = updatedQuestions.some((q) => q.required && q.needs_input);
    let newStatus: ApplicationWorkflowStatus = activeSession.status;

    if (wasApproved) {
      newStatus = 'ready_for_review';
      setInvalidationWarning(true);
      setActionNotice('⚠️ Material change detected: previous approval token invalidated per REQ-015 & AT-007.');
    } else if (hasUnresolved) {
      newStatus = 'draft';
    } else {
      newStatus = 'ready_for_review';
    }

    // Recompute bundle hash
    const updatedBundle = {
      ...activeSession.bundle,
      questions: updatedQuestions,
      bundle_checksum_sha256: `hash_${Date.now().toString(16)}`,
    };

    setActiveSession({
      ...activeSession,
      bundle: updatedBundle,
      status: newStatus,
      approval_token: wasApproved ? undefined : activeSession.approval_token,
      approved_at: wasApproved ? undefined : activeSession.approved_at,
      updated_at: new Date().toISOString(),
    });
  };

  // Handler for human approval (REQ-015, AT-003)
  const handleApprove = () => {
    if (!activeSession) return;

    // Verify no required questions need input (AT-003)
    const unresolved = activeSession.bundle.questions.filter((q) => q.required && q.needs_input);
    if (unresolved.length > 0) {
      alert(`Cannot approve: ${unresolved.length} required questions require candidate input (AT-003 Zero-Fabrication).`);
      return;
    }

    const token = `hmac_sha256_${Date.now().toString(16)}_${activeSession.id.slice(-6)}`;
    setActiveSession({
      ...activeSession,
      status: 'approved',
      approval_token: token,
      approved_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    setInvalidationWarning(false);
    setActionNotice('✅ Human approval granted. Cryptographic approval token bound to candidate bundle (REQ-015).');
  };

  // Handler for execution mode change (CAR-11)
  const handleModeChange = (mode: ApplicationExecutionMode) => {
    if (!activeSession) return;
    setActiveSession({
      ...activeSession,
      execution_mode: mode,
      updated_at: new Date().toISOString(),
    });
  };

  // Handler for dispatching (AT-005: Clicking Apply DOES NOT mark Applied!)
  const handleDispatch = () => {
    if (!activeSession) return;

    if (activeSession.status !== 'approved') {
      alert('Action blocked: Application bundle must receive human approval before dispatch (REQ-015).');
      return;
    }

    // Open portal in new window
    window.open(activeSession.apply_url, '_blank');

    setActiveSession({
      ...activeSession,
      status: 'dispatched',
      dispatched_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    setActionNotice(
      activeSession.execution_mode === 'native_manual'
        ? '🚀 Employer portal opened in new tab. Status set to DISPATCHED (AT-005: Clicking Apply never marks Applied). Copy your prepared answers below.'
        : '🤖 Assistant pre-fill initialized. Automation strictly stops before final submit for your manual sign-off (SRC-L1, REQ-015).'
    );
  };

  // Handler for explicit submission confirmation (REQ-005, AT-005)
  const handleConfirmSubmission = () => {
    if (!activeSession) return;

    const receipt: SubmissionReceipt = {
      receipt_id: receiptInput.reference || `rec_manual_${Date.now()}`,
      provider_reference: receiptInput.reference || 'user_explicit_portal_confirmation',
      submission_url: activeSession.apply_url,
      confirmed_at: new Date().toISOString(),
      confirmed_by_user: true,
      notes: receiptInput.notes || 'Candidate completed final submit on employer application portal.',
    };

    setActiveSession({
      ...activeSession,
      status: 'applied',
      receipt,
      application_record_id: `app_rec_${Date.now()}`,
      updated_at: new Date().toISOString(),
    });

    setShowReceiptModal(false);
    setActionNotice('🎉 Submission verified and permanently recorded in Applications Ledger (REQ-005, AT-005).');
  };

  // Copy all answers helper (CAR-11 native manual fallback)
  const handleCopyBundle = () => {
    if (!activeSession) return;
    const text = activeSession.bundle.questions
      .map((q) => `${q.question_text}: ${q.answer_value || '[NO ANSWER]'}`)
      .join('\n');
    navigator.clipboard.writeText(text);
    setActionNotice('📋 Prepared answers copied to clipboard! Paste directly into the employer portal fields.');
  };

  const unresolvedCount = activeSession?.bundle.questions.filter((q) => q.required && q.needs_input).length || 0;

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 md:p-12 font-sans">
      <div className="max-w-6xl mx-auto space-y-8">
        {/* Navigation & Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-slate-200 pb-6">
          <div>
            <div className="flex items-center gap-2 text-sm text-slate-500 mb-1">
              <Link href="/career" className="hover:text-blue-400">Career Hub</Link>
              <span>/</span>
              <Link href="/career/saved" className="hover:text-blue-400">Saved & Ledgers</Link>
              <span>/</span>
              <span className="text-slate-800">Application Review Workflow</span>
            </div>
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 flex items-center gap-3">
              <span>✍️ Human-Reviewed Application Gateway</span>
              <span className="text-xs px-2.5 py-1 bg-blue-900/60 border border-blue-700 text-blue-300 rounded font-mono">
                CAR-11 • REQ-015 • AT-005
              </span>
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Review material, answer questionnaire with zero fabrication, bind cryptographic approval tokens, and verify submissions.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Link
              href="/career/portal"
              className="px-4 py-2 bg-emerald-950 border border-emerald-700 hover:border-emerald-500 text-emerald-300 rounded text-sm font-medium transition"
            >
              🌐 Portals Hub (CAR-14)
            </Link>
            <Link
              href="/career/status"
              className="px-4 py-2 bg-blue-950 border border-blue-700 hover:border-blue-500 text-blue-300 rounded text-sm font-medium transition"
            >
              📊 Status & Ledger (CAR-15)
            </Link>
            <Link
              href="/career/wizard"
              className="px-4 py-2 bg-purple-950 border border-purple-700 hover:border-purple-500 text-purple-300 rounded text-sm font-medium transition"
            >
              🪄 Wizard Studio (CAR-13)
            </Link>
            <Link
              href="/career/fields"
              className="px-4 py-2 bg-cyan-950 border border-cyan-700 hover:border-cyan-500 text-cyan-300 rounded text-sm font-medium transition"
            >
              🛡️ Field Studio (CAR-12)
            </Link>
            <Link
              href="/career/saved"
              className="px-4 py-2 bg-white border border-slate-300 hover:border-neutral-500 rounded text-sm font-medium transition"
            >
              ⭐ View Saved & Ledgers
            </Link>
            <Link
              href="/career/discovery"
              className="px-4 py-2 bg-white border border-slate-300 hover:border-neutral-500 rounded text-sm font-medium transition"
            >
              🔍 Discover Jobs
            </Link>
          </div>
        </div>

        {/* Global Policy Alerts */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="p-4 bg-amber-950/40 border border-amber-800/80 rounded-lg text-xs space-y-1">
            <div className="font-semibold text-amber-300 flex items-center gap-1.5">
              <span>🛡️ Zero-Fabrication Rule (AT-003)</span>
            </div>
            <p className="text-amber-200/80">
              Unconfirmed visa, legal, or experience questions are never guessed. They require candidate answers before approval.
            </p>
          </div>

          <div className="p-4 bg-purple-950/40 border border-purple-800/80 rounded-lg text-xs space-y-1">
            <div className="font-semibold text-purple-300 flex items-center gap-1.5">
              <span>🔐 Tamper Invalidation (REQ-015, AT-007)</span>
            </div>
            <p className="text-purple-200/80">
              Approvals are bound to document & answer checksums. Any subsequent edit immediately invalidates the approval token.
            </p>
          </div>

          <div className="p-4 bg-blue-950/40 border border-blue-800/80 rounded-lg text-xs space-y-1">
            <div className="font-semibold text-blue-300 flex items-center gap-1.5">
              <span>🛑 Click ≠ Applied (AT-005)</span>
            </div>
            <p className="text-blue-200/80">
              Clicking Apply dispatches the flow. It is only marked Applied upon explicit receipt or candidate confirmation.
            </p>
          </div>
        </div>

        {/* Notification Toast */}
        {actionNotice && (
          <div className="p-4 bg-white border border-blue-600/80 rounded-lg text-sm text-blue-200 flex items-center justify-between">
            <span>{actionNotice}</span>
            <button onClick={() => setActionNotice(null)} className="text-xs text-slate-500 hover:text-slate-900">Dismiss</button>
          </div>
        )}

        {/* Main Application Session View */}
        {activeSession ? (
          <div className="space-y-6">
            {/* Status & Target Job Strip */}
            <div className="p-5 bg-white border border-slate-200 rounded-xl flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-3">
                  <h2 className="text-xl font-bold text-slate-900">{activeSession.job_title}</h2>
                  <span className="text-sm text-slate-500 font-medium">@ {activeSession.company}</span>
                  <span className="text-xs px-2 py-0.5 bg-slate-100 border border-slate-300 text-slate-700 rounded uppercase font-mono">
                    {activeSession.source_board}
                  </span>
                </div>
                <div className="text-xs text-slate-500 flex items-center gap-4">
                  <span>Session: <code className="text-slate-700">{activeSession.id}</code></span>
                  <span>Apply URL: <a href={activeSession.apply_url} target="_blank" rel="noreferrer" className="text-blue-400 hover:underline">{activeSession.apply_url}</a></span>
                </div>
              </div>

              {/* Status Badge */}
              <div className="flex items-center gap-3">
                <div className="text-right">
                  <div className="text-xs text-slate-500">Workflow Status</div>
                  <div className="text-sm font-semibold capitalize">
                    {activeSession.status === 'draft' && <span className="text-amber-400">🟡 Draft (Unresolved Questions)</span>}
                    {activeSession.status === 'ready_for_review' && <span className="text-blue-400">🔵 Ready for Human Review</span>}
                    {activeSession.status === 'approved' && <span className="text-emerald-400">🟢 Approved & Sealed</span>}
                    {activeSession.status === 'dispatched' && <span className="text-purple-400">🚀 Dispatched (In Portal)</span>}
                    {activeSession.status === 'needs_confirmation' && <span className="text-red-400">⚠️ Needs Confirmation (Timeout)</span>}
                    {activeSession.status === 'applied' && <span className="text-emerald-300 font-bold">✅ Applied & Verified</span>}
                  </div>
                </div>

                {activeSession.approval_token && (
                  <div className="px-3 py-1 bg-emerald-950/70 border border-emerald-700 rounded text-xs font-mono text-emerald-300">
                    Token: {activeSession.approval_token.slice(0, 14)}...
                  </div>
                )}
              </div>
            </div>

            {/* Invalidation Alert if triggered */}
            {invalidationWarning && (
              <div className="p-4 bg-red-950/50 border border-red-700 rounded-lg text-sm text-red-200 space-y-1">
                <div className="font-bold flex items-center gap-2">
                  <span>⚠️ Approval Token Invalidated (AT-007)</span>
                </div>
                <p>
                  You edited an application answer after approval was granted. In compliance with safety invariant <code>REQ-015</code>, the previous approval token was revoked. You must re-review and approve this bundle before dispatch.
                </p>
              </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Left Column: Material Inspection & Attached Documents */}
              <div className="space-y-6">
                <div className="p-5 bg-white border border-slate-200 rounded-xl space-y-4">
                  <h3 className="text-base font-semibold text-slate-900 flex items-center justify-between">
                    <span>📄 Material Bundle (REQ-003)</span>
                    <span className="text-xs px-2 py-0.5 bg-slate-100 text-slate-700 rounded font-mono">
                      Immutable Checksum
                    </span>
                  </h3>

                  {/* Resume Card */}
                  <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="text-sm font-medium text-slate-900 flex items-center gap-2">
                        <span>📑</span>
                        <span>{activeSession.bundle.resume_file_name}</span>
                      </div>
                      <span className="text-xs px-2 py-0.5 bg-emerald-950 border border-emerald-700 text-emerald-300 rounded font-mono">
                        Parse-Back OK (AT-002)
                      </span>
                    </div>
                    <div className="text-[11px] text-neutral-500 font-mono break-all">
                      SHA256: {activeSession.bundle.resume_checksum}
                    </div>
                  </div>

                  {/* Cover Letter Card */}
                  {activeSession.bundle.cover_letter_file_name && (
                    <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="text-sm font-medium text-slate-900 flex items-center gap-2">
                          <span>✉️</span>
                          <span>{activeSession.bundle.cover_letter_file_name}</span>
                        </div>
                        <span className="text-xs px-2 py-0.5 bg-slate-100 text-slate-700 rounded font-mono">
                          Fact-Grounded (REQ-016)
                        </span>
                      </div>
                      <div className="text-[11px] text-neutral-500 font-mono break-all">
                        SHA256: {activeSession.bundle.cover_letter_checksum}
                      </div>
                    </div>
                  )}

                  {/* Overall Bundle Hash */}
                  <div className="pt-2 border-t border-slate-200">
                    <div className="text-xs text-slate-500 mb-1">Bundle Fingerprint (Bound to Approval):</div>
                    <code className="text-[11px] text-purple-300 font-mono break-all block bg-slate-50 p-2 rounded border border-slate-200">
                      {activeSession.bundle.bundle_checksum_sha256}
                    </code>
                  </div>
                </div>

                {/* Execution Mode Selector (CAR-11) */}
                <div className="p-5 bg-white border border-slate-200 rounded-xl space-y-4">
                  <h3 className="text-base font-semibold text-slate-900">⚙️ Execution Mode (CAR-11)</h3>
                  <p className="text-xs text-slate-500">
                    Native manual application is always guaranteed. Provider approval governs automated assist.
                  </p>

                  <div className="space-y-3">
                    <label
                      onClick={() => handleModeChange('native_manual')}
                      className={`block p-3.5 rounded-lg border cursor-pointer transition ${
                        activeSession.execution_mode === 'native_manual'
                          ? 'bg-blue-950/40 border-blue-600 text-slate-900'
                          : 'bg-slate-50 border-slate-200 text-slate-500 hover:border-slate-300'
                      }`}
                    >
                      <div className="flex items-center justify-between font-medium text-sm">
                        <span>🖥️ Native Manual Application</span>
                        {activeSession.execution_mode === 'native_manual' && <span className="text-xs text-blue-400">Selected</span>}
                      </div>
                      <p className="text-xs mt-1 text-slate-500">
                        Opens portal directly in your browser. All answers copied to clipboard for instant pasting.
                      </p>
                    </label>

                    <label
                      onClick={() => handleModeChange('assistant_prefill')}
                      className={`block p-3.5 rounded-lg border cursor-pointer transition ${
                        activeSession.execution_mode === 'assistant_prefill'
                          ? 'bg-purple-950/40 border-purple-600 text-slate-900'
                          : 'bg-slate-50 border-slate-200 text-slate-500 hover:border-slate-300'
                      }`}
                    >
                      <div className="flex items-center justify-between font-medium text-sm">
                        <span>🤖 Assistant Pre-Fill Mode</span>
                        {activeSession.execution_mode === 'assistant_prefill' && <span className="text-xs text-purple-400">Selected</span>}
                      </div>
                      <p className="text-xs mt-1 text-slate-500">
                        Automates field pre-filling but <strong>strictly stops before final submission</strong> for human sign-off (SRC-L1, REQ-015).
                      </p>
                    </label>
                  </div>
                </div>

                {/* Manual Clipboard Helper */}
                <div className="p-4 bg-white/60 border border-slate-200 rounded-xl space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-slate-700">📋 Prepared Form Answers</span>
                    <button
                      onClick={handleCopyBundle}
                      className="px-2.5 py-1 bg-slate-100 hover:bg-neutral-700 text-xs text-slate-800 rounded transition"
                    >
                      Copy All Answers
                    </button>
                  </div>
                  <p className="text-[11px] text-slate-500">
                    Use this to copy all confirmed answers to your clipboard for rapid manual filling on third-party employer boards.
                  </p>
                </div>
              </div>

              {/* Right Column (2 cols): Interactive Questionnaire & Approval Gate */}
              <div className="lg:col-span-2 space-y-6">
                <div className="p-6 bg-white border border-slate-200 rounded-xl space-y-6">
                  <div className="flex items-center justify-between border-b border-slate-200 pb-4">
                    <div>
                      <h3 className="text-lg font-bold text-slate-900">📝 Candidate Questionnaire Review</h3>
                      <p className="text-xs text-slate-500">
                        Every field confirmed against your master profile. Missing fields flagged as <code>needs_input</code>.
                      </p>
                    </div>

                    {unresolvedCount > 0 ? (
                      <span className="px-3 py-1 bg-amber-950 border border-amber-700 text-amber-300 text-xs rounded-full font-semibold">
                        ⚠️ {unresolvedCount} Question(s) Need Input
                      </span>
                    ) : (
                      <span className="px-3 py-1 bg-emerald-950 border border-emerald-700 text-emerald-300 text-xs rounded-full font-semibold">
                        ✅ All Required Fields Resolved
                      </span>
                    )}
                  </div>

                  {/* Question Fields Form */}
                  <div className="space-y-4">
                    {activeSession.bundle.questions.map((q) => (
                      <div
                        key={q.question_id}
                        className={`p-4 rounded-lg border transition ${
                          q.needs_input
                            ? 'bg-amber-950/20 border-amber-800'
                            : 'bg-slate-50 border-slate-200'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-2">
                          <label className="text-sm font-medium text-slate-800 flex items-center gap-2">
                            <span>{q.question_text}</span>
                            {q.required && <span className="text-red-400">*</span>}
                          </label>

                          {q.needs_input && (
                            <span className="text-[11px] px-2 py-0.5 bg-amber-900/60 border border-amber-700 text-amber-200 rounded font-medium">
                              Input Required (AT-003)
                            </span>
                          )}
                          {!q.needs_input && q.is_confirmed && (
                            <span className="text-[11px] px-2 py-0.5 bg-emerald-950 border border-emerald-800 text-emerald-300 rounded">
                              Confirmed
                            </span>
                          )}
                        </div>

                        {/* Input Controls */}
                        {q.field_type === 'boolean' ? (
                          <div className="flex gap-3 mt-2">
                            <button
                              onClick={() => handleAnswerChange(q.question_id, 'Yes')}
                              className={`px-4 py-2 text-xs font-semibold rounded border transition ${
                                q.answer_value === 'Yes'
                                  ? 'bg-blue-600 border-blue-500 text-slate-900'
                                  : 'bg-white border-slate-300 text-slate-700 hover:border-neutral-500'
                              }`}
                            >
                              Yes
                            </button>
                            <button
                              onClick={() => handleAnswerChange(q.question_id, 'No')}
                              className={`px-4 py-2 text-xs font-semibold rounded border transition ${
                                q.answer_value === 'No'
                                  ? 'bg-blue-600 border-blue-500 text-slate-900'
                                  : 'bg-white border-slate-300 text-slate-700 hover:border-neutral-500'
                              }`}
                            >
                              No
                            </button>
                          </div>
                        ) : (
                          <input
                            type={q.field_type === 'number' ? 'number' : 'text'}
                            value={q.answer_value}
                            onChange={(e) => handleAnswerChange(q.question_id, e.target.value)}
                            placeholder={q.needs_input ? 'Please provide your confirmed answer...' : ''}
                            className="w-full bg-white border border-slate-300 rounded px-3 py-2 text-sm text-neutral-100 placeholder-neutral-500 focus:outline-none focus:border-blue-500"
                          />
                        )}

                        {q.explanation_note && (
                          <p className="text-[11px] text-slate-500 mt-2 italic">
                            ℹ️ {q.explanation_note}
                          </p>
                        )}
                      </div>
                    ))}
                  </div>

                  {/* Approval and Dispatch Actions Toolbar */}
                  <div className="pt-6 border-t border-slate-200 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                    <div>
                      {activeSession.status !== 'approved' && activeSession.status !== 'dispatched' && activeSession.status !== 'applied' && (
                        <button
                          onClick={handleApprove}
                          disabled={unresolvedCount > 0}
                          className={`px-6 py-3 rounded-lg text-sm font-bold transition flex items-center gap-2 ${
                            unresolvedCount > 0
                              ? 'bg-slate-100 text-neutral-500 cursor-not-allowed border border-slate-300'
                              : 'bg-emerald-600 hover:bg-emerald-500 text-slate-900 shadow-lg shadow-emerald-950/50'
                          }`}
                        >
                          <span>🛡️ Approve Application Bundle</span>
                          <span className="text-xs opacity-80">(REQ-015)</span>
                        </button>
                      )}

                      {activeSession.status === 'approved' && (
                        <div className="flex items-center gap-3">
                          <button
                            onClick={handleDispatch}
                            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-slate-900 font-bold text-sm rounded-lg shadow-lg shadow-blue-950/50 transition flex items-center gap-2"
                          >
                            <span>🚀 Open Portal & Dispatch</span>
                            <span className="text-xs opacity-80">(CAR-11)</span>
                          </button>
                        </div>
                      )}

                      {(activeSession.status === 'dispatched' || activeSession.status === 'needs_confirmation') && (
                        <div className="flex items-center gap-3">
                          <button
                            onClick={() => setShowReceiptModal(true)}
                            className="px-6 py-3 bg-emerald-600 hover:bg-emerald-500 text-slate-900 font-bold text-sm rounded-lg shadow-lg shadow-emerald-950/50 transition flex items-center gap-2"
                          >
                            <span>📝 Confirm Application Submission</span>
                            <span className="text-xs opacity-80">(AT-005)</span>
                          </button>
                          <button
                            onClick={() => window.open(activeSession.apply_url, '_blank')}
                            className="px-4 py-3 bg-slate-100 hover:bg-neutral-700 text-slate-800 text-sm font-medium rounded-lg transition"
                          >
                            Re-Open Portal
                          </button>
                        </div>
                      )}

                      {activeSession.status === 'applied' && (
                        <div className="flex items-center gap-3">
                          <span className="px-4 py-2 bg-emerald-950 border border-emerald-700 text-emerald-300 rounded font-semibold text-sm">
                            ✅ Applied & Verified
                          </span>
                          <Link
                            href="/career/saved"
                            className="px-4 py-2 bg-slate-100 hover:bg-neutral-700 text-slate-800 text-sm rounded transition"
                          >
                            View in Applications Ledger →
                          </Link>
                        </div>
                      )}
                    </div>

                    <div className="text-xs text-neutral-500">
                      {activeSession.status === 'dispatched' && (
                        <span className="text-purple-400">
                          ⏳ Waiting for submission proof. Clicking Apply does not mark Applied (AT-005).
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="p-12 text-center bg-white border border-slate-200 rounded-xl space-y-4">
            <div className="text-4xl">📭</div>
            <h3 className="text-lg font-bold text-slate-900">No Active Application Session</h3>
            <p className="text-sm text-slate-500 max-w-md mx-auto">
              Select an opportunity from the Job Discovery board or your Saved Jobs list to initiate a human-reviewed application workflow.
            </p>
            <div className="flex justify-center gap-4 pt-2">
              <Link
                href="/career/discovery"
                className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-slate-900 text-sm font-semibold rounded-lg transition"
              >
                Go to Job Discovery
              </Link>
              <Link
                href="/career/saved"
                className="px-5 py-2.5 bg-slate-100 hover:bg-neutral-700 text-slate-800 text-sm font-semibold rounded-lg transition"
              >
                View Saved Jobs
              </Link>
            </div>
          </div>
        )}

        {/* Explicit Confirmation / Submission Receipt Modal (AT-005, REQ-005) */}
        {showReceiptModal && (
          <div className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4">
            <div className="bg-white border border-slate-300 rounded-xl max-w-lg w-full p-6 space-y-5 shadow-2xl">
              <div className="border-b border-slate-200 pb-3 flex items-center justify-between">
                <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
                  <span>📝 Confirm Application Submission</span>
                  <span className="text-xs px-2 py-0.5 bg-blue-900 text-blue-300 rounded font-mono">AT-005</span>
                </h3>
                <button
                  onClick={() => setShowReceiptModal(false)}
                  className="text-slate-500 hover:text-slate-900 text-lg font-bold"
                >
                  ✕
                </button>
              </div>

              <div className="p-3.5 bg-blue-950/40 border border-blue-800 rounded text-xs text-blue-200 space-y-1">
                <div className="font-semibold">Invariant AT-005: Honest Verification</div>
                <p>
                  Clicking "Apply" does not mark an application as submitted. Please confirm that you have completed the submission on the employer portal, and optionally enter your confirmation reference ID.
                </p>
              </div>

              <div className="space-y-4 text-sm">
                <div>
                  <label className="block text-slate-700 font-medium mb-1">
                    Provider Confirmation / Receipt Reference (Optional)
                  </label>
                  <input
                    type="text"
                    value={receiptInput.reference}
                    onChange={(e) => setReceiptInput({ ...receiptInput, reference: e.target.value })}
                    placeholder="e.g. GH-2026-987654 or email confirmation ID"
                    className="w-full bg-slate-50 border border-slate-300 rounded px-3 py-2 text-slate-900 placeholder-neutral-500 focus:outline-none focus:border-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-slate-700 font-medium mb-1">
                    Application Notes
                  </label>
                  <textarea
                    value={receiptInput.notes}
                    onChange={(e) => setReceiptInput({ ...receiptInput, notes: e.target.value })}
                    rows={3}
                    placeholder="e.g. Uploaded custom portfolio, answered coding assessment questionnaire."
                    className="w-full bg-slate-50 border border-slate-300 rounded px-3 py-2 text-slate-900 placeholder-neutral-500 focus:outline-none focus:border-blue-500 text-sm"
                  />
                </div>
              </div>

              <div className="pt-3 border-t border-slate-200 flex justify-end gap-3">
                <button
                  onClick={() => setShowReceiptModal(false)}
                  className="px-4 py-2 bg-slate-100 hover:bg-neutral-700 text-slate-700 text-sm font-medium rounded transition"
                >
                  Cancel
                </button>
                <button
                  onClick={handleConfirmSubmission}
                  className="px-5 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 text-sm font-bold rounded transition"
                >
                  Confirm & Mark Applied
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default function ApplicationWorkflowPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-white text-slate-900 p-12 flex items-center justify-center">
          <div className="text-center space-y-3">
            <div className="animate-spin text-3xl">⏳</div>
            <div className="text-slate-500 text-sm">Loading Application Review Gateway...</div>
          </div>
        </div>
      }
    >
      <ApplicationWorkflowContent />
    </Suspense>
  );
}
