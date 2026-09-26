'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import {
  DailyReport,
  FollowUpRecommendation,
  UpcomingInterview,
  IncompleteProfileAlert,
  JobDiscoveryHighlight,
  CareerReminder,
  DailyReportConfig,
  ReminderStatus,
  ReminderPriority,
  ReminderType,
} from '@social-platform/contracts';
import {
  Calendar,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Bell,
  ArrowRight,
  TrendingUp,
  Briefcase,
  User,
  Settings,
  Plus,
  ExternalLink,
  Shield,
  FileText,
  Check,
  RotateCcw,
  Sparkles,
  MapPin,
  ChevronRight,
  Globe,
  Sliders,
  Send,
  X,
} from 'lucide-react';

export default function DailyReportsPage() {
  // Current active date and simulated report state
  const [currentDate, setCurrentDate] = useState('2026-09-20');
  const [activeTab, setActiveTab] = useState<'digest' | 'reminders' | 'config'>('digest');

  // Initial mock report reflecting groundtruth fixture
  const [report, setReport] = useState<DailyReport>({
    id: 'rpt_usr_01_2026-09-20',
    user_id: 'usr_candidate_01',
    workspace_id: 'ws_personal_01',
    report_date: '2026-09-20',
    timezone: 'America/New_York',
    activity: {
      applied_today: 2,
      interviews_scheduled: 1,
      follow_ups_due: 2,
      active_applications_total: 8,
      momentum_score: 88,
    },
    follow_ups: [
      {
        application_id: 'app_github_01',
        company: 'GitHub',
        role: 'Staff Backend Engineer',
        applied_at: '2026-09-12T10:00:00Z',
        days_since_applied: 8,
        suggested_action: 'Application submitted 8 days ago with no response. Send a polite check-in note to the recruiter.',
        recruiter_contact: 'sarah.connor@github.com',
      },
      {
        application_id: 'app_cloudflare_02',
        company: 'Cloudflare',
        role: 'Systems Engineer - Edge',
        applied_at: '2026-09-11T16:00:00Z',
        days_since_applied: 9,
        suggested_action: 'Application submitted 9 days ago. Follow up on status via recruiter LinkedIn message.',
      },
    ],
    interviews: [
      {
        id: 'int_001',
        company: 'Stripe',
        role: 'Senior Distributed Systems Engineer',
        round: 'System Design Round',
        scheduled_at: '2026-09-21T14:00:00-04:00',
        meeting_link: 'https://meet.google.com/xyz-stripe',
        prep_notes: 'Review confirmed project highlights on distributed ledger, raft consensus, and idempotency patterns.',
      },
    ],
    incomplete_profile_items: [
      {
        section: 'experience',
        issue: 'Work experience lacks quantified metrics (%, $, scale)',
        impact: 'ATS resume scoring weights power achievements with measurable impact higher',
        recommended_action: 'Include numbers (e.g., "reduced API latency by 35%", "scaled to 50k RPS")',
        severity: 'medium',
      },
      {
        section: 'skills',
        issue: 'Only 4 technical skills listed',
        impact: 'Low skill count reduces keyword match scores against job postings',
        recommended_action: 'Add at least 5-10 core skills, tools, and libraries you actively work with',
        severity: 'medium',
      },
    ],
    job_highlights: [
      {
        job_id: 'job_01',
        title: 'Principal Infrastructure Engineer',
        company: 'HashiCorp',
        location: 'Remote, US',
        match_score: 94,
        direct_apply_url: 'https://boards.greenhouse.io/hashicorp/jobs/12345',
        disclosed_salary: '$185,000 - $220,000 USD',
      },
      {
        job_id: 'job_02',
        title: 'Staff Backend Architect',
        company: 'Datadog',
        location: 'New York, NY',
        match_score: 91,
        direct_apply_url: 'https://boards.greenhouse.io/datadog/jobs/67890',
        disclosed_salary: '$190,000 - $230,000 USD',
      },
    ],
    generated_at: '2026-09-20T09:00:00Z',
  });

  // Config State
  const [config, setConfig] = useState<DailyReportConfig>({
    user_id: 'usr_candidate_01',
    workspace_id: 'ws_personal_01',
    enabled: true,
    scheduled_hour: 9,
    scheduled_minute: 0,
    timezone: 'America/New_York',
    delivery_days: 'weekdays',
    channels: {
      in_app: true,
      webhook_enabled: false,
      webhook_url: '',
      telegram_enabled: false,
      telegram_chat_id: '',
    },
    categories: {
      follow_ups: true,
      interviews: true,
      profile_gaps: true,
      job_discovery: true,
    },
    next_delivery_utc: '2026-09-21T13:00:00Z', // 09:00 EDT is 13:00 UTC
    updated_at: '2026-09-20T09:00:00Z',
  });

  // Reminders State
  const [reminders, setReminders] = useState<CareerReminder[]>([
    {
      id: 'rem_001',
      user_id: 'usr_candidate_01',
      workspace_id: 'ws_personal_01',
      type: 'interview',
      priority: 'high',
      title: 'Stripe System Design Interview Prep',
      description: 'Review distributed consensus, caching layers, and database sharding architectures.',
      due_at: '2026-09-21T12:00:00Z',
      status: 'pending',
      created_at: '2026-09-20T08:00:00Z',
      updated_at: '2026-09-20T08:00:00Z',
    },
    {
      id: 'rem_002',
      user_id: 'usr_candidate_01',
      workspace_id: 'ws_personal_01',
      type: 'follow_up',
      priority: 'medium',
      title: 'Follow up on GitHub Staff Backend application',
      description: 'Send follow-up note to recruiter Sarah Connor (applied 8 days ago).',
      due_at: '2026-09-20T17:00:00Z',
      status: 'pending',
      created_at: '2026-09-20T08:00:00Z',
      updated_at: '2026-09-20T08:00:00Z',
    },
    {
      id: 'rem_003',
      user_id: 'usr_candidate_01',
      workspace_id: 'ws_personal_01',
      type: 'profile_incomplete',
      priority: 'low',
      title: 'Add quantified metrics to Senior Engineer role',
      description: 'Update master profile with % latency improvements and QPS scales.',
      due_at: '2026-09-22T09:00:00Z',
      status: 'snoozed',
      snoozed_until: '2026-09-21T09:00:00Z',
      created_at: '2026-09-19T09:00:00Z',
      updated_at: '2026-09-20T08:30:00Z',
    },
  ]);

  // Modals & form state
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [showAddReminderModal, setShowAddReminderModal] = useState(false);
  const [newReminderTitle, setNewReminderTitle] = useState('');
  const [newReminderDesc, setNewReminderDesc] = useState('');
  const [newReminderPriority, setNewReminderPriority] = useState<ReminderPriority>('medium');
  const [newReminderType, setNewReminderType] = useState<ReminderType>('custom');

  // Action status message
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 4000);
  };

  const handleReminderAction = (id: string, action: ReminderStatus, snoozeHours?: number) => {
    setReminders(prev =>
      prev.map(rem => {
        if (rem.id === id) {
          const now = new Date();
          let snoozedUntil: string | undefined;
          if (action === 'snoozed') {
            const h = snoozeHours || 24;
            snoozedUntil = new Date(now.getTime() + h * 60 * 60 * 1000).toISOString();
          }
          return {
            ...rem,
            status: action,
            snoozed_until: snoozedUntil,
            updated_at: now.toISOString(),
          };
        }
        return rem;
      })
    );

    if (action === 'completed') {
      showToast('Reminder marked as completed!');
    } else if (action === 'snoozed') {
      showToast(`Reminder snoozed for ${snoozeHours || 24} hours.`);
    } else if (action === 'dismissed') {
      showToast('Reminder dismissed.');
    }
  };

  const handleCreateReminder = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newReminderTitle.trim()) return;

    const newRem: CareerReminder = {
      id: `rem_custom_${Date.now()}`,
      user_id: 'usr_candidate_01',
      workspace_id: 'ws_personal_01',
      type: newReminderType,
      priority: newReminderPriority,
      title: newReminderTitle.trim(),
      description: newReminderDesc.trim(),
      due_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      status: 'pending',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    setReminders([newRem, ...reminders]);
    setNewReminderTitle('');
    setNewReminderDesc('');
    setShowAddReminderModal(false);
    showToast('Custom career reminder created.');
  };

  const handleSaveConfig = (e: React.FormEvent) => {
    e.preventDefault();
    // Recompute next delivery
    setShowConfigModal(false);
    showToast(`Report delivery schedule updated for ${config.timezone} at ${String(config.scheduled_hour).padStart(2, '0')}:${String(config.scheduled_minute).padStart(2, '0')}.`);
  };

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 md:p-8 font-sans">
      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 bg-indigo-600 text-slate-900 px-5 py-3 rounded-lg shadow-xl flex items-center gap-3 border border-indigo-400/30 animate-in fade-in slide-in-from-bottom-5">
          <CheckCircle2 className="w-5 h-5 text-indigo-200" />
          <span className="text-sm font-medium">{toastMessage}</span>
        </div>
      )}

      {/* Header & Sub-navigation */}
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-200 pb-6">
          <div>
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-1">
              <Sparkles className="w-4 h-4" />
              Product 1: Career & LinkedIn Platform (yourdomain.com/career)
            </div>
            <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
              Daily Digest & Reminders
              <span className="text-xs bg-indigo-500/20 text-indigo-300 px-3 py-1 rounded-full border border-indigo-500/30 font-mono">
                IMP-CAR-19 • CAR-19 • AT-018
              </span>
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Automated daily progress briefs, follow-up alerts, interview briefs, and timezone-aware reminder schedules.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Link
              href="/career/status"
              className="text-xs text-slate-500 hover:text-slate-900 flex items-center gap-1 bg-white px-3 py-2 rounded-lg border border-slate-200 transition"
            >
              <ArrowRight className="w-3.5 h-3.5 rotate-180" /> Back to Status Hub
            </Link>

            <button
              onClick={() => setShowConfigModal(true)}
              className="flex items-center gap-2 bg-slate-100 hover:bg-slate-700 text-slate-900 text-xs font-medium px-4 py-2 rounded-lg border border-slate-300 transition"
            >
              <Settings className="w-4 h-4 text-indigo-400" />
              Schedule & Timezone
            </button>

            <button
              onClick={() => setShowAddReminderModal(true)}
              className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-slate-900 text-xs font-semibold px-4 py-2 rounded-lg shadow-lg shadow-indigo-600/20 transition"
            >
              <Plus className="w-4 h-4" />
              New Reminder
            </button>
          </div>
        </div>

        {/* View Tabs */}
        <div className="flex items-center gap-2 border-b border-slate-200 pb-2">
          <button
            onClick={() => setActiveTab('digest')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition flex items-center gap-2 ${
              activeTab === 'digest'
                ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <FileText className="w-4 h-4" />
            Today's Digest ({currentDate})
          </button>
          <button
            onClick={() => setActiveTab('reminders')}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition flex items-center gap-2 ${
              activeTab === 'reminders'
                ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <Bell className="w-4 h-4" />
            Active Reminders ({reminders.filter(r => r.status === 'pending').length})
          </button>
        </div>

        {/* TAB 1: DAILY DIGEST */}
        {activeTab === 'digest' && (
          <div className="space-y-6">
            {/* Momentum & Top Stats Banner */}
            <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
              <div className="bg-gradient-to-br from-indigo-950/60 to-slate-900 p-5 rounded-xl border border-indigo-500/30 col-span-1 md:col-span-2 flex flex-col justify-between">
                <div>
                  <div className="flex items-center justify-between text-xs text-indigo-300 font-semibold uppercase tracking-wider">
                    <span>Career Momentum</span>
                    <TrendingUp className="w-4 h-4 text-indigo-400" />
                  </div>
                  <div className="flex items-baseline gap-3 mt-2">
                    <span className="text-4xl font-extrabold text-slate-900">{report.activity.momentum_score}</span>
                    <span className="text-xs text-emerald-400 font-medium">+14 pts this week</span>
                  </div>
                  <p className="text-xs text-slate-500 mt-2">
                    High activity pacing: regular application follow-ups and scheduled interviews increase your offer likelihood.
                  </p>
                </div>
                <div className="w-full bg-slate-100 rounded-full h-2 mt-4 overflow-hidden">
                  <div
                    className="bg-indigo-500 h-2 rounded-full transition-all duration-500"
                    style={{ width: `${report.activity.momentum_score}%` }}
                  />
                </div>
              </div>

              <div className="bg-white p-5 rounded-xl border border-slate-200">
                <span className="text-xs text-slate-500 font-medium">Applied Today</span>
                <p className="text-3xl font-bold text-slate-900 mt-2">{report.activity.applied_today}</p>
                <span className="text-[11px] text-slate-500 mt-1 block">Within daily quota bounds</span>
              </div>

              <div className="bg-white p-5 rounded-xl border border-slate-200">
                <span className="text-xs text-slate-500 font-medium">Interviews Active</span>
                <p className="text-3xl font-bold text-emerald-400 mt-2">{report.activity.interviews_scheduled}</p>
                <span className="text-[11px] text-emerald-500/80 mt-1 block">1 upcoming tomorrow</span>
              </div>

              <div className="bg-white p-5 rounded-xl border border-slate-200">
                <span className="text-xs text-slate-500 font-medium">Follow-ups Due</span>
                <p className="text-3xl font-bold text-amber-400 mt-2">{report.activity.follow_ups_due}</p>
                <span className="text-[11px] text-amber-500/80 mt-1 block">&gt; 7 days without response</span>
              </div>
            </div>

            {/* Interviews & Follow-ups Section */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Upcoming Interviews Card */}
              <div className="bg-white/80 rounded-xl border border-slate-200 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-semibold text-slate-900 flex items-center gap-2">
                    <Calendar className="w-4 h-4 text-emerald-400" />
                    Upcoming Interviews
                  </h3>
                  <span className="text-xs bg-emerald-500/10 text-emerald-300 px-2.5 py-0.5 rounded-full border border-emerald-500/20">
                    {report.interviews.length} Scheduled
                  </span>
                </div>

                {report.interviews.length === 0 ? (
                  <p className="text-xs text-slate-500 py-6 text-center">No interviews scheduled today.</p>
                ) : (
                  <div className="space-y-3">
                    {report.interviews.map(int => (
                      <div key={int.id} className="p-4 bg-slate-50 rounded-lg border border-slate-200/80 space-y-2">
                        <div className="flex items-start justify-between">
                          <div>
                            <h4 className="text-sm font-semibold text-slate-900">{int.role}</h4>
                            <p className="text-xs text-slate-500 font-medium">{int.company} • {int.round}</p>
                          </div>
                          <span className="text-xs bg-slate-100 text-indigo-300 font-mono px-2 py-1 rounded">
                            Tomorrow 2:00 PM EDT
                          </span>
                        </div>
                        {int.prep_notes && (
                          <div className="text-xs bg-white/90 text-slate-700 p-2.5 rounded border border-slate-200 font-sans">
                            <span className="text-indigo-400 font-semibold block mb-0.5">Focus Brief:</span>
                            {int.prep_notes}
                          </div>
                        )}
                        {int.meeting_link && (
                          <div className="pt-1 flex justify-end">
                            <a
                              href={int.meeting_link}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1 font-medium"
                            >
                              Join Meeting Link <ExternalLink className="w-3 h-3" />
                            </a>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Follow-ups Due Card */}
              <div className="bg-white/80 rounded-xl border border-slate-200 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-semibold text-slate-900 flex items-center gap-2">
                    <Clock className="w-4 h-4 text-amber-400" />
                    Stalled Application Follow-ups (&gt;7 Days)
                  </h3>
                  <span className="text-xs bg-amber-500/10 text-amber-300 px-2.5 py-0.5 rounded-full border border-amber-500/20">
                    {report.follow_ups.length} Actionable
                  </span>
                </div>

                {report.follow_ups.length === 0 ? (
                  <p className="text-xs text-slate-500 py-6 text-center">All applications are actively in progress.</p>
                ) : (
                  <div className="space-y-3">
                    {report.follow_ups.map(f => (
                      <div key={f.application_id} className="p-4 bg-slate-50 rounded-lg border border-slate-200/80 space-y-2">
                        <div className="flex items-start justify-between">
                          <div>
                            <h4 className="text-sm font-semibold text-slate-900">{f.role}</h4>
                            <p className="text-xs text-slate-500">{f.company}</p>
                          </div>
                          <span className="text-xs bg-amber-950/40 text-amber-300 border border-amber-800/40 px-2 py-0.5 rounded-full font-medium">
                            {f.days_since_applied} days ago
                          </span>
                        </div>
                        <p className="text-xs text-slate-700">{f.suggested_action}</p>
                        <div className="pt-2 flex items-center justify-between text-xs border-t border-slate-900">
                          <span className="text-slate-500">Contact: {f.recruiter_contact || 'Via LinkedIn InMail'}</span>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => showToast(`Follow-up check-in drafted for ${f.company}.`)}
                              className="text-xs text-indigo-400 hover:text-indigo-300 font-medium px-2 py-1 bg-indigo-950/30 rounded border border-indigo-800/30"
                            >
                              Draft Note
                            </button>
                            <button
                              onClick={() => showToast(`Follow-up for ${f.company} snoozed 3 days.`)}
                              className="text-xs text-slate-500 hover:text-slate-900 px-2 py-1 bg-slate-100 rounded"
                            >
                              Snooze 3d
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Profile Gaps & Top Opportunities */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Profile Gaps Audit Card */}
              <div className="bg-white/80 rounded-xl border border-slate-200 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-semibold text-slate-900 flex items-center gap-2">
                    <AlertTriangle className="w-4 h-4 text-rose-400" />
                    ATS Profile Gaps & Recommended Fixes
                  </h3>
                  <Link
                    href="/career/profile"
                    className="text-xs text-indigo-400 hover:text-indigo-300 font-medium flex items-center gap-1"
                  >
                    Edit Profile <ChevronRight className="w-3.5 h-3.5" />
                  </Link>
                </div>

                <div className="space-y-3">
                  {report.incomplete_profile_items.map((gap, idx) => (
                    <div key={idx} className="p-4 bg-slate-50 rounded-lg border border-slate-200/80 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-semibold uppercase tracking-wider text-rose-400">
                          {gap.section} section
                        </span>
                        <span className="text-[11px] bg-rose-950/40 text-rose-300 px-2 py-0.5 rounded border border-rose-800/40">
                          Severity: {gap.severity}
                        </span>
                      </div>
                      <p className="text-xs text-slate-900 font-medium">{gap.issue}</p>
                      <p className="text-xs text-slate-500">{gap.recommended_action}</p>
                    </div>
                  ))}
                </div>
              </div>

              {/* Job Highlights */}
              <div className="bg-white/80 rounded-xl border border-slate-200 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-semibold text-slate-900 flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-indigo-400" />
                    Top Discovered Matches Today
                  </h3>
                  <Link
                    href="/career/discovery"
                    className="text-xs text-indigo-400 hover:text-indigo-300 font-medium flex items-center gap-1"
                  >
                    Explore All <ChevronRight className="w-3.5 h-3.5" />
                  </Link>
                </div>

                <div className="space-y-3">
                  {report.job_highlights.map(job => (
                    <div key={job.job_id} className="p-4 bg-slate-50 rounded-lg border border-slate-200/80 space-y-2">
                      <div className="flex items-start justify-between">
                        <div>
                          <h4 className="text-sm font-semibold text-slate-900">{job.title}</h4>
                          <p className="text-xs text-slate-500">{job.company} • {job.location}</p>
                        </div>
                        <span className="text-xs bg-indigo-950 text-indigo-300 font-bold px-2 py-1 rounded border border-indigo-800/50">
                          {job.match_score}% Match
                        </span>
                      </div>
                      {job.disclosed_salary && (
                        <p className="text-xs text-emerald-400 font-medium">Salary: {job.disclosed_salary}</p>
                      )}
                      <div className="pt-2 flex justify-end">
                        <Link
                          href={`/career/tailor?jobId=${job.job_id}`}
                          className="text-xs bg-indigo-600 hover:bg-indigo-500 text-slate-900 font-medium px-3 py-1.5 rounded transition"
                        >
                          Tailor Resume & Cover Letter
                        </Link>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* TAB 2: ACTIVE REMINDERS */}
        {activeTab === 'reminders' && (
          <div className="space-y-4">
            <div className="bg-white rounded-xl border border-slate-200 p-5">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-base font-semibold text-slate-900 flex items-center gap-2">
                  <Bell className="w-4 h-4 text-indigo-400" />
                  Scheduled Career Reminders ({reminders.length})
                </h3>
                <button
                  onClick={() => setShowAddReminderModal(true)}
                  className="text-xs bg-indigo-600 hover:bg-indigo-500 text-slate-900 font-medium px-3 py-1.5 rounded transition flex items-center gap-1"
                >
                  <Plus className="w-3.5 h-3.5" /> Add Reminder
                </button>
              </div>

              <div className="divide-y divide-slate-800/80">
                {reminders.map(rem => (
                  <div key={rem.id} className="py-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span
                          className={`text-[10px] font-semibold uppercase px-2 py-0.5 rounded border ${
                            rem.priority === 'high'
                              ? 'bg-rose-950 text-rose-300 border-rose-800'
                              : rem.priority === 'medium'
                              ? 'bg-amber-950 text-amber-300 border-amber-800'
                              : 'bg-slate-100 text-slate-700 border-slate-300'
                          }`}
                        >
                          {rem.priority} priority
                        </span>
                        <span className="text-xs text-slate-500 font-mono">[{rem.type}]</span>
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full ${
                            rem.status === 'completed'
                              ? 'bg-emerald-950 text-emerald-300'
                              : rem.status === 'snoozed'
                              ? 'bg-amber-950 text-amber-300'
                              : 'bg-indigo-950 text-indigo-300'
                          }`}
                        >
                          {rem.status}
                        </span>
                      </div>
                      <h4 className={`text-sm font-semibold ${rem.status === 'completed' ? 'line-through text-slate-500' : 'text-slate-900'}`}>
                        {rem.title}
                      </h4>
                      <p className="text-xs text-slate-500">{rem.description}</p>
                      <span className="text-[11px] text-slate-500 block">
                        Due: {new Date(rem.due_at).toLocaleString()}
                        {rem.snoozed_until && ` • Snoozed until: ${new Date(rem.snoozed_until).toLocaleString()}`}
                      </span>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      {rem.status !== 'completed' && (
                        <>
                          <button
                            onClick={() => handleReminderAction(rem.id, 'completed')}
                            className="text-xs bg-emerald-600 hover:bg-emerald-500 text-slate-900 font-medium px-3 py-1.5 rounded transition flex items-center gap-1"
                          >
                            <Check className="w-3.5 h-3.5" /> Done
                          </button>
                          <button
                            onClick={() => handleReminderAction(rem.id, 'snoozed', 24)}
                            className="text-xs bg-slate-100 hover:bg-slate-700 text-slate-700 font-medium px-3 py-1.5 rounded transition flex items-center gap-1"
                          >
                            <Clock className="w-3.5 h-3.5" /> Snooze 24h
                          </button>
                          <button
                            onClick={() => handleReminderAction(rem.id, 'dismissed')}
                            className="text-xs text-slate-500 hover:text-slate-700 px-2 py-1.5 transition"
                          >
                            Dismiss
                          </button>
                        </>
                      )}
                      {rem.status === 'completed' && (
                        <button
                          onClick={() => handleReminderAction(rem.id, 'pending')}
                          className="text-xs bg-slate-100 hover:bg-slate-700 text-slate-700 font-medium px-3 py-1.5 rounded transition flex items-center gap-1"
                        >
                          <RotateCcw className="w-3.5 h-3.5" /> Re-open
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* MODAL: SCHEDULE & TIMEZONE SETTINGS (AT-018) */}
        {showConfigModal && (
          <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-white border border-slate-200 rounded-xl max-w-lg w-full p-6 space-y-5 shadow-2xl animate-in fade-in zoom-in-95">
              <div className="flex items-center justify-between border-b border-slate-200 pb-3">
                <h3 className="text-base font-bold text-slate-900 flex items-center gap-2">
                  <Globe className="w-4 h-4 text-indigo-400" />
                  Daily Report Delivery & Timezone Settings (AT-018)
                </h3>
                <button onClick={() => setShowConfigModal(false)} className="text-slate-500 hover:text-slate-900">
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleSaveConfig} className="space-y-4 text-sm">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Timezone (AT-018 DST Compliant)</label>
                  <select
                    value={config.timezone}
                    onChange={e => setConfig({ ...config, timezone: e.target.value })}
                    className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-slate-900 text-xs focus:ring-1 focus:ring-indigo-500"
                  >
                    <option value="America/New_York">America/New_York (EDT / EST - Eastern Time)</option>
                    <option value="America/Chicago">America/Chicago (CDT / CST - Central Time)</option>
                    <option value="America/Los_Angeles">America/Los_Angeles (PDT / PST - Pacific Time)</option>
                    <option value="Europe/London">Europe/London (BST / GMT)</option>
                    <option value="Europe/Berlin">Europe/Berlin (CEST / CET)</option>
                    <option value="Asia/Dhaka">Asia/Dhaka (BST +06:00)</option>
                    <option value="UTC">UTC (Universal Coordinated Time)</option>
                  </select>
                  <span className="text-[11px] text-slate-500 mt-1 block">
                    Calculations automatically adapt to Daylight Saving Time without multiplying budgets or losing tasks.
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Delivery Hour (0-23)</label>
                    <input
                      type="number"
                      min={0}
                      max={23}
                      value={config.scheduled_hour}
                      onChange={e => setConfig({ ...config, scheduled_hour: parseInt(e.target.value) || 0 })}
                      className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-slate-900 text-xs"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Delivery Frequency</label>
                    <select
                      value={config.delivery_days}
                      onChange={e => setConfig({ ...config, delivery_days: e.target.value as 'weekdays' | 'daily' })}
                      className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-slate-900 text-xs"
                    >
                      <option value="weekdays">Weekdays Only (Mon-Fri)</option>
                      <option value="daily">Every Day (7 Days)</option>
                    </select>
                  </div>
                </div>

                <div className="border-t border-slate-200 pt-3 space-y-2">
                  <span className="text-xs font-semibold text-slate-700 block">Notification Channels (AT-010)</span>
                  <label className="flex items-center gap-2 text-xs text-slate-700">
                    <input
                      type="checkbox"
                      checked={config.channels.in_app}
                      disabled
                      className="rounded bg-slate-50 border-slate-200 text-indigo-600"
                    />
                    In-App Notification Drawer (Always Active)
                  </label>

                  <label className="flex items-center gap-2 text-xs text-slate-700">
                    <input
                      type="checkbox"
                      checked={config.channels.webhook_enabled}
                      onChange={e =>
                        setConfig({
                          ...config,
                          channels: { ...config.channels, webhook_enabled: e.target.checked },
                        })
                      }
                      className="rounded bg-slate-50 border-slate-200 text-indigo-600"
                    />
                    External Webhook (Slack / Discord / Zapier)
                  </label>
                  {config.channels.webhook_enabled && (
                    <input
                      type="url"
                      placeholder="https://hooks.slack.com/services/..."
                      value={config.channels.webhook_url || ''}
                      onChange={e =>
                        setConfig({
                          ...config,
                          channels: { ...config.channels, webhook_url: e.target.value },
                        })
                      }
                      className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2 text-slate-900 text-xs mt-1"
                    />
                  )}
                </div>

                <div className="flex justify-end gap-2 pt-4 border-t border-slate-200">
                  <button
                    type="button"
                    onClick={() => setShowConfigModal(false)}
                    className="px-4 py-2 bg-slate-100 hover:bg-slate-700 text-slate-700 rounded-lg text-xs"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-slate-900 font-semibold rounded-lg text-xs"
                  >
                    Save Preferences
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* MODAL: ADD CUSTOM REMINDER */}
        {showAddReminderModal && (
          <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-white border border-slate-200 rounded-xl max-w-md w-full p-6 space-y-4 shadow-2xl animate-in fade-in zoom-in-95">
              <div className="flex items-center justify-between border-b border-slate-200 pb-3">
                <h3 className="text-base font-bold text-slate-900 flex items-center gap-2">
                  <Plus className="w-4 h-4 text-indigo-400" />
                  Create Career Reminder
                </h3>
                <button onClick={() => setShowAddReminderModal(false)} className="text-slate-500 hover:text-slate-900">
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleCreateReminder} className="space-y-4 text-sm">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Reminder Title</label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Follow up on interview feedback..."
                    value={newReminderTitle}
                    onChange={e => setNewReminderTitle(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-slate-900 text-xs focus:ring-1 focus:ring-indigo-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Description & Target Notes</label>
                  <textarea
                    rows={3}
                    placeholder="Provide details or questions to ask..."
                    value={newReminderDesc}
                    onChange={e => setNewReminderDesc(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-slate-900 text-xs focus:ring-1 focus:ring-indigo-500"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Category</label>
                    <select
                      value={newReminderType}
                      onChange={e => setNewReminderType(e.target.value as ReminderType)}
                      className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2 text-slate-900 text-xs"
                    >
                      <option value="custom">General Reminder</option>
                      <option value="follow_up">Application Follow-up</option>
                      <option value="interview">Interview Preparation</option>
                      <option value="profile_incomplete">Profile Improvement</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Priority</label>
                    <select
                      value={newReminderPriority}
                      onChange={e => setNewReminderPriority(e.target.value as ReminderPriority)}
                      className="w-full bg-slate-50 border border-slate-200 rounded-lg p-2 text-slate-900 text-xs"
                    >
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                      <option value="urgent">Urgent</option>
                    </select>
                  </div>
                </div>

                <div className="flex justify-end gap-2 pt-4 border-t border-slate-200">
                  <button
                    type="button"
                    onClick={() => setShowAddReminderModal(false)}
                    className="px-4 py-2 bg-slate-100 hover:bg-slate-700 text-slate-700 rounded-lg text-xs"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-slate-900 font-semibold rounded-lg text-xs"
                  >
                    Create Reminder
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
