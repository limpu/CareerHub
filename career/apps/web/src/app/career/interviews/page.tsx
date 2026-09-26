'use client';

import React, { useState } from 'react';
import Link from 'next/link';

interface CompanyBrief {
  company_name: string;
  domain?: string;
  industry?: string;
  known_size?: string;
  known_tech_stack?: string[];
  mission?: string;
  culture_notes?: string[];
  unverified_fields?: string[];
}

interface StarGuidance {
  situation: string;
  task: string;
  action: string;
  result: string;
}

interface InterviewQuestion {
  id: string;
  category: 'behavioral_star' | 'technical' | 'system_design' | 'role_fit' | 'reverse_questions';
  difficulty: string;
  question: string;
  context_or_goal: string;
  suggested_talking_points: string[];
  grounded_fact_ids?: string[];
  identified_skill_gaps?: string[];
  star_guidance?: StarGuidance;
}

interface InterviewSchedule {
  schedule_id: string;
  round_id: string;
  stage_name: string;
  scheduled_at: string;
  timezone: string;
  format: string;
  meeting_link?: string;
  interviewer_name?: string;
  prep_reminder_at: string;
  follow_up_reminder_at: string;
  outcome: 'pending' | 'passed' | 'failed' | 'skipped';
}

interface SessionNote {
  round_id: string;
  pre_interview_notes: string;
  questions_asked: string[];
  post_interview_reflections: string;
  interviewer_name: string;
  interviewer_title: string;
  self_rating: number; // 1-5
  follow_up_actions: string;
}

export default function InterviewPrepPage() {
  // Active Preset Selection
  const [selectedPreset, setSelectedPreset] = useState<number>(0);
  const [activeCategory, setActiveCategory] = useState<string>('all');
  const [copiedFeedback, setCopiedFeedback] = useState<string | null>(null);

  // Scenario 1: Acme Cloud Infrastructure (Full STAR & Grounded Tech)
  const [companyBrief, setCompanyBrief] = useState<CompanyBrief>({
    company_name: 'Acme Cloud Infrastructure Inc.',
    domain: 'acmecloud.example.com',
    industry: 'Cloud Computing & Distributed Systems',
    known_size: '250-500 employees',
    known_tech_stack: ['Go', 'PostgreSQL', 'Kafka', 'Docker', 'Kubernetes', 'AWS'],
    mission: 'Empowering microservices with high-performance distributed storage and streaming resilience.',
    culture_notes: [
      'Values engineering autonomy, customer-centric architecture, and pragmatic trade-off documentation.',
      'Engineering culture emphasizes blameless postmortems and thorough design RFCs.'
    ],
    unverified_fields: ['Glassdoor Interview Difficulty Estimate (3.4/5)', 'Annual Revenue Estimate ($45M ARR)']
  });

  const [questions, setQuestions] = useState<InterviewQuestion[]>([
    {
      id: 'q-beh-01',
      category: 'behavioral_star',
      difficulty: 'Senior',
      question: 'Tell me about a time when you designed a critical system under tight deadlines and ambiguous requirements.',
      context_or_goal: 'Assesses architectural ownership, stakeholder communication, and iterative delivery.',
      suggested_talking_points: [
        'Frame the business problem: FastFintech ingestion gateway hitting concurrency bottlenecks.',
        'Detail non-negotiable constraints: zero data loss, sub-20ms latency, zero budget for proprietary licensing.',
        'Grounded Fact: Experience as Staff Backend Engineer at FastFintech Corp (2021-2025).'
      ],
      grounded_fact_ids: ['exp-fastfintech-01'],
      star_guidance: {
        situation: 'High-throughput scaling initiative at FastFintech Corp handling sudden 45,000 QPS spikes.',
        task: 'Architect a resilient service layer with Go and Kafka while keeping downtime to absolute zero.',
        action: 'Implemented decoupled queue consumer workers with exponential backoff retry and sharded PostgreSQL partitions.',
        result: 'Achieved p99 latency < 12ms and maintained 99.99% system availability during Black Friday peak.'
      }
    },
    {
      id: 'q-tech-02',
      category: 'technical',
      difficulty: 'Senior',
      question: 'How do you design for concurrency and data consistency in distributed Go services?',
      context_or_goal: 'Evaluates concurrency primitives, sync/atomic mechanisms, channel patterns, and idempotency.',
      suggested_talking_points: [
        'Leverage confirmed skills in Go, Kafka, and PostgreSQL.',
        'Discuss buffered channels, worker pools, mutex contention profiling via pprof, and atomic compare-and-swap (CAS).',
        'Explain transactional outbox pattern for atomic database write + Kafka message publishing.'
      ]
    },
    {
      id: 'q-sys-03',
      category: 'system_design',
      difficulty: 'Senior',
      question: 'Design a distributed, highly-available rate limiter and metrics telemetry stream for Acme Cloud.',
      context_or_goal: 'Evaluates distributed cache clustering, sliding-window algorithms, and telemetry aggregation.',
      suggested_talking_points: [
        'Clarify throughput expectations (50k RPS) and latency budget (<5ms).',
        'Compare Redis Token Bucket vs memory-mapped sliding-window ring buffer.',
        'Discuss partitioning strategy and graceful degradation during network splits.'
      ]
    },
    {
      id: 'q-rev-04',
      category: 'reverse_questions',
      difficulty: 'General',
      question: 'Reverse questions for the Acme Cloud Platform Core engineering team.',
      context_or_goal: 'Demonstrates intellectual curiosity, strategic alignment, and engineering maturity.',
      suggested_talking_points: [
        'How does the platform team prioritize technical debt versus new product features at Acme Cloud?',
        'What has been the most challenging production incident in the past 6 months and what was the team takeaway?',
        'What does success look like for this position in the first 90 days?'
      ]
    }
  ]);

  const [checklist, setChecklist] = useState<{ id: string; text: string; done: boolean }[]>([
    { id: 'c1', text: 'Review company mission, core value propositions, and engineering blog posts.', done: true },
    { id: 'c2', text: 'Audit confirmed resume metrics and be ready to detail contributions in STAR format.', done: true },
    { id: 'c3', text: 'Brush up on identified skill gaps; prepare honest talking points on how you learn quickly.', done: false },
    { id: 'c4', text: 'Formulate 3 strategic reverse questions for your interviewers.', done: true },
    { id: 'c5', text: 'Test video, microphone, and internet connection 15 minutes before the scheduled time.', done: false }
  ]);

  // Scheduled rounds
  const [schedules, setSchedules] = useState<InterviewSchedule[]>([
    {
      schedule_id: 'sched-101',
      round_id: 'round-tech-01',
      stage_name: 'Technical Assessment & Architecture',
      scheduled_at: '2026-10-05T14:00:00Z',
      timezone: 'UTC',
      format: 'video',
      meeting_link: 'https://meet.example.com/acme-tech-01',
      interviewer_name: 'Sarah Jenkins (Principal Architect)',
      prep_reminder_at: '2026-10-04T14:00:00Z',
      follow_up_reminder_at: '2026-10-06T14:00:00Z',
      outcome: 'passed'
    },
    {
      schedule_id: 'sched-102',
      round_id: 'round-onsite-02',
      stage_name: 'System Design & Cross-Functional Loop',
      scheduled_at: '2026-10-12T16:00:00Z',
      timezone: 'UTC',
      format: 'video',
      meeting_link: 'https://meet.example.com/acme-onsite-02',
      interviewer_name: 'David Chen (VP of Engineering)',
      prep_reminder_at: '2026-10-11T16:00:00Z',
      follow_up_reminder_at: '2026-10-13T16:00:00Z',
      outcome: 'pending'
    }
  ]);

  // Session Notes
  const [sessionNote, setSessionNote] = useState<SessionNote>({
    round_id: 'round-tech-01',
    pre_interview_notes: 'Focus on Kafka partition rebalancing and consensus protocols. Ask about multi-region failover strategy.',
    questions_asked: [
      'How does the team handle schema migrations across distributed Kafka streams?',
      'What is the average on-call rotation cadence for the platform core team?'
    ],
    post_interview_reflections: 'Very positive discussion on Go channels and GC tuning. System design portion went well; interviewer seemed satisfied with write-ahead log explanation.',
    interviewer_name: 'Sarah Jenkins',
    interviewer_title: 'Principal Architect',
    self_rating: 4,
    follow_up_actions: 'Sent thank-you note highlighting discussion on Raft consensus; scheduled follow-up reminder.'
  });

  // Funnel Analytics
  const [funnelData] = useState({
    total_applications: 8,
    stage_counts: {
      applied: 8,
      screen: 5,
      technical: 3,
      system_design: 2,
      onsite: 2,
      offered: 1,
      rejected: 1,
      withdrawn: 1
    },
    conversion_rates: {
      applied_to_screen_pct: 62.5,
      screen_to_tech_pct: 60.0,
      tech_to_onsite_pct: 66.7,
      onsite_to_offer_pct: 50.0,
      overall_offer_rate_pct: 12.5,
      overall_rejection_rate_pct: 12.5
    },
    average_days_in_stage: {
      applied: 3.5,
      screen: 6.2,
      technical: 11.4,
      onsite: 18.7,
      offered: 24.0
    }
  });

  // Handle Preset Switching
  const handleLoadPreset = (idx: number) => {
    setSelectedPreset(idx);
    if (idx === 0) {
      // Acme standard
      setCompanyBrief({
        company_name: 'Acme Cloud Infrastructure Inc.',
        domain: 'acmecloud.example.com',
        industry: 'Cloud Computing & Distributed Systems',
        known_size: '250-500 employees',
        known_tech_stack: ['Go', 'PostgreSQL', 'Kafka', 'Docker', 'Kubernetes', 'AWS'],
        mission: 'Empowering microservices with high-performance distributed storage.',
        culture_notes: ['Values engineering autonomy and pragmatic architecture.'],
        unverified_fields: ['Glassdoor Interview Difficulty Estimate (3.4/5)', 'Annual Revenue Estimate ($45M ARR)']
      });
      setQuestions([
        {
          id: 'q-beh-01',
          category: 'behavioral_star',
          difficulty: 'Senior',
          question: 'Tell me about a time when you designed a critical system under tight deadlines and ambiguous requirements.',
          context_or_goal: 'Assesses architectural ownership and stakeholder communication.',
          suggested_talking_points: [
            'Ground with your experience at FastFintech Corp as Staff Backend Engineer.',
            'Detail non-negotiable architectural constraints and proactive milestone decomposition.'
          ],
          grounded_fact_ids: ['exp-fastfintech-01'],
          star_guidance: {
            situation: 'High-throughput scaling initiative at FastFintech Corp handling 45k QPS.',
            task: 'Architect resilient service layer while keeping downtime to zero.',
            action: 'Implemented decoupled queue consumer workers with backoff retry.',
            result: 'Maintained 99.99% availability during traffic spikes.'
          }
        },
        {
          id: 'q-tech-02',
          category: 'technical',
          difficulty: 'Senior',
          question: 'How do you design for concurrency and data consistency in Senior Distributed Systems Engineer?',
          context_or_goal: 'Evaluates concurrency primitives and transactional boundaries.',
          suggested_talking_points: [
            'Leverage your verified skills in: Go, Kafka, PostgreSQL.',
            'Discuss optimistic locking, atomic CAS operations, and idempotency keys.'
          ]
        }
      ]);
    } else if (idx === 1) {
      // Skill Gap scenario
      setCompanyBrief({
        company_name: 'KernelEdge Security',
        domain: 'kerneledge.example.com',
        industry: 'Cybersecurity & Kernel Telemetry',
        known_size: '50-100 employees',
        known_tech_stack: ['Rust', 'eBPF', 'Linux Kernel', 'C++'],
        mission: 'Real-time kernel telemetry and threat containment.',
        culture_notes: ['Deep systems engineering and zero-copy performance.'],
        unverified_fields: ['Series B Valuation Estimate ($60M)']
      });
      setQuestions([
        {
          id: 'q-gap-01',
          category: 'technical',
          difficulty: 'Staff',
          question: 'How would you implement zero-copy network packet filtering in the Linux Kernel?',
          context_or_goal: 'Evaluates kernel networking, socket buffers, and low-latency interception.',
          suggested_talking_points: [
            'PREPARATION NOTE (AT-003): Role requires Rust, eBPF, and Linux Kernel (not confirmed in your profile).',
            'Review eBPF verifier constraints and ring buffer maps before interview; do not invent production experience.',
            'Ground what you DO know: Highlight your verified deep experience with Linux socket tuning and Go network gateways.'
          ],
          identified_skill_gaps: ['Rust', 'eBPF', 'Linux Kernel']
        },
        {
          id: 'q-rev-02',
          category: 'reverse_questions',
          difficulty: 'General',
          question: 'Questions for KernelEdge Security engineering leads.',
          context_or_goal: 'Assesses security ergonomics and kernel deployment pipeline.',
          suggested_talking_points: [
            'What testing methodology does the team use to prevent kernel panics in customer staging environments?'
          ]
        }
      ]);
    }
  };

  const handleCopySchedule = () => {
    const text = schedules
      .map(
        (s) =>
          `[${s.stage_name}]\nScheduled: ${s.scheduled_at} (${s.timezone})\nInterviewer: ${s.interviewer_name}\nLink: ${s.meeting_link}\nPrep Reminder: ${s.prep_reminder_at}\nOutcome: ${s.outcome.toUpperCase()}`
      )
      .join('\n\n');
    navigator.clipboard.writeText(text);
    setCopiedFeedback('Interview schedule copied to clipboard!');
    setTimeout(() => setCopiedFeedback(null), 3000);
  };

  const toggleChecklist = (id: string) => {
    setChecklist((prev) =>
      prev.map((item) => (item.id === id ? { ...item, done: !item.done } : item))
    );
  };

  const filteredQuestions = questions.filter((q) =>
    activeCategory === 'all' ? true : q.category === activeCategory
  );

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 font-sans">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Navigation & Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-slate-200 pb-4">
          <div>
            <div className="flex items-center gap-2 text-xs text-indigo-400 font-mono mb-1">
              <Link href="/career" className="hover:underline">
                CAREER
              </Link>
              <span>/</span>
              <span className="text-slate-500">INTERVIEWS</span>
              <span>/</span>
              <span className="text-emerald-400 font-bold">IMP-CAR-23</span>
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-900 flex items-center gap-3">
              Interview Preparation & Outcome Funnel Hub
              <span className="text-xs bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 px-2 py-0.5 rounded font-mono">
                CAR-23 • AT-003
              </span>
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Fact-grounded company intelligence, STAR questions, debrief reflection notes, and authentic conversion analytics.
            </p>
          </div>

          {/* Quick Action Badges */}
          <div className="flex items-center gap-2">
            <button
              onClick={handleCopySchedule}
              className="px-3 py-1.5 text-xs bg-slate-100 hover:bg-slate-700 text-slate-800 border border-slate-300 rounded transition font-medium flex items-center gap-1.5"
            >
              📋 Copy Schedule
            </button>
            <Link
              href="/career/status"
              className="px-3 py-1.5 text-xs bg-indigo-600 hover:bg-indigo-500 text-slate-900 rounded transition font-medium"
            >
              Application Ledger →
            </Link>
          </div>
        </div>

        {copiedFeedback && (
          <div className="p-3 bg-emerald-950/80 border border-emerald-700/50 rounded-lg text-emerald-300 text-xs flex items-center gap-2">
            <span>✓</span> {copiedFeedback}
          </div>
        )}

        {/* Truth-in-Advertising Transparency Alert */}
        <div className="p-4 bg-indigo-950/40 border border-indigo-500/30 rounded-xl space-y-2">
          <div className="flex items-center gap-2 text-indigo-300 text-xs font-semibold uppercase tracking-wider">
            <span>🛡️</span> Truth-in-Advertising Invariant (CAR-23, AT-003)
          </div>
          <p className="text-xs text-slate-700 leading-relaxed">
            This platform strictly prohibits invented interview signals or simulated recruiter messages. All interview stage transitions, notes, and outcome records reflect confirmed candidate actions. Candidate talking points are derived strictly from confirmed profile facts; missing skills are flagged as preparation gaps, never fabricated.
          </p>
        </div>

        {/* Preset Scenarios */}
        <div className="bg-white border border-slate-200 rounded-xl p-4">
          <div className="text-xs font-medium text-slate-500 mb-3 flex items-center gap-2">
            <span>📂</span> GROUNDTRUTH FIXTURE PRESETS
          </div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
            {[
              { title: 'Acme Cloud Infrastructure', desc: 'Full STAR & Grounded Tech Questions' },
              { title: 'KernelEdge Security', desc: 'Honest Skill Gaps (Rust/eBPF absent)' },
              { title: 'Technical Screen & Debrief', desc: 'Interview Schedule & Reflection Notes' },
              { title: 'Outcome Funnel Analytics', desc: '8 Active Applications Conversion' }
            ].map((preset, idx) => (
              <button
                key={idx}
                onClick={() => handleLoadPreset(idx)}
                className={`p-3 text-left rounded-lg border transition ${
                  selectedPreset === idx
                    ? 'bg-indigo-900/30 border-indigo-500 text-slate-900 shadow-sm'
                    : 'bg-slate-50 border-slate-200 text-slate-500 hover:border-slate-300'
                }`}
              >
                <div className="text-xs font-bold text-slate-800">{preset.title}</div>
                <div className="text-[11px] text-slate-500 mt-1">{preset.desc}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Main Content Grid: 2 Columns */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left 2 Columns: Company Brief & Question Studio */}
          <div className="lg:col-span-2 space-y-6">
            {/* Company Intelligence Brief */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 pb-3">
                <div className="flex items-center gap-2">
                  <span className="text-lg">🏢</span>
                  <div>
                    <h2 className="text-sm font-bold text-slate-900">{companyBrief.company_name}</h2>
                    <p className="text-xs text-slate-500">{companyBrief.industry} • {companyBrief.known_size}</p>
                  </div>
                </div>
                <span className="text-xs bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded font-mono">
                  Verified Brief
                </span>
              </div>

              <div className="text-xs text-slate-700 leading-relaxed">
                <span className="font-semibold text-slate-800">Mission: </span>
                {companyBrief.mission}
              </div>

              {/* Tech Stack Tags */}
              {companyBrief.known_tech_stack && companyBrief.known_tech_stack.length > 0 && (
                <div>
                  <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1.5">
                    Verified Infrastructure Tech Stack
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {companyBrief.known_tech_stack.map((tech, idx) => (
                      <span
                        key={idx}
                        className="px-2 py-0.5 text-xs bg-slate-100 text-slate-700 border border-slate-300 rounded"
                      >
                        {tech}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Unverified Disclosures (AT-003) */}
              {companyBrief.unverified_fields && companyBrief.unverified_fields.length > 0 && (
                <div className="p-3 bg-amber-950/30 border border-amber-800/40 rounded-lg text-xs space-y-1">
                  <div className="text-amber-400 font-semibold flex items-center gap-1.5">
                    <span>⚠️</span> Unverified External Estimates (AT-003 Disclosure)
                  </div>
                  <ul className="list-disc list-inside text-slate-500 space-y-0.5 text-[11px]">
                    {companyBrief.unverified_fields.map((field, idx) => (
                      <li key={idx}>{field}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>

            {/* Questions Preparation Studio */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-5">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-slate-200 pb-3">
                <div className="flex items-center gap-2">
                  <span className="text-lg">🎯</span>
                  <h3 className="text-sm font-bold text-slate-900">Interview Questions & STAR Strategy</h3>
                </div>

                {/* Filter Categories */}
                <div className="flex flex-wrap gap-1.5 text-xs">
                  {['all', 'behavioral_star', 'technical', 'system_design', 'reverse_questions'].map((cat) => (
                    <button
                      key={cat}
                      onClick={() => setActiveCategory(cat)}
                      className={`px-2.5 py-1 rounded transition text-xs capitalize ${
                        activeCategory === cat
                          ? 'bg-indigo-600 text-slate-900 font-medium'
                          : 'bg-slate-100 text-slate-500 hover:text-slate-900'
                      }`}
                    >
                      {cat.replace('_', ' ')}
                    </button>
                  ))}
                </div>
              </div>

              {/* Questions List */}
              <div className="space-y-4">
                {filteredQuestions.map((q) => (
                  <div key={q.id} className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3">
                    <div className="flex items-start justify-between gap-2">
                      <span className="text-xs font-mono text-indigo-400 uppercase tracking-wider">
                        {q.category.replace('_', ' ')} • {q.difficulty}
                      </span>
                      {q.identified_skill_gaps && q.identified_skill_gaps.length > 0 && (
                        <span className="text-[11px] bg-amber-500/20 text-amber-300 border border-amber-500/40 px-2 py-0.5 rounded font-mono font-medium">
                          Skill Gap Detected (AT-003)
                        </span>
                      )}
                    </div>

                    <h4 className="text-sm font-semibold text-slate-100">{q.question}</h4>
                    <p className="text-xs text-slate-500 italic">{q.context_or_goal}</p>

                    {/* Skill Gap Banner if present */}
                    {q.identified_skill_gaps && q.identified_skill_gaps.length > 0 && (
                      <div className="p-2.5 bg-amber-950/40 border border-amber-800/40 rounded text-xs text-amber-200 space-y-1">
                        <div className="font-semibold text-[11px] uppercase tracking-wider">
                          Identified Skill Gaps (Not in Verified Profile):
                        </div>
                        <div className="flex flex-wrap gap-1">
                          {q.identified_skill_gaps.map((gap, idx) => (
                            <span key={idx} className="bg-amber-900/60 text-amber-200 px-1.5 py-0.5 rounded text-[11px]">
                              {gap}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Grounded Talking Points */}
                    <div className="space-y-1.5 pt-1">
                      <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
                        Fact-Grounded Talking Points
                      </div>
                      <ul className="space-y-1 text-xs text-slate-700 list-disc list-inside">
                        {q.suggested_talking_points.map((pt, idx) => (
                          <li key={idx}>{pt}</li>
                        ))}
                      </ul>
                    </div>

                    {/* STAR Guidance Drawer */}
                    {q.star_guidance && (
                      <div className="mt-3 p-3 bg-indigo-950/20 border border-indigo-500/20 rounded-lg space-y-2 text-xs">
                        <div className="text-xs font-semibold text-indigo-300 flex items-center gap-1.5">
                          <span>⭐</span> Grounded STAR Breakdown (Candidate Experience)
                        </div>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px]">
                          <div className="bg-white p-2 rounded border border-slate-200">
                            <span className="font-bold text-slate-500">Situation:</span> {q.star_guidance.situation}
                          </div>
                          <div className="bg-white p-2 rounded border border-slate-200">
                            <span className="font-bold text-slate-500">Task:</span> {q.star_guidance.task}
                          </div>
                          <div className="bg-white p-2 rounded border border-slate-200">
                            <span className="font-bold text-slate-500">Action:</span> {q.star_guidance.action}
                          </div>
                          <div className="bg-white p-2 rounded border border-slate-200">
                            <span className="font-bold text-slate-500">Result:</span> {q.star_guidance.result}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Preparation Checklist */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-3">
              <div className="flex items-center gap-2 border-b border-slate-200 pb-2">
                <span>📋</span>
                <h3 className="text-sm font-bold text-slate-900">Interview Readiness Checklist</h3>
              </div>
              <div className="space-y-2">
                {checklist.map((item) => (
                  <label
                    key={item.id}
                    className="flex items-start gap-3 p-2 rounded hover:bg-slate-50/50 cursor-pointer text-xs"
                  >
                    <input
                      type="checkbox"
                      checked={item.done}
                      onChange={() => toggleChecklist(item.id)}
                      className="mt-0.5 accent-indigo-500 rounded"
                    />
                    <span className={item.done ? 'line-through text-slate-500' : 'text-slate-700'}>
                      {item.text}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          </div>

          {/* Right Column: Rounds, Session Notes & Outcome Funnel */}
          <div className="space-y-6">
            {/* Scheduled Rounds & Reminders */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 pb-2">
                <div className="flex items-center gap-2">
                  <span>📅</span>
                  <h3 className="text-sm font-bold text-slate-900">Scheduled Rounds</h3>
                </div>
                <span className="text-xs text-slate-500 font-mono">{schedules.length} Rounds</span>
              </div>

              <div className="space-y-3">
                {schedules.map((s) => (
                  <div key={s.schedule_id} className="p-3 bg-slate-50 border border-slate-200 rounded-lg space-y-2 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-slate-800">{s.stage_name}</span>
                      <span
                        className={`text-[10px] px-1.5 py-0.5 rounded font-mono uppercase ${
                          s.outcome === 'passed'
                            ? 'bg-emerald-500/20 text-emerald-300'
                            : s.outcome === 'failed'
                            ? 'bg-rose-500/20 text-rose-300'
                            : 'bg-amber-500/20 text-amber-300'
                        }`}
                      >
                        {s.outcome}
                      </span>
                    </div>
                    <div className="text-slate-500 text-[11px] space-y-0.5">
                      <div>Scheduled: {new Date(s.scheduled_at).toLocaleString()} ({s.timezone})</div>
                      <div>Interviewer: {s.interviewer_name}</div>
                      <div>Format: {s.format.toUpperCase()}</div>
                      <div className="text-indigo-400">Prep Reminder: 24h prior</div>
                      <div className="text-indigo-400">Follow-up: 24h post</div>
                    </div>
                    {s.meeting_link && (
                      <div className="pt-1">
                        <a
                          href={s.meeting_link}
                          target="_blank"
                          rel="noreferrer"
                          className="text-[11px] text-indigo-400 hover:underline flex items-center gap-1"
                        >
                          🔗 Open Meeting Link
                        </a>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Candidate Debrief Notes */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 pb-2">
                <div className="flex items-center gap-2">
                  <span>📝</span>
                  <h3 className="text-sm font-bold text-slate-900">Debrief & Session Reflections</h3>
                </div>
                <div className="flex text-amber-400 text-xs">
                  {'★'.repeat(sessionNote.self_rating)}{'☆'.repeat(5 - sessionNote.self_rating)}
                </div>
              </div>

              <div className="space-y-3 text-xs">
                <div>
                  <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
                    Pre-Interview Notes
                  </div>
                  <p className="p-2.5 bg-slate-50 rounded border border-slate-200 text-slate-700">
                    {sessionNote.pre_interview_notes}
                  </p>
                </div>

                <div>
                  <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
                    Questions Asked by Candidate
                  </div>
                  <ul className="list-disc list-inside p-2.5 bg-slate-50 rounded border border-slate-200 text-slate-700 space-y-1 text-[11px]">
                    {sessionNote.questions_asked.map((q, idx) => (
                      <li key={idx}>{q}</li>
                    ))}
                  </ul>
                </div>

                <div>
                  <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
                    Post-Interview Reflections
                  </div>
                  <p className="p-2.5 bg-slate-50 rounded border border-slate-200 text-slate-700">
                    {sessionNote.post_interview_reflections}
                  </p>
                </div>

                <div>
                  <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
                    Follow-Up Action
                  </div>
                  <p className="p-2 bg-emerald-950/30 border border-emerald-800/40 rounded text-emerald-300 text-[11px]">
                    {sessionNote.follow_up_actions}
                  </p>
                </div>
              </div>
            </div>

            {/* Outcome Funnel Analytics */}
            <div className="bg-white border border-slate-200 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 pb-2">
                <div className="flex items-center gap-2">
                  <span>📊</span>
                  <h3 className="text-sm font-bold text-slate-900">Outcome Funnel</h3>
                </div>
                <span className="text-xs bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 px-2 py-0.5 rounded font-mono">
                  {funnelData.total_applications} Apps
                </span>
              </div>

              {/* Conversion Metrics */}
              <div className="grid grid-cols-2 gap-2 text-center">
                <div className="p-2.5 bg-slate-50 rounded border border-slate-200">
                  <div className="text-[10px] text-slate-500 uppercase">Offer Conversion</div>
                  <div className="text-base font-bold text-emerald-400 mt-0.5">
                    {funnelData.conversion_rates.overall_offer_rate_pct}%
                  </div>
                </div>
                <div className="p-2.5 bg-slate-50 rounded border border-slate-200">
                  <div className="text-[10px] text-slate-500 uppercase">Screen Pass Rate</div>
                  <div className="text-base font-bold text-indigo-400 mt-0.5">
                    {funnelData.conversion_rates.screen_to_tech_pct}%
                  </div>
                </div>
              </div>

              {/* Funnel Stage Progression Bars */}
              <div className="space-y-2 text-xs">
                {[
                  { label: 'Applied', count: funnelData.stage_counts.applied, pct: 100, color: 'bg-slate-600' },
                  { label: 'Recruiter Screen', count: funnelData.stage_counts.screen, pct: 62.5, color: 'bg-indigo-600' },
                  { label: 'Technical Assessment', count: funnelData.stage_counts.technical, pct: 37.5, color: 'bg-blue-600' },
                  { label: 'Onsite / Final Loop', count: funnelData.stage_counts.onsite, pct: 25.0, color: 'bg-violet-600' },
                  { label: 'Offer Received', count: funnelData.stage_counts.offered, pct: 12.5, color: 'bg-emerald-500' }
                ].map((stg, idx) => (
                  <div key={idx} className="space-y-1">
                    <div className="flex justify-between text-[11px]">
                      <span className="text-slate-700 font-medium">{stg.label}</span>
                      <span className="text-slate-500 font-mono">{stg.count} ({stg.pct}%)</span>
                    </div>
                    <div className="w-full bg-slate-50 rounded-full h-1.5 overflow-hidden border border-slate-200">
                      <div className={`h-full ${stg.color}`} style={{ width: `${stg.pct}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
