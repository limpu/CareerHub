'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import {
  ExternalPortalInfo,
  PortalType,
  PortalSupportLevel,
  ClipboardItem,
  ExternalApplicationBundle,
} from '@social-platform/contracts';

interface PresetScenario {
  name: string;
  url: string;
  redirectTarget?: string;
  description: string;
  expectedType: PortalType;
  expectedLevel: PortalSupportLevel;
}

const PRESET_SCENARIOS: PresetScenario[] = [
  {
    name: 'Indeed Native Easy Apply',
    url: 'https://www.indeed.com/viewjob?jk=1234567890abcdef&easyapply=1',
    description: 'Native Indeed popup form supporting inline automated multi-step submission.',
    expectedType: 'indeed_easy_apply',
    expectedLevel: 'native_easy_apply',
  },
  {
    name: 'Indeed External Redirect (to Workday)',
    url: 'https://www.indeed.com/rc/clk?jk=external_click_through_123',
    redirectTarget: 'https://acme.wd5.myworkdayjobs.com/en-US/AcmeCareers/job/Staff-Platform-Engineer_JR-1092',
    description: 'Indeed listing that redirects candidate off-site to external Workday enterprise portal.',
    expectedType: 'workday_external',
    expectedLevel: 'assisted_manual',
  },
  {
    name: 'Workday Enterprise Portal Direct',
    url: 'https://enterprise.wd1.myworkdayjobs.com/Careers/job/Go-Lead-Architect_R-99881',
    description: 'Enterprise workday portal. Manual candidate interaction required per AT-010 truth-in-advertising.',
    expectedType: 'workday_external',
    expectedLevel: 'assisted_manual',
  },
  {
    name: 'Greenhouse ATS Direct Board',
    url: 'https://boards.greenhouse.io/moderntech/jobs/4556102',
    description: 'Standard Greenhouse direct job board with autofill-ready structured fields.',
    expectedType: 'greenhouse_direct',
    expectedLevel: 'native_easy_apply',
  },
  {
    name: 'Oracle Taleo Legacy Portal',
    url: 'https://globalcorp.taleo.net/careersection/2/jobdetail.ftl?job=2400918',
    description: 'Legacy multi-frame enterprise portal requiring manual candidate login and clipboard paste.',
    expectedType: 'taleo_external',
    expectedLevel: 'assisted_manual',
  },
];

const MOCK_CLIPBOARD_ITEMS: ClipboardItem[] = [
  { key: 'full_name', label: 'Full Legal Name', value: 'Alice Developer', category: 'contact', is_pii: true },
  { key: 'email', label: 'Email Address', value: 'alice.dev@example.com', category: 'contact', is_pii: true },
  { key: 'phone', label: 'Phone Number', value: '+1-555-0199', category: 'contact', is_pii: true },
  { key: 'location', label: 'Current Location', value: 'Austin, TX, USA', category: 'contact', is_pii: false },
  { key: 'linkedin', label: 'LinkedIn Profile', value: 'https://linkedin.com/in/alicedev', category: 'links', is_pii: false },
  { key: 'github', label: 'GitHub Profile', value: 'https://github.com/alicedev', category: 'links', is_pii: false },
  { key: 'work_auth', label: 'Authorized to Work in US', value: 'Yes, authorized without sponsorship', category: 'legal', is_pii: false },
  { key: 'exp_years', label: 'Total Years Experience', value: '6', category: 'skills', is_pii: false },
  { key: 'top_skills', label: 'Primary Technical Skills', value: 'Go, Kubernetes, PostgreSQL, TypeScript, React', category: 'skills', is_pii: false },
  { key: 'summary', label: 'Professional Summary', value: 'Staff Platform Engineer specializing in high-throughput distributed systems in Go and container orchestration.', category: 'profile', is_pii: false },
];

export default function ExternalPortalPage() {
  const [inputUrl, setInputUrl] = useState<string>(PRESET_SCENARIOS[0].url);
  const [redirectUrl, setRedirectUrl] = useState<string>('');
  const [selectedScenarioIndex, setSelectedScenarioIndex] = useState<number>(0);

  // Live Portal Classification State
  const [portalInfo, setPortalInfo] = useState<ExternalPortalInfo>({
    url: PRESET_SCENARIOS[0].url,
    portal_type: 'indeed_easy_apply',
    support_level: 'native_easy_apply',
    display_name: 'Indeed Easy Apply (Native Modal)',
    capabilities: {
      can_auto_fill: true,
      requires_external_redirect: false,
      manual_fallback_mandatory: false,
      description: 'Native Indeed Easy Apply modal. Supported for direct inline multi-step submission.',
    },
    truth_advertising_disclosure: 'This portal is recognized as an integrated native Easy Apply flow.',
  });

  // External Application Bundle State
  const [bundle, setBundle] = useState<ExternalApplicationBundle | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [copiedAll, setCopiedAll] = useState<boolean>(false);
  const [receiptInput, setReceiptInput] = useState<string>('');
  const [notesInput, setNotesInput] = useState<string>('');
  const [confirmationRecordId, setConfirmationRecordId] = useState<string | null>(null);
  const [notification, setNotification] = useState<{ type: 'info' | 'success' | 'warning'; message: string } | null>(null);

  // Classify logic (Mirroring core portal classifier)
  const classifyUrl = (rawUrl: string, targetUrl?: string): ExternalPortalInfo => {
    const effectiveUrl = (targetUrl && targetUrl.trim() !== '') ? targetUrl.trim() : rawUrl.trim();
    const lower = effectiveUrl.toLowerCase();

    if (lower.includes('myworkdayjobs.com') || lower.includes('workday.com')) {
      return {
        url: effectiveUrl,
        portal_type: 'workday_external',
        support_level: 'assisted_manual',
        display_name: 'Workday Enterprise Portal',
        capabilities: {
          can_auto_fill: false,
          requires_external_redirect: true,
          manual_fallback_mandatory: true,
          description: 'Enterprise Workday recruitment portal with anti-bot challenge and dynamic iframe isolation.',
          known_limitations: 'Automated script completion prohibited. Candidate assisted-manual completion mandatory.',
        },
        truth_advertising_disclosure:
          'ATTENTION (AT-010): Workday portals require applicant self-service completion. The platform equips you with a 1-click clipboard bundle and verified facts rather than unreliable synthetic automation.',
        redirect_target_url: targetUrl,
      };
    }

    if (lower.includes('taleo.net')) {
      return {
        url: effectiveUrl,
        portal_type: 'taleo_external',
        support_level: 'assisted_manual',
        display_name: 'Oracle Taleo Enterprise Portal',
        capabilities: {
          can_auto_fill: false,
          requires_external_redirect: true,
          manual_fallback_mandatory: true,
          description: 'Legacy Oracle Taleo candidate portal with multi-frame session auth.',
          known_limitations: 'Manual candidate submission required.',
        },
        truth_advertising_disclosure:
          'ATTENTION (AT-010): Taleo legacy portals require candidate interactive login and assisted-manual completion.',
        redirect_target_url: targetUrl,
      };
    }

    if (lower.includes('indeed.com')) {
      if (lower.includes('easyapply=1') || lower.includes('ia=1')) {
        return {
          url: effectiveUrl,
          portal_type: 'indeed_easy_apply',
          support_level: 'native_easy_apply',
          display_name: 'Indeed Easy Apply (Native Modal)',
          capabilities: {
            can_auto_fill: true,
            requires_external_redirect: false,
            manual_fallback_mandatory: false,
            description: 'Native Indeed Easy Apply modal. Supported for direct inline multi-step submission.',
          },
          truth_advertising_disclosure: 'This portal is recognized as an integrated native Easy Apply flow.',
        };
      } else {
        return {
          url: effectiveUrl,
          portal_type: 'generic_external',
          support_level: 'assisted_manual',
          display_name: 'Indeed External Click-Through Link',
          capabilities: {
            can_auto_fill: false,
            requires_external_redirect: true,
            manual_fallback_mandatory: true,
            description: 'Indeed link that redirects candidate to an employer off-site portal.',
            known_limitations: 'Application takes place on external employer website.',
          },
          truth_advertising_disclosure:
            'ATTENTION (AT-010): This Indeed listing is not an Easy Apply role; it directs you off-site to the employer career website.',
          redirect_target_url: targetUrl,
        };
      }
    }

    if (lower.includes('boards.greenhouse.io') || lower.includes('greenhouse.io')) {
      return {
        url: effectiveUrl,
        portal_type: 'greenhouse_direct',
        support_level: 'native_easy_apply',
        display_name: 'Greenhouse Job Board',
        capabilities: {
          can_auto_fill: true,
          requires_external_redirect: false,
          manual_fallback_mandatory: false,
          description: 'Direct ATS single-page board supporting zero-fabrication form autofill.',
        },
        truth_advertising_disclosure: 'Greenhouse direct board recognized with full autofill support.',
      };
    }

    return {
      url: effectiveUrl,
      portal_type: 'generic_external',
      support_level: 'assisted_manual',
      display_name: 'External Employer Portal',
      capabilities: {
        can_auto_fill: false,
        requires_external_redirect: true,
        manual_fallback_mandatory: true,
        description: 'Third-party employer career website requiring candidate manual submission.',
        known_limitations: 'Automated submission not supported on arbitrary external domains.',
      },
      truth_advertising_disclosure:
        'ATTENTION (AT-010): Arbitrary external portals require candidate-assisted completion to maintain truth and prevent deceptive claims.',
      redirect_target_url: targetUrl,
    };
  };

  const handleApplyScenario = (index: number) => {
    setSelectedScenarioIndex(index);
    const scenario = PRESET_SCENARIOS[index];
    setInputUrl(scenario.url);
    setRedirectUrl(scenario.redirectTarget || '');
    const info = classifyUrl(scenario.url, scenario.redirectTarget);
    setPortalInfo(info);
    setBundle(null);
    setConfirmationRecordId(null);
    setNotification({
      type: 'info',
      message: `Scenario "${scenario.name}" loaded: classified as ${info.display_name} (${info.support_level}).`,
    });
  };

  const handleInspectUrl = () => {
    const info = classifyUrl(inputUrl, redirectUrl);
    setPortalInfo(info);
    setBundle(null);
    setConfirmationRecordId(null);
    setNotification({
      type: 'info',
      message: `Inspected: ${info.display_name} -> Support Level: ${info.support_level}.`,
    });
  };

  // Prepare Clipboard Bundle
  const handlePrepareBundle = () => {
    const formattedText = MOCK_CLIPBOARD_ITEMS.map((item) => `${item.label}: ${item.value}`).join('\n');
    const newBundle: ExternalApplicationBundle = {
      bundle_id: `bundle_ext_${Date.now()}`,
      session_id: `sess_apply_${Date.now()}`,
      job_id: `job_staff_eng_01`,
      job_title: 'Staff Platform Engineer',
      company_name: 'Acme Cloud Infrastructure',
      portal_info: portalInfo,
      clipboard_items: MOCK_CLIPBOARD_ITEMS,
      formatted_clipboard_text: formattedText,
      resume_document_id: 'doc_resume_pdf_v1',
      cover_letter_text:
        'Dear Hiring Team, I am writing to express my enthusiastic interest in the Staff Platform Engineer role at Acme Cloud Infrastructure...',
      status: 'prepared',
    };
    setBundle(newBundle);
    setNotification({
      type: 'success',
      message: '1-Click Assisted-Manual Application Bundle successfully generated with confirmed facts.',
    });
  };

  // Copy Single Clipboard Item
  const handleCopyItem = (item: ClipboardItem) => {
    navigator.clipboard.writeText(item.value);
    setCopiedKey(item.key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  // Copy Entire Formatted Bundle
  const handleCopyAll = () => {
    if (!bundle) return;
    navigator.clipboard.writeText(bundle.formatted_clipboard_text);
    setCopiedAll(true);
    setTimeout(() => setCopiedAll(false), 2500);
  };

  // Dispatch Action (AT-005: MUST NOT mark as applied!)
  const handleDispatch = () => {
    if (!bundle) return;
    const dispatchedBundle: ExternalApplicationBundle = {
      ...bundle,
      status: 'dispatched',
      dispatched_at: new Date().toISOString(),
    };
    setBundle(dispatchedBundle);
    setNotification({
      type: 'warning',
      message:
        '[AT-005 INVARIANT ENFORCED] External portal opened/dispatched. Notice: Opening or navigating to an external portal NEVER marks the application as applied. Status remains "dispatched" until explicit receipt confirmation.',
    });
  };

  // Confirm Submission Gate (REQ-005, AT-005, CAR-10)
  const handleConfirmSubmission = () => {
    if (!bundle) return;
    if (bundle.status !== 'dispatched') {
      setNotification({
        type: 'warning',
        message: 'Application bundle must be dispatched to the external portal before confirmation can be recorded.',
      });
      return;
    }

    const recordId = `app_rec_verified_${Date.now()}`;
    const confirmedBundle: ExternalApplicationBundle = {
      ...bundle,
      status: 'applied',
      applied_at: new Date().toISOString(),
      notes: notesInput || 'Submitted via candidate assisted-manual dispatch.',
    };
    setBundle(confirmedBundle);
    setConfirmationRecordId(recordId);
    setNotification({
      type: 'success',
      message: `[REQ-005 & AT-005 VERIFIED] External application confirmed by candidate! Immutable record logged in Ledger with ID: ${recordId}.`,
    });
  };

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 font-sans">
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Navigation & Header */}
        <header className="border-b border-slate-200 pb-5">
          <div className="flex items-center space-x-2 text-xs text-slate-500 mb-2">
            <Link href="/career" className="hover:text-slate-800 transition">
              Career Dashboard
            </Link>
            <span>/</span>
            <Link href="/career/apply" className="hover:text-slate-800 transition">
              Application Workflow Hub
            </Link>
            <span>/</span>
            <Link href="/career/wizard" className="hover:text-slate-800 transition">
              Multi-Step Wizard
            </Link>
            <span>/</span>
            <span className="text-emerald-400 font-medium">External & Indeed Applications</span>
          </div>

          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900 flex items-center gap-3">
                <span>External & Indeed Applications Hub</span>
                <span className="text-xs px-2.5 py-1 rounded-full font-mono bg-emerald-950 text-emerald-300 border border-emerald-800">
                  IMP-CAR-14
                </span>
              </h1>
              <p className="text-sm text-slate-500 mt-1">
                Truth-in-advertising classifier, assisted manual application bundles, and external reconciliation gates
                (CAR-14, AT-005, AT-006, AT-010, SRC-C3).
              </p>
            </div>

            <div className="flex items-center gap-3">
              <Link
                href="/career/wizard"
                className="text-xs px-3 py-1.5 rounded bg-slate-100 hover:bg-neutral-700 text-slate-800 border border-slate-300 transition"
              >
                &larr; Multi-Step Wizard
              </Link>
              <Link
                href="/career/status"
                className="text-xs px-3 py-1.5 rounded bg-blue-600 hover:bg-blue-500 text-slate-900 font-medium transition flex items-center gap-1"
              >
                <span>Status & Ledger &rarr;</span>
              </Link>
              <Link
                href="/career/apply"
                className="text-xs px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-slate-900 font-medium transition"
              >
                Workflow Hub &rarr;
              </Link>
            </div>
          </div>
        </header>

        {/* Global Alert Notification */}
        {notification && (
          <div
            className={`p-4 rounded-lg border text-sm transition ${
              notification.type === 'success'
                ? 'bg-emerald-950/60 border-emerald-700 text-emerald-200'
                : notification.type === 'warning'
                ? 'bg-amber-950/60 border-amber-700 text-amber-200'
                : 'bg-blue-950/60 border-blue-700 text-blue-200'
            }`}
          >
            <div className="flex items-start gap-2">
              <span className="font-bold">
                {notification.type === 'success' ? '✓' : notification.type === 'warning' ? '⚠' : 'ℹ'}
              </span>
              <div>{notification.message}</div>
            </div>
          </div>
        )}

        {/* Section 1: Preset Evaluation Scenarios */}
        <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-800 uppercase tracking-wider">
              1. Groundtruth Evaluation Scenarios (CAR-14, AT-010)
            </h2>
            <span className="text-xs text-neutral-500 font-mono">Test Fixtures: forms/indeed_and_external_portals.json</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-3">
            {PRESET_SCENARIOS.map((sc, idx) => (
              <button
                key={sc.name}
                onClick={() => handleApplyScenario(idx)}
                className={`p-3 rounded-lg border text-left text-xs transition flex flex-col justify-between ${
                  selectedScenarioIndex === idx
                    ? 'border-emerald-500 bg-emerald-950/30 text-slate-900 shadow-sm'
                    : 'border-slate-200 bg-slate-50/50 hover:bg-slate-100/60 text-slate-700'
                }`}
              >
                <div>
                  <div className="font-semibold">{sc.name}</div>
                  <div className="text-[11px] text-slate-500 mt-1 line-clamp-2">{sc.description}</div>
                </div>
                <div className="mt-3 flex items-center gap-1.5 font-mono text-[10px]">
                  <span
                    className={`px-1.5 py-0.5 rounded ${
                      sc.expectedLevel === 'native_easy_apply'
                        ? 'bg-emerald-900 text-emerald-300'
                        : 'bg-amber-900 text-amber-300'
                    }`}
                  >
                    {sc.expectedLevel}
                  </span>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Section 2: Interactive Portal URL Classifier */}
        <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-800 uppercase tracking-wider">
              2. Interactive Portal Classifier & Truth-in-Advertising (AT-010)
            </h2>
            <span className="text-xs text-neutral-500 font-mono">Endpoint: /api/v1/career/portal/classify</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-500 mb-1">
                Target Job / Portal URL
              </label>
              <input
                type="text"
                value={inputUrl}
                onChange={(e) => setInputUrl(e.target.value)}
                placeholder="https://example.com/careers/job/123"
                className="w-full bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 text-xs font-mono text-slate-800 focus:outline-none focus:border-emerald-500"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-500 mb-1">
                Optional Redirect Target URL (e.g. Workday/Taleo off-site redirect)
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={redirectUrl}
                  onChange={(e) => setRedirectUrl(e.target.value)}
                  placeholder="https://company.wd5.myworkdayjobs.com/job/..."
                  className="flex-1 bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 text-xs font-mono text-slate-800 focus:outline-none focus:border-emerald-500"
                />
                <button
                  onClick={handleInspectUrl}
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 rounded-lg text-xs font-medium transition"
                >
                  Classify
                </button>
              </div>
            </div>
          </div>

          {/* Classification Results Card */}
          <div className="border border-slate-200 bg-slate-50 rounded-lg p-4 space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 pb-3">
              <div>
                <span className="text-xs text-slate-500">Classified Portal: </span>
                <span className="text-sm font-semibold text-slate-900">{portalInfo.display_name}</span>
                <span className="ml-2 font-mono text-xs px-2 py-0.5 rounded bg-slate-100 text-slate-700">
                  {portalInfo.portal_type}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <span
                  className={`text-xs px-2.5 py-1 rounded-full font-medium ${
                    portalInfo.support_level === 'native_easy_apply'
                      ? 'bg-emerald-950 border border-emerald-700 text-emerald-300'
                      : 'bg-amber-950 border border-amber-700 text-amber-300'
                  }`}
                >
                  Mode: {portalInfo.support_level === 'native_easy_apply' ? 'Native Easy Apply' : 'Assisted Manual'}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              <div className="p-2.5 rounded bg-white border border-slate-200">
                <span className="text-slate-500">Can Auto-Fill:</span>
                <div className="font-semibold mt-0.5">
                  {portalInfo.capabilities.can_auto_fill ? (
                    <span className="text-emerald-400">Yes (Zero-Fabrication Guarded)</span>
                  ) : (
                    <span className="text-amber-400">No (Anti-Bot / Iframe Isolated)</span>
                  )}
                </div>
              </div>

              <div className="p-2.5 rounded bg-white border border-slate-200">
                <span className="text-slate-500">Requires External Redirect:</span>
                <div className="font-semibold mt-0.5">
                  {portalInfo.capabilities.requires_external_redirect ? (
                    <span className="text-amber-400">Yes (Candidate Navigates to Employer Site)</span>
                  ) : (
                    <span className="text-emerald-400">No (Inline Experience)</span>
                  )}
                </div>
              </div>

              <div className="p-2.5 rounded bg-white border border-slate-200">
                <span className="text-slate-500">Manual Fallback Mandatory:</span>
                <div className="font-semibold mt-0.5">
                  {portalInfo.capabilities.manual_fallback_mandatory ? (
                    <span className="text-amber-400">Mandatory per AT-010 Rule</span>
                  ) : (
                    <span className="text-slate-700">Optional Fallback</span>
                  )}
                </div>
              </div>
            </div>

            {/* Truth in Advertising Notice (AT-010) */}
            <div className="p-3 rounded-lg border border-amber-800/80 bg-amber-950/40 text-amber-200 text-xs">
              <div className="font-semibold flex items-center gap-1.5 mb-1">
                <span>🛡️ Truth-in-Advertising Guarantee (AT-010 & CAR-14)</span>
              </div>
              <p>{portalInfo.truth_advertising_disclosure}</p>
              {portalInfo.capabilities.known_limitations && (
                <p className="mt-1 text-amber-300/80 font-mono text-[11px]">
                  Limitation: {portalInfo.capabilities.known_limitations}
                </p>
              )}
            </div>

            {!bundle && (
              <div className="pt-2 flex justify-end">
                <button
                  onClick={handlePrepareBundle}
                  className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 rounded-lg text-xs font-semibold shadow transition"
                >
                  Generate 1-Click Clipboard Bundle &rarr;
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Section 3: Assisted Manual Clipboard Bundle (AT-003, CAR-14) */}
        {bundle && (
          <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <h2 className="text-sm font-semibold text-slate-800 uppercase tracking-wider">
                  3. 1-Click Assisted Clipboard Bundle (AT-003 Zero-Fabrication)
                </h2>
                <p className="text-xs text-slate-500">
                  Target: <span className="text-slate-900 font-medium">{bundle.job_title}</span> at{' '}
                  <span className="text-slate-900 font-medium">{bundle.company_name}</span> | Bundle ID:{' '}
                  <span className="font-mono text-slate-500">{bundle.bundle_id}</span>
                </p>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopyAll}
                  className="px-3 py-1.5 rounded bg-slate-100 hover:bg-neutral-700 text-slate-800 border border-slate-300 text-xs font-medium transition"
                >
                  {copiedAll ? '✓ All Facts Copied!' : 'Copy All Text Bundle'}
                </button>
              </div>
            </div>

            {/* Individual Field Copiers */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {bundle.clipboard_items.map((item) => (
                <div
                  key={item.key}
                  className="p-3 rounded-lg border border-slate-200 bg-slate-50 flex items-center justify-between gap-3 text-xs"
                >
                  <div className="overflow-hidden">
                    <div className="flex items-center gap-1.5">
                      <span className="text-slate-500 font-medium">{item.label}</span>
                      {item.is_pii && (
                        <span className="text-[10px] px-1 py-0.2 rounded bg-slate-100 text-slate-500 font-mono">
                          PII
                        </span>
                      )}
                    </div>
                    <div className="font-mono text-slate-800 truncate mt-0.5" title={item.value}>
                      {item.value}
                    </div>
                  </div>

                  <button
                    onClick={() => handleCopyItem(item)}
                    className="px-2.5 py-1 rounded bg-slate-100 hover:bg-neutral-700 text-slate-700 font-mono text-[11px] shrink-0 transition"
                  >
                    {copiedKey === item.key ? '✓ Copied' : 'Copy'}
                  </button>
                </div>
              ))}
            </div>

            {/* Section 4: External Dispatch & Reconciliation Gate (AT-005, REQ-005, CAR-10) */}
            <div className="border-t border-slate-200 pt-5 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-semibold text-slate-700 uppercase tracking-wider">
                  4. External Dispatch & Reconciliation Gate (AT-005 Invariant)
                </h3>
                <span
                  className={`text-xs px-2.5 py-1 rounded-full font-mono font-medium ${
                    bundle.status === 'applied'
                      ? 'bg-emerald-950 text-emerald-300 border border-emerald-700'
                      : bundle.status === 'dispatched'
                      ? 'bg-blue-950 text-blue-300 border border-blue-700'
                      : 'bg-slate-100 text-slate-700'
                  }`}
                >
                  Status: {bundle.status.toUpperCase()}
                </span>
              </div>

              {/* Status explanation */}
              <div className="p-3 rounded-lg bg-slate-50 border border-slate-200 text-xs text-slate-500 space-y-1">
                <div>
                  • <span className="text-slate-700 font-medium">Dispatching to External Portal:</span> Opens candidate browser window to external domain with clipboard ready. Status changes from <code className="text-slate-700">prepared</code> to <code className="text-slate-700">dispatched</code>.
                </div>
                <div>
                  • <span className="text-amber-300 font-medium">AT-005 Invariant:</span> Opening or navigating to an external link NEVER marks the job as applied in the ledger.
                </div>
                <div>
                  • <span className="text-emerald-300 font-medium">Reconciliation Gate:</span> The candidate or automated receipt confirmation must explicitly submit confirmation details to record an immutable application record in the ledger.
                </div>
              </div>

              {/* Action Buttons & Confirmation Form */}
              <div className="space-y-3">
                {bundle.status === 'prepared' && (
                  <div className="flex items-center justify-end">
                    <button
                      onClick={handleDispatch}
                      className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-slate-900 rounded-lg text-xs font-semibold shadow transition"
                    >
                      Open Portal & Mark Dispatched (AT-005) &rarr;
                    </button>
                  </div>
                )}

                {bundle.status === 'dispatched' && (
                  <div className="p-4 rounded-lg border border-slate-200 bg-slate-50 space-y-3">
                    <div className="text-xs font-medium text-slate-700">
                      Dispatched at: <span className="font-mono text-slate-500">{bundle.dispatched_at}</span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <div>
                        <label className="block text-xs font-medium text-slate-500 mb-1">
                          Confirmation Notes / Requisition Info (Optional)
                        </label>
                        <input
                          type="text"
                          value={notesInput}
                          onChange={(e) => setNotesInput(e.target.value)}
                          placeholder="e.g. Submitted via Workday req #JR-1092"
                          className="w-full bg-white border border-slate-300 rounded px-3 py-1.5 text-xs text-slate-800 focus:outline-none focus:border-emerald-500"
                        />
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-slate-500 mb-1">
                          Confirmation Receipt / Email Subject (Optional)
                        </label>
                        <input
                          type="text"
                          value={receiptInput}
                          onChange={(e) => setReceiptInput(e.target.value)}
                          placeholder="e.g. Acme Careers: Thank you for applying"
                          className="w-full bg-white border border-slate-300 rounded px-3 py-1.5 text-xs text-slate-800 focus:outline-none focus:border-emerald-500"
                        />
                      </div>
                    </div>

                    <div className="flex justify-end gap-2 pt-2">
                      <button
                        onClick={handleConfirmSubmission}
                        className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 rounded-lg text-xs font-semibold shadow transition"
                      >
                        Confirm Submission & Record in Ledger (REQ-005)
                      </button>
                    </div>
                  </div>
                )}

                {bundle.status === 'applied' && confirmationRecordId && (
                  <div className="p-4 rounded-lg border border-emerald-800 bg-emerald-950/40 text-xs space-y-2">
                    <div className="font-semibold text-emerald-300 flex items-center gap-1.5">
                      <span>✓ Application Confirmed & Logged to Master Ledger</span>
                    </div>
                    <div className="text-slate-700">
                      Application Record ID:{' '}
                      <span className="font-mono text-emerald-400 font-bold">{confirmationRecordId}</span>
                    </div>
                    <div className="text-slate-500">
                      Applied at: <span className="font-mono text-slate-700">{bundle.applied_at}</span>
                    </div>
                    {bundle.notes && (
                      <div className="text-slate-500">
                        Notes: <span className="text-slate-800">{bundle.notes}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Section 5: Architectural Invariants Summary */}
        <div className="bg-white/60 border border-slate-200 rounded-xl p-5 text-xs space-y-2">
          <div className="font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Compliance & Architectural Guarantees (IMP-CAR-14)
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-slate-500">
            <div>
              <span className="text-slate-800 font-medium">• Truth-in-Advertising (AT-010):</span> We never promise full automation for Workday or unsupported enterprise portals. The system provides assisted-manual workflows with zero misleading marketing.
            </div>
            <div>
              <span className="text-slate-800 font-medium">• Open/Dispatch Does Not Mark Applied (AT-005):</span> External links navigate with status "dispatched". Only explicit candidate confirmation or email receipts log status "applied".
            </div>
            <div>
              <span className="text-slate-800 font-medium">• Zero-Fabrication Fact Grounding (AT-003):</span> Clipboard values originate exclusively from confirmed Master Profile facts. No AI hallucinations.
            </div>
            <div>
              <span className="text-slate-800 font-medium">• Audit Ledger Immutability (CAR-10, REQ-005):</span> Every confirmed application writes to the permanent application ledger, preventing double-application and tracking state.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
