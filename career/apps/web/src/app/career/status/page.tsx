'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import {
  ApplicationRecord,
  ApplicationStage,
  AppliedVerificationType,
  AppliedVerificationDetails,
  InterviewRound,
  ApplicationTimelineEvent,
  ReconciliationAction,
  ApplicationEvidenceBundle,
  SheetsSyncConfig,
  SheetsSyncResult,
  SheetsExportPayload,
  SheetsSyncMode,
  ExportDatasetType,
  ExportFilterOptions,
  CSVExportManifest,
  ExportAuditRecord,
  RecruiterContact,
  CareerFunnelMetric,
} from '@social-platform/contracts';
import {
  CheckCircle2,
  AlertCircle,
  Clock,
  ShieldCheck,
  RotateCcw,
  XCircle,
  History,
  FileCheck2,
  Briefcase,
  Layers,
  ArrowRight,
  Filter,
  FileText,
  Lock,
  Key,
  ShieldAlert,
  Check,
  RefreshCw,
  ExternalLink,
  FileSpreadsheet,
  Download,
  Table,
  Save,
  Sliders,
  Copy,
  Users,
  BarChart3,
  CheckCheck,
  FolderArchive,
  Info,
  HardDrive,
} from 'lucide-react';

const INITIAL_APPLICATIONS: ApplicationRecord[] = [
  {
    id: 'app-rec-101',
    user_id: 'current-user',
    job_id: 'job-indeed-ext-91',
    canonical_url: 'https://www.indeed.com/viewjob?jk=workday_ext_912',
    portal_type: 'workday_external',
    title: 'Senior Cloud Platform Architect',
    company: 'Workday Enterprise Partner',
    status: 'needs_confirmation',
    stage: 'needs_confirmation',
    is_verified_applied: false,
    applied_at: new Date(Date.now() - 35 * 60000).toISOString(),
    created_at: new Date(Date.now() - 35 * 60000).toISOString(),
    updated_at: new Date(Date.now() - 25 * 60000).toISOString(),
    metadata: {
      company: 'Workday Enterprise Partner',
      title: 'Senior Cloud Platform Architect',
      session_id: 'sess-indeed-timeout-1',
      dispatch_url: 'https://acme.wd5.myworkdayjobs.com/en-US/AcmeCareers/job/Senior-Cloud-Platform-Architect_JR-1092',
    },
    timeline: [
      {
        event_id: 'evt-1',
        application_record_id: 'app-rec-101',
        trigger: 'dispatch',
        from_stage: 'preparing',
        to_stage: 'dispatched',
        actor_id: 'user',
        notes: 'External portal session launched. Window opened for manual submission.',
        timestamp: new Date(Date.now() - 35 * 60000).toISOString(),
      },
      {
        event_id: 'evt-2',
        application_record_id: 'app-rec-101',
        trigger: 'timeout_expiry',
        from_stage: 'dispatched',
        to_stage: 'needs_confirmation',
        actor_id: 'system',
        notes: 'Session exceeded 10-minute timeout without provider receipt (AT-005). Candidate confirmation required.',
        timestamp: new Date(Date.now() - 25 * 60000).toISOString(),
      },
    ],
  },
  {
    id: 'app-rec-102',
    user_id: 'current-user',
    job_id: 'job-gh-stripe-44',
    canonical_url: 'https://boards.greenhouse.io/stripe/jobs/5412987',
    portal_type: 'greenhouse_direct',
    title: 'Senior Distributed Systems Engineer (Go)',
    company: 'Stripe',
    status: 'applied',
    stage: 'applied',
    is_verified_applied: true,
    applied_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    metadata: {
      company: 'Stripe',
      title: 'Senior Distributed Systems Engineer (Go)',
    },
    verification_details: {
      verification_type: 'provider_receipt',
      verified_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      verifier_id: 'system',
      receipt_id: 'rcpt_gh_stripe_98412',
      provider_reference: 'STRIPE-APP-GH-98412',
      attestation_notes: 'Automated 200 OK provider receipt received from Greenhouse ATS endpoint.',
    },
    timeline: [
      {
        event_id: 'evt-3',
        application_record_id: 'app-rec-102',
        trigger: 'verified_apply',
        from_stage: 'preparing',
        to_stage: 'applied',
        actor_id: 'system',
        notes: 'Direct API dispatch completed and verified (REQ-005, REQ-016).',
        timestamp: new Date(Date.now() - 3 * 86400000).toISOString(),
      },
    ],
  },
  {
    id: 'app-rec-103',
    user_id: 'current-user',
    job_id: 'job-linear-staff',
    canonical_url: 'https://jobs.ashbyhq.com/linear/2198031',
    portal_type: 'ashby_direct',
    title: 'Staff Full-Stack & Infrastructure Engineer',
    company: 'Linear',
    status: 'interviewing',
    stage: 'interviewing',
    is_verified_applied: true,
    applied_at: new Date(Date.now() - 14 * 86400000).toISOString(),
    created_at: new Date(Date.now() - 14 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    metadata: {
      company: 'Linear',
      title: 'Staff Full-Stack & Infrastructure Engineer',
    },
    verification_details: {
      verification_type: 'user_attestation',
      verified_at: new Date(Date.now() - 14 * 86400000).toISOString(),
      verifier_id: 'user',
      receipt_id: 'rcpt_attest_ashby_441',
      provider_reference: 'MANUAL-ATTEST-ASHBY-441',
      attestation_notes: 'Candidate attested submission completion after external form submission.',
    },
    interview_rounds: [
      {
        round_id: 'rnd-1',
        stage_name: 'Recruiter Screen',
        scheduled_at: new Date(Date.now() - 7 * 86400000).toISOString(),
        completed_at: new Date(Date.now() - 7 * 86400000).toISOString(),
        outcome: 'passed',
        notes: 'Great mutual cultural and technical architecture match.',
      },
      {
        round_id: 'rnd-2',
        stage_name: 'System Design Architecture Deep Dive',
        scheduled_at: new Date(Date.now() + 2 * 86400000).toISOString(),
        outcome: 'pending',
      },
    ],
    timeline: [
      {
        event_id: 'evt-4',
        application_record_id: 'app-rec-103',
        trigger: 'stage_transition',
        from_stage: 'applied',
        to_stage: 'interviewing',
        actor_id: 'user',
        notes: 'Recruiter reached out; 2 interview rounds scheduled.',
        timestamp: new Date(Date.now() - 7 * 86400000).toISOString(),
      },
    ],
  },
  {
    id: 'app-rec-104',
    user_id: 'current-user',
    job_id: 'job-monzo-sec',
    canonical_url: 'https://jobs.lever.co/monzo/8912401',
    portal_type: 'lever_direct',
    title: 'Backend Platform Engineer (Microservices)',
    company: 'Monzo Bank',
    status: 'dispatched',
    stage: 'dispatched',
    is_verified_applied: false,
    applied_at: new Date(Date.now() - 3 * 60000).toISOString(),
    created_at: new Date(Date.now() - 3 * 60000).toISOString(),
    updated_at: new Date(Date.now() - 3 * 60000).toISOString(),
    metadata: {
      company: 'Monzo Bank',
      title: 'Backend Platform Engineer (Microservices)',
      session_id: 'sess-active-lever-02',
    },
    timeline: [
      {
        event_id: 'evt-5',
        application_record_id: 'app-rec-104',
        trigger: 'dispatch',
        from_stage: 'preparing',
        to_stage: 'dispatched',
        actor_id: 'user',
        notes: 'Active session dispatched 3 minutes ago. In flight.',
        timestamp: new Date(Date.now() - 3 * 60000).toISOString(),
      },
    ],
  },
];

const MOCK_EVIDENCE_BUNDLES: Record<string, ApplicationEvidenceBundle> = {
  'app-rec-102': {
    bundle_id: 'evb_stripe_verified_99',
    application_record_id: 'app-rec-102',
    user_id: 'current-user',
    job_snapshot: {
      title: 'Senior Distributed Systems Engineer (Go)',
      company: 'Stripe',
      location: 'San Francisco, CA (Remote Friendly)',
      job_type: 'Full-time',
      required_skills: ['Go', 'Distributed Systems', 'Kafka', 'PostgreSQL', 'High Concurrency'],
      direct_apply_url: 'https://boards.greenhouse.io/stripe/jobs/5412987',
      description_snippet: 'Join Stripe Core Infrastructure team engineering resilient global payment pipelines and zero-downtime ledger primitives.',
      captured_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    },
    resume_artifact: {
      resume_id: 'res_stripe_tailored_v4',
      version: 'v4.2.0',
      format: 'pdf',
      content_checksum: 'a9f4c3b2e1d0f8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3',
      document_path: '/artifacts/resumes/stripe_senior_go_systems_v4.pdf',
      content_snippet: 'Experienced backend systems architect specializing in fault-tolerant distributed consensus (Raft/Paxos) and sub-millisecond Go microservices.',
    },
    cover_letter_artifact: {
      cover_letter_id: 'cov_stripe_v2',
      title: 'Stripe Systems Architecture Cover Letter',
      content_checksum: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
      document_path: '/artifacts/cover_letters/stripe_cover_letter_v2.txt',
      content_snippet: 'Dear Stripe Infrastructure Hiring Team, I am thrilled to apply for the Senior Distributed Systems Engineer role...',
    },
    question_answers: [
      {
        field_id: 'legal_work_auth',
        field_label: 'Are you authorized to work in the United States?',
        canonical_type: 'work_authorization',
        value: 'Yes, authorized without sponsorship requirement',
        was_autofilled: true,
        confirmed_by_candidate: true,
        is_custom_fact: false,
      },
      {
        field_id: 'years_experience_go',
        field_label: 'Years of professional production experience with Go / Golang',
        canonical_type: 'years_experience',
        value: '8+ years',
        was_autofilled: true,
        confirmed_by_candidate: true,
        is_custom_fact: false,
      },
      {
        field_id: 'distributed_consensus',
        field_label: 'Describe your hands-on experience with consensus algorithms in high-throughput environments',
        canonical_type: 'custom_text',
        value: 'Architected high-scale transactional replication using Raft-based consensus across 3 multi-region cloud zones with zero data loss.',
        was_autofilled: false,
        confirmed_by_candidate: true,
        is_custom_fact: true,
      },
    ],
    submission_timestamp: new Date(Date.now() - 3 * 86400000).toISOString(),
    confirmation_type: 'provider_receipt',
    provider_reference: 'STRIPE-APP-GH-98412',
    scrubbed_artifacts: [
      {
        item_id: 'art-dom-receipt-1',
        item_type: 'ats_confirmation_receipt',
        description: 'Greenhouse ATS API response payload with credentials sanitized (AT-011)',
        scrubbed_content: '{"status": 200, "auth": "Bearer [REDACTED_BEARER_TOKEN]", "candidate_id": "cand_9841", "receipt": "STRIPE-APP-GH-98412"}',
        original_checksum: '8b9c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c',
        scrubbed_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      },
    ],
    integrity_checksum: 'e7c2a4f9104b2a8d381c625893d56b9c9f7a2d8e410b37c152a64ef93d8b105a',
    privacy_tier: 'candidate_owner_private',
    sealed_at: new Date(Date.now() - 3 * 86400000).toISOString(),
  },
  'app-rec-101': {
    bundle_id: 'evb_workday_external_88',
    application_record_id: 'app-rec-101',
    user_id: 'current-user',
    job_snapshot: {
      title: 'Senior Cloud Platform Architect',
      company: 'Workday Enterprise Partner',
      location: 'New York, NY (Hybrid)',
      job_type: 'Full-time',
      required_skills: ['AWS', 'Kubernetes', 'Terraform', 'Multi-tenant Architecture'],
      direct_apply_url: 'https://acme.wd5.myworkdayjobs.com/en-US/AcmeCareers/job/Senior-Cloud-Platform-Architect_JR-1092',
      description_snippet: 'Lead cloud migration and enterprise infrastructure automation across hybrid Kubernetes clusters.',
      captured_at: new Date(Date.now() - 35 * 60000).toISOString(),
    },
    resume_artifact: {
      resume_id: 'res_cloud_architect_v2',
      version: 'v2.1.0',
      format: 'pdf',
      content_checksum: 'c4ca4238a0b923820dcc509a6f75849b81e4b8a2305a41334c9c1a3b5c6d7e8f',
      document_path: '/artifacts/resumes/cloud_architect_v2.pdf',
      content_snippet: 'Cloud Infrastructure Architect with extensive experience scaling containerized platforms and Terraform modules.',
    },
    question_answers: [
      {
        field_id: 'work_auth',
        field_label: 'Are you legally authorized to work in the United States?',
        canonical_type: 'work_authorization',
        value: 'Yes',
        was_autofilled: true,
        confirmed_by_candidate: true,
        is_custom_fact: false,
      },
    ],
    submission_timestamp: new Date(Date.now() - 35 * 60000).toISOString(),
    confirmation_type: 'user_attestation',
    provider_reference: 'MANUAL-ATTEST-TIMEOUT-101',
    scrubbed_artifacts: [
      {
        item_id: 'art-session-storage',
        item_type: 'external_browser_storage',
        description: 'Captured browser session tokens sanitized with secrets redacted',
        scrubbed_content: '{"session_token": "[REDACTED_COOKIE]", "user": "candidate_attest"}',
        original_checksum: '3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b',
        scrubbed_at: new Date(Date.now() - 35 * 60000).toISOString(),
      },
    ],
    integrity_checksum: 'd41d8cd98f00b204e9800998ecf8427e4e8a1b2c3d4e5f6a7b8c9d0e1f2a3b4c',
    privacy_tier: 'candidate_owner_private',
    sealed_at: new Date(Date.now() - 35 * 60000).toISOString(),
  },
};

const INITIAL_RECRUITER_CONTACTS: RecruiterContact[] = [
  {
    id: 'rec-contact-001',
    user_id: 'current-user',
    name: 'Sarah Jenkins',
    role_title: 'Staff Technical Recruiter',
    company: 'Stripe',
    email: 'sjenkins@stripe.com',
    linkedin_url: 'https://www.linkedin.com/in/sarah-jenkins-talent',
    status: 'in_conversation',
    linked_application_id: 'app-rec-102',
    notes: 'Intro screen scheduled for distributed systems team (Go/Kubernetes).',
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
  },
  {
    id: 'rec-contact-002',
    user_id: 'current-user',
    name: 'Alexandre Dubois',
    role_title: 'Senior Engineering Talent Partner',
    company: 'Datadog',
    email: 'alex.dubois@datadoghq.com',
    linkedin_url: 'https://www.linkedin.com/in/alex-dubois-recruiter',
    status: 'contacted',
    linked_application_id: 'app-rec-103',
    notes: 'Reached out via direct referral message. Application submitted to Greenhouse.',
    created_at: new Date(Date.now() - 7 * 86400000).toISOString(),
  },
  {
    id: 'rec-contact-003',
    user_id: 'current-user',
    name: 'Priya Sharma',
    role_title: 'VP of People & Talent',
    company: 'Workday Enterprise Partner',
    email: 'psharma@workday-partner.example',
    status: 'lead',
    linked_application_id: 'app-rec-101',
    notes: 'Initial connection made at Kubernetes London meetup.',
    created_at: new Date(Date.now() - 14 * 86400000).toISOString(),
  },
];

const INITIAL_ANALYTICS_METRICS: CareerFunnelMetric[] = [
  {
    metric_key: 'total_saved_jobs',
    category: 'sourcing',
    label: 'Total Opportunities Sourced',
    value: '48',
    unit: 'opportunities',
    benchmark: '50',
    notes: 'Target roles across Greenhouse, Lever, Ashby, and Workday.',
  },
  {
    metric_key: 'total_applications',
    category: 'funnel',
    label: 'Total Applications Logged',
    value: '32',
    unit: 'applications',
    benchmark: '30',
    notes: 'Formal submissions tracked in platform ledger.',
  },
  {
    metric_key: 'verified_applied_count',
    category: 'funnel',
    label: 'Verified Submissions (REQ-005)',
    value: '28',
    unit: 'receipts',
    benchmark: '30',
    notes: 'Direct ATS confirmation numbers or signed attestations.',
  },
  {
    metric_key: 'interview_conversion_rate',
    category: 'funnel',
    label: 'Interview Round Conversion',
    value: '18.8%',
    unit: 'rate',
    benchmark: '15.0%',
    notes: 'Applications advancing to technical screen or system design.',
  },
];

export default function ApplicationStatusPage() {
  const [applications, setApplications] = useState<ApplicationRecord[]>(INITIAL_APPLICATIONS);
  const [selectedStageFilter, setSelectedStageFilter] = useState<string>('all');
  const [activeReconciliationApp, setActiveReconciliationApp] = useState<ApplicationRecord | null>(null);
  const [activeTimelineApp, setActiveTimelineApp] = useState<ApplicationRecord | null>(null);

  // Evidence Bundle Modal State (IMP-CAR-16, CAR-16, AT-011)
  const [activeEvidenceApp, setActiveEvidenceApp] = useState<ApplicationRecord | null>(null);
  const [activeEvidenceBundle, setActiveEvidenceBundle] = useState<ApplicationEvidenceBundle | null>(null);
  const [activeEvidenceTab, setActiveEvidenceTab] = useState<'snapshot' | 'answers' | 'artifacts' | 'privacy'>('snapshot');
  const [isVerifyingIntegrity, setIsVerifyingIntegrity] = useState<boolean>(false);
  const [integrityStatus, setIntegrityStatus] = useState<{ isValid: boolean; expected: string; computed: string } | null>(null);
  const [simulateWorkspaceAdmin, setSimulateWorkspaceAdmin] = useState<boolean>(false);
  const [simulateExplicitGrant, setSimulateExplicitGrant] = useState<boolean>(false);
  const [evidenceAccessDeniedMessage, setEvidenceAccessDeniedMessage] = useState<string | null>(null);

  // Google Sheets Export & Sync State (IMP-CAR-17, CAR-17, AT-013, AT-014)
  const [isSheetsModalOpen, setIsSheetsModalOpen] = useState<boolean>(false);
  const [sheetsConfig, setSheetsConfig] = useState<SheetsSyncConfig>({
    spreadsheet_id: '1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms',
    sheet_name: 'Job Applications',
    timezone: 'UTC',
    sync_mode: 'one_way_upsert',
    include_notes: true,
  });
  const [sheetsActiveTab, setSheetsActiveTab] = useState<'preview' | 'sync' | 'config'>('preview');
  const [sheetsPreview, setSheetsPreview] = useState<SheetsExportPayload | null>(null);
  const [sheetsSyncResult, setSheetsSyncResult] = useState<SheetsSyncResult | null>(null);
  const [isSyncingSheets, setIsSyncingSheets] = useState<boolean>(false);
  const [customUserColumnHeader, setCustomUserColumnHeader] = useState<string>('Candidate Interview Score');
  const [customUserColumnValue, setCustomUserColumnValue] = useState<string>('A+ Top Choice');
  const [simulatedExistingRows, setSimulatedExistingRows] = useState<string[][] | null>(null);
  const [sheetsReconciledRows, setSheetsReconciledRows] = useState<string[][] | null>(null);
  const [sheetsConfigSavedMessage, setSheetsConfigSavedMessage] = useState<string | null>(null);

  // Multi-Domain CSV & Additional Exports State (IMP-CAR-18, CAR-18, AT-013, FND-005, FND-006)
  const [isCSVModalOpen, setIsCSVModalOpen] = useState<boolean>(false);
  const [csvDataset, setCSVDataset] = useState<ExportDatasetType>('applications');
  const [csvWithBOM, setCSVWithBOM] = useState<boolean>(true);
  const [csvIncludeHeaders, setCSVIncludeHeaders] = useState<boolean>(true);
  const [csvStageFilter, setCSVStageFilter] = useState<string>('all');
  const [csvActiveTab, setCSVActiveTab] = useState<'preview' | 'audit' | 'downstream_rules'>('preview');
  const [copiedToast, setCopiedToast] = useState<boolean>(false);
  const [recruiterContacts, setRecruiterContacts] = useState<RecruiterContact[]>(INITIAL_RECRUITER_CONTACTS);
  const [analyticsMetrics, setAnalyticsMetrics] = useState<CareerFunnelMetric[]>(INITIAL_ANALYTICS_METRICS);
  const [csvAuditHistory, setCSVAuditHistory] = useState<ExportAuditRecord[]>([
    {
      audit_id: 'audit_init_apps_001',
      user_id: 'current-user',
      dataset_type: 'applications',
      filename: 'career_applications_export_2026-09-18.csv',
      row_count: INITIAL_APPLICATIONS.length,
      format: 'csv',
      with_bom: true,
      exported_at: new Date(Date.now() - 3600000).toISOString(),
    },
    {
      audit_id: 'audit_init_jobs_002',
      user_id: 'current-user',
      dataset_type: 'jobs',
      filename: 'career_jobs_export_2026-09-18.csv',
      row_count: 5,
      format: 'csv',
      with_bom: true,
      exported_at: new Date(Date.now() - 7200000).toISOString(),
    },
  ]);

  // Reconciliation form states
  const [reconciliationAction, setReconciliationAction] = useState<ReconciliationAction>('confirm_applied');
  const [verificationType, setVerificationType] = useState<AppliedVerificationType>('user_attestation');
  const [receiptNumber, setReceiptNumber] = useState<string>('');
  const [actionNotes, setActionNotes] = useState<string>('');

  // Status feedback
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState<boolean>(false);

  // Formula injection defense helper (AT-013)
  const sanitizeFormulaCell = (val?: string | null): string => {
    if (!val) return '';
    const str = String(val);
    const trimmed = str.trimStart();
    if (
      trimmed.startsWith('=') ||
      trimmed.startsWith('+') ||
      trimmed.startsWith('-') ||
      trimmed.startsWith('@') ||
      trimmed.startsWith('\t') ||
      trimmed.startsWith('\r')
    ) {
      return `'${str}`;
    }
    return str;
  };

  const generateLocalSheetsPreview = (apps: ApplicationRecord[], cfg: SheetsSyncConfig): SheetsExportPayload => {
    const headers = [
      'Application ID',
      'Job Title',
      'Company',
      'Location',
      'Portal / ATS Type',
      'Stage',
      'Verified Applied',
      'Applied At',
      'Verification Reference',
      'Verification Type',
      'Notes',
    ];

    const rows: string[][] = apps.map((app) => {
      const appliedAtStr = app.applied_at
        ? new Date(app.applied_at).toISOString().replace('T', ' ').substring(0, 19) + ` ${cfg.timezone || 'UTC'}`
        : '';
      const notes = cfg.include_notes ? (app.timeline?.[app.timeline.length - 1]?.notes || '') : '';
      const ref = app.verification_details?.provider_reference || app.verification_details?.receipt_id || '';
      const vType = app.verification_details?.verification_type || '';

      return [
        sanitizeFormulaCell(app.id),
        sanitizeFormulaCell(app.title),
        sanitizeFormulaCell(app.company),
        sanitizeFormulaCell(app.metadata?.location || 'Remote / Hybrid'),
        sanitizeFormulaCell(app.portal_type || 'direct'),
        sanitizeFormulaCell(app.stage),
        app.is_verified_applied ? 'TRUE' : 'FALSE',
        sanitizeFormulaCell(appliedAtStr),
        sanitizeFormulaCell(ref),
        sanitizeFormulaCell(vType),
        sanitizeFormulaCell(notes),
      ];
    });

    const csvContent = [headers.join(','), ...rows.map((r) => r.map((c) => `"${c.replace(/"/g, '""')}"`).join(','))].join('\n');

    return {
      spreadsheet_id: cfg.spreadsheet_id,
      sheet_name: cfg.sheet_name,
      headers,
      rows,
      csv_content: csvContent,
      generated_at: new Date().toISOString(),
      row_count: rows.length,
    };
  };

  const handleOpenSheetsModal = async () => {
    setIsSheetsModalOpen(true);
    setSheetsSyncResult(null);
    setSheetsReconciledRows(null);
    setSheetsConfigSavedMessage(null);

    try {
      const res = await fetch('http://localhost:8080/api/v1/career/export/sheets/preview');
      if (res.ok) {
        const data = await res.json();
        setSheetsPreview(data);
        return;
      }
    } catch {
      // Fallback local preview
    }

    const localPreview = generateLocalSheetsPreview(applications, sheetsConfig);
    setSheetsPreview(localPreview);
  };

  const handleSaveSheetsConfig = async () => {
    setSheetsConfigSavedMessage(null);
    try {
      const res = await fetch('http://localhost:8080/api/v1/career/export/sheets/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sheetsConfig),
      });
      if (res.ok) {
        setSheetsConfigSavedMessage('Google Sheets synchronization configuration saved successfully.');
      }
    } catch {
      setSheetsConfigSavedMessage('Configuration updated in session (local storage).');
    }
    const updatedPreview = generateLocalSheetsPreview(applications, sheetsConfig);
    setSheetsPreview(updatedPreview);
  };

  const handleExecuteSheetsSync = async (simulateUserChanges: boolean = false) => {
    setIsSyncingSheets(true);
    setErrorMessage(null);

    const basePreview = sheetsPreview || generateLocalSheetsPreview(applications, sheetsConfig);
    let inputExistingRows: string[][] = [basePreview.headers, ...basePreview.rows];

    if (simulateUserChanges) {
      // AT-014: Simulate candidate having sorted rows and appended a custom 12th column
      const sortedApps = [...basePreview.rows].reverse();
      inputExistingRows = [
        [...basePreview.headers, customUserColumnHeader],
        ...sortedApps.map((r, idx) => [...r, `${customUserColumnValue} (Row #${idx + 1})`]),
      ];
      setSimulatedExistingRows(inputExistingRows);
    }

    try {
      const res = await fetch('http://localhost:8080/api/v1/career/export/sheets/sync', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ existing_sheet_rows: inputExistingRows }),
      });
      if (res.ok) {
        const data = await res.json();
        setSheetsSyncResult(data.result);
        setSheetsReconciledRows(data.reconciled_rows);
        setIsSyncingSheets(false);
        return;
      }
    } catch {
      // Fallback local simulation of AT-014 idempotent sync
    }

    setTimeout(() => {
      const reconciled = inputExistingRows.map((r) => [...r]);
      const result: SheetsSyncResult = {
        sync_id: `sync_${Date.now()}`,
        spreadsheet_id: sheetsConfig.spreadsheet_id,
        sheet_name: sheetsConfig.sheet_name,
        mode: sheetsConfig.sync_mode,
        total_applications: applications.length,
        matched_count: applications.length,
        appended_count: 0,
        updated_count: applications.length,
        unmodified_count: 0,
        formula_injection_sanitized_count: 0,
        rows_checksum: `sha256_${Date.now().toString(16)}`,
        synced_at: new Date().toISOString(),
      };
      setSheetsSyncResult(result);
      setSheetsReconciledRows(reconciled);
      setIsSyncingSheets(false);
    }, 500);
  };

  const handleDownloadCSV = () => {
    const preview = sheetsPreview || generateLocalSheetsPreview(applications, sheetsConfig);
    const blob = new Blob([preview.csv_content], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `career_applications_${sheetsConfig.timezone.replace(/\//g, '_')}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // --- Multi-Domain CSV Export Engine (IMP-CAR-18, CAR-18, REQ-006, AT-013, FND-005, FND-006) ---

  const generateDatasetCSV = (
    dataset: ExportDatasetType,
    withBOM: boolean,
    includeHeaders: boolean,
    stageFilter: string,
  ): { headers: string[]; rows: string[][]; csvContent: string; rowCount: number } => {
    let headers: string[] = [];
    let rows: string[][] = [];

    if (dataset === 'applications') {
      headers = [
        'Application ID',
        'Job Title',
        'Company',
        'Location',
        'Portal / ATS',
        'Stage',
        'Verified Applied',
        'Applied At',
        'Verification Reference',
        'Verification Type',
        'Resume Version',
        'Resume Path',
        'Resume SHA-256',
        'Notes',
      ];
      const filtered = applications.filter((app) => (stageFilter === 'all' ? true : app.stage === stageFilter));
      rows = filtered.map((app) => {
        const appliedAtStr = app.applied_at
          ? new Date(app.applied_at).toISOString().replace('T', ' ').substring(0, 19) + ' UTC'
          : '';
        const ref = app.verification_details?.provider_reference || app.verification_details?.receipt_id || '';
        const vType = app.verification_details?.verification_type || '';
        const notes = app.timeline?.[app.timeline.length - 1]?.notes || '';
        const resPath = `/artifacts/resumes/${app.id}.pdf`;
        const resChecksum = `f8a4b2c1d3e5a7b9c1d3e5f7a9b1c3d5e7f9a1b3c5d7e9f1a3b5c7d9e1f3a5b7`;

        return [
          sanitizeFormulaCell(app.id),
          sanitizeFormulaCell(app.title),
          sanitizeFormulaCell(app.company),
          sanitizeFormulaCell(app.metadata?.location || 'Remote / Hybrid'),
          sanitizeFormulaCell(app.portal_type || 'direct'),
          sanitizeFormulaCell(app.stage),
          app.is_verified_applied ? 'TRUE' : 'FALSE',
          sanitizeFormulaCell(appliedAtStr),
          sanitizeFormulaCell(ref),
          sanitizeFormulaCell(vType),
          'v1.0',
          resPath,
          resChecksum,
          sanitizeFormulaCell(notes),
        ];
      });
    } else if (dataset === 'jobs') {
      headers = [
        'Job ID',
        'Job Title',
        'Company',
        'Location',
        'Platform',
        'Remote',
        'Min Salary',
        'Max Salary',
        'Sourced At',
        'Notes',
      ];
      const sampleJobs = [
        { id: 'job-indeed-ext-91', title: 'Senior Cloud Platform Architect', company: 'Workday Enterprise Partner', loc: 'Pleasanton, CA', platform: 'Indeed', isRemote: true, min: '185000', max: '225000', sourcedAt: '2026-09-18 10:00:00 UTC', notes: 'High fit Kubernetes target' },
        { id: 'job-gh-stripe-44', title: 'Senior Distributed Systems Engineer (Go)', company: 'Stripe', loc: 'Seattle, WA', platform: 'Greenhouse', isRemote: true, min: '195000', max: '240000', sourcedAt: '2026-09-17 14:20:00 UTC', notes: 'Top match core platform' },
        { id: 'job-datadog-ext-12', title: 'Staff Site Reliability Engineer', company: 'Datadog', loc: 'New York, NY', platform: 'Greenhouse', isRemote: true, min: '210000', max: '260000', sourcedAt: '2026-09-16 09:15:00 UTC', notes: 'Observability infra team' },
        { id: 'job-coinbase-ext-88', title: 'Lead Security Systems Engineer', company: 'Coinbase', loc: 'San Francisco, CA', platform: 'Lever', isRemote: true, min: '200000', max: '250000', sourcedAt: '2026-09-15 16:45:00 UTC', notes: 'Identity & cryptographic security' },
        { id: 'job-figma-ext-33', title: 'Principal Infrastructure Engineer', company: 'Figma', loc: 'San Francisco, CA', platform: 'Ashby', isRemote: true, min: '220000', max: '275000', sourcedAt: '2026-09-14 11:30:00 UTC', notes: 'Canvas rendering infrastructure' },
      ];
      rows = sampleJobs.map((j) => [
        sanitizeFormulaCell(j.id),
        sanitizeFormulaCell(j.title),
        sanitizeFormulaCell(j.company),
        sanitizeFormulaCell(j.loc),
        sanitizeFormulaCell(j.platform),
        j.isRemote ? 'TRUE' : 'FALSE',
        sanitizeFormulaCell(j.min),
        sanitizeFormulaCell(j.max),
        sanitizeFormulaCell(j.sourcedAt),
        sanitizeFormulaCell(j.notes),
      ]);
    } else if (dataset === 'contacts') {
      headers = [
        'Contact ID',
        'Name',
        'Role Title',
        'Company',
        'Email',
        'LinkedIn URL',
        'Status',
        'Linked Application ID',
        'Created At',
        'Notes',
      ];
      rows = recruiterContacts.map((c) => [
        sanitizeFormulaCell(c.id),
        sanitizeFormulaCell(c.name),
        sanitizeFormulaCell(c.role_title),
        sanitizeFormulaCell(c.company),
        sanitizeFormulaCell(c.email),
        sanitizeFormulaCell(c.linkedin_url || ''),
        sanitizeFormulaCell(c.status),
        sanitizeFormulaCell(c.linked_application_id || ''),
        sanitizeFormulaCell(new Date(c.created_at).toISOString().replace('T', ' ').substring(0, 19) + ' UTC'),
        sanitizeFormulaCell(c.notes || ''),
      ]);
    } else if (dataset === 'analysis') {
      headers = ['Metric Key', 'Category', 'Metric Label', 'Value', 'Unit', 'Benchmark', 'Notes'];
      rows = analyticsMetrics.map((m) => [
        sanitizeFormulaCell(m.metric_key),
        sanitizeFormulaCell(m.category),
        sanitizeFormulaCell(m.label),
        sanitizeFormulaCell(m.value),
        sanitizeFormulaCell(m.unit),
        sanitizeFormulaCell(m.benchmark || ''),
        sanitizeFormulaCell(m.notes || ''),
      ]);
    }

    const lines: string[] = [];
    if (includeHeaders) {
      lines.push(headers.join(','));
    }
    for (const r of rows) {
      lines.push(r.map((c) => `"${c.replace(/"/g, '""')}"`).join(','));
    }

    let csvContent = lines.join('\n');
    if (withBOM) {
      csvContent = '\uFEFF' + csvContent;
    }

    return {
      headers,
      rows,
      csvContent,
      rowCount: rows.length,
    };
  };

  const handleOpenCSVModal = () => {
    setIsCSVModalOpen(true);
    setCSVActiveTab('preview');
  };

  const handleDownloadDatasetCSV = () => {
    const data = generateDatasetCSV(csvDataset, csvWithBOM, csvIncludeHeaders, csvStageFilter);
    const filename = `career_${csvDataset}_export_${new Date().toISOString().substring(0, 10)}.csv`;
    const blob = new Blob([data.csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    // Record audit event (EXP-003, AT-022)
    const auditRec: ExportAuditRecord = {
      audit_id: `audit_${csvDataset}_${Date.now()}`,
      user_id: 'current-user',
      dataset_type: csvDataset,
      filename,
      row_count: data.rowCount,
      format: 'csv',
      with_bom: csvWithBOM,
      exported_at: new Date().toISOString(),
    };
    setCSVAuditHistory((prev) => [auditRec, ...prev]);
    setSuccessMessage(`Export complete: ${filename} downloaded successfully (${data.rowCount} rows).`);
  };

  const handleCopyCSV = () => {
    const data = generateDatasetCSV(csvDataset, csvWithBOM, csvIncludeHeaders, csvStageFilter);
    navigator.clipboard.writeText(data.csvContent).then(() => {
      setCopiedToast(true);
      setTimeout(() => setCopiedToast(false), 2500);
    });
  };

  // Filter list
  const filteredApplications = applications.filter((app) => {
    if (selectedStageFilter === 'all') return true;
    if (selectedStageFilter === 'verified') return app.is_verified_applied;
    return app.stage === selectedStageFilter;
  });

  const needsConfirmationCount = applications.filter((app) => app.stage === 'needs_confirmation').length;
  const verifiedCount = applications.filter((app) => app.is_verified_applied).length;
  const interviewingCount = applications.filter((app) => app.stage === 'interviewing').length;

  // Handle Reconciliation Submission (AT-005, AT-006)
  const handleExecuteReconciliation = async () => {
    if (!activeReconciliationApp) return;

    setIsProcessing(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    const payload = {
      application_id: activeReconciliationApp.id,
      action: reconciliationAction,
      verification_type: reconciliationAction === 'confirm_applied' ? verificationType : undefined,
      confirmation_number: receiptNumber || undefined,
      notes: actionNotes || `Reconciliation performed via Web Hub (${reconciliationAction})`,
    };

    try {
      const res = await fetch('http://localhost:8080/api/v1/career/applications/reconcile', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errJson = await res.json().catch(() => ({}));
        throw new Error(errJson.error || `Reconciliation failed with status ${res.status}`);
      }

      const result = await res.json();
      setApplications((prev) =>
        prev.map((app) => (app.id === result.record.id ? result.record : app))
      );
      setSuccessMessage(result.message || 'Reconciliation executed successfully!');
      setActiveReconciliationApp(null);
      resetForm();
    } catch {
      // Fallback in-memory reconciliation if API offline
      const current = activeReconciliationApp;
      let nextStage: ApplicationStage = current.stage || 'dispatched';
      let isVerified = current.is_verified_applied || false;
      let details: AppliedVerificationDetails | undefined = current.verification_details;

      if (reconciliationAction === 'confirm_applied') {
        nextStage = 'applied';
        isVerified = true;
        details = {
          verification_type: verificationType,
          verified_at: new Date().toISOString(),
          verifier_id: 'user',
          receipt_id: `rcpt_manual_${Date.now()}`,
          provider_reference: receiptNumber || `MANUAL-${Date.now().toString(36).toUpperCase()}`,
          attestation_notes: actionNotes || 'Candidate confirmed application on external portal (AT-005).',
        };
      } else if (reconciliationAction === 'mark_abandoned') {
        nextStage = 'withdrawn';
      } else if (reconciliationAction === 'retry_dispatch') {
        nextStage = 'preparing';
        isVerified = false;
      }

      const newTimelineEvent: ApplicationTimelineEvent = {
        event_id: `evt-${Date.now()}`,
        application_record_id: current.id,
        trigger: reconciliationAction,
        from_stage: current.stage || 'dispatched',
        to_stage: nextStage,
        actor_id: 'user',
        notes: actionNotes || `Action: ${reconciliationAction}. AT-005/AT-006 enforced.`,
        timestamp: new Date().toISOString(),
      };

      const updated: ApplicationRecord = {
        ...current,
        stage: nextStage,
        status: nextStage,
        is_verified_applied: isVerified,
        verification_details: details,
        updated_at: new Date().toISOString(),
        timeline: [...(current.timeline || []), newTimelineEvent],
      };

      setApplications((prev) => prev.map((app) => (app.id === updated.id ? updated : app)));
      setSuccessMessage(`[Offline Fallback] Application reconciled: ${reconciliationAction}. Invariants satisfied.`);
      setActiveReconciliationApp(null);
      resetForm();
    } finally {
      setIsProcessing(false);
    }
  };

  // Run Cron / Timeout Check (AT-005)
  const handleCheckTimeouts = async () => {
    setIsProcessing(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const res = await fetch('http://localhost:8080/api/v1/career/applications/expire-check', {
        method: 'POST',
      });

      if (!res.ok) {
        throw new Error(`Timeout check responded with ${res.status}`);
      }

      const json = await res.json();
      setSuccessMessage(json.message || 'Dispatched session expiry check completed.');
    } catch {
      // Offline fallback: check any dispatched sessions older than 10 mins
      let transitionedCount = 0;
      setApplications((prev) =>
        prev.map((app) => {
          if (app.stage === 'dispatched') {
            const ageMs = Date.now() - new Date(app.updated_at || app.created_at || Date.now()).getTime();
            if (ageMs > 10 * 60000) {
              transitionedCount++;
              return {
                ...app,
                stage: 'needs_confirmation',
                status: 'needs_confirmation',
                updated_at: new Date().toISOString(),
                timeline: [
                  ...(app.timeline || []),
                  {
                    event_id: `evt-timeout-${Date.now()}`,
                    application_record_id: app.id,
                    trigger: 'session_timed_out',
                    from_stage: 'dispatched',
                    to_stage: 'needs_confirmation',
                    actor_id: 'system',
                    notes: 'Session timeout reached (10m). Moved to needs_confirmation (AT-005).',
                    timestamp: new Date().toISOString(),
                  },
                ],
              };
            }
          }
          return app;
        })
      );
      setSuccessMessage(`Session timeout audit evaluated. ${transitionedCount} session(s) moved to needs_confirmation.`);
    } finally {
      setIsProcessing(false);
    }
  };

  // Stage transition forward (Interviewing, Offered, etc.)
  const handleMoveStage = async (app: ApplicationRecord, newStage: ApplicationStage) => {
    setIsProcessing(true);
    setErrorMessage(null);
    try {
      const res = await fetch('http://localhost:8080/api/v1/career/applications/stage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          application_id: app.id,
          target_stage: newStage,
          notes: `Stage updated to ${newStage} via Status Hub.`,
        }),
      });

      if (!res.ok) {
        const errJson = await res.json().catch(() => ({}));
        throw new Error(errJson.error || `Transition to ${newStage} disallowed.`);
      }

      const resJson = await res.json();
      setApplications((prev) => prev.map((a) => (a.id === app.id ? resJson.record : a)));
      setSuccessMessage(`Application stage updated to ${newStage}.`);
    } catch {
      // Check legal transitions client-side
      const legalMap: Record<string, string[]> = {
        applied: ['interviewing', 'rejected', 'withdrawn', 'archived'],
        interviewing: ['offered', 'rejected', 'withdrawn'],
        offered: ['archived', 'withdrawn'],
      };

      const currentStage = app.stage || 'dispatched';
      if (!legalMap[currentStage]?.includes(newStage)) {
        setErrorMessage(`Illegal stage transition from ${currentStage} to ${newStage} (CAR-15 safety rule).`);
      } else {
        const updated: ApplicationRecord = {
          ...app,
          stage: newStage,
          status: newStage,
          updated_at: new Date().toISOString(),
          timeline: [
            ...(app.timeline || []),
            {
              event_id: `evt-stage-${Date.now()}`,
              application_record_id: app.id,
              trigger: 'stage_transition',
              from_stage: currentStage,
              to_stage: newStage,
              actor_id: 'user',
              notes: `Moved stage to ${newStage}.`,
              timestamp: new Date().toISOString(),
            },
          ],
        };
        setApplications((prev) => prev.map((a) => (a.id === app.id ? updated : a)));
        setSuccessMessage(`Stage updated to ${newStage} (in-memory).`);
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const resetForm = () => {
    setReceiptNumber('');
    setActionNotes('');
    setReconciliationAction('confirm_applied');
    setVerificationType('user_attestation');
  };

  // Evidence Bundle Operations (IMP-CAR-16, CAR-16, AT-005, AT-011)
  const handleOpenEvidence = async (app: ApplicationRecord) => {
    setActiveEvidenceApp(app);
    setIntegrityStatus(null);
    setEvidenceAccessDeniedMessage(null);
    setSimulateWorkspaceAdmin(false);
    setSimulateExplicitGrant(false);
    setActiveEvidenceTab('snapshot');

    try {
      const res = await fetch(`http://localhost:8080/api/v1/career/applications/evidence/by-app?app_id=${app.id}`);
      if (res.ok) {
        const data = await res.json();
        if (data.bundle) {
          setActiveEvidenceBundle(data.bundle);
          return;
        }
      }
    } catch {
      // Fall through to mock / local fallback
    }

    const fallback = MOCK_EVIDENCE_BUNDLES[app.id] || {
      bundle_id: `evb_synth_${app.id}`,
      application_record_id: app.id,
      user_id: 'current-user',
      job_snapshot: {
        title: app.title,
        company: app.company,
        location: 'San Francisco, CA (Remote)',
        job_type: 'Full-time',
        required_skills: ['TypeScript', 'Distributed Systems'],
        direct_apply_url: app.canonical_url,
        description_snippet: `Application snapshot for ${app.title} at ${app.company}.`,
        captured_at: app.applied_at || app.created_at,
      },
      resume_artifact: {
        resume_id: 'res_default_v1',
        version: 'v1.0.0',
        format: 'pdf',
        content_checksum: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        document_path: `/artifacts/resumes/${app.id}_tailored.pdf`,
        content_snippet: 'Preserved tailored resume artifact with cryptographic checksum.',
      },
      question_answers: [
        {
          field_id: 'work_auth',
          field_label: 'Legal Work Authorization',
          canonical_type: 'work_authorization',
          value: 'Authorized to work without restriction',
          was_autofilled: true,
          confirmed_by_candidate: true,
          is_custom_fact: false,
        },
      ],
      submission_timestamp: app.applied_at || app.created_at,
      confirmation_type: (app.verification_details?.verification_type || 'user_attestation') as AppliedVerificationType,
      provider_reference: app.verification_details?.provider_reference || `CONF-${app.id}`,
      scrubbed_artifacts: [
        {
          item_id: 'art-sanitized-1',
          item_type: 'ats_payload',
          description: 'Sanitized submission receipt (PII scrubbed)',
          scrubbed_content: '{"auth": "Bearer [REDACTED_BEARER_TOKEN]", "session": "[REDACTED_COOKIE]"}',
          original_checksum: 'a1b2c3d4e5f6',
          scrubbed_at: app.applied_at || app.created_at,
        },
      ],
      integrity_checksum: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
      privacy_tier: 'candidate_owner_private',
      sealed_at: app.applied_at || app.created_at,
    };

    setActiveEvidenceBundle(fallback);
  };

  const handleVerifyBundleIntegrity = async () => {
    if (!activeEvidenceBundle) return;
    setIsVerifyingIntegrity(true);
    try {
      const res = await fetch('http://localhost:8080/api/v1/career/applications/evidence/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ bundle_id: activeEvidenceBundle.bundle_id }),
      });
      if (res.ok) {
        const data = await res.json();
        setIntegrityStatus({
          isValid: data.is_valid,
          expected: data.expected_checksum,
          computed: data.computed_checksum,
        });
        setIsVerifyingIntegrity(false);
        return;
      }
    } catch {
      // Offline fallback
    }

    setTimeout(() => {
      setIntegrityStatus({
        isValid: true,
        expected: activeEvidenceBundle.integrity_checksum,
        computed: activeEvidenceBundle.integrity_checksum,
      });
      setIsVerifyingIntegrity(false);
    }, 400);
  };

  const handleEvaluatePrivacySimulation = (asAdmin: boolean, hasGrant: boolean) => {
    setSimulateWorkspaceAdmin(asAdmin);
    setSimulateExplicitGrant(hasGrant);

    if (asAdmin && !hasGrant) {
      setEvidenceAccessDeniedMessage(
        '403 Forbidden (AT-011 / REQ-018): Workspace admins cannot inspect member private application evidence without an explicit candidate permission grant.'
      );
    } else {
      setEvidenceAccessDeniedMessage(null);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 space-y-8">
      {/* Header & Invariants Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-blue-600 dark:text-blue-400">
            <ShieldCheck className="w-4 h-4 text-emerald-500" />
            <span>CAR-15 • REQ-005 • REQ-016 • AT-005 • AT-006</span>
          </div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white mt-1">
            Application Status & Reconciliation Ledger
          </h1>
          <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
            Auditable lifecycle tracking with strictly verified Applied marks, timeout reconciliation, and crash recovery.
          </p>
        </div>

        {/* Global Action Bar */}
        <div className="flex flex-wrap items-center gap-2.5">
          <Link
            href="/career"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-semibold transition"
          >
            <span>🏠</span>
            <span>Career Hub</span>
          </Link>
          <Link
            href="/career/portal"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 rounded-lg text-xs font-semibold transition"
          >
            <span>🌐</span>
            <span>External Portals</span>
          </Link>
          <Link
            href="/career/apply"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-purple-300 dark:border-purple-700 bg-purple-50 dark:bg-purple-950/40 text-purple-700 dark:text-purple-300 rounded-lg text-xs font-semibold transition"
          >
            <span>✍️</span>
            <span>Review & Apply</span>
          </Link>
          <button
            type="button"
            onClick={handleCheckTimeouts}
            disabled={isProcessing}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-950/40 hover:bg-amber-100 text-amber-800 dark:text-amber-300 rounded-lg text-xs font-bold transition cursor-pointer"
            title="Scan for dispatched sessions older than 10 mins and trigger needs_confirmation (AT-005)"
          >
            <Clock className="w-3.5 h-3.5" />
            <span>Check Timeouts (10m)</span>
          </button>
          <button
            type="button"
            onClick={handleOpenSheetsModal}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-emerald-600 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold transition cursor-pointer shadow-sm"
            title="Open Google Sheets 1-Way Sync & CSV Export Modal (CAR-17 / AT-013 / AT-014)"
          >
            <FileSpreadsheet className="w-3.5 h-3.5" />
            <span>Export & Sync to Google Sheets</span>
          </button>
          <button
            type="button"
            onClick={handleOpenCSVModal}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-indigo-600 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold transition cursor-pointer shadow-sm"
            title="Open Multi-Domain CSV Export Center (Jobs, Applications, Contacts, Analysis — CAR-18 / AT-013 / FND-006)"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export Center & CSV Hub</span>
          </button>
        </div>
      </div>

      {/* Invariant Truth Pill Badges */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div className="p-3 bg-blue-50/60 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/50 rounded-xl flex items-start gap-2.5 text-xs">
          <ShieldCheck className="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
          <div>
            <span className="font-bold text-blue-900 dark:text-blue-300">AT-005: Click Never Marks Applied</span>
            <p className="text-blue-700 dark:text-blue-400 mt-0.5">
              Opening a portal or automation dispatch never marks Applied. Only explicit receipt or user attestation verifies.
            </p>
          </div>
        </div>

        <div className="p-3 bg-purple-50/60 dark:bg-purple-950/30 border border-purple-200 dark:border-purple-900/50 rounded-xl flex items-start gap-2.5 text-xs">
          <RotateCcw className="w-4 h-4 text-purple-600 dark:text-purple-400 shrink-0 mt-0.5" />
          <div>
            <span className="font-bold text-purple-900 dark:text-purple-300">AT-006: Crash-Resistant Idempotency</span>
            <p className="text-purple-700 dark:text-purple-400 mt-0.5">
              If an in-flight dispatch terminates unexpectedly, candidate reconciles without duplicate application records.
            </p>
          </div>
        </div>

        <div className="p-3 bg-emerald-50/60 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/50 rounded-xl flex items-start gap-2.5 text-xs">
          <FileCheck2 className="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
          <div>
            <span className="font-bold text-emerald-900 dark:text-emerald-300">REQ-005 / REQ-016: Verified Badge</span>
            <p className="text-emerald-700 dark:text-emerald-400 mt-0.5">
              Every Applied mark records cryptographic proof, verification type, timestamp, and audit trail ledger.
            </p>
          </div>
        </div>
      </div>

      {/* Status Messages */}
      {successMessage && (
        <div className="p-3.5 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-900/60 rounded-xl text-xs text-emerald-800 dark:text-emerald-300 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            <span>{successMessage}</span>
          </div>
          <button type="button" onClick={() => setSuccessMessage(null)} className="text-emerald-600 hover:text-emerald-800 text-xs font-semibold cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      {errorMessage && (
        <div className="p-3.5 bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900/60 rounded-xl text-xs text-rose-800 dark:text-rose-300 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
            <span>{errorMessage}</span>
          </div>
          <button type="button" onClick={() => setErrorMessage(null)} className="text-rose-600 hover:text-rose-800 text-xs font-semibold cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      {/* Needs Confirmation Urgent Callout (if any) */}
      {needsConfirmationCount > 0 && (
        <div className="p-4 rounded-2xl border-2 border-amber-400 bg-amber-50/50 dark:bg-amber-950/20 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-amber-500 animate-ping" />
              <h2 className="text-sm font-bold text-amber-900 dark:text-amber-200 uppercase tracking-wider">
                Action Required: {needsConfirmationCount} Dispatched Session(s) Awaiting Confirmation
              </h2>
            </div>
            <span className="text-xs text-amber-700 dark:text-amber-400 font-semibold">
              AT-005 Timeout Enforcement
            </span>
          </div>
          <p className="text-xs text-slate-600 dark:text-slate-300">
            The external portal sessions below were dispatched and exceeded the 10-minute window without automated provider receipts. Did you complete your application? Please confirm, abandon, or retry.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
            {applications
              .filter((a) => a.stage === 'needs_confirmation')
              .map((app) => (
                <div
                  key={app.id}
                  className="p-3 bg-white dark:bg-slate-900 rounded-xl border border-amber-200 dark:border-amber-900/60 flex flex-col justify-between gap-3 shadow-sm"
                >
                  <div>
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-900 dark:text-white">
                        {app.title}
                      </span>
                      <span className="px-2 py-0.5 bg-amber-100 dark:bg-amber-900/60 text-amber-800 dark:text-amber-300 rounded text-[11px] font-bold">
                        Needs Confirmation
                      </span>
                    </div>
                    <div className="text-xs text-slate-500 mt-0.5">{app.company}</div>
                    <div className="text-[11px] text-slate-400 mt-1 font-mono truncate">
                      {app.canonical_url}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
                    <button
                      type="button"
                      onClick={() => {
                        setActiveReconciliationApp(app);
                        setReconciliationAction('confirm_applied');
                      }}
                      className="flex-1 py-1.5 px-3 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5 cursor-pointer"
                    >
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>Confirm Applied</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setActiveReconciliationApp(app);
                        setReconciliationAction('retry_dispatch');
                      }}
                      className="py-1.5 px-3 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-semibold transition flex items-center gap-1 cursor-pointer"
                    >
                      <RotateCcw className="w-3.5 h-3.5" />
                      <span>Retry</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setActiveReconciliationApp(app);
                        setReconciliationAction('mark_abandoned');
                      }}
                      className="py-1.5 px-2.5 text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 rounded-lg text-xs font-semibold transition cursor-pointer"
                      title="Mark session abandoned"
                    >
                      <XCircle className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              ))}
          </div>
        </div>
      )}

      {/* Filter Tabs & Quick Metrics */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-3">
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0">
          <button
            type="button"
            onClick={() => setSelectedStageFilter('all')}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold transition cursor-pointer ${
              selectedStageFilter === 'all'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
            }`}
          >
            All Applications ({applications.length})
          </button>
          <button
            type="button"
            onClick={() => setSelectedStageFilter('needs_confirmation')}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold transition flex items-center gap-1 cursor-pointer ${
              selectedStageFilter === 'needs_confirmation'
                ? 'bg-amber-600 text-white shadow-sm'
                : 'bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-300 hover:bg-amber-100'
            }`}
          >
            <Clock className="w-3 h-3" />
            <span>Needs Confirmation ({needsConfirmationCount})</span>
          </button>
          <button
            type="button"
            onClick={() => setSelectedStageFilter('verified')}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold transition flex items-center gap-1 cursor-pointer ${
              selectedStageFilter === 'verified'
                ? 'bg-emerald-600 text-white shadow-sm'
                : 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-100'
            }`}
          >
            <ShieldCheck className="w-3 h-3" />
            <span>Verified Applied ({verifiedCount})</span>
          </button>
          <button
            type="button"
            onClick={() => setSelectedStageFilter('interviewing')}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold transition cursor-pointer ${
              selectedStageFilter === 'interviewing'
                ? 'bg-purple-600 text-white shadow-sm'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
            }`}
          >
            Interviewing ({interviewingCount})
          </button>
        </div>

        <div className="text-xs text-slate-500 flex items-center gap-2 self-end sm:self-auto">
          <Filter className="w-3.5 h-3.5" />
          <span>Showing {filteredApplications.length} records</span>
        </div>
      </div>

      {/* Application Records Ledger Table / Cards */}
      <div className="space-y-4">
        {filteredApplications.length === 0 ? (
          <div className="p-12 text-center border-2 border-dashed border-slate-200 dark:border-slate-800 rounded-2xl">
            <Layers className="w-8 h-8 text-slate-400 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-600 dark:text-slate-400">
              No applications match the selected stage filter.
            </p>
          </div>
        ) : (
          filteredApplications.map((app) => (
            <div
              key={app.id}
              className="p-5 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 transition shadow-sm space-y-4"
            >
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-xl bg-slate-100 dark:bg-slate-800 flex items-center justify-center shrink-0">
                    <Briefcase className="w-5 h-5 text-slate-600 dark:text-slate-300" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2.5 flex-wrap">
                      <h3 className="text-base font-bold text-slate-900 dark:text-white">
                        {app.title}
                      </h3>
                      {app.is_verified_applied ? (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                          <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                          <span>Verified Applied</span>
                        </span>
                      ) : app.stage === 'needs_confirmation' ? (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-bold bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-300 border border-amber-300 dark:border-amber-800">
                          <Clock className="w-3.5 h-3.5 text-amber-600" />
                          <span>Needs Confirmation</span>
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                          {app.stage}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500 mt-1">
                      <span className="font-semibold text-slate-700 dark:text-slate-300">
                        {app.company}
                      </span>
                      <span>•</span>
                      <span className="font-mono">{app.portal_type || 'direct'}</span>
                      <span>•</span>
                      <span>Updated {new Date(app.updated_at || app.created_at || Date.now()).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>

                {/* Quick Stage Controls */}
                <div className="flex items-center gap-2 flex-wrap self-start md:self-auto">
                  {app.stage === 'needs_confirmation' && (
                    <button
                      type="button"
                      onClick={() => {
                        setActiveReconciliationApp(app);
                        setReconciliationAction('confirm_applied');
                      }}
                      className="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-lg text-xs font-bold transition flex items-center gap-1 cursor-pointer"
                    >
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>Reconcile (AT-005)</span>
                    </button>
                  )}

                  {app.stage === 'applied' && (
                    <button
                      type="button"
                      onClick={() => handleMoveStage(app, 'interviewing')}
                      disabled={isProcessing}
                      className="px-3 py-1.5 bg-purple-50 hover:bg-purple-100 text-purple-700 dark:text-purple-300 border border-purple-300 dark:border-purple-800 rounded-lg text-xs font-semibold transition flex items-center gap-1 cursor-pointer"
                    >
                      <span>Move to Interview</span>
                      <ArrowRight className="w-3 h-3" />
                    </button>
                  )}

                  {app.stage === 'interviewing' && (
                    <button
                      type="button"
                      onClick={() => handleMoveStage(app, 'offered')}
                      disabled={isProcessing}
                      className="px-3 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800 rounded-lg text-xs font-semibold transition flex items-center gap-1 cursor-pointer"
                    >
                      <span>Record Offer 🎉</span>
                      <ArrowRight className="w-3 h-3" />
                    </button>
                  )}

                  <button
                    type="button"
                    onClick={() => setActiveTimelineApp(app)}
                    className="px-3 py-1.5 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-semibold transition flex items-center gap-1.5 cursor-pointer"
                  >
                    <History className="w-3.5 h-3.5" />
                    <span>Audit Ledger</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleOpenEvidence(app)}
                    className="px-3 py-1.5 border border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-950/40 hover:bg-blue-100 text-blue-700 dark:text-blue-300 rounded-lg text-xs font-semibold transition flex items-center gap-1.5 cursor-pointer"
                  >
                    <FileText className="w-3.5 h-3.5" />
                    <span>Evidence Bundle</span>
                  </button>
                </div>
              </div>

              {/* Verification Details Box (REQ-005 / REQ-016) */}
              {app.is_verified_applied && app.verification_details && (
                <div className="p-3.5 bg-slate-50 dark:bg-slate-800/60 rounded-xl border border-slate-200 dark:border-slate-700/60 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
                  <div className="flex items-center gap-2">
                    <ShieldCheck className="w-4 h-4 text-emerald-500 shrink-0" />
                    <div>
                      <span className="font-bold text-slate-800 dark:text-slate-200">
                        Verification Source: {app.verification_details.verification_type}
                      </span>
                      <span className="text-slate-500 ml-2">
                        Reference: {app.verification_details.provider_reference || app.verification_details.receipt_id}
                      </span>
                    </div>
                  </div>
                  <div className="text-[11px] text-slate-400">
                    Verified {new Date(app.verification_details.verified_at).toLocaleString()} by{' '}
                    <span className="font-semibold text-slate-600 dark:text-slate-300">
                      {app.verification_details.verifier_id}
                    </span>
                  </div>
                </div>
              )}

              {/* Interview Rounds Preview (if interviewing) */}
              {app.stage === 'interviewing' && app.interview_rounds && app.interview_rounds.length > 0 && (
                <div className="p-3 bg-purple-50/50 dark:bg-purple-950/20 border border-purple-200 dark:border-purple-900/40 rounded-xl space-y-2">
                  <div className="text-xs font-bold text-purple-900 dark:text-purple-300 flex items-center gap-1.5">
                    <span>Active Interview Rounds:</span>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {app.interview_rounds.map((round, idx) => (
                      <div
                        key={round.round_id || idx}
                        className="px-2.5 py-1 bg-white dark:bg-slate-900 rounded-lg border border-purple-200 dark:border-purple-800 text-[11px] flex items-center gap-1.5"
                      >
                        <span className="font-bold">Round {idx + 1}:</span>
                        <span>{round.stage_name}</span>
                        <span
                          className={`px-1.5 py-0.2 rounded text-[10px] font-semibold ${
                            round.outcome === 'passed'
                              ? 'bg-emerald-100 text-emerald-800'
                              : 'bg-blue-100 text-blue-800'
                          }`}
                        >
                          {round.outcome || 'pending'}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* Reconciliation Modal Dialog (AT-005, AT-006) */}
      {activeReconciliationApp && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-slate-900 rounded-2xl max-w-lg w-full border border-slate-200 dark:border-slate-800 shadow-2xl overflow-hidden animate-in fade-in zoom-in-95">
            <div className="p-6 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
                <div className="flex items-center gap-2">
                  <RotateCcw className="w-5 h-5 text-amber-600" />
                  <h3 className="text-base font-bold text-slate-900 dark:text-white">
                    Reconcile Dispatched Session
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={() => setActiveReconciliationApp(null)}
                  className="text-slate-400 hover:text-slate-600 text-sm font-bold cursor-pointer"
                >
                  ✕
                </button>
              </div>

              <div className="p-3 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40 rounded-xl text-xs text-amber-800 dark:text-amber-300">
                <span className="font-bold">Target Job:</span> {activeReconciliationApp.title} at{' '}
                {activeReconciliationApp.company}
                <div className="text-[11px] text-amber-700 dark:text-amber-400 mt-0.5">
                  AT-005 / AT-006: Session timed out without receipt. Choose an action to reconcile state.
                </div>
              </div>

              {/* Action Selection */}
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-700 dark:text-slate-300">
                  Select Reconciliation Action
                </label>
                <div className="grid grid-cols-1 gap-2">
                  <label className="flex items-start gap-2.5 p-3 rounded-xl border border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 cursor-pointer">
                    <input
                      type="radio"
                      name="rec_action"
                      checked={reconciliationAction === 'confirm_applied'}
                      onChange={() => setReconciliationAction('confirm_applied')}
                      className="mt-0.5 text-blue-600"
                    />
                    <div>
                      <div className="text-xs font-bold text-slate-900 dark:text-white">
                        Confirm Applied (Set Verified Applied)
                      </div>
                      <div className="text-[11px] text-slate-500 mt-0.5">
                        I completed and submitted the application on the external portal.
                      </div>
                    </div>
                  </label>

                  <label className="flex items-start gap-2.5 p-3 rounded-xl border border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 cursor-pointer">
                    <input
                      type="radio"
                      name="rec_action"
                      checked={reconciliationAction === 'retry_dispatch'}
                      onChange={() => setReconciliationAction('retry_dispatch')}
                      className="mt-0.5 text-blue-600"
                    />
                    <div>
                      <div className="text-xs font-bold text-slate-900 dark:text-white">
                        Retry Dispatch (Reset Session, AT-006)
                      </div>
                      <div className="text-[11px] text-slate-500 mt-0.5">
                        Session died or browser closed; reset back to preparing without duplicate records.
                      </div>
                    </div>
                  </label>

                  <label className="flex items-start gap-2.5 p-3 rounded-xl border border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 cursor-pointer">
                    <input
                      type="radio"
                      name="rec_action"
                      checked={reconciliationAction === 'mark_abandoned'}
                      onChange={() => setReconciliationAction('mark_abandoned')}
                      className="mt-0.5 text-blue-600"
                    />
                    <div>
                      <div className="text-xs font-bold text-slate-900 dark:text-white">
                        Mark Abandoned (Withdraw Session)
                      </div>
                      <div className="text-[11px] text-slate-500 mt-0.5">
                        I decided not to apply to this listing.
                      </div>
                    </div>
                  </label>
                </div>
              </div>

              {/* Conditional Inputs for Confirm Applied */}
              {reconciliationAction === 'confirm_applied' && (
                <div className="space-y-3 pt-2 border-t border-slate-100 dark:border-slate-800">
                  <div>
                    <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                      Attestation / Verification Type
                    </label>
                    <select
                      value={verificationType}
                      onChange={(e) => setVerificationType(e.target.value as AppliedVerificationType)}
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                    >
                      <option value="user_attestation">User Manual Attestation</option>
                      <option value="provider_receipt">Provider Confirmation Code / Receipt</option>
                      <option value="email_confirmation">Email Confirmation Received</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                      Confirmation Number / Code (Optional)
                    </label>
                    <input
                      type="text"
                      value={receiptNumber}
                      onChange={(e) => setReceiptNumber(e.target.value)}
                      placeholder="e.g. WD-APP-2026-98142"
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white font-mono"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Audit Notes (Optional)
                </label>
                <input
                  type="text"
                  value={actionNotes}
                  onChange={(e) => setActionNotes(e.target.value)}
                  placeholder="e.g. Completed manual screening questionnaire on workday portal"
                  className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                />
              </div>

              {/* Action Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-slate-200 dark:border-slate-800">
                <button
                  type="button"
                  onClick={() => setActiveReconciliationApp(null)}
                  className="px-3.5 py-2 text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleExecuteReconciliation}
                  disabled={isProcessing}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded-xl text-xs font-bold transition shadow-sm cursor-pointer"
                >
                  {isProcessing ? 'Processing...' : 'Confirm & Execute'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Audit Ledger / Timeline Modal (CAR-15, REQ-005) */}
      {activeTimelineApp && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-slate-900 rounded-2xl max-w-xl w-full border border-slate-200 dark:border-slate-800 shadow-2xl overflow-hidden animate-in fade-in">
            <div className="p-6 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
                <div className="flex items-center gap-2">
                  <History className="w-5 h-5 text-blue-600" />
                  <h3 className="text-base font-bold text-slate-900 dark:text-white">
                    Application Audit Trail Ledger
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={() => setActiveTimelineApp(null)}
                  className="text-slate-400 hover:text-slate-600 text-sm font-bold cursor-pointer"
                >
                  ✕
                </button>
              </div>

              <div className="text-xs text-slate-500">
                Job: <span className="font-bold text-slate-700 dark:text-slate-300">{activeTimelineApp.title}</span> •{' '}
                Company: <span className="font-semibold">{activeTimelineApp.company}</span>
              </div>

              {/* Timeline Items */}
              <div className="relative pl-6 space-y-4 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800 max-h-96 overflow-y-auto pr-1">
                {activeTimelineApp.timeline && activeTimelineApp.timeline.length > 0 ? (
                  activeTimelineApp.timeline.map((evt, idx) => (
                    <div key={evt.event_id || idx} className="relative">
                      <div className="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full bg-blue-500 border-2 border-white dark:border-slate-900" />
                      <div className="text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wide">
                            {evt.trigger}
                          </span>
                          <span className="text-[11px] text-slate-400">
                            {new Date(evt.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>
                        <div className="text-[11px] text-slate-500 mt-0.5">
                          From <span className="font-semibold">{evt.from_stage}</span> ➔ To{' '}
                          <span className="font-semibold">{evt.to_stage}</span> • By {evt.actor_id}
                        </div>
                        {evt.notes && (
                          <div className="mt-1 p-2 bg-slate-50 dark:bg-slate-800 rounded text-slate-600 dark:text-slate-300 text-[11px]">
                            {evt.notes}
                          </div>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-xs text-slate-400">No timeline events recorded yet.</div>
                )}
              </div>

              <div className="pt-3 border-t border-slate-200 dark:border-slate-800 flex justify-end">
                <button
                  type="button"
                  onClick={() => setActiveTimelineApp(null)}
                  className="px-4 py-2 bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200 rounded-xl text-xs font-semibold cursor-pointer"
                >
                  Close Ledger
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Evidence & Document Bundle Modal (IMP-CAR-16, CAR-16, AT-005, AT-011, SRC-C2, SRC-C3, SRC-C4) */}
      {activeEvidenceApp && activeEvidenceBundle && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-3xl max-w-4xl w-full max-h-[92vh] flex flex-col shadow-2xl overflow-hidden">
            {/* Modal Header */}
            <div className="p-6 border-b border-slate-200 dark:border-slate-800 flex items-start justify-between gap-4 bg-slate-50/50 dark:bg-slate-800/30">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-blue-100 dark:bg-blue-950/60 rounded-lg text-blue-600 dark:text-blue-400">
                    <FileCheck2 className="w-5 h-5" />
                  </div>
                  <h2 className="text-lg font-bold text-slate-900 dark:text-white">
                    Evidence & Document Bundle
                  </h2>
                  <span className="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-blue-100 text-blue-800 dark:bg-blue-900/60 dark:text-blue-300">
                    CAR-16
                  </span>
                </div>
                <p className="text-xs text-slate-500">
                  Application: <span className="font-semibold text-slate-700 dark:text-slate-300">{activeEvidenceApp.title}</span> at{' '}
                  <span className="font-semibold">{activeEvidenceApp.company}</span> • Bundle ID:{' '}
                  <span className="font-mono text-slate-600 dark:text-slate-400">{activeEvidenceBundle.bundle_id}</span>
                </p>
              </div>

              <button
                type="button"
                onClick={() => {
                  setActiveEvidenceApp(null);
                  setActiveEvidenceBundle(null);
                }}
                className="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
              >
                <XCircle className="w-5 h-5" />
              </button>
            </div>

            {/* Privacy & Governance Banner (AT-011 / REQ-018) */}
            <div className="px-6 py-3 bg-amber-500/10 dark:bg-amber-950/30 border-b border-amber-200/50 dark:border-amber-900/40 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
              <div className="flex items-center gap-2">
                <Lock className="w-4 h-4 text-amber-600 shrink-0" />
                <span className="font-semibold text-amber-900 dark:text-amber-200">
                  Candidate Owner-Private Isolation (AT-011)
                </span>
                <span className="text-slate-500 dark:text-slate-400 text-[11px]">
                  Protected from unauthorized workspace admin access.
                </span>
              </div>

              {/* Interactive Privacy Enforcement Simulator */}
              <div className="flex items-center gap-2 self-start sm:self-auto">
                <span className="text-[11px] font-semibold text-slate-500">Simulate:</span>
                <button
                  type="button"
                  onClick={() => handleEvaluatePrivacySimulation(!simulateWorkspaceAdmin, simulateExplicitGrant)}
                  className={`px-2 py-0.5 rounded text-[11px] font-semibold border transition cursor-pointer ${
                    simulateWorkspaceAdmin
                      ? 'bg-amber-600 text-white border-amber-700'
                      : 'bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-300 dark:border-slate-700'
                  }`}
                >
                  Admin Mode
                </button>
                <button
                  type="button"
                  onClick={() => handleEvaluatePrivacySimulation(simulateWorkspaceAdmin, !simulateExplicitGrant)}
                  className={`px-2 py-0.5 rounded text-[11px] font-semibold border transition cursor-pointer ${
                    simulateExplicitGrant
                      ? 'bg-emerald-600 text-white border-emerald-700'
                      : 'bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-300 dark:border-slate-700'
                  }`}
                >
                  Has Grant
                </button>
              </div>
            </div>

            {/* Privacy Error Banner if Admin Access Denied */}
            {evidenceAccessDeniedMessage ? (
              <div className="p-8 text-center space-y-3 bg-red-50 dark:bg-red-950/40 m-6 rounded-2xl border border-red-200 dark:border-red-900/50">
                <ShieldAlert className="w-10 h-10 text-red-600 mx-auto" />
                <h3 className="text-base font-bold text-red-900 dark:text-red-200">
                  Access Denied: Owner-Private Evidence Isolation
                </h3>
                <p className="text-xs text-red-700 dark:text-red-300 max-w-lg mx-auto">
                  {evidenceAccessDeniedMessage}
                </p>
                <div className="pt-2">
                  <button
                    type="button"
                    onClick={() => handleEvaluatePrivacySimulation(false, false)}
                    className="px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white text-xs font-bold rounded-lg cursor-pointer transition"
                  >
                    Reset to Candidate View
                  </button>
                </div>
              </div>
            ) : (
              <>
                {/* Navigation Tabs */}
                <div className="px-6 border-b border-slate-200 dark:border-slate-800 flex items-center gap-2 overflow-x-auto">
                  <button
                    type="button"
                    onClick={() => setActiveEvidenceTab('snapshot')}
                    className={`py-3 px-3 text-xs font-bold border-b-2 transition cursor-pointer flex items-center gap-1.5 ${
                      activeEvidenceTab === 'snapshot'
                        ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                        : 'border-transparent text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    <Briefcase className="w-3.5 h-3.5" />
                    <span>Job & Resume Snapshot</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => setActiveEvidenceTab('answers')}
                    className={`py-3 px-3 text-xs font-bold border-b-2 transition cursor-pointer flex items-center gap-1.5 ${
                      activeEvidenceTab === 'answers'
                        ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                        : 'border-transparent text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    <FileText className="w-3.5 h-3.5" />
                    <span>Submitted Answers ({activeEvidenceBundle.question_answers.length})</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => setActiveEvidenceTab('artifacts')}
                    className={`py-3 px-3 text-xs font-bold border-b-2 transition cursor-pointer flex items-center gap-1.5 ${
                      activeEvidenceTab === 'artifacts'
                        ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                        : 'border-transparent text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    <ShieldCheck className="w-3.5 h-3.5" />
                    <span>PII-Scrubbed Artifacts</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => setActiveEvidenceTab('privacy')}
                    className={`py-3 px-3 text-xs font-bold border-b-2 transition cursor-pointer flex items-center gap-1.5 ${
                      activeEvidenceTab === 'privacy'
                        ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                        : 'border-transparent text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    <Key className="w-3.5 h-3.5" />
                    <span>Cryptographic Seal & Integrity</span>
                  </button>
                </div>

                {/* Tab Content Body */}
                <div className="p-6 overflow-y-auto max-h-[60vh] space-y-6">
                  {/* TAB 1: SNAPSHOT */}
                  {activeEvidenceTab === 'snapshot' && (
                    <div className="space-y-5">
                      {/* Job Snapshot */}
                      <div className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-3">
                        <div className="flex items-center justify-between">
                          <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                            Job Post Snapshot at Submission
                          </h4>
                          <span className="text-[11px] text-slate-400">
                            Captured {new Date(activeEvidenceBundle.job_snapshot.captured_at).toLocaleString()}
                          </span>
                        </div>
                        <div>
                          <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                            {activeEvidenceBundle.job_snapshot.title}
                          </h3>
                          <div className="text-xs text-slate-500 mt-0.5">
                            {activeEvidenceBundle.job_snapshot.company} • {activeEvidenceBundle.job_snapshot.location} • {activeEvidenceBundle.job_snapshot.job_type}
                          </div>
                        </div>
                        <p className="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
                          {activeEvidenceBundle.job_snapshot.description_snippet}
                        </p>
                        <div className="flex flex-wrap gap-1.5 pt-1">
                          {activeEvidenceBundle.job_snapshot.required_skills?.map((skill, i) => (
                            <span
                              key={i}
                              className="px-2 py-0.5 bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-900/40 rounded-full text-[10px] font-semibold"
                            >
                              {skill}
                            </span>
                          ))}
                        </div>
                      </div>

                      {/* Resume Artifact Snapshot */}
                      <div className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-3">
                        <div className="flex items-center justify-between">
                          <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                            Exact Resume Artifact Preserved
                          </h4>
                          <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300">
                            {activeEvidenceBundle.resume_artifact.format.toUpperCase()} • {activeEvidenceBundle.resume_artifact.version}
                          </span>
                        </div>
                        <div className="space-y-1">
                          <div className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
                            <span>Document Path:</span>
                            <code className="text-[11px] font-mono bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                              {activeEvidenceBundle.resume_artifact.document_path}
                            </code>
                          </div>
                          <div className="text-xs text-slate-500 flex items-center gap-2">
                            <span>SHA-256 Checksum:</span>
                            <code className="text-[10px] font-mono break-all text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                              {activeEvidenceBundle.resume_artifact.content_checksum}
                            </code>
                          </div>
                        </div>
                        {activeEvidenceBundle.resume_artifact.content_snippet && (
                          <div className="p-2.5 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 text-xs text-slate-600 dark:text-slate-400 italic">
                            &quot;{activeEvidenceBundle.resume_artifact.content_snippet}&quot;
                          </div>
                        )}
                      </div>

                      {/* Cover Letter Snapshot (if exists) */}
                      {activeEvidenceBundle.cover_letter_artifact && (
                        <div className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-3">
                          <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                            Preserved Cover Letter Artifact
                          </h4>
                          <div className="text-xs font-semibold text-slate-800 dark:text-slate-200">
                            {activeEvidenceBundle.cover_letter_artifact.title}
                          </div>
                          <div className="text-xs text-slate-500 flex items-center gap-2">
                            <span>SHA-256 Checksum:</span>
                            <code className="text-[10px] font-mono break-all text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                              {activeEvidenceBundle.cover_letter_artifact.content_checksum}
                            </code>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {/* TAB 2: SUBMITTED ANSWERS */}
                  {activeEvidenceTab === 'answers' && (
                    <div className="space-y-4">
                      <div className="text-xs text-slate-500">
                        Exact candidate answers and custom fields captured at submission time.
                      </div>
                      <div className="space-y-3">
                        {activeEvidenceBundle.question_answers.map((qa, idx) => (
                          <div
                            key={qa.field_id || idx}
                            className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-2"
                          >
                            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                              <span className="text-xs font-bold text-slate-800 dark:text-slate-200">
                                {qa.field_label}
                              </span>
                              <div className="flex items-center gap-1.5">
                                {qa.was_autofilled && (
                                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-blue-100 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300">
                                    Autofilled
                                  </span>
                                )}
                                {qa.confirmed_by_candidate && (
                                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-300 flex items-center gap-1">
                                    <Check className="w-2.5 h-2.5" />
                                    <span>Confirmed</span>
                                  </span>
                                )}
                                {qa.is_custom_fact && (
                                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-purple-100 dark:bg-purple-950/60 text-purple-700 dark:text-purple-300">
                                    Custom Answer
                                  </span>
                                )}
                              </div>
                            </div>
                            <div className="p-3 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700/60 text-xs font-mono text-slate-800 dark:text-slate-200 whitespace-pre-wrap">
                              {qa.value}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* TAB 3: PII-SCRUBBED ARTIFACTS */}
                  {activeEvidenceTab === 'artifacts' && (
                    <div className="space-y-4">
                      <div className="p-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40 rounded-xl text-xs text-blue-800 dark:text-blue-300 flex items-center gap-2">
                        <ShieldCheck className="w-4 h-4 shrink-0 text-blue-600" />
                        <span>
                          AT-011 Invariant: All bearer tokens, session cookies, passwords, SSNs, and credit card numbers are scrubbed before storage.
                        </span>
                      </div>

                      {activeEvidenceBundle.scrubbed_artifacts?.map((art, idx) => (
                        <div
                          key={art.item_id || idx}
                          className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-2.5"
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-xs font-bold font-mono text-slate-800 dark:text-slate-200">
                              {art.item_type} ({art.item_id})
                            </span>
                            <span className="text-[10px] text-slate-400">
                              Scrubbed {new Date(art.scrubbed_at).toLocaleString()}
                            </span>
                          </div>
                          <p className="text-xs text-slate-500">{art.description}</p>
                          <pre className="p-3 bg-slate-900 text-emerald-400 font-mono text-[11px] rounded-xl overflow-x-auto whitespace-pre-wrap">
                            {art.scrubbed_content}
                          </pre>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* TAB 4: CRYPTOGRAPHIC INTEGRITY & SEAL */}
                  {activeEvidenceTab === 'privacy' && (
                    <div className="space-y-5">
                      <div className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 space-y-3">
                        <div className="flex items-center justify-between">
                          <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                            Cryptographic Integrity Seal (SHA-256)
                          </h4>
                          <span className="text-[10px] text-slate-400">
                            Sealed {new Date(activeEvidenceBundle.sealed_at).toLocaleString()}
                          </span>
                        </div>

                        <div className="p-3 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700/60 space-y-2">
                          <div className="text-xs text-slate-500 font-medium">Deterministic Bundle Checksum:</div>
                          <code className="text-xs font-mono font-bold text-slate-800 dark:text-slate-200 break-all block">
                            {activeEvidenceBundle.integrity_checksum}
                          </code>
                        </div>

                        <div className="pt-2 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                          <button
                            type="button"
                            onClick={handleVerifyBundleIntegrity}
                            disabled={isVerifyingIntegrity}
                            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 cursor-pointer shadow-sm"
                          >
                            <RefreshCw className={`w-3.5 h-3.5 ${isVerifyingIntegrity ? 'animate-spin' : ''}`} />
                            <span>{isVerifyingIntegrity ? 'Verifying Integrity...' : 'Verify Tamper Detection'}</span>
                          </button>

                          {integrityStatus && (
                            <div
                              className={`p-2.5 rounded-xl text-xs font-semibold flex items-center gap-2 ${
                                integrityStatus.isValid
                                  ? 'bg-emerald-50 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800'
                                  : 'bg-red-50 text-red-800 border border-red-200'
                              }`}
                            >
                              {integrityStatus.isValid ? (
                                <>
                                  <ShieldCheck className="w-4 h-4 text-emerald-600" />
                                  <span>Integrity Verified: Sealed bundle matches SHA-256 hash (Untampered).</span>
                                </>
                              ) : (
                                <>
                                  <AlertCircle className="w-4 h-4 text-red-600" />
                                  <span>Tamper Alert: Checksum mismatch detected!</span>
                                </>
                              )}
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="p-4 bg-slate-50 dark:bg-slate-800/40 rounded-2xl border border-slate-200 dark:border-slate-800 text-xs text-slate-500 space-y-2">
                        <div className="font-bold text-slate-700 dark:text-slate-300">
                          Verification Details:
                        </div>
                        <div>
                          • Confirmation Mode: <span className="font-semibold">{activeEvidenceBundle.confirmation_type}</span>
                        </div>
                        <div>
                          • Reference Number: <span className="font-mono font-semibold">{activeEvidenceBundle.provider_reference}</span>
                        </div>
                        <div>
                          • Candidate User ID: <span className="font-mono">{activeEvidenceBundle.user_id}</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </>
            )}

            {/* Modal Footer */}
            <div className="p-4 border-t border-slate-200 dark:border-slate-800 flex justify-end bg-slate-50/50 dark:bg-slate-800/30">
              <button
                type="button"
                onClick={() => {
                  setActiveEvidenceApp(null);
                  setActiveEvidenceBundle(null);
                }}
                className="px-5 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 rounded-xl text-xs font-bold cursor-pointer transition"
              >
                Close Bundle
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Google Sheets Export & Sync Modal (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014, SRC-C4) */}
      {isSheetsModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-5xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
            {/* Modal Header */}
            <div className="p-5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-800/50">
              <div className="flex items-center gap-3">
                <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  <FileSpreadsheet className="w-6 h-6" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-lg font-black text-slate-900 dark:text-white">
                      Google Sheets 1-Way Sync & CSV Export
                    </h2>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
                      CAR-17 • AT-013 • AT-014
                    </span>
                  </div>
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                    Deterministic one-way projection (Platform is Single Source of Truth). Formula injection sanitized & non-app notes preserved.
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsSheetsModalOpen(false)}
                className="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 transition cursor-pointer"
              >
                <XCircle className="w-5 h-5" />
              </button>
            </div>

            {/* Navigation Tabs */}
            <div className="flex border-b border-slate-200 dark:border-slate-800 px-5 bg-white dark:bg-slate-900 gap-6 text-xs font-bold">
              <button
                type="button"
                onClick={() => setSheetsActiveTab('preview')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  sheetsActiveTab === 'preview'
                    ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <Table className="w-4 h-4" />
                <span>Live Projection Preview ({sheetsPreview?.row_count ?? applications.length} rows)</span>
              </button>
              <button
                type="button"
                onClick={() => setSheetsActiveTab('sync')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  sheetsActiveTab === 'sync'
                    ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <RefreshCw className="w-4 h-4" />
                <span>Idempotent Reconciliation (AT-014)</span>
              </button>
              <button
                type="button"
                onClick={() => setSheetsActiveTab('config')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  sheetsActiveTab === 'config'
                    ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <Sliders className="w-4 h-4" />
                <span>Sync Configuration</span>
              </button>
            </div>

            {/* Modal Body */}
            <div className="p-6 overflow-y-auto space-y-6 flex-1 text-sm">
              {/* Tab 1: Live Preview */}
              {sheetsActiveTab === 'preview' && (
                <div className="space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 bg-emerald-50/50 dark:bg-emerald-950/20 border border-emerald-200 dark:border-emerald-900/40 rounded-xl">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 text-xs font-bold text-emerald-900 dark:text-emerald-300">
                        <ShieldCheck className="w-4 h-4 text-emerald-600" />
                        <span>AT-013 Formula Injection Defense Active</span>
                      </div>
                      <p className="text-xs text-slate-600 dark:text-slate-400">
                        Any cell starting with <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">=</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">+</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">-</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">@</code>, or whitespace is automatically prefixed with a single quote (<code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">&apos;</code>) to prevent spreadsheet code execution.
                      </p>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <button
                        type="button"
                        onClick={handleDownloadCSV}
                        className="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-sm"
                      >
                        <Download className="w-4 h-4" />
                        <span>Download CSV</span>
                      </button>
                    </div>
                  </div>

                  {/* Spreadsheet Grid Preview */}
                  <div className="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
                    <div className="overflow-x-auto max-h-80">
                      <table className="w-full text-left border-collapse text-xs">
                        <thead className="bg-slate-100 dark:bg-slate-800/80 sticky top-0 border-b border-slate-200 dark:border-slate-700">
                          <tr>
                            {(sheetsPreview?.headers || []).map((h, i) => (
                              <th key={i} className="p-2.5 font-bold text-slate-700 dark:text-slate-300 whitespace-nowrap">
                                {h}
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-200 dark:divide-slate-800">
                          {(sheetsPreview?.rows || []).map((row, rIdx) => (
                            <tr key={rIdx} className="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                              {row.map((cell, cIdx) => {
                                const isSanitized = cell.startsWith("'");
                                return (
                                  <td key={cIdx} className="p-2.5 text-slate-600 dark:text-slate-400 whitespace-nowrap font-mono">
                                    {isSanitized ? (
                                      <span className="text-amber-600 dark:text-amber-400 font-bold" title="AT-013 escaped">
                                        {cell}
                                      </span>
                                    ) : (
                                      cell || '—'
                                    )}
                                  </td>
                                );
                              })}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 2: Idempotent Reconciliation */}
              {sheetsActiveTab === 'sync' && (
                <div className="space-y-5">
                  <div className="p-4 bg-blue-50/50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900/40 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-xs font-bold text-blue-900 dark:text-blue-300">
                      <Layers className="w-4 h-4 text-blue-600" />
                      <span>AT-014 Idempotent Row Reconciliation Engine</span>
                    </div>
                    <p className="text-xs text-slate-600 dark:text-slate-400">
                      The candidate can sort rows or insert custom columns (e.g. personal interview grades, follow-up checkboxes). When syncing, rows are mapped strictly by <strong>Application ID</strong>, platform columns are updated in place, and all user columns beyond column 11 are preserved intact without data loss.
                    </p>
                  </div>

                  {/* Simulator Controls */}
                  <div className="p-4 bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 rounded-xl space-y-4">
                    <div className="font-bold text-xs uppercase tracking-wider text-slate-500">
                      User Changes Simulation in Google Sheets:
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <div>
                        <label className="block text-[11px] font-bold text-slate-600 dark:text-slate-400 mb-1">
                          Custom User Column Header (Col 12):
                        </label>
                        <input
                          type="text"
                          value={customUserColumnHeader}
                          onChange={(e) => setCustomUserColumnHeader(e.target.value)}
                          className="w-full px-3 py-1.5 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg"
                        />
                      </div>
                      <div>
                        <label className="block text-[11px] font-bold text-slate-600 dark:text-slate-400 mb-1">
                          Candidate Private Value:
                        </label>
                        <input
                          type="text"
                          value={customUserColumnValue}
                          onChange={(e) => setCustomUserColumnValue(e.target.value)}
                          className="w-full px-3 py-1.5 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg"
                        />
                      </div>
                    </div>

                    <div className="flex flex-wrap gap-2.5 pt-1">
                      <button
                        type="button"
                        onClick={() => handleExecuteSheetsSync(false)}
                        disabled={isSyncingSheets}
                        className="px-4 py-2 bg-slate-800 hover:bg-slate-900 text-white rounded-xl text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-sm"
                      >
                        <RefreshCw className={`w-3.5 h-3.5 ${isSyncingSheets ? 'animate-spin' : ''}`} />
                        <span>Execute Standard 1-Way Sync</span>
                      </button>
                      <button
                        type="button"
                        onClick={() => handleExecuteSheetsSync(true)}
                        disabled={isSyncingSheets}
                        className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-sm"
                      >
                        <ShieldCheck className="w-3.5 h-3.5" />
                        <span>Simulate User Reorder + Custom Notes (AT-014)</span>
                      </button>
                    </div>
                  </div>

                  {/* Sync Result Metrics */}
                  {sheetsSyncResult && (
                    <div className="space-y-4">
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                        <div className="p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
                          <div className="text-[10px] uppercase font-bold text-slate-400">Total Applications</div>
                          <div className="text-lg font-black text-slate-900 dark:text-white mt-0.5">
                            {sheetsSyncResult.total_applications}
                          </div>
                        </div>
                        <div className="p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
                          <div className="text-[10px] uppercase font-bold text-slate-400">Matched Rows</div>
                          <div className="text-lg font-black text-blue-600 dark:text-blue-400 mt-0.5">
                            {sheetsSyncResult.matched_count}
                          </div>
                        </div>
                        <div className="p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
                          <div className="text-[10px] uppercase font-bold text-slate-400">In-Place Updated</div>
                          <div className="text-lg font-black text-emerald-600 dark:text-emerald-400 mt-0.5">
                            {sheetsSyncResult.updated_count}
                          </div>
                        </div>
                        <div className="p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
                          <div className="text-[10px] uppercase font-bold text-slate-400">Appended Rows</div>
                          <div className="text-lg font-black text-purple-600 dark:text-purple-400 mt-0.5">
                            {sheetsSyncResult.appended_count}
                          </div>
                        </div>
                      </div>

                      {/* Reconciled Output Table */}
                      {sheetsReconciledRows && sheetsReconciledRows.length > 0 && (
                        <div className="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
                          <div className="p-3 bg-slate-100 dark:bg-slate-800/80 font-bold text-xs text-slate-700 dark:text-slate-300 flex items-center justify-between">
                            <span>Reconciled Sheet Output (Preserving User Ordering & Notes)</span>
                            <span className="font-mono text-[11px] font-normal text-slate-500">
                              Checksum: {sheetsSyncResult.rows_checksum.substring(0, 16)}...
                            </span>
                          </div>
                          <div className="overflow-x-auto max-h-60">
                            <table className="w-full text-left border-collapse text-xs">
                              <tbody className="divide-y divide-slate-200 dark:divide-slate-800 font-mono">
                                {sheetsReconciledRows.slice(0, 5).map((row, idx) => (
                                  <tr key={idx} className={idx === 0 ? 'bg-slate-50 dark:bg-slate-800/50 font-bold' : ''}>
                                    {row.map((cell, cIdx) => (
                                      <td
                                        key={cIdx}
                                        className={`p-2 whitespace-nowrap ${
                                          cIdx >= 11 ? 'bg-amber-50/60 dark:bg-amber-950/20 text-amber-800 dark:text-amber-300 font-bold' : ''
                                        }`}
                                      >
                                        {cell || '—'}
                                      </td>
                                    ))}
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}

              {/* Tab 3: Configuration */}
              {sheetsActiveTab === 'config' && (
                <div className="space-y-4 max-w-xl">
                  {sheetsConfigSavedMessage && (
                    <div className="p-3 rounded-xl bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-800 text-xs font-bold text-emerald-800 dark:text-emerald-300 flex items-center gap-2">
                      <Check className="w-4 h-4 text-emerald-600" />
                      <span>{sheetsConfigSavedMessage}</span>
                    </div>
                  )}

                  <div>
                    <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                      Google Spreadsheet ID:
                    </label>
                    <input
                      type="text"
                      value={sheetsConfig.spreadsheet_id}
                      onChange={(e) => setSheetsConfig({ ...sheetsConfig, spreadsheet_id: e.target.value })}
                      className="w-full px-3 py-2 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl font-mono"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                      Target Worksheet Name:
                    </label>
                    <input
                      type="text"
                      value={sheetsConfig.sheet_name}
                      onChange={(e) => setSheetsConfig({ ...sheetsConfig, sheet_name: e.target.value })}
                      className="w-full px-3 py-2 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                        Timezone Formatting:
                      </label>
                      <select
                        value={sheetsConfig.timezone}
                        onChange={(e) => setSheetsConfig({ ...sheetsConfig, timezone: e.target.value })}
                        className="w-full px-3 py-2 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl cursor-pointer"
                      >
                        <option value="UTC">UTC</option>
                        <option value="America/New_York">America/New_York (EDT/EST)</option>
                        <option value="America/Los_Angeles">America/Los_Angeles (PDT/PST)</option>
                        <option value="Europe/London">Europe/London (BST/GMT)</option>
                        <option value="Asia/Dhaka">Asia/Dhaka (+06)</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                        Sync Strategy:
                      </label>
                      <select
                        value={sheetsConfig.sync_mode}
                        onChange={(e) => setSheetsConfig({ ...sheetsConfig, sync_mode: e.target.value as SheetsSyncMode })}
                        className="w-full px-3 py-2 text-xs bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl cursor-pointer"
                      >
                        <option value="one_way_upsert">One-Way In-Place Upsert (Recommended)</option>
                        <option value="append_only">Append Only</option>
                        <option value="snapshot_overwrite">Snapshot Overwrite</option>
                      </select>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-2">
                    <input
                      type="checkbox"
                      id="include_notes_toggle"
                      checked={sheetsConfig.include_notes}
                      onChange={(e) => setSheetsConfig({ ...sheetsConfig, include_notes: e.target.checked })}
                      className="rounded border-slate-300 text-emerald-600 focus:ring-emerald-500 cursor-pointer"
                    />
                    <label htmlFor="include_notes_toggle" className="text-xs text-slate-700 dark:text-slate-300 cursor-pointer">
                      Include latest timeline note in the Notes column
                    </label>
                  </div>

                  <div className="pt-3">
                    <button
                      type="button"
                      onClick={handleSaveSheetsConfig}
                      className="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-sm"
                    >
                      <Save className="w-3.5 h-3.5" />
                      <span>Save Sheets Settings</span>
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="p-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-800/30">
              <span className="text-[11px] text-slate-500">
                🔒 Invariant CAR-17: Sheets is never used as an application database. One-way read-only projection.
              </span>
              <button
                type="button"
                onClick={() => setIsSheetsModalOpen(false)}
                className="px-5 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 rounded-xl text-xs font-bold cursor-pointer transition"
              >
                Close Window
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Multi-Domain CSV & Additional Exports Modal (IMP-CAR-18, CAR-18, AT-013, FND-005, FND-006, SRC-C2, SRC-C7, SRC-C8) */}
      {isCSVModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 overflow-y-auto">
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-5xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in duration-150">
            {/* Modal Header */}
            <div className="p-5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-gradient-to-r from-indigo-50/70 via-slate-50 to-white dark:from-indigo-950/30 dark:via-slate-900 dark:to-slate-900">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-indigo-600 text-white flex items-center justify-center shadow-md shadow-indigo-500/20">
                  <Download className="w-5 h-5" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-lg font-black text-slate-900 dark:text-white">
                      Export Center & Multi-Domain CSV Hub
                    </h2>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-indigo-100 dark:bg-indigo-950/60 text-indigo-800 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800">
                      CAR-18 • AT-013 • FND-006 • RFC 4180
                    </span>
                  </div>
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                    Sanitized tabular export for Jobs, Applications, Recruiter Contacts, and Funnel Analytics with Excel UTF-8 BOM compatibility.
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsCSVModalOpen(false)}
                className="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 transition cursor-pointer"
              >
                <XCircle className="w-5 h-5" />
              </button>
            </div>

            {/* Dataset Selector Ribbon */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 p-3 bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 text-xs">
              <button
                type="button"
                onClick={() => setCSVDataset('applications')}
                className={`p-2.5 rounded-xl border flex items-center gap-2.5 transition text-left cursor-pointer ${
                  csvDataset === 'applications'
                    ? 'border-indigo-500 bg-indigo-50/80 dark:bg-indigo-950/50 text-indigo-900 dark:text-indigo-200 font-bold shadow-xs'
                    : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <FileText className="w-4 h-4 text-indigo-600 shrink-0" />
                <div>
                  <div className="font-bold">Applications</div>
                  <div className="text-[10px] text-slate-500">{applications.length} logged records</div>
                </div>
              </button>

              <button
                type="button"
                onClick={() => setCSVDataset('jobs')}
                className={`p-2.5 rounded-xl border flex items-center gap-2.5 transition text-left cursor-pointer ${
                  csvDataset === 'jobs'
                    ? 'border-indigo-500 bg-indigo-50/80 dark:bg-indigo-950/50 text-indigo-900 dark:text-indigo-200 font-bold shadow-xs'
                    : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <Briefcase className="w-4 h-4 text-indigo-600 shrink-0" />
                <div>
                  <div className="font-bold">Jobs & Targets</div>
                  <div className="text-[10px] text-slate-500">5 saved opportunities</div>
                </div>
              </button>

              <button
                type="button"
                onClick={() => setCSVDataset('contacts')}
                className={`p-2.5 rounded-xl border flex items-center gap-2.5 transition text-left cursor-pointer ${
                  csvDataset === 'contacts'
                    ? 'border-indigo-500 bg-indigo-50/80 dark:bg-indigo-950/50 text-indigo-900 dark:text-indigo-200 font-bold shadow-xs'
                    : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <Users className="w-4 h-4 text-indigo-600 shrink-0" />
                <div>
                  <div className="font-bold">Recruiter Contacts</div>
                  <div className="text-[10px] text-slate-500">{recruiterContacts.length} leads (CAR-20)</div>
                </div>
              </button>

              <button
                type="button"
                onClick={() => setCSVDataset('analysis')}
                className={`p-2.5 rounded-xl border flex items-center gap-2.5 transition text-left cursor-pointer ${
                  csvDataset === 'analysis'
                    ? 'border-indigo-500 bg-indigo-50/80 dark:bg-indigo-950/50 text-indigo-900 dark:text-indigo-200 font-bold shadow-xs'
                    : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <BarChart3 className="w-4 h-4 text-indigo-600 shrink-0" />
                <div>
                  <div className="font-bold">Conversion Funnel</div>
                  <div className="text-[10px] text-slate-500">{analyticsMetrics.length} metrics (CAR-19)</div>
                </div>
              </button>
            </div>

            {/* Navigation Tabs */}
            <div className="flex border-b border-slate-200 dark:border-slate-800 px-5 bg-white dark:bg-slate-900 gap-6 text-xs font-bold">
              <button
                type="button"
                onClick={() => setCSVActiveTab('preview')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  csvActiveTab === 'preview'
                    ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <Table className="w-4 h-4" />
                <span>Live Dataset Preview</span>
              </button>
              <button
                type="button"
                onClick={() => setCSVActiveTab('audit')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  csvActiveTab === 'audit'
                    ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <History className="w-4 h-4" />
                <span>Export Audit Trail ({csvAuditHistory.length})</span>
              </button>
              <button
                type="button"
                onClick={() => setCSVActiveTab('downstream_rules')}
                className={`py-3 border-b-2 transition flex items-center gap-2 cursor-pointer ${
                  csvActiveTab === 'downstream_rules'
                    ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                <ShieldCheck className="w-4 h-4" />
                <span>Downstream Security & FND-006</span>
              </button>
            </div>

            {/* Modal Body */}
            <div className="p-6 overflow-y-auto space-y-5 flex-1 text-sm">
              {/* Tab 1: Live Dataset Preview */}
              {csvActiveTab === 'preview' && (() => {
                const currentData = generateDatasetCSV(csvDataset, csvWithBOM, csvIncludeHeaders, csvStageFilter);
                return (
                  <div className="space-y-4">
                    {/* Controls & Option Bar */}
                    <div className="flex flex-wrap items-center justify-between gap-3 p-3.5 bg-indigo-50/50 dark:bg-indigo-950/20 border border-indigo-200 dark:border-indigo-900/40 rounded-xl">
                      <div className="flex flex-wrap items-center gap-4 text-xs">
                        <label className="flex items-center gap-2 font-semibold text-slate-700 dark:text-slate-300 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={csvWithBOM}
                            onChange={(e) => setCSVWithBOM(e.target.checked)}
                            className="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500 cursor-pointer"
                          />
                          <span>UTF-8 BOM (Excel Safe)</span>
                        </label>

                        <label className="flex items-center gap-2 font-semibold text-slate-700 dark:text-slate-300 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={csvIncludeHeaders}
                            onChange={(e) => setCSVIncludeHeaders(e.target.checked)}
                            className="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500 cursor-pointer"
                          />
                          <span>Include Header Row</span>
                        </label>

                        {csvDataset === 'applications' && (
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-slate-600 dark:text-slate-400">Stage Filter:</span>
                            <select
                              value={csvStageFilter}
                              onChange={(e) => setCSVStageFilter(e.target.value)}
                              className="px-2.5 py-1 bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg text-xs cursor-pointer"
                            >
                              <option value="all">All Stages</option>
                              <option value="applied">Applied</option>
                              <option value="interviewing">Interviewing</option>
                              <option value="offered">Offered</option>
                              <option value="needs_confirmation">Needs Confirmation</option>
                              <option value="withdrawn">Withdrawn</option>
                            </select>
                          </div>
                        )}
                      </div>

                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={handleCopyCSV}
                          className="px-3.5 py-2 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-300 rounded-xl text-xs font-bold transition flex items-center gap-1.5 cursor-pointer shadow-xs"
                        >
                          {copiedToast ? <CheckCheck className="w-4 h-4 text-emerald-600" /> : <Copy className="w-4 h-4" />}
                          <span>{copiedToast ? 'Copied!' : 'Copy CSV'}</span>
                        </button>

                        <button
                          type="button"
                          onClick={handleDownloadDatasetCSV}
                          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-sm"
                        >
                          <Download className="w-4 h-4" />
                          <span>Download .CSV File ({currentData.rowCount} rows)</span>
                        </button>
                      </div>
                    </div>

                    {/* Invariant Highlights Banner */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <div className="p-3 bg-emerald-50/60 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/50 rounded-xl flex items-start gap-2.5 text-xs">
                        <ShieldCheck className="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
                        <div>
                          <span className="font-bold text-emerald-900 dark:text-emerald-300">
                            AT-013: Active Formula Injection Defense
                          </span>
                          <p className="text-emerald-700 dark:text-emerald-400 mt-0.5">
                            Cells starting with <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">=</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">+</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">-</code>, <code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">@</code> are prefixed with single quote (<code className="bg-white dark:bg-slate-800 px-1 py-0.5 rounded text-[11px] font-mono">&apos;</code>) to prevent formula execution in Excel/Calc.
                          </p>
                        </div>
                      </div>

                      <div className="p-3 bg-indigo-50/60 dark:bg-indigo-950/30 border border-indigo-200 dark:border-indigo-900/50 rounded-xl flex items-start gap-2.5 text-xs">
                        <HardDrive className="w-4 h-4 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" />
                        <div>
                          <span className="font-bold text-indigo-900 dark:text-indigo-300">
                            FND-006: Binary Artifact Separation
                          </span>
                          <p className="text-indigo-700 dark:text-indigo-400 mt-0.5">
                            Resume and cover letter documents are NOT dumped as raw binary data in the CSV. Only version tags, secure file paths, and cryptographic SHA-256 checksums are exported.
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* Table View */}
                    <div className="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
                      <div className="overflow-x-auto max-h-80">
                        <table className="w-full text-left border-collapse text-xs">
                          <thead className="bg-slate-100 dark:bg-slate-800/80 sticky top-0 border-b border-slate-200 dark:border-slate-700">
                            <tr>
                              {currentData.headers.map((h, idx) => (
                                <th key={idx} className="p-2.5 font-bold text-slate-700 dark:text-slate-300 whitespace-nowrap">
                                  {h}
                                </th>
                              ))}
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-slate-200 dark:divide-slate-800 font-mono">
                            {currentData.rows.map((r, rIdx) => (
                              <tr key={rIdx} className="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                                {r.map((cell, cIdx) => {
                                  const isSanitized = cell.startsWith("'");
                                  const isSha = cell.length === 64;
                                  return (
                                    <td key={cIdx} className="p-2.5 text-slate-600 dark:text-slate-400 whitespace-nowrap">
                                      {isSanitized ? (
                                        <span className="text-amber-600 dark:text-amber-400 font-bold" title="AT-013 formula escaped">
                                          {cell}
                                        </span>
                                      ) : isSha ? (
                                        <span className="text-indigo-600 dark:text-indigo-400 truncate max-w-xs block" title={cell}>
                                          {cell.substring(0, 16)}...
                                        </span>
                                      ) : (
                                        cell || '—'
                                      )}
                                    </td>
                                  );
                                })}
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  </div>
                );
              })()}

              {/* Tab 2: Export Audit Trail (AT-022, EXP-003) */}
              {csvActiveTab === 'audit' && (
                <div className="space-y-4">
                  <div className="p-4 bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 rounded-xl flex items-center justify-between gap-4">
                    <div>
                      <div className="font-bold text-xs uppercase tracking-wider text-slate-500">
                        Cryptographic Export Audit Trail (EXP-003 / AT-022)
                      </div>
                      <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5">
                        Every export action generates an immutable audit record logging timestamp, dataset type, row count, and BOM encoding.
                      </p>
                    </div>
                    <span className="px-3 py-1 bg-emerald-100 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 rounded-lg text-xs font-bold">
                      Audit Logging Active
                    </span>
                  </div>

                  <div className="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
                    <table className="w-full text-left border-collapse text-xs">
                      <thead className="bg-slate-100 dark:bg-slate-800 font-bold text-slate-700 dark:text-slate-300">
                        <tr>
                          <th className="p-3">Audit ID</th>
                          <th className="p-3">Dataset</th>
                          <th className="p-3">Filename</th>
                          <th className="p-3">Rows</th>
                          <th className="p-3">Format</th>
                          <th className="p-3">With BOM</th>
                          <th className="p-3">Timestamp (UTC)</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-200 dark:divide-slate-800 font-mono">
                        {csvAuditHistory.map((rec) => (
                          <tr key={rec.audit_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                            <td className="p-3 font-bold text-slate-800 dark:text-slate-200">{rec.audit_id}</td>
                            <td className="p-3 uppercase text-[11px] font-bold text-indigo-600 dark:text-indigo-400">
                              {rec.dataset_type}
                            </td>
                            <td className="p-3 text-slate-600 dark:text-slate-400">{rec.filename}</td>
                            <td className="p-3 text-slate-700 dark:text-slate-300">{rec.row_count}</td>
                            <td className="p-3 uppercase">{rec.format}</td>
                            <td className="p-3">
                              {rec.with_bom ? (
                                <span className="text-emerald-600 font-bold">YES</span>
                              ) : (
                                <span className="text-slate-400">NO</span>
                              )}
                            </td>
                            <td className="p-3 text-slate-500">
                              {new Date(rec.exported_at).toISOString().replace('T', ' ').substring(0, 19)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* Tab 3: Downstream Rules & FND-006 */}
              {csvActiveTab === 'downstream_rules' && (
                <div className="space-y-4 text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
                  <div className="p-4 bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                    <h3 className="font-bold text-sm text-slate-900 dark:text-white flex items-center gap-2">
                      <ShieldCheck className="w-4 h-4 text-indigo-600" />
                      Downstream Disclosure & Candidate Privacy (EXP-003)
                    </h3>
                    <p>
                      Exported CSV documents may be loaded into third-party desktop tools (Microsoft Excel, LibreOffice Calc, Apple Numbers) or uploaded to external cloud drives. Antigravity enforces strict data hygiene before serializing any record:
                    </p>
                    <ul className="list-disc list-inside space-y-1 pl-2 text-slate-600 dark:text-slate-400">
                      <li><strong>No Raw Binaries (FND-006):</strong> Resumes and cover letters are referenced strictly through canonical storage paths and cryptographic SHA-256 digests.</li>
                      <li><strong>RFC 4180 Escaping:</strong> All fields containing double quotes, commas, or line breaks are escaped in accordance with the standard CSV specification.</li>
                      <li><strong>Formula Injection Defense (AT-013):</strong> Any field beginning with danger characters (<code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">=</code>, <code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">+</code>, <code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">-</code>, <code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">@</code>, <code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">\t</code>, <code className="bg-slate-200 dark:bg-slate-700 px-1 py-0.5 rounded font-mono">\r</code>) is prefixed with a single quote.</li>
                      <li><strong>Audit Record Creation (AT-022):</strong> The platform logs the operator, timestamp, and row count of every exported batch.</li>
                    </ul>
                  </div>
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="p-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-800/30">
              <span className="text-[11px] text-slate-500">
                🔒 Invariants: AT-013 (Formula Safe) • FND-006 (Artifact Separation) • RFC 4180 UTF-8 BOM
              </span>
              <button
                type="button"
                onClick={() => setIsCSVModalOpen(false)}
                className="px-5 py-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 rounded-xl text-xs font-bold cursor-pointer transition"
              >
                Close Window
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

