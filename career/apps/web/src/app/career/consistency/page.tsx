'use client';

import React, { useState } from 'react';
import Link from 'next/link';

export type FactSource = 'master_resume' | 'master_profile' | 'linkedin_profile';
export type MismatchedFieldType = 'job_title' | 'employment_date' | 'company_name' | 'skill_coverage' | 'education';
export type MismatchSeverity = 'high' | 'medium' | 'low';
export type ResolutionAction = 'retain_resume' | 'retain_linkedin' | 'retain_profile' | 'custom_override';

export interface FactSourceValue {
  source: FactSource;
  value: string;
  context?: string;
  observed_at: string;
}

export interface ResolutionDecision {
  action: ResolutionAction;
  chosen_value: string;
  resolved_at?: string;
  user_justification: string;
}

export interface ConsistencyMismatch {
  mismatch_id: string;
  field_type: MismatchedFieldType;
  severity: MismatchSeverity;
  entity_key: string;
  description: string;
  source_values: FactSourceValue[];
  resolution?: ResolutionDecision;
}

export interface ConsistencyAuditReport {
  report_id: string;
  user_id: string;
  mismatches: ConsistencyMismatch[];
  total_mismatches: number;
  high_severity_count: number;
  medium_severity_count: number;
  low_severity_count: number;
  resolved_count: number;
  overall_consistency_score: number;
  generated_at: string;
}

interface AuditLogEntry {
  id: string;
  field: string;
  old_value: string;
  new_value: string;
  action: string;
  reason: string;
  timestamp: string;
}

const PRESET_SCENARIOS: { name: string; description: string; report: ConsistencyAuditReport }[] = [
  {
    name: 'Discrepancy Scenario (Title & Date Mismatch)',
    description: 'High and medium severity title & date divergence between Resume, Profile, and LinkedIn.',
    report: {
      report_id: 'rep-scenario-01',
      user_id: 'usr-cand-101',
      total_mismatches: 3,
      high_severity_count: 1,
      medium_severity_count: 2,
      low_severity_count: 0,
      resolved_count: 0,
      overall_consistency_score: 72.5,
      generated_at: new Date().toISOString(),
      mismatches: [
        {
          mismatch_id: 'mis-title-fastfintech',
          field_type: 'job_title',
          severity: 'high',
          entity_key: 'experience:FastFintech Corp',
          description: 'Significant title discrepancy: Resume indicates Staff Backend Engineer while LinkedIn and Profile state Senior Backend Engineer.',
          source_values: [
            { source: 'master_resume', value: 'Staff Backend Engineer', context: 'Latest Tailored Master Resume', observed_at: '2026-03-15T10:00:00Z' },
            { source: 'master_profile', value: 'Senior Backend Engineer', context: 'Canonical Profile Experience', observed_at: '2026-02-01T12:00:00Z' },
            { source: 'linkedin_profile', value: 'Senior Backend Engineer', context: 'LinkedIn Snapshot (Synced)', observed_at: '2026-03-10T14:30:00Z' }
          ]
        },
        {
          mismatch_id: 'mis-date-fastfintech',
          field_type: 'employment_date',
          severity: 'medium',
          entity_key: 'experience:FastFintech Corp',
          description: 'Start date variance of 2 months between resume and profile records.',
          source_values: [
            { source: 'master_resume', value: '2021-03-01', context: 'Resume start date', observed_at: '2026-03-15T10:00:00Z' },
            { source: 'master_profile', value: '2021-01-15', context: 'Canonical profile start date', observed_at: '2026-02-01T12:00:00Z' },
            { source: 'linkedin_profile', value: '2021-01-01', context: 'LinkedIn start month', observed_at: '2026-03-10T14:30:00Z' }
          ]
        },
        {
          mismatch_id: 'mis-title-cloudscale',
          field_type: 'job_title',
          severity: 'medium',
          entity_key: 'experience:CloudScale Systems',
          description: 'Seniority nuance: "Systems Engineer II" vs "Software Engineer".',
          source_values: [
            { source: 'master_resume', value: 'Systems Engineer II', context: 'Resume title', observed_at: '2026-03-15T10:00:00Z' },
            { source: 'master_profile', value: 'Software Engineer', context: 'Profile entry', observed_at: '2026-02-01T12:00:00Z' },
            { source: 'linkedin_profile', value: 'Systems Engineer', context: 'LinkedIn Headline tag', observed_at: '2026-03-10T14:30:00Z' }
          ]
        }
      ]
    }
  },
  {
    name: 'Skill Coverage Differences',
    description: 'Minor low-severity gaps where LinkedIn lists extra technical skills omitted from Master Resume.',
    report: {
      report_id: 'rep-scenario-02',
      user_id: 'usr-cand-102',
      total_mismatches: 2,
      high_severity_count: 0,
      medium_severity_count: 0,
      low_severity_count: 2,
      resolved_count: 0,
      overall_consistency_score: 89.0,
      generated_at: new Date().toISOString(),
      mismatches: [
        {
          mismatch_id: 'mis-skill-rust',
          field_type: 'skill_coverage',
          severity: 'low',
          entity_key: 'skill:Rust',
          description: 'Skill "Rust" is featured on LinkedIn profile snapshot but missing from canonical Master Profile.',
          source_values: [
            { source: 'linkedin_profile', value: 'Rust (Proficient)', context: 'LinkedIn Endorsed Skills', observed_at: '2026-03-12T09:00:00Z' },
            { source: 'master_profile', value: 'Not Listed', context: 'Master Profile Skills Section', observed_at: '2026-03-10T12:00:00Z' },
            { source: 'master_resume', value: 'Not Listed', context: 'Active Resume', observed_at: '2026-03-15T10:00:00Z' }
          ]
        },
        {
          mismatch_id: 'mis-skill-terraform',
          field_type: 'skill_coverage',
          severity: 'low',
          entity_key: 'skill:Terraform',
          description: 'Skill "Terraform" is featured on LinkedIn profile snapshot but missing from canonical Master Profile.',
          source_values: [
            { source: 'linkedin_profile', value: 'Terraform (Intermediate)', context: 'LinkedIn Endorsed Skills', observed_at: '2026-03-12T09:00:00Z' },
            { source: 'master_profile', value: 'Not Listed', context: 'Master Profile Skills Section', observed_at: '2026-03-10T12:00:00Z' },
            { source: 'master_resume', value: 'Terraform (Projects)', context: 'Tailored Resume Bullet', observed_at: '2026-03-15T10:00:00Z' }
          ]
        }
      ]
    }
  },
  {
    name: 'Perfect Consistency (Fully Aligned)',
    description: 'All 3 sources (Resume, Profile, and LinkedIn) reflect identical career facts and chronology.',
    report: {
      report_id: 'rep-scenario-03',
      user_id: 'usr-cand-103',
      total_mismatches: 0,
      high_severity_count: 0,
      medium_severity_count: 0,
      low_severity_count: 0,
      resolved_count: 0,
      overall_consistency_score: 100.0,
      generated_at: new Date().toISOString(),
      mismatches: []
    }
  }
];

export default function ConsistencyAuditPage() {
  const [selectedPresetIndex, setSelectedPresetIndex] = useState<number>(0);
  const [currentReport, setCurrentReport] = useState<ConsistencyAuditReport>(PRESET_SCENARIOS[0].report);
  const [auditTrail, setAuditTrail] = useState<AuditLogEntry[]>([]);
  const [resolutionForms, setResolutionForms] = useState<Record<string, { action: ResolutionAction; customValue: string; justification: string }>>({});
  const [copiedNotification, setCopiedNotification] = useState<string | null>(null);
  const [filterSeverity, setFilterSeverity] = useState<string>('all');

  const handleSelectPreset = (index: number) => {
    setSelectedPresetIndex(index);
    setCurrentReport(JSON.parse(JSON.stringify(PRESET_SCENARIOS[index].report)));
    setResolutionForms({});
  };

  const handleFormChange = (mismatchId: string, field: 'action' | 'customValue' | 'justification', value: string) => {
    setResolutionForms((prev) => ({
      ...prev,
      [mismatchId]: {
        action: (prev[mismatchId]?.action || 'retain_profile') as ResolutionAction,
        customValue: prev[mismatchId]?.customValue || '',
        justification: prev[mismatchId]?.justification || '',
        [field]: value
      }
    }));
  };

  const handleResolveMismatch = (mismatch: ConsistencyMismatch) => {
    const form = resolutionForms[mismatch.mismatch_id] || {
      action: 'retain_profile',
      customValue: '',
      justification: ''
    };

    if (!form.justification.trim()) {
      alert('Candidate justification is required before reconciliation (AT-007 Human-in-the-loop requirement).');
      return;
    }

    let chosenVal = '';
    if (form.action === 'retain_resume') {
      chosenVal = mismatch.source_values.find((s) => s.source === 'master_resume')?.value || '';
    } else if (form.action === 'retain_linkedin') {
      chosenVal = mismatch.source_values.find((s) => s.source === 'linkedin_profile')?.value || '';
    } else if (form.action === 'retain_profile') {
      chosenVal = mismatch.source_values.find((s) => s.source === 'master_profile')?.value || '';
    } else if (form.action === 'custom_override') {
      chosenVal = form.customValue.trim();
      if (!chosenVal) {
        alert('Custom override value cannot be empty.');
        return;
      }
    }

    const decision: ResolutionDecision = {
      action: form.action,
      chosen_value: chosenVal,
      resolved_at: new Date().toISOString(),
      user_justification: form.justification.trim()
    };

    // Update Report State
    const updatedMismatches = currentReport.mismatches.map((m) => {
      if (m.mismatch_id === mismatch.mismatch_id) {
        return { ...m, resolution: decision };
      }
      return m;
    });

    const newResolvedCount = updatedMismatches.filter((m) => m.resolution).length;
    const remainingUnresolved = updatedMismatches.length - newResolvedCount;
    // Recalculate Health Score
    const recalculatedScore = Math.min(100, Math.round(currentReport.overall_consistency_score + (100 - currentReport.overall_consistency_score) * (newResolvedCount / (updatedMismatches.length || 1))));

    setCurrentReport({
      ...currentReport,
      mismatches: updatedMismatches,
      resolved_count: newResolvedCount,
      overall_consistency_score: recalculatedScore
    });

    // Record Profile Field Audit Entry (AT-007)
    const newAuditEntry: AuditLogEntry = {
      id: `audit-${Date.now()}`,
      field: mismatch.entity_key,
      old_value: mismatch.source_values.find((s) => s.source === 'master_profile')?.value || 'N/A',
      new_value: chosenVal,
      action: form.action,
      reason: form.justification.trim(),
      timestamp: new Date().toLocaleTimeString()
    };

    setAuditTrail((prev) => [newAuditEntry, ...prev]);
  };

  const handleCopyText = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopiedNotification(label);
    setTimeout(() => setCopiedNotification(null), 3000);
  };

  const filteredMismatches = currentReport.mismatches.filter((m) => {
    if (filterSeverity === 'all') return true;
    return m.severity === filterSeverity;
  });

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      {/* Header Bar */}
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
          <div>
            <div className="flex items-center space-x-3">
              <Link href="/career" className="text-sm font-medium text-indigo-600 hover:text-indigo-800">
                &larr; Career Hub
              </Link>
              <span className="text-slate-300">|</span>
              <span className="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-semibold text-amber-800 border border-amber-200">
                IMP-CAR-24 &bull; CAR-24 &bull; AT-007
              </span>
            </div>
            <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
              Cross-Source Career Consistency Check
            </h1>
            <p className="mt-1 text-sm text-slate-500">
              Verify and align titles, dates, and skill coverage across Master Resume, Canonical Profile, and LinkedIn snapshot.
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <button
              onClick={() => handleCopyText(JSON.stringify(currentReport, null, 2), 'Audit Report JSON')}
              className="inline-flex items-center rounded-md border border-slate-300 bg-white px-3.5 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              Export Report JSON
            </button>
          </div>
        </div>
      </header>

      {copiedNotification && (
        <div className="fixed bottom-4 right-4 z-50 rounded-lg bg-emerald-600 px-4 py-3 text-sm font-medium text-white shadow-lg">
          Copied {copiedNotification} to clipboard!
        </div>
      )}

      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
        {/* Scenario Switcher */}
        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
            Interactive Test Fixtures & Presets
          </h2>
          <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-3">
            {PRESET_SCENARIOS.map((preset, idx) => (
              <button
                key={preset.name}
                onClick={() => handleSelectPreset(idx)}
                className={`flex flex-col text-left p-4 rounded-lg border transition-all ${
                  selectedPresetIndex === idx
                    ? 'border-indigo-600 bg-indigo-50/50 shadow-sm'
                    : 'border-slate-200 hover:border-slate-300 bg-white'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-slate-900 text-sm">{preset.name}</span>
                  {selectedPresetIndex === idx && (
                    <span className="h-2 w-2 rounded-full bg-indigo-600"></span>
                  )}
                </div>
                <p className="mt-1 text-xs text-slate-500">{preset.description}</p>
                <div className="mt-2 text-xs font-medium text-indigo-700">
                  Initial Score: {preset.report.overall_consistency_score}% &bull; {preset.report.total_mismatches} Mismatches
                </div>
              </button>
            ))}
          </div>
        </section>

        {/* Health & Score Overview Banner */}
        <section className="grid grid-cols-1 gap-6 sm:grid-cols-4">
          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Consistency Score</span>
            <div className="mt-2 flex items-baseline space-x-2">
              <span className={`text-4xl font-extrabold ${
                currentReport.overall_consistency_score >= 90
                  ? 'text-emerald-600'
                  : currentReport.overall_consistency_score >= 75
                  ? 'text-amber-600'
                  : 'text-rose-600'
              }`}>
                {currentReport.overall_consistency_score}%
              </span>
              <span className="text-xs text-slate-500">
                {currentReport.overall_consistency_score >= 90 ? 'Aligned' : 'Action Required'}
              </span>
            </div>
            <div className="mt-3 w-full bg-slate-100 rounded-full h-2">
              <div
                className={`h-2 rounded-full ${
                  currentReport.overall_consistency_score >= 90
                    ? 'bg-emerald-500'
                    : currentReport.overall_consistency_score >= 75
                    ? 'bg-amber-500'
                    : 'bg-rose-500'
                }`}
                style={{ width: `${currentReport.overall_consistency_score}%` }}
              ></div>
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">High Severity</span>
            <div className="mt-2 text-3xl font-bold text-rose-600">
              {currentReport.high_severity_count}
            </div>
            <p className="mt-1 text-xs text-slate-500">
              Conflicting company names or divergent titles.
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Medium Severity</span>
            <div className="mt-2 text-3xl font-bold text-amber-600">
              {currentReport.medium_severity_count}
            </div>
            <p className="mt-1 text-xs text-slate-500">
              Employment date overlaps or title seniority nuances.
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Reconciled Status</span>
            <div className="mt-2 text-3xl font-bold text-indigo-600">
              {currentReport.resolved_count} / {currentReport.total_mismatches}
            </div>
            <p className="mt-1 text-xs text-slate-500">
              Human-in-the-loop decisions confirmed (AT-007).
            </p>
          </div>
        </section>

        {/* Truth-in-Advertising & LinkedIn Sync Policy Notice (CAR-24, AT-010) */}
        <section className="rounded-xl border border-blue-200 bg-blue-50/50 p-5">
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 text-blue-600 pt-0.5">
              <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="text-xs text-blue-900 leading-relaxed">
              <p className="font-semibold text-sm text-blue-950">Truth-in-Advertising & LinkedIn Write Limitations (AT-010):</p>
              Automated third-party background mutation of live LinkedIn profiles violates LinkedIn terms of service without enterprise OAuth write scopes.
              When reconciling facts, the platform updates your <strong>Canonical Master Career Profile</strong> and <strong>Master Resume</strong>. For LinkedIn, use the generated <em>Manual Sync Snippet</em> below to copy-paste exact reconciled facts.
            </div>
          </div>
        </section>

        {/* Discrepancies & Resolution Matrix */}
        <section className="rounded-xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-200 px-6 py-4 flex flex-wrap items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-bold text-slate-900">Detected Fact Discrepancies</h2>
              <p className="text-xs text-slate-500">Inspect mismatched records and apply your authoritative decision.</p>
            </div>
            <div className="flex items-center space-x-2">
              <span className="text-xs font-medium text-slate-500">Filter Severity:</span>
              <select
                value={filterSeverity}
                onChange={(e) => setFilterSeverity(e.target.value)}
                className="rounded-md border border-slate-300 bg-white px-2.5 py-1 text-xs text-slate-700 shadow-sm focus:border-indigo-500 focus:outline-none"
              >
                <option value="all">All Severities</option>
                <option value="high">High Severity Only</option>
                <option value="medium">Medium Severity Only</option>
                <option value="low">Low Severity Only</option>
              </select>
            </div>
          </div>

          <div className="divide-y divide-slate-200">
            {filteredMismatches.length === 0 ? (
              <div className="p-12 text-center">
                <svg className="mx-auto h-12 w-12 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h3 className="mt-2 text-sm font-semibold text-slate-900">Zero Inconsistencies Found</h3>
                <p className="mt-1 text-xs text-slate-500">All cross-source records are fully reconciled and aligned.</p>
              </div>
            ) : (
              filteredMismatches.map((mismatch) => {
                const isResolved = Boolean(mismatch.resolution);
                const formState = resolutionForms[mismatch.mismatch_id] || {
                  action: 'retain_profile',
                  customValue: '',
                  justification: ''
                };

                return (
                  <div key={mismatch.mismatch_id} className={`p-6 transition-colors ${isResolved ? 'bg-emerald-50/20' : 'bg-white'}`}>
                    <div className="flex items-start justify-between">
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${
                            mismatch.severity === 'high'
                              ? 'bg-rose-100 text-rose-800'
                              : mismatch.severity === 'medium'
                              ? 'bg-amber-100 text-amber-800'
                              : 'bg-slate-100 text-slate-800'
                          }`}>
                            {mismatch.severity} Severity
                          </span>
                          <span className="font-mono text-xs text-slate-400">&bull; {mismatch.entity_key}</span>
                          <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-medium text-slate-600">
                            Field: {mismatch.field_type.replace('_', ' ')}
                          </span>
                        </div>
                        <h3 className="mt-2 font-semibold text-slate-900 text-base">{mismatch.description}</h3>
                      </div>
                      {isResolved && (
                        <span className="inline-flex items-center rounded-md bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">
                          &check; Reconciled & Synced
                        </span>
                      )}
                    </div>

                    {/* Tri-Source Inspection Table */}
                    <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
                      {mismatch.source_values.map((src) => (
                        <div key={src.source} className="rounded-lg border border-slate-200 bg-slate-50/70 p-3.5">
                          <div className="flex items-center justify-between">
                            <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
                              {src.source === 'master_resume' && 'Master Resume'}
                              {src.source === 'master_profile' && 'Canonical Profile'}
                              {src.source === 'linkedin_profile' && 'LinkedIn Snapshot'}
                            </span>
                            <span className="text-[10px] text-slate-400">
                              {new Date(src.observed_at).toLocaleDateString()}
                            </span>
                          </div>
                          <div className="mt-1 font-semibold text-slate-900 text-sm break-words">
                            {src.value}
                          </div>
                          {src.context && (
                            <div className="mt-1 text-xs text-slate-500 italic">
                              Context: {src.context}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>

                    {/* Human Resolution Gate (AT-007) */}
                    {!isResolved ? (
                      <div className="mt-5 rounded-lg border border-indigo-100 bg-indigo-50/30 p-4">
                        <h4 className="text-xs font-bold uppercase tracking-wider text-indigo-900">
                          Human-in-the-Loop Resolution Gate (AT-007)
                        </h4>
                        <p className="mt-1 text-xs text-slate-600">
                          Select the authoritative fact to retain or specify a custom reconciled value. This will update your profile and write to the immutable field audit trail.
                        </p>

                        <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-4">
                          <label className={`flex items-center space-x-2 rounded-md border p-2.5 text-xs font-medium cursor-pointer ${
                            formState.action === 'retain_profile' ? 'border-indigo-600 bg-white text-indigo-900' : 'border-slate-200 bg-slate-50 text-slate-700'
                          }`}>
                            <input
                              type="radio"
                              name={`action-${mismatch.mismatch_id}`}
                              value="retain_profile"
                              checked={formState.action === 'retain_profile'}
                              onChange={(e) => handleFormChange(mismatch.mismatch_id, 'action', e.target.value)}
                              className="text-indigo-600 focus:ring-indigo-500"
                            />
                            <span>Retain Profile Fact</span>
                          </label>

                          <label className={`flex items-center space-x-2 rounded-md border p-2.5 text-xs font-medium cursor-pointer ${
                            formState.action === 'retain_resume' ? 'border-indigo-600 bg-white text-indigo-900' : 'border-slate-200 bg-slate-50 text-slate-700'
                          }`}>
                            <input
                              type="radio"
                              name={`action-${mismatch.mismatch_id}`}
                              value="retain_resume"
                              checked={formState.action === 'retain_resume'}
                              onChange={(e) => handleFormChange(mismatch.mismatch_id, 'action', e.target.value)}
                              className="text-indigo-600 focus:ring-indigo-500"
                            />
                            <span>Retain Resume Fact</span>
                          </label>

                          <label className={`flex items-center space-x-2 rounded-md border p-2.5 text-xs font-medium cursor-pointer ${
                            formState.action === 'retain_linkedin' ? 'border-indigo-600 bg-white text-indigo-900' : 'border-slate-200 bg-slate-50 text-slate-700'
                          }`}>
                            <input
                              type="radio"
                              name={`action-${mismatch.mismatch_id}`}
                              value="retain_linkedin"
                              checked={formState.action === 'retain_linkedin'}
                              onChange={(e) => handleFormChange(mismatch.mismatch_id, 'action', e.target.value)}
                              className="text-indigo-600 focus:ring-indigo-500"
                            />
                            <span>Retain LinkedIn Fact</span>
                          </label>

                          <label className={`flex items-center space-x-2 rounded-md border p-2.5 text-xs font-medium cursor-pointer ${
                            formState.action === 'custom_override' ? 'border-indigo-600 bg-white text-indigo-900' : 'border-slate-200 bg-slate-50 text-slate-700'
                          }`}>
                            <input
                              type="radio"
                              name={`action-${mismatch.mismatch_id}`}
                              value="custom_override"
                              checked={formState.action === 'custom_override'}
                              onChange={(e) => handleFormChange(mismatch.mismatch_id, 'action', e.target.value)}
                              className="text-indigo-600 focus:ring-indigo-500"
                            />
                            <span>Custom Override</span>
                          </label>
                        </div>

                        {formState.action === 'custom_override' && (
                          <div className="mt-3">
                            <label className="block text-xs font-semibold text-slate-700">Specified Reconciled Value</label>
                            <input
                              type="text"
                              value={formState.customValue}
                              onChange={(e) => handleFormChange(mismatch.mismatch_id, 'customValue', e.target.value)}
                              placeholder="e.g. Lead Backend Engineer & System Architect"
                              className="mt-1 w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-xs text-slate-900 shadow-sm focus:border-indigo-500 focus:outline-none"
                            />
                          </div>
                        )}

                        <div className="mt-3">
                          <label className="block text-xs font-semibold text-slate-700">
                            Candidate Justification / Change Notes (Required for Audit Trail)
                          </label>
                          <input
                            type="text"
                            value={formState.justification}
                            onChange={(e) => handleFormChange(mismatch.mismatch_id, 'justification', e.target.value)}
                            placeholder="e.g. Master resume has exact official corporate promotion title verified by HR."
                            className="mt-1 w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-xs text-slate-900 shadow-sm focus:border-indigo-500 focus:outline-none"
                          />
                        </div>

                        <div className="mt-4 flex justify-end">
                          <button
                            onClick={() => handleResolveMismatch(mismatch)}
                            className="inline-flex items-center rounded-md bg-indigo-600 px-4 py-2 text-xs font-bold text-white shadow hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                          >
                            Confirm Decision & Reconcile Profile
                          </button>
                        </div>
                      </div>
                    ) : (
                      <div className="mt-4 rounded-lg bg-emerald-50 border border-emerald-200 p-3.5">
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-bold text-emerald-900">
                            Authoritative Retained Value: &ldquo;{mismatch.resolution?.chosen_value}&rdquo;
                          </span>
                          <span className="text-emerald-700 font-mono">
                            Action: {mismatch.resolution?.action}
                          </span>
                        </div>
                        <p className="mt-1 text-xs text-emerald-800 italic">
                          Justification: &ldquo;{mismatch.resolution?.user_justification}&rdquo;
                        </p>
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </section>

        {/* Audit Log / Propagation History */}
        {auditTrail.length > 0 && (
          <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <h3 className="text-sm font-bold text-slate-900">Field Audit Trail Log (AT-007)</h3>
            <p className="mt-1 text-xs text-slate-500">Granular modifications recorded with candidate justification and timestamps.</p>
            <div className="mt-4 overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200 text-xs">
                <thead>
                  <tr className="bg-slate-50">
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">Timestamp</th>
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">Entity Field</th>
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">Old Value</th>
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">New Reconciled Value</th>
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">Action</th>
                    <th className="px-3 py-2 text-left font-semibold text-slate-600">Candidate Justification</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 bg-white">
                  {auditTrail.map((entry) => (
                    <tr key={entry.id}>
                      <td className="whitespace-nowrap px-3 py-2 font-mono text-slate-500">{entry.timestamp}</td>
                      <td className="whitespace-nowrap px-3 py-2 font-medium text-slate-900">{entry.field}</td>
                      <td className="whitespace-nowrap px-3 py-2 text-slate-500">{entry.old_value}</td>
                      <td className="whitespace-nowrap px-3 py-2 font-semibold text-indigo-700">{entry.new_value}</td>
                      <td className="whitespace-nowrap px-3 py-2 text-slate-600">{entry.action}</td>
                      <td className="px-3 py-2 text-slate-600 max-w-xs truncate" title={entry.reason}>{entry.reason}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
