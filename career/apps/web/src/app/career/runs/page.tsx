'use client';

import React, { useState } from 'react';
import {
  Play,
  Pause,
  RotateCcw,
  Square,
  AlertTriangle,
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  Clock,
  Activity,
  Layers,
  Terminal,
  Zap,
  RefreshCw,
  Gauge,
  HelpCircle,
} from 'lucide-react';
import type {
  CareerRun,
  CareerRunStatus,
  CareerRunType,
  StageSafetyLevel,
} from '@social-platform/contracts';

const INITIAL_RUNS: CareerRun[] = [
  {
    id: 'run_disc_77810',
    user_id: 'usr_demo_1',
    workspace_id: 'ws_prod',
    run_type: 'discovery_run',
    status: 'running',
    current_stage: 'evaluate_and_score_matches',
    stage_safety: 'safe_to_retry',
    fencing_token: 3,
    items_total: 25,
    items_processed: 12,
    items_succeeded: 12,
    items_failed: 0,
    hourly_limit: 50,
    recent_submissions: [
      new Date(Date.now() - 40 * 60 * 1000).toISOString(),
      new Date(Date.now() - 25 * 60 * 1000).toISOString(),
      new Date(Date.now() - 10 * 60 * 1000).toISOString(),
    ],
    checkpoints: [
      {
        stage_name: 'fetch_job_listings',
        safety: 'safe_to_retry',
        started_at: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        completed_at: new Date(Date.now() - 14 * 60 * 1000).toISOString(),
        status: 'success',
        items_processed: 25,
        diagnostics: 'Fetched 25 matches from configured boards',
      },
      {
        stage_name: 'evaluate_and_score_matches',
        safety: 'safe_to_retry',
        started_at: new Date(Date.now() - 14 * 60 * 1000).toISOString(),
        status: 'running',
        items_processed: 12,
        diagnostics: '12 items evaluated with heuristic match scoring',
      },
    ],
    audit_log: [
      {
        timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        level: 'info',
        stage: 'initialization',
        message: 'Run queued with initial fencing token 1',
        fencing_token: 1,
      },
      {
        timestamp: new Date(Date.now() - 14 * 60 * 1000).toISOString(),
        level: 'info',
        stage: 'fetch_job_listings',
        message: 'Stage fetch_job_listings completed cleanly',
        fencing_token: 1,
      },
      {
        timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        level: 'warn',
        stage: 'evaluate_and_score_matches',
        message: 'Worker paused and resumed; fencing token incremented to 3 (AT-021)',
        fencing_token: 3,
      },
    ],
    heartbeat_at: new Date(Date.now() - 30 * 1000).toISOString(),
    lease_expires_at: new Date(Date.now() + 90 * 1000).toISOString(),
    requires_reconciliation: false,
    created_at: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
    updated_at: new Date(Date.now() - 30 * 1000).toISOString(),
  },
  {
    id: 'run_apply_99201',
    user_id: 'usr_demo_1',
    workspace_id: 'ws_prod',
    run_type: 'application_session',
    status: 'needs_reconciliation',
    current_stage: 'submit_application_to_ats',
    stage_safety: 'uncertain_write_requires_reconciliation',
    fencing_token: 4,
    items_total: 1,
    items_processed: 1,
    items_succeeded: 0,
    items_failed: 1,
    hourly_limit: 10,
    recent_submissions: [
      new Date(Date.now() - 55 * 60 * 1000).toISOString(),
      new Date(Date.now() - 42 * 60 * 1000).toISOString(),
      new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      new Date(Date.now() - 18 * 60 * 1000).toISOString(),
      new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    ],
    checkpoints: [
      {
        stage_name: 'prepare_application_bundle',
        safety: 'safe_to_retry',
        started_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
        completed_at: new Date(Date.now() - 19 * 60 * 1000).toISOString(),
        status: 'success',
        items_processed: 1,
        diagnostics: 'Resume, tailored cover letter, and Q&A bundled',
      },
      {
        stage_name: 'submit_application_to_ats',
        safety: 'uncertain_write_requires_reconciliation',
        started_at: new Date(Date.now() - 19 * 60 * 1000).toISOString(),
        status: 'interrupted',
        items_processed: 1,
        diagnostics: 'Network timeout during POST /candidates/apply; lease expired',
      },
    ],
    audit_log: [
      {
        timestamp: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
        level: 'info',
        stage: 'initialization',
        message: 'Application session initiated',
        fencing_token: 1,
      },
      {
        timestamp: new Date(Date.now() - 19 * 60 * 1000).toISOString(),
        level: 'info',
        stage: 'prepare_application_bundle',
        message: 'Materials verified with 0 PII leak',
        fencing_token: 1,
      },
      {
        timestamp: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
        level: 'error',
        stage: 'submit_application_to_ats',
        message: 'CRASH DETECTED during uncertain write: automatic blind retry blocked (AT-006). Reconciliation required.',
        fencing_token: 4,
      },
    ],
    heartbeat_at: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
    lease_expires_at: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
    requires_reconciliation: true,
    reconciliation_notes: 'Worker crashed during uncertain ATS submission at CloudScale. Manual verification required before retry.',
    created_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
    updated_at: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
  },
];

export default function CareerRunsPage() {
  const [runs, setRuns] = useState<CareerRun[]>(INITIAL_RUNS);
  const [selectedRunId, setSelectedRunId] = useState<string>(INITIAL_RUNS[0].id);
  const [filterType, setFilterType] = useState<string>('all');
  const [notice, setNotice] = useState<{ type: 'success' | 'warning' | 'error'; message: string } | null>(null);

  // New Run Creation Modal / Form State
  const [newRunType, setNewRunType] = useState<CareerRunType>('discovery_run');
  const [newRunItems, setNewRunItems] = useState<number>(20);
  const [newRunLimit, setNewRunLimit] = useState<number>(15);

  // Reconciliation Dialog State
  const [reconcilingRun, setReconcilingRun] = useState<CareerRun | null>(null);
  const [reconcileResolution, setReconcileResolution] = useState<'confirm_completed' | 'abandon_attempt' | 'force_retry'>('confirm_completed');
  const [reconcileNotes, setReconcileNotes] = useState<string>('Verified application state on external ATS portal');

  const selectedRun = runs.find((r) => r.id === selectedRunId) || runs[0];

  const triggerNotice = (type: 'success' | 'warning' | 'error', message: string) => {
    setNotice({ type, message });
    setTimeout(() => setNotice(null), 5000);
  };

  const handleCreateRun = (e: React.FormEvent) => {
    e.preventDefault();
    const id = `run_${newRunType.slice(0, 4)}_${Math.floor(10000 + Math.random() * 90000)}`;
    const now = new Date().toISOString();

    const stageNames: Record<CareerRunType, { name: string; safety: StageSafetyLevel }[]> = {
      discovery_run: [
        { name: 'fetch_job_listings', safety: 'safe_to_retry' },
        { name: 'evaluate_and_score_matches', safety: 'safe_to_retry' },
        { name: 'persist_saved_jobs', safety: 'safe_to_retry' },
      ],
      tailoring_batch: [
        { name: 'retrieve_master_facts', safety: 'safe_to_retry' },
        { name: 'synthesize_tailored_artifacts', safety: 'safe_to_retry' },
      ],
      application_session: [
        { name: 'prepare_application_bundle', safety: 'safe_to_retry' },
        { name: 'submit_application_to_ats', safety: 'uncertain_write_requires_reconciliation' },
      ],
      feed_scan: [
        { name: 'scan_linkedin_posts', safety: 'safe_to_retry' },
        { name: 'extract_hiring_leads', safety: 'safe_to_retry' },
      ],
    };

    const runStages = stageNames[newRunType] || stageNames.discovery_run;

    const newRun: CareerRun = {
      id,
      user_id: 'usr_demo_1',
      workspace_id: 'ws_prod',
      run_type: newRunType,
      status: 'running',
      current_stage: runStages[0].name,
      stage_safety: runStages[0].safety,
      fencing_token: 1,
      items_total: newRunItems,
      items_processed: 0,
      items_succeeded: 0,
      items_failed: 0,
      hourly_limit: newRunLimit,
      recent_submissions: [],
      checkpoints: [
        {
          stage_name: runStages[0].name,
          safety: runStages[0].safety,
          started_at: now,
          status: 'running',
          items_processed: 0,
          diagnostics: `Initiating ${runStages[0].name} for ${newRunItems} targets`,
        },
      ],
      audit_log: [
        {
          timestamp: now,
          level: 'info',
          stage: 'initialization',
          message: `Run created with ${newRunItems} items, limit ${newRunLimit}/hr, fencing token 1`,
          fencing_token: 1,
        },
      ],
      heartbeat_at: now,
      lease_expires_at: new Date(Date.now() + 120 * 1000).toISOString(),
      requires_reconciliation: false,
      created_at: now,
      updated_at: now,
    };

    setRuns([newRun, ...runs]);
    setSelectedRunId(newRun.id);
    triggerNotice('success', `Created run ${newRun.id} with fencing token 1`);
  };

  const handleControlAction = (runId: string, action: 'pause' | 'resume' | 'drain' | 'cancel') => {
    setRuns((prev) =>
      prev.map((run) => {
        if (run.id !== runId) return run;

        const now = new Date().toISOString();
        let newStatus: CareerRunStatus = run.status;
        let newToken = run.fencing_token;
        let newLevel = 'info';
        let msg = '';

        if (action === 'pause') {
          newToken++;
          newStatus = 'paused';
          newLevel = 'warn';
          msg = `User paused run; fencing token bumped to ${newToken} (AT-021: invalidates stale workers)`;
        } else if (action === 'resume') {
          // Check rate limit per FND-011
          const oneHourAgo = Date.now() - 60 * 60 * 1000;
          const used = run.recent_submissions.filter((t) => new Date(t).getTime() > oneHourAgo).length;
          if (used >= run.hourly_limit) {
            triggerNotice('error', `Resume throttled: ${used}/${run.hourly_limit} hourly submissions exhausted (FND-011)`);
            return run;
          }
          newToken++;
          newStatus = 'running';
          msg = `Run resumed from safe checkpoint; lease renewed under token ${newToken}`;
        } else if (action === 'drain') {
          newStatus = 'draining';
          msg = `Graceful draining initiated (SRC-C3, REQ-017): completing current batch item and stopping`;
        } else if (action === 'cancel') {
          newToken++;
          newStatus = 'cancelled';
          newLevel = 'error';
          msg = `Run terminated by user; fencing token ${newToken} rejects in-flight updates`;
        }

        return {
          ...run,
          status: newStatus,
          fencing_token: newToken,
          updated_at: now,
          audit_log: [
            ...run.audit_log,
            {
              timestamp: now,
              level: newLevel,
              stage: run.current_stage,
              message: msg,
              fencing_token: newToken,
            },
          ],
        };
      })
    );

    triggerNotice('warning', `Action '${action}' executed on run ${runId}`);
  };

  const handleSimulateCrash = (runId: string) => {
    setRuns((prev) =>
      prev.map((run) => {
        if (run.id !== runId) return run;

        const now = new Date().toISOString();
        const newToken = run.fencing_token + 1;

        if (run.stage_safety === 'uncertain_write_requires_reconciliation') {
          return {
            ...run,
            status: 'needs_reconciliation',
            fencing_token: newToken,
            requires_reconciliation: true,
            reconciliation_notes: `Crash detected during uncertain write in stage '${run.current_stage}'. Zero-duplicate invariant (AT-006) prevents blind retry.`,
            updated_at: now,
            audit_log: [
              ...run.audit_log,
              {
                timestamp: now,
                level: 'error',
                stage: run.current_stage,
                message: `WORKER CRASH: Uncertain write requires manual reconciliation before retry (AT-006)`,
                fencing_token: newToken,
              },
            ],
          };
        } else {
          return {
            ...run,
            status: 'paused',
            fencing_token: newToken,
            updated_at: now,
            audit_log: [
              ...run.audit_log,
              {
                timestamp: now,
                level: 'warn',
                stage: run.current_stage,
                message: `WORKER CRASH in idempotent stage '${run.current_stage}'. Run paused at checkpoint for safe resumption.`,
                fencing_token: newToken,
              },
            ],
          };
        }
      })
    );

    triggerNotice('error', `Simulated worker lease timeout crash on run ${runId}`);
  };

  const handleCommitReconciliation = () => {
    if (!reconcilingRun) return;

    const now = new Date().toISOString();
    setRuns((prev) =>
      prev.map((run) => {
        if (run.id !== reconcilingRun.id) return run;

        let finalStatus: CareerRunStatus = 'completed';
        let succeededDelta = 0;
        let failedDelta = 0;
        let logMsg = '';

        if (reconcileResolution === 'confirm_completed') {
          finalStatus = 'completed';
          succeededDelta = 1;
          logMsg = `Reconciled: User confirmed application completed on ATS portal. Notes: ${reconcileNotes}`;
        } else if (reconcileResolution === 'abandon_attempt') {
          finalStatus = 'failed';
          failedDelta = 1;
          logMsg = `Reconciled: User marked submission attempt abandoned. Notes: ${reconcileNotes}`;
        } else if (reconcileResolution === 'force_retry') {
          finalStatus = 'running';
          logMsg = `Reconciled: Explicit human override to force retry. Notes: ${reconcileNotes}`;
        }

        return {
          ...run,
          status: finalStatus,
          fencing_token: run.fencing_token + 1,
          requires_reconciliation: false,
          reconciliation_notes: reconcileNotes,
          items_succeeded: run.items_succeeded + succeededDelta,
          items_failed: run.items_failed + failedDelta,
          updated_at: now,
          audit_log: [
            ...run.audit_log,
            {
              timestamp: now,
              level: 'info',
              stage: run.current_stage,
              message: logMsg,
              fencing_token: run.fencing_token + 1,
            },
          ],
        };
      })
    );

    triggerNotice('success', `Reconciled run ${reconcilingRun.id} as '${reconcileResolution}'`);
    setReconcilingRun(null);
  };

  const filteredRuns = runs.filter((r) => {
    if (filterType === 'all') return true;
    if (filterType === 'active') return r.status === 'running' || r.status === 'queued' || r.status === 'draining';
    if (filterType === 'attention') return r.status === 'needs_reconciliation' || r.status === 'paused';
    if (filterType === 'finished') return r.status === 'completed' || r.status === 'cancelled' || r.status === 'failed';
    return true;
  });

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 lg:p-10 font-sans">
      {/* Top Banner & Header */}
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-slate-200 pb-6">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2.5 py-0.5 rounded text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                IMP-CAR-21
              </span>
              <span className="px-2.5 py-0.5 rounded text-xs font-semibold bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                AT-006 & AT-021 Invariants
              </span>
              <span className="px-2.5 py-0.5 rounded text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                FND-011 Rate Limiting
              </span>
            </div>
            <h1 className="text-2xl lg:text-3xl font-bold tracking-tight text-slate-900 flex items-center gap-3">
              <Activity className="w-8 h-8 text-cyan-400" />
              Run Controls & Crash Recovery Dashboard
            </h1>
            <p className="text-slate-500 text-sm mt-1">
              Pause/cancel/resume safe stages, monotonic fencing token worker protection, zero-duplicate uncertain write recovery, and rolling hourly rate limits.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <div className="bg-white border border-slate-200 rounded-lg p-2 px-3 text-right">
              <div className="text-xs text-slate-500">Total Managed Runs</div>
              <div className="text-lg font-bold text-slate-900">{runs.length}</div>
            </div>
            <div className="bg-white border border-slate-200 rounded-lg p-2 px-3 text-right">
              <div className="text-xs text-slate-500">Needs Reconciliation</div>
              <div className="text-lg font-bold text-amber-400">
                {runs.filter((r) => r.status === 'needs_reconciliation').length}
              </div>
            </div>
          </div>
        </div>

        {/* Global Notice Alert */}
        {notice && (
          <div
            className={`p-4 rounded-lg flex items-center gap-3 text-sm border animate-fade-in ${
              notice.type === 'success'
                ? 'bg-emerald-950/40 border-emerald-600 text-emerald-200'
                : notice.type === 'warning'
                ? 'bg-amber-950/40 border-amber-600 text-amber-200'
                : 'bg-rose-950/40 border-rose-600 text-rose-200'
            }`}
          >
            {notice.type === 'success' && <CheckCircle2 className="w-5 h-5 flex-shrink-0" />}
            {notice.type === 'warning' && <AlertTriangle className="w-5 h-5 flex-shrink-0" />}
            {notice.type === 'error' && <ShieldAlert className="w-5 h-5 flex-shrink-0" />}
            <span>{notice.message}</span>
          </div>
        )}

        {/* Invariant Highlights Card */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white/80 border border-slate-200 rounded-xl p-4 flex gap-3">
            <ShieldCheck className="w-6 h-6 text-emerald-400 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="text-sm font-semibold text-slate-900">AT-006: Zero-Duplicate Rule</h4>
              <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                Uncertain writes (final ATS submit, direct outreach) are <strong className="text-amber-300">never blindly retried</strong> upon worker crash or timeout. They require explicit human reconciliation.
              </p>
            </div>
          </div>

          <div className="bg-white/80 border border-slate-200 rounded-xl p-4 flex gap-3">
            <Zap className="w-6 h-6 text-cyan-400 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="text-sm font-semibold text-slate-900">AT-021: Fencing Tokens</h4>
              <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                Each pause, resume, or cancellation increments the run&apos;s monotonically increasing fencing token. Late worker updates with stale tokens are automatically rejected.
              </p>
            </div>
          </div>

          <div className="bg-white/80 border border-slate-200 rounded-xl p-4 flex gap-3">
            <Gauge className="w-6 h-6 text-indigo-400 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="text-sm font-semibold text-slate-900">FND-011: Rolling Hourly Limits</h4>
              <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                Sliding 60-minute rate limiting window prevents platform abuse and protects user accounts from detection or throttling.
              </p>
            </div>
          </div>
        </div>

        {/* Main Content Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column: Runs List & Creator */}
          <div className="space-y-6 lg:col-span-1">
            {/* Create New Run Card */}
            <div className="bg-white border border-slate-200 rounded-xl p-4">
              <h3 className="text-sm font-semibold text-slate-900 mb-3 flex items-center gap-2">
                <Play className="w-4 h-4 text-emerald-400" />
                Initialize New Run
              </h3>
              <form onSubmit={handleCreateRun} className="space-y-3 text-xs">
                <div>
                  <label className="block text-slate-500 mb-1">Run Execution Type</label>
                  <select
                    value={newRunType}
                    onChange={(e) => setNewRunType(e.target.value as CareerRunType)}
                    className="w-full bg-slate-50 border border-slate-300 rounded-lg p-2 text-slate-800 focus:outline-none focus:border-cyan-500"
                  >
                    <option value="discovery_run">Job Discovery & Match Scoring (Safe)</option>
                    <option value="tailoring_batch">Artifact Tailoring Batch (Safe)</option>
                    <option value="application_session">Application Session (Uncertain Write Stage)</option>
                    <option value="feed_scan">Hiring Lead Feed Scan (Safe)</option>
                  </select>
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-slate-500 mb-1">Total Items</label>
                    <input
                      type="number"
                      min={1}
                      max={100}
                      value={newRunItems}
                      onChange={(e) => setNewRunItems(Number(e.target.value))}
                      className="w-full bg-slate-50 border border-slate-300 rounded-lg p-2 text-slate-800 focus:outline-none focus:border-cyan-500"
                    />
                  </div>
                  <div>
                    <label className="block text-slate-500 mb-1">Hourly Limit</label>
                    <input
                      type="number"
                      min={1}
                      max={100}
                      value={newRunLimit}
                      onChange={(e) => setNewRunLimit(Number(e.target.value))}
                      className="w-full bg-slate-50 border border-slate-300 rounded-lg p-2 text-slate-800 focus:outline-none focus:border-cyan-500"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  className="w-full py-2 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-slate-900 font-medium rounded-lg text-xs transition flex items-center justify-center gap-1.5 shadow"
                >
                  <Play className="w-3.5 h-3.5 fill-current" />
                  Launch Safe Run
                </button>
              </form>
            </div>

            {/* Runs Filter & List */}
            <div className="bg-white border border-slate-200 rounded-xl p-4 space-y-3">
              <div className="flex items-center justify-between pb-2 border-b border-slate-200">
                <h3 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
                  <Layers className="w-4 h-4 text-cyan-400" />
                  Managed Runs ({filteredRuns.length})
                </h3>
                <div className="flex gap-1 text-[11px]">
                  {['all', 'active', 'attention', 'finished'].map((tab) => (
                    <button
                      key={tab}
                      onClick={() => setFilterType(tab)}
                      className={`px-2 py-0.5 rounded capitalize ${
                        filterType === tab ? 'bg-cyan-500/20 text-cyan-400 font-semibold' : 'text-slate-500 hover:text-slate-800'
                      }`}
                    >
                      {tab}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-2 max-h-[500px] overflow-y-auto pr-1">
                {filteredRuns.map((run) => {
                  const isSelected = run.id === selectedRunId;
                  const statusColors: Record<CareerRunStatus, string> = {
                    queued: 'bg-slate-500/10 text-slate-500 border-slate-500/20',
                    running: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20 animate-pulse',
                    paused: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
                    draining: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
                    completed: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
                    failed: 'bg-rose-500/10 text-rose-400 border-rose-500/20',
                    cancelled: 'bg-slate-700/10 text-slate-500 border-slate-300/20',
                    needs_reconciliation: 'bg-rose-500/20 text-rose-300 border-rose-500/40 font-bold animate-bounce',
                    crashed: 'bg-rose-600/20 text-rose-400 border-rose-600/30',
                  };

                  return (
                    <div
                      key={run.id}
                      onClick={() => setSelectedRunId(run.id)}
                      className={`p-3 rounded-lg border text-left cursor-pointer transition ${
                        isSelected
                          ? 'bg-slate-100/80 border-cyan-500/50 shadow-md'
                          : 'bg-slate-50/40 border-slate-200 hover:bg-slate-100/40'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="font-mono text-xs text-cyan-300 font-medium">{run.id}</span>
                        <span className={`text-[10px] px-2 py-0.5 rounded border uppercase tracking-wide ${statusColors[run.status]}`}>
                          {run.status.replace('_', ' ')}
                        </span>
                      </div>

                      <div className="flex items-center justify-between text-xs text-slate-500">
                        <span className="capitalize">{run.run_type.replace('_', ' ')}</span>
                        <span>Token #{run.fencing_token}</span>
                      </div>

                      <div className="mt-2 w-full bg-slate-100 h-1.5 rounded-full overflow-hidden">
                        <div
                          className={`h-full ${run.status === 'needs_reconciliation' ? 'bg-amber-500' : 'bg-cyan-500'}`}
                          style={{
                            width: `${Math.min(100, Math.round((run.items_processed / Math.max(1, run.items_total)) * 100))}%`,
                          }}
                        />
                      </div>

                      <div className="flex items-center justify-between text-[10px] text-slate-500 mt-1">
                        <span>
                          {run.items_processed}/{run.items_total} processed
                        </span>
                        <span>{new Date(run.updated_at).toLocaleTimeString()}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Right Column: Selected Run Details, Controls & Logs */}
          <div className="lg:col-span-2 space-y-6">
            {selectedRun && (
              <>
                {/* Run Inspector Card */}
                <div className="bg-white border border-slate-200 rounded-xl p-6 space-y-5">
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-slate-200 pb-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <h2 className="text-xl font-bold text-slate-900 font-mono">{selectedRun.id}</h2>
                        <span className="px-2 py-0.5 rounded text-xs bg-slate-100 text-slate-700 border border-slate-300">
                          Token #{selectedRun.fencing_token}
                        </span>
                        <span
                          className={`px-2 py-0.5 rounded text-xs font-semibold ${
                            selectedRun.stage_safety === 'safe_to_retry'
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                          }`}
                        >
                          {selectedRun.stage_safety === 'safe_to_retry' ? 'Safe to Retry' : 'Uncertain Write Stage'}
                        </span>
                      </div>
                      <p className="text-xs text-slate-500 mt-1">
                        Current Active Stage: <strong className="text-cyan-400">{selectedRun.current_stage}</strong>
                      </p>
                    </div>

                    {/* Operational Action Buttons */}
                    <div className="flex items-center gap-2 flex-wrap">
                      {selectedRun.status === 'running' && (
                        <>
                          <button
                            onClick={() => handleControlAction(selectedRun.id, 'pause')}
                            className="px-3 py-1.5 bg-amber-600/20 hover:bg-amber-600/30 text-amber-300 border border-amber-500/30 rounded text-xs font-medium flex items-center gap-1.5 transition"
                            title="Pause run and increment fencing token (AT-021)"
                          >
                            <Pause className="w-3.5 h-3.5" />
                            Pause
                          </button>
                          <button
                            onClick={() => handleControlAction(selectedRun.id, 'drain')}
                            className="px-3 py-1.5 bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 border border-indigo-500/30 rounded text-xs font-medium flex items-center gap-1.5 transition"
                            title="Gracefully complete current batch item and stop (SRC-C3)"
                          >
                            <Clock className="w-3.5 h-3.5" />
                            Drain
                          </button>
                        </>
                      )}

                      {selectedRun.status === 'paused' && (
                        <button
                          onClick={() => handleControlAction(selectedRun.id, 'resume')}
                          className="px-3 py-1.5 bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/30 rounded text-xs font-medium flex items-center gap-1.5 transition"
                          title="Resume run after sliding rate limit check (FND-011)"
                        >
                          <Play className="w-3.5 h-3.5" />
                          Resume
                        </button>
                      )}

                      {(selectedRun.status === 'running' || selectedRun.status === 'paused' || selectedRun.status === 'draining') && (
                        <button
                          onClick={() => handleControlAction(selectedRun.id, 'cancel')}
                          className="px-3 py-1.5 bg-rose-600/20 hover:bg-rose-600/30 text-rose-300 border border-rose-500/30 rounded text-xs font-medium flex items-center gap-1.5 transition"
                          title="Permanently cancel run and fence late workers (AT-021)"
                        >
                          <Square className="w-3.5 h-3.5" />
                          Cancel
                        </button>
                      )}

                      {/* Crash Simulator Trigger */}
                      {(selectedRun.status === 'running' || selectedRun.status === 'draining') && (
                        <button
                          onClick={() => handleSimulateCrash(selectedRun.id)}
                          className="px-3 py-1.5 bg-purple-600/20 hover:bg-purple-600/30 text-purple-300 border border-purple-500/30 rounded text-xs font-medium flex items-center gap-1.5 transition"
                          title="Simulate worker crash and lease expiration to test AT-006 & AT-021"
                        >
                          <Zap className="w-3.5 h-3.5" />
                          Simulate Crash
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Needs Reconciliation Banner */}
                  {selectedRun.status === 'needs_reconciliation' && (
                    <div className="bg-amber-950/40 border border-amber-500/50 rounded-lg p-4 space-y-3">
                      <div className="flex items-start gap-3">
                        <AlertTriangle className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
                        <div>
                          <h4 className="text-sm font-semibold text-amber-300">
                            Reconciliation Required (AT-006 Zero-Duplicate Invariant)
                          </h4>
                          <p className="text-xs text-amber-200/80 mt-1 leading-relaxed">
                            {selectedRun.reconciliation_notes ||
                              'A worker crash or timeout occurred during an uncertain write. Automatic retry is strictly blocked to prevent submitting duplicate applications to ATS or sending duplicate outreach.'}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2 pt-2 border-t border-amber-500/20">
                        <button
                          onClick={() => setReconcilingRun(selectedRun)}
                          className="px-3 py-1.5 bg-amber-500 text-slate-950 font-semibold rounded text-xs hover:bg-amber-400 transition"
                        >
                          Resolve & Reconcile Now
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Rate Limiter & Progress Metrics */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                    <div className="bg-slate-50/60 p-3 rounded-lg border border-slate-200">
                      <span className="text-slate-500 block mb-1">Items Processed</span>
                      <span className="text-base font-bold text-slate-900">
                        {selectedRun.items_processed} / {selectedRun.items_total}
                      </span>
                    </div>
                    <div className="bg-slate-50/60 p-3 rounded-lg border border-slate-200">
                      <span className="text-slate-500 block mb-1">Success / Failures</span>
                      <span className="text-base font-bold text-emerald-400">
                        {selectedRun.items_succeeded}{' '}
                        <span className="text-rose-400 font-normal">/ {selectedRun.items_failed}</span>
                      </span>
                    </div>
                    <div className="bg-slate-50/60 p-3 rounded-lg border border-slate-200">
                      <span className="text-slate-500 block mb-1">Hourly Limit (FND-011)</span>
                      <span className="text-base font-bold text-cyan-400">
                        {selectedRun.recent_submissions.length} / {selectedRun.hourly_limit}/hr
                      </span>
                    </div>
                    <div className="bg-slate-50/60 p-3 rounded-lg border border-slate-200">
                      <span className="text-slate-500 block mb-1">Lease Status</span>
                      <span className="text-xs font-mono text-slate-700">
                        {new Date(selectedRun.lease_expires_at) > new Date() ? 'Active Lease' : 'Expired'}
                      </span>
                    </div>
                  </div>

                  {/* Checkpoints Breakdown */}
                  <div className="space-y-3">
                    <h3 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
                      <Layers className="w-4 h-4 text-cyan-400" />
                      Stage Checkpoints
                    </h3>
                    <div className="space-y-2">
                      {selectedRun.checkpoints.map((cp, idx) => (
                        <div
                          key={idx}
                          className="bg-slate-50/50 border border-slate-200 rounded-lg p-3 text-xs flex items-center justify-between"
                        >
                          <div className="space-y-1">
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-slate-800">{cp.stage_name}</span>
                              <span
                                className={`text-[10px] px-2 py-0.2 rounded border ${
                                  cp.safety === 'safe_to_retry'
                                    ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                                    : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                                }`}
                              >
                                {cp.safety}
                              </span>
                            </div>
                            <div className="text-slate-500 text-[11px]">{cp.diagnostics}</div>
                          </div>
                          <div className="text-right">
                            <span
                              className={`text-[10px] px-2 py-0.5 rounded capitalize font-medium ${
                                cp.status === 'success'
                                  ? 'text-emerald-400 bg-emerald-500/10'
                                  : cp.status === 'running'
                                  ? 'text-cyan-400 bg-cyan-500/10'
                                  : 'text-rose-400 bg-rose-500/10'
                              }`}
                            >
                              {cp.status}
                            </span>
                            <div className="text-[10px] text-slate-500 mt-1">
                              {new Date(cp.started_at).toLocaleTimeString()}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Immutable Diagnostics Audit Log */}
                  <div className="space-y-3">
                    <h3 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
                      <Terminal className="w-4 h-4 text-cyan-400" />
                      Immutable Diagnostic Log ({selectedRun.audit_log.length} records)
                    </h3>
                    <div className="bg-slate-50 rounded-lg p-3 border border-slate-200 max-h-60 overflow-y-auto space-y-2 font-mono text-xs">
                      {selectedRun.audit_log.map((log, idx) => {
                        const levelColors: Record<string, string> = {
                          info: 'text-cyan-400',
                          warn: 'text-amber-400',
                          error: 'text-rose-400',
                          security: 'text-purple-400',
                        };

                        return (
                          <div key={idx} className="flex items-start gap-2 border-b border-slate-900 pb-1.5">
                            <span className="text-slate-500 text-[10px]">
                              {new Date(log.timestamp).toLocaleTimeString()}
                            </span>
                            <span className={`text-[10px] uppercase font-bold ${levelColors[log.level] || 'text-slate-500'}`}>
                              [{log.level}]
                            </span>
                            <span className="text-slate-500 text-[10px]">T#{log.fencing_token}</span>
                            <span className="text-slate-700 flex-1">{log.message}</span>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Reconciliation Dialog Modal */}
      {reconcilingRun && (
        <div className="fixed inset-0 z-50 bg-slate-50/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white border border-slate-300 rounded-xl p-6 max-w-lg w-full space-y-4 shadow-2xl text-xs">
            <div className="flex items-center gap-2 text-amber-400 text-sm font-bold border-b border-slate-200 pb-3">
              <AlertTriangle className="w-5 h-5" />
              Reconcile Crashed Run: {reconcilingRun.id}
            </div>

            <p className="text-slate-700 leading-relaxed">
              Per <strong className="text-slate-900">AT-006</strong>, this uncertain write stage crashed. Please inspect the external ATS portal or communication channel, and select how to reconcile this attempt:
            </p>

            <div className="space-y-2">
              <label className="block text-slate-500">Reconciliation Resolution</label>
              <div className="space-y-2">
                <label className="flex items-start gap-2 p-2.5 rounded-lg border border-slate-200 bg-slate-50/50 cursor-pointer hover:border-slate-300">
                  <input
                    type="radio"
                    name="resolution"
                    checked={reconcileResolution === 'confirm_completed'}
                    onChange={() => setReconcileResolution('confirm_completed')}
                    className="mt-0.5 text-cyan-500"
                  />
                  <div>
                    <span className="font-semibold text-slate-900 block">Confirm Completed</span>
                    <span className="text-slate-500 text-[11px]">
                      Application was successfully received on ATS. Mark item succeeded and run completed without resubmitting.
                    </span>
                  </div>
                </label>

                <label className="flex items-start gap-2 p-2.5 rounded-lg border border-slate-200 bg-slate-50/50 cursor-pointer hover:border-slate-300">
                  <input
                    type="radio"
                    name="resolution"
                    checked={reconcileResolution === 'abandon_attempt'}
                    onChange={() => setReconcileResolution('abandon_attempt')}
                    className="mt-0.5 text-cyan-500"
                  />
                  <div>
                    <span className="font-semibold text-slate-900 block">Abandon Attempt</span>
                    <span className="text-slate-500 text-[11px]">
                      Submission failed or was rejected. Mark attempt failed and do not retry.
                    </span>
                  </div>
                </label>

                <label className="flex items-start gap-2 p-2.5 rounded-lg border border-slate-200 bg-slate-50/50 cursor-pointer hover:border-slate-300">
                  <input
                    type="radio"
                    name="resolution"
                    checked={reconcileResolution === 'force_retry'}
                    onChange={() => setReconcileResolution('force_retry')}
                    className="mt-0.5 text-cyan-500"
                  />
                  <div>
                    <span className="font-semibold text-slate-900 block">Explicit Force Retry</span>
                    <span className="text-slate-500 text-[11px]">
                      Verified application was not submitted. Explicitly authorize a fresh submission under a new fencing token.
                    </span>
                  </div>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-slate-500 mb-1">Reconciliation Notes & Verification Evidence</label>
              <textarea
                rows={3}
                value={reconcileNotes}
                onChange={(e) => setReconcileNotes(e.target.value)}
                className="w-full bg-slate-50 border border-slate-300 rounded-lg p-2 text-slate-800 focus:outline-none focus:border-cyan-500 text-xs"
                placeholder="Describe how external portal was verified..."
              />
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200">
              <button
                onClick={() => setReconcilingRun(null)}
                className="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-700 transition"
              >
                Cancel
              </button>
              <button
                onClick={handleCommitReconciliation}
                className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-slate-900 font-medium rounded-lg transition"
              >
                Commit Reconciliation
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
