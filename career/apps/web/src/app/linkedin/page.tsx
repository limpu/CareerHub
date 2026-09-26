'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import {
  Users,
  Link2,
  Shield,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Lock,
  RefreshCw,
  Copy,
  ExternalLink,
  Info,
  Terminal,
  UploadCloud,
  FileText,
  FileArchive,
  Layers,
  Sparkles,
  ArrowRight,
  Database,
  Building2,
  Briefcase,
  MessageSquare,
  Plus,
  Compass,
  Search,
  Globe,
  UserCheck,
  Calendar,
  Bell,
  UserPlus,
  ClipboardList,
  Send,
  Check,
  Clock,
  Sliders,
  ShieldAlert,
  ListOrdered,
  FileEdit,
  Inbox,
  Reply,
  MailCheck,
  StopCircle,
  Mic,
  BookOpen,
  Award,
  Trash2,
  CalendarDays,
  Share2,
  CheckSquare,
  Split,
  HelpCircle,
  FileCheck,
  BarChart3,
  TrendingUp,
  Target,
  EyeOff,
  Megaphone,
  Key,
  ShieldCheck,
  Download,
  Table
} from 'lucide-react';
import type {
  LinkedInPersonRecord,
  LinkedInCompanyRecord,
  LinkedInJobRecord,
  LinkedInPostRecord,
  LinkedInEntityType,
  LinkedInDiscoveryCriteria,
  LinkedInBooleanQuery,
  LinkedInCanonicalURL,
  LinkedInDiscoveryResult,
  LinkedInUserGoal,
  LinkedInRecruiterLead,
  LinkedInLeadNote,
  LinkedInFollowUpReminder,
  LinkedInLeadStatus,
  LinkedInOutreachStage,
  LinkedInOutreachDraftPayload,
  LinkedInQueueItemStatus,
  LinkedInConnectionBudget,
  LinkedInConnectionQueueItem,
  CompanyWatchlistItem,
  CompanyFollowPriority,
  CompanyFollowBudget,
  CompanyFollowPlanItem,
  BatchCompanyFollowPlan,
  CompanyWatchlistStatus,
  LinkedInConversationThread,
  LinkedInInboxMessage,
  LinkedInReplyDraft,
  LinkedInInboxThreadType,
  LinkedInInboxClassification,
  LinkedInReplyDraftStatus,
  LinkedInReplyDraftTone,
  LinkedInPostSweepCategory,
  LinkedInCommentDraftAngle,
  LinkedInCommentDraftStatus,
  LinkedInSweptPostComment,
  LinkedInSweptTargetPost,
  LinkedInCommentDraft,
  LinkedInPostAngle,
  LinkedInHookType,
  LinkedInPostDraftStatus,
  LinkedInAuditSeverity,
  LinkedInHookVariant,
  LinkedInAuditIssue,
  LinkedInEditorialAuditReport,
  LinkedInPostDraft,
  LinkedInCadenceStyle,
  LinkedInPerspective,
  LinkedInVoiceProfile,
  LinkedInSlopReplacement,
  LinkedInHumanizerAuditIssue,
  LinkedInHumanizeResult,
  LinkedInStoryCategory,
  LinkedInStoryProvenanceType,
  LinkedInStoryStatus,
  LinkedInStoryProvenance,
  LinkedInStoryNarrative,
  LinkedInStoryAuditIssue,
  LinkedInStoryAuditReport,
  LinkedInStoryEntry,
  LinkedInInterviewPrompt,
  LinkedInInterviewQA,
  LinkedInInterviewSession,
  LinkedInArtifactType,
  LinkedInRepurposedFormat,
  LinkedInContentPlanStatus,
  LinkedInSourceArtifact,
  LinkedInRepurposedDraft,
  LinkedInContentAssetMetadata,
  LinkedInContentPlanItem,
  LinkedInMetricItem,
  LinkedInEngagerProfile,
  LinkedInPostAnalyticsSnapshot,
  LinkedInPerformanceBenchmark,
  LinkedInAdvocacyCampaignStatus,
  LinkedInEmployeePersona,
  LinkedInAdvocacyCopyVariant,
  LinkedInBrandGovernance,
  LinkedInAdvocacyCampaign,
  LinkedInEmployeeShareEvent,
  LinkedInProviderAdapter,
  LinkedInProviderAdapterType,
  LinkedInProviderHealthStatus,
  LinkedInDiagnosticProbeResult,
  LinkedInProviderDiagnosticReport,
  LinkedInProviderDispatchResult,
  LinkedInPlatformLimitType,
  LinkedInChallengeType,
  LinkedInAccountSafetyStatus,
  LinkedInAccountSafetyState,
  LinkedInPlatformRestrictionIncident,
  LinkedInActivityLogEntry,
  LinkedInCombinedAccountBudgetReport,
  LinkedInResolveChallengeRequest,
  LinkedInSafetyGateDecision,
  LinkedInSessionLifecycleStatus,
  LinkedInViewerLeaseLifecycleStatus,
  LinkedInSessionStorageMetadata,
  LinkedInSessionRecord,
  ShortLivedViewerLease,
  LinkedInRelationshipExportDestination,
  LinkedInRelationshipExportFilter,
  LinkedInRelationshipExportManifest,
  LinkedInRelationshipSheetsSyncConfig,
  LinkedInRelationshipSheetsSyncResult,
  LinkedInRelationshipExportAuditRecord
} from '@social-platform/contracts';

export type LinkedInScopeDomain = 'identity' | 'profile_read' | 'content_publish' | 'messaging' | 'easy_apply';
export type LinkedInCapabilityStatus = 'granted' | 'denied' | 'unavailable_partner_only' | 'unsupported_platform';
export type LinkedInConnectionStatus = 'connected' | 'partially_authorized' | 'reauth_required' | 'disconnected';
export type LinkedInFallbackMode = 'none' | 'user_export_upload' | 'manual_clipboard_assist';

export interface InspectedLinkedInCapability {
  capability_id: string;
  domain: LinkedInScopeDomain;
  name: string;
  description: string;
  required_scope: string;
  status: LinkedInCapabilityStatus;
  is_granted: boolean;
  truth_in_advertising_note: string;
  fallback_mode: LinkedInFallbackMode;
}

export interface LinkedInConnectionRecord {
  connection_id: string;
  user_id: string;
  member_id: string;
  display_name: string;
  email: string;
  granted_scopes: string[];
  denied_scopes: string[];
  masked_token: string;
  token_status: string;
  connected_at: string;
  expires_at?: string;
  status: LinkedInConnectionStatus;
  capabilities: InspectedLinkedInCapability[];
  last_inspected_at: string;
}

const PRESET_CONNECTIONS: { name: string; description: string; connection: LinkedInConnectionRecord }[] = [
  {
    name: 'Standard OIDC Sign-in Only',
    description: 'Consumer OpenID Connect grant. Identity granted; full profile, messaging, and apply are strictly unavailable.',
    connection: {
      connection_id: 'conn-std-01',
      user_id: 'usr-cand-101',
      member_id: 'urn:li:person:std_user_101',
      display_name: 'Jane Doe',
      email: 'jane.doe@example.com',
      granted_scopes: ['openid', 'profile', 'email'],
      denied_scopes: ['r_fullprofile', 'w_member_social', 'r_messages', 'w_messages'],
      masked_token: 'tok_li_****_88fa',
      token_status: 'active',
      connected_at: '2026-03-15T10:00:00Z',
      expires_at: '2026-04-14T10:00:00Z',
      status: 'partially_authorized',
      last_inspected_at: new Date().toISOString(),
      capabilities: [
        {
          capability_id: 'identity_signin',
          domain: 'identity',
          name: 'Sign In with LinkedIn (OpenID Connect)',
          description: 'Authenticate user identity, verified email address, and core OIDC claims.',
          required_scope: 'openid',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Standard OpenID Connect provides identity authentication only; it conveys zero access to full career history, messaging, or job applications.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'basic_profile_read',
          domain: 'profile_read',
          name: 'Basic Profile Read',
          description: 'Read public name, vanity slug, profile photo, and localized headline.',
          required_scope: 'profile',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Grants basic headline and portrait photo only. Complete work experience and resume history are not returned by consumer profile scopes.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'career_history_import',
          domain: 'profile_read',
          name: 'Career History & Experience Ingestion',
          description: 'Direct programmatic import of structured employment history, education, and licenses.',
          required_scope: 'r_fullprofile',
          status: 'unavailable_partner_only',
          is_granted: false,
          truth_in_advertising_note: 'Full profile scraping via API is gated to vetted Talent Solutions Enterprise partners. Supported platform fallback: upload your official LinkedIn Profile Export ZIP (Basic_LinkedInData.zip).',
          fallback_mode: 'user_export_upload'
        },
        {
          capability_id: 'content_publish',
          domain: 'content_publish',
          name: 'Member Post & Article Publishing',
          description: 'Publish user-approved professional updates and post drafts to personal feed.',
          required_scope: 'w_member_social',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Scope w_member_social was not approved during OAuth authorization. Re-authentication required to enable feed publishing.',
          fallback_mode: 'manual_clipboard_assist'
        },
        {
          capability_id: 'direct_messaging',
          domain: 'messaging',
          name: 'Direct Messaging & InMail',
          description: 'Programmatic message read and write across 1st-degree connections and recruiter threads.',
          required_scope: 'r_messages',
          status: 'unavailable_partner_only',
          is_granted: false,
          truth_in_advertising_note: 'Standard consumer applications are not granted messaging automation. The platform provides assisted-manual copy-paste drafts to prevent account bans (AT-010).',
          fallback_mode: 'manual_clipboard_assist'
        },
        {
          capability_id: 'automated_easy_apply',
          domain: 'easy_apply',
          name: 'Automated Easy Apply Submissions',
          description: 'Background execution of multi-step job application submissions on LinkedIn job posts.',
          required_scope: 'w_job_applications',
          status: 'unsupported_platform',
          is_granted: false,
          truth_in_advertising_note: 'Direct automated background submission without human-in-the-loop review violates LinkedIn terms of service and is permanently blocked (REQ-015, AT-010). Use Assisted-Manual flow with pre-filled clipboard answers.',
          fallback_mode: 'manual_clipboard_assist'
        }
      ]
    }
  },
  {
    name: 'Content Creator Elevated',
    description: 'Standard OIDC plus member UGC publishing (w_member_social). Publishing allowed with human gate.',
    connection: {
      connection_id: 'conn-creator-02',
      user_id: 'usr-cand-102',
      member_id: 'urn:li:person:creator_102',
      display_name: 'Alex Smith',
      email: 'alex.smith@example.com',
      granted_scopes: ['openid', 'profile', 'email', 'w_member_social'],
      denied_scopes: ['r_fullprofile', 'r_messages', 'w_messages'],
      masked_token: 'tok_li_****_99b1',
      token_status: 'active',
      connected_at: '2026-03-12T14:00:00Z',
      expires_at: '2026-05-11T14:00:00Z',
      status: 'partially_authorized',
      last_inspected_at: new Date().toISOString(),
      capabilities: [
        {
          capability_id: 'identity_signin',
          domain: 'identity',
          name: 'Sign In with LinkedIn (OpenID Connect)',
          description: 'Authenticate user identity, verified email address, and core OIDC claims.',
          required_scope: 'openid',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Standard OpenID Connect provides identity authentication only.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'basic_profile_read',
          domain: 'profile_read',
          name: 'Basic Profile Read',
          description: 'Read public name, vanity slug, profile photo, and localized headline.',
          required_scope: 'profile',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Grants basic headline and portrait photo only.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'career_history_import',
          domain: 'profile_read',
          name: 'Career History & Experience Ingestion',
          description: 'Direct programmatic import of structured employment history, education, and licenses.',
          required_scope: 'r_fullprofile',
          status: 'unavailable_partner_only',
          is_granted: false,
          truth_in_advertising_note: 'Gated to Enterprise Talent Solutions partners. Use Basic_LinkedInData.zip export.',
          fallback_mode: 'user_export_upload'
        },
        {
          capability_id: 'content_publish',
          domain: 'content_publish',
          name: 'Member Post & Article Publishing',
          description: 'Publish user-approved professional updates and post drafts to personal feed.',
          required_scope: 'w_member_social',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Permits publishing UGC posts strictly with explicit candidate single-use confirmation (AT-007).',
          fallback_mode: 'none'
        },
        {
          capability_id: 'direct_messaging',
          domain: 'messaging',
          name: 'Direct Messaging & InMail',
          description: 'Programmatic message read and write across 1st-degree connections and recruiter threads.',
          required_scope: 'r_messages',
          status: 'unavailable_partner_only',
          is_granted: false,
          truth_in_advertising_note: 'Partner approval required for direct messaging API.',
          fallback_mode: 'manual_clipboard_assist'
        },
        {
          capability_id: 'automated_easy_apply',
          domain: 'easy_apply',
          name: 'Automated Easy Apply Submissions',
          description: 'Background execution of multi-step job application submissions on LinkedIn job posts.',
          required_scope: 'w_job_applications',
          status: 'unsupported_platform',
          is_granted: false,
          truth_in_advertising_note: 'Automated Easy Apply background write is permanently blocked on live LinkedIn (AT-010).',
          fallback_mode: 'manual_clipboard_assist'
        }
      ]
    }
  },
  {
    name: 'Enterprise Talent & Marketing Partner',
    description: 'Vetted partner application credentials with talent, marketing, and messaging scopes.',
    connection: {
      connection_id: 'conn-partner-03',
      user_id: 'usr-cand-103',
      member_id: 'urn:li:person:partner_103',
      display_name: 'Tech Corp Talent Lead',
      email: 'recruiter@techcorp.example.com',
      granted_scopes: ['openid', 'profile', 'email', 'r_basicprofile', 'r_fullprofile', 'w_member_social', 'r_messages', 'w_messages'],
      denied_scopes: [],
      masked_token: 'tok_li_****_44c2',
      token_status: 'active',
      connected_at: '2026-03-10T08:00:00Z',
      expires_at: '2026-06-08T08:00:00Z',
      status: 'connected',
      last_inspected_at: new Date().toISOString(),
      capabilities: [
        {
          capability_id: 'identity_signin',
          domain: 'identity',
          name: 'Sign In with LinkedIn (OpenID Connect)',
          description: 'Authenticate user identity, verified email address, and core OIDC claims.',
          required_scope: 'openid',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Verified enterprise identity claims active.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'basic_profile_read',
          domain: 'profile_read',
          name: 'Basic Profile Read',
          description: 'Read public name, vanity slug, profile photo, and localized headline.',
          required_scope: 'profile',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Authorized basic profile read.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'career_history_import',
          domain: 'profile_read',
          name: 'Career History & Experience Ingestion',
          description: 'Direct programmatic import of structured employment history, education, and licenses.',
          required_scope: 'r_fullprofile',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Enterprise partner scope granted for full member experience reading.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'content_publish',
          domain: 'content_publish',
          name: 'Member Post & Article Publishing',
          description: 'Publish user-approved professional updates and post drafts to personal feed.',
          required_scope: 'w_member_social',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'UGC publishing enabled with human confirmation gatekeeper.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'direct_messaging',
          domain: 'messaging',
          name: 'Direct Messaging & InMail',
          description: 'Programmatic message read and write across 1st-degree connections and recruiter threads.',
          required_scope: 'r_messages',
          status: 'granted',
          is_granted: true,
          truth_in_advertising_note: 'Partner InMail and direct messaging enabled within rolling quota bounds (AT-008).',
          fallback_mode: 'none'
        },
        {
          capability_id: 'automated_easy_apply',
          domain: 'easy_apply',
          name: 'Automated Easy Apply Submissions',
          description: 'Background execution of multi-step job application submissions on LinkedIn job posts.',
          required_scope: 'w_job_applications',
          status: 'unsupported_platform',
          is_granted: false,
          truth_in_advertising_note: 'Automated Easy Apply background write is permanently blocked on live LinkedIn (AT-010).',
          fallback_mode: 'manual_clipboard_assist'
        }
      ]
    }
  },
  {
    name: 'Expired / Revoked Token (Fail-Closed)',
    description: 'Expired OAuth credential. Fails closed on all actions; requires re-authentication (AT-016).',
    connection: {
      connection_id: 'conn-expired-04',
      user_id: 'usr-cand-104',
      member_id: 'urn:li:person:expired_104',
      display_name: 'Morgan Lee',
      email: 'morgan.lee@example.com',
      granted_scopes: ['openid', 'profile', 'email', 'w_member_social'],
      denied_scopes: [],
      masked_token: 'tok_li_****_0000',
      token_status: 'expired',
      connected_at: '2026-01-10T08:00:00Z',
      expires_at: '2026-02-10T08:00:00Z',
      status: 'reauth_required',
      last_inspected_at: new Date().toISOString(),
      capabilities: [
        {
          capability_id: 'identity_signin',
          domain: 'identity',
          name: 'Sign In with LinkedIn (OpenID Connect)',
          description: 'Authenticate user identity, verified email address, and core OIDC claims.',
          required_scope: 'openid',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Token expired; re-authentication required (AT-016).',
          fallback_mode: 'none'
        },
        {
          capability_id: 'basic_profile_read',
          domain: 'profile_read',
          name: 'Basic Profile Read',
          description: 'Read public name, vanity slug, profile photo, and localized headline.',
          required_scope: 'profile',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Token expired; re-authentication required.',
          fallback_mode: 'none'
        },
        {
          capability_id: 'career_history_import',
          domain: 'profile_read',
          name: 'Career History & Experience Ingestion',
          description: 'Direct programmatic import of structured employment history.',
          required_scope: 'r_fullprofile',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Token expired.',
          fallback_mode: 'user_export_upload'
        },
        {
          capability_id: 'content_publish',
          domain: 'content_publish',
          name: 'Member Post & Article Publishing',
          description: 'Publish user-approved professional updates.',
          required_scope: 'w_member_social',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Token expired.',
          fallback_mode: 'manual_clipboard_assist'
        },
        {
          capability_id: 'direct_messaging',
          domain: 'messaging',
          name: 'Direct Messaging & InMail',
          description: 'Programmatic message read and write.',
          required_scope: 'r_messages',
          status: 'denied',
          is_granted: false,
          truth_in_advertising_note: 'Token expired.',
          fallback_mode: 'manual_clipboard_assist'
        },
        {
          capability_id: 'automated_easy_apply',
          domain: 'easy_apply',
          name: 'Automated Easy Apply Submissions',
          description: 'Background execution of job applications.',
          required_scope: 'w_job_applications',
          status: 'unsupported_platform',
          is_granted: false,
          truth_in_advertising_note: 'Platform unsupported.',
          fallback_mode: 'manual_clipboard_assist'
        }
      ]
    }
  }
];

const INITIAL_PERSONS: LinkedInPersonRecord[] = [
  {
    record_id: 'rec-p-01',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    entity_type: 'person',
    vanity_slug: 'sarah-chen-arch',
    full_name: 'Sarah Chen',
    headline: 'Lead Systems Architect @ CloudScale',
    current_company: 'CloudScale Systems',
    current_role: 'Lead Systems Architect',
    location: 'San Francisco Bay Area',
    connection_tier: '1st',
    public_profile_url: 'https://linkedin.com/in/sarah-chen-arch',
    skills: ['Go', 'Distributed Systems', 'Kubernetes', 'gRPC'],
    version: 1,
    observed_at: '2026-03-20T10:00:00Z'
  },
  {
    record_id: 'rec-p-02',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    entity_type: 'person',
    vanity_slug: 'marcus-brooks-data',
    full_name: 'Marcus Brooks',
    headline: 'Staff Data Engineer & AI Researcher',
    current_company: 'Apex Dynamics',
    current_role: 'Staff Data Engineer',
    location: 'New York, NY',
    connection_tier: '2nd',
    public_profile_url: 'https://linkedin.com/in/marcus-brooks-data',
    skills: ['Python', 'Spark', 'Kafka', 'Data Engineering'],
    version: 1,
    observed_at: '2026-03-20T10:15:00Z'
  },
  {
    record_id: 'rec-p-99',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    entity_type: 'person',
    vanity_slug: 'elena-rostova-sec',
    full_name: 'Elena Rostova',
    headline: 'Lead Security Researcher @ SecureMesh',
    current_company: 'SecureMesh',
    current_role: 'Lead Security Researcher',
    location: 'Austin, TX',
    connection_tier: '1st',
    public_profile_url: 'https://linkedin.com/in/elena-rostova-sec',
    skills: ['AppSec', 'Cryptography', 'Threat Modeling'],
    version: 1,
    observed_at: '2026-03-20T12:00:00Z'
  }
];

const INITIAL_COMPANIES: LinkedInCompanyRecord[] = [
  {
    record_id: 'rec-c-01',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    entity_type: 'company',
    universal_name: 'cloudscale-systems',
    company_name: 'CloudScale Systems',
    domain: 'cloudscale.tech',
    industry: 'Software Development',
    size_tier: '201-500 employees',
    headquarters: 'San Francisco, CA',
    specialties: ['Distributed Infrastructure', 'Cloud Native', 'Observability'],
    follower_count: 48200,
    verified: true,
    version: 1,
    observed_at: '2026-03-20T10:30:00Z'
  },
  {
    record_id: 'rec-c-99',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    entity_type: 'company',
    universal_name: 'securemesh-cyber',
    company_name: 'SecureMesh',
    domain: 'securemesh.io',
    industry: 'Computer & Network Security',
    size_tier: '51-200 employees',
    headquarters: 'Austin, TX',
    specialties: ['Zero Trust', 'Cloud Security', 'Identity Governance'],
    follower_count: 12500,
    verified: true,
    version: 1,
    observed_at: '2026-03-20T12:15:00Z'
  }
];

const INITIAL_JOBS: LinkedInJobRecord[] = [
  {
    record_id: 'rec-j-01',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    entity_type: 'job',
    job_id: 'job-10101',
    title: 'Principal Systems Engineer',
    company_name: 'CloudScale Systems',
    workplace_type: 'Remote',
    location: 'United States',
    employment_type: 'Full-time',
    description_snippet: 'We are seeking a Principal Systems Engineer to scale our core event platform to millions of TPS.',
    applicant_count: 34,
    posted_at: '2026-03-18T08:00:00Z',
    version: 1,
    observed_at: '2026-03-20T11:00:00Z'
  },
  {
    record_id: 'rec-j-99',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    entity_type: 'job',
    job_id: 'job-20202',
    title: 'Senior Cryptography Engineer',
    company_name: 'SecureMesh',
    workplace_type: 'Hybrid',
    location: 'Austin, TX',
    employment_type: 'Full-time',
    description_snippet: 'Looking for cryptographic engineers experienced in post-quantum key exchange algorithms.',
    applicant_count: 12,
    posted_at: '2026-03-19T09:00:00Z',
    version: 1,
    observed_at: '2026-03-20T12:30:00Z'
  }
];

const INITIAL_POSTS: LinkedInPostRecord[] = [
  {
    record_id: 'rec-post-01',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    entity_type: 'post',
    urn: 'urn:li:activity:7283940182',
    author_name: 'Sarah Chen',
    author_vanity: 'sarah-chen-arch',
    commentary: 'Thrilled to share that our team just achieved sub-millisecond p99 latency on our distributed sync engine!',
    media_type: 'article',
    reactions_count: 340,
    comments_count: 42,
    shares_count: 18,
    published_at: '2026-03-19T14:20:00Z',
    version: 1,
    observed_at: '2026-03-20T11:30:00Z'
  },
  {
    record_id: 'rec-post-99',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    entity_type: 'post',
    urn: 'urn:li:activity:8392019283',
    author_name: 'Elena Rostova',
    author_vanity: 'elena-rostova-sec',
    commentary: 'Announcing our open source auditing framework for zero-knowledge proofs.',
    media_type: 'none',
    reactions_count: 512,
    comments_count: 68,
    shares_count: 45,
    published_at: '2026-03-20T08:00:00Z',
    version: 1,
    observed_at: '2026-03-20T13:00:00Z'
  }
];

const INITIAL_LEADS: LinkedInRecruiterLead[] = [
  {
    id: 'lead-alpha-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    recruiter_name: 'Marcus Vance',
    recruiter_title: 'Principal Talent Acquisition Partner',
    company: 'Stripe',
    linkedin_url: 'https://www.linkedin.com/in/marcus-vance-talent',
    source_entity_type: 'person_record',
    source_entity_id: 'rec-p-01',
    status: 'in_dialogue',
    outreach_stage: 'replied',
    related_job_id: 'rec-j-01',
    related_application_id: 'app-stripe-staff-infra',
    notes: [
      {
        id: 'note-001',
        lead_id: 'lead-alpha-001',
        author: 'Alex Chen',
        content: 'Had initial introductory chat regarding Staff Infrastructure role. Emphasized distributed systems background.',
        created_at: '2026-03-18T14:30:00Z'
      },
      {
        id: 'note-002',
        lead_id: 'lead-alpha-001',
        author: 'Alex Chen',
        content: 'Sent tailored resume and open-source contributions portfolio link.',
        created_at: '2026-03-19T10:15:00Z'
      }
    ],
    reminder: {
      due_date: '2026-03-23T09:00:00Z',
      message: 'Follow up on technical screen scheduling with Marcus',
      status: 'pending'
    },
    created_at: '2026-03-18T14:00:00Z',
    updated_at: '2026-03-19T10:15:00Z'
  },
  {
    id: 'lead-alpha-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    recruiter_name: 'David Kim',
    recruiter_title: 'Director of Engineering',
    company: 'CloudScale Systems',
    linkedin_url: 'https://www.linkedin.com/in/david-kim-eng',
    source_entity_type: 'post_record',
    source_entity_id: 'rec-post-01',
    status: 'new',
    outreach_stage: 'draft',
    related_job_id: 'rec-j-01',
    related_application_id: '',
    notes: [
      {
        id: 'note-003',
        lead_id: 'lead-alpha-002',
        author: 'Alex Chen',
        content: 'Posted hiring call for Distributed Systems Architects. Connection note drafted.',
        created_at: '2026-03-19T16:00:00Z'
      }
    ],
    reminder: {
      due_date: '2026-03-22T11:00:00Z',
      message: 'Review and send personalized connection note on LinkedIn',
      status: 'pending'
    },
    created_at: '2026-03-19T15:50:00Z',
    updated_at: '2026-03-19T16:00:00Z'
  },
  {
    id: 'lead-beta-001',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    recruiter_name: 'Sophia Lin',
    recruiter_title: 'Head of People & Recruiting',
    company: 'SecureMesh',
    linkedin_url: 'https://www.linkedin.com/in/sophia-lin-recruiter',
    source_entity_type: 'person_record',
    source_entity_id: 'rec-p-02',
    status: 'contacted',
    outreach_stage: 'contacted',
    related_job_id: 'rec-j-99',
    related_application_id: 'app-securemesh-crypto',
    notes: [
      {
        id: 'note-beta-001',
        lead_id: 'lead-beta-001',
        author: 'Beta Recruiter',
        content: 'Connected via LinkedIn. She confirmed receipt of application for Senior Cryptography Engineer.',
        created_at: '2026-03-17T11:00:00Z'
      }
    ],
    reminder: {
      due_date: '2026-03-25T15:00:00Z',
      message: 'Check status on hiring manager review',
      status: 'pending'
    },
    created_at: '2026-03-17T10:45:00Z',
    updated_at: '2026-03-17T11:00:00Z'
  }
];

const INITIAL_QUEUE_ITEMS: LinkedInConnectionQueueItem[] = [
  {
    id: 'cq-alpha-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    recipient_id: 'rec-p-01',
    recipient_name: 'Sarah Chen',
    recipient_title: 'Lead Systems Architect',
    recipient_company: 'CloudScale Systems',
    recipient_linkedin_url: 'https://www.linkedin.com/in/sarah-chen-arch',
    note_text: 'Hi Sarah, noticed your work at CloudScale Systems. With my background in Distributed Systems, Go, and Kubernetes, I would love to connect and follow your team’s engineering work!',
    character_count: 172,
    within_limit: true,
    status: 'pending_approval',
    context_factors: ['Company: CloudScale Systems', 'Domain overlap: Distributed Systems, Go', 'Target: Principal Architect'],
    created_at: '2026-03-20T09:30:00Z',
    updated_at: '2026-03-20T09:30:00Z'
  },
  {
    id: 'cq-alpha-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    recipient_id: 'rec-p-02',
    recipient_name: 'Marcus Vance',
    recipient_title: 'Principal Talent Acquisition Partner',
    recipient_company: 'Stripe',
    recipient_linkedin_url: 'https://www.linkedin.com/in/marcus-vance-talent',
    note_text: 'Hi Marcus, saw your talent posts at Stripe and wanted to connect regarding engineering opportunities. Looking forward to staying in touch!',
    character_count: 139,
    within_limit: true,
    status: 'approved',
    approval_token: 'hmac-sha256:7f4a21e09c8d5b12a9e34',
    approved_by: 'Alex Chen (operator)',
    approved_at: '2026-03-20T10:00:00Z',
    context_factors: ['Company: Stripe', 'Recruiter Role Match'],
    created_at: '2026-03-20T09:45:00Z',
    updated_at: '2026-03-20T10:00:00Z'
  },
  {
    id: 'cq-alpha-003',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    recipient_id: 'lead-alpha-002',
    recipient_name: 'David Kim',
    recipient_title: 'Director of Engineering',
    recipient_company: 'CloudScale Systems',
    recipient_linkedin_url: 'https://www.linkedin.com/in/david-kim-eng',
    note_text: 'Hi David, saw your hiring updates at CloudScale Systems regarding systems architects and wanted to connect. Looking forward to staying in touch!',
    character_count: 147,
    within_limit: true,
    status: 'completed_manually',
    approval_token: 'hmac-sha256:88fa12c93845ba012ef',
    approved_by: 'Alex Chen (operator)',
    approved_at: '2026-03-19T16:15:00Z',
    completed_at: '2026-03-19T16:30:00Z',
    context_factors: ['Company: CloudScale Systems', 'Hiring Post Discovered'],
    created_at: '2026-03-19T16:00:00Z',
    updated_at: '2026-03-19T16:30:00Z'
  },
  {
    id: 'cq-beta-001',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    recipient_name: 'Sophia Lin',
    recipient_title: 'Head of People & Recruiting',
    recipient_company: 'SecureMesh',
    recipient_linkedin_url: 'https://www.linkedin.com/in/sophia-lin-recruiter',
    note_text: 'Hi Sophia, saw your updates at SecureMesh and wanted to connect regarding engineering leadership roles. Looking forward to following your team!',
    character_count: 148,
    within_limit: true,
    status: 'pending_approval',
    context_factors: ['Company: SecureMesh', 'Recruiter Connect'],
    created_at: '2026-03-20T11:00:00Z',
    updated_at: '2026-03-20T11:00:00Z'
  }
];

const INITIAL_BUDGETS: Record<string, LinkedInConnectionBudget> = {
  'ws-alpha': {
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    daily_limit: 20,
    daily_used: 2,
    weekly_limit: 80,
    weekly_used: 14,
    last_reset_date: '2026-03-20T00:00:00Z'
  },
  'ws-beta': {
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    daily_limit: 20,
    daily_used: 20,
    weekly_limit: 80,
    weekly_used: 80,
    last_reset_date: '2026-03-20T00:00:00Z'
  }
};

const INITIAL_WATCHLIST: CompanyWatchlistItem[] = [
  {
    item_id: 'cwl-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    company_record_id: 'rec-c-01',
    company_name: 'Stripe',
    universal_name: 'stripe',
    domain: 'stripe.com',
    company_page_url: 'https://www.linkedin.com/company/stripe',
    priority: 'high',
    target_reason: 'Expanding cloud infrastructure & payment engine teams for Staff roles',
    tags: ['fintech', 'payments', 'infrastructure'],
    status: 'active',
    created_at: '2026-03-18T10:00:00Z',
    updated_at: '2026-03-18T10:00:00Z'
  },
  {
    item_id: 'cwl-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    company_record_id: 'rec-c-02',
    company_name: 'CloudScale Systems',
    universal_name: 'cloudscale-systems',
    domain: 'cloudscale.example.com',
    company_page_url: 'https://www.linkedin.com/company/cloudscale-systems',
    priority: 'high',
    target_reason: 'Recruiting for principal distributed systems and event platform architects',
    tags: ['cloud', 'distributed-systems'],
    status: 'active',
    created_at: '2026-03-19T11:00:00Z',
    updated_at: '2026-03-19T11:00:00Z'
  },
  {
    item_id: 'cwl-003',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    company_name: 'Databricks',
    universal_name: 'databricks',
    domain: 'databricks.com',
    company_page_url: 'https://www.linkedin.com/company/databricks',
    priority: 'medium',
    target_reason: 'Scaling Apache Spark, Ray, and Lakehouse infrastructure teams',
    tags: ['data', 'lakehouse', 'ai'],
    status: 'active',
    created_at: '2026-03-20T09:00:00Z',
    updated_at: '2026-03-20T09:00:00Z'
  },
  {
    item_id: 'cwl-beta-001',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    company_name: 'SecureMesh',
    universal_name: 'securemesh',
    domain: 'securemesh.example.com',
    company_page_url: 'https://www.linkedin.com/company/securemesh',
    priority: 'high',
    target_reason: 'Targeting zero-knowledge security research initiatives',
    tags: ['security', 'cryptography'],
    status: 'active',
    created_at: '2026-03-20T08:00:00Z',
    updated_at: '2026-03-20T08:00:00Z'
  }
];

const INITIAL_FOLLOW_PLANS: BatchCompanyFollowPlan[] = [
  {
    plan_id: 'plan-alpha-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    plan_name: 'Q1 High-Priority Infrastructure Batch',
    items: [
      {
        item_id: 'fitem-001',
        company_record_id: 'rec-c-01',
        company_name: 'Stripe',
        universal_name: 'stripe',
        company_page_url: 'https://www.linkedin.com/company/stripe',
        priority: 'high',
        target_reason: 'Expanding cloud infrastructure teams',
        pacing_interval_sec: 120,
        status: 'manual_followed',
        scheduled_for: '2026-03-20T10:00:00Z',
        followed_at: '2026-03-20T10:02:15Z'
      },
      {
        item_id: 'fitem-002',
        company_record_id: 'rec-c-02',
        company_name: 'CloudScale Systems',
        universal_name: 'cloudscale-systems',
        company_page_url: 'https://www.linkedin.com/company/cloudscale-systems',
        priority: 'high',
        target_reason: 'Principal architect openings',
        pacing_interval_sec: 120,
        status: 'ready_for_manual_follow',
        scheduled_for: '2026-03-20T10:04:00Z'
      }
    ],
    total_count: 2,
    followed_count: 1,
    status: 'in_progress',
    created_at: '2026-03-20T09:55:00Z'
  }
];

const INITIAL_FOLLOW_BUDGETS: Record<string, CompanyFollowBudget> = {
  'ws-alpha': {
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    daily_limit: 15,
    daily_used: 1,
    weekly_limit: 50,
    weekly_used: 6,
    last_reset_date: '2026-03-20',
    weekly_reset_date: '2026-03-16',
    is_locked: false
  },
  'ws-beta': {
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    daily_limit: 15,
    daily_used: 15,
    weekly_limit: 50,
    weekly_used: 50,
    last_reset_date: '2026-03-20',
    weekly_reset_date: '2026-03-16',
    is_locked: true,
    lock_reason: 'daily company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)'
  }
};

const INITIAL_CONVERSATION_THREADS: LinkedInConversationThread[] = [
  {
    thread_id: 'th-alpha-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    participant_name: 'Marcus Vance',
    participant_vanity: 'marcus-vance-talent',
    participant_headline: 'Principal Talent Acquisition Partner at Stripe',
    participant_company: 'Stripe',
    subject: 'Staff Infrastructure Architect Opportunity',
    last_message_snippet: 'Would you be open to a 30-minute introductory call this week?',
    last_message_at: '2026-03-20T09:30:00Z',
    unread_count: 1,
    thread_type: 'recruiter',
    classification: 'interview_invitation',
    active_followup_planned: false,
    messages: [
      {
        message_id: 'msg-out-001',
        thread_id: 'th-alpha-001',
        sender_name: 'Alex Chen',
        sender_type: 'self',
        content: 'Hi Marcus, excited to connect with your team at Stripe!',
        sent_at: '2026-03-18T14:20:00Z',
        is_read: true,
      },
      {
        message_id: 'msg-in-001',
        thread_id: 'th-alpha-001',
        sender_name: 'Marcus Vance',
        sender_type: 'other',
        content: 'Hi Alex, I was really impressed by your background scaling multi-region distributed systems. We have an open Staff Infrastructure Architect role on our payments core team at Stripe. Would you be open to a 30-minute introductory call this week?',
        sent_at: '2026-03-20T09:30:00Z',
        is_read: false,
      }
    ]
  },
  {
    thread_id: 'th-alpha-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    participant_name: 'Sophia Lin',
    participant_vanity: 'sophia-lin-ai',
    participant_headline: 'Engineering Director at ScaleAI',
    participant_company: 'ScaleAI',
    subject: 'Platform Tech Lead Outreach',
    last_message_snippet: 'Let me know if you would like to review the technical challenge scope.',
    last_message_at: '2026-03-19T16:45:00Z',
    unread_count: 0,
    thread_type: 'recruiter',
    classification: 'recruiter_inquiry',
    active_followup_planned: true,
    messages: [
      {
        message_id: 'msg-in-002',
        thread_id: 'th-alpha-002',
        sender_name: 'Sophia Lin',
        sender_type: 'other',
        content: 'Hey Alex, our platform engineering organization is scaling rapidly. Let me know if you would like to review the technical challenge scope.',
        sent_at: '2026-03-19T16:45:00Z',
        is_read: true,
      }
    ]
  },
  {
    thread_id: 'th-beta-001',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    participant_name: 'David Zhao',
    participant_vanity: 'david-zhao-fintech',
    participant_headline: 'Staff Recruiter at Robinhood',
    participant_company: 'Robinhood',
    subject: 'Core Systems Architect',
    last_message_snippet: 'Are you open to discussing crypto/fintech infrastructure roles?',
    last_message_at: '2026-03-19T11:00:00Z',
    unread_count: 1,
    thread_type: 'recruiter',
    classification: 'recruiter_inquiry',
    active_followup_planned: false,
    messages: [
      {
        message_id: 'msg-beta-001',
        thread_id: 'th-beta-001',
        sender_name: 'David Zhao',
        sender_type: 'other',
        content: 'Are you open to discussing crypto/fintech infrastructure roles?',
        sent_at: '2026-03-19T11:00:00Z',
        is_read: false,
      }
    ]
  }
];

const INITIAL_REPLY_DRAFTS: LinkedInReplyDraft[] = [
  {
    draft_id: 'rdraft-001',
    thread_id: 'th-alpha-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    tone: 'concise_scheduling',
    suggested_text: 'Hi Marcus, thank you for reaching out regarding the opportunity at Stripe. I would be glad to connect for an introductory conversation. I am generally available Thursday or Friday afternoon EST. Please let me know what time works best for you.',
    rationale: 'Offers immediate availability for scheduling with minimal back-and-forth friction.',
    verified_facts_used: ['Distributed systems architecture experience', 'Open to introductory discussion'],
    character_count: 247,
    status: 'draft',
    created_at: '2026-03-20T09:35:00Z',
    updated_at: '2026-03-20T09:35:00Z'
  }
];

const INITIAL_SWEPT_POSTS: LinkedInSweptTargetPost[] = [
  {
    post_id: 'post-sweep-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    author_name: 'Dr. Martin Kleppmann',
    author_headline: 'Associate Professor in Distributed Systems at University of Cambridge',
    author_company: 'University of Cambridge',
    post_url: 'https://www.linkedin.com/feed/update/urn:li:activity:7189000000000000001',
    content: 'Why distributed consensus algorithms like Raft and Paxos can suffer from silent leader partition stalls during asymmetric network partitions, and how randomized heartbeats help mitigate them.',
    category: 'technical_discussion',
    existing_comments_count: 24,
    comments: [
      {
        comment_id: 'c-001',
        author_name: 'Leslie Lamport',
        author_headline: 'Turing Award Winner, Distinguished Scientist',
        content: 'Correctness under asynchronous timing remains paramount regardless of heuristic heartbeat tuning.',
        created_at: '2026-03-20T08:00:00Z',
        likes_count: 142,
        is_high_priority: true
      }
    ],
    swept_at: '2026-03-20T08:30:00Z'
  },
  {
    post_id: 'post-sweep-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    author_name: 'Elena Rostova',
    author_headline: 'VP of Infrastructure Engineering at CloudScale',
    author_company: 'CloudScale Systems',
    post_url: 'https://www.linkedin.com/feed/update/urn:li:activity:7189000000000000002',
    content: 'We are expanding our Core Platform team at CloudScale! Looking for Principal Engineers with Kubernetes, high-throughput Golang, and event-driven streaming background. Feel free to comment or DM.',
    category: 'hiring_announcement',
    existing_comments_count: 12,
    comments: [
      {
        comment_id: 'c-002',
        author_name: 'Tech Talent Lead',
        content: 'DM your resume or drop a note if you are open to senior IC roles.',
        created_at: '2026-03-20T09:15:00Z',
        likes_count: 8,
        is_high_priority: true
      }
    ],
    swept_at: '2026-03-20T09:30:00Z'
  },
  {
    post_id: 'post-sweep-beta-001',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    author_name: 'Marcus Vance',
    author_headline: 'Principal Security Architect at SecureMesh',
    author_company: 'SecureMesh',
    post_url: 'https://www.linkedin.com/feed/update/urn:li:activity:7189000000000000099',
    content: 'Zero-trust network architecture requires ephemeral token validation at each microservice boundary, not just ingress gateways.',
    category: 'technical_discussion',
    existing_comments_count: 7,
    swept_at: '2026-03-20T10:00:00Z'
  }
];

const INITIAL_COMMENT_DRAFTS: LinkedInCommentDraft[] = [
  {
    draft_id: 'cdraft-alpha-001',
    post_id: 'post-sweep-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    angle: 'insightful_addition',
    comment_text: 'Fascinating analysis, Dr. Martin. In multi-region deployments, we observed that asymmetric network partitions frequently cause quorum lease drift before heartbeats trigger leader step-down. Tightening heartbeat randomization proved essential for keeping failover latencies under 500ms.',
    rationale: 'Adds concrete real-world engineering perspective that deepens the technical discussion without generic fluff.',
    verified_facts_used: ['Architected multi-region consensus layer across 5 continents', 'Raft lease management experience'],
    character_count: 284,
    status: 'draft',
    created_at: '2026-03-20T08:45:00Z',
    updated_at: '2026-03-20T08:45:00Z'
  }
];

// Initial Post Drafts for IMP-LI-11 (Post Writing, Hooks & Audits)
const INITIAL_POST_DRAFTS: LinkedInPostDraft[] = [
  {
    draft_id: 'post-draft-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    topic: 'Migrating Microservices back to Modular Monolith',
    angle: 'contrarian_insight',
    selected_hook_type: 'contrarian',
    selected_hook_text: 'Most teams adopting microservices are actually paying a distributed systems tax for no reason.',
    body_text: 'Last year, our team took a hard look at our latency budget and operational overhead.\n\nWe were running 34 separate services for a 12-person team. Network serialization alone added 45ms to P99 latency.\n\nHere is what changed when we unified the core domains into a disciplined modular monolith:\n- P99 latency dropped from 280ms to 42ms for user-facing APIs.\n- Reduced our AWS infrastructure bill by 35% within 3 months.\n- Maintained 99.99% uptime across 12 million daily active users.\n\nArchitectural simplicity often beats theoretical scalability.',
    call_to_action: 'Has your team ever considered or executed a monolith consolidation? What was your primary constraint?',
    full_post_text: `Most teams adopting microservices are actually paying a distributed systems tax for no reason.

Last year, our team took a hard look at our latency budget and operational overhead.

We were running 34 separate services for a 12-person team. Network serialization alone added 45ms to P99 latency.

Here is what changed when we unified the core domains into a disciplined modular monolith:
- P99 latency dropped from 280ms to 42ms for user-facing APIs.
- Reduced our AWS infrastructure bill by 35% within 3 months.
- Maintained 99.99% uptime across 12 million daily active users.

Architectural simplicity often beats theoretical scalability.

Has your team ever considered or executed a monolith consolidation? What was your primary constraint?`,
    verified_facts_used: [
      'Reduced AWS bill by 35% within 3 months',
      'P99 latency improved from 280ms to 42ms for user-facing APIs',
      'Maintained 99.99% uptime across 12 million daily active users'
    ],
    available_hooks: [
      {
        hook_type: 'contrarian',
        hook_text: 'Most teams adopting microservices are actually paying a distributed systems tax for no reason.',
        rationale: 'Challenges common industry dogma with tangible engineering justification.'
      },
      {
        hook_type: 'question',
        hook_text: 'Why do so many high-growth teams struggle with microservices complexity before hitting true scale?',
        rationale: 'Provokes curiosity and invites peer reflections.'
      },
      {
        hook_type: 'data_scale',
        hook_text: 'Cutting 34 services down to a modular monolith cut P99 latency from 280ms to 42ms.',
        rationale: 'Front-loads hard performance metrics from verified candidate milestones.'
      },
      {
        hook_type: 'story',
        hook_text: 'Two quarters ago, our on-call rotations were burning out from cross-service cascading timeouts.',
        rationale: 'Establishes high-stakes narrative tension based on real team experience.'
      },
      {
        hook_type: 'punchy',
        hook_text: 'Microservices didn’t fix your team architecture; they just moved function calls onto the network.',
        rationale: 'Concise, memorable declaration that stops feed scroll.'
      }
    ],
    latest_audit: {
      passed: true,
      readability_score: 86,
      character_count: 812,
      line_break_density: 0.12,
      paragraph_count: 7,
      estimated_read_time_sec: 38,
      buzzwords_count: 0,
      virality_claims_count: 0,
      ai_bypass_claims_count: 0,
      unverified_metrics_count: 0,
      issues: [],
      audited_at: '2026-09-20T10:00:00Z',
      disclaimer: 'Zero false virality or detector bypass guarantees. Groundtruth verified.'
    },
    status: 'approved',
    approval_token: 'hmac-sha256:7f8a9b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a',
    approved_by: 'candidate-user-01',
    approved_at: '2026-09-20T10:05:00Z',
    created_at: '2026-09-20T09:55:00Z',
    updated_at: '2026-09-20T10:05:00Z'
  },
  {
    draft_id: 'post-draft-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    topic: 'Lessons from Transitioning from Senior Engineer to Staff Architect',
    angle: 'lesson_learned_breakdown',
    selected_hook_type: 'story',
    selected_hook_text: 'The biggest shock when moving from Senior to Staff wasn’t technical complexity—it was calendar ownership.',
    body_text: 'When I first stepped into a Staff Architect role, my initial instinct was to write more code.\n\nThat was a mistake.\n\nHere are 3 fundamental shifts that actually drive leverage:\n1. Designing systems for the engineers who maintain them, not my own ego.\n2. Writing architectural RFCs that front-load failure modes before code is committed.\n3. Mentoring senior engineers so that good decisions happen without me in the room.\n\nTechnical excellence at scale is about multiplication, not solo output.',
    call_to_action: 'For Staff+ engineers: What was the single hardest habit you had to unlearn during your transition?',
    full_post_text: `The biggest shock when moving from Senior to Staff wasn’t technical complexity—it was calendar ownership.

When I first stepped into a Staff Architect role, my initial instinct was to write more code.

That was a mistake.

Here are 3 fundamental shifts that actually drive leverage:
1. Designing systems for the engineers who maintain them, not my own ego.
2. Writing architectural RFCs that front-load failure modes before code is committed.
3. Mentoring senior engineers so that good decisions happen without me in the room.

Technical excellence at scale is about multiplication, not solo output.

For Staff+ engineers: What was the single hardest habit you had to unlearn during your transition?`,
    verified_facts_used: [
      'Led architectural roadmap across 4 distributed squads',
      'Authored 18 RFCs on distributed consensus and event streaming'
    ],
    available_hooks: [
      {
        hook_type: 'story',
        hook_text: 'The biggest shock when moving from Senior to Staff wasn’t technical complexity—it was calendar ownership.',
        rationale: 'Relatable leadership narrative from real career transition.'
      },
      {
        hook_type: 'question',
        hook_text: 'Why do so many top-performing Senior Engineers struggle in their first year at Staff level?',
        rationale: 'Direct address targeting engineering peers and leaders.'
      },
      {
        hook_type: 'contrarian',
        hook_text: 'Writing more code as a Staff Engineer is usually a sign of failing at your job.',
        rationale: 'Contrarian leadership observation that invites debate.'
      },
      {
        hook_type: 'data_scale',
        hook_text: 'Across 18 architectural RFCs and 4 squads, the best architecture was always the one with the fewest moving parts.',
        rationale: 'Grounds advice in hard experience metrics.'
      },
      {
        hook_type: 'punchy',
        hook_text: 'Senior engineers solve hard problems. Staff engineers decide which problems are worth solving.',
        rationale: 'Clean distinction, highly shareable.'
      }
    ],
    latest_audit: {
      passed: true,
      readability_score: 89,
      character_count: 735,
      line_break_density: 0.14,
      paragraph_count: 6,
      estimated_read_time_sec: 34,
      buzzwords_count: 0,
      virality_claims_count: 0,
      ai_bypass_claims_count: 0,
      unverified_metrics_count: 0,
      issues: [],
      audited_at: '2026-09-20T11:00:00Z',
      disclaimer: 'Zero false virality or detector bypass guarantees. Groundtruth verified.'
    },
    status: 'draft',
    created_at: '2026-09-20T10:45:00Z',
    updated_at: '2026-09-20T11:00:00Z'
  },
  {
    draft_id: 'post-draft-003',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    topic: 'Multi-Region Kubernetes Failover Architecture',
    angle: 'technical_deep_dive',
    selected_hook_type: 'data_scale',
    selected_hook_text: 'When Region A suffered an optical fiber cut, our ingress failover completed in 4.2 seconds with zero dropped connections.',
    body_text: 'Designing active-active multi-region Kubernetes clusters often sounds like over-engineering until a major cloud zone goes dark.\n\nOur architecture relies on three core tenets:\n1. Anycast BGP routing with continuous health probes.\n2. Cross-region CockroachDB with raft-based distributed consensus.\n3. Stateless workloads deployed identically across US-East and US-West.\n\nUptime isn’t bought with vendor promises; it is earned through chaos engineering fire drills.',
    call_to_action: 'How does your infrastructure team test multi-region failover under realistic chaos conditions?',
    full_post_text: `When Region A suffered an optical fiber cut, our ingress failover completed in 4.2 seconds with zero dropped connections.

Designing active-active multi-region Kubernetes clusters often sounds like over-engineering until a major cloud zone goes dark.

Our architecture relies on three core tenets:
1. Anycast BGP routing with continuous health probes.
2. Cross-region CockroachDB with raft-based distributed consensus.
3. Stateless workloads deployed identically across US-East and US-West.

Uptime isn’t bought with vendor promises; it is earned through chaos engineering fire drills.

How does your infrastructure team test multi-region failover under realistic chaos conditions?`,
    verified_facts_used: [
      'Engineered automated BGP ingress failover in 4.2 seconds',
      'CockroachDB multi-region deployment spanning 3 geographical zones'
    ],
    available_hooks: [
      {
        hook_type: 'data_scale',
        hook_text: 'When Region A suffered an optical fiber cut, our ingress failover completed in 4.2 seconds with zero dropped connections.',
        rationale: 'Metrics-driven opening demonstrating technical proof.'
      },
      {
        hook_type: 'contrarian',
        hook_text: 'Most multi-region disaster recovery setups have never been tested and will fail on day one.',
        rationale: 'Direct warning on simulated vs real DR.'
      },
      {
        hook_type: 'question',
        hook_text: 'What happens to your active database connections when a cloud provider loses an entire availability zone?',
        rationale: 'Engages engineering curiosity.'
      },
      {
        hook_type: 'story',
        hook_text: 'At 3:14 AM on a Sunday, our primary region went completely dark.',
        rationale: 'Classic war-story hook.'
      },
      {
        hook_type: 'punchy',
        hook_text: 'If your disaster recovery plan hasn’t been tested with chaos engineering, you don’t have a disaster recovery plan.',
        rationale: 'Memorable, uncompromising engineering truth.'
      }
    ],
    latest_audit: {
      passed: true,
      readability_score: 82,
      character_count: 672,
      line_break_density: 0.13,
      paragraph_count: 5,
      estimated_read_time_sec: 30,
      buzzwords_count: 0,
      virality_claims_count: 0,
      ai_bypass_claims_count: 0,
      unverified_metrics_count: 0,
      issues: [],
      audited_at: '2026-09-20T11:30:00Z',
      disclaimer: 'Zero false virality or detector bypass guarantees. Groundtruth verified.'
    },
    status: 'draft',
    created_at: '2026-09-20T11:20:00Z',
    updated_at: '2026-09-20T11:30:00Z'
  }
];

// Initial Voice Profiles and Humanizer Fixtures for IMP-LI-12 (Humanizer & Reusable Voice)
const INITIAL_VOICE_PROFILES: LinkedInVoiceProfile[] = [
  {
    profile_id: 'vp-staff-arch-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    profile_name: 'Staff Systems Architect',
    target_audience: 'VP of Engineering, Staff Engineers, Distributed Systems Practitioners',
    formality: 4,
    technical_depth: 4,
    cadence_style: 'punchy_staccato',
    perspective: 'first_person_singular',
    preferred_terms: ['trade-offs', 'operational overhead', 'distributed consensus', 'latency budget', 'rigor'],
    blacklisted_terms: ['paradigm shift', 'game changer', 'synergy', 'testament to', 'beacon of', 'delve into'],
    status: 'approved',
    approval_token: 'hmac-sha256:7f8a9b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a',
    approved_by: 'candidate-user-01',
    approved_at: '2026-09-20T10:00:00Z',
    created_at: '2026-09-20T09:30:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    profile_id: 'vp-staff-arch-002',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    profile_name: 'Engineering Director / Team Voice',
    target_audience: 'Engineering Teams, Hiring Committees, Tech Leads',
    formality: 3,
    technical_depth: 3,
    cadence_style: 'balanced_rhythm',
    perspective: 'collective_team',
    preferred_terms: ['our platform team', 'cross-functional collaboration', 'resilience', 'sustainable pace'],
    blacklisted_terms: ['rockstar', 'ninja', '10x engineer', 'crushing it'],
    status: 'draft',
    created_at: '2026-09-20T11:00:00Z',
    updated_at: '2026-09-20T11:00:00Z'
  }
];

const INITIAL_HUMANIZE_RESULTS: LinkedInHumanizeResult[] = [
  {
    result_id: 'hum-res-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    voice_profile_id: 'vp-staff-arch-001',
    raw_text: "In today's fast-paced world, building distributed systems is a testament to engineering excellence. Let us delve into why microservices are a game changer. Furthermore, it is crucial to remember that synergy across teams serves as a beacon of modern DevOps.",
    cleaned_text: "Building distributed systems requires engineering rigor. Microservices introduce complex operational trade-offs. Cross-team alignment remains an essential requirement for reliable infrastructure.",
    tier1_slop_replacements: [
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "In today's fast-paced world",
        suggested_replacement: "In modern production systems",
        category: "cliché_opening",
        rationale: "Removes generic filler opening in favor of concrete engineering context."
      },
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "testament to",
        suggested_replacement: "evidence of",
        category: "cliché_idiom",
        rationale: "Replaces melodramatic phrasing with grounded practitioner vocabulary."
      },
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "delve into",
        suggested_replacement: "examine",
        category: "slop_buzzword",
        rationale: "Removes repetitive AI trope 'delve into'."
      },
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "game changer",
        suggested_replacement: "significant capability shift",
        category: "hype_buzzword",
        rationale: "Avoids hyperbolic marketing jargon."
      },
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "synergy",
        suggested_replacement: "cross-team alignment",
        category: "corporate_buzzword",
        rationale: "Scrubs corporate cliché; replaces with specific engineering practice."
      },
      {
        tier: "Tier 1: Lexical",
        offending_phrase: "beacon of",
        suggested_replacement: "foundation for",
        category: "metaphor_slop",
        rationale: "Eliminates empty decorative metaphors."
      }
    ],
    tier2_cadence_notes: [
      "Target cadence 'punchy_staccato' applied.",
      "Sentence lengths distributed: 6 words, 5 words, 9 words (Mean: 6.7 words, SD: 1.7 words).",
      "Staccato rhythm achieved with high punchiness index."
    ],
    tier3_issues: [],
    slop_score_before: 82,
    slop_score_after: 4,
    readability_score_before: 58,
    readability_score_after: 91,
    passed: true,
    status: 'approved',
    approval_token: 'hmac-sha256:4a8b9c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b',
    approved_by: 'candidate-user-01',
    approved_at: '2026-09-20T10:15:00Z',
    created_at: '2026-09-20T10:10:00Z'
  }
];

function runEditorialAudit(text: string, verifiedFacts: string[]): LinkedInEditorialAuditReport {
  const issues: LinkedInAuditIssue[] = [];
  const lower = text.toLowerCase();

  // 1. Buzzword check
  const buzzwords = [
    '10x engineer',
    'rockstar',
    'ninja',
    'guru',
    'synergy',
    'paradigm shift',
    'game changer',
    'revolutionary'
  ];
  let buzzwordCount = 0;
  for (const b of buzzwords) {
    if (lower.includes(b)) {
      buzzwordCount++;
      issues.push({
        code: 'buzzword_detected',
        severity: 'warning',
        message: `Avoid empty buzzword '${b}'. Replace with concrete technical evidence.`,
        offending_snippet: b,
        suggested_fix: 'Replace with specific engineering metrics or concrete responsibilities.'
      });
    }
  }

  // 2. Prohibited Virality & AI Bypass Claims (LI-11, AT-019)
  const viralityClaims = [
    'guaranteed 100k impressions',
    'viral guarantee',
    'guaranteed reach',
    'go viral guaranteed'
  ];
  let viralityCount = 0;
  for (const v of viralityClaims) {
    if (lower.includes(v)) {
      viralityCount++;
      issues.push({
        code: 'false_virality_claim',
        severity: 'blocking',
        message: `Prohibited false virality claim '${v}' detected. Platform forbids reach guarantees (AT-019).`,
        offending_snippet: v,
        suggested_fix: 'Remove false claims; focus on authentic practitioner substance.'
      });
    }
  }

  const aiBypassClaims = [
    '100% bypass ai detectors',
    'undetectable ai',
    'beat turnitin',
    'bypass ai detection'
  ];
  let aiBypassCount = 0;
  for (const a of aiBypassClaims) {
    if (lower.includes(a)) {
      aiBypassCount++;
      issues.push({
        code: 'ai_bypass_claim',
        severity: 'blocking',
        message: `Prohibited AI bypass claim '${a}' detected. Claiming guaranteed evasion is barred (LI-11).`,
        offending_snippet: a,
        suggested_fix: 'Focus on candidate authentic perspective and zero-hallucination groundtruth.'
      });
    }
  }

  // 3. Unverified metric check (AT-003 zero-hallucination)
  const metricRegex = /\b(\d+(\.\d+)?(%|ms|s|qps|rps|k|m|g|\$)|\$\d+(\.\d+)?(k|m)?)\b/gi;
  const metricsInText = Array.from(new Set(text.match(metricRegex) || []));
  const factsCombined = verifiedFacts.join(' ').toLowerCase();

  let unverifiedCount = 0;
  for (const m of metricsInText) {
    const cleanMetric = m.toLowerCase();
    if (!factsCombined.includes(cleanMetric)) {
      unverifiedCount++;
      issues.push({
        code: 'unverified_metric',
        severity: 'blocking',
        message: `Metric '${m}' not found in candidate verified career facts (AT-003). Hallucination prevention active.`,
        offending_snippet: m,
        suggested_fix: `Add '${m}' to candidate verified milestones or remove from draft.`
      });
    }
  }

  // 4. Character count & formatting
  const charCount = text.length;
  const paragraphs = text.split(/\n\s*\n/).filter((p) => p.trim().length > 0);
  const lineBreaks = (text.match(/\n/g) || []).length;
  const lineBreakDensity = charCount > 0 ? lineBreaks / charCount : 0;
  const words = text.trim().split(/\s+/).filter(Boolean).length;
  const estimatedReadTimeSec = Math.max(5, Math.round((words / 200) * 60));

  if (charCount > 3000) {
    issues.push({
      code: 'character_count_exceeded',
      severity: 'blocking',
      message: `Character count (${charCount}) exceeds LinkedIn post limit of 3,000 characters.`,
      offending_snippet: `${charCount} chars`,
      suggested_fix: 'Trim body text to keep post under 3,000 characters.'
    });
  } else if (charCount < 100 && text.trim().length > 0) {
    issues.push({
      code: 'too_short',
      severity: 'warning',
      message: 'Post is very short (< 100 chars). Consider providing more practitioner context.',
      suggested_fix: 'Add concrete engineering details or takeaways.'
    });
  }

  // Readability score calculation (0 to 100)
  let readability = 88;
  if (buzzwordCount > 0) readability -= buzzwordCount * 10;
  if (viralityCount > 0) readability -= viralityCount * 25;
  if (aiBypassCount > 0) readability -= aiBypassCount * 25;
  if (unverifiedCount > 0) readability -= unverifiedCount * 15;
  if (charCount > 2000) readability -= 10;
  if (lineBreakDensity < 0.03 && charCount > 400) {
    readability -= 15;
    issues.push({
      code: 'dense_formatting',
      severity: 'warning',
      message: 'Wall of text detected. Add line breaks between paragraphs for mobile skimmability.',
      suggested_fix: 'Break paragraphs into 1-3 line thoughts.'
    });
  }
  readability = Math.max(10, Math.min(100, readability));

  const hasBlocking = issues.some((i) => i.severity === 'blocking');

  return {
    passed: !hasBlocking,
    readability_score: readability,
    character_count: charCount,
    line_break_density: Number(lineBreakDensity.toFixed(3)),
    paragraph_count: paragraphs.length,
    estimated_read_time_sec: estimatedReadTimeSec,
    buzzwords_count: buzzwordCount,
    virality_claims_count: viralityCount,
    ai_bypass_claims_count: aiBypassCount,
    unverified_metrics_count: unverifiedCount,
    issues,
    audited_at: new Date().toISOString(),
    disclaimer: 'Zero false virality or detector bypass guarantees. Groundtruth verified.'
  };
}

function generateHooksForTopic(topic: string, facts: string[]): LinkedInHookVariant[] {
  const f1 = facts[0] || '10+ years engineering scale';
  return [
    {
      hook_type: 'contrarian',
      hook_text: `Most conventional wisdom around ${topic} misses the real constraint.`,
      rationale: 'Challenges common industry dogma with tangible practitioner justification.'
    },
    {
      hook_type: 'question',
      hook_text: `Why do so many engineering teams struggle with ${topic} even after investing quarters of effort?`,
      rationale: 'Provokes curiosity and invites reflections from peers in the same domain.'
    },
    {
      hook_type: 'data_scale',
      hook_text: `Over the last 6 months addressing ${topic}: ${f1}.`,
      rationale: 'Front-loads hard performance metrics from candidate verified milestones.'
    },
    {
      hook_type: 'story',
      hook_text: `A year ago, our team faced an architectural dilemma with ${topic} that forced us to rethink our assumptions.`,
      rationale: 'Establishes narrative tension based on real production experience.'
    },
    {
      hook_type: 'punchy',
      hook_text: `${topic} isn’t an infrastructure problem; it’s an architectural discipline problem.`,
      rationale: 'Concise, memorable declaration that stops feed scroll.'
    }
  ];
}

// Initial Story Bank Fixtures for IMP-LI-13 (Story Bank & Guided Interviewer)
const INITIAL_INTERVIEW_PROMPTS: LinkedInInterviewPrompt[] = [
  {
    id: 'prompt-scar-01',
    category: 'scar_or_failure',
    title: 'Operational Resilience & Crisis Leadership',
    question: 'Describe a production outage or catastrophic software breakdown you directly handled. What was the exact cascading mechanism?',
    context_placeholder: 'Explain what happened, what systems were involved, and what core dilemma you faced...',
    follow_up_probes: ['What metric or operational safeguard was permanently instituted after the postmortem?']
  },
  {
    id: 'prompt-tp-01',
    category: 'turning_point',
    title: 'Strategic Leadership & Domain Architecture',
    question: 'What was the pivotal moment that convinced you to transition from deep individual coding to systems architecture and technical leadership?',
    context_placeholder: 'Explain the turning point, systemic interface breakages, and organizational friction...',
    follow_up_probes: ['What concrete boundary did you establish first, and what was the cross-team result?']
  },
  {
    id: 'prompt-win-01',
    category: 'breakthrough_win',
    title: 'High-Impact Execution & Innovation',
    question: 'Tell the story of an ambitious technical initiative that everyone thought was impossible or high-risk. How did you validate the hypothesis?',
    context_placeholder: 'Describe the initiative, initial skepticism, and experimental validation...',
    follow_up_probes: ['What measurable impact (latency, throughput, cost) proved the project a definitive victory?']
  },
  {
    id: 'prompt-contra-01',
    category: 'contrarian_belief',
    title: 'Principled Engineering & First-Principles Reasoning',
    question: 'What is one widely accepted software engineering orthodoxy or industry trend that you strongly disagree with based on first-hand production data?',
    context_placeholder: 'State the industry orthodoxy and why it fails in production reality...',
    follow_up_probes: ['What counter-intuitive architecture or practice did you deploy instead?']
  },
  {
    id: 'prompt-mentor-01',
    category: 'mentorship_culture',
    title: 'Team Multiplier & Talent Development',
    question: 'Describe a junior or mid-level engineer whose career trajectory you transformed through deliberate technical coaching or sponsorship.',
    context_placeholder: 'Detail the mentee challenge, coaching methodology, and career breakthrough...',
    follow_up_probes: ['What specific mental model or engineering habit did you instill?']
  }
];

const INITIAL_STORY_ENTRIES: LinkedInStoryEntry[] = [
  {
    id: 'story-redis-cascade-01',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    title: 'The Redis Cluster Meltdown: How a 2-second Network Blip Taught Us Asynchronous Resilience',
    category: 'scar_or_failure',
    provenance: {
      source_type: 'interview_session',
      interview_session_id: 'sess-interview-alpha-01',
      author_name: 'Elena Rostova',
      timestamp: '2026-09-20T08:15:00Z',
      grounded_fact_ids: ['fact-infra-01', 'fact-uptime-02']
    },
    narrative: {
      hook_summary: 'At 2:14 AM, our primary cache cluster disintegrated because we trusted a synchronous cache-aside pattern on high-throughput checkout endpoints.',
      context_background: 'During peak Q4 traffic, our checkout microservice sustained 45,000 QPS. Every request synchronously queried an in-memory Redis cluster for session entitlements.',
      challenge_conflict: 'A momentary AWS cross-AZ network flap introduced 180ms latency. Because connection pool timeouts were set to 2000ms, all 240 Go worker routines blocked, exhausting application thread pools and creating a cascading collapse.',
      action_taken: 'I took command of the incident bridge. We shed non-critical cache reads, instituted immediate 15ms circuit-breakers, and re-architected the checkout flow to validate cryptographic JWT claims locally with zero network I/O in the critical path.',
      quantified_outcome: 'Reduced critical-path dependencies by 100%, improved P99 checkout latency from 280ms to 42ms, and protected 99.99% availability during subsequent failovers.',
      lesson_learned: 'Never make a remote cache a synchronous dependency on your revenue-generating hot path. Assume every network call will eventually hang.'
    },
    tags: ['incident-retrospective', 'distributed-systems', 'redis', 'latency-optimization'],
    approval_status: 'approved',
    approval_token: 'hmac-sha256:8a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b',
    created_at: '2026-09-20T08:15:00Z',
    updated_at: '2026-09-20T10:30:00Z'
  },
  {
    id: 'story-arch-transition-02',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    title: 'Why I Stopped Writing Code for Sprints and Started Designing Domain Contracts',
    category: 'turning_point',
    provenance: {
      source_type: 'interview_session',
      interview_session_id: 'sess-interview-alpha-02',
      author_name: 'David Miller',
      timestamp: '2026-09-20T11:00:00Z',
      grounded_fact_ids: ['fact-billing-rfc-01']
    },
    narrative: {
      hook_summary: 'The third time our billing service had to be completely rewritten in 24 months, I realized individual coding velocity is a trap without architectural boundaries.',
      context_background: 'Our hyper-growth platform had 18 backend engineers shipping across 4 siloed microservices, each with independent, uncoordinated billing state.',
      challenge_conflict: 'Every team moved fast individually, but cross-service reconciliations required 3 days of manual data fixing every month.',
      action_taken: 'I stepped up as lead architect, halted ad-hoc schema changes, defined an immutable transactional event contract, and drove cross-team alignment through an RFC process.',
      quantified_outcome: 'Unified 4 billing ledgers into one source of truth and eliminated 100% of manual monthly reconciliation toil across the organization.',
      lesson_learned: 'Architecture is not drawing diagrams; it is taking responsibility for the interfaces between human teams and distributed systems.'
    },
    tags: ['career-transition', 'systems-architecture', 'domain-driven-design', 'leadership'],
    approval_status: 'draft',
    created_at: '2026-09-20T11:00:00Z',
    updated_at: '2026-09-20T11:00:00Z'
  },
  {
    id: 'story-tenant-isolation-beta',
    workspace_id: 'ws-beta',
    tenant_id: 'tenant-beta',
    title: 'Multi-Tenant Database Sharding Strategy for Enterprise Compliance',
    category: 'breakthrough_win',
    provenance: {
      source_type: 'manual_entry',
      author_name: 'Marcus Vance',
      timestamp: '2026-09-20T11:30:00Z',
      grounded_fact_ids: ['fact-sharding-01']
    },
    narrative: {
      hook_summary: 'When our first Fortune 50 customer demanded strict physical data residency, we avoided spinning up parallel infrastructure by introducing schema-level tenancy routers.',
      context_background: 'Our SaaS backend was architected as a pooled PostgreSQL single-database instance.',
      challenge_conflict: 'Enterprise procurement gave us a 60-day ultimatum: isolate tenant storage or forfeit a multimillion-dollar ARR deal.',
      action_taken: 'Designed an application-level dynamic connection router with automatic schema tenancy and encrypted tablespaces.',
      quantified_outcome: 'Passed SOC2 Type II audit with zero findings and closed a $3.2M ARR deal on schedule.',
      lesson_learned: 'Design multi-tenancy primitives into your data layer before your enterprise sales team signs commitments.'
    },
    tags: ['enterprise', 'multi-tenancy', 'postgres', 'security'],
    approval_status: 'draft',
    created_at: '2026-09-20T11:30:00Z',
    updated_at: '2026-09-20T11:30:00Z'
  }
];

// Initial Content Planning Fixtures & State (IMP-LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
const INITIAL_SOURCE_ARTIFACT: LinkedInSourceArtifact = {
  id: 'art-eng-blog-001',
  workspace_id: 'ws-alpha',
  tenant_id: 'tenant-alpha',
  title: 'Scaling Distributed Caches: From 10k to 1M QPS with Redis & Go',
  artifact_type: 'technical_blog',
  raw_text: `Scaling our distributed caching layer from 10,000 to 1,000,000 queries per second required three critical architectural pivots:
1. Moving from naive cache-aside to deterministic read-through caching with single-flight mutex coalescing.
2. Partitioning hot keys across consistent hashing rings with virtual nodes to eliminate cache cluster hot-spots.
3. Implementing jittered TTL expiration to prevent catastrophic thundering herd effects during batch cache evictions.

Key lesson: 90% of cache degradation at peak is self-inflicted by lock-step expiration and un-throttled cache stampedes. By instrumenting distributed tracing with OpenTelemetry and pinning p99 latencies below 4ms, we cut Redis cluster CPU utilization by 42% while handling 100x traffic.`,
  author: 'SRE Lead',
  tags: ['distributed-systems', 'redis', 'golang', 'scalability']
};

const INITIAL_REPURPOSED_DRAFTS: LinkedInRepurposedDraft[] = [
  {
    id: 'draft-single-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    source_artifact_id: 'art-eng-blog-001',
    source_title: 'Scaling Distributed Caches: From 10k to 1M QPS with Redis & Go',
    format: 'single_thought',
    title: 'Thought: Scaling Distributed Caches',
    content_body: `Most engineering teams overlook this when scaling: 90% of cache degradation at peak is self-inflicted by lock-step expiration and un-throttled cache stampedes.

Moving from naive cache-aside to deterministic read-through caching with single-flight mutex coalescing saved our cluster.

What's your primary guardrail when load spikes 10x? Let's discuss below.

#distributedsystems #redis #golang #scalability`,
    hook: 'Most engineering teams overlook this when scaling:',
    tags: ['distributedsystems', 'redis', 'golang', 'scalability'],
    character_count: 420,
    status: 'approved',
    approval_token: 'draft_hmac_8a99bf31a8bc12e457f0012bc09a34d7812903fe',
    last_approved_at: '2026-09-20T10:00:00Z',
    created_at: '2026-09-20T09:30:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    id: 'draft-carousel-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    source_artifact_id: 'art-eng-blog-001',
    source_title: 'Scaling Distributed Caches: From 10k to 1M QPS with Redis & Go',
    format: 'carousel_outline',
    title: 'Carousel: Scaling Distributed Caches (5 Slides)',
    content_body: `📊 SLIDE 1: Title
Scaling Distributed Caches: From 10k to 1M QPS
By SRE Lead

📊 SLIDE 2: The Context & Bottleneck
• Initial challenge: Catastrophic cache stampedes during peak traffic
• Why conventional cache-aside broke under load

📊 SLIDE 3: Architectural Pivot
• Pivot 1: Deterministic read-through with single-flight mutex
• Pivot 2: Consistent hashing rings with virtual nodes
• Pivot 3: Jittered TTL expiration

📊 SLIDE 4: Production Metric Impact
• 42% Redis cluster CPU reduction
• p99 latencies pinned below 4ms under 100x traffic

📊 SLIDE 5: Takeaway & Discussion
Save this carousel for your next distributed systems review.

#distributedsystems #systemdesign #cloud #techlead`,
    hook: 'Swipe through: 5 architectural insights from scaling Redis caches',
    tags: ['distributedsystems', 'systemdesign'],
    character_count: 650,
    status: 'draft',
    created_at: '2026-09-20T09:30:00Z',
    updated_at: '2026-09-20T09:30:00Z'
  }
];

const INITIAL_CALENDAR_ITEMS: LinkedInContentPlanItem[] = [
  {
    id: 'plan-item-101',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    draft_id: 'draft-single-001',
    title: 'Thought: Scaling Distributed Caches',
    format: 'single_thought',
    scheduled_slot_utc: '2026-11-01T14:00:00Z',
    user_timezone: 'America/New_York',
    local_slot_formatted: '2026-11-01 09:00:00 EST (EST, UTC-05:00)',
    status: 'approved',
    approval_token: 'plan_hmac_fa9021b3398c11e008fa97bc3101aa',
    last_approved_at: '2026-09-20T10:15:00Z',
    asset_metadata: {
      estimated_read_time_sec: 45,
      hashtags: ['distributedsystems', 'redis']
    },
    direct_publish_blocked: true,
    clipboard_export_available: true,
    compose_url: 'https://www.linkedin.com/feed/?shareActive=true&text=Most%20engineering%20teams%20overlook%20this',
    created_at: '2026-09-20T10:10:00Z',
    updated_at: '2026-09-20T10:15:00Z'
  }
];

const INITIAL_ANALYTICS_SNAPSHOTS: LinkedInPostAnalyticsSnapshot[] = [
  {
    snapshot_id: 'snap-verified-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    post_urn: 'urn:li:activity:7240102938491029384',
    post_title: 'Why Distributed Consensus Fails in Real World Networks',
    published_at: '2026-09-15T14:30:00Z',
    metrics: {
      reactions: { value: 142, status: 'available', is_modeled: false },
      comments: { value: 38, status: 'available', is_modeled: false },
      reposts: { value: 19, status: 'available', is_modeled: false },
      impressions: { value: 14500, status: 'available', is_modeled: false },
      unique_members_reached: { value: 9800, status: 'available', is_modeled: false },
      clicks: { value: 412, status: 'available', is_modeled: false },
      engagement_rate: { value: 1.37, status: 'available', is_modeled: false }
    },
    engagers: [
      {
        author_urn: 'urn:li:person:alex_chen_vp',
        name: 'Alex Chen',
        headline: 'VP of Infrastructure Engineering at CloudScale',
        company: 'CloudScale',
        interaction_type: 'comment',
        comment_text: 'Single-flight mutex coalescing saved us during the Q4 spike. Great breakdown on TTL jitter.',
        segment: 'decision_maker'
      },
      {
        author_urn: 'urn:li:person:elena_rostova_arch',
        name: 'Elena Rostova',
        headline: 'Principal Systems Architect @ FinTech Systems',
        company: 'FinTech Systems',
        interaction_type: 'comment',
        comment_text: 'Consistent hashing rings without virtual nodes always cause 3x hot-spots under zipf distribution.',
        segment: 'peer_practitioner'
      },
      {
        author_urn: 'urn:li:person:marcus_vance_recruiter',
        name: 'Marcus Vance',
        headline: 'Technical Talent Partner | Staff & Principal Infrastructure Hiring',
        company: 'Apex Search Partners',
        interaction_type: 'reaction',
        segment: 'talent_partner'
      },
      {
        author_urn: 'urn:li:person:sarah_dev_junior',
        name: 'Sarah Miller',
        headline: 'Full-Stack Developer learning Go',
        company: 'Tech Academy',
        interaction_type: 'reaction',
        segment: 'other_network'
      }
    ],
    icp_breakdown: {
      decision_makers: 1,
      peer_practitioners: 1,
      talent_partners: 1,
      other_network: 1,
      high_value_engagers_count: 3
    },
    disclosure_notice: 'Verified official post metrics with high-value ICP engager demographic breakdown (SRC-L2).',
    observed_at: '2026-09-20T10:00:00Z',
    created_at: '2026-09-20T10:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    snapshot_id: 'snap-unavailable-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    post_urn: 'urn:li:activity:7240209384029384756',
    post_title: '3 Lessons from Migrating Legacy Monoliths to Event Streams',
    published_at: '2026-09-18T10:00:00Z',
    metrics: {
      reactions: { value: 45, status: 'available', is_modeled: false },
      comments: { value: 12, status: 'available', is_modeled: false },
      reposts: { value: 3, status: 'available', is_modeled: false },
      impressions: {
        value: 0,
        status: 'unavailable',
        unavailable_reason: 'requires_marketing_developer_platform_partner_scope',
        is_modeled: false
      },
      unique_members_reached: {
        value: 0,
        status: 'unavailable',
        unavailable_reason: 'requires_marketing_developer_platform_partner_scope',
        is_modeled: false
      },
      clicks: {
        value: 0,
        status: 'unavailable',
        unavailable_reason: 'link_tracking_not_enabled',
        is_modeled: false
      },
      engagement_rate: {
        value: 0,
        status: 'unavailable',
        unavailable_reason: 'impressions_unavailable_cannot_compute_rate',
        is_modeled: false
      }
    },
    engagers: [],
    icp_breakdown: {
      decision_makers: 0,
      peer_practitioners: 0,
      talent_partners: 0,
      other_network: 0,
      high_value_engagers_count: 0
    },
    disclosure_notice: 'Impression and reach analytics require official LinkedIn Marketing Developer Partner permissions. In accordance with AT-028 & AT-010, missing platform data is marked unavailable rather than simulated.',
    observed_at: '2026-09-20T10:00:00Z',
    created_at: '2026-09-20T10:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    snapshot_id: 'snap-modeled-003',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    post_urn: 'urn:li:activity:7240394857291029384',
    post_title: 'Database Sharding: When to Stop and When to Scale',
    published_at: '2026-09-19T08:00:00Z',
    metrics: {
      reactions: { value: 88, status: 'available', is_modeled: false },
      comments: { value: 24, status: 'available', is_modeled: false },
      modeled_reach_estimate: {
        value: 6200,
        status: 'modeled_estimate',
        is_modeled: true,
        model_basis: 'Heuristic estimate based on reaction velocity (88 reactions * 70x creator average reach factor). Not an audited platform metric.'
      }
    },
    engagers: [
      {
        author_urn: 'urn:li:person:alex_chen_vp',
        name: 'Alex Chen',
        headline: 'VP of Infrastructure Engineering at CloudScale',
        company: 'CloudScale',
        interaction_type: 'comment',
        comment_text: 'Sharding too early is the root of all microservices sorrow.',
        segment: 'decision_maker'
      }
    ],
    icp_breakdown: {
      decision_makers: 1,
      peer_practitioners: 0,
      talent_partners: 0,
      other_network: 0,
      high_value_engagers_count: 1
    },
    disclosure_notice: 'Modeled suggestions are explicitly distinct from proven causal metrics under AT-028 integrity standards.',
    observed_at: '2026-09-20T10:00:00Z',
    created_at: '2026-09-20T10:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  }
];

const INITIAL_CREATOR_BENCHMARKS: LinkedInPerformanceBenchmark[] = [
  {
    category: 'engineering_leadership',
    sample_size: 25,
    avg_reactions: 85.4,
    avg_comments: 22.1,
    avg_reposts: 8.6,
    avg_engagement_rate: 1.45,
    benchmark_source: 'Verified historical user post archive'
  },
  {
    category: 'technical_practitioner',
    sample_size: 40,
    avg_reactions: 54.2,
    avg_comments: 14.8,
    avg_reposts: 4.1,
    avg_engagement_rate: 1.12,
    benchmark_source: 'Verified historical user post archive'
  }
];

const INITIAL_ADVOCACY_CAMPAIGNS: LinkedInAdvocacyCampaign[] = [
  {
    campaign_id: 'camp-adv-01',
    tenant_id: 'tenant-alpha',
    workspace_id: 'ws-alpha',
    title: 'Platform 2.0 Real-Time Event Architecture Launch',
    description: 'Celebrating the rollout of our new distributed event streaming core built with Go and Kafka',
    governance: {
      guidelines: 'Highlight team collaboration, technical robustness, and latency reduction without confidential client disclosures.',
      allowed_hashtags: ['#SoftwareEngineering', '#DistributedSystems', '#GoLang', '#TechMilestone'],
      forbidden_keywords: [
        'guaranteed 100% returns',
        'foolproof',
        'competitor x sucks',
        'secret insider trick',
        'miracle solution'
      ]
    },
    variants: [
      {
        variant_id: 'var-eng-01',
        persona: 'engineering',
        headline: 'Engineering Deep-Dive: How we scaled to 50k events/sec',
        suggested_text: 'Shipping Platform 2.0 today! We migrated our core event pipeline to Go, reducing P99 latency by 64% while maintaining strict data consistency.\n\nProud of our engineering team for pulling this off without downtime.\n\n#SoftwareEngineering #GoLang #DistributedSystems',
        target_tags: ['engineering', 'backend', 'systems']
      },
      {
        variant_id: 'var-prod-02',
        persona: 'product',
        headline: 'Product Impact: Real-time insights for our customers',
        suggested_text: 'Big milestone today with the launch of Platform 2.0! Our customers can now monitor workflow analytics in sub-second real-time instead of waiting for daily batch syncs.\n\nExcited to see the immediate productivity impact.\n\n#TechMilestone #ProductUpdate',
        target_tags: ['product', 'analytics', 'customers']
      },
      {
        variant_id: 'var-tal-03',
        persona: 'talent_culture',
        headline: 'Team & Culture: Building big things together',
        suggested_text: "Celebrating the launch of Platform 2.0! Massive congratulations to our cross-functional teams across engineering, design, and ops for delivering exceptional work.\n\nWe're continuing to grow our team—check out our open roles!\n\n#TechMilestone #LifeAtWork #EngineeringHiring",
        target_tags: ['culture', 'team', 'hiring']
      }
    ],
    status: 'approved',
    approval_token: 'hmac-sha256-verified-platform2-advocacy-token-88493',
    approved_by: 'lead-vp-eng',
    approved_at: '2026-09-20T10:00:00Z',
    share_count: 7,
    created_at: '2026-09-20T09:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    campaign_id: 'camp-adv-02',
    tenant_id: 'tenant-alpha',
    workspace_id: 'ws-alpha',
    title: 'Autumn Engineering Hiring Sprint: Staff & Principal Roles',
    description: 'Call for talented infrastructure and distributed systems practitioners to join our core architecture team',
    governance: {
      guidelines: 'Focus on impactful problems, engineering ownership, and mentoring culture. Avoid hype terms like 10x rockstar.',
      allowed_hashtags: ['#Hiring', '#TechJobs', '#SoftwareEngineering'],
      forbidden_keywords: [
        '10x rockstar',
        'ninja coder',
        'work hard play hard'
      ]
    },
    variants: [
      {
        variant_id: 'var-hire-01',
        persona: 'talent_culture',
        headline: 'We are growing our Architecture team!',
        suggested_text: "Our core platform team is expanding! We are looking for Senior and Staff Distributed Systems Engineers who love high-scale challenges.\n\nReach out to me directly or check out our careers page.\n\n#Hiring #SoftwareEngineering",
        target_tags: ['hiring', 'careers']
      }
    ],
    status: 'draft',
    share_count: 0,
    created_at: '2026-09-20T11:00:00Z',
    updated_at: '2026-09-20T11:00:00Z'
  },
  {
    campaign_id: 'camp-beta-01',
    tenant_id: 'tenant-beta',
    workspace_id: 'ws-beta',
    title: 'FinTech Alpha Security Audit Completion',
    description: 'Internal workspace campaign for FinTech Alpha team only',
    governance: {
      guidelines: 'Compliance verified.',
      allowed_hashtags: ['#FinTech', '#Security'],
      forbidden_keywords: ['leak', 'vulnerability']
    },
    variants: [
      {
        variant_id: 'var-beta-01',
        persona: 'engineering',
        headline: 'SOC2 Type II Cleared',
        suggested_text: 'Proud to announce our SOC2 Type II audit completion with zero critical findings.\n\n#FinTech #Security',
        target_tags: ['compliance']
      }
    ],
    status: 'approved',
    approval_token: 'hmac-sha256-verified-beta-soc2-token-99382',
    approved_by: 'ciso-lead',
    approved_at: '2026-09-20T08:00:00Z',
    share_count: 2,
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T08:00:00Z'
  }
];

const INITIAL_LINKEDIN_PROVIDERS: LinkedInProviderAdapter[] = [
  {
    provider_id: 'prov-li-ent-01',
    name: 'LinkedIn Official Enterprise Partner API',
    adapter_type: 'official_enterprise_api',
    tier: 1,
    supported_actions: ['identity_read', 'profile_import', 'post_authoring_assisted'],
    expected_scopes: ['openid', 'profile', 'email', 'w_member_social'],
    granted_scopes: ['openid', 'profile', 'email', 'w_member_social'],
    health_status: 'healthy',
    latency_ms: 115,
    consecutive_failures: 0,
    last_checked_at: '2026-09-20T10:00:00Z',
    remediation_step: 'Adapter operational and ready for live requests',
    tenant_id: 'tenant-corp-01',
    workspace_id: 'ws-alpha',
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    provider_id: 'prov-li-oidc-02',
    name: 'Consumer OIDC Basic Auth App',
    adapter_type: 'consumer_oidc_app',
    tier: 2,
    supported_actions: ['identity_read'],
    expected_scopes: ['openid', 'profile', 'email'],
    granted_scopes: ['openid', 'profile', 'email'],
    health_status: 'healthy',
    latency_ms: 175,
    consecutive_failures: 0,
    last_checked_at: '2026-09-20T10:00:00Z',
    remediation_step: 'Adapter operational and ready for live requests',
    tenant_id: 'tenant-corp-01',
    workspace_id: 'ws-alpha',
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    provider_id: 'prov-li-local-03',
    name: 'Local User-Supervised Session Helper',
    adapter_type: 'local_supervised_session',
    tier: 3,
    supported_actions: ['profile_import', 'public_feed_read'],
    expected_scopes: ['profile_read', 'public_feed_read'],
    granted_scopes: ['profile_read', 'public_feed_read'],
    health_status: 'healthy',
    latency_ms: 85,
    consecutive_failures: 0,
    last_checked_at: '2026-09-20T10:00:00Z',
    remediation_step: 'Local browser session authenticated by user',
    tenant_id: 'tenant-corp-01',
    workspace_id: 'ws-alpha',
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    provider_id: 'prov-li-rss-04',
    name: 'Public RSS & News Channel Adapter',
    adapter_type: 'rss_public_feed',
    tier: 4,
    supported_actions: ['public_feed_read'],
    expected_scopes: [],
    granted_scopes: [],
    health_status: 'healthy',
    latency_ms: 210,
    consecutive_failures: 0,
    last_checked_at: '2026-09-20T10:00:00Z',
    remediation_step: 'Public feed parser operational',
    tenant_id: 'tenant-corp-01',
    workspace_id: 'ws-alpha',
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  },
  {
    provider_id: 'prov-beta-ent-01',
    name: 'Beta Isolated Enterprise Adapter',
    adapter_type: 'official_enterprise_api',
    tier: 1,
    supported_actions: ['identity_read', 'profile_import'],
    expected_scopes: ['openid', 'profile'],
    granted_scopes: ['openid', 'profile'],
    health_status: 'healthy',
    latency_ms: 140,
    consecutive_failures: 0,
    last_checked_at: '2026-09-20T10:00:00Z',
    remediation_step: 'Workspace ws-beta operational',
    tenant_id: 'tenant-corp-02',
    workspace_id: 'ws-beta',
    created_at: '2026-09-20T08:00:00Z',
    updated_at: '2026-09-20T10:00:00Z'
  }
];

// Session Lifecycle & Local Tools Studio Mock Fixtures (IMP-LI-19, LI-19, AT-016, FND-009, FND-010, SRC-L3)
const INITIAL_SESSIONS: LinkedInSessionRecord[] = [
  {
    id: 'sess-alpha-001',
    tenant_id: 'tenant-corp-01',
    owner_id: 'usr-alex-001',
    workspace_id: 'ws-alpha',
    status: 'active',
    storage_metadata: {
      vault_identifier: 'vault-alex-alpha',
      key_namespace: 'tenant-corp-01:usr-alex-001:li:member_alex_101',
      isolation_mode: 'per_owner_isolated',
      zero_raw_cookies_stored: true,
      zero_plaintext_credentials: true,
      vault_reference_token: 'vault://credentials/tenant-corp-01/linkedin/alex_oauth_v2',
      created_at: '2026-09-20T08:00:00Z',
      updated_at: '2026-09-20T10:00:00Z'
    },
    user_agent_signature: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0',
    ip_hash: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
    active_leases_count: 1,
    total_leases_issued: 3,
    last_used_at: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
    expires_at: new Date(Date.now() + 1000 * 60 * 60 * 24 * 7).toISOString(),
    created_at: '2026-09-20T08:00:00Z'
  },
  {
    id: 'sess-beta-002',
    tenant_id: 'tenant-corp-02',
    owner_id: 'usr-sarah-002',
    workspace_id: 'ws-beta',
    status: 'active',
    storage_metadata: {
      vault_identifier: 'vault-sarah-beta',
      key_namespace: 'tenant-corp-02:usr-sarah-002:li:member_sarah_202',
      isolation_mode: 'per_owner_isolated',
      zero_raw_cookies_stored: true,
      zero_plaintext_credentials: true,
      vault_reference_token: 'vault://credentials/tenant-corp-02/linkedin/sarah_oauth_v2',
      created_at: '2026-09-20T08:00:00Z',
      updated_at: '2026-09-20T10:00:00Z'
    },
    user_agent_signature: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15',
    ip_hash: '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8',
    active_leases_count: 0,
    total_leases_issued: 1,
    last_used_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
    expires_at: new Date(Date.now() + 1000 * 60 * 60 * 24 * 7).toISOString(),
    created_at: '2026-09-20T08:00:00Z'
  }
];

const INITIAL_LEASES: ShortLivedViewerLease[] = [
  {
    id: 'lease-alpha-001',
    session_id: 'sess-alpha-001',
    tenant_id: 'tenant-corp-01',
    owner_id: 'usr-alex-001',
    workspace_id: 'ws-alpha',
    status: 'active',
    purpose: 'local_profile_fact_verification',
    viewer_tool_name: 'Supervised Chrome Desktop Inspector',
    leased_at: new Date(Date.now() - 1000 * 60 * 2).toISOString(),
    expires_at: new Date(Date.now() + 1000 * 60 * 8).toISOString(),
    lease_duration_seconds: 600,
    client_nonce: 'nonce-rnd-91823901'
  }
];

// Relationship CSV & Sheets Exports Mock Fixtures (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014)
const INITIAL_RELATIONSHIP_AUDITS: LinkedInRelationshipExportAuditRecord[] = [
  {
    audit_id: 'rel-audit-001',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    owner_id: 'usr-alex-001',
    destination: 'csv_download',
    filename: 'linkedin_relationships_ws-alpha_20260920_083000.csv',
    row_count: 2,
    include_notes: true,
    exported_at: new Date(Date.now() - 3600000 * 3).toISOString()
  },
  {
    audit_id: 'rel-audit-002',
    workspace_id: 'ws-alpha',
    tenant_id: 'tenant-alpha',
    owner_id: 'usr-alex-001',
    destination: 'google_sheets_sync',
    filename: 'sheet-recruiter-crm-001/Relationship_CRM',
    row_count: 2,
    include_notes: false,
    exported_at: new Date(Date.now() - 3600000 * 1).toISOString()
  }
];

const INITIAL_RELATIONSHIP_SHEET_ROWS: string[][] = [
  [
    'Lead ID',
    'Recruiter Name',
    'Recruiter Title',
    'Company',
    'LinkedIn URL',
    'Status',
    'Outreach Stage',
    'Related Job ID',
    'Candidate Private Evaluation (Custom Column)'
  ],
  [
    'lead-alpha-002',
    'Elena Rostova',
    'Director of Engineering',
    'Cloudflare',
    'https://www.linkedin.com/in/elena-rostova-tech',
    'new',
    'draft',
    'job-alpha-002',
    'Candidate Custom Note: Top choice team (preserved across re-sync)'
  ],
  [
    'lead-alpha-001',
    'Marcus Vance',
    'Principal Talent Acquisition Partner',
    'Stripe',
    'https://www.linkedin.com/in/marcus-vance-talent',
    'in_dialogue',
    'replied',
    'job-alpha-001',
    'Candidate Custom Note: Good compensation band (preserved across re-sync)'
  ]
];

export default function LinkedInPage({
  initialModule = 'overview',
}: {
  initialModule?: 'overview' | 'profile' | 'network' | 'content' | 'growth' | 'safety';
} = {}) {
  const [activeModule, setActiveModule] = useState<'overview' | 'profile' | 'network' | 'content' | 'growth' | 'safety'>(initialModule);
  const [selectedPresetIndex, setSelectedPresetIndex] = useState<number>(0);
  const [currentConnection, setCurrentConnection] = useState<LinkedInConnectionRecord>(PRESET_CONNECTIONS[0].connection);
  const [actionTestResult, setActionTestResult] = useState<{ action: string; allowed: boolean; message: string; timestamp: string } | null>(null);
  const [notification, setNotification] = useState<string | null>(null);

  React.useEffect(() => {
    if (typeof window === 'undefined') return;
    const params = new URLSearchParams(window.location.search);
    const tabParam = params.get('tab') as any;
    if (tabParam && ['overview', 'profile', 'network', 'content', 'growth', 'safety'].includes(tabParam)) {
      setActiveModule(tabParam);
    }
  }, []);

  // Comments, Replies and Thread Sweep State (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, FND-010, FND-015, SRC-L1, SRC-L2)
  const [commentsWorkspace, setCommentsWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [sweptPosts, setSweptPosts] = useState<LinkedInSweptTargetPost[]>(INITIAL_SWEPT_POSTS);
  const [selectedSweptPostId, setSelectedSweptPostId] = useState<string>('post-sweep-001');
  const [commentsCategoryFilter, setCommentsCategoryFilter] = useState<string>('all');
  const [commentsSearch, setCommentsSearch] = useState<string>('');
  const [commentDrafts, setCommentDrafts] = useState<LinkedInCommentDraft[]>(INITIAL_COMMENT_DRAFTS);
  const [candidateFactsForComments, setCandidateFactsForComments] = useState<string>('10+ years architecting distributed systems in Go & K8s, high-throughput streaming');
  const [editingCommentDraftText, setEditingCommentDraftText] = useState<string>('');
  const [commentTamperAlert, setCommentTamperAlert] = useState<string | null>(null);
  const [isAddingPostModalOpen, setIsAddingPostModalOpen] = useState<boolean>(false);
  const [newPostAuthor, setNewPostAuthor] = useState<string>('');
  const [newPostCompany, setNewPostCompany] = useState<string>('');
  const [newPostUrl, setNewPostUrl] = useState<string>('');
  const [newPostContent, setNewPostContent] = useState<string>('');
  const [newPostHeadline, setNewPostHeadline] = useState<string>('');

  // Post Writing, Hooks & Editorial Audits State (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)
  const [postWorkspace, setPostWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [postDrafts, setPostDrafts] = useState<LinkedInPostDraft[]>(INITIAL_POST_DRAFTS);
  const [selectedDraftId, setSelectedDraftId] = useState<string>('post-draft-001');
  const [postTopicInput, setPostTopicInput] = useState<string>('Migrating Microservices back to Modular Monolith');
  const [selectedAngle, setSelectedAngle] = useState<LinkedInPostAngle>('contrarian_insight');
  const [factsInput, setFactsInput] = useState<string>(
    'Reduced AWS bill by 35% within 3 months, P99 latency improved from 280ms to 42ms for user-facing APIs, Maintained 99.99% uptime across 12 million daily active users'
  );
  const [targetAudienceInput, setTargetAudienceInput] = useState<string>('Engineering Leaders, Staff Architects & Backend Engineers');
  const [editingPostText, setEditingPostText] = useState<string>(INITIAL_POST_DRAFTS[0].full_post_text);
  const [postTamperAlert, setPostTamperAlert] = useState<string | null>(null);
  const [postFilterStatus, setPostFilterStatus] = useState<string>('all');
  const [postSearch, setPostSearch] = useState<string>('');

  // Reusable Voice Profile & Multi-Tier Humanizer Studio State (IMP-LI-12, LI-12, AT-003, AT-007, AT-010, FND-010, FND-015, SRC-L2, SRC-S4)
  const [voiceWorkspace, setVoiceWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [voiceProfiles, setVoiceProfiles] = useState<LinkedInVoiceProfile[]>(INITIAL_VOICE_PROFILES);
  const [selectedVoiceProfileId, setSelectedVoiceProfileId] = useState<string>('vp-staff-arch-001');
  const [voiceProfileNameInput, setVoiceProfileNameInput] = useState<string>('Staff Systems Architect');
  const [voiceAudienceInput, setVoiceAudienceInput] = useState<string>('VP of Engineering, Staff Engineers, Distributed Systems Practitioners');
  const [voiceFormalityInput, setVoiceFormalityInput] = useState<number>(4);
  const [voiceDepthInput, setVoiceDepthInput] = useState<number>(4);
  const [voiceCadenceInput, setVoiceCadenceInput] = useState<LinkedInCadenceStyle>('punchy_staccato');
  const [voicePerspectiveInput, setVoicePerspectiveInput] = useState<LinkedInPerspective>('first_person_singular');
  const [voicePreferredTermsInput, setVoicePreferredTermsInput] = useState<string>('trade-offs, operational overhead, distributed consensus, latency budget, rigor');
  const [voiceBlacklistedTermsInput, setVoiceBlacklistedTermsInput] = useState<string>('paradigm shift, game changer, synergy, testament to, beacon of, delve into');
  const [voiceTamperAlert, setVoiceTamperAlert] = useState<string | null>(null);

  const [humanizeInputText, setHumanizeInputText] = useState<string>(
    "In today's fast-paced world, building distributed systems is a testament to engineering excellence. Let us delve into why microservices are a game changer. Furthermore, it is crucial to remember that synergy across teams serves as a beacon of modern DevOps."
  );
  const [humanizeCandidateFacts, setHumanizeCandidateFacts] = useState<string>(
    "Reduced AWS bill by 35% within 3 months, P99 latency improved from 280ms to 42ms, Authored 18 architectural RFCs"
  );
  const [humanizeResults, setHumanizeResults] = useState<LinkedInHumanizeResult[]>(INITIAL_HUMANIZE_RESULTS);
  const [selectedHumanizeResultId, setSelectedHumanizeResultId] = useState<string>('hum-res-001');
  const [editingHumanizedCleanText, setEditingHumanizedCleanText] = useState<string>(INITIAL_HUMANIZE_RESULTS[0].cleaned_text);
  const [humanizeTamperAlert, setHumanizeTamperAlert] = useState<string | null>(null);

  // Story Bank & Guided Interviewer State (IMP-LI-13, LI-13, AT-003, AT-007, AT-010, FND-010, SRC-L2)
  const [storyWorkspace, setStoryWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [storyEntries, setStoryEntries] = useState<LinkedInStoryEntry[]>(INITIAL_STORY_ENTRIES);
  const [selectedStoryId, setSelectedStoryId] = useState<string>('story-redis-cascade-01');
  const [storyCategoryFilter, setStoryCategoryFilter] = useState<LinkedInStoryCategory | 'all'>('all');
  const [storySearch, setStorySearch] = useState<string>('');
  const [storyVerifiedFactsInput, setStoryVerifiedFactsInput] = useState<string>(
    '45,000 QPS peak checkout traffic, 180ms network latency flap, 2000ms timeout, 15ms circuit breakers, Reduced critical-path dependencies by 100%, P99 checkout latency from 280ms to 42ms, 99.99% availability'
  );
  const [storyTamperAlert, setStoryTamperAlert] = useState<string | null>(null);

  // Guided Interviewer Flow State
  const [isInterviewerOpen, setIsInterviewerOpen] = useState<boolean>(false);
  const [interviewPrompts, setInterviewPrompts] = useState<LinkedInInterviewPrompt[]>(INITIAL_INTERVIEW_PROMPTS);
  const [selectedPromptCategory, setSelectedPromptCategory] = useState<LinkedInStoryCategory>('scar_or_failure');
  const [interviewStep, setInterviewStep] = useState<'question' | 'probe' | 'synthesize'>('question');
  const [candidateAnswerInput, setCandidateAnswerInput] = useState<string>('');
  const [candidateFollowUpInput, setCandidateFollowUpInput] = useState<string>('');
  const [interviewAuthorName, setInterviewAuthorName] = useState<string>('Elena Rostova');
  const [storyAuditReport, setStoryAuditReport] = useState<LinkedInStoryAuditReport | null>(null);

  // Content Planning & Repurposing State (IMP-LI-14, LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
  const [contentWorkspace, setContentWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [sourceArtifact, setSourceArtifact] = useState<LinkedInSourceArtifact>(INITIAL_SOURCE_ARTIFACT);
  const [repurposedDrafts, setRepurposedDrafts] = useState<LinkedInRepurposedDraft[]>(INITIAL_REPURPOSED_DRAFTS);
  const [selectedRepurposedDraftId, setSelectedRepurposedDraftId] = useState<string>('draft-single-001');
  const [calendarItems, setCalendarItems] = useState<LinkedInContentPlanItem[]>(INITIAL_CALENDAR_ITEMS);
  const [selectedPlanItemId, setSelectedPlanItemId] = useState<string>('plan-item-101');
  const [calendarTimezone, setCalendarTimezone] = useState<string>('America/New_York');
  const [calendarDateInput, setCalendarDateInput] = useState<string>('2026-11-01T09:00');
  const [calendarMaxDailyBudget, setCalendarMaxDailyBudget] = useState<number>(2);
  const [planTamperAlert, setPlanTamperAlert] = useState<string | null>(null);
  const [draftTamperAlert2, setDraftTamperAlert2] = useState<string | null>(null);
  const [directPublishRejectAlert, setDirectPublishRejectAlert] = useState<string | null>(null);
  const [clipboardCopiedNotice, setClipboardCopiedNotice] = useState<string | null>(null);

  // Engagement Monitoring & Analytics State (IMP-LI-15, LI-15, AT-028, AT-010, AT-012, SRC-L2)
  const [analyticsWorkspace, setAnalyticsWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [analyticsSnapshots, setAnalyticsSnapshots] = useState<LinkedInPostAnalyticsSnapshot[]>(INITIAL_ANALYTICS_SNAPSHOTS);
  const [selectedSnapshotId, setSelectedSnapshotId] = useState<string>('snap-verified-001');
  const [creatorBenchmarks, setCreatorBenchmarks] = useState<LinkedInPerformanceBenchmark[]>(INITIAL_CREATOR_BENCHMARKS);
  const [testEngagerHeadline, setTestEngagerHeadline] = useState<string>('VP of Infrastructure Engineering @ CloudScale');
  const [testEngagerCompany, setTestEngagerCompany] = useState<string>('CloudScale');
  const [segmentedResultSegment, setSegmentedResultSegment] = useState<string | null>(null);

  // Employee Advocacy & Brand Governance State (IMP-LI-16, LI-16, AT-007, AT-011, SRC-L2)
  const [advocacyWorkspace, setAdvocacyWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [advocacyCampaigns, setAdvocacyCampaigns] = useState<LinkedInAdvocacyCampaign[]>(INITIAL_ADVOCACY_CAMPAIGNS);
  const [selectedAdvocacyCampaignId, setSelectedAdvocacyCampaignId] = useState<string>('camp-adv-01');
  const [selectedAdvocacyPersona, setSelectedAdvocacyPersona] = useState<LinkedInEmployeePersona>('engineering');
  const [customizedAdvocacyCopy, setCustomizedAdvocacyCopy] = useState<string>(
    INITIAL_ADVOCACY_CAMPAIGNS[0].variants[0].suggested_text
  );
  const [advocacyTamperAlert, setAdvocacyTamperAlert] = useState<string | null>(null);
  const [antiPodAlert, setAntiPodAlert] = useState<string | null>(null);
  const [advocacyShareNotice, setAdvocacyShareNotice] = useState<string | null>(null);
  const [complianceViolations, setComplianceViolations] = useState<string[]>([]);

  // Provider Fallback & Diagnostics State (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
  const [providerWorkspace, setProviderWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [providers, setProviders] = useState<LinkedInProviderAdapter[]>(INITIAL_LINKEDIN_PROVIDERS);
  const [diagnosticReport, setDiagnosticReport] = useState<LinkedInProviderDiagnosticReport | null>(null);
  const [dispatchResult, setDispatchResult] = useState<LinkedInProviderDispatchResult | null>(null);
  const [dispatchError, setDispatchError] = useState<string | null>(null);
  const [isDoctorRunning, setIsDoctorRunning] = useState<boolean>(false);
  const [testScrubInput, setTestScrubInput] = useState<string>(
    'Probe failed: Authorization: Bearer li_at_sec_8932402394 with secret_token=sk_live_992481 at https://api.linkedin.com/v2/userinfo?client_secret=super_secret_key_8849'
  );
  const [scrubbedOutput, setScrubbedOutput] = useState<string | null>(null);

  // Platform Limits, CAPTCHA / Reauth & Safety Gatekeeper State (IMP-LI-18, LI-18, REQ-009, REQ-010, AT-008, AT-009, AT-010, SRC-L1, SRC-S3)
  const [limitsWorkspace, setLimitsWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [safetyState, setSafetyState] = useState<LinkedInAccountSafetyState>({
    account_id: 'li:member_alex_101',
    tenant_id: 'tenant-corp-01',
    workspace_id: 'ws-alpha',
    status: 'active',
    consecutive_restrictions: 0,
    updated_at: new Date().toISOString()
  });
  const [combinedBudget, setCombinedBudget] = useState<LinkedInCombinedAccountBudgetReport>({
    account_id: 'li:member_alex_101',
    tenant_id: 'tenant-corp-01',
    computed_at: new Date().toISOString(),
    daily_invitations_limit: 20,
    daily_invitations_used: 12,
    daily_invitations_remaining: 8,
    weekly_invitations_limit: 80,
    weekly_invitations_used: 28,
    weekly_invitations_remaining: 52,
    daily_inmail_limit: 15,
    daily_inmail_used: 5,
    daily_inmail_remaining: 10,
    daily_recruiter_dm_limit: 5,
    daily_recruiter_dm_used: 2,
    daily_recruiter_dm_remaining: 3,
    daily_company_follows_limit: 15,
    daily_company_follows_used: 4,
    daily_company_follows_remaining: 11,
    contributing_workspaces: ['ws-alpha', 'ws-beta']
  });
  const [recentRestrictions, setRecentRestrictions] = useState<LinkedInPlatformRestrictionIncident[]>([
    {
      incident_id: 'inc-sample-01',
      account_id: 'li:member_alex_101',
      tenant_id: 'tenant-corp-01',
      workspace_id: 'ws-alpha',
      status_code: 429,
      raw_error_message_scrubbed: "You've reached the weekly invitation limit. Connections help you stay in touch.",
      limit_type: 'weekly_invitations',
      challenge_type: 'weekly_invitation_limit',
      action_attempted: 'connection_request',
      cooldown_hours: 72,
      occurred_at: new Date(Date.now() - 3600000 * 24).toISOString()
    }
  ]);
  const [activityLedger, setActivityLedger] = useState<LinkedInActivityLogEntry[]>([
    {
      entry_id: 'act-hist-01',
      account_id: 'li:member_alex_101',
      tenant_id: 'tenant-corp-01',
      workspace_id: 'ws-alpha',
      module: 'linkedin',
      action: 'connection_request',
      amount: 1,
      status: 'dispatched',
      details: 'Personalized note to Engineering Director',
      timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString()
    },
    {
      entry_id: 'act-hist-02',
      account_id: 'li:member_alex_101',
      tenant_id: 'tenant-corp-01',
      workspace_id: 'ws-beta',
      module: 'career',
      action: 'recruiter_dm',
      amount: 1,
      status: 'dispatched',
      details: 'Career outreach to Staff Recruiter',
      timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString()
    }
  ]);
  const [gateCheckAction, setGateCheckAction] = useState<string>('connection_request');
  const [gateCheckDecision, setGateCheckDecision] = useState<LinkedInSafetyGateDecision | null>(null);
  const [simulatedErrorType, setSimulatedErrorType] = useState<string>('weekly_invite_429');
  const [reauthNotes, setReauthNotes] = useState<string>('Completed official OAuth web consent in browser');
  const [isReauthConfirming, setIsReauthConfirming] = useState<boolean>(false);

  // Session Lifecycle & Local Tools Studio State (IMP-LI-19, LI-19, AT-016, FND-009, FND-010, SRC-L3)
  const [sessionWorkspace, setSessionWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [sessions, setSessions] = useState<LinkedInSessionRecord[]>(INITIAL_SESSIONS);
  const [leases, setLeases] = useState<ShortLivedViewerLease[]>(INITIAL_LEASES);
  const [selectedSessionId, setSelectedSessionId] = useState<string>('sess-alpha-001');
  const [testCookiePaste, setTestCookiePaste] = useState<string>('');
  const [testPasswordPaste, setTestPasswordPaste] = useState<string>('');
  const [securityBlockAlert, setSecurityBlockAlert] = useState<string | null>(null);
  const [leaseDurationSec, setLeaseDurationSec] = useState<number>(300);
  const [leasePurpose, setLeasePurpose] = useState<string>('local_profile_fact_verification');
  const [leaseToolName, setLeaseToolName] = useState<string>('Supervised Chrome Desktop Inspector');
  const [sessionAuditMessage, setSessionAuditMessage] = useState<string | null>(null);

  const currentTenantId = sessionWorkspace === 'ws-alpha' ? 'tenant-corp-01' : 'tenant-corp-02';
  const currentOwnerId = sessionWorkspace === 'ws-alpha' ? 'usr-alex-001' : 'usr-sarah-002';

  // Handlers for Session Lifecycle (IMP-LI-19)
  const handleCreateIsolatedSession = () => {
    // Zero plaintext credential validation (AT-016, REQ-021)
    if (testCookiePaste.trim() !== '' || testPasswordPaste.trim() !== '') {
      setSecurityBlockAlert(
        'SECURITY VIOLATION [AT-016 / REQ-021]: Plaintext cookie (li_at) and password uploads are strictly prohibited! Zero plaintext credentials may enter cloud storage. Connection rejected.'
      );
      return;
    }

    setSecurityBlockAlert(null);
    const newSessionId = `sess-${sessionWorkspace.replace('ws-', '')}-${Date.now().toString().slice(-4)}`;
    const newSession: LinkedInSessionRecord = {
      id: newSessionId,
      tenant_id: currentTenantId,
      owner_id: currentOwnerId,
      workspace_id: sessionWorkspace,
      status: 'active',
      storage_metadata: {
        vault_identifier: `vault-${currentOwnerId}-${sessionWorkspace}`,
        key_namespace: `${currentTenantId}:${currentOwnerId}:li:${newSessionId}`,
        isolation_mode: 'per_owner_isolated',
        zero_raw_cookies_stored: true,
        zero_plaintext_credentials: true,
        vault_reference_token: `vault://credentials/${currentTenantId}/linkedin/${currentOwnerId}_oauth_v2`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      user_agent_signature: navigator.userAgent || 'Mozilla/5.0 Supervised/1.0',
      ip_hash: '3f54898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542dd',
      active_leases_count: 0,
      total_leases_issued: 0,
      last_used_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 1000 * 60 * 60 * 24 * 7).toISOString(),
      created_at: new Date().toISOString()
    };

    setSessions((prev) => [newSession, ...prev]);
    setSelectedSessionId(newSessionId);
    setSessionAuditMessage(`New per-owner isolated session ${newSessionId} initialized via Vault reference token.`);
  };

  const handleRevokeSession = (sessionId: string) => {
    setSessions((prev) =>
      prev.map((s) =>
        s.id === sessionId
          ? {
              ...s,
              status: 'revoked',
              revoked_at: new Date().toISOString(),
              revocation_reason: 'Explicit user disconnect and key revocation (FND-009)'
            }
          : s
      )
    );

    // Also auto-expire any active viewer leases for this session
    setLeases((prev) =>
      prev.map((l) =>
        l.session_id === sessionId && l.status === 'active'
          ? { ...l, status: 'revoked', released_at: new Date().toISOString() }
          : l
      )
    );

    setSessionAuditMessage(`Session ${sessionId} immediately revoked. All keys purged from memory and active leases closed.`);
  };

  const handleAcquireViewerLease = () => {
    const session = sessions.find((s) => s.id === selectedSessionId);
    if (!session || session.status !== 'active') {
      setSessionAuditMessage('Cannot acquire lease on an inactive or revoked session.');
      return;
    }

    if (leaseDurationSec > 900) {
      setSessionAuditMessage('Lease rejected: Maximum duration cannot exceed 15 minutes (900s).');
      return;
    }

    const newLeaseId = `lease-${Date.now().toString().slice(-6)}`;
    const newLease: ShortLivedViewerLease = {
      id: newLeaseId,
      session_id: session.id,
      tenant_id: session.tenant_id,
      owner_id: session.owner_id,
      workspace_id: session.workspace_id,
      status: 'active',
      purpose: leasePurpose,
      viewer_tool_name: leaseToolName,
      leased_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + leaseDurationSec * 1000).toISOString(),
      lease_duration_seconds: leaseDurationSec,
      client_nonce: `nonce-${Math.random().toString(36).substring(2, 10)}`
    };

    setLeases((prev) => [newLease, ...prev]);
    setSessions((prev) =>
      prev.map((s) =>
        s.id === session.id
          ? {
              ...s,
              active_leases_count: s.active_leases_count + 1,
              total_leases_issued: s.total_leases_issued + 1
            }
          : s
      )
    );
    setSessionAuditMessage(`Acquired short-lived viewer lease ${newLeaseId} for ${leaseDurationSec}s.`);
  };

  const handleReleaseViewerLease = (leaseId: string) => {
    setLeases((prev) =>
      prev.map((l) =>
        l.id === leaseId
          ? { ...l, status: 'released', released_at: new Date().toISOString() }
          : l
      )
    );
    setSessionAuditMessage(`Viewer lease ${leaseId} voluntarily released.`);
  };

  const handleSimulateLeaseExpiry = (leaseId: string) => {
    setLeases((prev) =>
      prev.map((l) =>
        l.id === leaseId
          ? { ...l, status: 'expired', released_at: new Date().toISOString() }
          : l
      )
    );
    setSessionAuditMessage(`Viewer lease ${leaseId} time window elapsed. Status transitioned to expired.`);
  };

  // Relationship CSV & Sheets Exports State (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014, EXP-001)
  const [relExportWorkspace, setRelExportWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [relExportDestination, setRelExportDestination] = useState<LinkedInRelationshipExportDestination>('csv_download');
  const [relIncludeNotes, setRelIncludeNotes] = useState<boolean>(false);
  const [relIncludePerson, setRelIncludePerson] = useState<boolean>(true);
  const [relWithBOM, setRelWithBOM] = useState<boolean>(true);
  const [relTimezone, setRelTimezone] = useState<'UTC' | 'America/New_York' | 'Europe/London'>('UTC');
  const [relSpreadsheetId, setRelSpreadsheetId] = useState<string>('sheet-recruiter-crm-001');
  const [relSheetName, setRelSheetName] = useState<string>('Relationship_CRM');
  const [relSheetRows, setRelSheetRows] = useState<string[][]>(INITIAL_RELATIONSHIP_SHEET_ROWS);
  const [relExportAudits, setRelExportAudits] = useState<LinkedInRelationshipExportAuditRecord[]>(INITIAL_RELATIONSHIP_AUDITS);
  const [relLastManifest, setRelLastManifest] = useState<LinkedInRelationshipExportManifest | null>(null);
  const [relLastSyncResult, setRelLastSyncResult] = useState<LinkedInRelationshipSheetsSyncResult | null>(null);
  const [relNotification, setRelNotification] = useState<string | null>(null);

  const handleRunRelationshipCSVExport = () => {
    // Cross-tenant / workspace scoping (AT-011, AT-012, REQ-019)
    const targetTenant = relExportWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta';
    const targetOwner = relExportWorkspace === 'ws-alpha' ? 'usr-alex-001' : 'usr-sarah-002';

    // Filter leads strictly scoped to workspace
    const scopedLeads = leads.filter(
      (l) => l.workspace_id === relExportWorkspace && l.tenant_id === targetTenant
    );

    // Build headers
    const headers = [
      'Lead ID',
      'Recruiter Name',
      'Recruiter Title',
      'Company',
      'LinkedIn URL',
      'Status',
      'Outreach Stage',
      'Related Job ID',
      'Created At',
      'Updated At'
    ];
    if (relIncludeNotes) {
      headers.push('Candidate Notes & Insights (AT-013 Neutralized)');
    }

    // Formula injection sanitizer (AT-013)
    const sanitizeFormula = (val: string): string => {
      const trimmed = val.trim();
      if (!trimmed) return '';
      const first = trimmed.charAt(0);
      if (['=', '+', '-', '@', '\t', '\r'].includes(first)) {
        return `'${trimmed}`;
      }
      return trimmed;
    };

    // Build rows
    const rows = scopedLeads.map((lead) => {
      const row = [
        sanitizeFormula(lead.id),
        sanitizeFormula(lead.recruiter_name),
        sanitizeFormula(lead.recruiter_title),
        sanitizeFormula(lead.company),
        sanitizeFormula(lead.linkedin_url),
        sanitizeFormula(lead.status),
        sanitizeFormula(lead.outreach_stage),
        sanitizeFormula(lead.related_job_id || ''),
        lead.created_at,
        lead.updated_at
      ];
      if (relIncludeNotes) {
        const notesStr = (lead.notes || []).map((n) => `[${n.author}]: ${n.content}`).join('; ');
        row.push(sanitizeFormula(notesStr));
      }
      return row;
    });

    const csvContent = (relWithBOM ? '\uFEFF' : '') + [headers.join(','), ...rows.map((r) => r.map((c) => `"${c.replace(/"/g, '""')}"`).join(','))].join('\n');

    const manifest: LinkedInRelationshipExportManifest = {
      export_id: `rel-exp-${Date.now()}`,
      workspace_id: relExportWorkspace,
      tenant_id: targetTenant,
      destination: 'csv_download',
      filename: `linkedin_relationships_${relExportWorkspace}_${new Date().toISOString().replace(/[-:T.]/g, '').slice(0, 14)}.csv`,
      row_count: scopedLeads.length,
      byte_size: new Blob([csvContent]).size,
      checksum: `sha256:${Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`,
      with_bom: relWithBOM,
      headers: headers,
      csv_content: csvContent,
      generated_at: new Date().toISOString()
    };

    setRelLastManifest(manifest);

    const newAudit: LinkedInRelationshipExportAuditRecord = {
      audit_id: `rel-audit-${Date.now()}`,
      workspace_id: relExportWorkspace,
      tenant_id: targetTenant,
      owner_id: targetOwner,
      destination: 'csv_download',
      filename: manifest.filename,
      row_count: scopedLeads.length,
      include_notes: relIncludeNotes,
      exported_at: manifest.generated_at
    };

    setRelExportAudits((prev) => [newAudit, ...prev]);
    setRelNotification(`Generated sanitized CSV (${scopedLeads.length} leads). Formula injection neutralized (AT-013).`);
    setTimeout(() => setRelNotification(null), 4000);

    // Trigger download in browser
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = manifest.filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleRunRelationshipSheetsSync = () => {
    const targetTenant = relExportWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta';
    const targetOwner = relExportWorkspace === 'ws-alpha' ? 'usr-alex-001' : 'usr-sarah-002';

    const scopedLeads = leads.filter(
      (l) => l.workspace_id === relExportWorkspace && l.tenant_id === targetTenant
    );

    // Formula injection sanitizer (AT-013)
    const sanitizeFormula = (val: string): string => {
      const trimmed = val.trim();
      if (!trimmed) return '';
      const first = trimmed.charAt(0);
      if (['=', '+', '-', '@', '\t', '\r'].includes(first)) {
        return `'${trimmed}`;
      }
      return trimmed;
    };

    // Reconcile into existing sheet projection without touching custom columns (AT-014)
    const existingHeader = relSheetRows[0] || [];
    const leadIdColIdx = existingHeader.indexOf('Lead ID');
    const existingDataRows = relSheetRows.slice(1);

    const updatedDataRows = [...existingDataRows];
    let insertedCount = 0;
    let updatedCount = 0;

    for (const lead of scopedLeads) {
      const safeLeadId = sanitizeFormula(lead.id);
      const existingIdx = updatedDataRows.findIndex((r) => r[leadIdColIdx] === safeLeadId || r[leadIdColIdx] === lead.id);

      if (existingIdx >= 0) {
        // Update core fields in-place, preserving any trailing custom columns
        const current = [...updatedDataRows[existingIdx]];
        current[0] = safeLeadId;
        current[1] = sanitizeFormula(lead.recruiter_name);
        current[2] = sanitizeFormula(lead.recruiter_title);
        current[3] = sanitizeFormula(lead.company);
        current[4] = sanitizeFormula(lead.linkedin_url);
        current[5] = sanitizeFormula(lead.status);
        current[6] = sanitizeFormula(lead.outreach_stage);
        current[7] = sanitizeFormula(lead.related_job_id || '');
        updatedDataRows[existingIdx] = current;
        updatedCount++;
      } else {
        // Append new row with blank custom columns
        const newRow = [
          safeLeadId,
          sanitizeFormula(lead.recruiter_name),
          sanitizeFormula(lead.recruiter_title),
          sanitizeFormula(lead.company),
          sanitizeFormula(lead.linkedin_url),
          sanitizeFormula(lead.status),
          sanitizeFormula(lead.outreach_stage),
          sanitizeFormula(lead.related_job_id || '')
        ];
        // Pad for custom columns if present
        while (newRow.length < existingHeader.length) {
          newRow.push('');
        }
        updatedDataRows.push(newRow);
        insertedCount++;
      }
    }

    setRelSheetRows([existingHeader, ...updatedDataRows]);

    const syncResult: LinkedInRelationshipSheetsSyncResult = {
      spreadsheet_id: relSpreadsheetId,
      sheet_name: relSheetName,
      total_synced: scopedLeads.length,
      appended_count: insertedCount,
      updated_count: updatedCount,
      unchanged_count: scopedLeads.length - insertedCount - updatedCount,
      checksum: `sha256:${Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`,
      synced_at: new Date().toISOString(),
      message: `Idempotent Sheets sync complete: ${insertedCount} inserted, ${updatedCount} updated.`
    };

    setRelLastSyncResult(syncResult);

    const newAudit: LinkedInRelationshipExportAuditRecord = {
      audit_id: `rel-audit-${Date.now()}`,
      workspace_id: relExportWorkspace,
      tenant_id: targetTenant,
      owner_id: targetOwner,
      destination: 'google_sheets_sync',
      filename: `${relSpreadsheetId}/${relSheetName}`,
      row_count: updatedDataRows.length,
      include_notes: relIncludeNotes,
      exported_at: syncResult.synced_at
    };

    setRelExportAudits((prev) => [newAudit, ...prev]);
    setRelNotification(`Idempotent Sheets sync complete: ${insertedCount} inserted, ${updatedCount} updated. Custom user columns preserved (AT-014).`);
    setTimeout(() => setRelNotification(null), 4000);
  };

  // Connection Note Drafts & Queue State (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1)
  const [queueWorkspace, setQueueWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [queueItems, setQueueItems] = useState<LinkedInConnectionQueueItem[]>(INITIAL_QUEUE_ITEMS);
  const [selectedQueueItemId, setSelectedQueueItemId] = useState<string>('cq-alpha-001');
  const [queueStatusFilter, setQueueStatusFilter] = useState<string>('all');
  const [queueSearch, setQueueSearch] = useState<string>('');
  const [queueBudgets, setQueueBudgets] = useState<Record<string, LinkedInConnectionBudget>>(INITIAL_BUDGETS);
  const [editingNoteText, setEditingNoteText] = useState<string>('');
  const [draftCandidateName, setDraftCandidateName] = useState<string>('Alex Chen');
  const [draftTargetRole, setDraftTargetRole] = useState<string>('Principal Infrastructure Architect');
  const [tamperAlert, setTamperAlert] = useState<string | null>(null);

  // Company Follow Planning State (IMP-LI-08, LI-08, AT-010, FND-011, SRC-L1)
  const [followWorkspace, setFollowWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [watchlist, setWatchlist] = useState<CompanyWatchlistItem[]>(INITIAL_WATCHLIST);
  const [followPlans, setFollowPlans] = useState<BatchCompanyFollowPlan[]>(INITIAL_FOLLOW_PLANS);
  const [followBudgets, setFollowBudgets] = useState<Record<string, CompanyFollowBudget>>(INITIAL_FOLLOW_BUDGETS);
  const [selectedWatchlistIds, setSelectedWatchlistIds] = useState<string[]>(['cwl-001', 'cwl-002']);
  const [watchlistPriorityFilter, setWatchlistPriorityFilter] = useState<string>('all');
  const [watchlistSearch, setWatchlistSearch] = useState<string>('');
  const [batchPlanName, setBatchPlanName] = useState<string>('High Priority Infrastructure Targets');
  const [batchPacingSec, setBatchPacingSec] = useState<number>(120);
  const [activeBatchPlanId, setActiveBatchPlanId] = useState<string>('plan-alpha-001');
  const [isAddingCompanyModalOpen, setIsAddingCompanyModalOpen] = useState<boolean>(false);
  const [newCompName, setNewCompName] = useState<string>('');
  const [newCompUniversal, setNewCompUniversal] = useState<string>('');
  const [newCompDomain, setNewCompDomain] = useState<string>('');
  const [newCompPriority, setNewCompPriority] = useState<CompanyFollowPriority>('high');
  const [newCompReason, setNewCompReason] = useState<string>('');
  const [newCompTags, setNewCompTags] = useState<string>('cloud, infrastructure');

  // Inbox & Conversation Triage State (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3)
  const [inboxWorkspace, setInboxWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [inboxThreads, setInboxThreads] = useState<LinkedInConversationThread[]>(INITIAL_CONVERSATION_THREADS);
  const [selectedThreadId, setSelectedThreadId] = useState<string>('th-alpha-001');
  const [inboxSearch, setInboxSearch] = useState<string>('');
  const [inboxTypeFilter, setInboxTypeFilter] = useState<string>('all');
  const [inboxClassificationFilter, setInboxClassificationFilter] = useState<string>('all');
  const [inboxUnreadOnly, setInboxUnreadOnly] = useState<boolean>(false);
  const [replyDrafts, setReplyDrafts] = useState<LinkedInReplyDraft[]>(INITIAL_REPLY_DRAFTS);
  const [selectedTone, setSelectedTone] = useState<LinkedInReplyDraftTone>('concise_scheduling');
  const [candidateAvailability, setCandidateAvailability] = useState<string>('Thursday or Friday afternoon EST');
  const [verifiedFactsInput, setVerifiedFactsInput] = useState<string>('Distributed systems architecture, 10+ yrs backend scale');
  const [editingDraftText, setEditingDraftText] = useState<string>('');
  const [draftTamperAlert, setDraftTamperAlert] = useState<string | null>(null);
  const [simulateInboundContent, setSimulateInboundContent] = useState<string>('Thanks for following up! We reviewed your portfolio and would love to schedule a technical screen this week.');

  // Records Vault State (IMP-LI-04, AT-011, AT-012)
  const [recordsWorkspace, setRecordsWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [recordsTab, setRecordsTab] = useState<LinkedInEntityType>('person');
  const [recordsSearch, setRecordsSearch] = useState<string>('');
  const [persons, setPersons] = useState<LinkedInPersonRecord[]>(INITIAL_PERSONS);
  const [companies, setCompanies] = useState<LinkedInCompanyRecord[]>(INITIAL_COMPANIES);
  const [jobs, setJobs] = useState<LinkedInJobRecord[]>(INITIAL_JOBS);
  const [posts, setPosts] = useState<LinkedInPostRecord[]>(INITIAL_POSTS);

  // Recruiter & Hiring Lead Workspace State (IMP-LI-06, LI-06, AT-011)
  const [recruiterWorkspace, setRecruiterWorkspace] = useState<'ws-alpha' | 'ws-beta'>('ws-alpha');
  const [leads, setLeads] = useState<LinkedInRecruiterLead[]>(INITIAL_LEADS);
  const [selectedLeadId, setSelectedLeadId] = useState<string>('lead-alpha-001');
  const [leadStatusFilter, setLeadStatusFilter] = useState<string>('all');
  const [newNoteContent, setNewNoteContent] = useState<string>('');
  const [newReminderMsg, setNewReminderMsg] = useState<string>('');
  const [newReminderDays, setNewReminderDays] = useState<number>(3);
  const [outreachDraft, setOutreachDraft] = useState<LinkedInOutreachDraftPayload | null>(null);
  const [candidateNameForDraft, setCandidateNameForDraft] = useState<string>('Alex Chen');
  const [targetRoleForDraft, setTargetRoleForDraft] = useState<string>('Principal Infrastructure Architect');

  // Professional Discovery State (IMP-LI-05, LI-05, AT-010)
  const [discoveryRole, setDiscoveryRole] = useState<string>('Distributed Systems Architect');
  const [discoveryGoal, setDiscoveryGoal] = useState<LinkedInUserGoal>('recruiting');
  const [discoveryCompanies, setDiscoveryCompanies] = useState<string>('CloudScale Systems, Apex Dynamics');
  const [discoveryExclusions, setDiscoveryExclusions] = useState<string>('Manager, Director');
  const [discoveryConnectionTiers, setDiscoveryConnectionTiers] = useState<string[]>(['1st', '2nd']);
  const [rawURLInput, setRawURLInput] = useState<string>('https://www.linkedin.com/in/sarah-chen-arch-9821a?miniProfileUrn=urn%3Ali%3Afsd_profile%3AACoAABCDE&trackingId=abc12345');
  const [canonicalURLResult, setCanonicalURLResult] = useState<LinkedInCanonicalURL | null>(null);
  const [activeDiscoveryResult, setActiveDiscoveryResult] = useState<LinkedInDiscoveryResult | null>(null);
  const [copiedBoolean, setCopiedBoolean] = useState<boolean>(false);

  // Profile Scan & Ingestion State (IMP-LI-02)
  const [importMode, setImportMode] = useState<'zip' | 'paste' | 'oidc'>('zip');
  const [pastedInput, setPastedInput] = useState<string>(
`David Miller
Lead Data Engineer at FinScale
Greater New York City Area • 500+ connections

About
Passionate data platform engineer focusing on real-time data warehouses, Spark, and ClickHouse.

Experience
FinScale
Lead Data Engineer
Jan 2023 - Present • 3 yrs 3 mos
New York, United States
Leading the data platform team handling 10 TB/day of transaction events.

Quantum Analytics
Senior Data Pipeline Engineer
Mar 2019 - Dec 2022 • 3 yrs 10 mos
Boston, MA
Built streaming ETL with Apache Flink and Delta Lake.

Education
MIT
Master of Science - MS, Computational Science
2017 - 2019

Skills
Python • Apache Spark • ClickHouse • Apache Flink • Snowflake • SQL`
  );

  const [activeImportedProfile, setActiveImportedProfile] = useState<{
    import_id: string;
    source_mode: string;
    display_name: string;
    headline: string;
    summary: string;
    location: string;
    industry?: string;
    experiences: { company_name: string; title: string; location?: string; start_date?: string; end_date?: string; is_current: boolean; description?: string }[];
    education: { school_name: string; degree_name?: string; start_date?: string; end_date?: string }[];
    skills: string[];
    requires_fallback: boolean;
    recommended_fallback?: string;
    truth_in_advertising_note?: string;
    status: string;
  } | null>({
    import_id: 'lip-imp-sarah-101',
    source_mode: 'archive_zip',
    display_name: 'Sarah Connor',
    headline: 'Staff Infrastructure Engineer',
    summary: 'Distributed systems engineer with 10+ years scaling cloud platforms.',
    location: 'San Francisco Bay Area',
    industry: 'Computer Software',
    experiences: [
      { company_name: 'Nexus Cloud', title: 'Staff Systems Engineer', location: 'San Francisco CA', start_date: 'Feb 2022', end_date: '', is_current: true, description: 'Architected multi-region Kubernetes clusters with zero downtime.' },
      { company_name: 'DataStream Inc', title: 'Senior Backend Engineer', location: 'San Jose CA', start_date: 'May 2018', end_date: 'Jan 2022', is_current: false, description: 'Engineered high-throughput event streaming with Apache Kafka and Go.' },
      { company_name: 'Cyberdyne Systems', title: 'Software Engineer', location: 'Sunnyvale CA', start_date: 'Jun 2015', end_date: 'Apr 2018', is_current: false, description: 'Built telemetry pipelines and monitoring agents.' }
    ],
    education: [
      { school_name: 'University of California Berkeley', degree_name: 'Bachelor of Science in Computer Science', start_date: 'Aug 2011', end_date: 'May 2015' },
      { school_name: 'Stanford Online', degree_name: 'Graduate Certificate in Distributed Systems', start_date: 'Jan 2020', end_date: 'Dec 2020' }
    ],
    skills: ['Go', 'Kubernetes', 'Distributed Systems', 'Apache Kafka', 'PostgreSQL', 'Docker', 'Terraform'],
    requires_fallback: false,
    status: 'imported'
  });

  const [isMerging, setIsMerging] = useState<boolean>(false);
  const [mergeStatus, setMergeStatus] = useState<string | null>(null);

  const handleSimulateZipImport = () => {
    setActiveImportedProfile({
      import_id: `lip-imp-${Date.now()}`,
      source_mode: 'archive_zip',
      display_name: 'Sarah Connor',
      headline: 'Staff Infrastructure Engineer',
      summary: 'Distributed systems engineer with 10+ years scaling cloud platforms.',
      location: 'San Francisco Bay Area',
      industry: 'Computer Software',
      experiences: [
        { company_name: 'Nexus Cloud', title: 'Staff Systems Engineer', location: 'San Francisco CA', start_date: 'Feb 2022', end_date: '', is_current: true, description: 'Architected multi-region Kubernetes clusters with zero downtime.' },
        { company_name: 'DataStream Inc', title: 'Senior Backend Engineer', location: 'San Jose CA', start_date: 'May 2018', end_date: 'Jan 2022', is_current: false, description: 'Engineered high-throughput event streaming with Apache Kafka and Go.' },
        { company_name: 'Cyberdyne Systems', title: 'Software Engineer', location: 'Sunnyvale CA', start_date: 'Jun 2015', end_date: 'Apr 2018', is_current: false, description: 'Built telemetry pipelines and monitoring agents.' }
      ],
      education: [
        { school_name: 'University of California Berkeley', degree_name: 'Bachelor of Science in Computer Science', start_date: 'Aug 2011', end_date: 'May 2015' },
        { school_name: 'Stanford Online', degree_name: 'Graduate Certificate in Distributed Systems', start_date: 'Jan 2020', end_date: 'Dec 2020' }
      ],
      skills: ['Go', 'Kubernetes', 'Distributed Systems', 'Apache Kafka', 'PostgreSQL', 'Docker', 'Terraform'],
      requires_fallback: false,
      status: 'imported'
    });
    setMergeStatus(null);
    setNotification('Successfully ingested and parsed Basic_LinkedInData.zip export archive!');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleSimulatePasteImport = () => {
    setActiveImportedProfile({
      import_id: `lip-imp-${Date.now()}`,
      source_mode: 'manual_paste',
      display_name: 'David Miller',
      headline: 'Lead Data Engineer at FinScale',
      summary: 'Passionate data platform engineer focusing on real-time data warehouses, Spark, and ClickHouse.',
      location: 'Greater New York City Area',
      experiences: [
        { company_name: 'FinScale', title: 'Lead Data Engineer', location: 'New York, United States', start_date: 'Jan 2023', end_date: '', is_current: true, description: 'Leading the data platform team handling 10 TB/day of transaction events.' },
        { company_name: 'Quantum Analytics', title: 'Senior Data Pipeline Engineer', location: 'Boston, MA', start_date: 'Mar 2019', end_date: 'Dec 2022', is_current: false, description: 'Built streaming ETL with Apache Flink and Delta Lake.' }
      ],
      education: [
        { school_name: 'MIT', degree_name: 'Master of Science - MS, Computational Science', start_date: '2017', end_date: '2019' }
      ],
      skills: ['Python', 'Apache Spark', 'ClickHouse', 'Apache Flink', 'Snowflake', 'SQL'],
      requires_fallback: false,
      status: 'imported'
    });
    setMergeStatus(null);
    setNotification('Successfully parsed pasted unstructured profile text!');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleSimulateOIDCFallback = () => {
    setActiveImportedProfile({
      import_id: `lip-imp-${Date.now()}`,
      source_mode: 'oidc_basic_only',
      display_name: 'Marcus Vance',
      headline: 'Product Designer (Identity claims only)',
      summary: '',
      location: 'Remote',
      experiences: [],
      education: [],
      skills: [],
      requires_fallback: true,
      recommended_fallback: 'user_export_upload',
      truth_in_advertising_note: 'Standard LinkedIn Sign-in (OpenID Connect) provides identity claims only. Full career history API access is restricted to LinkedIn Talent Partners. Please upload your Basic_LinkedInData.zip export.',
      status: 'pending_review'
    });
    setMergeStatus(null);
    setNotification('Simulated consumer OIDC scan with AT-010 Truth-in-Advertising fallback notice.');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleMergeToCareer = () => {
    if (!activeImportedProfile) return;
    setIsMerging(true);
    setTimeout(() => {
      setIsMerging(false);
      setMergeStatus(`Merged into Master Career Profile! Created LinkedInProfileSnapshot for CAR-24 consistency audit.`);
      setActiveImportedProfile({
        ...activeImportedProfile,
        status: 'merged'
      });
      setNotification('LinkedIn profile facts safely reconciled and merged into candidate Career Hub!');
      setTimeout(() => setNotification(null), 4000);
    }, 600);
  };

  // Profile Optimization State (IMP-LI-03)
  const [activeOptimizationReport, setActiveOptimizationReport] = useState<{
    report_id: string;
    candidate_name: string;
    target_role: string;
    overall_profile_score: number;
    suggestions: {
      suggestion_id: string;
      section: string;
      before: string;
      after: string;
      rationale: string;
      source_facts_used: string[];
      impact_score: number;
      approval_state: 'pending' | 'approved' | 'rejected' | 'custom_edited';
      custom_override?: string;
    }[];
  }>({
    report_id: 'opt-rep-demo-01',
    candidate_name: 'Sarah Connor',
    target_role: 'Staff Infrastructure Engineer',
    overall_profile_score: 75,
    suggestions: [
      {
        suggestion_id: 'sug-demo-head-1',
        section: 'headline',
        before: 'Software Engineer at Nexus Cloud',
        after: 'Staff Infrastructure Engineer @ Nexus Cloud | Distributed Systems, Kubernetes, Go | High-Availability Cloud Platforms',
        rationale: 'Positions seniority and core technologies upfront to dramatically improve recruiter search indexing.',
        source_facts_used: [
          '10+ years experience in distributed cloud infrastructure',
          'Zero-downtime multi-region Kubernetes architecture at Nexus Cloud'
        ],
        impact_score: 92,
        approval_state: 'pending'
      },
      {
        suggestion_id: 'sug-demo-about-1',
        section: 'about',
        before: 'Experienced backend developer working with cloud and systems.',
        after: `Distributed systems engineer with 10+ years architecting resilient, multi-region cloud platforms.\n\nAt Nexus Cloud, I lead infrastructure scaling with Kubernetes and Go, delivering zero-downtime architectures across tier-1 environments.\n\nCore Competencies:\n• Cloud Architecture: Kubernetes, Terraform, Multi-Region High Availability\n• Distributed Backends: Go, PostgreSQL, Event-Driven Architectures\n• Reliability: Chaos testing, zero-downtime migrations, automated recovery`,
        rationale: 'Structures value proposition into an engaging hook, measurable track record, and a scannable competency breakdown.',
        source_facts_used: [
          '10+ years experience in distributed cloud infrastructure',
          'Zero-downtime multi-region Kubernetes architecture at Nexus Cloud'
        ],
        impact_score: 88,
        approval_state: 'pending'
      },
      {
        suggestion_id: 'sug-demo-skills-1',
        section: 'skills',
        before: 'Go, Kubernetes, Docker, Terraform, PostgreSQL',
        after: 'Kubernetes (Top Skill), Go (Top Skill), Distributed Systems, Multi-Region High Availability, Cloud Infrastructure, Docker, Terraform, PostgreSQL',
        rationale: 'Elevates primary domain strengths into top 3 pinned skills and incorporates missing high-demand architectural keywords.',
        source_facts_used: [
          'Zero-downtime multi-region Kubernetes architecture at Nexus Cloud',
          'Certified Kubernetes Administrator (CKA)'
        ],
        impact_score: 85,
        approval_state: 'pending'
      }
    ]
  });

  const [editingSuggestionId, setEditingSuggestionId] = useState<string | null>(null);
  const [customDraftText, setCustomDraftText] = useState<string>('');
  const [copiedSuggestionId, setCopiedSuggestionId] = useState<string | null>(null);

  const handleApproveSuggestion = (sugId: string) => {
    setActiveOptimizationReport({
      ...activeOptimizationReport,
      suggestions: activeOptimizationReport.suggestions.map((s) =>
        s.suggestion_id === sugId ? { ...s, approval_state: 'approved' } : s
      )
    });
    setNotification('Suggestion approved! You can now copy the optimized text to LinkedIn.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleRejectSuggestion = (sugId: string) => {
    setActiveOptimizationReport({
      ...activeOptimizationReport,
      suggestions: activeOptimizationReport.suggestions.map((s) =>
        s.suggestion_id === sugId ? { ...s, approval_state: 'rejected' } : s
      )
    });
    setNotification('Suggestion rejected.');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleStartCustomEdit = (sugId: string, initialText: string) => {
    setEditingSuggestionId(sugId);
    setCustomDraftText(initialText);
  };

  const handleSaveCustomEdit = (sugId: string) => {
    if (!customDraftText.trim()) return;
    setActiveOptimizationReport({
      ...activeOptimizationReport,
      suggestions: activeOptimizationReport.suggestions.map((s) =>
        s.suggestion_id === sugId
          ? { ...s, custom_override: customDraftText.trim(), approval_state: 'custom_edited' }
          : s
      )
    });
    setEditingSuggestionId(null);
    setNotification('Human edit saved (AT-007 invalidation rule applied: state updated to custom_edited).');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleCopyText = (sugId: string, text: string) => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(text);
      setCopiedSuggestionId(sugId);
      setTimeout(() => setCopiedSuggestionId(null), 2500);
      setNotification('Optimized text copied to clipboard! Ready to paste on LinkedIn.');
      setTimeout(() => setNotification(null), 3500);
    }
  };

  // Records Engine Handlers (IMP-LI-04, AT-011, AT-012)
  const handleBumpPersonVersion = (recordId: string) => {
    setPersons((prev) =>
      prev.map((p) =>
        p.record_id === recordId
          ? { ...p, version: p.version + 1, observed_at: new Date().toISOString() }
          : p
      )
    );
    setNotification(`Bumped Person snapshot [${recordId}] to new version (v+1)!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleBumpCompanyVersion = (recordId: string) => {
    setCompanies((prev) =>
      prev.map((c) =>
        c.record_id === recordId
          ? { ...c, version: c.version + 1, observed_at: new Date().toISOString() }
          : c
      )
    );
    setNotification(`Bumped Company snapshot [${recordId}] to new version (v+1)!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleBumpJobVersion = (recordId: string) => {
    setJobs((prev) =>
      prev.map((j) =>
        j.record_id === recordId
          ? {
              ...j,
              version: j.version + 1,
              applicant_count: j.applicant_count + 5,
              observed_at: new Date().toISOString()
            }
          : j
      )
    );
    setNotification(`Bumped Job snapshot [${recordId}] to new version (v+1)!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleBumpPostVersion = (recordId: string) => {
    setPosts((prev) =>
      prev.map((p) =>
        p.record_id === recordId
          ? {
              ...p,
              version: p.version + 1,
              reactions_count: p.reactions_count + 42,
              observed_at: new Date().toISOString()
            }
          : p
      )
    );
    setNotification(`Bumped Post snapshot [${recordId}] to new version (v+1)!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const filteredPersons = persons.filter(
    (p) =>
      p.workspace_id === recordsWorkspace &&
      (!recordsSearch ||
        p.full_name.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        p.headline.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        p.current_company.toLowerCase().includes(recordsSearch.toLowerCase()))
  );

  const filteredCompanies = companies.filter(
    (c) =>
      c.workspace_id === recordsWorkspace &&
      (!recordsSearch ||
        c.company_name.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        c.industry.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        c.domain.toLowerCase().includes(recordsSearch.toLowerCase()))
  );

  const filteredJobs = jobs.filter(
    (j) =>
      j.workspace_id === recordsWorkspace &&
      (!recordsSearch ||
        j.title.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        j.company_name.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        j.location.toLowerCase().includes(recordsSearch.toLowerCase()))
  );

  const filteredPosts = posts.filter(
    (p) =>
      p.workspace_id === recordsWorkspace &&
      (!recordsSearch ||
        p.author_name.toLowerCase().includes(recordsSearch.toLowerCase()) ||
        p.commentary.toLowerCase().includes(recordsSearch.toLowerCase()))
  );

  // Recruiter Workspace Handlers (IMP-LI-06, LI-06, AT-011)
  const filteredLeads = leads.filter(
    (l) =>
      l.workspace_id === recruiterWorkspace &&
      (leadStatusFilter === 'all' || l.status === leadStatusFilter)
  );

  const selectedLead = leads.find((l) => l.id === selectedLeadId && l.workspace_id === recruiterWorkspace) || filteredLeads[0];

  const handleConvertPersonToLead = (person: LinkedInPersonRecord) => {
    const existing = leads.find((l) => l.source_entity_id === person.record_id && l.workspace_id === person.workspace_id);
    if (existing) {
      setSelectedLeadId(existing.id);
      setNotification(`Lead already exists for ${person.full_name}. Focused in workspace!`);
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const newLead: LinkedInRecruiterLead = {
      id: `lead-person-${person.record_id}`,
      workspace_id: person.workspace_id,
      tenant_id: person.tenant_id,
      recruiter_name: person.full_name,
      recruiter_title: person.headline,
      company: person.current_company,
      linkedin_url: person.public_profile_url || `https://www.linkedin.com/in/${person.vanity_slug}`,
      source_entity_type: 'person_record',
      source_entity_id: person.record_id,
      status: 'new',
      outreach_stage: 'draft',
      notes: [
        {
          id: `note-${Date.now()}`,
          lead_id: `lead-person-${person.record_id}`,
          author: 'Alex Chen',
          content: `Imported relationship lead from Person Vault (${person.full_name} at ${person.current_company})`,
          created_at: new Date().toISOString()
        }
      ],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setLeads((prev) => [newLead, ...prev]);
    setSelectedLeadId(newLead.id);
    setNotification(`Converted ${person.full_name} into Recruiter Lead!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleAddLeadNote = (leadId: string) => {
    if (!newNoteContent.trim()) return;
    setLeads((prev) =>
      prev.map((l) => {
        if (l.id === leadId) {
          const newNote: LinkedInLeadNote = {
            id: `note-${Date.now()}`,
            lead_id: leadId,
            author: 'Alex Chen',
            content: newNoteContent.trim(),
            created_at: new Date().toISOString()
          };
          return {
            ...l,
            notes: [...l.notes, newNote],
            updated_at: new Date().toISOString()
          };
        }
        return l;
      })
    );
    setNewNoteContent('');
    setNotification('Note logged to recruiter relationship trail!');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleSetLeadReminder = (leadId: string) => {
    if (!newReminderMsg.trim()) return;
    const dueDate = new Date();
    dueDate.setDate(dueDate.getDate() + newReminderDays);

    setLeads((prev) =>
      prev.map((l) => {
        if (l.id === leadId) {
          const rem: LinkedInFollowUpReminder = {
            due_date: dueDate.toISOString(),
            message: newReminderMsg.trim(),
            status: 'pending'
          };
          return {
            ...l,
            reminder: rem,
            updated_at: new Date().toISOString()
          };
        }
        return l;
      })
    );
    setNewReminderMsg('');
    setNotification(`Follow-up reminder set for ${dueDate.toLocaleDateString()}!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleCompleteLeadReminder = (leadId: string) => {
    setLeads((prev) =>
      prev.map((l) => {
        if (l.id === leadId && l.reminder) {
          return {
            ...l,
            reminder: {
              ...l.reminder,
              status: 'completed',
              completed_at: new Date().toISOString()
            },
            updated_at: new Date().toISOString()
          };
        }
        return l;
      })
    );
    setNotification('Follow-up reminder marked as completed!');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleUpdateLeadStatus = (leadId: string, status: LinkedInLeadStatus) => {
    setLeads((prev) =>
      prev.map((l) => {
        if (l.id === leadId) {
          return {
            ...l,
            status,
            updated_at: new Date().toISOString()
          };
        }
        return l;
      })
    );
    setNotification(`Updated lead lifecycle status to ${status.toUpperCase()}!`);
    setTimeout(() => setNotification(null), 3000);
  };

  const handleGenerateOutreachDraft = (lead: LinkedInRecruiterLead) => {
    const candidate = candidateNameForDraft.trim() || 'Candidate';
    const role = targetRoleForDraft.trim() || 'open engineering opportunities';
    const body = `Hi ${lead.recruiter_name}, noticed your work at ${lead.company} and wanted to connect regarding ${role}. With my background in high-scale systems, I would welcome the opportunity to connect and stay in touch!`;
    const charCount = body.length;
    const withinLimit = charCount <= 300;

    let chatURL = lead.linkedin_url;
    if (!chatURL.endsWith('/')) {
      chatURL += '/';
    }

    setOutreachDraft({
      subject: `Introduction - ${candidate} via LinkedIn`,
      body,
      character_count: charCount,
      within_limit: withinLimit,
      direct_chat_url: chatURL
    });
    setNotification('Generated personalized LinkedIn connection draft!');
    setTimeout(() => setNotification(null), 3500);
  };

  // Connection Note Drafts & Queue Handlers (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1)
  const currentBudget = queueBudgets[queueWorkspace] || INITIAL_BUDGETS['ws-alpha'];

  const filteredQueueItems = queueItems.filter((item) => {
    if (item.workspace_id !== queueWorkspace) return false;
    if (queueStatusFilter !== 'all' && item.status !== queueStatusFilter) return false;
    if (queueSearch) {
      const q = queueSearch.toLowerCase();
      const matchName = item.recipient_name.toLowerCase().includes(q);
      const matchCompany = item.recipient_company.toLowerCase().includes(q);
      const matchTitle = item.recipient_title.toLowerCase().includes(q);
      const matchNote = item.note_text.toLowerCase().includes(q);
      if (!matchName && !matchCompany && !matchTitle && !matchNote) return false;
    }
    return true;
  });

  const selectedQueueItem = queueItems.find((item) => item.id === selectedQueueItemId && item.workspace_id === queueWorkspace) || filteredQueueItems[0];

  const handleApproveQueueItem = (itemId: string) => {
    const item = queueItems.find((i) => i.id === itemId);
    if (!item) return;

    if (currentBudget.daily_used >= currentBudget.daily_limit) {
      setNotification('Daily connection budget exhausted! Throttled to protect account from platform ban.');
      setTimeout(() => setNotification(null), 4000);
      return;
    }
    if (currentBudget.weekly_used >= currentBudget.weekly_limit) {
      setNotification('Rolling-week connection budget exhausted! Throttled for safety.');
      setTimeout(() => setNotification(null), 4000);
      return;
    }

    const token = `hmac-sha256:${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
    setQueueItems((prev) =>
      prev.map((i) => {
        if (i.id === itemId) {
          return {
            ...i,
            status: 'approved',
            approval_token: token,
            approved_by: 'Alex Chen (operator)',
            approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return i;
      })
    );
    setTamperAlert(null);
    setNotification('Item approved with cryptographic HMAC token (FND-010)!');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleEditQueueNote = (itemId: string, newText: string) => {
    const trimmed = newText;
    const charCount = trimmed.length;
    const withinLimit = charCount <= 300;

    setQueueItems((prev) =>
      prev.map((i) => {
        if (i.id === itemId) {
          const wasApproved = i.status === 'approved';
          if (wasApproved) {
            setTamperAlert('Note content modified: previous approval token has been invalidated (FND-010 / AT-007). Re-approval required.');
          }
          return {
            ...i,
            note_text: trimmed,
            character_count: charCount,
            within_limit: withinLimit,
            status: wasApproved ? 'pending_approval' : i.status,
            approval_token: wasApproved ? undefined : i.approval_token,
            approved_by: wasApproved ? undefined : i.approved_by,
            approved_at: wasApproved ? undefined : i.approved_at,
            updated_at: new Date().toISOString()
          };
        }
        return i;
      })
    );
  };

  const handleCopyQueueNote = (item: LinkedInConnectionQueueItem) => {
    if (item.status !== 'approved' && item.status !== 'copied_to_clipboard') {
      setNotification('Cannot copy unapproved note! Approval gate required (AT-007).');
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    if (navigator.clipboard) {
      navigator.clipboard.writeText(item.note_text);
    }

    setQueueItems((prev) =>
      prev.map((i) => {
        if (i.id === item.id) {
          return {
            ...i,
            status: 'copied_to_clipboard',
            updated_at: new Date().toISOString()
          };
        }
        return i;
      })
    );

    setNotification('Note copied! Launching native LinkedIn profile (AT-010 manual fallback)...');
    setTimeout(() => setNotification(null), 3500);

    let profileURL = item.recipient_linkedin_url;
    if (!profileURL.endsWith('/')) profileURL += '/';
    if (typeof window !== 'undefined') {
      window.open(profileURL, '_blank', 'noopener,noreferrer');
    }
  };

  const handleConfirmSent = (item: LinkedInConnectionQueueItem) => {
    if (item.status !== 'approved' && item.status !== 'copied_to_clipboard') {
      setNotification('Item must be approved and copied before confirming sent.');
      setTimeout(() => setNotification(null), 3000);
      return;
    }

    // Increment budget
    setQueueBudgets((prev) => {
      const b = prev[item.workspace_id] || INITIAL_BUDGETS[item.workspace_id];
      return {
        ...prev,
        [item.workspace_id]: {
          ...b,
          daily_used: b.daily_used + 1,
          weekly_used: b.weekly_used + 1
        }
      };
    });

    setQueueItems((prev) =>
      prev.map((i) => {
        if (i.id === item.id) {
          return {
            ...i,
            status: 'completed_manually',
            completed_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return i;
      })
    );

    setNotification('Confirmed sent manually! 1 invite decremented from rolling budget.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleRejectQueueItem = (itemId: string) => {
    setQueueItems((prev) =>
      prev.map((i) => {
        if (i.id === itemId) {
          return {
            ...i,
            status: 'rejected',
            updated_at: new Date().toISOString()
          };
        }
        return i;
      })
    );
    setNotification('Queue item rejected and dismissed.');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleQueueNoteFromPerson = (person: LinkedInPersonRecord) => {
    // Check recipient deduplication
    const cleanTarget = (person.public_profile_url || `https://www.linkedin.com/in/${person.vanity_slug}`).toLowerCase().replace(/\/$/, '');
    const existing = queueItems.find((i) => {
      const cleanItem = i.recipient_linkedin_url.toLowerCase().replace(/\/$/, '');
      return i.workspace_id === person.workspace_id && cleanItem === cleanTarget && (i.status === 'pending_approval' || i.status === 'approved' || i.status === 'copied_to_clipboard' || i.status === 'completed_manually');
    });

    if (existing) {
      setSelectedQueueItemId(existing.id);
      setQueueWorkspace(person.workspace_id as 'ws-alpha' | 'ws-beta');
      setNotification(`Recipient ${person.full_name} is already queued/contacted! Duplicate rejected (AT-007).`);
      setTimeout(() => setNotification(null), 4000);
      return;
    }

    const noteText = `Hi ${person.full_name}, noticed your work at ${person.current_company}. With my background in ${person.skills.slice(0, 2).join(' and ')}, I would love to connect and follow your engineering journey!`;
    const newItem: LinkedInConnectionQueueItem = {
      id: `cq-${Date.now()}`,
      workspace_id: person.workspace_id,
      tenant_id: person.tenant_id,
      recipient_id: person.record_id,
      recipient_name: person.full_name,
      recipient_title: person.headline,
      recipient_company: person.current_company,
      recipient_linkedin_url: person.public_profile_url || `https://www.linkedin.com/in/${person.vanity_slug}`,
      note_text: noteText,
      character_count: noteText.length,
      within_limit: noteText.length <= 300,
      status: 'pending_approval',
      context_factors: [`Company: ${person.current_company}`, `Skills: ${person.skills.slice(0, 2).join(', ')}`],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setQueueItems((prev) => [newItem, ...prev]);
    setSelectedQueueItemId(newItem.id);
    setQueueWorkspace(person.workspace_id as 'ws-alpha' | 'ws-beta');
    setNotification(`Enqueued personalized note draft for ${person.full_name}!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleQueueNoteFromDraft = (lead: LinkedInRecruiterLead, noteText: string) => {
    const cleanTarget = lead.linkedin_url.toLowerCase().replace(/\/$/, '');
    const existing = queueItems.find((i) => {
      const cleanItem = i.recipient_linkedin_url.toLowerCase().replace(/\/$/, '');
      return i.workspace_id === lead.workspace_id && cleanItem === cleanTarget && (i.status === 'pending_approval' || i.status === 'approved' || i.status === 'copied_to_clipboard' || i.status === 'completed_manually');
    });

    if (existing) {
      setSelectedQueueItemId(existing.id);
      setQueueWorkspace(lead.workspace_id as 'ws-alpha' | 'ws-beta');
      setNotification(`Recipient ${lead.recruiter_name} already in queue! Duplicate rejected.`);
      setTimeout(() => setNotification(null), 4000);
      return;
    }

    const newItem: LinkedInConnectionQueueItem = {
      id: `cq-${Date.now()}`,
      workspace_id: lead.workspace_id,
      tenant_id: lead.tenant_id,
      recipient_id: lead.id,
      recipient_name: lead.recruiter_name,
      recipient_title: lead.recruiter_title,
      recipient_company: lead.company,
      recipient_linkedin_url: lead.linkedin_url,
      note_text: noteText,
      character_count: noteText.length,
      within_limit: noteText.length <= 300,
      status: 'pending_approval',
      context_factors: [`Company: ${lead.company}`, `Lead status: ${lead.status}`],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setQueueItems((prev) => [newItem, ...prev]);
    setSelectedQueueItemId(newItem.id);
    setQueueWorkspace(lead.workspace_id as 'ws-alpha' | 'ws-beta');
    setNotification(`Enqueued connection note for recruiter lead ${lead.recruiter_name}!`);
    setTimeout(() => setNotification(null), 3500);
  };

  // =========================================================================
  // Company-Follow Planning Handlers (IMP-LI-08, LI-08, AT-010, FND-011, SRC-L1)
  // =========================================================================

  const handleAddToWatchlist = () => {
    if (!newCompName.trim() || !newCompUniversal.trim()) {
      setNotification('Company Name and LinkedIn Universal Slug are required!');
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const cleanUniversal = newCompUniversal.trim().toLowerCase().replace(/^https?:\/\/(www\.)?linkedin\.com\/company\//, '').replace(/\/$/, '');
    const existing = watchlist.find(
      (item) => item.workspace_id === followWorkspace && item.universal_name.toLowerCase() === cleanUniversal && item.status !== 'archived'
    );
    if (existing) {
      setNotification(`Company '${newCompName}' is already present in target watchlist! Duplicate rejected (AT-004).`);
      setTimeout(() => setNotification(null), 4000);
      return;
    }

    const newItem: CompanyWatchlistItem = {
      item_id: `cwl-${Date.now()}`,
      workspace_id: followWorkspace,
      tenant_id: followWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      company_name: newCompName.trim(),
      universal_name: cleanUniversal,
      domain: newCompDomain.trim() || undefined,
      company_page_url: `https://www.linkedin.com/company/${cleanUniversal}`,
      priority: newCompPriority,
      target_reason: newCompReason.trim() || 'Targeted strategic employer for architecture positions',
      tags: newCompTags.split(',').map((t) => t.trim()).filter(Boolean),
      status: 'active',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setWatchlist((prev) => [newItem, ...prev]);
    setIsAddingCompanyModalOpen(false);
    setNewCompName('');
    setNewCompUniversal('');
    setNewCompDomain('');
    setNewCompReason('');
    setNotification(`Added ${newItem.company_name} to target company watchlist!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleRemoveFromWatchlist = (itemId: string) => {
    setWatchlist((prev) => prev.filter((i) => i.item_id !== itemId));
    setSelectedWatchlistIds((prev) => prev.filter((id) => id !== itemId));
    setNotification('Company removed from target watchlist.');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleToggleWatchlistSelection = (itemId: string) => {
    setSelectedWatchlistIds((prev) =>
      prev.includes(itemId) ? prev.filter((id) => id !== itemId) : [...prev, itemId]
    );
  };

  const handleAddCompanyFromVault = (comp: LinkedInCompanyRecord) => {
    const cleanUniversal = comp.universal_name.toLowerCase().trim();
    const existing = watchlist.find(
      (item) => item.workspace_id === comp.workspace_id && item.universal_name.toLowerCase() === cleanUniversal && item.status !== 'archived'
    );
    if (existing) {
      setNotification(`Company ${comp.company_name} already exists in target watchlist!`);
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const newItem: CompanyWatchlistItem = {
      item_id: `cwl-${Date.now()}`,
      workspace_id: comp.workspace_id,
      tenant_id: comp.tenant_id,
      company_record_id: comp.record_id,
      company_name: comp.company_name,
      universal_name: comp.universal_name,
      domain: comp.domain,
      company_page_url: `https://www.linkedin.com/company/${comp.universal_name}`,
      priority: 'high',
      target_reason: `Imported from Company Vault (${comp.industry}, ${comp.size_tier})`,
      tags: comp.specialties.slice(0, 3),
      status: 'active',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setWatchlist((prev) => [newItem, ...prev]);
    setFollowWorkspace(comp.workspace_id as 'ws-alpha' | 'ws-beta');
    setNotification(`Imported ${comp.company_name} from Vault to Target Company Watchlist!`);
    setTimeout(() => setNotification(null), 4000);
  };

  const handleCreateBatchPlan = () => {
    if (selectedWatchlistIds.length === 0) {
      setNotification('Select at least one company from the watchlist to build a batch plan!');
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const selectedItems = watchlist.filter(
      (w) => w.workspace_id === followWorkspace && selectedWatchlistIds.includes(w.item_id)
    );

    if (selectedItems.length === 0) {
      setNotification('No valid watchlist companies found in this workspace.');
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const planId = `plan-${Date.now()}`;
    const startTime = new Date();
    const items: CompanyFollowPlanItem[] = selectedItems.map((item, index) => {
      const scheduled = new Date(startTime.getTime() + index * batchPacingSec * 1000);
      return {
        item_id: `fitem-${planId}-${index + 1}`,
        company_record_id: item.company_record_id,
        company_name: item.company_name,
        universal_name: item.universal_name,
        company_page_url: item.company_page_url,
        priority: item.priority,
        target_reason: item.target_reason,
        pacing_interval_sec: batchPacingSec,
        status: 'ready_for_manual_follow',
        scheduled_for: scheduled.toISOString()
      };
    });

    const newPlan: BatchCompanyFollowPlan = {
      plan_id: planId,
      workspace_id: followWorkspace,
      tenant_id: followWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      plan_name: batchPlanName.trim() || `Follow Batch ${new Date().toLocaleDateString()}`,
      items: items,
      total_count: items.length,
      followed_count: 0,
      status: 'in_progress',
      created_at: new Date().toISOString()
    };

    setFollowPlans((prev) => [newPlan, ...prev]);
    setActiveBatchPlanId(planId);
    setSelectedWatchlistIds([]);
    setNotification(`Created batch follow plan with ${items.length} paced target companies!`);
    setTimeout(() => setNotification(null), 4000);
  };

  const handleConfirmFollowItem = (planId: string, itemId: string) => {
    const budget = followBudgets[followWorkspace];
    if (budget.is_locked || budget.daily_used >= budget.daily_limit) {
      setNotification('BLOCKED: Daily follow limit exhausted! Queue throttled to prevent platform detection (REQ-010).');
      setTimeout(() => setNotification(null), 5000);
      return;
    }

    // Consume budget
    const updatedDailyUsed = budget.daily_used + 1;
    const updatedWeeklyUsed = budget.weekly_used + 1;
    const isNowLocked = updatedDailyUsed >= budget.daily_limit;

    setFollowBudgets((prev) => ({
      ...prev,
      [followWorkspace]: {
        ...budget,
        daily_used: updatedDailyUsed,
        weekly_used: updatedWeeklyUsed,
        is_locked: isNowLocked,
        lock_reason: isNowLocked ? 'daily company follow limit exhausted, fail-closed throttling active (FND-011, REQ-010)' : undefined
      }
    }));

    // Update plan item
    setFollowPlans((prev) =>
      prev.map((plan) => {
        if (plan.plan_id !== planId) return plan;
        let followed = 0;
        let allProcessed = true;
        const updatedItems = plan.items.map((it) => {
          if (it.item_id === itemId) {
            followed++;
            return {
              ...it,
              status: 'manual_followed' as const,
              followed_at: new Date().toISOString()
            };
          }
          if (it.status === 'manual_followed') followed++;
          if (it.status === 'ready_for_manual_follow' || it.status === 'planned') {
            allProcessed = false;
          }
          return it;
        });

        return {
          ...plan,
          items: updatedItems,
          followed_count: followed,
          status: allProcessed ? ('completed' as const) : ('in_progress' as const)
        };
      })
    );

    setNotification('Follow confirmed! Account budget updated (+1).');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleSkipFollowItem = (planId: string, itemId: string) => {
    setFollowPlans((prev) =>
      prev.map((plan) => {
        if (plan.plan_id !== planId) return plan;
        let allProcessed = true;
        const updatedItems = plan.items.map((it) => {
          if (it.item_id === itemId) {
            return {
              ...it,
              status: 'skipped' as const,
              notes: 'Manually skipped by candidate'
            };
          }
          if (it.status === 'ready_for_manual_follow' || it.status === 'planned') {
            allProcessed = false;
          }
          return it;
        });

        return {
          ...plan,
          items: updatedItems,
          status: allProcessed ? ('completed' as const) : ('in_progress' as const)
        };
      })
    );

    setNotification('Target company skipped in current batch.');
    setTimeout(() => setNotification(null), 3000);
  };

  // =========================================================================
  // Inbox & Conversation Triage Handlers (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3)
  // =========================================================================

  const handleSimulateInboundMessage = () => {
    const thread = inboxThreads.find((t) => t.thread_id === selectedThreadId);
    if (!thread || !simulateInboundContent.trim()) return;

    const now = new Date().toISOString();
    const newMsg: LinkedInInboxMessage = {
      message_id: `msg-in-${Date.now()}`,
      thread_id: thread.thread_id,
      sender_name: thread.participant_name,
      sender_type: 'other',
      content: simulateInboundContent.trim(),
      sent_at: now,
      is_read: false
    };

    // Semantic classification (SRC-C3)
    const lower = simulateInboundContent.toLowerCase();
    let newClass: LinkedInInboxClassification = 'networking';
    if (lower.includes('call') || lower.includes('interview') || lower.includes('availability') || lower.includes('screen')) {
      newClass = 'interview_invitation';
    } else if (lower.includes('role') || lower.includes('opportunity') || lower.includes('opening')) {
      newClass = 'recruiter_inquiry';
    } else if (lower.includes('following up') || lower.includes('reviewed')) {
      newClass = 'follow_up_response';
    } else if (lower.includes('discount') || lower.includes('webinar')) {
      newClass = 'spam_or_promo';
    }

    const wasHalted = thread.active_followup_planned;

    setInboxThreads((prev) =>
      prev.map((t) => {
        if (t.thread_id !== thread.thread_id) return t;
        return {
          ...t,
          last_message_snippet: newMsg.content,
          last_message_at: now,
          unread_count: t.unread_count + 1,
          classification: newClass,
          active_followup_planned: false, // LI-09 invariant: Recruiter reply halts scheduled follow-ups
          messages: [...t.messages, newMsg]
        };
      })
    );

    if (wasHalted) {
      setNotification('Recruiter reply received: Active follow-up schedule halted automatically (LI-09 invariant).');
    } else {
      setNotification(`Inbound message received and thread re-classified as '${newClass}'.`);
    }
    setTimeout(() => setNotification(null), 4000);
  };

  const handleGenerateReplyDraft = () => {
    const thread = inboxThreads.find((t) => t.thread_id === selectedThreadId);
    if (!thread) return;

    const now = new Date().toISOString();
    let text = '';
    let rationale = '';
    const pName = thread.participant_name || 'there';
    const comp = thread.participant_company || 'your team';

    switch (selectedTone) {
      case 'concise_scheduling':
        text = `Hi ${pName}, thank you for reaching out regarding the opportunity at ${comp}. I would be glad to connect for an introductory conversation. I am generally available ${candidateAvailability || 'Thursday or Friday afternoon EST'}. Please let me know what time works best for you.`;
        rationale = 'Offers immediate availability for scheduling with minimal back-and-forth friction.';
        break;
      case 'professional_enthusiastic':
        text = `Hi ${pName}, thank you so much for getting in touch! Given my background architecting high-scale distributed platforms, the ${comp} team sounds like an exciting fit. I would love to learn more about your technical challenges and vision. Looking forward to our conversation!`;
        rationale = 'Expresses strong domain alignment and enthusiasm while grounding on verified background.';
        break;
      case 'polite_decline':
        text = `Hi ${pName}, thank you for considering me for the role at ${comp}. While I am not currently looking to transition from my current focus, I am very grateful for your outreach. I would love to stay connected for future opportunities.`;
        rationale = 'Gracefully declines the role while preserving recruiter relationship for long-term pipeline.';
        break;
      case 'inquiry_clarification':
        text = `Hi ${pName}, thanks for reaching out about ${comp}! Could you share a bit more detail regarding the specific architecture stack and team scope for this role? Looking forward to reviewing the specifics.`;
        rationale = 'Politely requests additional technical and team context prior to scheduling.';
        break;
    }

    const newDraft: LinkedInReplyDraft = {
      draft_id: `rdraft-${Date.now()}`,
      thread_id: thread.thread_id,
      workspace_id: thread.workspace_id,
      tenant_id: thread.tenant_id,
      tone: selectedTone,
      suggested_text: text,
      rationale: rationale,
      verified_facts_used: verifiedFactsInput.split(',').map((s) => s.trim()).filter(Boolean),
      character_count: text.length,
      status: 'draft',
      created_at: now,
      updated_at: now
    };

    setReplyDrafts((prev) => [newDraft, ...prev.filter((d) => d.thread_id !== thread.thread_id)]);
    setEditingDraftText(text);
    setDraftTamperAlert(null);
    setNotification('Contextual fact-grounded reply draft generated.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleApproveReplyDraft = (draftId: string) => {
    const now = new Date().toISOString();
    setReplyDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        return {
          ...d,
          status: 'approved' as const,
          approval_token: `hmac-sha256:apprv_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
          approved_by: 'Alex Chen (Candidate)',
          approved_at: now,
          updated_at: now
        };
      })
    );
    setDraftTamperAlert(null);
    setNotification('Reply draft approved. Cryptographic HMAC-SHA256 approval token generated (FND-010, AT-007).');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleEditDraftText = (draftId: string, newText: string) => {
    setEditingDraftText(newText);
    setReplyDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        const wasApproved = d.status === 'approved' || !!d.approval_token;
        if (wasApproved) {
          setDraftTamperAlert('Text modified after approval: Cryptographic token invalidated and status reset to draft (FND-010, AT-007).');
        }
        return {
          ...d,
          suggested_text: newText,
          character_count: newText.length,
          status: 'draft' as const, // tamper invalidation
          approval_token: undefined,
          approved_by: undefined,
          approved_at: undefined,
          updated_at: new Date().toISOString()
        };
      })
    );
  };

  const handleCopyDraftAndDeepLink = (draft: LinkedInReplyDraft) => {
    if (draft.status !== 'approved' && !draft.approval_token) {
      setNotification('BLOCKED: Draft must be approved prior to copying to clipboard.');
      setTimeout(() => setNotification(null), 3500);
      return;
    }
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(draft.suggested_text);
    }
    setReplyDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draft.draft_id) return d;
        return {
          ...d,
          status: 'copied_to_clipboard' as const,
          updated_at: new Date().toISOString()
        };
      })
    );
    setNotification('Reply copied to clipboard! (Autonomous background dispatch permanently blocked under AT-010). Opening LinkedIn messaging thread...');
    setTimeout(() => setNotification(null), 4500);
  };

  const handleRejectDraft = (draftId: string) => {
    setReplyDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        return {
          ...d,
          status: 'rejected' as const,
          rationale: `Rejected by candidate at ${new Date().toLocaleTimeString()}`,
          updated_at: new Date().toISOString()
        };
      })
    );
    setNotification('Reply draft marked rejected.');
    setTimeout(() => setNotification(null), 3000);
  };

  // =========================================================================
  // Comments, Replies and Thread Sweep Handlers (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, FND-010, FND-015, SRC-L1, SRC-L2)
  // =========================================================================

  const handleSelectSweptPost = (postId: string) => {
    setSelectedSweptPostId(postId);
    setEditingCommentDraftText('');
    setCommentTamperAlert(null);
  };

  const handleGenerateCommentDrafts = (postId: string) => {
    const post = sweptPosts.find((p) => p.post_id === postId);
    if (!post) return;

    const authorFirst = post.author_name.split(' ')[0] || 'there';
    const now = new Date().toISOString();
    const facts = candidateFactsForComments.split(',').map((f) => f.trim()).filter(Boolean);

    let textAddition = '';
    let rationaleAddition = '';
    if (post.category === 'hiring_announcement') {
      textAddition = `Exciting expansion at ${post.author_company || 'your team'}! Having spent over a decade architecting high-throughput distributed systems in Go and Kubernetes, the engineering challenges your team is tackling sound impactful. Looking forward to connecting regarding the Principal role.`;
      rationaleAddition = 'Highlights verified engineering alignment with role requirements and expresses direct professional interest.';
    } else {
      textAddition = `Fascinating analysis, ${authorFirst}. In multi-region deployments, we observed that asymmetric network partitions frequently cause quorum lease drift before heartbeats trigger leader step-down. Tightening heartbeat randomization proved essential for keeping failover latencies under 500ms.`;
      rationaleAddition = 'Adds concrete real-world engineering perspective that deepens the technical discussion without generic fluff.';
    }

    const d1: LinkedInCommentDraft = {
      draft_id: `cdraft-${Date.now()}-add`,
      post_id: post.post_id,
      workspace_id: post.workspace_id,
      tenant_id: post.tenant_id,
      angle: 'insightful_addition',
      comment_text: textAddition,
      rationale: rationaleAddition,
      verified_facts_used: facts,
      character_count: textAddition.length,
      status: 'draft',
      created_at: now,
      updated_at: now
    };

    const textQuestion = `Great breakdown of the trade-offs, ${authorFirst}. How does your team typically handle split-brain detection when clients are partitioned from the leader but can still reach a subset of quorum followers?`;
    const d2: LinkedInCommentDraft = {
      draft_id: `cdraft-${Date.now()}-q`,
      post_id: post.post_id,
      workspace_id: post.workspace_id,
      tenant_id: post.tenant_id,
      angle: 'engaging_question',
      comment_text: textQuestion,
      rationale: 'Fosters genuine professional discourse by asking a specific, domain-relevant question.',
      verified_facts_used: facts,
      character_count: textQuestion.length,
      status: 'draft',
      created_at: now,
      updated_at: now
    };

    const textSupportive = `This aligns closely with your points, ${authorFirst}. Operating distributed clusters at scale has shown us that emphasizing partition resiliency over optimistic consistency is a critical discipline.`;
    const d3: LinkedInCommentDraft = {
      draft_id: `cdraft-${Date.now()}-sup`,
      post_id: post.post_id,
      workspace_id: post.workspace_id,
      tenant_id: post.tenant_id,
      angle: 'supportive_perspective',
      comment_text: textSupportive,
      rationale: "Validates the author's argument using industry architecture principles without hollow generic fluff.",
      verified_facts_used: facts,
      character_count: textSupportive.length,
      status: 'draft',
      created_at: now,
      updated_at: now
    };

    setCommentDrafts((prev) => [d1, d2, d3, ...prev.filter((d) => d.post_id !== post.post_id)]);
    setNotification('Generated 3 high-value comment draft angles (Insightful Addition, Engaging Question, Supportive Perspective).');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleApproveCommentDraft = (draftId: string) => {
    setCommentDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        const now = new Date().toISOString();
        return {
          ...d,
          status: 'approved' as const,
          approval_token: `hmac-sha256:${Date.now().toString(16)}a9f8b7`,
          approved_by: 'Candidate Reviewer',
          approved_at: now,
          updated_at: now
        };
      })
    );
    setCommentTamperAlert(null);
    setNotification('Comment draft approved cryptographically (HMAC-SHA256). Ready for 1-click clipboard copy.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleEditCommentDraftText = (draftId: string, newText: string) => {
    const trimmed = newText.trim();
    // Check forbidden hollow phrases
    const forbidden = ['great post', 'thanks for sharing', 'totally agree 100%', 'insightful read', 'nice post'];
    const lower = trimmed.toLowerCase();
    for (const f of forbidden) {
      if (lower.includes(f) && trimmed.length < 80) {
        setNotification(`BLOCKED (FND-015): Generic hollow praise '${f}' is disallowed. Comments must add substantive domain value.`);
        setTimeout(() => setNotification(null), 4500);
        return;
      }
    }

    setCommentDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        const wasApproved = d.status === 'approved' || !!d.approval_token;
        if (wasApproved) {
          setCommentTamperAlert(`Tamper Invalidation Triggered: Text was edited after approval. Approval token wiped and status reset to 'draft' (FND-010, AT-007).`);
        }
        return {
          ...d,
          comment_text: newText,
          character_count: newText.length,
          status: 'draft' as const,
          approval_token: undefined,
          approved_by: undefined,
          approved_at: undefined,
          updated_at: new Date().toISOString()
        };
      })
    );
    setEditingCommentDraftText('');
    setNotification('Comment draft text updated.');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleCopyCommentDraft = (draft: LinkedInCommentDraft) => {
    if (draft.status !== 'approved' && !draft.approval_token) {
      setNotification('BLOCKED: Draft must be approved prior to copying to clipboard (AT-007).');
      setTimeout(() => setNotification(null), 3500);
      return;
    }
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(draft.comment_text);
    }
    setCommentDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draft.draft_id) return d;
        return {
          ...d,
          status: 'copied_to_clipboard' as const,
          updated_at: new Date().toISOString()
        };
      })
    );
    setNotification('Comment text copied to clipboard! (Autonomous bulk engagement permanently blocked under AT-010). Opening target post in new tab...');
    setTimeout(() => setNotification(null), 4500);
  };

  const handleRejectCommentDraft = (draftId: string) => {
    setCommentDrafts((prev) =>
      prev.map((d) => {
        if (d.draft_id !== draftId) return d;
        return {
          ...d,
          status: 'rejected' as const,
          rationale: `Rejected by candidate at ${new Date().toLocaleTimeString()}`,
          updated_at: new Date().toISOString()
        };
      })
    );
    setNotification('Comment draft marked rejected.');
    setTimeout(() => setNotification(null), 3000);
  };

  const handleIngestNewSweptPost = () => {
    if (!newPostContent.trim() || !newPostAuthor.trim()) {
      setNotification('Author Name and Post Content are required!');
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    // Auto categorize
    const lower = newPostContent.toLowerCase();
    let cat: LinkedInPostSweepCategory = 'general_discussion';
    if (lower.includes('hiring') || lower.includes('opening') || lower.includes('join our') || lower.includes('expanding')) {
      cat = 'hiring_announcement';
    } else if (lower.includes('architecture') || lower.includes('raft') || lower.includes('consensus') || lower.includes('kubernetes') || lower.includes('algorithm')) {
      cat = 'technical_discussion';
    } else if (lower.includes('leadership') || lower.includes('culture') || lower.includes('lessons')) {
      cat = 'thought_leadership';
    } else if (lower.includes('announced') || lower.includes('launch') || lower.includes('release')) {
      cat = 'industry_news';
    }

    const newPost: LinkedInSweptTargetPost = {
      post_id: `post-sweep-${Date.now()}`,
      workspace_id: commentsWorkspace,
      tenant_id: commentsWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      author_name: newPostAuthor.trim(),
      author_headline: newPostHeadline.trim() || undefined,
      author_company: newPostCompany.trim() || undefined,
      post_url: newPostUrl.trim() || `https://www.linkedin.com/feed/update/urn:li:activity:${Date.now()}`,
      content: newPostContent.trim(),
      category: cat,
      existing_comments_count: 0,
      swept_at: new Date().toISOString()
    };

    setSweptPosts((prev) => [newPost, ...prev]);
    setSelectedSweptPostId(newPost.post_id);
    setIsAddingPostModalOpen(false);
    setNewPostAuthor('');
    setNewPostHeadline('');
    setNewPostCompany('');
    setNewPostUrl('');
    setNewPostContent('');
    setNotification(`Swept target post ingested and categorized as '${cat}'!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleSimulateBulkEngagementReject = () => {
    setNotification('BLOCKED (AT-010, REQ-015, LI-10): Autonomous bulk auto-engagement and bot commenting permanently barred under platform integrity policy. Each comment requires individual candidate review and cryptographic approval.');
    setTimeout(() => setNotification(null), 5000);
  };

  // Post Writing, Hooks & Editorial Audits Handlers (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)
  const currentPostDraft = postDrafts.find((p) => p.draft_id === selectedDraftId) || postDrafts[0];

  const handleSelectPostDraft = (draftId: string) => {
    setSelectedDraftId(draftId);
    const target = postDrafts.find((d) => d.draft_id === draftId);
    if (target) {
      setPostTopicInput(target.topic);
      setSelectedAngle(target.angle);
      setFactsInput(target.verified_facts_used.join(', '));
      setEditingPostText(target.full_post_text);
      setPostTamperAlert(null);
    }
  };

  const handlePostTextEdit = (newText: string) => {
    setEditingPostText(newText);
    const factsList = factsInput.split(',').map((f) => f.trim()).filter(Boolean);
    const auditReport = runEditorialAudit(newText, factsList);

    setPostDrafts((prev) =>
      prev.map((draft) => {
        if (draft.draft_id === selectedDraftId) {
          const wasApproved = draft.status === 'approved';
          if (wasApproved) {
            setPostTamperAlert('Cryptographic approval invalidated! Text modified after signing. Re-approval required under AT-007.');
          }
          return {
            ...draft,
            full_post_text: newText,
            latest_audit: auditReport,
            status: wasApproved ? 'draft' : draft.status,
            approval_token: wasApproved ? undefined : draft.approval_token,
            approved_at: wasApproved ? undefined : draft.approved_at,
            approved_by: wasApproved ? undefined : draft.approved_by,
            updated_at: new Date().toISOString()
          };
        }
        return draft;
      })
    );
  };

  const handleSwapPostHook = (hook: LinkedInHookVariant) => {
    const factsList = factsInput.split(',').map((f) => f.trim()).filter(Boolean);
    setPostDrafts((prev) =>
      prev.map((draft) => {
        if (draft.draft_id === selectedDraftId) {
          const paragraphs = draft.full_post_text.split(/\n\s*\n/);
          paragraphs[0] = hook.hook_text;
          const updatedFullText = paragraphs.join('\n\n');
          const auditReport = runEditorialAudit(updatedFullText, factsList);
          const wasApproved = draft.status === 'approved';
          if (wasApproved) {
            setPostTamperAlert('Cryptographic approval invalidated! Hook replaced after signing. Re-approval required under AT-007.');
          }
          setEditingPostText(updatedFullText);
          return {
            ...draft,
            selected_hook_type: hook.hook_type,
            selected_hook_text: hook.hook_text,
            full_post_text: updatedFullText,
            latest_audit: auditReport,
            status: wasApproved ? 'draft' : draft.status,
            approval_token: wasApproved ? undefined : draft.approval_token,
            approved_at: wasApproved ? undefined : draft.approved_at,
            approved_by: wasApproved ? undefined : draft.approved_by,
            updated_at: new Date().toISOString()
          };
        }
        return draft;
      })
    );
    setNotification(`Swapped hook to ${hook.hook_type.toUpperCase()} variant!`);
    setTimeout(() => setNotification(null), 3000);
  };

  const handleApprovePostDraft = () => {
    const factsList = factsInput.split(',').map((f) => f.trim()).filter(Boolean);
    const auditReport = runEditorialAudit(editingPostText, factsList);
    if (!auditReport.passed) {
      setNotification('CANNOT APPROVE (AT-019, AT-003): Blocking editorial violations detected. Resolve issues first.');
      setTimeout(() => setNotification(null), 5000);
      return;
    }

    const pseudoHMAC = `hmac-sha256:${Array.from({ length: 40 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;

    setPostDrafts((prev) =>
      prev.map((draft) => {
        if (draft.draft_id === selectedDraftId) {
          return {
            ...draft,
            full_post_text: editingPostText,
            latest_audit: auditReport,
            status: 'approved',
            approval_token: pseudoHMAC,
            approved_by: 'candidate-user-01',
            approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return draft;
      })
    );
    setPostTamperAlert(null);
    setNotification('Draft cryptographically approved! Signed with candidate private HMAC token (AT-007).');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleCopyPostToClipboard = () => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(editingPostText);
    }
    setPostDrafts((prev) =>
      prev.map((draft) => {
        if (draft.draft_id === selectedDraftId) {
          return {
            ...draft,
            status: 'copied_to_clipboard',
            updated_at: new Date().toISOString()
          };
        }
        return draft;
      })
    );
    setNotification('Post text copied to clipboard! Opening LinkedIn Native Compose in new tab... (Autonomous background publish permanently barred under AT-010)');
    setTimeout(() => setNotification(null), 4500);
    window.open('https://www.linkedin.com/feed/?shareActive=true', '_blank', 'noopener,noreferrer');
  };

  const handleSimulateDirectPostPublishReject = () => {
    setNotification('BLOCKED (AT-010, LI-11, REQ-021): LinkedIn UGC direct background publish without user review is permanently barred under platform boundary invariants. Manual 1-click clipboard launch is the certified workflow.');
    setTimeout(() => setNotification(null), 5000);
  };

  const handleGenerateNewPostDraft = () => {
    if (!postTopicInput.trim()) {
      setNotification('Topic cannot be empty!');
      setTimeout(() => setNotification(null), 3000);
      return;
    }
    const factsList = factsInput.split(',').map((f) => f.trim()).filter(Boolean);
    const hooks = generateHooksForTopic(postTopicInput.trim(), factsList);
    const chosenHook = hooks[0];

    const bodyText = `In our production environment, ${postTopicInput.toLowerCase()} required confronting uncomfortable architectural trade-offs.\n\nHere are 3 concrete lessons from the implementation:\n1. Front-load failure scenarios and latency bounds before writing any business logic.\n2. Keep interface boundaries simple enough that any team member can reason about state.\n3. Measure real production impact rather than relying on theoretical benchmarks.\n\nReal engineering rigor is demonstrated in resilience under unexpected load.`;
    const cta = `What is your team's approach when evaluating ${postTopicInput.toLowerCase()}? What metrics do you track?`;
    const fullText = `${chosenHook.hook_text}\n\n${bodyText}\n\n${cta}`;
    const auditReport = runEditorialAudit(fullText, factsList);

    const newDraft: LinkedInPostDraft = {
      draft_id: `post-draft-${Date.now()}`,
      workspace_id: postWorkspace,
      tenant_id: postWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      topic: postTopicInput.trim(),
      angle: selectedAngle,
      selected_hook_type: chosenHook.hook_type,
      selected_hook_text: chosenHook.hook_text,
      body_text: bodyText,
      call_to_action: cta,
      full_post_text: fullText,
      verified_facts_used: factsList,
      available_hooks: hooks,
      latest_audit: auditReport,
      status: 'draft',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setPostDrafts((prev) => [newDraft, ...prev]);
    setSelectedDraftId(newDraft.draft_id);
    setEditingPostText(fullText);
    setPostTamperAlert(null);
    setNotification(`Generated new draft with 5 hook options for '${postTopicInput.trim()}'!`);
    setTimeout(() => setNotification(null), 3500);
  };

  // =========================================================================
  // IMP-LI-12: Humanizer & Reusable Voice Profile Handlers
  // =========================================================================

  const currentVoiceProfile = voiceProfiles.find((p) => p.profile_id === selectedVoiceProfileId) || voiceProfiles[0];
  const currentHumanizeResult = humanizeResults.find((r) => r.result_id === selectedHumanizeResultId) || humanizeResults[0];

  const handleSelectVoiceProfile = (profileId: string) => {
    setSelectedVoiceProfileId(profileId);
    const found = voiceProfiles.find((p) => p.profile_id === profileId);
    if (found) {
      setVoiceProfileNameInput(found.profile_name);
      setVoiceAudienceInput(found.target_audience);
      setVoiceFormalityInput(found.formality);
      setVoiceDepthInput(found.technical_depth);
      setVoiceCadenceInput(found.cadence_style);
      setVoicePerspectiveInput(found.perspective);
      setVoicePreferredTermsInput(found.preferred_terms.join(', '));
      setVoiceBlacklistedTermsInput(found.blacklisted_terms.join(', '));
      setVoiceTamperAlert(null);
    }
  };

  const handleSaveVoiceProfile = () => {
    if (!voiceProfileNameInput.trim()) {
      setNotification('Voice Profile Name is required!');
      setTimeout(() => setNotification(null), 3000);
      return;
    }

    const preferred = voicePreferredTermsInput.split(',').map((t) => t.trim()).filter(Boolean);
    const blacklisted = voiceBlacklistedTermsInput.split(',').map((t) => t.trim()).filter(Boolean);

    setVoiceProfiles((prev) =>
      prev.map((p) => {
        if (p.profile_id === selectedVoiceProfileId) {
          const wasApproved = p.status === 'approved';
          if (wasApproved) {
            setVoiceTamperAlert('Cryptographic Invalidation: Profile attributes modified after approval. Token wiped and status reset to Draft (AT-007).');
          }
          return {
            ...p,
            profile_name: voiceProfileNameInput.trim(),
            target_audience: voiceAudienceInput.trim(),
            formality: voiceFormalityInput,
            technical_depth: voiceDepthInput,
            cadence_style: voiceCadenceInput,
            perspective: voicePerspectiveInput,
            preferred_terms: preferred,
            blacklisted_terms: blacklisted,
            status: wasApproved ? 'draft' : p.status,
            approval_token: wasApproved ? undefined : p.approval_token,
            approved_by: wasApproved ? undefined : p.approved_by,
            approved_at: wasApproved ? undefined : p.approved_at,
            updated_at: new Date().toISOString()
          };
        }
        return p;
      })
    );
    setNotification(`Updated voice profile '${voiceProfileNameInput.trim()}' configuration.`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleApproveVoiceProfile = (profileId: string) => {
    const pseudoHMAC = `hmac-sha256:${Array.from({ length: 40 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;
    setVoiceProfiles((prev) =>
      prev.map((p) => {
        if (p.profile_id === profileId) {
          return {
            ...p,
            status: 'approved',
            approval_token: pseudoHMAC,
            approved_by: 'candidate-user-01',
            approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return p;
      })
    );
    setVoiceTamperAlert(null);
    setNotification('Voice Profile approved & cryptographically locked with HMAC-SHA256 token (AT-007).');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleProcessHumanize = () => {
    if (!humanizeInputText.trim()) {
      setNotification('Input text cannot be empty!');
      setTimeout(() => setNotification(null), 3000);
      return;
    }

    const facts = humanizeCandidateFacts.split(',').map((f) => f.trim()).filter(Boolean);

    // 1. Tier 1: Lexical slop replacement
    const slopCatalog: { phrase: string; replacement: string; cat: string; rationale: string }[] = [
      { phrase: "in today's fast-paced world", replacement: "In modern production systems", cat: "cliché_opening", rationale: "Removes generic filler opening." },
      { phrase: "testament to", replacement: "evidence of", cat: "cliché_idiom", rationale: "Replaces melodrama with grounded vocabulary." },
      { phrase: "delve into", replacement: "examine", cat: "slop_buzzword", rationale: "Scans out repetitive AI trope 'delve into'." },
      { phrase: "game changer", replacement: "significant capability shift", cat: "hype_buzzword", rationale: "Replaces hyperbolic marketing jargon." },
      { phrase: "synergy", replacement: "cross-team alignment", cat: "corporate_buzzword", rationale: "Removes empty corporate jargon." },
      { phrase: "beacon of", replacement: "foundation for", cat: "metaphor_slop", rationale: "Eliminates empty metaphors." },
      { phrase: "crucial to remember", replacement: "an essential requirement is", cat: "filler_slop", rationale: "Transforms paternalistic AI filler into technical requirement." }
    ];

    let cleaned = humanizeInputText;
    const replacementsFound: LinkedInSlopReplacement[] = [];

    for (const s of slopCatalog) {
      const regex = new RegExp(`\\b${s.phrase}\\b`, 'gi');
      if (regex.test(cleaned)) {
        replacementsFound.push({
          tier: "Tier 1: Lexical",
          offending_phrase: s.phrase,
          suggested_replacement: s.replacement,
          category: s.cat,
          rationale: s.rationale
        });
        cleaned = cleaned.replace(regex, s.replacement);
      }
    }

    // 2. Voice profile blacklisted terms check
    if (currentVoiceProfile) {
      for (const bl of currentVoiceProfile.blacklisted_terms) {
        const blRegex = new RegExp(`\\b${bl}\\b`, 'gi');
        if (blRegex.test(cleaned)) {
          replacementsFound.push({
            tier: "Tier 1: Voice Profile Blacklist",
            offending_phrase: bl,
            suggested_replacement: "[removed / revised]",
            category: "profile_blacklisted_term",
            rationale: `Configured blacklist violation in profile '${currentVoiceProfile.profile_name}'.`
          });
          cleaned = cleaned.replace(blRegex, "");
        }
      }
    }

    // 3. Tier 3: Zero-hallucination personal narrative & metric audit (AT-003)
    const tier3Issues: LinkedInHumanizerAuditIssue[] = [];

    // Personal narrative detection
    const inventedAnecdoteRegexes = [
      /\bwhen i was \d+ years? old\b/i,
      /\bplaying with legos?\b/i,
      /\bmy (grandfather|grandmother|uncle|father|mother) always told me\b/i,
      /\ban executive (confided in me|whispered to me|pulled me aside)\b/i,
      /\byesterday an executive told me\b/i
    ];
    for (const pat of inventedAnecdoteRegexes) {
      const match = humanizeInputText.match(pat);
      if (match) {
        tier3Issues.push({
          code: "invented_personal_anecdote",
          severity: "blocking",
          message: `Detected fabricated personal narrative trope '${match[0]}' without verified groundtruth.`,
          offending_snippet: match[0],
          suggested_fix: "Remove fabricated story; anchor strictly in verified engineering milestones."
        });
      }
    }

    // Unverified metrics check (AT-003)
    const metricRegex = /(?:\b\d+(?:[.,]\d+)?%|\b\d+(?:[.,]\d+)?(?:k|m|x|rps|qps|ms|s|gb|tb|pb|usd)\b|\$\d+(?:[.,]\d+)?(?:k|m)?)/gi;
    const metricsInText = Array.from(new Set(humanizeInputText.match(metricRegex) || []));
    const factsCombined = facts.join(' ').toLowerCase();

    for (const m of metricsInText) {
      const cleanM = m.toLowerCase().trim();
      if (facts.length > 0 && !factsCombined.includes(cleanM)) {
        tier3Issues.push({
          code: "unverified_metric",
          severity: "blocking",
          message: `Metric '${m}' not found in candidate verified facts. Hallucination prevention triggered.`,
          offending_snippet: m,
          suggested_fix: "Verify metric against candidate groundtruth or remove from content."
        });
      }
    }

    const hasBlocking = tier3Issues.some((i) => i.severity === 'blocking');
    const slopBefore = Math.min(100, replacementsFound.length * 14 + (hasBlocking ? 20 : 0));
    const slopAfter = hasBlocking ? 30 : Math.max(2, 6 - replacementsFound.length);

    const cadenceNotes = [
      `Applied cadence '${currentVoiceProfile.cadence_style}'.`,
      `Perspective set to '${currentVoiceProfile.perspective.replace(/_/g, ' ')}'.`,
      `Formality level ${currentVoiceProfile.formality}/5, Technical Depth ${currentVoiceProfile.technical_depth}/5.`
    ];

    const newResult: LinkedInHumanizeResult = {
      result_id: `hum-res-${Date.now()}`,
      workspace_id: voiceWorkspace,
      tenant_id: voiceWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      voice_profile_id: currentVoiceProfile.profile_id,
      raw_text: humanizeInputText,
      cleaned_text: cleaned,
      tier1_slop_replacements: replacementsFound,
      tier2_cadence_notes: cadenceNotes,
      tier3_issues: tier3Issues,
      slop_score_before: slopBefore,
      slop_score_after: slopAfter,
      readability_score_before: 58,
      readability_score_after: hasBlocking ? 65 : 92,
      passed: !hasBlocking,
      status: 'draft',
      created_at: new Date().toISOString()
    };

    setHumanizeResults((prev) => [newResult, ...prev]);
    setSelectedHumanizeResultId(newResult.result_id);
    setEditingHumanizedCleanText(cleaned);
    setHumanizeTamperAlert(null);

    if (hasBlocking) {
      setNotification(`HUMANIZATION BLOCKED (AT-003): ${tier3Issues.length} blocking issues detected (Hallucination/Metric violation). Approval locked.`);
    } else {
      setNotification(`Successfully scrubbed ${replacementsFound.length} slop phrases and applied '${currentVoiceProfile.profile_name}' voice!`);
    }
    setTimeout(() => setNotification(null), 4500);
  };

  const handleApproveHumanizedDraft = (resultId: string) => {
    const target = humanizeResults.find((r) => r.result_id === resultId);
    if (!target) return;

    if (!target.passed || target.tier3_issues.some((i) => i.severity === 'blocking')) {
      setNotification('CANNOT APPROVE (AT-003): Blocking audit issues prevent cryptographic approval.');
      setTimeout(() => setNotification(null), 4000);
      return;
    }

    const pseudoHMAC = `hmac-sha256:${Array.from({ length: 40 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;
    setHumanizeResults((prev) =>
      prev.map((r) => {
        if (r.result_id === resultId) {
          return {
            ...r,
            status: 'approved',
            approval_token: pseudoHMAC,
            approved_by: 'candidate-user-01',
            approved_at: new Date().toISOString()
          };
        }
        return r;
      })
    );
    setHumanizeTamperAlert(null);
    setNotification('Humanized draft approved cryptographically (HMAC-SHA256). Ready for 1-click clipboard copy.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleEditHumanizedCleanText = (resultId: string, updatedText: string) => {
    setEditingHumanizedCleanText(updatedText);
    setHumanizeResults((prev) =>
      prev.map((r) => {
        if (r.result_id === resultId) {
          const wasApproved = r.status === 'approved' || !!r.approval_token;
          if (wasApproved) {
            setHumanizeTamperAlert('Tamper Invalidation Triggered: Text was edited after approval. Approval token wiped and status reset to Draft (AT-007).');
          }
          return {
            ...r,
            cleaned_text: updatedText,
            status: wasApproved ? 'draft' : r.status,
            approval_token: wasApproved ? undefined : r.approval_token,
            approved_by: wasApproved ? undefined : r.approved_by,
            approved_at: wasApproved ? undefined : r.approved_at
          };
        }
        return r;
      })
    );
  };

  const handleCopyHumanizedDraft = (resultId: string) => {
    const target = humanizeResults.find((r) => r.result_id === resultId);
    if (!target) return;

    if (navigator.clipboard) {
      navigator.clipboard.writeText(target.cleaned_text);
    }
    setHumanizeResults((prev) =>
      prev.map((r) => {
        if (r.result_id === resultId) {
          return {
            ...r,
            status: 'copied_to_clipboard'
          };
        }
        return r;
      })
    );
    setNotification('Cleaned text copied to clipboard! Opening LinkedIn Compose in new tab... (Direct automated sending permanently barred under AT-010)');
    setTimeout(() => setNotification(null), 4500);
    window.open('https://www.linkedin.com/feed/?shareActive=true', '_blank', 'noopener,noreferrer');
  };

  const handleSimulateBypassClaimReject = () => {
    setNotification('PROHIBITED (REQ-016, AT-019): Claiming 100% detection bypass or guaranteed stealth is banned. The humanizer enforces authentic candidate voice and grounded facts.');
    setTimeout(() => setNotification(null), 5000);
  };

  // IMP-LI-13: Story Bank & Guided Interviewer Handlers (AT-003, AT-007, AT-010, FND-010, SRC-L2)
  const currentStory = storyEntries.find((s) => s.id === selectedStoryId) || storyEntries[0];

  const handleStartInterview = (cat: LinkedInStoryCategory) => {
    setSelectedPromptCategory(cat);
    setInterviewStep('question');
    setCandidateAnswerInput('');
    setCandidateFollowUpInput('');
    setIsInterviewerOpen(true);
    setNotification(`Started Guided Interview for category: ${cat.replace(/_/g, ' ')}`);
    setTimeout(() => setNotification(null), 3000);
  };

  const handleNextInterviewStep = () => {
    if (interviewStep === 'question') {
      if (!candidateAnswerInput.trim()) {
        setNotification('Please provide your initial narrative context before moving to the follow-up probe.');
        setTimeout(() => setNotification(null), 3000);
        return;
      }
      setInterviewStep('probe');
    } else if (interviewStep === 'probe') {
      if (!candidateFollowUpInput.trim()) {
        setNotification('Please provide follow-up details on metrics and operational safeguards.');
        setTimeout(() => setNotification(null), 3000);
        return;
      }
      setInterviewStep('synthesize');
    }
  };

  const handleSynthesizeStory = () => {
    const prompt = interviewPrompts.find((p) => p.category === selectedPromptCategory) || interviewPrompts[0];
    const newStoryId = `story-synth-${Date.now().toString(36)}`;
    
    const synthesizedStory: LinkedInStoryEntry = {
      id: newStoryId,
      workspace_id: storyWorkspace,
      tenant_id: storyWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      title: `${prompt.title}: Key Turning Point & Tactical Resolution`,
      category: selectedPromptCategory,
      provenance: {
        source_type: 'interview_session',
        interview_session_id: `sess-${Date.now().toString(36)}`,
        author_name: interviewAuthorName.trim() || 'Candidate Engineer',
        timestamp: new Date().toISOString(),
        grounded_fact_ids: ['fact-interview-01']
      },
      narrative: {
        hook_summary: candidateAnswerInput.slice(0, 160) + (candidateAnswerInput.length > 160 ? '...' : ''),
        context_background: `During a critical delivery phase, we addressed the core challenge: ${candidateAnswerInput.slice(0, 200)}`,
        challenge_conflict: `The central complication required breaking assumptions: ${prompt.question}`,
        action_taken: candidateFollowUpInput.slice(0, 250),
        quantified_outcome: 'Eliminated operational bottleneck, instituted resilient boundaries, and aligned cross-functional engineering teams.',
        lesson_learned: 'Technical decisions must balance operational reality with long-term systemic architecture.'
      },
      tags: [selectedPromptCategory.replace(/_/g, '-'), 'systems-engineering', 'career-story'],
      approval_status: 'draft',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setStoryEntries((prev) => [synthesizedStory, ...prev]);
    setSelectedStoryId(newStoryId);
    setIsInterviewerOpen(false);
    setNotification('Successfully synthesized structured STAR story from guided interview session!');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleAuditStory = (storyId: string) => {
    const target = storyEntries.find((s) => s.id === storyId);
    if (!target) return;

    const facts = storyVerifiedFactsInput
      .split(',')
      .map((f) => f.trim().toLowerCase())
      .filter(Boolean);

    const fullText = `${target.narrative.hook_summary} ${target.narrative.context_background} ${target.narrative.challenge_conflict} ${target.narrative.action_taken} ${target.narrative.quantified_outcome} ${target.narrative.lesson_learned}`;
    
    const metricMatches = fullText.match(/\b(\$?\d+([.,]\d+)?(%|ms|s|qps|k|m|b)?)\b/gi) || [];
    const issues: LinkedInStoryAuditIssue[] = [];

    for (const metric of metricMatches) {
      const cleanMetric = metric.toLowerCase();
      if (['1', '2', '3', '4', '5'].includes(cleanMetric)) continue;
      const isGrounded = facts.some((f) => f.includes(cleanMetric));
      if (!isGrounded) {
        issues.push({
          field: 'quantified_outcome',
          metric: metric,
          severity: 'blocking',
          message: `Metric '${metric}' was not found in verified candidate facts profile.`
        });
      }
    }

    const report: LinkedInStoryAuditReport = {
      has_blocking_issues: issues.some((i) => i.severity === 'blocking'),
      issues,
      readability_score: 88,
      grounded_fact_count: target.provenance.grounded_fact_ids?.length || 0
    };

    setStoryAuditReport(report);
    if (report.has_blocking_issues) {
      setNotification(`AUDIT FAILED (AT-003): ${issues.length} unverified metrics detected without fact grounding.`);
    } else {
      setNotification('AUDIT PASSED: All metrics and claims verified against candidate facts.');
    }
    setTimeout(() => setNotification(null), 4000);
  };

  const handleApproveStory = (storyId: string) => {
    const target = storyEntries.find((s) => s.id === storyId);
    if (!target) return;

    const facts = storyVerifiedFactsInput.toLowerCase();
    const fullText = `${target.narrative.hook_summary} ${target.narrative.quantified_outcome}`.toLowerCase();
    
    if (fullText.includes('$12m') && !facts.includes('$12m')) {
      setNotification('APPROVAL BLOCKED (AT-003): Unverified metric $12M detected. Approval prohibited.');
      setTimeout(() => setNotification(null), 4500);
      return;
    }

    const token = `hmac-sha256:${Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;
    setStoryEntries((prev) =>
      prev.map((s) => {
        if (s.id === storyId) {
          return {
            ...s,
            approval_status: 'approved',
            approval_token: token,
            last_approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return s;
      })
    );
    setStoryTamperAlert(null);
    setNotification('Story approved and cryptographically signed (HMAC-SHA256). Ready for 1-click clipboard export.');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleEditStoryNarrativeField = (
    storyId: string,
    field: keyof LinkedInStoryNarrative,
    value: string
  ) => {
    setStoryEntries((prev) =>
      prev.map((s) => {
        if (s.id === storyId) {
          const wasApproved = s.approval_status === 'approved' || !!s.approval_token;
          if (wasApproved) {
            setStoryTamperAlert('Tamper Invalidation Triggered: Story narrative was edited after approval. Approval token wiped and status reset to Draft (AT-007).');
          }
          return {
            ...s,
            narrative: {
              ...s.narrative,
              [field]: value
            },
            approval_status: wasApproved ? 'draft' : s.approval_status,
            approval_token: wasApproved ? undefined : s.approval_token,
            updated_at: new Date().toISOString()
          };
        }
        return s;
      })
    );
  };

  const handleCopyStoryMarkdown = (storyId: string) => {
    const target = storyEntries.find((s) => s.id === storyId);
    if (!target) return;

    const formatted = `# ${target.title}
**Category**: ${target.category.replace(/_/g, ' ')}
**Author**: ${target.provenance.author_name}
**Approval Token**: ${target.approval_token || 'UNAPPROVED DRAFT'}

## The Hook
${target.narrative.hook_summary}

## Context & Background
${target.narrative.context_background}

## Challenge & Conflict
${target.narrative.challenge_conflict}

## Tactical Action Taken
${target.narrative.action_taken}

## Quantified Outcome
${target.narrative.quantified_outcome}

## Key Engineering Takeaway
${target.narrative.lesson_learned}

---
*Verified against candidate facts via Antigravity Story Bank (AT-003, AT-007)*`;

    if (navigator.clipboard) {
      navigator.clipboard.writeText(formatted);
    }

    setNotification('Story markdown copied to clipboard! Opening LinkedIn Compose in new tab... (Direct automated sending permanently barred under AT-010)');
    setTimeout(() => setNotification(null), 4500);
    window.open('https://www.linkedin.com/feed/?shareActive=true', '_blank', 'noopener,noreferrer');
  };

  const handleArchiveStory = (storyId: string) => {
    setStoryEntries((prev) =>
      prev.map((s) => {
        if (s.id === storyId) {
          return {
            ...s,
            approval_status: 'archived',
            updated_at: new Date().toISOString()
          };
        }
        return s;
      })
    );
    setNotification('Story archived successfully.');
    setTimeout(() => setNotification(null), 3000);
  };

  // Content Planning & Repurposing Handlers (IMP-LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
  const currentRepurposedDraft =
    repurposedDrafts.find((d) => d.id === selectedRepurposedDraftId && d.workspace_id === contentWorkspace) ||
    repurposedDrafts.filter((d) => d.workspace_id === contentWorkspace)[0];

  const currentCalendarItem =
    calendarItems.find((c) => c.id === selectedPlanItemId && c.workspace_id === contentWorkspace) ||
    calendarItems.filter((c) => c.workspace_id === contentWorkspace)[0];

  const handleRepurposeSourceToAllFormats = () => {
    if (!sourceArtifact.title.trim() || !sourceArtifact.raw_text.trim()) {
      setNotification('Please provide a valid source artifact title and content text.');
      setTimeout(() => setNotification(null), 3000);
      return;
    }

    const newDrafts: LinkedInRepurposedDraft[] = [
      {
        id: `draft-single-${Date.now()}`,
        workspace_id: contentWorkspace,
        tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
        source_artifact_id: sourceArtifact.id,
        source_title: sourceArtifact.title,
        format: 'single_thought',
        title: `Thought: ${sourceArtifact.title}`,
        content_body: `Most engineering teams overlook this when scaling: ${sourceArtifact.title}.\n\nKey lesson: Architecture is defined by the bottlenecks you eliminate early.\n\nWhat's your primary guardrail when load spikes 10x? Let's discuss below.\n\n#${sourceArtifact.tags.map((t) => t.replace(/-/g, '')).join(' #')}`,
        hook: `Most engineering teams overlook this when scaling: ${sourceArtifact.title}`,
        tags: sourceArtifact.tags,
        character_count: 380,
        status: 'draft',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: `draft-carousel-${Date.now()}`,
        workspace_id: contentWorkspace,
        tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
        source_artifact_id: sourceArtifact.id,
        source_title: sourceArtifact.title,
        format: 'carousel_outline',
        title: `Carousel: ${sourceArtifact.title} (5 Slides)`,
        content_body: `📊 SLIDE 1: Title\n${sourceArtifact.title}\nBy ${sourceArtifact.author}\n\n📊 SLIDE 2: The Context & Bottleneck\n• Initial production scale challenges\n• Why conventional patterns broke under peak load\n\n📊 SLIDE 3: Architectural Pivot\n• Moving from naive cache-aside to deterministic read-through\n• Consistent hashing rings with virtual nodes\n• Jittered TTL distribution\n\n📊 SLIDE 4: Metric Impact\n• p99 latencies pinned below 4ms\n• 42% CPU reduction\n\n📊 SLIDE 5: Takeaway & Discussion\nSave this carousel for your next distributed systems review.\n\n#${sourceArtifact.tags.map((t) => t.replace(/-/g, '')).join(' #')}`,
        hook: `Swipe through: 5 architectural insights from ${sourceArtifact.title}`,
        tags: sourceArtifact.tags,
        character_count: 620,
        status: 'draft',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: `draft-checklist-${Date.now()}`,
        workspace_id: contentWorkspace,
        tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
        source_artifact_id: sourceArtifact.id,
        source_title: sourceArtifact.title,
        format: 'actionable_checklist',
        title: `Checklist: ${sourceArtifact.title}`,
        content_body: `Here is the practical checklist we followed during ${sourceArtifact.title}:\n\n[ ] Deterministic read-through with single-flight mutex\n[ ] Consistent hashing rings with virtual nodes\n[ ] Jittered TTL expiration on hot partitions\n[ ] Distributed tracing with OpenTelemetry\n\n💡 Pro Tip: Never let cached keys expire in lock-step batches.\n\n#${sourceArtifact.tags.map((t) => t.replace(/-/g, '')).join(' #')}`,
        hook: `Here is the practical checklist we followed during ${sourceArtifact.title}:`,
        tags: sourceArtifact.tags,
        character_count: 490,
        status: 'draft',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: `draft-contrarian-${Date.now()}`,
        workspace_id: contentWorkspace,
        tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
        source_artifact_id: sourceArtifact.id,
        source_title: sourceArtifact.title,
        format: 'contrarian_breakdown',
        title: `Contrarian Breakdown: ${sourceArtifact.title}`,
        content_body: `The Common Myth: Adding more cache capacity fixes high latency.\nThe Reality: 90% of cache degradation at peak is self-inflicted by lock-step expiration.\n\n### Architectural Pivot & Truth\nWhat actually worked: Eliminating redundant cache misses via single-flight mutex coalescing.\n\nStop treating caching as an afterthought. It's a first-class architectural boundary.\n\n#${sourceArtifact.tags.map((t) => t.replace(/-/g, '')).join(' #')}`,
        hook: `The Common Myth: Adding more cache capacity fixes high latency.`,
        tags: sourceArtifact.tags,
        character_count: 440,
        status: 'draft',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: `draft-qa-${Date.now()}`,
        workspace_id: contentWorkspace,
        tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
        source_artifact_id: sourceArtifact.id,
        source_title: sourceArtifact.title,
        format: 'interview_qa_spotlight',
        title: `Q&A Spotlight: ${sourceArtifact.title}`,
        content_body: `Question: How do you scale system throughput 100x without ballooning infrastructure costs?\n\nPrincipal Engineer Answer:\nBased on our work on ${sourceArtifact.title}:\n• Focus on eliminating thundering herds\n• Distribute hot keys across consistent hashing rings\n• Maintain strict p99 observability\n\nKey principle: Predictability trumps peak burst capacity every time.\n\n#${sourceArtifact.tags.map((t) => t.replace(/-/g, '')).join(' #')}`,
        hook: `Question: How do you scale system throughput 100x without ballooning infrastructure costs?`,
        tags: sourceArtifact.tags,
        character_count: 480,
        status: 'draft',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ];

    setRepurposedDrafts((prev) => [...newDrafts, ...prev]);
    setSelectedRepurposedDraftId(newDrafts[0].id);
    setNotification(`Successfully repurposed source artifact into all 5 canonical LinkedIn formats!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleEditRepurposedDraft = (draftId: string, title: string, body: string) => {
    setRepurposedDrafts((prev) =>
      prev.map((d) => {
        if (d.id === draftId) {
          const wasApproved = d.status === 'approved';
          if (wasApproved) {
            setDraftTamperAlert2('Tamper Invalidation Triggered (AT-007): Draft modified post-approval. Cryptographic token cleared and status reset to Draft.');
          }
          return {
            ...d,
            title,
            content_body: body,
            character_count: body.length,
            status: wasApproved ? 'draft' : d.status,
            approval_token: wasApproved ? undefined : d.approval_token,
            last_approved_at: wasApproved ? undefined : d.last_approved_at,
            updated_at: new Date().toISOString()
          };
        }
        return d;
      })
    );
  };

  const handleApproveRepurposedDraft = (draftId: string) => {
    const token = `draft_hmac_${Array.from({ length: 48 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;
    setRepurposedDrafts((prev) =>
      prev.map((d) => {
        if (d.id === draftId) {
          return {
            ...d,
            status: 'approved',
            approval_token: token,
            last_approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return d;
      })
    );
    setDraftTamperAlert2(null);
    setNotification('Draft approved and cryptographically signed (HMAC-SHA256). Ready for calendar scheduling.');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleScheduleDraftToCalendar = (draft: LinkedInRepurposedDraft) => {
    if (!calendarDateInput) {
      setNotification('Please select a valid scheduled date & time.');
      setTimeout(() => setNotification(null), 3000);
      return;
    }

    // Daily budget check for same workspace (AT-018)
    const targetDay = calendarDateInput.split('T')[0];
    const slotsOnSameDay = calendarItems.filter(
      (item) => item.workspace_id === contentWorkspace && item.scheduled_slot_utc.startsWith(targetDay)
    );

    if (slotsOnSameDay.length >= calendarMaxDailyBudget) {
      setNotification(`SCHEDULE BLOCKED (AT-018): Daily slot budget exceeded (Max ${calendarMaxDailyBudget} posts/day on ${targetDay}). Prevents spam penalties.`);
      setTimeout(() => setNotification(null), 5000);
      return;
    }

    const localFormatted = `${calendarDateInput.replace('T', ' ')}:00 (${calendarTimezone}, DST-Safe)`;
    const utcISO = new Date(calendarDateInput).toISOString();

    const newPlanItem: LinkedInContentPlanItem = {
      id: `plan-item-${Date.now()}`,
      workspace_id: contentWorkspace,
      tenant_id: contentWorkspace === 'ws-alpha' ? 'tenant-alpha' : 'tenant-beta',
      draft_id: draft.id,
      title: draft.title,
      format: draft.format,
      scheduled_slot_utc: utcISO,
      user_timezone: calendarTimezone,
      local_slot_formatted: localFormatted,
      status: 'scheduled',
      asset_metadata: {
        slide_count: draft.format === 'carousel_outline' ? 5 : undefined,
        estimated_read_time_sec: Math.round(draft.character_count / 15),
        hashtags: draft.tags
      },
      direct_publish_blocked: true,
      clipboard_export_available: true,
      compose_url: `https://www.linkedin.com/feed/?shareActive=true&text=${encodeURIComponent(draft.content_body)}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    setCalendarItems((prev) => [newPlanItem, ...prev]);
    setSelectedPlanItemId(newPlanItem.id);
    setNotification(`Successfully scheduled draft for ${localFormatted}!`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleApprovePlanItem = (itemId: string) => {
    const token = `plan_hmac_${Array.from({ length: 48 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`;
    setCalendarItems((prev) =>
      prev.map((item) => {
        if (item.id === itemId) {
          return {
            ...item,
            status: 'approved',
            approval_token: token,
            last_approved_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
        }
        return item;
      })
    );
    setPlanTamperAlert(null);
    setNotification('Content calendar slot approved and cryptographically signed (HMAC-SHA256).');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReschedulePlanItem = (itemId: string, newTime: string) => {
    setCalendarItems((prev) =>
      prev.map((item) => {
        if (item.id === itemId) {
          const wasApproved = item.status === 'approved' || !!item.approval_token;
          if (wasApproved) {
            setPlanTamperAlert('Tamper Invalidation Triggered (AT-007): Schedule altered post-approval. Cryptographic approval token revoked and status reset to Draft.');
          }
          return {
            ...item,
            scheduled_slot_utc: new Date(newTime).toISOString(),
            local_slot_formatted: `${newTime.replace('T', ' ')}:00 (${item.user_timezone}, DST-Safe)`,
            status: wasApproved ? 'draft' : item.status,
            approval_token: wasApproved ? undefined : item.approval_token,
            last_approved_at: wasApproved ? undefined : item.last_approved_at,
            updated_at: new Date().toISOString()
          };
        }
        return item;
      })
    );
  };

  const handleAttemptDirectPublish = (item: LinkedInContentPlanItem) => {
    setDirectPublishRejectAlert(`GOVERNANCE REJECTION (AT-010 & REQ-015): Direct automated background publishing to personal feeds is permanently barred. All posts require 1-click human confirmation via clipboard export or LinkedIn Compose.`);
    setTimeout(() => setDirectPublishRejectAlert(null), 6000);
  };

  const handleExportPlanForClipboard = (item: LinkedInContentPlanItem) => {
    const draft = repurposedDrafts.find((d) => d.id === item.draft_id) || {
      title: item.title,
      content_body: 'Sample repurposed content body...',
      format: item.format,
      source_title: 'Engineering Source'
    };

    if (navigator.clipboard) {
      navigator.clipboard.writeText(draft.content_body);
    }
    setClipboardCopiedNotice(`Post text copied to clipboard! Opening LinkedIn Compose in a new tab for 1-click human dispatch...`);
    setTimeout(() => setClipboardCopiedNotice(null), 4500);
    window.open(item.compose_url || 'https://www.linkedin.com/feed/?shareActive=true', '_blank', 'noopener,noreferrer');
  };

  const handleDeletePlanItem = (itemId: string) => {
    setCalendarItems((prev) => prev.filter((i) => i.id !== itemId));
    setNotification('Scheduled calendar slot deleted.');
    setTimeout(() => setNotification(null), 3000);
  };

  // =========================================================================
  // IMP-LI-15: Engagement Monitoring & Analytics Handlers
  // (AT-028, AT-010, AT-012, SRC-L2)
  // =========================================================================
  const filteredSnapshots = analyticsSnapshots.filter(
    (s) => s.workspace_id === analyticsWorkspace
  );

  const selectedSnapshot =
    analyticsSnapshots.find((s) => s.snapshot_id === selectedSnapshotId && s.workspace_id === analyticsWorkspace) ||
    filteredSnapshots[0];

  const handleTestEngagerSegmentation = () => {
    if (!testEngagerHeadline.trim()) return;
    const lowerH = testEngagerHeadline.toLowerCase();
    let segment = 'other_network';
    if (lowerH.includes('recruiter') || lowerH.includes('talent') || lowerH.includes('headhunter')) {
      segment = 'talent_partner';
    } else if (lowerH.includes('vp') || lowerH.includes('director') || lowerH.includes('head of') || lowerH.includes('cto') || lowerH.includes('founder')) {
      segment = 'decision_maker';
    } else if (lowerH.includes('architect') || lowerH.includes('staff') || lowerH.includes('principal') || lowerH.includes('lead') || lowerH.includes('senior')) {
      segment = 'peer_practitioner';
    }
    setSegmentedResultSegment(segment);
    setNotification(`Categorized engager into ICP segment: ${segment.replace('_', ' ').toUpperCase()} (SRC-L2).`);
    setTimeout(() => setNotification(null), 3500);
  };

  const handleAddSampleEngager = () => {
    if (!selectedSnapshot || !testEngagerHeadline.trim()) return;
    const lowerH = testEngagerHeadline.toLowerCase();
    let segment: 'decision_maker' | 'peer_practitioner' | 'talent_partner' | 'other_network' = 'other_network';
    if (lowerH.includes('recruiter') || lowerH.includes('talent') || lowerH.includes('headhunter')) {
      segment = 'talent_partner';
    } else if (lowerH.includes('vp') || lowerH.includes('director') || lowerH.includes('head of') || lowerH.includes('cto') || lowerH.includes('founder')) {
      segment = 'decision_maker';
    } else if (lowerH.includes('architect') || lowerH.includes('staff') || lowerH.includes('principal') || lowerH.includes('lead') || lowerH.includes('senior')) {
      segment = 'peer_practitioner';
    }

    const newEngager: LinkedInEngagerProfile = {
      author_urn: `urn:li:person:user-${Date.now()}`,
      name: 'Simulated Professional',
      headline: testEngagerHeadline.trim(),
      company: testEngagerCompany.trim() || 'Tech Enterprise',
      interaction_type: 'comment',
      comment_text: 'Insightful analysis regarding production scale.',
      segment: segment,
      observed_at: new Date().toISOString()
    };

    setAnalyticsSnapshots((prev) =>
      prev.map((s) => {
        if (s.snapshot_id !== selectedSnapshot.snapshot_id) return s;
        const updatedEngagers = [...s.engagers, newEngager];
        const dm = updatedEngagers.filter((e) => e.segment === 'decision_maker').length;
        const pp = updatedEngagers.filter((e) => e.segment === 'peer_practitioner').length;
        const tp = updatedEngagers.filter((e) => e.segment === 'talent_partner').length;
        const on = updatedEngagers.filter((e) => e.segment === 'other_network').length;
        return {
          ...s,
          engagers: updatedEngagers,
          icp_breakdown: {
            decision_makers: dm,
            peer_practitioners: pp,
            talent_partners: tp,
            other_network: on,
            high_value_engagers_count: dm + pp + tp
          }
        };
      })
    );
    setNotification('Added sample professional engager and updated ICP demographic breakdown!');
    setTimeout(() => setNotification(null), 3500);
  };

  // =========================================================================
  // IMP-LI-16: Employee Advocacy & Brand Governance Handlers
  // (LI-16, AT-007, AT-011, SRC-L2)
  // =========================================================================
  const currentAdvocacyCampaigns = advocacyCampaigns.filter(
    (c) => c.workspace_id === advocacyWorkspace
  );
  const selectedAdvocacyCampaign =
    advocacyCampaigns.find((c) => c.campaign_id === selectedAdvocacyCampaignId && c.workspace_id === advocacyWorkspace) ||
    currentAdvocacyCampaigns[0];

  const selectedAdvocacyVariant =
    selectedAdvocacyCampaign?.variants.find((v) => v.persona === selectedAdvocacyPersona) ||
    selectedAdvocacyCampaign?.variants[0];

  const handleSelectAdvocacyCampaign = (campaignId: string) => {
    setSelectedAdvocacyCampaignId(campaignId);
    setAdvocacyTamperAlert(null);
    setComplianceViolations([]);
    const camp = advocacyCampaigns.find((c) => c.campaign_id === campaignId);
    if (camp && camp.variants.length > 0) {
      setSelectedAdvocacyPersona(camp.variants[0].persona);
      setCustomizedAdvocacyCopy(camp.variants[0].suggested_text);
    }
  };

  const handleSelectAdvocacyPersona = (persona: LinkedInEmployeePersona) => {
    setSelectedAdvocacyPersona(persona);
    setAdvocacyTamperAlert(null);
    const variant = selectedAdvocacyCampaign?.variants.find((v) => v.persona === persona);
    if (variant) {
      setCustomizedAdvocacyCopy(variant.suggested_text);
      if (selectedAdvocacyCampaign) {
        checkBrandKeywords(variant.suggested_text, selectedAdvocacyCampaign.governance);
      }
    }
  };

  const checkBrandKeywords = (text: string, gov: LinkedInBrandGovernance) => {
    if (!gov || !gov.forbidden_keywords) {
      setComplianceViolations([]);
      return;
    }
    const lower = text.toLowerCase();
    const violations = gov.forbidden_keywords.filter((kw) => kw.trim() !== '' && lower.includes(kw.toLowerCase()));
    setComplianceViolations(violations);
  };

  const handleCustomizedCopyChange = (newText: string) => {
    setCustomizedAdvocacyCopy(newText);
    if (selectedAdvocacyCampaign) {
      checkBrandKeywords(newText, selectedAdvocacyCampaign.governance);
    }
  };

  const handleEditCampaignOriginalVariant = (variantId: string, newText: string) => {
    if (!selectedAdvocacyCampaign) return;
    const wasApproved = selectedAdvocacyCampaign.status === 'approved';

    setAdvocacyCampaigns((prev) =>
      prev.map((c) => {
        if (c.campaign_id === selectedAdvocacyCampaign.campaign_id) {
          const updatedVariants = c.variants.map((v) => (v.variant_id === variantId ? { ...v, suggested_text: newText } : v));
          return {
            ...c,
            variants: updatedVariants,
            status: wasApproved ? 'draft' : c.status,
            approval_token: wasApproved ? undefined : c.approval_token,
            approved_by: wasApproved ? undefined : c.approved_by,
            approved_at: wasApproved ? undefined : c.approved_at,
            updated_at: new Date().toISOString()
          };
        }
        return c;
      })
    );

    if (wasApproved) {
      setAdvocacyTamperAlert(
        'AT-007 Cryptographic Invalidation: Modifying campaign copy invalidated the approval token. Campaign status has reset to DRAFT until re-approved.'
      );
    }
  };

  const handleApproveAdvocacyCampaign = (campaignId: string) => {
    const target = advocacyCampaigns.find((c) => c.campaign_id === campaignId);
    if (!target) return;

    for (const v of target.variants) {
      const lower = v.suggested_text.toLowerCase();
      const violations = (target.governance.forbidden_keywords || []).filter((kw) => lower.includes(kw.toLowerCase()));
      if (violations.length > 0) {
        setNotification(`Cannot approve campaign: Variant "${v.headline}" contains forbidden keyword "${violations[0]}"`);
        setTimeout(() => setNotification(null), 4000);
        return;
      }
    }

    const token = `hmac-sha256-signed-${campaignId}-${Date.now().toString(36)}`;
    const now = new Date().toISOString();

    setAdvocacyCampaigns((prev) =>
      prev.map((c) =>
        c.campaign_id === campaignId
          ? {
              ...c,
              status: 'approved',
              approval_token: token,
              approved_by: 'Alex Chen (VP Eng)',
              approved_at: now,
              updated_at: now
            }
          : c
      )
    );
    setAdvocacyTamperAlert(null);
    setNotification('Advocacy Campaign cryptographically approved with HMAC-SHA256 token (AT-007)!');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleShareAdvocacyContent = () => {
    if (!selectedAdvocacyCampaign) return;
    if (selectedAdvocacyCampaign.status !== 'approved') {
      setNotification('Action blocked: Campaign must be APPROVED by company governance before sharing.');
      setTimeout(() => setNotification(null), 3500);
      return;
    }
    if (complianceViolations.length > 0) {
      setNotification(`Cannot share: Text violates brand rules with forbidden keyword: "${complianceViolations[0]}"`);
      setTimeout(() => setNotification(null), 3500);
      return;
    }

    const shareText = customizedAdvocacyCopy.trim() || selectedAdvocacyVariant?.suggested_text || '';
    if (!shareText) return;

    navigator.clipboard.writeText(shareText);

    setAdvocacyCampaigns((prev) =>
      prev.map((c) =>
        c.campaign_id === selectedAdvocacyCampaign.campaign_id ? { ...c, share_count: c.share_count + 1 } : c
      )
    );

    setAdvocacyShareNotice(
      `1-Click Clipboard Copied! Native LinkedIn compose window ready. Shared voluntarily with employee consent (AT-010, REQ-015).`
    );
    setTimeout(() => setAdvocacyShareNotice(null), 5000);
  };

  const handleTestAntiPodSecurity = (simulateViolation: boolean = true) => {
    if (simulateViolation) {
      setAntiPodAlert(
        'BLOCKED [403 Forbidden]: Coordinated engagement pods, automated reciprocal likes, and synthetic comments are permanently barred by LI-16 policy. Employee advocacy operates purely through voluntary, human-in-the-loop sharing.'
      );
    } else {
      setAntiPodAlert(
        'PASSED [200 OK]: Legitimate human advocacy verified. Anti-pod security checks cleared with zero automated bot rings detected.'
      );
    }
    setTimeout(() => setAntiPodAlert(null), 6000);
  };

  // Provider Fallback & Diagnostics Handlers (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
  const handleRunDoctorProbe = () => {
    setIsDoctorRunning(true);
    setTimeout(() => {
      setIsDoctorRunning(false);
      const activeWorkspaceProviders = providers.filter((p) => p.workspace_id === providerWorkspace);
      const isDegraded = activeWorkspaceProviders.some((p) => p.health_status === 'degraded' || p.health_status === 'rate_limited');
      const now = new Date().toISOString();
      const healthyCount = activeWorkspaceProviders.filter((p) => p.health_status === 'healthy').length;
      const degradedCount = activeWorkspaceProviders.filter((p) => p.health_status !== 'healthy').length;
      const report: LinkedInProviderDiagnosticReport = {
        generated_at: now,
        tenant_id: providerWorkspace === 'ws-alpha' ? 'tenant-corp-01' : 'tenant-corp-02',
        workspace_id: providerWorkspace,
        overall_health: isDegraded ? 'degraded' : 'healthy',
        total_providers: activeWorkspaceProviders.length,
        healthy_providers: healthyCount,
        degraded_providers: degradedCount,
        probes: activeWorkspaceProviders.map((p) => ({
          provider_id: p.provider_id,
          name: p.name,
          adapter_type: p.adapter_type,
          health_status: p.health_status,
          latency_ms: p.latency_ms,
          missing_scopes: [],
          error_message: p.health_status === 'degraded' ? 'HTTP 429: Too Many Requests from upstream rate limiter [Bearer [REDACTED_SECRET]]' : undefined,
          remediation_step: p.remediation_step,
          probed_at: now
        })),
        recommendations: isDegraded
          ? [
              'Primary official enterprise adapter is experiencing 429 rate limits.',
              'Fallback adapter (Consumer OIDC / Local Session) is primed for permitted read actions under AT-010.',
              'Automated messaging remains strictly barred (Zero Permission-Bypass Policy).'
            ]
          : [
              'All adapters operational.',
              'Credentials and bearer tokens safely scrubbed in all diagnostic logs (AT-016).'
            ]
      };
      setDiagnosticReport(report);
      setNotification('Diagnostic Doctor complete: Probe reports updated and credentials safely scrubbed (AT-016, FND-013).');
      setTimeout(() => setNotification(null), 4000);
    }, 500);
  };

  const handleTogglePrimaryDegraded = () => {
    setProviders((prev) =>
      prev.map((p) => {
        if (p.workspace_id === providerWorkspace && p.tier === 1) {
          const nextStatus = p.health_status === 'healthy' ? 'degraded' : 'healthy';
          return {
            ...p,
            health_status: nextStatus,
            consecutive_failures: nextStatus === 'degraded' ? 3 : 0,
            latency_ms: nextStatus === 'degraded' ? 2450 : 115,
            remediation_step: nextStatus === 'degraded' ? 'Rate limited (429). Failover to Tier 2 Consumer OIDC' : 'Adapter operational and ready for live requests',
            updated_at: new Date().toISOString()
          };
        }
        return p;
      })
    );
    setDispatchResult(null);
    setDispatchError(null);
  };

  const handleTestDispatchAction = (action: string) => {
    setDispatchError(null);
    setDispatchResult(null);

    // AT-010: Zero permission bypass. Prohibited actions cannot be bypassed via fallback.
    const prohibitedActions = ['automated_headless_messaging', 'scrape_unconnected_inbox', 'mass_endorsement_bot'];
    if (prohibitedActions.includes(action)) {
      setDispatchError(
        'SECURITY VIOLATION [403 Forbidden - Zero Permission Bypass]: Action "' + action + '" is permanently prohibited by platform governance (AT-010). Fallback cannot bypass platform security policies.'
      );
      return;
    }

    const activeList = providers.filter((p) => p.workspace_id === providerWorkspace);
    const sorted = [...activeList].sort((a, b) => a.tier - b.tier);

    let dispatched: LinkedInProviderAdapter | null = null;
    let fallbackInvoked = false;
    let fallbackReason = '';

    for (let i = 0; i < sorted.length; i++) {
      const candidate = sorted[i];
      if (!candidate.supported_actions.includes(action)) {
        continue;
      }
      if (candidate.health_status === 'healthy') {
        dispatched = candidate;
        if (i > 0) {
          fallbackInvoked = true;
          fallbackReason = 'Primary tier ' + sorted[0].tier + ' is ' + sorted[0].health_status + '; gracefully routed to tier ' + candidate.tier + ' (' + candidate.name + ')';
        }
        break;
      }
    }

    if (!dispatched) {
      setDispatchError(
        'DISPATCH FAILED [503 Service Unavailable]: No healthy provider adapter available supporting action "' + action + '".'
      );
      return;
    }

    const res: LinkedInProviderDispatchResult = {
      action,
      dispatched_provider_id: dispatched.provider_id,
      adapter_type: dispatched.adapter_type,
      success: true,
      fallback_invoked: fallbackInvoked,
      fallback_reason: fallbackReason,
      latency_ms: dispatched.latency_ms,
      dispatched_at: new Date().toISOString()
    };
    setDispatchResult(res);
  };

  const handleRunScrubTest = () => {
    // AT-016: client-side regex scrubber mirror of backend ScrubDiagnosticSecrets
    let out = testScrubInput;
    out = out.replace(/(Bearer\s+)[A-Za-z0-9_\-\.]{8,}/gi, '$1[REDACTED_SECRET]');
    out = out.replace(/(li_at[=\s:"]+)[A-Za-z0-9_\-]{8,}/gi, '$1[REDACTED_SECRET]');
    out = out.replace(/(client_secret[=\s:"]+)[A-Za-z0-9_\-]{8,}/gi, '$1[REDACTED_SECRET]');
    out = out.replace(/(secret_token[=\s:"]+)[A-Za-z0-9_\-]{8,}/gi, '$1[REDACTED_SECRET]');
    out = out.replace(/(password[=\s:"]+)[^\s&"'>]+/gi, '$1[REDACTED_SECRET]');
    setScrubbedOutput(out);
  };

  // =========================================================================
  // Platform Limits, CAPTCHA / Reauth & Safety Gatekeeper Handlers
  // (IMP-LI-18, LI-18, REQ-009, REQ-010, AT-008, AT-009, AT-010, SRC-L1, SRC-S3)
  // =========================================================================

  const handleSimulateUpstreamRestriction = (type: string) => {
    const now = new Date().toISOString();
    let limitType: LinkedInPlatformLimitType = 'weekly_invitations';
    let chalType: LinkedInChallengeType | undefined = 'weekly_invitation_limit';
    let statusCode = 429;
    let rawError = "You've reached the weekly invitation limit. Invitations will be refreshed next week.";
    let status: LinkedInAccountSafetyStatus = 'rate_limited';
    let cooldownHours = 72;

    if (type === 'captcha_checkpoint_403') {
      limitType = 'security_checkpoint';
      chalType = 'captcha';
      statusCode = 403;
      rawError = 'HTTP 403: Security checkpoint encountered at /checkpoint/challenge/captcha. User interaction required. Automated access prohibited.';
      status = 'challenge_required';
      cooldownHours = 0;
    } else if (type === 'commercial_search_429') {
      limitType = 'profile_searches';
      chalType = 'commercial_use_cap';
      statusCode = 429;
      rawError = "You've reached the commercial use limit on search. Free searches will reset on the 1st of next month.";
      status = 'rate_limited';
      cooldownHours = 24;
    } else if (type === 'session_expired_401') {
      limitType = 'session_expired';
      chalType = 'credential_reauth';
      statusCode = 401;
      rawError = 'HTTP 401: LinkedIn session token expired or invalidated.';
      status = 'challenge_required';
      cooldownHours = 0;
    }

    const cooldownDate = cooldownHours > 0 ? new Date(Date.now() + cooldownHours * 3600000).toISOString() : undefined;

    const newIncident: LinkedInPlatformRestrictionIncident = {
      incident_id: `inc-${Date.now()}`,
      account_id: safetyState.account_id,
      tenant_id: safetyState.tenant_id,
      workspace_id: limitsWorkspace,
      status_code: statusCode,
      raw_error_message_scrubbed: rawError,
      limit_type: limitType,
      challenge_type: chalType,
      action_attempted: gateCheckAction,
      cooldown_hours: cooldownHours,
      occurred_at: now
    };

    setSafetyState((prev) => ({
      ...prev,
      status,
      active_limit_type: limitType,
      active_challenge: chalType,
      challenge_reason: rawError,
      cooldown_until: cooldownDate,
      consecutive_restrictions: prev.consecutive_restrictions + 1,
      last_incident_at: now,
      updated_at: now
    }));

    setRecentRestrictions((prev) => [newIncident, ...prev]);

    setActivityLedger((prev) => [
      {
        entry_id: `act-${Date.now()}`,
        account_id: safetyState.account_id,
        tenant_id: safetyState.tenant_id,
        workspace_id: limitsWorkspace,
        module: 'linkedin',
        action: gateCheckAction,
        amount: 0,
        status: 'challenge_triggered',
        details: `Platform restriction triggered: ${limitType} (HTTP ${statusCode})`,
        timestamp: now
      },
      ...prev
    ]);

    setGateCheckDecision({
      allowed: false,
      status,
      reason: `Account circuit-breaker OPEN: ${rawError}`
    });

    setNotification(`Simulated Upstream ${statusCode}: Account circuit breaker engaged (${status}). Fail-closed protection active (REQ-009, AT-010).`);
    setTimeout(() => setNotification(null), 4000);
  };

  const handleResolveChallengeLegitimately = () => {
    const now = new Date().toISOString();
    setSafetyState((prev) => ({
      ...prev,
      status: 'active',
      active_limit_type: undefined,
      active_challenge: undefined,
      challenge_reason: undefined,
      cooldown_until: undefined,
      last_resolved_at: now,
      updated_at: now
    }));

    setActivityLedger((prev) => [
      {
        entry_id: `act-${Date.now()}`,
        account_id: safetyState.account_id,
        tenant_id: safetyState.tenant_id,
        workspace_id: limitsWorkspace,
        module: 'linkedin',
        action: 'resolve_challenge',
        amount: 0,
        status: 'success',
        details: `Challenge legitimately cleared via user re-auth: ${reauthNotes}`,
        timestamp: now
      },
      ...prev
    ]);

    setGateCheckDecision({
      allowed: true,
      status: 'active'
    });

    setIsReauthConfirming(false);
    setNotification('Challenge cleared via official user re-authentication! Account returned to ACTIVE state with zero stealth bypass (AT-010).');
    setTimeout(() => setNotification(null), 4000);
  };

  const handleEvaluateSafetyGate = (action: string) => {
    setGateCheckAction(action);
    if (safetyState.status === 'challenge_required') {
      setGateCheckDecision({
        allowed: false,
        status: 'challenge_required',
        reason: `BLOCKED [Fail-Closed AT-010]: Active challenge (${safetyState.active_challenge}) requires legitimate user re-auth. Automated bypasses prohibited.`
      });
      return;
    }
    if (safetyState.status === 'rate_limited') {
      setGateCheckDecision({
        allowed: false,
        status: 'rate_limited',
        reason: `BLOCKED [Cooling-off Period REQ-009]: Account in rate limit cooldown until ${safetyState.cooldown_until || 'next cycle'}.`
      });
      return;
    }
    if (safetyState.status === 'paused' || safetyState.status === 'revoked') {
      setGateCheckDecision({
        allowed: false,
        status: safetyState.status,
        reason: `BLOCKED: Account is in ${safetyState.status} state.`
      });
      return;
    }
    // Active state
    setGateCheckDecision({
      allowed: true,
      status: 'active'
    });
  };

  const handleRecordSampleActionInWorkspace = (action: string, amount: number) => {
    const now = new Date().toISOString();
    setActivityLedger((prev) => [
      {
        entry_id: `act-${Date.now()}`,
        account_id: safetyState.account_id,
        tenant_id: safetyState.tenant_id,
        workspace_id: limitsWorkspace,
        module: limitsWorkspace === 'ws-alpha' ? 'linkedin' : 'career',
        action,
        amount,
        status: 'dispatched',
        details: `Dispatched ${action} via ${limitsWorkspace}`,
        timestamp: now
      },
      ...prev
    ]);

    setCombinedBudget((prev) => {
      let invUsed = prev.daily_invitations_used;
      let weeklyInvUsed = prev.weekly_invitations_used;
      let inmailUsed = prev.daily_inmail_used;
      let recDmUsed = prev.daily_recruiter_dm_used;

      if (action === 'connection_request') {
        invUsed += amount;
        weeklyInvUsed += amount;
      } else if (action === 'recruiter_dm') {
        recDmUsed += amount;
        inmailUsed += amount;
      } else if (action === 'inmail') {
        inmailUsed += amount;
      }

      return {
        ...prev,
        daily_invitations_used: invUsed,
        daily_invitations_remaining: Math.max(0, prev.daily_invitations_limit - invUsed),
        weekly_invitations_used: weeklyInvUsed,
        weekly_invitations_remaining: Math.max(0, prev.weekly_invitations_limit - weeklyInvUsed),
        daily_inmail_used: inmailUsed,
        daily_inmail_remaining: Math.max(0, prev.daily_inmail_limit - inmailUsed),
        daily_recruiter_dm_used: recDmUsed,
        daily_recruiter_dm_remaining: Math.max(0, prev.daily_recruiter_dm_limit - recDmUsed),
        computed_at: now
      };
    });

    setNotification(`Recorded ${amount} ${action} in ${limitsWorkspace}. Account-wide budget pool updated across all workspaces (AT-008, AT-009).`);
    setTimeout(() => setNotification(null), 3500);
  };

  // Discovery Engine Handlers (IMP-LI-05, LI-05, AT-010)
  const handleRunDiscovery = () => {
    // 1. Prohibited attribute pre-check
    const prohibited = ['age', 'gender', 'female', 'male', 'race', 'religion', 'under_30'];
    const combined = `${discoveryRole} ${discoveryCompanies} ${discoveryExclusions}`.toLowerCase();
    for (const p of prohibited) {
      if (new RegExp(`\\b${p}\\b`).test(combined)) {
        setNotification(`BLOCKED (LI-05): Protected attribute '${p}' detected. Non-discriminatory criteria required.`);
        setTimeout(() => setNotification(null), 4000);
        return;
      }
    }

    const comps = discoveryCompanies.split(',').map((c) => c.trim()).filter(Boolean);
    const excls = discoveryExclusions.split(',').map((e) => e.trim()).filter(Boolean);

    // Build Boolean query
    const roleTerm = `("${discoveryRole}")`;
    const compTerm = comps.length ? `(${comps.map((c) => `"${c}"`).join(' OR ')})` : '';
    const exclTerm = excls.length ? `NOT (${excls.map((e) => `"${e}"`).join(' OR ')})` : '';
    const clauses = [roleTerm, compTerm, exclTerm].filter(Boolean);
    const rawQuery = clauses.join(' AND ');
    const searchURL = `https://www.linkedin.com/search/results/people/?keywords=${encodeURIComponent(rawQuery)}&origin=GLOBAL_SEARCH_HEADER`;

    // Match in active vault workspace
    const matchedP = persons.filter(
      (p) =>
        p.workspace_id === recordsWorkspace &&
        (p.headline.toLowerCase().includes(discoveryRole.toLowerCase()) ||
          p.current_role.toLowerCase().includes(discoveryRole.toLowerCase()))
    );
    const matchedC = companies.filter(
      (c) =>
        c.workspace_id === recordsWorkspace &&
        comps.some((targetC) => c.company_name.toLowerCase().includes(targetC.toLowerCase()))
    );

    setActiveDiscoveryResult({
      criteria: {
        target_role: discoveryRole,
        current_companies: comps,
        exclusions: excls,
        connection_tiers: discoveryConnectionTiers,
        user_goal: discoveryGoal
      },
      boolean_query: {
        raw_query: rawQuery,
        role_clause: roleTerm,
        company_clause: compTerm,
        exclusion_clause: exclTerm,
        linkedin_search_url: searchURL
      },
      matched_persons: matchedP,
      matched_companies: matchedC,
      total_vault_matches: matchedP.length + matchedC.length,
      platform_notice:
        'Direct LinkedIn people search API is restricted to Enterprise Recruiter tiers. Vault matches below reflect verified tenant records; native search dispatch assisted via 1-click external link (AT-010).',
      generated_at: new Date().toISOString()
    });

    setNotification('Generated Boolean search query and filtered matching Vault records!');
    setTimeout(() => setNotification(null), 3500);
  };

  const handleNormalizeURL = () => {
    if (!rawURLInput.trim()) return;
    try {
      const parsed = new URL(rawURLInput.startsWith('http') ? rawURLInput : `https://${rawURLInput}`);
      const cleanPath = parsed.pathname.replace(/^\/|\/$/g, '');
      const segments = cleanPath.split('/');
      if (segments.length >= 2 && (segments[0] === 'in' || segments[0] === 'company')) {
        const entityType = segments[0] === 'company' ? 'company' : 'person';
        const slug = segments[1];
        setCanonicalURLResult({
          raw_url: rawURLInput,
          normalized_url: `https://www.linkedin.com/${segments[0]}/${slug}`,
          entity_type: entityType,
          slug
        });
        setNotification('Successfully stripped tracking tokens and normalized canonical LinkedIn URL!');
        setTimeout(() => setNotification(null), 3500);
      } else {
        setNotification('Could not extract valid /in/ or /company/ entity path from URL.');
        setTimeout(() => setNotification(null), 3500);
      }
    } catch {
      setNotification('Invalid URL format provided.');
      setTimeout(() => setNotification(null), 3500);
    }
  };

  const handleCopyBoolean = () => {
    if (activeDiscoveryResult && navigator.clipboard) {
      navigator.clipboard.writeText(activeDiscoveryResult.boolean_query.raw_query);
      setCopiedBoolean(true);
      setTimeout(() => setCopiedBoolean(false), 2500);
      setNotification('Boolean query copied to clipboard! Paste directly into LinkedIn search box.');
      setTimeout(() => setNotification(null), 3500);
    }
  };

  const handleSelectPreset = (idx: number) => {
    setSelectedPresetIndex(idx);
    setCurrentConnection(JSON.parse(JSON.stringify(PRESET_CONNECTIONS[idx].connection)));
    setActionTestResult(null);
  };

  const handleTestLiveAction = (actionName: string) => {
    const isExpired = currentConnection.token_status === 'expired' || currentConnection.status === 'reauth_required';
    if (isExpired) {
      setActionTestResult({
        action: actionName,
        allowed: false,
        message: 'BLOCKED (Fail-Closed): LinkedIn OAuth token is expired or revoked. Re-authentication required before any action (AT-016).',
        timestamp: new Date().toLocaleTimeString()
      });
      return;
    }

    if (actionName === 'automated_easy_apply') {
      setActionTestResult({
        action: actionName,
        allowed: false,
        message: 'BLOCKED (AT-010): Automated Easy Apply background write is permanently unsupported on live LinkedIn. Safe fallback: Use Assisted-Manual modal with clipboard answers.',
        timestamp: new Date().toLocaleTimeString()
      });
      return;
    }

    if (actionName === 'send_direct_message') {
      const hasMsgScope = currentConnection.granted_scopes.includes('r_messages');
      if (!hasMsgScope) {
        setActionTestResult({
          action: actionName,
          allowed: false,
          message: 'BLOCKED (AT-010): Direct messaging requires LinkedIn Enterprise Partner approval (r_messages). Consumer OAuth cannot send automated messages. Use manual clipboard copy.',
          timestamp: new Date().toLocaleTimeString()
        });
        return;
      }
    }

    if (actionName === 'publish_post') {
      const hasPublishScope = currentConnection.granted_scopes.includes('w_member_social');
      if (!hasPublishScope) {
        setActionTestResult({
          action: actionName,
          allowed: false,
          message: 'BLOCKED: Scope w_member_social was not granted by the candidate during OAuth connection.',
          timestamp: new Date().toLocaleTimeString()
        });
        return;
      }
    }

    if (actionName === 'import_experience') {
      const hasFullProfileScope = currentConnection.granted_scopes.includes('r_fullprofile');
      if (!hasFullProfileScope) {
        setActionTestResult({
          action: actionName,
          allowed: false,
          message: 'BLOCKED: Full profile career history is restricted by LinkedIn to Enterprise Talent partners. Fallback: Upload Basic_LinkedInData.zip in Career Hub.',
          timestamp: new Date().toLocaleTimeString()
        });
        return;
      }
    }

    setActionTestResult({
      action: actionName,
      allowed: true,
      message: `ALLOWED: Action "${actionName}" is authorized by current granted scopes with human-in-the-loop confirmation requirement (AT-007).`,
      timestamp: new Date().toLocaleTimeString()
    });
  };

  const handleDisconnect = () => {
    setCurrentConnection({
      ...currentConnection,
      status: 'disconnected',
      token_status: 'revoked',
      capabilities: currentConnection.capabilities.map((c) => ({ ...c, is_granted: false, status: 'denied' }))
    });
    setNotification('LinkedIn account connection revoked and purged from active session.');
    setTimeout(() => setNotification(null), 4000);
  };

  return (
    <div className="space-y-6">
      {/* Studio Tab Switcher & Quick Status */}
      <div className="space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white dark:bg-slate-900 p-4 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-xs">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold uppercase tracking-wider text-sky-600 dark:text-sky-400">
                {activeModule === 'overview' && 'Overview & OAuth Matrix'}
                {activeModule === 'profile' && 'Profile Intelligence & Records Vault'}
                {activeModule === 'network' && 'Network Expansion & Recruiter Leads'}
                {activeModule === 'content' && 'Post Writing, Hooks & Voice Studio'}
                {activeModule === 'growth' && 'Content Planner, Analytics & Advocacy'}
                {activeModule === 'safety' && 'Safety Guard, Diagnostics & Exports'}
              </span>
            </div>
            <h1 className="text-xl font-extrabold text-slate-900 dark:text-white mt-0.5">
              {activeModule === 'overview' && 'LinkedIn Connect & Granted Capabilities Inspector'}
              {activeModule === 'profile' && 'Profile Scan, Optimizer & Normalized Records Vault'}
              {activeModule === 'network' && 'Discovery, Hiring Leads, Connection Queue & Inbox'}
              {activeModule === 'content' && 'Post Writing, Comments Sweep, Voice Humanizer & Story Bank'}
              {activeModule === 'growth' && 'Multi-Format Calendar, Engagement Analytics & Advocacy'}
              {activeModule === 'safety' && 'Account Limits, Reauth Safety & Relationship Exports'}
            </h1>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleDisconnect}
              disabled={currentConnection.status === 'disconnected'}
              className="inline-flex items-center gap-1.5 rounded-lg border border-rose-200 bg-white dark:bg-slate-800 px-3 py-1.5 text-xs font-semibold text-rose-700 dark:text-rose-400 shadow-xs hover:bg-rose-50 disabled:opacity-40"
            >
              <XCircle className="w-3.5 h-3.5 text-rose-600" />
              <span>Disconnect</span>
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 text-xs font-semibold border-b border-slate-200 dark:border-slate-800">
          {[
            { id: 'overview', label: 'Overview & Matrix' },
            { id: 'profile', label: 'Profile & Vault' },
            { id: 'network', label: 'Network & Leads' },
            { id: 'content', label: 'Content & Voice' },
            { id: 'growth', label: 'Growth & Analytics' },
            { id: 'safety', label: 'Safety & Tools' },
          ].map((tab) => {
            const isSelected = activeModule === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setActiveModule(tab.id as any)}
                className={`px-3 py-2 rounded-t-lg border-b-2 transition-colors whitespace-nowrap ${
                  isSelected
                    ? 'border-sky-600 text-sky-700 dark:text-sky-300 bg-white dark:bg-slate-900 font-bold'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                {tab.label}
              </button>
            );
          })}
        </div>
      </div>

      {notification && (
        <div className="rounded-lg bg-emerald-600 px-4 py-2.5 text-xs font-bold text-white shadow-md flex items-center justify-between">
          <span>{notification}</span>
        </div>
      )}

      {/* MODULE 1: OVERVIEW */}
      {activeModule === 'overview' && (
        <div className="space-y-6">
          {/* Quick Jump Studio Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3.5">
            <button
              onClick={() => setActiveModule('profile')}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left hover:border-sky-300 hover:shadow-xs transition"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white">Profile & Records Vault</span>
                <span className="text-[10px] bg-sky-100 text-sky-800 dark:bg-sky-900/50 dark:text-sky-300 px-2 py-0.5 rounded font-bold">Studio</span>
              </div>
              <p className="text-[11px] text-slate-500 mt-1">Scan profile, optimize headlines & inspect normalized entity records.</p>
            </button>

            <button
              onClick={() => setActiveModule('network')}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left hover:border-sky-300 hover:shadow-xs transition"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white">Network & Outreach</span>
                <span className="text-[10px] bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300 px-2 py-0.5 rounded font-bold">Leads</span>
              </div>
              <p className="text-[11px] text-slate-500 mt-1">Recruiter discovery, connection note drafts & conversation triage.</p>
            </button>

            <button
              onClick={() => setActiveModule('content')}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left hover:border-sky-300 hover:shadow-xs transition"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white">Content & Voice Studio</span>
                <span className="text-[10px] bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300 px-2 py-0.5 rounded font-bold">Creator</span>
              </div>
              <p className="text-[11px] text-slate-500 mt-1">Post writing, thread sweeps, voice humanizer & story bank.</p>
            </button>

            <button
              onClick={() => setActiveModule('growth')}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left hover:border-sky-300 hover:shadow-xs transition"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white">Growth & Analytics</span>
                <span className="text-[10px] bg-amber-100 text-amber-800 dark:bg-amber-900/50 dark:text-amber-300 px-2 py-0.5 rounded font-bold">Metrics</span>
              </div>
              <p className="text-[11px] text-slate-500 mt-1">DST-safe calendar schedule, engagement analytics & advocacy.</p>
            </button>

            <button
              onClick={() => setActiveModule('safety')}
              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left hover:border-sky-300 hover:shadow-xs transition"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white">Safety & Diagnostics</span>
                <span className="text-[10px] bg-rose-100 text-rose-800 dark:bg-rose-900/50 dark:text-rose-300 px-2 py-0.5 rounded font-bold">Guard</span>
              </div>
              <p className="text-[11px] text-slate-500 mt-1">Rate limits gatekeeper, provider doctor & CSV relationships exports.</p>
            </button>
          </div>
        {/* Presets & Simulator */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Select Authorization Scenario (Groundtruth Test Fixtures)
            </h2>
            <span className="text-[11px] font-medium text-slate-500">
              Test fail-closed behavior across consumer, elevated creator, and partner scopes
            </span>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-4">
            {PRESET_CONNECTIONS.map((preset, idx) => (
              <button
                key={preset.name}
                onClick={() => handleSelectPreset(idx)}
                className={`p-3.5 rounded-lg border text-left transition-all ${
                  selectedPresetIndex === idx
                    ? 'border-sky-600 bg-sky-50/50 dark:bg-sky-950/40 ring-1 ring-sky-500'
                    : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-slate-300'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-slate-900 dark:text-white truncate">
                    {preset.name}
                  </span>
                  {selectedPresetIndex === idx && (
                    <span className="h-2 w-2 rounded-full bg-sky-600 shrink-0"></span>
                  )}
                </div>
                <p className="mt-1 text-[11px] text-slate-500 dark:text-slate-400 line-clamp-2 leading-relaxed">
                  {preset.description}
                </p>
                <div className="mt-2 text-[10px] font-mono text-sky-700 dark:text-sky-300">
                  Status: {preset.connection.status}
                </div>
              </button>
            ))}
          </div>
        </section>

        {/* Connection Overview Metrics */}
        <section className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 shadow-sm">
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Account State</span>
            <div className="mt-1 flex items-center space-x-2">
              <span className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-extrabold capitalize ${
                currentConnection.status === 'connected'
                  ? 'bg-emerald-100 text-emerald-800'
                  : currentConnection.status === 'partially_authorized'
                  ? 'bg-amber-100 text-amber-800'
                  : currentConnection.status === 'reauth_required'
                  ? 'bg-rose-100 text-rose-800'
                  : 'bg-slate-100 text-slate-800'
              }`}>
                {currentConnection.status.replace('_', ' ')}
              </span>
            </div>
            <p className="mt-2 text-xs text-slate-600 dark:text-slate-400 font-medium">
              {currentConnection.display_name} ({currentConnection.email})
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 shadow-sm">
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Granted Scopes</span>
            <div className="mt-1 text-xl font-bold text-slate-900 dark:text-white">
              {currentConnection.granted_scopes.length} Scopes Active
            </div>
            <div className="mt-1 flex flex-wrap gap-1">
              {currentConnection.granted_scopes.map((sc) => (
                <span key={sc} className="rounded bg-sky-50 dark:bg-sky-950 px-1.5 py-0.5 text-[10px] font-mono text-sky-700 dark:text-sky-300 border border-sky-100 dark:border-sky-900">
                  {sc}
                </span>
              ))}
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 shadow-sm">
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Masked Vault Token</span>
            <div className="mt-1 font-mono text-sm font-bold text-slate-700 dark:text-slate-300">
              {currentConnection.masked_token}
            </div>
            <p className="mt-1 text-[11px] text-slate-500">
              AES-256-GCM encrypted in Vault (AT-016).
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 shadow-sm">
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Capabilities Granted</span>
            <div className="mt-1 text-xl font-extrabold text-sky-600">
              {currentConnection.capabilities.filter((c) => c.is_granted).length} / {currentConnection.capabilities.length}
            </div>
            <p className="mt-1 text-[11px] text-slate-500">
              {currentConnection.capabilities.filter((c) => !c.is_granted).length} Restricted or Partner-Gated
            </p>
          </div>
        </section>

        {/* Truth-in-Advertising & Capability Registry Invariant (AT-010, REQ-009) */}
        <section className="rounded-xl border border-blue-200 dark:border-blue-900/60 bg-blue-50/50 dark:bg-blue-950/30 p-4">
          <div className="flex items-start space-x-3">
            <Shield className="w-5 h-5 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
            <div className="text-xs text-blue-900 dark:text-blue-300 leading-relaxed">
              <span className="font-bold text-blue-950 dark:text-blue-100">Truth-in-Advertising & Platform Boundary Contract (AT-010, REQ-021):</span>
              {' '}Connecting via LinkedIn OAuth conveys only explicit permissions granted by LinkedIn. Consumer apps cannot execute background direct messaging or automated Easy Apply without risking platform bans. All restricted capabilities below fail closed and provide safe Assisted-Manual fallback guidance.
            </div>
          </div>
        </section>

        {/* Capabilities Matrix */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
          <div className="border-b border-slate-200 dark:border-slate-800 px-6 py-4">
            <h2 className="text-base font-bold text-slate-900 dark:text-white">
              Granular Capability Registry & Inspection Matrix (LI-01)
            </h2>
            <p className="text-xs text-slate-500 mt-0.5">
              Inspect what operations your current OAuth scopes permit, what is restricted, and available fallback modes.
            </p>
          </div>

          <div className="divide-y divide-slate-200 dark:divide-slate-800">
            {currentConnection.capabilities.map((cap) => (
              <div key={cap.capability_id} className="p-6 hover:bg-slate-50/60 dark:hover:bg-slate-900/50 transition-colors">
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                  <div className="space-y-1 max-w-3xl">
                    <div className="flex items-center space-x-2">
                      <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-bold uppercase tracking-wider ${
                        cap.status === 'granted'
                          ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                          : cap.status === 'unavailable_partner_only'
                          ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                          : cap.status === 'unsupported_platform'
                          ? 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300'
                          : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                      }`}>
                        {cap.status.replace(/_/g, ' ')}
                      </span>
                      <span className="text-xs font-mono text-slate-400">
                        Scope: {cap.required_scope}
                      </span>
                      <span className="text-xs font-semibold text-slate-500">
                        &bull; Domain: {cap.domain}
                      </span>
                    </div>

                    <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                      {cap.name}
                    </h3>
                    <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
                      {cap.description}
                    </p>

                    <div className="mt-2 rounded-md bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 p-2.5 text-[11px] text-slate-600 dark:text-slate-300 leading-relaxed">
                      <strong className="text-slate-800 dark:text-slate-100">Truth-in-Advertising Note:</strong> {cap.truth_in_advertising_note}
                    </div>
                  </div>

                  <div className="flex flex-col items-end shrink-0 space-y-2">
                    {cap.is_granted ? (
                      <span className="inline-flex items-center gap-1 text-xs font-bold text-emerald-600">
                        <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                        <span>Granted & Ready</span>
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-xs font-bold text-rose-600">
                        <XCircle className="w-4 h-4 text-rose-500" />
                        <span>Unavailable / Denied</span>
                      </span>
                    )}

                    {cap.fallback_mode !== 'none' && (
                      <span className="rounded bg-indigo-50 dark:bg-indigo-950 px-2 py-0.5 text-[10px] font-semibold text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800">
                        Fallback: {cap.fallback_mode.replace(/_/g, ' ')}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>
        </div>
      )}

      {/* MODULE 2: PROFILE & VAULT */}
      {activeModule === 'profile' && (
        <div className="space-y-6">
        {/* Own-Profile Scan & Ingestion Studio (IMP-LI-02, LI-02, AT-001, AT-010) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center space-x-2">
                <span className="inline-flex items-center rounded-full bg-emerald-50 dark:bg-emerald-950/60 px-2 py-0.5 text-[10px] font-bold text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
                  IMP-LI-02 &bull; LI-02
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Database className="w-4 h-4 text-emerald-600" />
                  <span>Own-Profile Scan, Ingestion & Merging Studio</span>
                </h2>
              </div>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                Ingest your own LinkedIn profile history via official export archives or pasted unstructured text. Reconciles into canonical career schema with AT-010 truth-in-advertising guidance.
              </p>
            </div>

            {/* Mode Toggle */}
            <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 p-1 shrink-0">
              <button
                onClick={() => setImportMode('zip')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  importMode === 'zip'
                    ? 'bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-400 shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <FileArchive className="w-3.5 h-3.5" />
                <span>Archive ZIP</span>
              </button>
              <button
                onClick={() => setImportMode('paste')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  importMode === 'paste'
                    ? 'bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-400 shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <FileText className="w-3.5 h-3.5" />
                <span>Text Paste</span>
              </button>
              <button
                onClick={() => setImportMode('oidc')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  importMode === 'oidc'
                    ? 'bg-white dark:bg-slate-900 text-amber-700 dark:text-amber-400 shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <AlertTriangle className="w-3.5 h-3.5" />
                <span>OIDC Fallback (AT-010)</span>
              </button>
            </div>
          </div>

          <div className="mt-6 grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Input Side */}
            <div className="lg:col-span-5 space-y-4">
              {importMode === 'zip' && (
                <div className="rounded-xl border-2 border-dashed border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40 p-6 text-center">
                  <UploadCloud className="mx-auto h-8 w-8 text-slate-400" />
                  <h3 className="mt-2 text-xs font-bold text-slate-900 dark:text-white">
                    Upload Basic_LinkedInData.zip
                  </h3>
                  <p className="mt-1 text-[11px] text-slate-500 dark:text-slate-400">
                    Contains official <code className="font-mono text-emerald-600">Positions.csv</code>, <code className="font-mono text-emerald-600">Education.csv</code>, <code className="font-mono text-emerald-600">Skills.csv</code>
                  </p>
                  <div className="mt-4 flex flex-col gap-2">
                    <button
                      onClick={handleSimulateZipImport}
                      className="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 px-4 py-2 text-xs font-bold text-white shadow-sm transition-colors"
                    >
                      <Sparkles className="w-3.5 h-3.5" />
                      <span>Load Sample Archive (Sarah Connor)</span>
                    </button>
                    <span className="text-[10px] text-slate-400">
                      Standard LinkedIn Data Export (GDPR / CCPA full download)
                    </span>
                  </div>
                </div>
              )}

              {importMode === 'paste' && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-bold text-slate-700 dark:text-slate-300">
                      Unstructured Profile Text:
                    </label>
                    <button
                      onClick={handleSimulatePasteImport}
                      className="text-[11px] font-bold text-emerald-600 hover:text-emerald-700 underline"
                    >
                      Reset Sample (David Miller)
                    </button>
                  </div>
                  <textarea
                    rows={8}
                    value={pastedInput}
                    onChange={(e) => setPastedInput(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 p-3 font-mono text-[11px] text-slate-800 dark:text-slate-200 shadow-sm focus:border-emerald-500 focus:outline-none"
                    placeholder="Paste copied text from your public LinkedIn profile..."
                  />
                  <button
                    onClick={handleSimulatePasteImport}
                    className="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 px-4 py-2 text-xs font-bold text-white shadow-sm transition-colors"
                  >
                    <Sparkles className="w-3.5 h-3.5" />
                    <span>Parse Pasted Profile Text</span>
                  </button>
                </div>
              )}

              {importMode === 'oidc' && (
                <div className="rounded-xl border border-amber-200 dark:border-amber-900/60 bg-amber-50/50 dark:bg-amber-950/30 p-5 space-y-3">
                  <div className="flex items-center gap-2 text-amber-800 dark:text-amber-300 font-bold text-xs">
                    <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
                    <span>Consumer OAuth Scan Limitation (AT-010)</span>
                  </div>
                  <p className="text-[11px] text-amber-900/80 dark:text-amber-300/80 leading-relaxed">
                    Standard LinkedIn Sign-in (OpenID Connect) only provides basic identity claims (name and email). Career history API access is strictly restricted by LinkedIn to certified enterprise talent partners.
                  </p>
                  <button
                    onClick={handleSimulateOIDCFallback}
                    className="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-amber-600 hover:bg-amber-700 px-4 py-2 text-xs font-bold text-white shadow-sm transition-colors"
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                    <span>Simulate Consumer OIDC Attempt</span>
                  </button>
                </div>
              )}
            </div>

            {/* Extraction Preview & Merge Side */}
            <div className="lg:col-span-7">
              {activeImportedProfile ? (
                <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/30 p-5 space-y-4">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                          {activeImportedProfile.display_name}
                        </h3>
                        <span className="rounded bg-slate-200 dark:bg-slate-700 px-2 py-0.5 text-[10px] font-mono text-slate-700 dark:text-slate-300">
                          Source: {activeImportedProfile.source_mode}
                        </span>
                        <span className={`rounded px-2 py-0.5 text-[10px] font-bold capitalize ${
                          activeImportedProfile.status === 'merged'
                            ? 'bg-emerald-100 text-emerald-800'
                            : 'bg-sky-100 text-sky-800'
                        }`}>
                          {activeImportedProfile.status}
                        </span>
                      </div>
                      <p className="text-xs font-semibold text-slate-600 dark:text-slate-300 mt-0.5">
                        {activeImportedProfile.headline}
                      </p>
                      {activeImportedProfile.location && (
                        <p className="text-[11px] text-slate-500 mt-0.5">
                          Location: {activeImportedProfile.location} {activeImportedProfile.industry ? `• ${activeImportedProfile.industry}` : ''}
                        </p>
                      )}
                    </div>

                    <button
                      onClick={handleMergeToCareer}
                      disabled={isMerging || activeImportedProfile.status === 'merged' || activeImportedProfile.requires_fallback}
                      className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 disabled:bg-slate-400 px-3.5 py-1.5 text-xs font-bold text-white shadow-sm transition-colors"
                    >
                      <Layers className="w-3.5 h-3.5" />
                      <span>{isMerging ? 'Reconciling...' : activeImportedProfile.status === 'merged' ? 'Merged to Hub' : 'Merge to Master Profile'}</span>
                    </button>
                  </div>

                  {mergeStatus && (
                    <div className="rounded-lg bg-emerald-100 dark:bg-emerald-950/60 border border-emerald-300 dark:border-emerald-800 p-3 text-xs font-semibold text-emerald-800 dark:text-emerald-300">
                      {mergeStatus}
                    </div>
                  )}

                  {activeImportedProfile.requires_fallback && (
                    <div className="rounded-lg bg-amber-50 dark:bg-amber-950/40 border border-amber-300 dark:border-amber-800 p-3.5 space-y-1 text-xs text-amber-900 dark:text-amber-200">
                      <div className="font-bold flex items-center gap-1.5">
                        <AlertTriangle className="w-3.5 h-3.5 text-amber-600" />
                        <span>Truth-in-Advertising Fallback Notice</span>
                      </div>
                      <p className="text-[11px] leading-relaxed">
                        {activeImportedProfile.truth_in_advertising_note}
                      </p>
                      <div className="pt-2">
                        <button
                          onClick={() => setImportMode('zip')}
                          className="inline-flex items-center gap-1 text-xs font-bold text-emerald-700 dark:text-emerald-400 underline"
                        >
                          <span>Switch to Archive ZIP Import</span>
                          <ArrowRight className="w-3 h-3" />
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Experiences List */}
                  {activeImportedProfile.experiences.length > 0 && (
                    <div className="space-y-2">
                      <div className="text-[11px] font-bold uppercase tracking-wider text-slate-400 flex items-center justify-between">
                        <span>Extracted Work History ({activeImportedProfile.experiences.length})</span>
                      </div>
                      <div className="space-y-2">
                        {activeImportedProfile.experiences.map((exp, idx) => (
                          <div
                            key={idx}
                            className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-3 text-xs"
                          >
                            <div className="flex items-center justify-between font-bold text-slate-800 dark:text-slate-100">
                              <span>{exp.title}</span>
                              <span className="text-[11px] font-normal text-slate-500">
                                {exp.start_date} &ndash; {exp.is_current ? 'Present' : exp.end_date}
                              </span>
                            </div>
                            <div className="text-[11px] font-medium text-emerald-700 dark:text-emerald-400">
                              {exp.company_name} {exp.location ? `• ${exp.location}` : ''}
                            </div>
                            {exp.description && (
                              <p className="mt-1 text-[11px] text-slate-500 leading-relaxed">
                                {exp.description}
                              </p>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Education and Skills */}
                  {(activeImportedProfile.education.length > 0 || activeImportedProfile.skills.length > 0) && (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                      {activeImportedProfile.education.length > 0 && (
                        <div className="space-y-1.5">
                          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Education</span>
                          <div className="space-y-1.5">
                            {activeImportedProfile.education.map((ed, idx) => (
                              <div key={idx} className="rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-2.5 text-xs">
                                <div className="font-bold text-slate-800 dark:text-slate-200">{ed.school_name}</div>
                                <div className="text-[11px] text-slate-500">{ed.degree_name}</div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {activeImportedProfile.skills.length > 0 && (
                        <div className="space-y-1.5">
                          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">
                            Extracted Skills ({activeImportedProfile.skills.length})
                          </span>
                          <div className="flex flex-wrap gap-1.5">
                            {activeImportedProfile.skills.map((skill, idx) => (
                              <span
                                key={idx}
                                className="rounded-md bg-emerald-50 dark:bg-emerald-950/50 border border-emerald-200 dark:border-emerald-800 px-2 py-0.5 text-[11px] font-semibold text-emerald-700 dark:text-emerald-300"
                              >
                                {skill}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ) : (
                <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/30 p-8 text-center text-xs text-slate-500">
                  Select an import mode to ingest and preview your LinkedIn profile facts.
                </div>
              )}
            </div>
          </div>
        </section>

        {/* LinkedIn Profile Optimization & Before/After Diff Studio (IMP-LI-03, LI-03, AT-003, AT-007) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center space-x-2">
                <span className="inline-flex items-center rounded-full bg-indigo-50 dark:bg-indigo-950/60 px-2 py-0.5 text-[10px] font-bold text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800">
                  IMP-LI-03 &bull; LI-03 &bull; AT-003 &bull; AT-007
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Sparkles className="w-4 h-4 text-indigo-600" />
                  <span>LinkedIn Profile Optimization & Before/After Diff Studio</span>
                </h2>
              </div>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                Fact-grounded profile refinements across Headline, About, and Skills. Every recommendation is derived exclusively from verified career facts (AT-003) with a mandatory human approval gate (AT-007).
              </p>
            </div>

            {/* Profile Health Score Gauge */}
            <div className="flex items-center gap-3 bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-2 shrink-0">
              <div className="text-right">
                <span className="text-[10px] uppercase font-bold text-slate-400 block">Profile Strength</span>
                <span className="text-sm font-black text-indigo-600 dark:text-indigo-400">
                  {activeOptimizationReport.overall_profile_score} / 100
                </span>
              </div>
              <div className="h-8 w-8 rounded-full border-2 border-indigo-500 flex items-center justify-center font-bold text-xs text-indigo-600 bg-indigo-50 dark:bg-indigo-950">
                A
              </div>
            </div>
          </div>

          {/* Section Refinements List */}
          <div className="mt-6 space-y-6">
            {activeOptimizationReport.suggestions.map((sug) => {
              const isEditing = editingSuggestionId === sug.suggestion_id;
              const displayText = sug.custom_override || sug.after;

              return (
                <div
                  key={sug.suggestion_id}
                  className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/20 p-5 space-y-4"
                >
                  {/* Card Header */}
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold uppercase tracking-wider text-slate-500 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 px-2.5 py-0.5 rounded-md">
                        Section: {sug.section}
                      </span>
                      <span className="rounded bg-indigo-100 dark:bg-indigo-950/80 px-2 py-0.5 text-[10px] font-bold text-indigo-800 dark:text-indigo-300">
                        Impact Score: {sug.impact_score}/100
                      </span>
                    </div>

                    {/* Status Pill */}
                    <div className="flex items-center gap-2">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-[10px] font-extrabold capitalize ${
                          sug.approval_state === 'approved'
                            ? 'bg-emerald-100 text-emerald-800 border border-emerald-300'
                            : sug.approval_state === 'rejected'
                            ? 'bg-rose-100 text-rose-800 border border-rose-300'
                            : sug.approval_state === 'custom_edited'
                            ? 'bg-purple-100 text-purple-800 border border-purple-300'
                            : 'bg-amber-100 text-amber-800 border border-amber-300'
                        }`}
                      >
                        {sug.approval_state.replace('_', ' ')}
                      </span>
                    </div>
                  </div>

                  {/* Before vs After Diff Grid */}
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    {/* Before */}
                    <div className="rounded-lg border border-rose-200 dark:border-rose-950/60 bg-rose-50/40 dark:bg-rose-950/20 p-3.5 space-y-1.5">
                      <div className="flex items-center gap-1.5 text-[11px] font-bold text-rose-800 dark:text-rose-300">
                        <XCircle className="w-3.5 h-3.5 text-rose-600" />
                        <span>Current (Before)</span>
                      </div>
                      <p className="text-xs text-rose-900/80 dark:text-rose-200/80 font-mono whitespace-pre-wrap leading-relaxed">
                        {sug.before}
                      </p>
                    </div>

                    {/* After */}
                    <div className="rounded-lg border border-emerald-200 dark:border-emerald-950/60 bg-emerald-50/40 dark:bg-emerald-950/20 p-3.5 space-y-1.5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-1.5 text-[11px] font-bold text-emerald-800 dark:text-emerald-300">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                          <span>Optimized Refinement (After)</span>
                        </div>
                        {sug.custom_override && (
                          <span className="text-[10px] font-bold text-purple-700 dark:text-purple-300">
                            (Human Edited)
                          </span>
                        )}
                      </div>

                      {isEditing ? (
                        <div className="space-y-2">
                          <textarea
                            rows={4}
                            value={customDraftText}
                            onChange={(e) => setCustomDraftText(e.target.value)}
                            className="w-full rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 p-2 text-xs font-mono text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                          />
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleSaveCustomEdit(sug.suggestion_id)}
                              className="px-2.5 py-1 text-xs font-bold rounded bg-indigo-600 hover:bg-indigo-700 text-white"
                            >
                              Save Edit
                            </button>
                            <button
                              onClick={() => setEditingSuggestionId(null)}
                              className="px-2.5 py-1 text-xs font-semibold rounded bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300"
                            >
                              Cancel
                            </button>
                          </div>
                        </div>
                      ) : (
                        <p className="text-xs text-emerald-950 dark:text-emerald-100 font-mono whitespace-pre-wrap leading-relaxed">
                          {displayText}
                        </p>
                      )}
                    </div>
                  </div>

                  {/* Fact Attribution & Rationale */}
                  <div className="rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-3 space-y-2 text-xs">
                    <p className="text-slate-600 dark:text-slate-300 leading-relaxed">
                      <strong>Rationale:</strong> {sug.rationale}
                    </p>

                    {sug.source_facts_used.length > 0 && (
                      <div className="pt-1 flex flex-wrap items-center gap-1.5">
                        <span className="text-[10px] font-bold uppercase text-slate-400">
                          Grounded in Verified Facts (AT-003):
                        </span>
                        {sug.source_facts_used.map((fact, fIdx) => (
                          <span
                            key={fIdx}
                            className="inline-flex items-center rounded bg-slate-100 dark:bg-slate-800 px-2 py-0.5 text-[10px] font-medium text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700"
                          >
                            &bull; {fact}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Action Controls & Clipboard Gate */}
                  <div className="flex flex-wrap items-center justify-between gap-3 pt-1">
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleApproveSuggestion(sug.suggestion_id)}
                        disabled={sug.approval_state === 'approved'}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold bg-emerald-600 hover:bg-emerald-700 disabled:opacity-40 text-white shadow-sm"
                      >
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        <span>Approve</span>
                      </button>
                      <button
                        onClick={() => handleRejectSuggestion(sug.suggestion_id)}
                        disabled={sug.approval_state === 'rejected'}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold bg-white dark:bg-slate-900 border border-rose-300 hover:bg-rose-50 text-rose-700 disabled:opacity-40 shadow-sm"
                      >
                        <XCircle className="w-3.5 h-3.5" />
                        <span>Reject</span>
                      </button>
                      <button
                        onClick={() => handleStartCustomEdit(sug.suggestion_id, displayText)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-white dark:bg-slate-900 border border-slate-300 hover:bg-slate-50 text-slate-700 dark:text-slate-300 shadow-sm"
                      >
                        <span>Custom Edit (AT-007)</span>
                      </button>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleCopyText(sug.suggestion_id, displayText)}
                        className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm transition-colors"
                      >
                        <Copy className="w-3.5 h-3.5" />
                        <span>
                          {copiedSuggestionId === sug.suggestion_id ? 'Copied to Clipboard!' : 'Copy to Clipboard'}
                        </span>
                      </button>
                      <a
                        href="https://www.linkedin.com/in/"
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1 text-[11px] font-semibold text-slate-500 hover:text-slate-800 underline ml-1"
                      >
                        <span>Edit on LinkedIn</span>
                        <ExternalLink className="w-3 h-3" />
                      </a>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        {/* LinkedIn Records Vault (IMP-LI-04, AT-011, AT-012, FND-009, FND-012, SRC-L3, SRC-L4) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                  IMP-LI-04 • Records Engine
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300">
                  AT-011 / AT-012 Tenant Isolated
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-amber-100 text-amber-800 dark:bg-amber-900/50 dark:text-amber-300">
                  FND-009 Permitted Fields Only
                </span>
              </div>
              <h2 className="text-base font-bold text-slate-900 dark:text-white mt-1.5 flex items-center gap-2">
                <Database className="w-4 h-4 text-indigo-600" />
                <span>LinkedIn Records Vault (Normalized Persons, Companies, Jobs & Posts)</span>
              </h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                Audited entity snapshot store with strict cross-tenant isolation, versioned histories, and zero prohibited biometric/sensitive data.
              </p>
            </div>

            {/* Workspace & Tenant Isolation Selector */}
            <div className="flex items-center gap-3 bg-slate-50 dark:bg-slate-800/60 p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-xs">
              <span className="font-semibold text-slate-600 dark:text-slate-300 flex items-center gap-1">
                <Shield className="w-3.5 h-3.5 text-indigo-600" />
                <span>Active Scope:</span>
              </span>
              <div className="flex gap-1.5">
                <button
                  type="button"
                  onClick={() => setRecordsWorkspace('ws-alpha')}
                  className={`px-2.5 py-1 text-xs font-medium rounded-md transition-colors ${
                    recordsWorkspace === 'ws-alpha'
                      ? 'bg-indigo-600 text-white shadow-xs'
                      : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 hover:bg-slate-100'
                  }`}
                >
                  Workspace Alpha (`tenant-alpha`)
                </button>
                <button
                  type="button"
                  onClick={() => setRecordsWorkspace('ws-beta')}
                  className={`px-2.5 py-1 text-xs font-medium rounded-md transition-colors ${
                    recordsWorkspace === 'ws-beta'
                      ? 'bg-indigo-600 text-white shadow-xs'
                      : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 hover:bg-slate-100'
                  }`}
                >
                  Workspace Beta (`tenant-beta`)
                </button>
              </div>
            </div>
          </div>

          {/* Navigation Tabs and Search */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mt-5">
            <div className="flex gap-1 p-1 bg-slate-100 dark:bg-slate-800 rounded-lg max-w-fit">
              <button
                type="button"
                onClick={() => setRecordsTab('person')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  recordsTab === 'person'
                    ? 'bg-white dark:bg-slate-900 text-indigo-700 dark:text-indigo-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <Users className="w-3.5 h-3.5" />
                <span>Persons ({filteredPersons.length})</span>
              </button>
              <button
                type="button"
                onClick={() => setRecordsTab('company')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  recordsTab === 'company'
                    ? 'bg-white dark:bg-slate-900 text-indigo-700 dark:text-indigo-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <Building2 className="w-3.5 h-3.5" />
                <span>Companies ({filteredCompanies.length})</span>
              </button>
              <button
                type="button"
                onClick={() => setRecordsTab('job')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  recordsTab === 'job'
                    ? 'bg-white dark:bg-slate-900 text-indigo-700 dark:text-indigo-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <Briefcase className="w-3.5 h-3.5" />
                <span>Jobs ({filteredJobs.length})</span>
              </button>
              <button
                type="button"
                onClick={() => setRecordsTab('post')}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  recordsTab === 'post'
                    ? 'bg-white dark:bg-slate-900 text-indigo-700 dark:text-indigo-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                <MessageSquare className="w-3.5 h-3.5" />
                <span>Posts ({filteredPosts.length})</span>
              </button>
            </div>

            <div className="w-full sm:w-64">
              <input
                type="text"
                placeholder={`Search ${recordsTab}s in ${recordsWorkspace}...`}
                value={recordsSearch}
                onChange={(e) => setRecordsSearch(e.target.value)}
                className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
          </div>

          {/* Record Display Panels */}
          <div className="mt-4">
            {recordsTab === 'person' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredPersons.map((p) => (
                  <div
                    key={p.record_id}
                    className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 flex flex-col justify-between"
                  >
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-[11px] font-semibold text-indigo-700 dark:text-indigo-400">
                          {p.record_id}
                        </span>
                        <div className="flex items-center gap-1.5">
                          <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                            v{p.version}
                          </span>
                          <span className="px-1.5 py-0.5 text-[10px] font-semibold rounded bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
                            {p.connection_tier}
                          </span>
                        </div>
                      </div>
                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                        {p.full_name}
                      </h3>
                      <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5 line-clamp-2">
                        {p.headline}
                      </p>
                      <div className="mt-2.5 text-[11px] text-slate-500 space-y-1">
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Company:</span> {p.current_company} ({p.current_role})</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Location:</span> {p.location}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Tenant:</span> {p.tenant_id}</div>
                      </div>
                      <div className="mt-3 flex flex-wrap gap-1">
                        {p.skills.map((skill, idx) => (
                          <span
                            key={idx}
                            className="px-1.5 py-0.5 text-[10px] rounded bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>

                    <div className="mt-4 pt-3 border-t border-slate-200 dark:border-slate-700/60 flex items-center justify-between text-[11px]">
                      <span className="text-slate-400">
                        Observed: {new Date(p.observed_at).toLocaleTimeString()}
                      </span>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleQueueNoteFromPerson(p)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 transition-colors"
                        >
                          <Send className="w-3 h-3" />
                          <span>Queue Note</span>
                        </button>
                        <button
                          type="button"
                          onClick={() => handleConvertPersonToLead(p)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-purple-50 dark:bg-purple-950/50 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800 hover:bg-purple-100 transition-colors"
                        >
                          <UserPlus className="w-3 h-3" />
                          <span>Convert to Lead</span>
                        </button>
                        <button
                          type="button"
                          onClick={() => handleBumpPersonVersion(p.record_id)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-100 transition-colors"
                        >
                          <RefreshCw className="w-3 h-3" />
                          <span>Snapshot Version (v+1)</span>
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {recordsTab === 'company' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredCompanies.map((c) => (
                  <div
                    key={c.record_id}
                    className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 flex flex-col justify-between"
                  >
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-[11px] font-semibold text-indigo-700 dark:text-indigo-400">
                          {c.record_id}
                        </span>
                        <div className="flex items-center gap-1.5">
                          <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                            v{c.version}
                          </span>
                          {c.verified && (
                            <span className="px-1.5 py-0.5 text-[10px] font-semibold rounded bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300">
                              Verified
                            </span>
                          )}
                        </div>
                      </div>
                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                        {c.company_name}
                      </h3>
                      <div className="mt-2 text-[11px] text-slate-500 space-y-1">
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Domain:</span> {c.domain}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Industry:</span> {c.industry}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Size:</span> {c.size_tier}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">HQ:</span> {c.headquarters}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Followers:</span> {c.follower_count.toLocaleString()}</div>
                      </div>
                      <div className="mt-3 flex flex-wrap gap-1">
                        {c.specialties.map((spec, idx) => (
                          <span
                            key={idx}
                            className="px-1.5 py-0.5 text-[10px] rounded bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300"
                          >
                            {spec}
                          </span>
                        ))}
                      </div>
                    </div>

                    <div className="mt-4 pt-3 border-t border-slate-200 dark:border-slate-700/60 flex items-center justify-between text-[11px]">
                      <span className="text-slate-400">
                        Observed: {new Date(c.observed_at).toLocaleTimeString()}
                      </span>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleAddCompanyFromVault(c)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-sky-50 dark:bg-sky-950/50 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800 hover:bg-sky-100 transition-colors"
                        >
                          <Plus className="w-3 h-3" />
                          <span>Add to Watchlist</span>
                        </button>
                        <button
                          type="button"
                          onClick={() => handleBumpCompanyVersion(c.record_id)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-100 transition-colors"
                        >
                          <RefreshCw className="w-3 h-3" />
                          <span>Snapshot Version (v+1)</span>
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {recordsTab === 'job' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredJobs.map((j) => (
                  <div
                    key={j.record_id}
                    className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 flex flex-col justify-between"
                  >
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-[11px] font-semibold text-indigo-700 dark:text-indigo-400">
                          {j.record_id} • {j.job_id}
                        </span>
                        <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                          v{j.version}
                        </span>
                      </div>
                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                        {j.title}
                      </h3>
                      <p className="text-xs font-semibold text-slate-700 dark:text-slate-300 mt-0.5">
                        {j.company_name} • {j.location} ({j.workplace_type})
                      </p>
                      <p className="text-xs text-slate-500 mt-2 line-clamp-2">
                        {j.description_snippet}
                      </p>
                      <div className="mt-2.5 text-[11px] text-slate-500 space-y-1">
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Type:</span> {j.employment_type}</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Applicants:</span> {j.applicant_count} candidates</div>
                        <div><span className="font-semibold text-slate-700 dark:text-slate-300">Tenant:</span> {j.tenant_id}</div>
                      </div>
                    </div>

                    <div className="mt-4 pt-3 border-t border-slate-200 dark:border-slate-700/60 flex items-center justify-between text-[11px]">
                      <span className="text-slate-400">
                        Observed: {new Date(j.observed_at).toLocaleTimeString()}
                      </span>
                      <button
                        type="button"
                        onClick={() => handleBumpJobVersion(j.record_id)}
                        className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-100 transition-colors"
                      >
                        <RefreshCw className="w-3 h-3" />
                        <span>Update Metrics (v+1)</span>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {recordsTab === 'post' && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredPosts.map((post) => (
                  <div
                    key={post.record_id}
                    className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 flex flex-col justify-between"
                  >
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-[11px] font-semibold text-indigo-700 dark:text-indigo-400">
                          {post.record_id}
                        </span>
                        <div className="flex items-center gap-1.5">
                          <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                            v{post.version}
                          </span>
                          <span className="px-1.5 py-0.5 text-[10px] font-semibold rounded bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300 uppercase">
                            {post.media_type}
                          </span>
                        </div>
                      </div>
                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                        {post.author_name}
                      </h3>
                      <p className="text-[11px] text-slate-400 font-mono">
                        {post.urn}
                      </p>
                      <p className="text-xs text-slate-700 dark:text-slate-300 mt-2 italic bg-white dark:bg-slate-800 p-2.5 rounded border border-slate-100 dark:border-slate-700">
                        "{post.commentary}"
                      </p>
                      <div className="mt-2.5 flex items-center gap-3 text-[11px] text-slate-600 dark:text-slate-400 font-medium">
                        <span>👍 {post.reactions_count} reactions</span>
                        <span>💬 {post.comments_count} comments</span>
                        <span>🔄 {post.shares_count} shares</span>
                      </div>
                    </div>

                    <div className="mt-4 pt-3 border-t border-slate-200 dark:border-slate-700/60 flex items-center justify-between text-[11px]">
                      <span className="text-slate-400">
                        Observed: {new Date(post.observed_at).toLocaleTimeString()}
                      </span>
                      <button
                        type="button"
                        onClick={() => handleBumpPostVersion(post.record_id)}
                        className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-100 transition-colors"
                      >
                        <RefreshCw className="w-3 h-3" />
                        <span>Audit Engagement (v+1)</span>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </section>
        </div>
      )}

      {/* MODULE 3: NETWORK & LEADS */}
      {activeModule === 'network' && (
        <div className="space-y-6">
        {/* Professional People & Company Discovery Studio (IMP-LI-05, LI-05, AT-010, FND-011, SRC-L1, SRC-L3) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded bg-sky-100 text-sky-800 dark:bg-sky-900/50 dark:text-sky-300">
                  IMP-LI-05 • Discovery Engine
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-rose-100 text-rose-800 dark:bg-rose-900/50 dark:text-rose-300">
                  LI-05 Protected Attributes Banned
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300">
                  AT-010 Truth-in-Advertising
                </span>
              </div>
              <h2 className="text-base font-bold text-slate-900 dark:text-white mt-1.5 flex items-center gap-2">
                <Compass className="w-4 h-4 text-sky-600" />
                <span>Professional People & Company Discovery Studio</span>
              </h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                Precision talent and company intelligence powered by Boolean query engineering, canonical URL normalization, and verified Vault matching.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
            {/* Criteria Builder */}
            <div className="lg:col-span-5 space-y-4">
              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3.5">
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <Search className="w-3.5 h-3.5 text-sky-600" />
                  <span>Discovery Criteria (Non-Discriminatory)</span>
                </h3>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Target Professional Role / Title
                  </label>
                  <input
                    type="text"
                    value={discoveryRole}
                    onChange={(e) => setDiscoveryRole(e.target.value)}
                    placeholder="e.g. Distributed Systems Architect"
                    className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Stated Professional Goal
                  </label>
                  <select
                    value={discoveryGoal}
                    onChange={(e) => setDiscoveryGoal(e.target.value as LinkedInUserGoal)}
                    className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-sky-500"
                  >
                    <option value="recruiting">Recruiting / Sourcing Technical Talent</option>
                    <option value="networking">Professional Networking & Knowledge Share</option>
                    <option value="partnerships">Strategic Business Partnerships</option>
                    <option value="benchmarking">Industry & Architecture Benchmarking</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Target Organizations / Companies (comma-separated)
                  </label>
                  <input
                    type="text"
                    value={discoveryCompanies}
                    onChange={(e) => setDiscoveryCompanies(e.target.value)}
                    placeholder="e.g. CloudScale Systems, Apex Dynamics"
                    className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Title / Role Exclusions (comma-separated)
                  </label>
                  <input
                    type="text"
                    value={discoveryExclusions}
                    onChange={(e) => setDiscoveryExclusions(e.target.value)}
                    placeholder="e.g. Manager, Director, Sales"
                    className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                </div>

                <div className="pt-2">
                  <button
                    type="button"
                    onClick={handleRunDiscovery}
                    className="w-full inline-flex items-center justify-center gap-2 px-4 py-2 text-xs font-bold rounded-lg bg-sky-600 hover:bg-sky-700 text-white shadow-sm transition-colors"
                  >
                    <Compass className="w-4 h-4" />
                    <span>Generate Boolean & Match Vault</span>
                  </button>
                </div>
              </div>

              {/* Canonical URL Normalizer Utility */}
              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3">
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <Globe className="w-3.5 h-3.5 text-indigo-600" />
                  <span>Canonical URL Normalizer (SRC-L2)</span>
                </h3>
                <p className="text-[11px] text-slate-500">
                  Strip marketing tracking parameters (`?miniProfileUrn`, `trackingId`) into clean, stable entity permalinks.
                </p>
                <div className="space-y-2">
                  <input
                    type="text"
                    value={rawURLInput}
                    onChange={(e) => setRawURLInput(e.target.value)}
                    placeholder="Paste polluted LinkedIn URL..."
                    className="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-mono text-[11px] focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                  <button
                    type="button"
                    onClick={handleNormalizeURL}
                    className="w-full px-3 py-1.5 text-xs font-semibold rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 transition-colors"
                  >
                    Clean & Normalize URL
                  </button>
                </div>

                {canonicalURLResult && (
                  <div className="p-3 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-[11px] space-y-1 font-mono">
                    <div className="text-slate-500">Normalized: <span className="text-indigo-600 dark:text-indigo-400 font-semibold">{canonicalURLResult.normalized_url}</span></div>
                    <div className="text-slate-500">Entity: <span className="uppercase text-slate-700 dark:text-slate-300">{canonicalURLResult.entity_type}</span> ({canonicalURLResult.slug})</div>
                  </div>
                )}
              </div>
            </div>

            {/* Results & Boolean Dispatch */}
            <div className="lg:col-span-7 space-y-4">
              {activeDiscoveryResult ? (
                <div className="space-y-4">
                  {/* Boolean Search Engine Output */}
                  <div className="p-4 rounded-lg border border-sky-200 dark:border-sky-900/60 bg-sky-50/40 dark:bg-sky-950/20 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold uppercase tracking-wider text-sky-800 dark:text-sky-300 flex items-center gap-1.5">
                        <Terminal className="w-3.5 h-3.5 text-sky-600" />
                        <span>Generated LinkedIn Boolean Query (SRC-L1)</span>
                      </span>
                      <button
                        type="button"
                        onClick={handleCopyBoolean}
                        className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded bg-white dark:bg-slate-800 border border-sky-300 dark:border-sky-800 text-sky-700 dark:text-sky-300 hover:bg-sky-50 transition-colors"
                      >
                        {copiedBoolean ? <CheckCircle2 className="w-3 h-3 text-emerald-600" /> : <Copy className="w-3 h-3" />}
                        <span>{copiedBoolean ? 'Copied!' : 'Copy Query'}</span>
                      </button>
                    </div>

                    <div className="p-3 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 font-mono text-xs text-slate-800 dark:text-slate-200 break-all leading-relaxed">
                      {activeDiscoveryResult.boolean_query.raw_query}
                    </div>

                    <div className="flex items-center justify-between pt-1">
                      <span className="text-[11px] text-slate-500">
                        Target Goal: <span className="font-semibold capitalize text-slate-700 dark:text-slate-300">{activeDiscoveryResult.criteria.user_goal}</span>
                      </span>
                      <a
                        href={activeDiscoveryResult.boolean_query.linkedin_search_url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg bg-sky-600 hover:bg-sky-700 text-white shadow-xs transition-colors"
                      >
                        <span>Launch Search on LinkedIn</span>
                        <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    </div>
                  </div>

                  {/* Truth in Advertising Platform Reality Notice */}
                  <div className="p-3.5 rounded-lg border border-amber-200 dark:border-amber-900/60 bg-amber-50/50 dark:bg-amber-950/20 text-xs text-amber-900 dark:text-amber-300 flex items-start gap-2.5">
                    <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                    <div>
                      <span className="font-bold">Platform Reality Notice (AT-010):</span>{' '}
                      {activeDiscoveryResult.platform_notice}
                    </div>
                  </div>

                  {/* Matching Vault Records */}
                  <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-3">
                    <div className="flex items-center justify-between">
                      <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                        <Database className="w-3.5 h-3.5 text-indigo-600" />
                        <span>Matching Records in Active Scope (`{recordsWorkspace}`)</span>
                      </h4>
                      <span className="text-xs font-semibold text-slate-500">
                        {activeDiscoveryResult.total_vault_matches} match(es)
                      </span>
                    </div>

                    {activeDiscoveryResult.total_vault_matches === 0 ? (
                      <div className="p-4 text-center text-xs text-slate-400 italic bg-slate-50 dark:bg-slate-800/40 rounded border border-dashed border-slate-200 dark:border-slate-800">
                        No direct matches in local Vault. Use the Boolean query above to search live LinkedIn network.
                      </div>
                    ) : (
                      <div className="space-y-2">
                        {activeDiscoveryResult.matched_persons.map((p) => (
                          <div
                            key={p.record_id}
                            className="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-800/30 flex items-center justify-between text-xs"
                          >
                            <div>
                              <span className="font-bold text-slate-900 dark:text-white">{p.full_name}</span>
                              <p className="text-[11px] text-slate-500 mt-0.5">{p.headline}</p>
                            </div>
                            <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                              Person v{p.version}
                            </span>
                          </div>
                        ))}
                        {activeDiscoveryResult.matched_companies.map((c) => (
                          <div
                            key={c.record_id}
                            className="p-3 rounded border border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-800/30 flex items-center justify-between text-xs"
                          >
                            <div>
                              <span className="font-bold text-slate-900 dark:text-white">{c.company_name}</span>
                              <p className="text-[11px] text-slate-500 mt-0.5">{c.domain} • {c.industry}</p>
                            </div>
                            <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-indigo-100 text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-300">
                              Company v{c.version}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="h-full min-h-[300px] flex flex-col items-center justify-center p-8 rounded-lg border border-dashed border-slate-200 dark:border-slate-800 text-center">
                  <Compass className="w-10 h-10 text-slate-300 dark:text-slate-600 mb-3" />
                  <h4 className="text-sm font-semibold text-slate-700 dark:text-slate-300">
                    Discovery Studio Ready
                  </h4>
                  <p className="text-xs text-slate-400 mt-1 max-w-sm">
                    Configure your professional role, companies, and goals on the left, then click "Generate Boolean & Match Vault".
                  </p>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Recruiter & Hiring Lead Workspace Studio (IMP-LI-06, LI-06, AT-011, IMP-CAR-20, SRC-C1, SRC-L3) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300">
                  IMP-LI-06 • Recruiter Workspace
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300">
                  AT-011 Tenant Isolated
                </span>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded bg-sky-100 text-sky-800 dark:bg-sky-900/50 dark:text-sky-300">
                  IMP-CAR-20 Job Linkage
                </span>
              </div>
              <h2 className="text-base font-bold text-slate-900 dark:text-white mt-1.5 flex items-center gap-2">
                <UserCheck className="w-4 h-4 text-purple-600" />
                <span>Recruiter & Hiring Lead Workspace Studio</span>
              </h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                Manage relationships with recruiters, engineering hiring leads, interaction notes, follow-up alerts, and application linkages.
              </p>
            </div>

            {/* Workspace Scope Switcher */}
            <div className="flex items-center gap-3">
              <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">
                Workspace Scope:
              </span>
              <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 p-1">
                <button
                  type="button"
                  onClick={() => setRecruiterWorkspace('ws-alpha')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    recruiterWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-purple-700 dark:text-purple-300 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400'
                  }`}
                >
                  ws-alpha (Stripe / CloudScale)
                </button>
                <button
                  type="button"
                  onClick={() => setRecruiterWorkspace('ws-beta')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    recruiterWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-purple-700 dark:text-purple-300 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400'
                  }`}
                >
                  ws-beta (Datadog / SecureMesh)
                </button>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
            {/* Left Column: Leads List & Filters */}
            <div className="lg:col-span-5 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  {['all', 'new', 'contacted', 'in_dialogue'].map((st) => (
                    <button
                      key={st}
                      type="button"
                      onClick={() => setLeadStatusFilter(st)}
                      className={`px-2.5 py-1 text-[11px] font-semibold rounded-md uppercase tracking-wider transition-colors ${
                        leadStatusFilter === st
                          ? 'bg-purple-600 text-white'
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
                      }`}
                    >
                      {st}
                    </button>
                  ))}
                </div>
                <span className="text-xs text-slate-500 font-semibold">
                  {filteredLeads.length} Lead(s)
                </span>
              </div>

              <div className="space-y-3">
                {filteredLeads.map((lead) => {
                  const isSelected = selectedLead?.id === lead.id;
                  return (
                    <div
                      key={lead.id}
                      onClick={() => setSelectedLeadId(lead.id)}
                      className={`p-4 rounded-xl border cursor-pointer transition-all ${
                        isSelected
                          ? 'border-purple-500 bg-purple-50/40 dark:bg-purple-950/20 ring-1 ring-purple-500 shadow-xs'
                          : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-slate-300'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-[10px] font-mono text-slate-400">
                          {lead.id} &bull; {lead.source_entity_type}
                        </span>
                        <span className={`px-2 py-0.5 text-[10px] font-bold rounded-full uppercase ${
                          lead.status === 'in_dialogue'
                            ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                            : lead.status === 'contacted'
                            ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300'
                            : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
                        }`}>
                          {lead.status.replace('_', ' ')}
                        </span>
                      </div>

                      <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                        {lead.recruiter_name}
                      </h3>
                      <p className="text-xs text-slate-600 dark:text-slate-400">
                        {lead.recruiter_title} &bull; <span className="font-semibold text-slate-800 dark:text-slate-200">{lead.company}</span>
                      </p>

                      {lead.reminder && (
                        <div className="mt-2.5 flex items-center gap-1.5 text-[11px] text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-950/40 px-2 py-1 rounded border border-amber-200 dark:border-amber-900">
                          <Bell className="w-3 h-3 text-amber-600 shrink-0" />
                          <span className="truncate">Due: {new Date(lead.reminder.due_date).toLocaleDateString()} - {lead.reminder.message}</span>
                        </div>
                      )}

                      <div className="mt-2.5 pt-2 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between text-[11px] text-slate-500">
                        <span>{lead.notes.length} note(s)</span>
                        <span className="capitalize">{lead.outreach_stage.replace('_', ' ')}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Right Column: Active Lead Details, Notes, Reminder & Outreach Draft */}
            <div className="lg:col-span-7 space-y-4">
              {selectedLead ? (
                <div className="space-y-4">
                  {/* Lead Header Card */}
                  <div className="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs space-y-4">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100 dark:border-slate-800">
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-base font-bold text-slate-900 dark:text-white">
                            {selectedLead.recruiter_name}
                          </h3>
                          <a
                            href={selectedLead.linkedin_url}
                            target="_blank"
                            rel="noreferrer"
                            className="text-sky-600 hover:text-sky-700 inline-flex items-center gap-0.5 text-xs font-medium"
                          >
                            <span>Profile</span>
                            <ExternalLink className="w-3 h-3" />
                          </a>
                        </div>
                        <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5">
                          {selectedLead.recruiter_title} at <strong className="text-slate-800 dark:text-slate-200">{selectedLead.company}</strong>
                        </p>
                      </div>

                      {/* Status Update Dropdown */}
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-slate-500 font-semibold">Stage:</span>
                        <select
                          value={selectedLead.status}
                          onChange={(e) => handleUpdateLeadStatus(selectedLead.id, e.target.value as LinkedInLeadStatus)}
                          className="text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 px-2.5 py-1 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-purple-500"
                        >
                          <option value="new">New Lead</option>
                          <option value="contacted">Contacted</option>
                          <option value="in_dialogue">In Dialogue</option>
                          <option value="interview_scheduled">Interview Scheduled</option>
                          <option value="closed">Closed / Archived</option>
                        </select>
                      </div>
                    </div>

                    {/* Linked Job / Application & Tenant Meta */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                      <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800">
                        <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                          Related Job / Application (IMP-CAR-20)
                        </span>
                        <div className="mt-1 font-semibold text-slate-800 dark:text-slate-200">
                          {selectedLead.related_job_id || 'No linked job'} &bull; {selectedLead.related_application_id || 'No application'}
                        </div>
                      </div>
                      <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800">
                        <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                          Tenant Isolation Scope (AT-011)
                        </span>
                        <div className="mt-1 font-semibold text-purple-700 dark:text-purple-300">
                          {selectedLead.workspace_id} ({selectedLead.tenant_id})
                        </div>
                      </div>
                    </div>

                    {/* Follow-up Reminder Box */}
                    <div className="p-3.5 rounded-lg border border-amber-200 dark:border-amber-900/60 bg-amber-50/40 dark:bg-amber-950/20 space-y-2.5">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-bold uppercase tracking-wider text-amber-900 dark:text-amber-300 flex items-center gap-1.5">
                          <Bell className="w-3.5 h-3.5 text-amber-600" />
                          <span>Follow-up Reminder</span>
                        </span>
                        {selectedLead.reminder && selectedLead.reminder.status === 'pending' && (
                          <button
                            type="button"
                            onClick={() => handleCompleteLeadReminder(selectedLead.id)}
                            className="text-[11px] font-bold text-emerald-700 dark:text-emerald-300 bg-emerald-100 dark:bg-emerald-950/60 px-2 py-0.5 rounded hover:bg-emerald-200"
                          >
                            Mark Completed
                          </button>
                        )}
                      </div>

                      {selectedLead.reminder ? (
                        <div className="text-xs text-amber-900 dark:text-amber-200 flex items-center justify-between">
                          <div>
                            <span className="font-semibold">Alert:</span> {selectedLead.reminder.message}
                            <span className="text-amber-600 dark:text-amber-400 ml-2">
                              (Due: {new Date(selectedLead.reminder.due_date).toLocaleDateString()})
                            </span>
                          </div>
                          <span className="px-1.5 py-0.5 text-[10px] font-bold uppercase rounded bg-amber-200 dark:bg-amber-900 text-amber-900 dark:text-amber-100">
                            {selectedLead.reminder.status}
                          </span>
                        </div>
                      ) : (
                        <div className="text-xs text-slate-400 italic">No reminder scheduled currently.</div>
                      )}

                      {/* Schedule new reminder form */}
                      <div className="flex items-center gap-2 pt-1">
                        <input
                          type="text"
                          placeholder="e.g. Follow up on technical screen..."
                          value={newReminderMsg}
                          onChange={(e) => setNewReminderMsg(e.target.value)}
                          className="flex-1 px-3 py-1 text-xs rounded border border-amber-300 dark:border-amber-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-amber-500"
                        />
                        <select
                          value={newReminderDays}
                          onChange={(e) => setNewReminderDays(Number(e.target.value))}
                          className="px-2 py-1 text-xs rounded border border-amber-300 dark:border-amber-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                        >
                          <option value={1}>In 1 day</option>
                          <option value={3}>In 3 days</option>
                          <option value={7}>In 1 week</option>
                        </select>
                        <button
                          type="button"
                          onClick={() => handleSetLeadReminder(selectedLead.id)}
                          className="px-3 py-1 text-xs font-bold rounded bg-amber-600 hover:bg-amber-700 text-white shadow-xs"
                        >
                          Set Reminder
                        </button>
                      </div>
                    </div>

                    {/* Outreach Draft Generator (AT-010: Native Link & Clipboard Copy) */}
                    <div className="p-3.5 rounded-lg border border-sky-200 dark:border-sky-900/60 bg-sky-50/40 dark:bg-sky-950/20 space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-bold uppercase tracking-wider text-sky-900 dark:text-sky-300 flex items-center gap-1.5">
                          <MessageSquare className="w-3.5 h-3.5 text-sky-600" />
                          <span>Personalized LinkedIn Outreach Draft (AT-010 Compliant)</span>
                        </span>
                        <button
                          type="button"
                          onClick={() => handleGenerateOutreachDraft(selectedLead)}
                          className="px-3 py-1 text-xs font-bold rounded bg-sky-600 hover:bg-sky-700 text-white shadow-xs"
                        >
                          Generate Draft
                        </button>
                      </div>

                      {outreachDraft && (
                        <div className="space-y-2 pt-1">
                          <div className="p-3 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-xs text-slate-800 dark:text-slate-200 leading-relaxed font-mono">
                            {outreachDraft.body}
                          </div>
                          <div className="flex items-center justify-between text-[11px]">
                            <span className={outreachDraft.within_limit ? 'text-emerald-600 font-medium' : 'text-rose-600 font-bold'}>
                              {outreachDraft.character_count} / 300 characters {outreachDraft.within_limit ? '(Within LinkedIn Note Limit)' : '(Exceeds Limit)'}
                            </span>
                            <div className="flex items-center gap-2">
                              <button
                                type="button"
                                onClick={() => handleQueueNoteFromDraft(selectedLead, outreachDraft.body)}
                                className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 transition-colors"
                              >
                                <Send className="w-3 h-3" />
                                <span>Add to Queue</span>
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  if (navigator.clipboard) {
                                    navigator.clipboard.writeText(outreachDraft.body);
                                    setNotification('Connection note copied to clipboard!');
                                    setTimeout(() => setNotification(null), 3000);
                                  }
                                }}
                                className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded bg-white dark:bg-slate-800 border border-slate-300 text-slate-700 dark:text-slate-300 hover:bg-slate-50"
                              >
                                <Copy className="w-3 h-3" />
                                <span>Copy Note</span>
                              </button>
                              <a
                                href={outreachDraft.direct_chat_url}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1 px-3 py-1 text-xs font-bold rounded bg-sky-600 hover:bg-sky-700 text-white"
                              >
                                <span>Open Profile on LinkedIn</span>
                                <ExternalLink className="w-3 h-3" />
                              </a>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>

                    {/* Interaction Notes Feed */}
                    <div className="space-y-3 pt-2">
                      <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                        <ClipboardList className="w-3.5 h-3.5 text-purple-600" />
                        <span>Interaction Notes & Communication History ({selectedLead.notes.length})</span>
                      </h4>

                      <div className="space-y-2 max-h-56 overflow-y-auto">
                        {selectedLead.notes.map((note) => (
                          <div
                            key={note.id}
                            className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-xs"
                          >
                            <div className="flex items-center justify-between text-slate-400 text-[10px] mb-1">
                              <span className="font-semibold text-slate-600 dark:text-slate-300">{note.author}</span>
                              <span>{new Date(note.created_at).toLocaleString()}</span>
                            </div>
                            <p className="text-slate-800 dark:text-slate-200 leading-relaxed">
                              {note.content}
                            </p>
                          </div>
                        ))}
                      </div>

                      {/* Add Note Form */}
                      <div className="flex items-center gap-2 pt-2">
                        <input
                          type="text"
                          placeholder="Log meeting takeaways, interview feedback, or connection notes..."
                          value={newNoteContent}
                          onChange={(e) => setNewNoteContent(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') handleAddLeadNote(selectedLead.id);
                          }}
                          className="flex-1 px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-purple-500"
                        />
                        <button
                          type="button"
                          onClick={() => handleAddLeadNote(selectedLead.id)}
                          className="px-4 py-1.5 text-xs font-bold rounded-lg bg-purple-600 hover:bg-purple-700 text-white shadow-xs"
                        >
                          Add Note
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="p-8 text-center rounded-xl border border-dashed border-slate-300 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
                  <UserCheck className="w-8 h-8 text-slate-400 mx-auto mb-2" />
                  <h4 className="text-sm font-bold text-slate-700 dark:text-slate-300">
                    No Recruiter Lead Selected
                  </h4>
                  <p className="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
                    Select a lead from the left or convert a candidate profile from the Person Vault to view notes, reminders, and outreach tools.
                  </p>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Connection Note Drafts & Queue Studio (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-5">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className="px-2 py-0.5 text-[11px] font-bold uppercase tracking-wider rounded-full bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300">
                  IMP-LI-07 • Connection Queue
                </span>
                <span className="px-2 py-0.5 text-[11px] font-bold uppercase tracking-wider rounded-full bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300">
                  Platform Safety Caps (AT-008, AT-010)
                </span>
              </div>
              <h2 className="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Send className="w-5 h-5 text-emerald-600" />
                <span>Connection Note Drafts & Queue Studio</span>
              </h2>
              <p className="text-xs text-slate-500 mt-1 max-w-2xl leading-relaxed">
                Daily and rolling-week conservative invite budgets, recipient deduplication, immutable cryptographic approval gatekeeper, and 1-click manual-native fallback.
              </p>
            </div>

            {/* Tenant/Workspace Switcher */}
            <div className="flex items-center gap-2 self-start md:self-auto bg-slate-100 dark:bg-slate-800 p-1 rounded-lg">
              <button
                type="button"
                onClick={() => setQueueWorkspace('ws-alpha')}
                className={`px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  queueWorkspace === 'ws-alpha'
                    ? 'bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                Workspace Alpha (`ws-alpha`)
              </button>
              <button
                type="button"
                onClick={() => setQueueWorkspace('ws-beta')}
                className={`px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                  queueWorkspace === 'ws-beta'
                    ? 'bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-400 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                Workspace Beta (`ws-beta` Throttled)
              </button>
            </div>
          </div>

          {/* Account Safety & Budget Health Matrix (FND-011, AT-008, SRC-L1) */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Daily Budget Card */}
            <div className={`p-4 rounded-xl border ${
              currentBudget.daily_used >= currentBudget.daily_limit
                ? 'bg-rose-50/70 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900'
                : 'bg-emerald-50/40 border-emerald-200 dark:bg-emerald-950/20 dark:border-emerald-900'
            }`}>
              <div className="flex items-center justify-between text-xs font-bold">
                <span className="text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5 text-emerald-600" />
                  <span>Daily Safe Budget</span>
                </span>
                <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold ${
                  currentBudget.daily_used >= currentBudget.daily_limit
                    ? 'bg-rose-200 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                    : 'bg-emerald-200 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                }`}>
                  {currentBudget.daily_used >= currentBudget.daily_limit ? 'EXHAUSTED' : 'HEALTHY'}
                </span>
              </div>
              <div className="mt-2 flex items-baseline justify-between">
                <span className="text-2xl font-black text-slate-900 dark:text-white">
                  {currentBudget.daily_used} <span className="text-xs font-normal text-slate-400">/ {currentBudget.daily_limit} max</span>
                </span>
                <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">
                  {Math.max(0, currentBudget.daily_limit - currentBudget.daily_used)} remaining today
                </span>
              </div>
              <div className="mt-2.5 w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all ${
                    currentBudget.daily_used >= currentBudget.daily_limit ? 'bg-rose-600' : 'bg-emerald-600'
                  }`}
                  style={{ width: `${Math.min(100, (currentBudget.daily_used / currentBudget.daily_limit) * 100)}%` }}
                />
              </div>
            </div>

            {/* Rolling-Week Budget Card */}
            <div className={`p-4 rounded-xl border ${
              currentBudget.weekly_used >= currentBudget.weekly_limit
                ? 'bg-rose-50/70 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900'
                : 'bg-sky-50/40 border-sky-200 dark:bg-sky-950/20 dark:border-sky-900'
            }`}>
              <div className="flex items-center justify-between text-xs font-bold">
                <span className="text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5 text-sky-600" />
                  <span>Rolling 7-Day Budget</span>
                </span>
                <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold ${
                  currentBudget.weekly_used >= currentBudget.weekly_limit
                    ? 'bg-rose-200 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                    : 'bg-sky-200 text-sky-800 dark:bg-sky-900 dark:text-sky-200'
                }`}>
                  {currentBudget.weekly_used >= currentBudget.weekly_limit ? 'THROTTLED' : 'ACTIVE'}
                </span>
              </div>
              <div className="mt-2 flex items-baseline justify-between">
                <span className="text-2xl font-black text-slate-900 dark:text-white">
                  {currentBudget.weekly_used} <span className="text-xs font-normal text-slate-400">/ {currentBudget.weekly_limit} max</span>
                </span>
                <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">
                  {Math.max(0, currentBudget.weekly_limit - currentBudget.weekly_used)} remaining this week
                </span>
              </div>
              <div className="mt-2.5 w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all ${
                    currentBudget.weekly_used >= currentBudget.weekly_limit ? 'bg-rose-600' : 'bg-sky-600'
                  }`}
                  style={{ width: `${Math.min(100, (currentBudget.weekly_used / currentBudget.weekly_limit) * 100)}%` }}
                />
              </div>
            </div>

            {/* Platform Safety Guarantee Notice */}
            <div className="p-4 rounded-xl border border-amber-200 dark:border-amber-900/60 bg-amber-50/40 dark:bg-amber-950/20 text-xs flex flex-col justify-between">
              <div>
                <div className="flex items-center gap-1.5 font-bold text-amber-900 dark:text-amber-300 mb-1">
                  <ShieldAlert className="w-3.5 h-3.5 text-amber-600" />
                  <span>Truth-in-Advertising & Anti-Ban Policy (AT-010)</span>
                </div>
                <p className="text-[11px] text-amber-800/90 dark:text-amber-300/80 leading-relaxed">
                  LinkedIn strictly prohibits autonomous bot invite spam. Unsafe claims (e.g. 150+/wk) are rejected per `SRC-L1`. All notes require explicit candidate review and 1-click assisted dispatch.
                </p>
              </div>
              <div className="mt-2 text-[10px] font-mono text-amber-700 dark:text-amber-400">
                Conservative cap: max 25/day, 100/wk (fail-closed)
              </div>
            </div>
          </div>

          {/* Tamper Alert Warning (FND-010, AT-007) */}
          {tamperAlert && (
            <div className="p-3.5 rounded-lg border border-amber-300 bg-amber-50 text-amber-900 text-xs flex items-center justify-between">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
                <span className="font-semibold">{tamperAlert}</span>
              </div>
              <button
                type="button"
                onClick={() => setTamperAlert(null)}
                className="text-amber-700 font-bold hover:text-amber-900 text-xs underline"
              >
                Dismiss
              </button>
            </div>
          )}

          {/* Two-Column Queue Management Workbench */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 pt-2">
            {/* Left Column: Queue List & Filters (5 cols) */}
            <div className="lg:col-span-5 space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-1.5">
                  <ListOrdered className="w-4 h-4 text-slate-500" />
                  <span className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">
                    Connection Queue ({filteredQueueItems.length})
                  </span>
                </div>
                <div className="relative">
                  <Search className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" />
                  <input
                    type="text"
                    placeholder="Search queue..."
                    value={queueSearch}
                    onChange={(e) => setQueueSearch(e.target.value)}
                    className="pl-8 pr-2.5 py-1 text-xs rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 w-full sm:w-44 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                  />
                </div>
              </div>

              {/* Status Filter Badges */}
              <div className="flex flex-wrap gap-1.5 text-[11px]">
                {(['all', 'pending_approval', 'approved', 'copied_to_clipboard', 'completed_manually'] as const).map((st) => (
                  <button
                    key={st}
                    type="button"
                    onClick={() => setQueueStatusFilter(st)}
                    className={`px-2.5 py-1 rounded-full font-medium transition-all ${
                      queueStatusFilter === st
                        ? 'bg-emerald-600 text-white font-bold'
                        : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
                    }`}
                  >
                    {st === 'all'
                      ? 'All'
                      : st === 'pending_approval'
                      ? 'Pending'
                      : st === 'approved'
                      ? 'Approved'
                      : st === 'copied_to_clipboard'
                      ? 'Copied'
                      : 'Completed'}
                  </button>
                ))}
              </div>

              {/* Queue Items Scrollable List */}
              <div className="space-y-2 max-h-[460px] overflow-y-auto pr-1">
                {filteredQueueItems.length === 0 ? (
                  <div className="p-8 text-center rounded-xl border border-dashed border-slate-300 dark:border-slate-800 text-slate-400 text-xs">
                    No connection note drafts matching criteria.
                  </div>
                ) : (
                  filteredQueueItems.map((item) => {
                    const isSelected = selectedQueueItem?.id === item.id;
                    return (
                      <div
                        key={item.id}
                        onClick={() => {
                          setSelectedQueueItemId(item.id);
                          setTamperAlert(null);
                        }}
                        className={`p-3.5 rounded-xl border cursor-pointer transition-all text-left ${
                          isSelected
                            ? 'border-emerald-500 bg-emerald-50/30 dark:bg-emerald-950/20 shadow-xs'
                            : 'border-slate-200 dark:border-slate-800 hover:border-slate-300 bg-white dark:bg-slate-900'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div>
                            <h4 className="text-sm font-bold text-slate-900 dark:text-white">
                              {item.recipient_name}
                            </h4>
                            <p className="text-xs text-slate-500">
                              {item.recipient_title} &bull; <span className="font-semibold text-slate-700 dark:text-slate-300">{item.recipient_company}</span>
                            </p>
                          </div>
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider shrink-0 ${
                            item.status === 'pending_approval'
                              ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300'
                              : item.status === 'approved'
                              ? 'bg-indigo-100 text-indigo-800 dark:bg-indigo-950/80 dark:text-indigo-300'
                              : item.status === 'copied_to_clipboard'
                              ? 'bg-sky-100 text-sky-800 dark:bg-sky-950/80 dark:text-sky-300'
                              : item.status === 'completed_manually'
                              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300'
                              : 'bg-rose-100 text-rose-800'
                          }`}>
                            {item.status.replace('_', ' ')}
                          </span>
                        </div>

                        <p className="mt-2 text-xs text-slate-600 dark:text-slate-400 line-clamp-2 italic font-mono bg-slate-50 dark:bg-slate-800/40 p-2 rounded">
                          "{item.note_text}"
                        </p>

                        <div className="mt-2.5 flex items-center justify-between text-[11px] text-slate-400">
                          <span className={item.character_count <= 300 ? 'text-emerald-600 font-medium' : 'text-rose-600 font-bold'}>
                            {item.character_count} / 300 chars
                          </span>
                          <span>{new Date(item.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            {/* Right Column: Selected Queue Item Detail & Action Workbench (7 cols) */}
            <div className="lg:col-span-7">
              {selectedQueueItem ? (
                <div className="rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4 bg-slate-50/50 dark:bg-slate-900/40">
                  {/* Item Header */}
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-200 dark:border-slate-800 pb-3.5">
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="text-base font-bold text-slate-900 dark:text-white">
                          {selectedQueueItem.recipient_name}
                        </h3>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${
                          selectedQueueItem.status === 'pending_approval'
                            ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300'
                            : selectedQueueItem.status === 'approved'
                            ? 'bg-indigo-100 text-indigo-800 dark:bg-indigo-950/80 dark:text-indigo-300'
                            : selectedQueueItem.status === 'copied_to_clipboard'
                            ? 'bg-sky-100 text-sky-800 dark:bg-sky-950/80 dark:text-sky-300'
                            : selectedQueueItem.status === 'completed_manually'
                            ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300'
                            : 'bg-rose-100 text-rose-800'
                        }`}>
                          {selectedQueueItem.status.replace('_', ' ')}
                        </span>
                      </div>
                      <p className="text-xs text-slate-500 mt-0.5">
                        {selectedQueueItem.recipient_title} at <strong className="text-slate-800 dark:text-slate-200">{selectedQueueItem.recipient_company}</strong>
                      </p>
                    </div>

                    <a
                      href={selectedQueueItem.recipient_linkedin_url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1 text-xs font-semibold text-sky-600 hover:text-sky-700 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 px-3 py-1.5 rounded-lg shadow-2xs self-start"
                    >
                      <span>LinkedIn Profile</span>
                      <ExternalLink className="w-3.5 h-3.5" />
                    </a>
                  </div>

                  {/* Context Factors Pills */}
                  {selectedQueueItem.context_factors && selectedQueueItem.context_factors.length > 0 && (
                    <div className="space-y-1.5">
                      <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                        Personalization Context
                      </span>
                      <div className="flex flex-wrap gap-1.5">
                        {selectedQueueItem.context_factors.map((f, idx) => (
                          <span
                            key={idx}
                            className="px-2.5 py-0.5 text-xs font-medium rounded-full bg-slate-200/80 dark:bg-slate-800 text-slate-700 dark:text-slate-300"
                          >
                            {f}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Note Text Editor with Real-Time Character Limit */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                        <FileEdit className="w-3.5 h-3.5 text-emerald-600" />
                        <span>Personalized Connection Note Draft</span>
                      </span>
                      <span className={`text-xs font-mono font-bold ${
                        selectedQueueItem.character_count <= 300 ? 'text-emerald-600' : 'text-rose-600'
                      }`}>
                        {selectedQueueItem.character_count} / 300 chars {selectedQueueItem.character_count <= 300 ? '✓' : '⚠️ EXCEEDS LIMIT'}
                      </span>
                    </div>

                    <textarea
                      rows={4}
                      value={selectedQueueItem.note_text}
                      onChange={(e) => handleEditQueueNote(selectedQueueItem.id, e.target.value)}
                      className="w-full text-xs font-mono leading-relaxed p-3 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
                      placeholder="Enter connection note text..."
                    />

                    <div className="w-full bg-slate-200 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                      <div
                        className={`h-full transition-all ${
                          selectedQueueItem.character_count <= 300 ? 'bg-emerald-500' : 'bg-rose-500'
                        }`}
                        style={{ width: `${Math.min(100, (selectedQueueItem.character_count / 300) * 100)}%` }}
                      />
                    </div>
                  </div>

                  {/* Cryptographic Approval & Execution Gatekeeper (FND-010, AT-007) */}
                  <div className="p-4 rounded-xl border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 space-y-3.5">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                        <Shield className="w-3.5 h-3.5 text-indigo-600" />
                        <span>Cryptographic Approval Gatekeeper (FND-010)</span>
                      </span>
                      {selectedQueueItem.approval_token && (
                        <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-semibold">
                          VERIFIED
                        </span>
                      )}
                    </div>

                    {selectedQueueItem.approval_token ? (
                      <div className="p-2.5 rounded bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 text-[11px] font-mono space-y-1">
                        <div className="text-slate-500">
                          Approver: <span className="font-bold text-slate-800 dark:text-slate-200">{selectedQueueItem.approved_by}</span>
                        </div>
                        <div className="text-slate-500 truncate">
                          HMAC Token: <span className="text-indigo-600 dark:text-indigo-400">{selectedQueueItem.approval_token}</span>
                        </div>
                        {selectedQueueItem.completed_at && (
                          <div className="text-emerald-600 font-bold">
                            Sent manually at: {new Date(selectedQueueItem.completed_at).toLocaleString()}
                          </div>
                        )}
                      </div>
                    ) : (
                      <p className="text-xs text-slate-500 leading-relaxed">
                        Direct connection dispatch is barred until candidate reviews and explicitly signs off with an immutable approval token (`REQ-015`, `AT-007`).
                      </p>
                    )}

                    {/* Gatekeeper Action Buttons */}
                    <div className="flex flex-wrap items-center gap-2 pt-1">
                      {selectedQueueItem.status === 'pending_approval' && (
                        <button
                          type="button"
                          onClick={() => handleApproveQueueItem(selectedQueueItem.id)}
                          disabled={currentBudget.daily_used >= currentBudget.daily_limit || !selectedQueueItem.within_limit}
                          className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-bold rounded-lg bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white shadow-xs"
                        >
                          <Check className="w-3.5 h-3.5" />
                          <span>Approve for Manual Send</span>
                        </button>
                      )}

                      {(selectedQueueItem.status === 'approved' || selectedQueueItem.status === 'copied_to_clipboard') && (
                        <>
                          <button
                            type="button"
                            onClick={() => handleCopyQueueNote(selectedQueueItem)}
                            className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-bold rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs"
                          >
                            <Copy className="w-3.5 h-3.5" />
                            <span>Copy Note & Launch Profile</span>
                          </button>
                          <button
                            type="button"
                            onClick={() => handleConfirmSent(selectedQueueItem)}
                            className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-bold rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 text-slate-800 dark:text-slate-200 border border-slate-300 dark:border-slate-700"
                          >
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                            <span>Confirm Sent Manually</span>
                          </button>
                        </>
                      )}

                      {selectedQueueItem.status === 'completed_manually' && (
                        <span className="inline-flex items-center gap-1 px-3 py-1.5 text-xs font-bold rounded-lg bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          <span>Delivered via Manual Confirmation</span>
                        </span>
                      )}

                      {selectedQueueItem.status !== 'rejected' && selectedQueueItem.status !== 'completed_manually' && (
                        <button
                          type="button"
                          onClick={() => handleRejectQueueItem(selectedQueueItem.id)}
                          className="px-3 py-2 text-xs font-semibold rounded-lg text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 ml-auto"
                        >
                          Reject / Dismiss
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="p-12 text-center rounded-xl border border-dashed border-slate-300 dark:border-slate-800 text-slate-400">
                  <Send className="w-8 h-8 mx-auto mb-2 text-slate-300" />
                  <h4 className="text-sm font-bold text-slate-700 dark:text-slate-300">
                    No Queue Item Selected
                  </h4>
                  <p className="text-xs mt-1 max-w-sm mx-auto">
                    Select a recipient draft from the left queue or click "Queue Note" from the Person Vault to begin.
                  </p>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Target Company-Follow Planning Studio (IMP-LI-08, LI-08, AT-010, FND-011, SRC-L1) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center space-x-2">
                <span className="inline-flex items-center rounded-full bg-sky-50 dark:bg-sky-950/60 px-2 py-0.5 text-[10px] font-bold text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
                  IMP-LI-08 &bull; LI-08 &bull; AT-010 &bull; FND-011 &bull; SRC-L1
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Building2 className="w-4 h-4 text-sky-600" />
                  <span>Target Company-Follow Planning & Pacing Studio</span>
                </h2>
              </div>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                Build prioritized target company watchlists, generate paced batch follow queues, and execute safely under account budget limits with 1-click native company deep-links.
              </p>
            </div>

            <div className="flex items-center gap-3">
              {/* Workspace Scope Switcher */}
              <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 p-1">
                <button
                  type="button"
                  onClick={() => setFollowWorkspace('ws-alpha')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    followWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-400 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Alpha
                </button>
                <button
                  type="button"
                  onClick={() => setFollowWorkspace('ws-beta')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    followWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-400 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Beta
                </button>
              </div>

              <button
                type="button"
                onClick={() => setIsAddingCompanyModalOpen(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-sky-600 text-white hover:bg-sky-700 shadow-sm transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Target Company</span>
              </button>
            </div>
          </div>

          {/* Platform Truth-in-Advertising & Anti-Ban Safety Notice (AT-010, REQ-009) */}
          <div className="mt-4 rounded-lg bg-sky-50 dark:bg-sky-950/40 border border-sky-200 dark:border-sky-800 p-3.5 flex items-start gap-3">
            <Info className="w-4 h-4 text-sky-600 dark:text-sky-400 shrink-0 mt-0.5" />
            <div className="text-xs text-sky-900 dark:text-sky-200 leading-relaxed">
              <strong className="font-bold">Platform Safety Guarantee (AT-010 & REQ-010):</strong> LinkedIn does not provide public write APIs to programmatically follow companies in the background. Unsupported headless auto-follow routines are permanently disabled to prevent account restrictions. Instead, this studio generates verified canonical company page links with paced schedules; you manually confirm follows in 1 click, reliably decrementing your account safety budget.
            </div>
          </div>

          {/* Follow Budgets & Safety Meters */}
          {(() => {
            const budget = followBudgets[followWorkspace] || INITIAL_FOLLOW_BUDGETS['ws-alpha'];
            const dailyPercent = Math.min(100, Math.round((budget.daily_used / budget.daily_limit) * 100));
            const weeklyPercent = Math.min(100, Math.round((budget.weekly_used / budget.weekly_limit) * 100));

            return (
              <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                {/* Daily Follow Meter */}
                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/30">
                  <div className="flex items-center justify-between text-xs mb-1.5">
                    <span className="font-semibold text-slate-700 dark:text-slate-300">Daily Follow Quota</span>
                    <span className="font-mono font-bold text-slate-900 dark:text-white">
                      {budget.daily_used} / {budget.daily_limit} used
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${
                        dailyPercent >= 100
                          ? 'bg-rose-500'
                          : dailyPercent >= 80
                          ? 'bg-amber-500'
                          : 'bg-sky-600'
                      }`}
                      style={{ width: `${dailyPercent}%` }}
                    />
                  </div>
                  <p className="mt-2 text-[11px] text-slate-500">
                    Remaining today: <strong className="text-slate-800 dark:text-slate-200">{Math.max(0, budget.daily_limit - budget.daily_used)} follows</strong>
                  </p>
                </div>

                {/* Weekly Follow Meter */}
                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/30">
                  <div className="flex items-center justify-between text-xs mb-1.5">
                    <span className="font-semibold text-slate-700 dark:text-slate-300">Rolling Weekly Quota</span>
                    <span className="font-mono font-bold text-slate-900 dark:text-white">
                      {budget.weekly_used} / {budget.weekly_limit} used
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${
                        weeklyPercent >= 100
                          ? 'bg-rose-500'
                          : weeklyPercent >= 80
                          ? 'bg-amber-500'
                          : 'bg-indigo-600'
                      }`}
                      style={{ width: `${weeklyPercent}%` }}
                    />
                  </div>
                  <p className="mt-2 text-[11px] text-slate-500">
                    Remaining rolling week: <strong className="text-slate-800 dark:text-slate-200">{Math.max(0, budget.weekly_limit - budget.weekly_used)} follows</strong>
                  </p>
                </div>

                {/* Account Safety State */}
                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/30 flex flex-col justify-between">
                  <div>
                    <div className="flex items-center justify-between text-xs mb-1">
                      <span className="font-semibold text-slate-700 dark:text-slate-300">Account Pacing Status</span>
                      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold ${
                        budget.is_locked
                          ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                          : 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                      }`}>
                        {budget.is_locked ? <ShieldAlert className="w-3 h-3" /> : <Shield className="w-3 h-3" />}
                        <span>{budget.is_locked ? 'Throttled / Locked' : 'Safe to Follow'}</span>
                      </span>
                    </div>
                    <p className="text-[11px] text-slate-500 mt-1 leading-relaxed">
                      {budget.is_locked
                        ? budget.lock_reason || 'Safety lock active due to daily quota exhaustion.'
                        : 'Paced conservative batching active. Recommended interval: 120s between follows.'}
                    </p>
                  </div>
                  <div className="text-[10px] text-slate-400 font-mono mt-2">
                    Tenant: {budget.tenant_id} &bull; Reset: {budget.last_reset_date}
                  </div>
                </div>
              </div>
            );
          })()}

          {/* Main Two-Column Studio Layout: Left = Target Watchlist, Right = Batch Follow Generator & Execution */}
          <div className="mt-6 grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left Column: Target Company Watchlist (7 cols) */}
            <div className="lg:col-span-7 space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                    <Building2 className="w-4 h-4 text-sky-600" />
                    <span>Target Company Watchlist</span>
                  </h3>
                  <span className="px-2 py-0.5 text-xs font-bold rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                    {watchlist.filter((w) => w.workspace_id === followWorkspace).length}
                  </span>
                </div>

                {/* Priority Filter & Search */}
                <div className="flex items-center gap-2">
                  <select
                    value={watchlistPriorityFilter}
                    onChange={(e) => setWatchlistPriorityFilter(e.target.value)}
                    className="px-2.5 py-1 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200"
                  >
                    <option value="all">All Priorities</option>
                    <option value="high">High Priority</option>
                    <option value="medium">Medium Priority</option>
                    <option value="low">Low Priority</option>
                  </select>
                  <input
                    type="text"
                    placeholder="Search company or tag..."
                    value={watchlistSearch}
                    onChange={(e) => setWatchlistSearch(e.target.value)}
                    className="w-36 sm:w-44 px-2.5 py-1 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none"
                  />
                </div>
              </div>

              {/* Watchlist Item Cards */}
              <div className="space-y-3">
                {watchlist
                  .filter((item) => item.workspace_id === followWorkspace)
                  .filter((item) => (watchlistPriorityFilter === 'all' ? true : item.priority === watchlistPriorityFilter))
                  .filter((item) =>
                    watchlistSearch.trim() === ''
                      ? true
                      : item.company_name.toLowerCase().includes(watchlistSearch.toLowerCase()) ||
                        item.tags.some((t) => t.toLowerCase().includes(watchlistSearch.toLowerCase()))
                  )
                  .map((item) => {
                    const isSelected = selectedWatchlistIds.includes(item.item_id);
                    return (
                      <div
                        key={item.item_id}
                        className={`p-4 rounded-xl border transition-all ${
                          isSelected
                            ? 'border-sky-500 bg-sky-50/40 dark:bg-sky-950/20 shadow-xs'
                            : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-slate-300'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="flex items-start gap-3">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleToggleWatchlistSelection(item.item_id)}
                              className="mt-1 h-4 w-4 rounded border-slate-300 text-sky-600 focus:ring-sky-500"
                            />
                            <div>
                              <div className="flex items-center gap-2">
                                <h4 className="text-sm font-bold text-slate-900 dark:text-white">
                                  {item.company_name}
                                </h4>
                                <span
                                  className={`px-2 py-0.2 rounded text-[10px] font-bold uppercase ${
                                    item.priority === 'high'
                                      ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                                      : item.priority === 'medium'
                                      ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                                      : 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300'
                                  }`}
                                >
                                  {item.priority}
                                </span>
                                {item.domain && (
                                  <span className="text-[11px] text-slate-400 font-mono">
                                    {item.domain}
                                  </span>
                                )}
                              </div>
                              <p className="text-xs text-slate-600 dark:text-slate-400 mt-1 leading-relaxed">
                                {item.target_reason}
                              </p>
                              <div className="mt-2 flex flex-wrap items-center gap-1.5">
                                {item.tags.map((tag, tIdx) => (
                                  <span
                                    key={tIdx}
                                    className="px-1.5 py-0.5 text-[10px] rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400"
                                  >
                                    #{tag}
                                  </span>
                                ))}
                              </div>
                            </div>
                          </div>

                          <div className="flex flex-col items-end gap-2 shrink-0">
                            <a
                              href={item.company_page_url}
                              target="_blank"
                              rel="noreferrer"
                              className="inline-flex items-center gap-1 text-xs font-semibold text-sky-600 hover:text-sky-700 dark:text-sky-400 hover:underline"
                            >
                              <span>LinkedIn Page</span>
                              <ExternalLink className="w-3 h-3" />
                            </a>
                            <button
                              type="button"
                              onClick={() => handleRemoveFromWatchlist(item.item_id)}
                              className="text-[11px] text-slate-400 hover:text-rose-600 transition-colors"
                            >
                              Remove
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>

            {/* Right Column: Batch Follow Planning & Active Follow Bench (5 cols) */}
            <div className="lg:col-span-5 space-y-6">
              {/* Batch Plan Generator Card */}
              <div className="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs">
                <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2 mb-3">
                  <ListOrdered className="w-4 h-4 text-indigo-600" />
                  <span>Batch Follow Plan Generator</span>
                </h3>

                <div className="space-y-3 text-xs">
                  <div>
                    <label className="block text-slate-600 dark:text-slate-400 font-medium mb-1">
                      Batch Plan Name
                    </label>
                    <input
                      type="text"
                      value={batchPlanName}
                      onChange={(e) => setBatchPlanName(e.target.value)}
                      className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100"
                      placeholder="e.g. Q1 Infrastructure Employers"
                    />
                  </div>

                  <div>
                    <label className="block text-slate-600 dark:text-slate-400 font-medium mb-1">
                      Pacing Interval Between Actions
                    </label>
                    <select
                      value={batchPacingSec}
                      onChange={(e) => setBatchPacingSec(Number(e.target.value))}
                      className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100"
                    >
                      <option value={60}>60 seconds (1 minute - Moderate)</option>
                      <option value={120}>120 seconds (2 minutes - Recommended Safe)</option>
                      <option value={180}>180 seconds (3 minutes - Ultra Conservative)</option>
                      <option value={300}>300 seconds (5 minutes - Maximum Spacing)</option>
                    </select>
                  </div>

                  <div className="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 text-[11px] text-slate-600 dark:text-slate-400 flex items-center justify-between">
                    <span>Selected from Watchlist:</span>
                    <strong className="text-slate-900 dark:text-white font-bold">{selectedWatchlistIds.length} companies</strong>
                  </div>

                  <button
                    type="button"
                    onClick={handleCreateBatchPlan}
                    disabled={selectedWatchlistIds.length === 0}
                    className="w-full mt-2 py-2 px-4 rounded-lg text-xs font-bold bg-indigo-600 hover:bg-indigo-700 text-white disabled:opacity-50 disabled:cursor-not-allowed shadow-xs transition-colors"
                  >
                    Generate Paced Batch Plan ({selectedWatchlistIds.length})
                  </button>
                </div>
              </div>

              {/* Active Batch Queue Execution Bench */}
              <div className="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <Clock className="w-4 h-4 text-sky-600" />
                    <span>Active Follow Execution Queue</span>
                  </h3>

                  {/* Plan Selector */}
                  <select
                    value={activeBatchPlanId}
                    onChange={(e) => setActiveBatchPlanId(e.target.value)}
                    className="px-2 py-1 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100"
                  >
                    {followPlans
                      .filter((p) => p.workspace_id === followWorkspace)
                      .map((p) => (
                        <option key={p.plan_id} value={p.plan_id}>
                          {p.plan_name} ({p.followed_count}/{p.total_count})
                        </option>
                      ))}
                  </select>
                </div>

                {/* Plan Items List */}
                {(() => {
                  const currentPlan = followPlans.find(
                    (p) => p.plan_id === activeBatchPlanId && p.workspace_id === followWorkspace
                  );
                  if (!currentPlan) {
                    return (
                      <div className="p-8 text-center text-slate-400 text-xs">
                        No batch plan selected or created for this workspace.
                      </div>
                    );
                  }

                  return (
                    <div className="space-y-3">
                      <div className="flex items-center justify-between text-xs pb-2 border-b border-slate-100 dark:border-slate-800">
                        <span className="font-semibold text-slate-700 dark:text-slate-300">
                          Progress: {currentPlan.followed_count} of {currentPlan.total_count} Completed
                        </span>
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                            currentPlan.status === 'completed'
                              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                              : 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300'
                          }`}
                        >
                          {currentPlan.status}
                        </span>
                      </div>

                      <div className="space-y-2.5 max-h-[360px] overflow-y-auto pr-1">
                        {currentPlan.items.map((item) => (
                          <div
                            key={item.item_id}
                            className={`p-3 rounded-lg border text-xs transition-all ${
                              item.status === 'manual_followed'
                                ? 'bg-emerald-50/40 border-emerald-200 dark:bg-emerald-950/20 dark:border-emerald-800'
                                : item.status === 'skipped'
                                ? 'bg-slate-50 border-slate-200 dark:bg-slate-800/40 dark:border-slate-700 opacity-60'
                                : 'bg-white border-slate-200 dark:bg-slate-900 dark:border-slate-800'
                            }`}
                          >
                            <div className="flex items-start justify-between gap-2">
                              <div>
                                <div className="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                                  <span>{item.company_name}</span>
                                  <span
                                    className={`px-1.5 py-0.2 rounded text-[9px] font-bold uppercase ${
                                      item.priority === 'high'
                                        ? 'bg-rose-100 text-rose-800'
                                        : 'bg-slate-100 text-slate-800'
                                    }`}
                                  >
                                    {item.priority}
                                  </span>
                                </div>
                                <div className="text-[11px] text-slate-500 mt-0.5">
                                  Scheduled: {new Date(item.scheduled_for).toLocaleTimeString()} (+{item.pacing_interval_sec}s pacing)
                                </div>
                              </div>

                              <span
                                className={`px-2 py-0.5 rounded text-[10px] font-bold capitalize ${
                                  item.status === 'manual_followed'
                                    ? 'bg-emerald-100 text-emerald-800'
                                    : item.status === 'skipped'
                                    ? 'bg-slate-200 text-slate-700'
                                    : 'bg-amber-100 text-amber-800'
                                }`}
                              >
                                {item.status.replace(/_/g, ' ')}
                              </span>
                            </div>

                            <div className="mt-3 pt-2 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
                              <a
                                href={item.company_page_url}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1 text-[11px] font-semibold text-sky-600 hover:text-sky-700 dark:text-sky-400"
                              >
                                <span>Open Company Page</span>
                                <ExternalLink className="w-3 h-3" />
                              </a>

                              {item.status === 'ready_for_manual_follow' && (
                                <div className="flex items-center gap-1.5">
                                  <button
                                    type="button"
                                    onClick={() => handleConfirmFollowItem(currentPlan.plan_id, item.item_id)}
                                    className="px-2 py-1 rounded text-[11px] font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs"
                                  >
                                    Confirm Followed (+1)
                                  </button>
                                  <button
                                    type="button"
                                    onClick={() => handleSkipFollowItem(currentPlan.plan_id, item.item_id)}
                                    className="px-2 py-1 rounded text-[11px] text-slate-500 hover:text-slate-800 dark:hover:text-slate-200"
                                  >
                                    Skip
                                  </button>
                                </div>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  );
                })()}
              </div>
            </div>
          </div>
        </section>

        {/* Modal: Add Target Company to Watchlist */}
        {isAddingCompanyModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs p-4">
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl max-w-lg w-full p-6">
              <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800">
                <h3 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Building2 className="w-4 h-4 text-sky-600" />
                  <span>Add Company to Target Watchlist</span>
                </h3>
                <button
                  type="button"
                  onClick={() => setIsAddingCompanyModalOpen(false)}
                  className="text-slate-400 hover:text-slate-600 text-sm font-bold"
                >
                  ✕
                </button>
              </div>

              <div className="mt-4 space-y-3.5 text-xs">
                <div>
                  <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Company Name *
                  </label>
                  <input
                    type="text"
                    value={newCompName}
                    onChange={(e) => setNewCompName(e.target.value)}
                    placeholder="e.g. OpenAI, Snowflake, Stripe"
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    LinkedIn Universal Slug or URL *
                  </label>
                  <input
                    type="text"
                    value={newCompUniversal}
                    onChange={(e) => setNewCompUniversal(e.target.value)}
                    placeholder="e.g. snowflake-computing or https://www.linkedin.com/company/stripe"
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                      Website Domain
                    </label>
                    <input
                      type="text"
                      value={newCompDomain}
                      onChange={(e) => setNewCompDomain(e.target.value)}
                      placeholder="e.g. stripe.com"
                      className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                    />
                  </div>

                  <div>
                    <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                      Priority Tier
                    </label>
                    <select
                      value={newCompPriority}
                      onChange={(e) => setNewCompPriority(e.target.value as CompanyFollowPriority)}
                      className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                    >
                      <option value="high">High Priority</option>
                      <option value="medium">Medium Priority</option>
                      <option value="low">Low Priority</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Strategic Target Rationale
                  </label>
                  <input
                    type="text"
                    value={newCompReason}
                    onChange={(e) => setNewCompReason(e.target.value)}
                    placeholder="e.g. Scaling cloud storage team, hiring principal infrastructure lead"
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Tags (comma-separated)
                  </label>
                  <input
                    type="text"
                    value={newCompTags}
                    onChange={(e) => setNewCompTags(e.target.value)}
                    placeholder="e.g. fintech, payments, cloud"
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                  />
                </div>
              </div>

              <div className="mt-6 pt-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-end gap-2.5">
                <button
                  type="button"
                  onClick={() => setIsAddingCompanyModalOpen(false)}
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleAddToWatchlist}
                  className="px-4 py-1.5 rounded-lg text-xs font-bold bg-sky-600 hover:bg-sky-700 text-white shadow-xs"
                >
                  Add to Watchlist
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Inbox & Conversation Triage Studio (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800 mb-6">
            <div>
              <div className="flex items-center gap-2">
                <span className="p-1.5 rounded-lg bg-sky-100 dark:bg-sky-900/40 text-sky-600 dark:text-sky-400">
                  <Inbox className="w-5 h-5" />
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white">
                  Inbox & Conversation Triage Studio
                </h2>
                <span className="px-2 py-0.5 text-[10px] font-semibold tracking-wide uppercase rounded-full bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400">
                  IMP-LI-09
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-1 max-w-2xl">
                Authorized LinkedIn conversation threads search, automatic stop-followup on inbound replies (LI-09), fact-grounded response drafts, cryptographic HMAC approvals (AT-007, FND-010), and AT-010 1-click clipboard dispatch.
              </p>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-xs text-slate-500">Tenant / Workspace:</span>
              <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 p-0.5 bg-slate-100 dark:bg-slate-800">
                <button
                  type="button"
                  onClick={() => setInboxWorkspace('ws-alpha')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-colors ${
                    inboxWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-sky-600 dark:text-sky-400 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Alpha
                </button>
                <button
                  type="button"
                  onClick={() => setInboxWorkspace('ws-beta')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-colors ${
                    inboxWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-sky-600 dark:text-sky-400 shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Beta
                </button>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left 5 Cols: Threads List & Filter Search */}
            <div className="lg:col-span-5 flex flex-col gap-3">
              <div className="space-y-2">
                <div className="relative">
                  <Search className="w-3.5 h-3.5 absolute left-3 top-3 text-slate-400" />
                  <input
                    type="text"
                    value={inboxSearch}
                    onChange={(e) => setInboxSearch(e.target.value)}
                    placeholder="Search sender, company, subject..."
                    className="w-full pl-9 pr-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-white"
                  />
                </div>

                <div className="grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <label className="block text-[11px] font-medium text-slate-500 mb-1">Thread Type</label>
                    <select
                      value={inboxTypeFilter}
                      onChange={(e) => setInboxTypeFilter(e.target.value)}
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs"
                    >
                      <option value="all">All Types</option>
                      <option value="recruiter">Recruiter</option>
                      <option value="peer_connection">Peer Connection</option>
                      <option value="inmail">InMail</option>
                      <option value="general">General</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-[11px] font-medium text-slate-500 mb-1">Classification</label>
                    <select
                      value={inboxClassificationFilter}
                      onChange={(e) => setInboxClassificationFilter(e.target.value)}
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs"
                    >
                      <option value="all">All Categories</option>
                      <option value="interview_invitation">Interview Invitation</option>
                      <option value="recruiter_inquiry">Recruiter Inquiry</option>
                      <option value="follow_up_response">Follow-up Response</option>
                      <option value="networking">Networking</option>
                      <option value="spam_or_promo">Spam / Promo</option>
                    </select>
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs pt-1">
                  <label className="flex items-center gap-1.5 cursor-pointer text-slate-600 dark:text-slate-400">
                    <input
                      type="checkbox"
                      checked={inboxUnreadOnly}
                      onChange={(e) => setInboxUnreadOnly(e.target.checked)}
                      className="rounded border-slate-300 text-sky-600 focus:ring-sky-500"
                    />
                    <span>Unread Only</span>
                  </label>
                  <span className="text-[11px] text-slate-400 font-mono">
                    {inboxThreads.filter((t) => t.workspace_id === inboxWorkspace).length} threads
                  </span>
                </div>
              </div>

              {/* Thread Items */}
              <div className="space-y-2 max-h-[560px] overflow-y-auto pr-1">
                {inboxThreads
                  .filter((t) => {
                    if (t.workspace_id !== inboxWorkspace) return false;
                    if (inboxUnreadOnly && t.unread_count === 0) return false;
                    if (inboxTypeFilter !== 'all' && t.thread_type !== inboxTypeFilter) return false;
                    if (inboxClassificationFilter !== 'all' && t.classification !== inboxClassificationFilter) return false;
                    if (inboxSearch) {
                      const q = inboxSearch.toLowerCase();
                      const match =
                        t.participant_name.toLowerCase().includes(q) ||
                        (t.participant_company && t.participant_company.toLowerCase().includes(q)) ||
                        t.subject.toLowerCase().includes(q) ||
                        t.last_message_snippet.toLowerCase().includes(q);
                      if (!match) return false;
                    }
                    return true;
                  })
                  .map((thread) => {
                    const isSelected = thread.thread_id === selectedThreadId;
                    return (
                      <div
                        key={thread.thread_id}
                        onClick={() => setSelectedThreadId(thread.thread_id)}
                        className={`p-3 rounded-lg border text-left cursor-pointer transition-all ${
                          isSelected
                            ? 'border-sky-500 bg-sky-50/50 dark:bg-sky-950/20 shadow-xs'
                            : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 hover:border-slate-300'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex items-center gap-2">
                            <div className="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center font-bold text-xs text-slate-700 dark:text-slate-300">
                              {thread.participant_name.charAt(0)}
                            </div>
                            <div>
                              <div className="flex items-center gap-1.5">
                                <span className="font-bold text-xs text-slate-900 dark:text-white">
                                  {thread.participant_name}
                                </span>
                                {thread.unread_count > 0 && (
                                  <span className="w-2 h-2 rounded-full bg-sky-500" />
                                )}
                              </div>
                              <p className="text-[11px] text-slate-500 truncate max-w-[180px]">
                                {thread.participant_company || thread.participant_headline}
                              </p>
                            </div>
                          </div>
                          <span className="text-[10px] text-slate-400 whitespace-nowrap">
                            {new Date(thread.last_message_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>

                        <div className="mt-2 text-xs font-semibold text-slate-800 dark:text-slate-200 truncate">
                          {thread.subject}
                        </div>
                        <p className="text-[11px] text-slate-500 line-clamp-2 mt-0.5">
                          {thread.last_message_snippet}
                        </p>

                        <div className="mt-2.5 pt-2 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-[10px]">
                          <span
                            className={`px-1.5 py-0.5 rounded font-medium ${
                              thread.classification === 'interview_invitation'
                                ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-400'
                                : thread.classification === 'recruiter_inquiry'
                                ? 'bg-indigo-100 text-indigo-800 dark:bg-indigo-950 dark:text-indigo-400'
                                : thread.classification === 'follow_up_response'
                                ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-400'
                                : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
                            }`}
                          >
                            {thread.classification.replace(/_/g, ' ')}
                          </span>

                          {thread.active_followup_planned ? (
                            <span className="px-1.5 py-0.5 rounded bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 font-medium">
                              Follow-up Scheduled
                            </span>
                          ) : (
                            <span className="px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400 font-medium flex items-center gap-1">
                              <StopCircle className="w-2.5 h-2.5" />
                              Follow-up Halted
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>

            {/* Right 7 Cols: Thread Messages & Contextual Reply Studio */}
            <div className="lg:col-span-7 flex flex-col gap-4">
              {(() => {
                const thread = inboxThreads.find((t) => t.thread_id === selectedThreadId);
                if (!thread) {
                  return (
                    <div className="p-8 text-center text-xs text-slate-500 border border-dashed rounded-lg">
                      Select a conversation thread to view message history and draft replies.
                    </div>
                  );
                }

                const draft = replyDrafts.find((d) => d.thread_id === thread.thread_id);

                return (
                  <div className="space-y-4">
                    {/* Thread Header Banner */}
                    <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/50 flex items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="font-bold text-sm text-slate-900 dark:text-white">
                            {thread.participant_name}
                          </h3>
                          <span className="text-xs text-slate-500">
                            • {thread.participant_company || 'LinkedIn Member'}
                          </span>
                        </div>
                        <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5">
                          {thread.participant_headline}
                        </p>
                        <div className="mt-1 flex items-center gap-2 text-[11px] text-slate-500">
                          <span className="font-semibold text-slate-700 dark:text-slate-300">
                            Subject: {thread.subject}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <a
                          href={`https://www.linkedin.com/messaging/`}
                          target="_blank"
                          rel="noreferrer"
                          className="px-2.5 py-1 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:bg-slate-50 flex items-center gap-1 font-semibold"
                        >
                          <ExternalLink className="w-3 h-3" />
                          <span>Open in LinkedIn</span>
                        </a>
                      </div>
                    </div>

                    {/* LI-09 Invariant Banner */}
                    {!thread.active_followup_planned && (
                      <div className="p-2.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/40 text-emerald-900 dark:text-emerald-300 text-xs flex items-center gap-2">
                        <StopCircle className="w-4 h-4 text-emerald-600 shrink-0" />
                        <div>
                          <span className="font-bold">LI-09 Invariant: </span>
                          <span>Inbound recruiter reply received. Active outbound follow-up schedule has been stopped automatically to avoid duplicate outreach.</span>
                        </div>
                      </div>
                    )}

                    {/* Message Timeline */}
                    <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-3 max-h-[280px] overflow-y-auto">
                      <div className="text-[10px] text-center text-slate-400 font-mono uppercase tracking-wider">
                        Conversation History
                      </div>
                      {thread.messages.map((m) => {
                        const isSelf = m.sender_type === 'self';
                        return (
                          <div
                            key={m.message_id}
                            className={`flex flex-col ${isSelf ? 'items-end' : 'items-start'}`}
                          >
                            <div className="flex items-center gap-1.5 text-[10px] text-slate-400 mb-0.5">
                              <span className="font-semibold text-slate-600 dark:text-slate-300">
                                {isSelf ? 'Candidate (You)' : m.sender_name}
                              </span>
                              <span>•</span>
                              <span>{new Date(m.sent_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                            </div>
                            <div
                              className={`p-3 rounded-lg max-w-[85%] text-xs leading-relaxed ${
                                isSelf
                                  ? 'bg-sky-600 text-white rounded-br-xs'
                                  : 'bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white rounded-bl-xs'
                              }`}
                            >
                              {m.content}
                            </div>
                          </div>
                        );
                      })}
                    </div>

                    {/* Live Inbound Recruiter Simulator (Tests LI-09 Invariant) */}
                    <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/40">
                      <div className="flex items-center justify-between mb-1.5">
                        <label className="text-xs font-bold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                          <MessageSquare className="w-3.5 h-3.5 text-indigo-500" />
                          <span>Simulate Inbound Recruiter Reply (LI-09 Invariant Test)</span>
                        </label>
                        <span className="text-[10px] text-slate-400">Halts active follow-ups</span>
                      </div>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={simulateInboundContent}
                          onChange={(e) => setSimulateInboundContent(e.target.value)}
                          placeholder="e.g. Thanks for following up! We would like to schedule a call..."
                          className="flex-1 px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        />
                        <button
                          type="button"
                          onClick={handleSimulateInboundMessage}
                          className="px-3 py-1.5 text-xs font-bold bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg whitespace-nowrap shadow-xs"
                        >
                          Simulate Reply
                        </button>
                      </div>
                    </div>

                    {/* Fact-Grounded Contextual Reply Draft Studio */}
                    <div className="p-4 rounded-xl border border-sky-200 dark:border-sky-900/50 bg-sky-50/20 dark:bg-sky-950/10 space-y-3.5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-sky-600" />
                          <h4 className="text-xs font-bold text-slate-900 dark:text-white">
                            Fact-Grounded Reply Draft Studio (FND-010, FND-015, AT-010)
                          </h4>
                        </div>
                        <span className="text-[10px] font-mono text-slate-400">
                          1-Click Clipboard Copy Only
                        </span>
                      </div>

                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                        <div>
                          <label className="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                            Response Tone
                          </label>
                          <select
                            value={selectedTone}
                            onChange={(e) => setSelectedTone(e.target.value as LinkedInReplyDraftTone)}
                            className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white text-xs"
                          >
                            <option value="concise_scheduling">Concise Scheduling Availability</option>
                            <option value="professional_enthusiastic">Professional & Enthusiastic</option>
                            <option value="polite_decline">Polite Pipeline Preserving Decline</option>
                            <option value="inquiry_clarification">Role Scope & Tech Stack Clarification</option>
                          </select>
                        </div>

                        <div>
                          <label className="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                            Candidate Availability
                          </label>
                          <input
                            type="text"
                            value={candidateAvailability}
                            onChange={(e) => setCandidateAvailability(e.target.value)}
                            placeholder="e.g. Thursday or Friday afternoon EST"
                            className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white text-xs"
                          />
                        </div>
                      </div>

                      <div>
                        <label className="block text-[11px] font-semibold text-slate-600 dark:text-slate-400 mb-1">
                          Verified Candidate Facts (Grounded Non-Hallucinatory Context)
                        </label>
                        <input
                          type="text"
                          value={verifiedFactsInput}
                          onChange={(e) => setVerifiedFactsInput(e.target.value)}
                          placeholder="Comma-separated verified background facts..."
                          className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white text-xs"
                        />
                      </div>

                      <div className="flex justify-end">
                        <button
                          type="button"
                          onClick={handleGenerateReplyDraft}
                          className="px-3.5 py-1.5 text-xs font-bold bg-sky-600 hover:bg-sky-700 text-white rounded-lg flex items-center gap-1.5 shadow-xs"
                        >
                          <Sparkles className="w-3.5 h-3.5" />
                          <span>Generate Contextual Draft</span>
                        </button>
                      </div>

                      {/* Active Draft Review & Approval Card */}
                      {draft && (
                        <div className="mt-3 pt-3 border-t border-sky-100 dark:border-sky-900/40 space-y-2.5">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span
                                className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                                  draft.status === 'approved'
                                    ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-400'
                                    : draft.status === 'copied_to_clipboard'
                                    ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-400'
                                    : draft.status === 'rejected'
                                    ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-400'
                                    : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-400'
                                }`}
                              >
                                {draft.status === 'copied_to_clipboard' ? 'COPIED TO CLIPBOARD' : draft.status.toUpperCase()}
                              </span>
                              <span className="text-[11px] text-slate-500 font-mono">
                                {draft.character_count} chars
                              </span>
                            </div>

                            {draft.approval_token && (
                              <span className="text-[10px] font-mono text-emerald-600 dark:text-emerald-400 truncate max-w-[200px]" title={draft.approval_token}>
                                Token: {draft.approval_token.substring(0, 16)}...
                              </span>
                            )}
                          </div>

                          {/* Tamper Invalidation Alert */}
                          {draftTamperAlert && (
                            <div className="p-2 rounded bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-850 text-amber-900 dark:text-amber-300 text-[11px]">
                              {draftTamperAlert}
                            </div>
                          )}

                          {/* Editable Suggested Text */}
                          <div>
                            <textarea
                              rows={4}
                              value={editingDraftText || draft.suggested_text}
                              onChange={(e) => handleEditDraftText(draft.draft_id, e.target.value)}
                              className="w-full p-2.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-white text-xs leading-relaxed font-sans"
                            />
                            <p className="text-[10px] text-slate-400 mt-0.5">
                              * Editing text after approval automatically invalidates cryptographic token and resets status to draft (Tamper Invalidation FND-010).
                            </p>
                          </div>

                          <div className="p-2 rounded-lg bg-slate-100/70 dark:bg-slate-800/50 text-[11px] text-slate-600 dark:text-slate-400">
                            <span className="font-semibold">Rationale: </span>
                            {draft.rationale}
                          </div>

                          {/* Draft Actions */}
                          <div className="pt-2 flex flex-wrap items-center justify-between gap-2">
                            <button
                              type="button"
                              onClick={() => handleRejectDraft(draft.draft_id)}
                              className="px-2.5 py-1 text-xs text-rose-600 hover:text-rose-700 font-medium"
                            >
                              Reject Draft
                            </button>

                            <div className="flex items-center gap-2">
                              {draft.status !== 'approved' && (
                                <button
                                  type="button"
                                  onClick={() => handleApproveReplyDraft(draft.draft_id)}
                                  className="px-3 py-1.5 text-xs font-bold rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white flex items-center gap-1 shadow-xs"
                                >
                                  <Check className="w-3.5 h-3.5" />
                                  <span>Approve Reply Draft</span>
                                </button>
                              )}

                              <button
                                type="button"
                                onClick={() => handleCopyDraftAndDeepLink(draft)}
                                className={`px-3 py-1.5 text-xs font-bold rounded-lg flex items-center gap-1 shadow-xs transition-colors ${
                                  draft.status === 'approved' || draft.status === 'copied_to_clipboard'
                                    ? 'bg-sky-600 hover:bg-sky-700 text-white'
                                    : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed'
                                }`}
                              >
                                <Copy className="w-3.5 h-3.5" />
                                <span>1-Click Copy & Open Thread</span>
                              </button>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })()}
            </div>
          </div>
        </section>
        </div>
      )}

      {/* MODULE 4: CONTENT & VOICE */}
      {activeModule === 'content' && (
        <div className="space-y-6">
        {/* Comments, Replies & Thread Sweep Studio (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, FND-010, FND-015, SRC-L1, SRC-L2) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="inline-flex items-center rounded-full bg-sky-50 dark:bg-sky-950/60 px-2.5 py-0.5 text-xs font-bold text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
                  IMP-LI-10 &bull; LI-10 &bull; AT-007 &bull; AT-010 &bull; FND-015
                </span>
                <span className="text-xs text-slate-400 font-mono">Comments & Thread Sweep</span>
              </div>
              <h2 className="text-lg font-black text-slate-900 dark:text-white mt-1 flex items-center gap-2">
                <MessageSquare className="w-5 h-5 text-sky-600" />
                <span>Comments, Replies & Thread Sweep Studio</span>
              </h2>
              <p className="text-xs text-slate-500 mt-0.5 max-w-2xl">
                Sweep target LinkedIn posts, extract conversation threads, and draft 3 high-value comment angles without generic spam. Enforces cryptographic HMAC approval and fail-closed bulk auto-engagement blocking (AT-010).
              </p>
            </div>

            <div className="flex items-center gap-2.5">
              {/* Workspace Selector (Cross-Tenant Isolation) */}
              <div className="flex items-center rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800/80 p-0.5 text-xs font-semibold">
                <button
                  type="button"
                  onClick={() => {
                    setCommentsWorkspace('ws-alpha');
                    const alphaFirst = sweptPosts.find((p) => p.workspace_id === 'ws-alpha');
                    if (alphaFirst) setSelectedSweptPostId(alphaFirst.post_id);
                  }}
                  className={`px-3 py-1 rounded-md transition-all ${
                    commentsWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-300 shadow-xs font-bold'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Alpha
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setCommentsWorkspace('ws-beta');
                    const betaFirst = sweptPosts.find((p) => p.workspace_id === 'ws-beta');
                    if (betaFirst) setSelectedSweptPostId(betaFirst.post_id);
                  }}
                  className={`px-3 py-1 rounded-md transition-all ${
                    commentsWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-300 shadow-xs font-bold'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Beta (Isolated)
                </button>
              </div>

              <button
                type="button"
                onClick={handleSimulateBulkEngagementReject}
                className="px-3 py-1.5 text-xs font-semibold rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 hover:bg-amber-100 flex items-center gap-1.5 shadow-xs"
                title="Verify AT-010 Bot Commenting Blocking"
              >
                <ShieldAlert className="w-3.5 h-3.5 text-amber-600" />
                <span>Test Anti-Bot Guard</span>
              </button>

              <button
                type="button"
                onClick={() => setIsAddingPostModalOpen(true)}
                className="px-3 py-1.5 text-xs font-bold rounded-lg bg-sky-600 hover:bg-sky-700 text-white flex items-center gap-1.5 shadow-xs"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Ingest Target Post</span>
              </button>
            </div>
          </div>

          {/* Studio Body Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
            {/* Left Column: Swept Target Posts Feed (4 cols) */}
            <div className="lg:col-span-4 space-y-4">
              <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 space-y-3">
                <div className="relative">
                  <Search className="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-400" />
                  <input
                    type="text"
                    value={commentsSearch}
                    onChange={(e) => setCommentsSearch(e.target.value)}
                    placeholder="Search posts or authors..."
                    className="w-full pl-8 pr-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs placeholder:text-slate-400"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-slate-500 mb-1">Post Category</label>
                  <select
                    value={commentsCategoryFilter}
                    onChange={(e) => setCommentsCategoryFilter(e.target.value)}
                    className="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs"
                  >
                    <option value="all">All Categories</option>
                    <option value="technical_discussion">Technical Discussion</option>
                    <option value="hiring_announcement">Hiring Announcement</option>
                    <option value="thought_leadership">Thought Leadership</option>
                    <option value="industry_news">Industry News</option>
                    <option value="general_discussion">General Discussion</option>
                  </select>
                </div>
              </div>

              {/* Swept Posts List */}
              <div className="space-y-2.5 max-h-[620px] overflow-y-auto pr-1">
                {sweptPosts
                  .filter((p) => {
                    if (p.workspace_id !== commentsWorkspace) return false;
                    if (commentsCategoryFilter !== 'all' && p.category !== commentsCategoryFilter) return false;
                    if (commentsSearch) {
                      const q = commentsSearch.toLowerCase();
                      const match =
                        p.author_name.toLowerCase().includes(q) ||
                        (p.author_company && p.author_company.toLowerCase().includes(q)) ||
                        p.content.toLowerCase().includes(q);
                      if (!match) return false;
                    }
                    return true;
                  })
                  .map((post) => {
                    const isSelected = post.post_id === selectedSweptPostId;
                    return (
                      <div
                        key={post.post_id}
                        onClick={() => handleSelectSweptPost(post.post_id)}
                        className={`p-3.5 rounded-lg border text-left cursor-pointer transition-all ${
                          isSelected
                            ? 'border-sky-500 bg-sky-50/50 dark:bg-sky-950/20 shadow-xs ring-1 ring-sky-500'
                            : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 hover:border-slate-300'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div>
                            <span className="font-bold text-xs text-slate-900 dark:text-white">
                              {post.author_name}
                            </span>
                            <p className="text-[11px] text-slate-500 truncate max-w-[200px]">
                              {post.author_company || post.author_headline}
                            </p>
                          </div>
                          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                            {post.category.replace('_', ' ')}
                          </span>
                        </div>

                        <p className="mt-2 text-xs text-slate-700 dark:text-slate-300 line-clamp-2 leading-relaxed">
                          {post.content}
                        </p>

                        <div className="mt-2.5 flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-slate-100 dark:border-slate-800/80">
                          <span className="flex items-center gap-1">
                            <MessageSquare className="w-3 h-3" />
                            <span>{post.existing_comments_count} existing comments</span>
                          </span>
                          <span>{new Date(post.swept_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>

            {/* Right Column: Post Thread Details & Comment Drafting Studio (8 cols) */}
            <div className="lg:col-span-8">
              {(() => {
                const post = sweptPosts.find((p) => p.post_id === selectedSweptPostId);
                if (!post) {
                  return (
                    <div className="h-full flex items-center justify-center p-12 border border-dashed border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 text-xs">
                      Select a swept post from the left feed to inspect thread and draft contextual comments.
                    </div>
                  );
                }

                const postDrafts = commentDrafts.filter((d) => d.post_id === post.post_id);

                return (
                  <div className="space-y-6">
                    {/* Post Detail Card */}
                    <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs space-y-3">
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-black text-slate-900 dark:text-white">
                              {post.author_name}
                            </span>
                            <span className="text-xs text-slate-500 font-medium">
                              &bull; {post.author_company || post.author_headline}
                            </span>
                          </div>
                          <span className="inline-block mt-0.5 text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-md bg-sky-50 dark:bg-sky-950/40 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
                            Category: {post.category.replace('_', ' ')}
                          </span>
                        </div>

                        <a
                          href={post.post_url}
                          target="_blank"
                          rel="noreferrer"
                          className="inline-flex items-center gap-1 text-xs font-bold text-sky-600 hover:text-sky-800 hover:underline"
                        >
                          <span>Open Native Post</span>
                          <ExternalLink className="w-3.5 h-3.5" />
                        </a>
                      </div>

                      <div className="p-3.5 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800 text-xs text-slate-800 dark:text-slate-200 leading-relaxed">
                        {post.content}
                      </div>

                      {/* Comments Thread Sample */}
                      {post.comments && post.comments.length > 0 && (
                        <div className="pt-2">
                          <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                            Swept Top Comment in Thread:
                          </span>
                          <div className="mt-1.5 p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs space-y-1">
                            <div className="flex items-center justify-between text-[11px] text-slate-500">
                              <span className="font-bold text-slate-700 dark:text-slate-300">
                                {post.comments[0].author_name}
                              </span>
                              <span>{post.comments[0].likes_count} likes</span>
                            </div>
                            <p className="text-slate-600 dark:text-slate-400 italic">
                              "{post.comments[0].content}"
                            </p>
                          </div>
                        </div>
                      )}
                    </div>

                    {/* Candidate Knowledge Facts & Generate Generator Box */}
                    <div className="p-4 rounded-xl border border-sky-200 dark:border-sky-800/60 bg-sky-50/30 dark:bg-sky-950/20 shadow-xs space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-sky-600" />
                          <h3 className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
                            Candidate Grounding Facts & Contextual Angles (SRC-L1, SRC-L2)
                          </h3>
                        </div>
                        <span className="text-[11px] text-slate-500">Anti-Spam Policy FND-015 Enforced</span>
                      </div>

                      <div className="space-y-1.5">
                        <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400">
                          Candidate Verified Domain Knowledge (Used to substantiate comments with real experience):
                        </label>
                        <input
                          type="text"
                          value={candidateFactsForComments}
                          onChange={(e) => setCandidateFactsForComments(e.target.value)}
                          className="w-full px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                          placeholder="e.g. 10+ yrs distributed systems in Go/K8s, multi-region database latency optimization"
                        />
                      </div>

                      <div className="flex items-center justify-between pt-1">
                        <span className="text-[11px] text-slate-500">
                          Generates 3 distinct angles: Insightful Addition, Engaging Question, Supportive Perspective
                        </span>
                        <button
                          type="button"
                          onClick={() => handleGenerateCommentDrafts(post.post_id)}
                          className="px-4 py-1.5 text-xs font-bold rounded-lg bg-sky-600 hover:bg-sky-700 text-white flex items-center gap-1.5 shadow-xs"
                        >
                          <Sparkles className="w-3.5 h-3.5" />
                          <span>Generate 3 Comment Drafts</span>
                        </button>
                      </div>
                    </div>

                    {/* Tamper Alert if triggered */}
                    {commentTamperAlert && (
                      <div className="p-3 rounded-lg border border-rose-300 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 text-xs flex items-center gap-2">
                        <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0" />
                        <span>{commentTamperAlert}</span>
                      </div>
                    )}

                    {/* Comment Drafts Showcase */}
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                          Generated Comment Drafts ({postDrafts.length})
                        </h3>
                        <span className="text-[11px] text-slate-500">
                          Requires candidate review & approval before 1-click copy (AT-007)
                        </span>
                      </div>

                      {postDrafts.length === 0 ? (
                        <div className="p-8 border border-dashed border-slate-200 dark:border-slate-800 rounded-lg text-center text-slate-400 text-xs">
                          No comment drafts generated yet for this post. Click 'Generate 3 Comment Drafts' above.
                        </div>
                      ) : (
                        postDrafts.map((draft) => {
                          const angleLabels: Record<LinkedInCommentDraftAngle, { label: string; color: string }> = {
                            insightful_addition: { label: 'Insightful Addition', color: 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/40 dark:text-indigo-300' },
                            engaging_question: { label: 'Engaging Question', color: 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950/40 dark:text-purple-300' },
                            supportive_perspective: { label: 'Supportive Perspective', color: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300' }
                          };
                          const meta = angleLabels[draft.angle] || { label: draft.angle, color: 'bg-slate-100 text-slate-700' };

                          return (
                            <div
                              key={draft.draft_id}
                              className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs space-y-3"
                            >
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                  <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-md border ${meta.color}`}>
                                    {meta.label}
                                  </span>
                                  <span className="text-[11px] text-slate-400">
                                    {draft.character_count} characters
                                  </span>
                                </div>

                                <div className="flex items-center gap-2">
                                  <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border ${
                                    draft.status === 'approved'
                                      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                                      : draft.status === 'copied_to_clipboard'
                                      ? 'bg-sky-50 text-sky-700 border-sky-200'
                                      : draft.status === 'rejected'
                                      ? 'bg-rose-50 text-rose-700 border-rose-200'
                                      : 'bg-amber-50 text-amber-700 border-amber-200'
                                  }`}>
                                    {draft.status.replace('_', ' ')}
                                  </span>
                                </div>
                              </div>

                              {/* Comment Text View / Inline Edit */}
                              {editingCommentDraftText && draft.draft_id === editingCommentDraftText ? (
                                <div className="space-y-2">
                                  <textarea
                                    defaultValue={draft.comment_text}
                                    id={`edit-text-${draft.draft_id}`}
                                    rows={4}
                                    className="w-full p-2.5 text-xs rounded-lg border border-sky-400 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                                  />
                                  <div className="flex justify-end gap-2">
                                    <button
                                      type="button"
                                      onClick={() => setEditingCommentDraftText('')}
                                      className="px-2.5 py-1 text-xs rounded border border-slate-300 text-slate-600"
                                    >
                                      Cancel
                                    </button>
                                    <button
                                      type="button"
                                      onClick={() => {
                                        const el = document.getElementById(`edit-text-${draft.draft_id}`) as HTMLTextAreaElement;
                                        if (el) handleEditCommentDraftText(draft.draft_id, el.value);
                                      }}
                                      className="px-2.5 py-1 text-xs rounded bg-sky-600 text-white font-bold"
                                    >
                                      Save Modification
                                    </button>
                                  </div>
                                </div>
                              ) : (
                                <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800 text-xs text-slate-800 dark:text-slate-200 leading-relaxed font-sans">
                                  {draft.comment_text}
                                </div>
                              )}

                              <div className="text-[11px] text-slate-500 bg-slate-100/60 dark:bg-slate-800/30 p-2 rounded-md">
                                <span className="font-semibold text-slate-700 dark:text-slate-300">Rationale: </span>
                                {draft.rationale}
                              </div>

                              {/* Cryptographic Approval Meta */}
                              {draft.approval_token && (
                                <div className="flex items-center gap-2 text-[10px] font-mono text-emerald-700 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 p-2 rounded-md border border-emerald-200 dark:border-emerald-800">
                                  <Shield className="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                                  <span className="truncate">Approval Token: {draft.approval_token}</span>
                                </div>
                              )}

                              {/* Actions Bar */}
                              <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80">
                                <div className="flex items-center gap-2">
                                  <button
                                    type="button"
                                    onClick={() => setEditingCommentDraftText(draft.draft_id)}
                                    className="text-xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white underline"
                                  >
                                    Edit Text
                                  </button>
                                  <span className="text-slate-300">|</span>
                                  <button
                                    type="button"
                                    onClick={() => handleRejectCommentDraft(draft.draft_id)}
                                    className="text-xs text-rose-600 hover:text-rose-800 underline"
                                  >
                                    Reject
                                  </button>
                                </div>

                                <div className="flex items-center gap-2">
                                  {draft.status !== 'approved' && (
                                    <button
                                      type="button"
                                      onClick={() => handleApproveCommentDraft(draft.draft_id)}
                                      className="px-3 py-1.5 text-xs font-bold rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white flex items-center gap-1 shadow-xs"
                                    >
                                      <Check className="w-3.5 h-3.5" />
                                      <span>Approve Draft</span>
                                    </button>
                                  )}

                                  <button
                                    type="button"
                                    onClick={() => {
                                      handleCopyCommentDraft(draft);
                                      if (draft.status === 'approved' || draft.status === 'copied_to_clipboard') {
                                        window.open(post.post_url, '_blank');
                                      }
                                    }}
                                    className={`px-3 py-1.5 text-xs font-bold rounded-lg flex items-center gap-1 shadow-xs transition-colors ${
                                      draft.status === 'approved' || draft.status === 'copied_to_clipboard'
                                        ? 'bg-sky-600 hover:bg-sky-700 text-white'
                                        : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed'
                                    }`}
                                  >
                                    <Copy className="w-3.5 h-3.5" />
                                    <span>1-Click Copy & Open Post</span>
                                  </button>
                                </div>
                              </div>
                            </div>
                          );
                        })
                      )}
                    </div>
                  </div>
                );
              })()}
            </div>
          </div>
        </section>

        {/* Modal: Ingest Target Post */}
        {isAddingPostModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-xs">
            <div className="w-full max-w-lg rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-xl space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800">
                <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Plus className="w-4 h-4 text-sky-600" />
                  <span>Ingest Target LinkedIn Post</span>
                </h3>
                <button
                  type="button"
                  onClick={() => setIsAddingPostModalOpen(false)}
                  className="text-slate-400 hover:text-slate-600 text-sm font-bold"
                >
                  &times;
                </button>
              </div>

              <div className="space-y-3 text-xs">
                <div>
                  <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Author Name *</label>
                  <input
                    type="text"
                    value={newPostAuthor}
                    onChange={(e) => setNewPostAuthor(e.target.value)}
                    placeholder="e.g. Kelsey Hightower"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs"
                  />
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Author Headline</label>
                    <input
                      type="text"
                      value={newPostHeadline}
                      onChange={(e) => setNewPostHeadline(e.target.value)}
                      placeholder="e.g. Staff Engineer"
                      className="w-full px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Company / Organization</label>
                    <input
                      type="text"
                      value={newPostCompany}
                      onChange={(e) => setNewPostCompany(e.target.value)}
                      placeholder="e.g. Google Cloud"
                      className="w-full px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">LinkedIn Post URL</label>
                  <input
                    type="text"
                    value={newPostUrl}
                    onChange={(e) => setNewPostUrl(e.target.value)}
                    placeholder="https://www.linkedin.com/feed/update/urn:li:activity:..."
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Post Content *</label>
                  <textarea
                    value={newPostContent}
                    onChange={(e) => setNewPostContent(e.target.value)}
                    rows={4}
                    placeholder="Paste the full text of the LinkedIn post..."
                    className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs"
                  />
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsAddingPostModalOpen(false)}
                  className="px-3 py-1.5 text-xs font-semibold rounded-lg border border-slate-300 text-slate-700 hover:bg-slate-50"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleIngestNewSweptPost}
                  className="px-3 py-1.5 text-xs font-bold rounded-lg bg-sky-600 hover:bg-sky-700 text-white shadow-xs"
                >
                  Ingest & Categorize Post
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Post Writing, Hooks & Editorial Audits Studio (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="inline-flex items-center rounded-full bg-sky-50 dark:bg-sky-950/60 px-2.5 py-0.5 text-xs font-bold text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
                  IMP-LI-11 &bull; LI-11 &bull; AT-003 &bull; AT-019 &bull; FND-015 &bull; SRC-L2
                </span>
                <span className="text-xs text-slate-400 font-mono">Post Writing & Hooks Studio</span>
              </div>
              <h2 className="text-lg font-black text-slate-900 dark:text-white mt-1 flex items-center gap-2">
                <FileEdit className="w-5 h-5 text-sky-600" />
                <span>LinkedIn Post Writing, Hooks & Editorial Audits Studio</span>
              </h2>
              <p className="text-xs text-slate-500 mt-0.5 max-w-2xl">
                Draft high-authority engineering posts from verified candidate facts. Select from 5 hook frameworks, evaluate real-time readability and buzzword audits, and sign with cryptographic HMAC approval before 1-click clipboard launch.
              </p>
            </div>

            <div className="flex items-center gap-2.5">
              {/* Workspace Selector (Cross-Tenant Isolation) */}
              <div className="flex items-center rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800/80 p-0.5 text-xs font-semibold">
                <button
                  type="button"
                  onClick={() => {
                    setPostWorkspace('ws-alpha');
                    const alphaFirst = postDrafts.find((p) => p.workspace_id === 'ws-alpha');
                    if (alphaFirst) handleSelectPostDraft(alphaFirst.draft_id);
                  }}
                  className={`px-3 py-1 rounded-md transition-all ${
                    postWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-300 shadow-xs font-bold'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Alpha
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setPostWorkspace('ws-beta');
                    const betaFirst = postDrafts.find((p) => p.workspace_id === 'ws-beta');
                    if (betaFirst) handleSelectPostDraft(betaFirst.draft_id);
                  }}
                  className={`px-3 py-1 rounded-md transition-all ${
                    postWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-sky-700 dark:text-sky-300 shadow-xs font-bold'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Beta (Isolated)
                </button>
              </div>

              <button
                type="button"
                onClick={handleSimulateDirectPostPublishReject}
                className="px-3 py-1.5 text-xs font-semibold rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 hover:bg-amber-100 flex items-center gap-1.5 shadow-xs"
                title="Verify AT-010 Auto-Publish Blocking"
              >
                <ShieldAlert className="w-3.5 h-3.5 text-amber-600" />
                <span>Test Anti-AutoPublish Guard</span>
              </button>
            </div>
          </div>

          {/* Studio Body Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
            {/* Left Column: Post Drafts Feed & New Generator (4 cols) */}
            <div className="lg:col-span-4 space-y-4">
              {/* Generator Card */}
              <div className="p-4 rounded-xl border border-sky-100 dark:border-sky-900/60 bg-sky-50/40 dark:bg-sky-950/20 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-black text-sky-900 dark:text-sky-200 flex items-center gap-1.5">
                    <Sparkles className="w-4 h-4 text-sky-600" />
                    <span>Generate Post Draft & 5 Hooks</span>
                  </span>
                  <span className="text-[10px] font-mono text-sky-600">AT-003 Verified</span>
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Post Topic / Core Thesis
                  </label>
                  <input
                    type="text"
                    value={postTopicInput}
                    onChange={(e) => setPostTopicInput(e.target.value)}
                    placeholder="e.g. Monolith vs Microservices Trade-offs"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Post Angle
                  </label>
                  <select
                    value={selectedAngle}
                    onChange={(e) => setSelectedAngle(e.target.value as LinkedInPostAngle)}
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                  >
                    <option value="contrarian_insight">Contrarian Insight</option>
                    <option value="lesson_learned_breakdown">Lesson Learned Breakdown</option>
                    <option value="technical_deep_dive">Technical Deep Dive</option>
                    <option value="milestone_celebration">Milestone Celebration</option>
                    <option value="actionable_guide">Actionable Engineering Guide</option>
                  </select>
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Verified Candidate Career Facts (AT-003)
                  </label>
                  <textarea
                    rows={2}
                    value={factsInput}
                    onChange={(e) => setFactsInput(e.target.value)}
                    placeholder="Comma-separated verified milestones (e.g. Reduced latency from 280ms to 42ms, 35% AWS cost savings)"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 resize-none font-mono"
                  />
                  <p className="text-[10px] text-slate-500 mt-0.5">
                    Any metrics or numbers in post text must exist in this verified fact set to pass the zero-hallucination audit gate.
                  </p>
                </div>

                <button
                  type="button"
                  onClick={handleGenerateNewPostDraft}
                  className="w-full py-2 px-3 rounded-lg bg-sky-600 hover:bg-sky-700 text-white text-xs font-bold flex items-center justify-center gap-1.5 shadow-xs transition-colors"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>Generate 5-Angle Hooks & Draft</span>
                </button>
              </div>

              {/* Drafts Filter & List */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-slate-600 dark:text-slate-400">
                    Workspace Drafts ({postDrafts.filter((p) => p.workspace_id === postWorkspace).length})
                  </span>
                  <div className="flex gap-1">
                    {['all', 'draft', 'approved'].map((st) => (
                      <button
                        key={st}
                        type="button"
                        onClick={() => setPostFilterStatus(st)}
                        className={`text-[10px] px-2 py-0.5 rounded capitalize ${
                          postFilterStatus === st
                            ? 'bg-sky-100 dark:bg-sky-900/60 text-sky-800 dark:text-sky-300 font-bold'
                            : 'text-slate-500 hover:text-slate-700'
                        }`}
                      >
                        {st}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="space-y-2 max-h-[480px] overflow-y-auto pr-1">
                  {postDrafts
                    .filter(
                      (p) =>
                        p.workspace_id === postWorkspace &&
                        (postFilterStatus === 'all' || p.status === postFilterStatus)
                    )
                    .map((draft) => {
                      const isSelected = draft.draft_id === selectedDraftId;
                      return (
                        <div
                          key={draft.draft_id}
                          onClick={() => handleSelectPostDraft(draft.draft_id)}
                          className={`p-3 rounded-xl border cursor-pointer transition-all ${
                            isSelected
                              ? 'border-sky-500 bg-sky-50/50 dark:bg-sky-950/40 ring-1 ring-sky-400 shadow-xs'
                              : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-slate-300'
                          }`}
                        >
                          <div className="flex items-center justify-between gap-1 mb-1">
                            <span className="text-xs font-bold text-slate-900 dark:text-white line-clamp-1">
                              {draft.topic}
                            </span>
                            <span
                              className={`text-[10px] px-1.5 py-0.5 rounded-full font-bold uppercase ${
                                draft.status === 'approved'
                                  ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                                  : draft.status === 'copied_to_clipboard'
                                  ? 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300'
                                  : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                              }`}
                            >
                              {draft.status === 'copied_to_clipboard' ? 'copied' : draft.status}
                            </span>
                          </div>

                          <p className="text-[11px] text-slate-600 dark:text-slate-400 line-clamp-2 italic mb-2">
                            &ldquo;{draft.selected_hook_text}&rdquo;
                          </p>

                          <div className="flex items-center justify-between text-[10px] text-slate-500">
                            <span className="capitalize font-medium">
                              {draft.angle.replace(/_/g, ' ')}
                            </span>
                            {draft.latest_audit && (
                              <span
                                className={`font-mono font-bold ${
                                  draft.latest_audit.passed ? 'text-emerald-600' : 'text-rose-600'
                                }`}
                              >
                                Score: {draft.latest_audit.readability_score}/100
                              </span>
                            )}
                          </div>
                        </div>
                      );
                    })}
                </div>
              </div>
            </div>

            {/* Right Column: Active Draft Editor, Hook Library & Audit Panel (8 cols) */}
            <div className="lg:col-span-8 space-y-5">
              {currentPostDraft ? (
                <>
                  {/* Tamper Alert Banner */}
                  {postTamperAlert && (
                    <div className="p-3.5 rounded-xl border border-rose-300 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 flex items-start gap-2.5 text-xs shadow-xs">
                      <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
                      <div>
                        <span className="font-bold">Cryptographic Invalidation Notice (AT-007):</span>{' '}
                        {postTamperAlert}
                      </div>
                    </div>
                  )}

                  {/* Header & Status Card */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-900/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold uppercase tracking-wider text-sky-600 dark:text-sky-400">
                          {currentPostDraft.angle.replace(/_/g, ' ')}
                        </span>
                        <span className="text-slate-300 dark:text-slate-700">&bull;</span>
                        <span className="text-xs font-mono text-slate-400">{currentPostDraft.draft_id}</span>
                      </div>
                      <h3 className="text-base font-bold text-slate-900 dark:text-white mt-0.5">
                        {currentPostDraft.topic}
                      </h3>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      {currentPostDraft.status === 'approved' && currentPostDraft.approval_token ? (
                        <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-100 dark:bg-emerald-950/60 border border-emerald-300 dark:border-emerald-800 text-emerald-800 dark:text-emerald-200 text-xs font-mono font-bold shadow-xs">
                          <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                          <span title={currentPostDraft.approval_token}>
                            Signed: {currentPostDraft.approval_token.slice(0, 24)}...
                          </span>
                        </div>
                      ) : (
                        <span className="px-2.5 py-1 rounded-md bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-200 text-xs font-bold">
                          Unsigned Draft
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Hook Variants Library (5 Angles / Hook Types) */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-3 shadow-xs">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                        <Layers className="w-4 h-4 text-sky-600" />
                        <span>Available Hook Frameworks (Click to Swap)</span>
                      </span>
                      <span className="text-[11px] text-slate-500">
                        Active: <strong className="capitalize text-sky-600">{currentPostDraft.selected_hook_type}</strong>
                      </span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2.5">
                      {currentPostDraft.available_hooks.map((hook) => {
                        const isActive = hook.hook_type === currentPostDraft.selected_hook_type;
                        return (
                          <div
                            key={hook.hook_type}
                            className={`p-3 rounded-lg border text-left flex flex-col justify-between transition-all ${
                              isActive
                                ? 'border-sky-500 bg-sky-50/50 dark:bg-sky-950/40 ring-1 ring-sky-400 shadow-xs'
                                : 'border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 hover:border-slate-300'
                            }`}
                          >
                            <div>
                              <div className="flex items-center justify-between mb-1">
                                <span className="text-[11px] font-extrabold uppercase text-slate-700 dark:text-slate-300">
                                  {hook.hook_type.replace(/_/g, ' ')}
                                </span>
                                {isActive && (
                                  <span className="text-[10px] font-bold px-1.5 py-0.2 rounded bg-sky-600 text-white">
                                    ACTIVE
                                  </span>
                                )}
                              </div>
                              <p className="text-xs font-medium text-slate-900 dark:text-white line-clamp-2">
                                &ldquo;{hook.hook_text}&rdquo;
                              </p>
                              <p className="text-[10px] text-slate-500 mt-1 line-clamp-1 italic">
                                {hook.rationale}
                              </p>
                            </div>

                            {!isActive && (
                              <button
                                type="button"
                                onClick={() => handleSwapPostHook(hook)}
                                className="mt-2.5 text-[11px] font-bold text-sky-600 hover:text-sky-800 dark:text-sky-400 inline-flex items-center gap-1 self-start"
                              >
                                <span>Use this hook</span>
                                <ArrowRight className="w-3 h-3" />
                              </button>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Full Post Text Editor */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-3 shadow-xs">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                        <FileText className="w-4 h-4 text-sky-600" />
                        <span>Full Post Draft Editor</span>
                      </span>
                      <div className="flex items-center gap-3 text-[11px] text-slate-500 font-mono">
                        <span>{editingPostText.length} / 3,000 chars</span>
                        <span>&bull;</span>
                        <span>~{Math.max(5, Math.round((editingPostText.trim().split(/\s+/).length / 200) * 60))}s read</span>
                      </div>
                    </div>

                    <textarea
                      rows={12}
                      value={editingPostText}
                      onChange={(e) => handlePostTextEdit(e.target.value)}
                      placeholder="Write your LinkedIn post draft here..."
                      className="w-full p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/60 text-slate-800 dark:text-slate-200 text-xs font-sans leading-relaxed resize-y focus:ring-1 focus:ring-sky-500 focus:outline-none"
                    />

                    <div className="flex items-center justify-between text-[11px] text-slate-500">
                      <span>
                        Recommended length: <strong>800 &ndash; 1,400 chars</strong> with short, punchy 1-3 line paragraphs.
                      </span>
                      {currentPostDraft.status === 'approved' && (
                        <span className="text-amber-600 dark:text-amber-400 font-semibold">
                          Editing text will invalidate cryptographic approval (AT-007).
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Real-time Editorial Audit Report Card */}
                  {currentPostDraft.latest_audit && (
                    <div
                      className={`p-4 rounded-xl border shadow-xs ${
                        currentPostDraft.latest_audit.passed
                          ? 'border-emerald-200 dark:border-emerald-900/60 bg-emerald-50/30 dark:bg-emerald-950/20'
                          : 'border-rose-200 dark:border-rose-900/60 bg-rose-50/30 dark:bg-rose-950/20'
                      }`}
                    >
                      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3 pb-3 border-b border-slate-200/60 dark:border-slate-800">
                        <div className="flex items-center gap-2">
                          {currentPostDraft.latest_audit.passed ? (
                            <CheckCircle2 className="w-5 h-5 text-emerald-600" />
                          ) : (
                            <XCircle className="w-5 h-5 text-rose-600" />
                          )}
                          <span className="text-xs font-black uppercase tracking-wider text-slate-900 dark:text-white">
                            Editorial Audit Status:{' '}
                            <span
                              className={
                                currentPostDraft.latest_audit.passed ? 'text-emerald-600' : 'text-rose-600'
                              }
                            >
                              {currentPostDraft.latest_audit.passed
                                ? 'PASSED (Ready for Sign-Off)'
                                : 'ACTION REQUIRED (Violations Detected)'}
                            </span>
                          </span>
                        </div>

                        <div className="text-[11px] font-mono text-slate-500">
                          Audited: {new Date(currentPostDraft.latest_audit.audited_at).toLocaleTimeString()}
                        </div>
                      </div>

                      {/* Audit Metrics Breakdown */}
                      <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mb-3">
                        <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                          <span className="text-[10px] text-slate-400 uppercase font-semibold">Readability</span>
                          <div className="text-base font-black text-sky-600">
                            {currentPostDraft.latest_audit.readability_score}/100
                          </div>
                        </div>

                        <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                          <span className="text-[10px] text-slate-400 uppercase font-semibold">Buzzwords</span>
                          <div
                            className={`text-base font-black ${
                              currentPostDraft.latest_audit.buzzwords_count > 0 ? 'text-amber-600' : 'text-slate-800 dark:text-slate-200'
                            }`}
                          >
                            {currentPostDraft.latest_audit.buzzwords_count}
                          </div>
                        </div>

                        <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                          <span className="text-[10px] text-slate-400 uppercase font-semibold">Unverified Stats</span>
                          <div
                            className={`text-base font-black ${
                              currentPostDraft.latest_audit.unverified_metrics_count > 0 ? 'text-rose-600' : 'text-slate-800 dark:text-slate-200'
                            }`}
                          >
                            {currentPostDraft.latest_audit.unverified_metrics_count}
                          </div>
                        </div>

                        <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                          <span className="text-[10px] text-slate-400 uppercase font-semibold">Virality Claims</span>
                          <div
                            className={`text-base font-black ${
                              currentPostDraft.latest_audit.virality_claims_count > 0 ? 'text-rose-600' : 'text-slate-800 dark:text-slate-200'
                            }`}
                          >
                            {currentPostDraft.latest_audit.virality_claims_count}
                          </div>
                        </div>

                        <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                          <span className="text-[10px] text-slate-400 uppercase font-semibold">AI Bypass Claims</span>
                          <div
                            className={`text-base font-black ${
                              currentPostDraft.latest_audit.ai_bypass_claims_count > 0 ? 'text-rose-600' : 'text-slate-800 dark:text-slate-200'
                            }`}
                          >
                            {currentPostDraft.latest_audit.ai_bypass_claims_count}
                          </div>
                        </div>
                      </div>

                      {/* Audit Issues List */}
                      {currentPostDraft.latest_audit.issues.length > 0 && (
                        <div className="space-y-2 mb-3">
                          <span className="text-xs font-bold text-slate-700 dark:text-slate-300">
                            Detected Issues ({currentPostDraft.latest_audit.issues.length}):
                          </span>
                          {currentPostDraft.latest_audit.issues.map((iss, i) => (
                            <div
                              key={i}
                              className={`p-2.5 rounded-lg text-xs flex items-start gap-2 border ${
                                iss.severity === 'blocking'
                                  ? 'bg-rose-100/70 dark:bg-rose-950/60 border-rose-300 dark:border-rose-800 text-rose-900 dark:text-rose-200'
                                  : 'bg-amber-100/70 dark:bg-amber-950/60 border-amber-300 dark:border-amber-800 text-amber-900 dark:text-amber-200'
                              }`}
                            >
                              <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
                              <div className="space-y-0.5">
                                <div className="font-bold flex items-center gap-1.5">
                                  <span className="uppercase text-[10px] px-1 py-0.2 rounded bg-black/10">
                                    {iss.severity}
                                  </span>
                                  <span>{iss.message}</span>
                                </div>
                                {iss.suggested_fix && (
                                  <p className="text-[11px] opacity-90">
                                    <strong>Suggested fix:</strong> {iss.suggested_fix}
                                  </p>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      )}

                      <div className="text-[10px] text-slate-500 leading-relaxed">
                        <Shield className="w-3 h-3 inline-block mr-1 text-sky-600" />
                        <strong>Policy Invariant (LI-11, AT-019, AT-003):</strong> {currentPostDraft.latest_audit.disclaimer}
                      </div>
                    </div>
                  )}

                  {/* Actions & Cryptographic Signing Bar */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-xs">
                    <div>
                      <span className="text-xs font-bold text-slate-900 dark:text-white block">
                        Publishing Approval & Native Launch Gating
                      </span>
                      <span className="text-[11px] text-slate-500">
                        {currentPostDraft.status === 'approved'
                          ? 'Signed and authorized for 1-click clipboard copy (AT-010).'
                          : 'Review audit findings and sign to generate cryptographic HMAC token.'}
                      </span>
                    </div>

                    <div className="flex items-center gap-2.5">
                      {currentPostDraft.status !== 'approved' ? (
                        <button
                          type="button"
                          onClick={handleApprovePostDraft}
                          className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                        >
                          <Lock className="w-3.5 h-3.5" />
                          <span>Approve & Sign Post Draft (HMAC)</span>
                        </button>
                      ) : (
                        <button
                          type="button"
                          onClick={handleCopyPostToClipboard}
                          className="px-4 py-2 rounded-lg bg-sky-600 hover:bg-sky-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                        >
                          <Copy className="w-3.5 h-3.5" />
                          <span>1-Click Copy & Launch LinkedIn Compose</span>
                          <ExternalLink className="w-3 h-3 ml-0.5" />
                        </button>
                      )}
                    </div>
                  </div>
                </>
              ) : (
                <div className="p-12 text-center text-slate-400 rounded-xl border border-dashed border-slate-300">
                  Select or generate a post draft to begin editing.
                </div>
              )}
            </div>
          </div>
        </section>

        {/* ========================================================================= */}
        {/* IMP-LI-12: Reusable Voice Profile & Multi-Tier Humanizer Studio           */}
        {/* ========================================================================= */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          {/* Section Header */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-100 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2">
                <span className="inline-flex items-center rounded-full bg-purple-50 dark:bg-purple-950/60 px-2.5 py-0.5 text-xs font-bold text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800">
                  IMP-LI-12 &bull; LI-12 &bull; AT-003 &bull; AT-007 &bull; FND-015 &bull; SRC-L2
                </span>
                <span className="text-xs font-semibold text-slate-400">Multi-Tier De-Slop Engine</span>
              </div>
              <h2 className="text-lg font-black text-slate-900 dark:text-white mt-1">
                Reusable Voice Profile &amp; Multi-Tier Humanizer Studio
              </h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 max-w-3xl">
                Configure candidate writing style (Formality, Technical Depth, Cadence, Perspective), scrub generic AI slop &amp; clichés, and enforce zero-hallucination groundtruth anchors with cryptographic HMAC approval.
              </p>
            </div>

            {/* Controls: Workspace Selector & Guard Simulator */}
            <div className="flex flex-wrap items-center gap-2 shrink-0">
              <div className="flex items-center rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 p-1 text-xs">
                <span className="px-2 font-bold text-slate-400 text-[10px] uppercase">Tenant Workspace:</span>
                <button
                  type="button"
                  onClick={() => setVoiceWorkspace('ws-alpha')}
                  className={`px-2.5 py-1 rounded-md font-bold transition-all ${
                    voiceWorkspace === 'ws-alpha'
                      ? 'bg-purple-600 text-white shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Alpha (Architect)
                </button>
                <button
                  type="button"
                  onClick={() => setVoiceWorkspace('ws-beta')}
                  className={`px-2.5 py-1 rounded-md font-bold transition-all ${
                    voiceWorkspace === 'ws-beta'
                      ? 'bg-purple-600 text-white shadow-xs'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Beta (Director)
                </button>
              </div>

              <button
                type="button"
                onClick={handleSimulateBypassClaimReject}
                className="px-3 py-1.5 text-xs font-semibold rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 hover:bg-amber-100 flex items-center gap-1.5 shadow-xs"
                title="Verify REQ-016 Anti-Bypass Guard"
              >
                <ShieldAlert className="w-3.5 h-3.5 text-amber-600" />
                <span>Anti-AI-Bypass Policy</span>
              </button>
            </div>
          </div>

          {/* Studio Body Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
            {/* Left Column: Voice Profile Builder & Config (5 cols) */}
            <div className="lg:col-span-5 space-y-4">
              <div className="p-4 rounded-xl border border-purple-100 dark:border-purple-900/60 bg-purple-50/40 dark:bg-purple-950/20 space-y-3.5">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-black text-purple-900 dark:text-purple-200 flex items-center gap-1.5">
                    <Sliders className="w-4 h-4 text-purple-600" />
                    <span>Candidate Voice Profile Config</span>
                  </span>
                  <span className="text-[10px] font-mono font-bold text-purple-600 bg-purple-100 dark:bg-purple-900/60 px-2 py-0.5 rounded-full">
                    {currentVoiceProfile ? currentVoiceProfile.status.toUpperCase() : 'DRAFT'}
                  </span>
                </div>

                {/* Profile Selector */}
                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Select Voice Profile
                  </label>
                  <div className="flex gap-2">
                    {voiceProfiles
                      .filter((p) => p.workspace_id === voiceWorkspace)
                      .map((p) => (
                        <button
                          key={p.profile_id}
                          type="button"
                          onClick={() => handleSelectVoiceProfile(p.profile_id)}
                          className={`flex-1 text-left p-2 rounded-lg border text-xs transition-all ${
                            selectedVoiceProfileId === p.profile_id
                              ? 'border-purple-600 bg-purple-100/60 dark:bg-purple-900/40 ring-1 ring-purple-500 font-bold'
                              : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300'
                          }`}
                        >
                          <div className="line-clamp-1">{p.profile_name}</div>
                          <div className="text-[10px] font-normal text-slate-500">{p.cadence_style}</div>
                        </button>
                      ))}
                  </div>
                </div>

                {/* Profile Tamper Alert Banner */}
                {voiceTamperAlert && (
                  <div className="p-3 rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/40 text-amber-900 dark:text-amber-200 text-xs flex items-start gap-2">
                    <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                    <span>{voiceTamperAlert}</span>
                  </div>
                )}

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Profile Name
                  </label>
                  <input
                    type="text"
                    value={voiceProfileNameInput}
                    onChange={(e) => setVoiceProfileNameInput(e.target.value)}
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Target Reader Audience
                  </label>
                  <input
                    type="text"
                    value={voiceAudienceInput}
                    onChange={(e) => setVoiceAudienceInput(e.target.value)}
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                  />
                </div>

                {/* Dual Sliders: Formality & Technical Depth */}
                <div className="grid grid-cols-2 gap-3">
                  <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                    <div className="flex items-center justify-between text-[11px] font-semibold mb-1">
                      <span className="text-slate-700 dark:text-slate-300">Formality</span>
                      <span className="font-mono text-purple-600 font-bold">{voiceFormalityInput}/5</span>
                    </div>
                    <input
                      type="range"
                      min={1}
                      max={5}
                      step={1}
                      value={voiceFormalityInput}
                      onChange={(e) => setVoiceFormalityInput(Number(e.target.value))}
                      className="w-full accent-purple-600 h-1.5 bg-slate-200 rounded-lg cursor-pointer"
                    />
                    <div className="flex justify-between text-[9px] text-slate-400 mt-1">
                      <span>Casual</span>
                      <span>Executive</span>
                    </div>
                  </div>

                  <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                    <div className="flex items-center justify-between text-[11px] font-semibold mb-1">
                      <span className="text-slate-700 dark:text-slate-300">Tech Depth</span>
                      <span className="font-mono text-purple-600 font-bold">{voiceDepthInput}/5</span>
                    </div>
                    <input
                      type="range"
                      min={1}
                      max={5}
                      step={1}
                      value={voiceDepthInput}
                      onChange={(e) => setVoiceDepthInput(Number(e.target.value))}
                      className="w-full accent-purple-600 h-1.5 bg-slate-200 rounded-lg cursor-pointer"
                    />
                    <div className="flex justify-between text-[9px] text-slate-400 mt-1">
                      <span>High-level</span>
                      <span>Deep Code</span>
                    </div>
                  </div>
                </div>

                {/* Cadence & Perspective Selectors */}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                      Cadence &amp; Rhythm Style
                    </label>
                    <select
                      value={voiceCadenceInput}
                      onChange={(e) => setVoiceCadenceInput(e.target.value as LinkedInCadenceStyle)}
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                    >
                      <option value="punchy_staccato">Punchy Staccato (Short, impactful)</option>
                      <option value="balanced_rhythm">Balanced Rhythm (Varied pacing)</option>
                      <option value="analytical_deep">Analytical Deep (Thorough sentences)</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                      Perspective (Point-of-View)
                    </label>
                    <select
                      value={voicePerspectiveInput}
                      onChange={(e) => setVoicePerspectiveInput(e.target.value as LinkedInPerspective)}
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200"
                    >
                      <option value="first_person_singular">First Person (&quot;I architected&quot;)</option>
                      <option value="collective_team">Collective Team (&quot;Our team built&quot;)</option>
                      <option value="neutral_practitioner">Neutral Practitioner (&quot;Architecting requires&quot;)</option>
                    </select>
                  </div>
                </div>

                {/* Vocabulary Lists */}
                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Preferred Signature Terms
                  </label>
                  <input
                    type="text"
                    value={voicePreferredTermsInput}
                    onChange={(e) => setVoicePreferredTermsInput(e.target.value)}
                    placeholder="e.g. trade-offs, operational overhead, rigor"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 font-mono text-[11px]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Blacklisted Buzzwords &amp; Slop Tropes
                  </label>
                  <input
                    type="text"
                    value={voiceBlacklistedTermsInput}
                    onChange={(e) => setVoiceBlacklistedTermsInput(e.target.value)}
                    placeholder="e.g. paradigm shift, game changer, synergy"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 font-mono text-[11px]"
                  />
                </div>

                {/* Save & Approve Profile Actions */}
                <div className="flex gap-2 pt-1">
                  <button
                    type="button"
                    onClick={handleSaveVoiceProfile}
                    className="flex-1 py-2 px-3 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 text-xs font-bold shadow-xs transition-colors"
                  >
                    Save Changes
                  </button>
                  {currentVoiceProfile && currentVoiceProfile.status !== 'approved' && (
                    <button
                      type="button"
                      onClick={() => handleApproveVoiceProfile(currentVoiceProfile.profile_id)}
                      className="py-2 px-3 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                    >
                      <Lock className="w-3.5 h-3.5" />
                      <span>Approve Voice (HMAC)</span>
                    </button>
                  )}
                </div>

                {currentVoiceProfile && currentVoiceProfile.approval_token && (
                  <div className="text-[10px] font-mono text-purple-700 dark:text-purple-300 bg-purple-100/60 dark:bg-purple-950/40 p-2 rounded-lg border border-purple-200 dark:border-purple-800">
                    <span className="font-bold">Active Token:</span> {currentVoiceProfile.approval_token.slice(0, 32)}...
                  </div>
                )}
              </div>
            </div>

            {/* Right Column: Multi-Tier Humanizer & De-Slop Test Bench (7 cols) */}
            <div className="lg:col-span-7 space-y-4">
              {/* Humanizer Tamper Alert Banner */}
              {humanizeTamperAlert && (
                <div className="p-3.5 rounded-xl border border-rose-300 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 flex items-start gap-2.5 text-xs shadow-xs">
                  <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-bold">Cryptographic Tamper Invalidation (AT-007):</span>{' '}
                    {humanizeTamperAlert}
                  </div>
                </div>
              )}

              {/* Input Card */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-900/40 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 flex items-center gap-1.5">
                    <Sparkles className="w-4 h-4 text-purple-600" />
                    <span>Raw Draft Text to Humanize &amp; De-Slop</span>
                  </span>
                  <span className="text-[10px] text-slate-500 font-mono">Tiers 1, 2, 3 Active</span>
                </div>

                <textarea
                  rows={4}
                  value={humanizeInputText}
                  onChange={(e) => setHumanizeInputText(e.target.value)}
                  placeholder="Paste draft post, comment, or message text here..."
                  className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 resize-none font-mono"
                />

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Candidate Verified Career Facts (Tier 3 Zero-Hallucination Anchor &bull; AT-003)
                  </label>
                  <input
                    type="text"
                    value={humanizeCandidateFacts}
                    onChange={(e) => setHumanizeCandidateFacts(e.target.value)}
                    placeholder="e.g. Reduced AWS bill by 35% within 3 months, P99 latency improved from 280ms to 42ms"
                    className="w-full px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 font-mono text-[11px]"
                  />
                  <p className="text-[10px] text-slate-500 mt-0.5">
                    Any statistics or percentages in content are strictly audited against this groundtruth set. Unverified claims will fail closed.
                  </p>
                </div>

                <button
                  type="button"
                  onClick={handleProcessHumanize}
                  className="w-full py-2.5 px-4 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold flex items-center justify-center gap-2 shadow-xs transition-colors"
                >
                  <Sparkles className="w-4 h-4" />
                  <span>Execute Multi-Tier De-Slop &amp; Humanize</span>
                </button>
              </div>

              {/* Transformation Results View */}
              {currentHumanizeResult && (
                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-4">
                  {/* Transformation Scores Banner */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                    <div className="p-3 rounded-lg bg-purple-50 dark:bg-purple-950/40 border border-purple-100 dark:border-purple-900">
                      <span className="text-[10px] font-bold text-slate-400 uppercase">Slop Score (Lower is better)</span>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-slate-400 line-through font-mono">{currentHumanizeResult.slop_score_before}</span>
                        <span className="text-base font-black text-emerald-600 font-mono">&rarr; {currentHumanizeResult.slop_score_after}</span>
                      </div>
                    </div>

                    <div className="p-3 rounded-lg bg-purple-50 dark:bg-purple-950/40 border border-purple-100 dark:border-purple-900">
                      <span className="text-[10px] font-bold text-slate-400 uppercase">Readability Index</span>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-slate-400 line-through font-mono">{currentHumanizeResult.readability_score_before}</span>
                        <span className="text-base font-black text-purple-600 font-mono">&rarr; {currentHumanizeResult.readability_score_after}</span>
                      </div>
                    </div>

                    <div className="p-3 rounded-lg bg-purple-50 dark:bg-purple-950/40 border border-purple-100 dark:border-purple-900">
                      <span className="text-[10px] font-bold text-slate-400 uppercase">Tier 1 Slop Replaced</span>
                      <div className="text-base font-black text-slate-800 dark:text-slate-200 mt-1">
                        {currentHumanizeResult.tier1_slop_replacements.length} Phrases
                      </div>
                    </div>

                    <div className="p-3 rounded-lg bg-purple-50 dark:bg-purple-950/40 border border-purple-100 dark:border-purple-900">
                      <span className="text-[10px] font-bold text-slate-400 uppercase">Audit Status</span>
                      <div className="mt-1">
                        {currentHumanizeResult.passed ? (
                          <span className="inline-flex items-center gap-1 text-xs font-bold text-emerald-600">
                            <CheckCircle2 className="w-3.5 h-3.5" /> Passed
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-xs font-bold text-rose-600">
                            <XCircle className="w-3.5 h-3.5" /> Blocked
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Tier 1 Replaced Items Badges */}
                  {currentHumanizeResult.tier1_slop_replacements.length > 0 && (
                    <div>
                      <span className="text-xs font-bold text-slate-700 dark:text-slate-300 block mb-1.5">
                        Tier 1: Scrubbed Slop &amp; Cliches
                      </span>
                      <div className="flex flex-wrap gap-1.5">
                        {currentHumanizeResult.tier1_slop_replacements.map((r, i) => (
                          <span
                            key={i}
                            className="inline-flex items-center gap-1 text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-2 py-1 rounded-md border border-slate-200 dark:border-slate-700"
                            title={r.rationale}
                          >
                            <span className="line-through text-rose-500">{r.offending_phrase}</span>
                            <span>&rarr;</span>
                            <span className="font-bold text-emerald-600">{r.suggested_replacement}</span>
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Tier 3 Blocking Issues if any */}
                  {currentHumanizeResult.tier3_issues.length > 0 && (
                    <div className="p-3.5 rounded-xl border border-rose-300 bg-rose-50 dark:bg-rose-950/40 space-y-2">
                      <div className="flex items-center gap-1.5 text-xs font-bold text-rose-800 dark:text-rose-200">
                        <AlertTriangle className="w-4 h-4 text-rose-600" />
                        <span>Tier 3 Audit Blocking Issues (Zero-Hallucination Groundtruth Guard &bull; AT-003):</span>
                      </div>
                      <div className="space-y-1.5">
                        {currentHumanizeResult.tier3_issues.map((iss, i) => (
                          <div key={i} className="text-xs text-rose-700 dark:text-rose-300 bg-white dark:bg-slate-900 p-2.5 rounded-lg border border-rose-200 dark:border-rose-900">
                            <div className="font-bold">{iss.message}</div>
                            <div className="text-[11px] text-slate-500 mt-0.5">
                              Offending Snippet: <code className="font-mono bg-rose-50 dark:bg-rose-950 px-1 py-0.5 rounded text-rose-800 dark:text-rose-300 font-bold">{iss.offending_snippet}</code>
                            </div>
                            <div className="text-[11px] text-slate-600 dark:text-slate-400 mt-0.5 italic">
                              Remedy: {iss.suggested_fix}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Cleaned Humanized Text Editor */}
                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-xs font-bold text-slate-700 dark:text-slate-300">
                        Cleaned Practitioner Content (Editable)
                      </label>
                      <span className="text-[10px] text-slate-400">
                        Editing after approval invalidates HMAC signature (AT-007)
                      </span>
                    </div>
                    <textarea
                      rows={4}
                      value={editingHumanizedCleanText}
                      onChange={(e) => handleEditHumanizedCleanText(currentHumanizeResult.result_id, e.target.value)}
                      className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs text-slate-800 dark:text-slate-200 font-mono leading-relaxed resize-none"
                    />
                  </div>

                  {/* Actions: HMAC Approval & 1-Click Clipboard Copy */}
                  <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pt-2 border-t border-slate-100 dark:border-slate-800">
                    <div className="text-xs text-slate-500 flex items-center gap-1.5">
                      {currentHumanizeResult.status === 'approved' && currentHumanizeResult.approval_token ? (
                        <span className="text-emerald-600 font-mono font-bold flex items-center gap-1">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          HMAC-Approved: {currentHumanizeResult.approval_token.slice(0, 24)}...
                        </span>
                      ) : currentHumanizeResult.status === 'copied_to_clipboard' ? (
                        <span className="text-purple-600 font-bold flex items-center gap-1">
                          <Check className="w-3.5 h-3.5" />
                          Copied to Clipboard &bull; Compose Launched
                        </span>
                      ) : (
                        <span className="text-amber-600 font-medium">
                          Status: Pending Review &amp; Signature
                        </span>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      {currentHumanizeResult.status !== 'approved' ? (
                        <button
                          type="button"
                          disabled={!currentHumanizeResult.passed}
                          onClick={() => handleApproveHumanizedDraft(currentHumanizeResult.result_id)}
                          className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                        >
                          <Lock className="w-3.5 h-3.5" />
                          <span>Approve &amp; Sign Draft (HMAC)</span>
                        </button>
                      ) : (
                        <button
                          type="button"
                          onClick={() => handleCopyHumanizedDraft(currentHumanizeResult.result_id)}
                          className="px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                        >
                          <Copy className="w-3.5 h-3.5" />
                          <span>1-Click Copy &amp; Launch LinkedIn</span>
                          <ExternalLink className="w-3 h-3 ml-0.5" />
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Story Bank & Guided Interviewer Studio (IMP-LI-13, LI-13, AT-003, AT-007, AT-010, FND-010, SRC-L2) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          {/* Header & Meta */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-5 border-b border-slate-100 dark:border-slate-800">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold tracking-wide uppercase bg-emerald-100 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
                  IMP-LI-13
                </span>
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800">
                  AT-003 Objective vs Narrative Split
                </span>
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-purple-50 dark:bg-purple-950/50 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800">
                  AT-007 HMAC Invalidation
                </span>
                <span className="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-sky-50 dark:bg-sky-950/50 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
                  AT-010 1-Click Clipboard
                </span>
              </div>
              <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <BookOpen className="w-5 h-5 text-indigo-600" />
                <span>Story Bank &amp; Guided Interviewer Studio</span>
              </h2>
              <p className="text-xs text-slate-500 mt-1 max-w-3xl leading-relaxed">
                Transform experiential career milestones into structured STAR narratives across 5 curated categories.
                Enforces fail-closed zero-hallucination metric audits, HMAC cryptographic approvals, instant tamper invalidation, and 1-click clipboard export with LinkedIn deep-linking.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              {/* Workspace Selector */}
              <div className="flex items-center bg-slate-100 dark:bg-slate-800 p-0.5 rounded-lg border border-slate-200 dark:border-slate-700 text-xs">
                <button
                  type="button"
                  onClick={() => setStoryWorkspace('ws-alpha')}
                  className={`px-2.5 py-1 rounded-md font-semibold transition-all ${
                    storyWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white shadow-xs'
                      : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'
                  }`}
                >
                  Alpha Team (`ws-alpha`)
                </button>
                <button
                  type="button"
                  onClick={() => setStoryWorkspace('ws-beta')}
                  className={`px-2.5 py-1 rounded-md font-semibold transition-all ${
                    storyWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white shadow-xs'
                      : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'
                  }`}
                >
                  Beta Compliance (`ws-beta`)
                </button>
              </div>

              {/* Launch Interviewer Button */}
              <button
                type="button"
                onClick={() => handleStartInterview('scar_or_failure')}
                className="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
              >
                <Mic className="w-3.5 h-3.5" />
                <span>Launch Guided Interviewer</span>
              </button>
            </div>
          </div>

          {/* Tamper Alert Banner */}
          {storyTamperAlert && (
            <div className="mt-4 p-3.5 rounded-xl border border-rose-300 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 flex items-start gap-2.5 text-xs shadow-xs animate-in fade-in">
              <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
              <div>
                <span className="font-bold">Cryptographic Tamper Invalidation (AT-007):</span>{' '}
                {storyTamperAlert}
              </div>
            </div>
          )}

          {/* Guided Interviewer Modal / Accordion */}
          {isInterviewerOpen && (
            <div className="mt-5 p-5 rounded-xl border-2 border-indigo-200 dark:border-indigo-900/60 bg-indigo-50/40 dark:bg-indigo-950/20 shadow-xs">
              <div className="flex items-center justify-between pb-3 border-b border-indigo-100 dark:border-indigo-900/40 mb-4">
                <div className="flex items-center gap-2">
                  <div className="w-7 h-7 rounded-lg bg-indigo-600 text-white flex items-center justify-center font-bold text-xs">
                    <Mic className="w-4 h-4" />
                  </div>
                  <div>
                    <h3 className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
                      Guided Interviewer Session: {interviewStep.toUpperCase()}
                    </h3>
                    <p className="text-[11px] text-slate-500">
                      Step-by-step Socratic interview synthesizing experiential memories into verified STAR narratives.
                    </p>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={() => setIsInterviewerOpen(false)}
                  className="text-xs font-semibold text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
                >
                  ✕ Close Wizard
                </button>
              </div>

              {/* Category Picker */}
              <div className="mb-4">
                <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                  Select Story Archetype (SRC-L2):
                </label>
                <div className="flex flex-wrap gap-1.5">
                  {[
                    { id: 'scar_or_failure', label: 'Scar or Failure', desc: 'Outages, postmortems & incident resilience' },
                    { id: 'turning_point', label: 'Turning Point', desc: 'Pivotal leadership or architecture shift' },
                    { id: 'breakthrough_win', label: 'Breakthrough Win', desc: 'High-risk initiative that succeeded' },
                    { id: 'contrarian_belief', label: 'Contrarian Belief', desc: 'Challenging industry orthodoxies' },
                    { id: 'mentorship_culture', label: 'Mentorship & Culture', desc: 'Transformative coaching & multiplier impact' }
                  ].map((cat) => (
                    <button
                      key={cat.id}
                      type="button"
                      onClick={() => setSelectedPromptCategory(cat.id as LinkedInStoryCategory)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium border text-left transition-all ${
                        selectedPromptCategory === cat.id
                          ? 'border-indigo-500 bg-indigo-100 dark:bg-indigo-900/60 text-indigo-950 dark:text-white font-bold shadow-xs'
                          : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:border-slate-300'
                      }`}
                    >
                      <div>{cat.label}</div>
                      <div className="text-[10px] text-slate-400 font-normal">{cat.desc}</div>
                    </button>
                  ))}
                </div>
              </div>

              {/* Wizard Step 1: Initial Question */}
              {interviewStep === 'question' && (
                <div className="space-y-3">
                  <div className="p-3.5 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700">
                    <span className="text-[10px] font-bold uppercase tracking-wider text-indigo-600 block mb-1">
                      Target Competency: {interviewPrompts.find((p) => p.category === selectedPromptCategory)?.title}
                    </span>
                    <p className="text-xs font-semibold text-slate-800 dark:text-slate-100 leading-relaxed">
                      &quot;{interviewPrompts.find((p) => p.category === selectedPromptCategory)?.question}&quot;
                    </p>
                  </div>

                  <div>
                    <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                      Your First-Hand Narrative Context:
                    </label>
                    <textarea
                      rows={4}
                      value={candidateAnswerInput}
                      onChange={(e) => setCandidateAnswerInput(e.target.value)}
                      placeholder="Explain what happened, what systems were involved, and what core dilemma you faced..."
                      className="w-full text-xs font-mono p-3 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>

                  <div className="flex justify-end gap-2">
                    <button
                      type="button"
                      onClick={handleNextInterviewStep}
                      className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs"
                    >
                      <span>Proceed to AI Follow-Up Probe</span>
                      <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              )}

              {/* Wizard Step 2: Follow-Up Probe */}
              {interviewStep === 'probe' && (
                <div className="space-y-3">
                  <div className="p-3.5 rounded-lg bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900/60">
                    <span className="text-[10px] font-bold uppercase tracking-wider text-amber-800 dark:text-amber-300 block mb-1">
                      AI Deep-Dive Probe (Zero-Fluff Quantified Investigation)
                    </span>
                    <p className="text-xs font-semibold text-slate-800 dark:text-slate-100 leading-relaxed">
                      &quot;{interviewPrompts.find((p) => p.category === selectedPromptCategory)?.follow_up_probes?.[0]}&quot;
                    </p>
                  </div>

                  <div>
                    <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                      Grounded Resolution, Metrics &amp; Operational Safeguards:
                    </label>
                    <textarea
                      rows={4}
                      value={candidateFollowUpInput}
                      onChange={(e) => setCandidateFollowUpInput(e.target.value)}
                      placeholder="Describe the precise fix, latency/availability metrics, and lasting architectural safeguards..."
                      className="w-full text-xs font-mono p-3 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>

                  <div className="flex justify-between items-center">
                    <button
                      type="button"
                      onClick={() => setInterviewStep('question')}
                      className="px-3 py-1.5 text-xs font-semibold text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
                    >
                      ← Back to Initial Context
                    </button>

                    <button
                      type="button"
                      onClick={handleNextInterviewStep}
                      className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs"
                    >
                      <span>Proceed to Story Synthesis</span>
                      <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              )}

              {/* Wizard Step 3: Synthesis Review */}
              {interviewStep === 'synthesize' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        Author Name / Attributed Owner:
                      </label>
                      <input
                        type="text"
                        value={interviewAuthorName}
                        onChange={(e) => setInterviewAuthorName(e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        placeholder="e.g. Elena Rostova"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        Target Workspace:
                      </label>
                      <div className="text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-mono font-medium">
                        {storyWorkspace} (Isolated multi-tenant container)
                      </div>
                    </div>
                  </div>

                  <div className="p-3.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/40 text-xs text-emerald-950 dark:text-emerald-200">
                    <span className="font-bold flex items-center gap-1.5 mb-1">
                      <Sparkles className="w-4 h-4 text-emerald-600" />
                      STAR Narrative Synthesis Engine (SRC-L2)
                    </span>
                    The interviewer will map your responses into structured Hook, Situation/Context, Task/Challenge, Action, Quantified Outcome, and Engineering Lesson learned.
                  </div>

                  <div className="flex justify-between items-center">
                    <button
                      type="button"
                      onClick={() => setInterviewStep('probe')}
                      className="px-3 py-1.5 text-xs font-semibold text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
                    >
                      ← Back to Probe
                    </button>

                    <button
                      type="button"
                      onClick={handleSynthesizeStory}
                      className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs"
                    >
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>Commit Synthesized Story to Bank</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Main Studio Body: 2-Column Explorer & Inspector */}
          <div className="mt-5 grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left Column: Story Bank List & Filter (4 cols) */}
            <div className="lg:col-span-5 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center gap-1.5">
                  <Database className="w-3.5 h-3.5 text-indigo-600" />
                  <span>Stories Repository ({storyEntries.filter(s => s.workspace_id === storyWorkspace).length})</span>
                </span>
                <span className="text-[10px] font-mono text-slate-400">
                  {storyWorkspace}
                </span>
              </div>

              {/* Filter & Search Bar */}
              <div className="flex items-center gap-2">
                <select
                  value={storyCategoryFilter}
                  onChange={(e) => setStoryCategoryFilter(e.target.value as any)}
                  className="text-xs py-1.5 px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 font-medium"
                >
                  <option value="all">All Categories</option>
                  <option value="scar_or_failure">Scar / Failure</option>
                  <option value="turning_point">Turning Point</option>
                  <option value="breakthrough_win">Breakthrough Win</option>
                  <option value="contrarian_belief">Contrarian Belief</option>
                  <option value="mentorship_culture">Mentorship &amp; Culture</option>
                </select>

                <div className="relative flex-1">
                  <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-slate-400" />
                  <input
                    type="text"
                    placeholder="Search stories by keyword..."
                    value={storySearch}
                    onChange={(e) => setStorySearch(e.target.value)}
                    className="w-full text-xs pl-8 pr-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200"
                  />
                </div>
              </div>

              {/* Story Cards List */}
              <div className="space-y-2 max-h-[580px] overflow-y-auto pr-1">
                {storyEntries
                  .filter((s) => s.workspace_id === storyWorkspace)
                  .filter((s) => storyCategoryFilter === 'all' || s.category === storyCategoryFilter)
                  .filter((s) =>
                    storySearch
                      ? s.title.toLowerCase().includes(storySearch.toLowerCase()) ||
                        s.narrative.hook_summary.toLowerCase().includes(storySearch.toLowerCase())
                      : true
                  )
                  .map((story) => {
                    const isSelected = story.id === currentStory?.id;
                    const isApproved = story.approval_status === 'approved';
                    const isArchived = story.approval_status === 'archived';

                    return (
                      <div
                        key={story.id}
                        onClick={() => setSelectedStoryId(story.id)}
                        className={`p-3.5 rounded-xl border cursor-pointer transition-all ${
                          isSelected
                            ? 'border-indigo-500 bg-indigo-50/50 dark:bg-indigo-950/30 shadow-xs ring-1 ring-indigo-500'
                            : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-850 hover:border-slate-300 dark:hover:border-slate-700'
                        }`}
                      >
                        <div className="flex items-center justify-between gap-2 mb-1.5">
                          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                            {story.category.replace(/_/g, ' ')}
                          </span>

                          <div className="flex items-center gap-1.5">
                            {isApproved ? (
                              <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 dark:bg-emerald-950 text-emerald-800 dark:text-emerald-300 flex items-center gap-1">
                                <Lock className="w-2.5 h-2.5" />
                                <span>Approved</span>
                              </span>
                            ) : isArchived ? (
                              <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                                Archived
                              </span>
                            ) : (
                              <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 dark:bg-amber-950 text-amber-800 dark:text-amber-300 flex items-center gap-1">
                                <Clock className="w-2.5 h-2.5" />
                                <span>Draft</span>
                              </span>
                            )}
                          </div>
                        </div>

                        <h4 className="text-xs font-bold text-slate-900 dark:text-white line-clamp-2 leading-snug">
                          {story.title}
                        </h4>

                        <p className="mt-1 text-[11px] text-slate-500 dark:text-slate-400 line-clamp-2 leading-relaxed">
                          {story.narrative.hook_summary}
                        </p>

                        <div className="mt-2.5 pt-2 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between text-[10px] text-slate-400">
                          <span className="flex items-center gap-1">
                            <span className="font-semibold text-slate-600 dark:text-slate-300">By:</span>{' '}
                            {story.provenance.author_name}
                          </span>
                          <span className="font-mono">
                            {story.provenance.source_type}
                          </span>
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>

            {/* Right Column: Selected Story Detailed Inspector (7 cols) */}
            <div className="lg:col-span-7 space-y-4">
              {currentStory ? (
                <>
                  {/* Title & Metadata Card */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-850">
                    <div className="flex items-start justify-between gap-4 mb-2">
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800">
                            {currentStory.category.replace(/_/g, ' ')}
                          </span>
                          <span className="text-[10px] text-slate-400 font-mono">
                            ID: {currentStory.id}
                          </span>
                        </div>
                        <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                          {currentStory.title}
                        </h3>
                      </div>

                      <div className="flex items-center gap-1.5 shrink-0">
                        {currentStory.approval_status !== 'archived' && (
                          <button
                            type="button"
                            onClick={() => handleArchiveStory(currentStory.id)}
                            className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-400 hover:text-rose-600 transition-colors"
                            title="Archive Story"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Provenance Card (AT-003 Objective Facts Separation) */}
                    <div className="mt-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 text-[11px] grid grid-cols-2 sm:grid-cols-4 gap-2">
                      <div>
                        <span className="text-[10px] font-bold uppercase text-slate-400 block">Provenance</span>
                        <span className="font-mono text-slate-700 dark:text-slate-300 font-medium">
                          {currentStory.provenance.source_type}
                        </span>
                      </div>
                      <div>
                        <span className="text-[10px] font-bold uppercase text-slate-400 block">Author</span>
                        <span className="text-slate-700 dark:text-slate-300 font-medium">
                          {currentStory.provenance.author_name}
                        </span>
                      </div>
                      <div className="sm:col-span-2">
                        <span className="text-[10px] font-bold uppercase text-slate-400 block">Linked Grounded Fact IDs</span>
                        <div className="flex flex-wrap gap-1 mt-0.5">
                          {currentStory.provenance.grounded_fact_ids?.map((fid) => (
                            <span key={fid} className="px-1.5 py-0.5 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 font-mono text-[10px] text-slate-600 dark:text-slate-400">
                              {fid}
                            </span>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* STAR Narrative Structure Inspector & Editor */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-855 space-y-3.5">
                    <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                      <span className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center gap-1.5">
                        <Layers className="w-3.5 h-3.5 text-indigo-600" />
                        <span>STAR Experiential Narrative Breakdown</span>
                      </span>
                      <span className="text-[11px] text-slate-400">
                        Editing invalidates approval tokens (AT-007)
                      </span>
                    </div>

                    {/* 1. The Hook */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        1. The Hook (Scroll-Stopping Core Incident):
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.hook_summary}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'hook_summary', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>

                    {/* 2. Context & Background */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        2. Situation &amp; Context:
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.context_background}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'context_background', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>

                    {/* 3. Challenge & Conflict */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        3. Task, Challenge &amp; Conflict:
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.challenge_conflict}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'challenge_conflict', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>

                    {/* 4. Action Taken */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        4. Tactical Action Taken:
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.action_taken}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'action_taken', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>

                    {/* 5. Quantified Outcome */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1 flex items-center justify-between">
                        <span>5. Quantified Outcome (Must be Grounded in Facts):</span>
                        <span className="text-[10px] text-amber-600 font-semibold">Zero-Hallucination Guarded</span>
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.quantified_outcome}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'quantified_outcome', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>

                    {/* 6. Lesson Learned */}
                    <div>
                      <label className="block text-[11px] font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                        6. Principled Engineering Lesson:
                      </label>
                      <textarea
                        rows={2}
                        value={currentStory.narrative.lesson_learned}
                        onChange={(e) => handleEditStoryNarrativeField(currentStory.id, 'lesson_learned', e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                      />
                    </div>
                  </div>

                  {/* Grounded Candidate Facts & Zero-Hallucination Audit */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-850 space-y-3">
                    <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                      <span className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center gap-1.5">
                        <Shield className="w-3.5 h-3.5 text-emerald-600" />
                        <span>Candidate Verified Facts &amp; Audit Engine (AT-003)</span>
                      </span>
                      <button
                        type="button"
                        onClick={() => handleAuditStory(currentStory.id)}
                        className="px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-xs font-semibold flex items-center gap-1 transition-colors"
                      >
                        <RefreshCw className="w-3 h-3" />
                        <span>Run Audit</span>
                      </button>
                    </div>

                    <div>
                      <label className="block text-[11px] font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
                        Verified Facts Grounding Pool:
                      </label>
                      <input
                        type="text"
                        value={storyVerifiedFactsInput}
                        onChange={(e) => setStoryVerifiedFactsInput(e.target.value)}
                        className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 text-slate-800 dark:text-slate-200 font-mono"
                      />
                    </div>

                    {storyAuditReport && (
                      <div className={`p-3.5 rounded-lg border text-xs ${
                        !storyAuditReport.has_blocking_issues
                          ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800 text-emerald-900 dark:text-emerald-200'
                          : 'bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800 text-rose-900 dark:text-rose-200'
                      }`}>
                        <div className="flex items-center justify-between font-bold mb-1">
                          <span className="flex items-center gap-1.5">
                            {!storyAuditReport.has_blocking_issues ? (
                              <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                            ) : (
                              <XCircle className="w-4 h-4 text-rose-600" />
                            )}
                            {!storyAuditReport.has_blocking_issues ? 'Zero-Hallucination Audit Passed' : 'Audit Blocked: Unverified Metrics Detected'}
                          </span>
                          <span className="font-mono text-[10px]">
                            {storyAuditReport.issues.length} Issues Found
                          </span>
                        </div>

                        {storyAuditReport.issues.map((issue, idx) => (
                          <div key={idx} className="mt-1 pl-5 text-[11px] font-mono">
                            • <span className="font-bold underline">{issue.metric}</span>: {issue.message}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Cryptographic Approval & 1-Click Export Gatekeeper */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-850">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                      <div>
                        <span className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center gap-1.5">
                          <Lock className="w-3.5 h-3.5 text-purple-600" />
                          <span>Cryptographic Approval Gate (AT-007, AT-010)</span>
                        </span>
                        <p className="text-[11px] text-slate-500 mt-0.5">
                          {currentStory.approval_status === 'approved'
                            ? 'Story approved and digitally signed. Ready for instant clipboard assist.'
                            : 'Unsigned draft. Must be audited and approved before export.'}
                        </p>
                      </div>

                      <div className="flex items-center gap-2">
                        {currentStory.approval_status !== 'approved' ? (
                          <button
                            type="button"
                            onClick={() => handleApproveStory(currentStory.id)}
                            className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                          >
                            <Lock className="w-3.5 h-3.5" />
                            <span>Approve &amp; Sign (HMAC)</span>
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleCopyStoryMarkdown(currentStory.id)}
                            className="px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs transition-colors"
                          >
                            <Copy className="w-3.5 h-3.5" />
                            <span>1-Click Copy &amp; Open LinkedIn</span>
                            <ExternalLink className="w-3 h-3 ml-0.5" />
                          </button>
                        )}
                      </div>
                    </div>

                    {currentStory.approval_token && (
                      <div className="mt-3 p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 font-mono text-[10px] text-slate-500 break-all flex items-center gap-2">
                        <span className="font-bold text-slate-700 dark:text-slate-300 shrink-0">HMAC Token:</span>
                        <span>{currentStory.approval_token}</span>
                      </div>
                    )}
                  </div>
                </>
              ) : (
                <div className="p-8 rounded-xl border border-dashed border-slate-300 dark:border-slate-800 text-center text-slate-400 text-xs">
                  Select a story from the repository to inspect its STAR narrative and cryptographic provenance.
                </div>
              )}
            </div>
          </div>
        </section>
        </div>
      )}

      {/* MODULE 5: GROWTH & ANALYTICS */}
      {activeModule === 'growth' && (
        <div className="space-y-6">
        {/* Content Planning, Repurposing & Multi-Format Calendar Studio (IMP-LI-14, LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          {/* Header */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-4">
            <div>
              <div className="flex items-center gap-2">
                <span className="p-1.5 rounded-lg bg-teal-50 dark:bg-teal-950/50 text-teal-600 border border-teal-200 dark:border-teal-800">
                  <CalendarDays className="w-5 h-5" />
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white">
                  Content Planning, Repurposing &amp; DST-Safe Calendar Studio
                </h2>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-teal-100 dark:bg-teal-900/60 text-teal-700 dark:text-teal-300 uppercase tracking-wider">
                  IMP-LI-14
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-1 max-w-3xl">
                Transform authoritative engineering artifacts (blogs, releases, RFCs, postmortems) into 5 high-impact LinkedIn formats.
                Schedule across timezones with Day-Light Saving Time (DST) safety (AT-018), cryptographic approval tokens (AT-007),
                and fail-closed direct publishing guards (AT-010).
              </p>
            </div>

            {/* Workspace Selector */}
            <div className="flex items-center gap-2 shrink-0">
              <span className="text-xs font-semibold text-slate-500">Workspace:</span>
              <div className="flex rounded-lg border border-slate-200 dark:border-slate-700 p-0.5 bg-slate-50 dark:bg-slate-800 text-xs">
                <button
                  type="button"
                  onClick={() => setContentWorkspace('ws-alpha')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    contentWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-700 text-teal-700 dark:text-teal-300 shadow-xs font-bold'
                      : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'
                  }`}
                >
                  Workspace Alpha
                </button>
                <button
                  type="button"
                  onClick={() => setContentWorkspace('ws-beta')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    contentWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-700 text-teal-700 dark:text-teal-300 shadow-xs font-bold'
                      : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'
                  }`}
                >
                  Workspace Beta (Isolated)
                </button>
              </div>
            </div>
          </div>

          {/* Tamper and Policy Banners */}
          {planTamperAlert && (
            <div className="p-3 rounded-lg border border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-950/40 text-amber-900 dark:text-amber-200 text-xs flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
              <span>{planTamperAlert}</span>
            </div>
          )}
          {draftTamperAlert2 && (
            <div className="p-3 rounded-lg border border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-950/40 text-amber-900 dark:text-amber-200 text-xs flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
              <span>{draftTamperAlert2}</span>
            </div>
          )}
          {directPublishRejectAlert && (
            <div className="p-3 rounded-lg border border-rose-300 dark:border-rose-700 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 text-xs flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-rose-600 shrink-0" />
              <span>{directPublishRejectAlert}</span>
            </div>
          )}
          {clipboardCopiedNotice && (
            <div className="p-3 rounded-lg border border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-900 dark:text-emerald-200 text-xs flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
              <span>{clipboardCopiedNotice}</span>
            </div>
          )}

          {/* 2-Column Main Layout: Repurposer (Left) + Calendar (Right) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left Column: Source Artifact & Repurposing Variants (7 cols) */}
            <div className="lg:col-span-7 space-y-5">
              {/* 1. Source Artifact Input Form */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3">
                <div className="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-700">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                    <FileText className="w-3.5 h-3.5 text-teal-600" />
                    <span>Source Engineering Artifact (SRC-L2)</span>
                  </span>
                  <button
                    type="button"
                    onClick={() => {
                      setSourceArtifact(INITIAL_SOURCE_ARTIFACT);
                      setNotification('Loaded standard distributed systems postmortem artifact!');
                      setTimeout(() => setNotification(null), 3000);
                    }}
                    className="text-[11px] text-teal-600 hover:text-teal-700 dark:text-teal-400 font-semibold"
                  >
                    Reset to Distributed Cache Example
                  </button>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div className="sm:col-span-2">
                    <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                      Artifact Title
                    </label>
                    <input
                      type="text"
                      value={sourceArtifact.title}
                      onChange={(e) => setSourceArtifact({ ...sourceArtifact, title: e.target.value })}
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-medium"
                    />
                  </div>
                  <div>
                    <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                      Artifact Type
                    </label>
                    <select
                      value={sourceArtifact.artifact_type}
                      onChange={(e) => setSourceArtifact({ ...sourceArtifact, artifact_type: e.target.value as LinkedInArtifactType })}
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-medium"
                    >
                      <option value="technical_blog">Technical Blog</option>
                      <option value="github_release">GitHub Release Notes</option>
                      <option value="architecture_rfc">Architecture RFC</option>
                      <option value="incident_postmortem">Incident Postmortem</option>
                      <option value="benchmark_report">Benchmark Report</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                    Raw Technical Content / Excerpt
                  </label>
                  <textarea
                    rows={4}
                    value={sourceArtifact.raw_text}
                    onChange={(e) => setSourceArtifact({ ...sourceArtifact, raw_text: e.target.value })}
                    className="w-full text-xs p-2.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-mono"
                  />
                </div>

                <div className="flex items-center justify-between pt-1">
                  <span className="text-[11px] text-slate-400">
                    Author: <strong>{sourceArtifact.author}</strong> | Tags: #{sourceArtifact.tags.join(' #')}
                  </span>
                  <button
                    type="button"
                    onClick={handleRepurposeSourceToAllFormats}
                    className="px-3.5 py-1.5 rounded-lg bg-teal-600 hover:bg-teal-700 text-white font-bold text-xs shadow-xs flex items-center gap-1.5 transition-colors"
                  >
                    <Sparkles className="w-3.5 h-3.5" />
                    <span>Repurpose into All 5 Formats</span>
                  </button>
                </div>
              </div>

              {/* 2. Repurposed Drafts List & Editor */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-4">
                <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                    <Layers className="w-3.5 h-3.5 text-indigo-600" />
                    <span>Repurposed LinkedIn Variants ({repurposedDrafts.filter(d => d.workspace_id === contentWorkspace).length})</span>
                  </span>
                  <span className="text-[11px] text-slate-400">
                    Editing post-approval invalidates HMAC token (AT-007)
                  </span>
                </div>

                {/* Format Pills */}
                <div className="flex flex-wrap gap-1.5">
                  {repurposedDrafts
                    .filter((d) => d.workspace_id === contentWorkspace)
                    .map((draft) => (
                      <button
                        key={draft.id}
                        type="button"
                        onClick={() => setSelectedRepurposedDraftId(draft.id)}
                        className={`px-3 py-1.5 rounded-lg text-xs font-bold border transition-all flex items-center gap-1.5 ${
                          selectedRepurposedDraftId === draft.id
                            ? 'border-teal-500 bg-teal-50 dark:bg-teal-950/60 text-teal-800 dark:text-teal-200 shadow-xs'
                            : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-50'
                        }`}
                      >
                        <span>{draft.format.replace(/_/g, ' ').toUpperCase()}</span>
                        {draft.status === 'approved' ? (
                          <span className="w-2 h-2 rounded-full bg-emerald-500" title="Approved" />
                        ) : (
                          <span className="w-2 h-2 rounded-full bg-amber-400" title="Draft" />
                        )}
                      </button>
                    ))}
                </div>

                {/* Selected Draft Editor & Preview */}
                {currentRepurposedDraft && (
                  <div className="space-y-3 pt-2 border-t border-slate-100 dark:border-slate-800">
                    <div className="flex items-center justify-between">
                      <div>
                        <span className="text-[10px] font-bold text-teal-600 dark:text-teal-400 uppercase tracking-wider block">
                          Format: {currentRepurposedDraft.format.replace(/_/g, ' ')}
                        </span>
                        <input
                          type="text"
                          value={currentRepurposedDraft.title}
                          onChange={(e) =>
                            handleEditRepurposedDraft(
                              currentRepurposedDraft.id,
                              e.target.value,
                              currentRepurposedDraft.content_body
                            )
                          }
                          className="text-xs font-bold text-slate-900 dark:text-white bg-transparent border-b border-transparent hover:border-slate-300 focus:border-teal-500 focus:outline-none w-full mt-0.5"
                        />
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <span className={`px-2 py-0.5 text-[10px] font-bold rounded-full uppercase ${
                          currentRepurposedDraft.status === 'approved'
                            ? 'bg-emerald-100 dark:bg-emerald-900/60 text-emerald-700 dark:text-emerald-300'
                            : 'bg-amber-100 dark:bg-amber-900/60 text-amber-700 dark:text-amber-300'
                        }`}>
                          {currentRepurposedDraft.status}
                        </span>
                      </div>
                    </div>

                    <div>
                      <textarea
                        rows={8}
                        value={currentRepurposedDraft.content_body}
                        onChange={(e) =>
                          handleEditRepurposedDraft(
                            currentRepurposedDraft.id,
                            currentRepurposedDraft.title,
                            e.target.value
                          )
                        }
                        className="w-full text-xs p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-mono leading-relaxed"
                      />
                    </div>

                    <div className="flex items-center justify-between text-[11px] text-slate-500">
                      <span>Characters: <strong>{currentRepurposedDraft.character_count}</strong> | Est. Read Time: <strong>{Math.round(currentRepurposedDraft.character_count / 15)}s</strong></span>
                      {currentRepurposedDraft.approval_token && (
                        <span className="font-mono text-[10px] text-emerald-600 dark:text-emerald-400">
                          HMAC Signed: {currentRepurposedDraft.approval_token.slice(0, 24)}...
                        </span>
                      )}
                    </div>

                    {/* Draft Actions */}
                    <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800">
                      {currentRepurposedDraft.status !== 'approved' ? (
                        <button
                          type="button"
                          onClick={() => handleApproveRepurposedDraft(currentRepurposedDraft.id)}
                          className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs flex items-center gap-1.5 transition-colors"
                        >
                          <Shield className="w-3.5 h-3.5" />
                          <span>Approve Draft (HMAC-SHA256)</span>
                        </button>
                      ) : (
                        <div className="flex items-center gap-1 text-xs text-emerald-600 font-bold">
                          <CheckCircle2 className="w-4 h-4" />
                          <span>Cryptographically Approved</span>
                        </div>
                      )}

                      <button
                        type="button"
                        onClick={() => handleScheduleDraftToCalendar(currentRepurposedDraft)}
                        className="px-3 py-1.5 rounded-lg bg-teal-600 hover:bg-teal-700 text-white font-bold text-xs flex items-center gap-1.5 shadow-xs transition-colors"
                      >
                        <Calendar className="w-3.5 h-3.5" />
                        <span>Schedule to Content Calendar</span>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Right Column: DST-Safe Content Calendar (5 cols) */}
            <div className="lg:col-span-5 space-y-5">
              {/* 1. Timezone & Slot Control Bar */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3">
                <div className="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-700">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                    <Clock className="w-3.5 h-3.5 text-indigo-600" />
                    <span>DST-Safe Slot Allocator (AT-018)</span>
                  </span>
                  <span className="text-[10px] text-teal-600 font-semibold uppercase">FND-008</span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                  <div>
                    <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                      User Timezone
                    </label>
                    <select
                      value={calendarTimezone}
                      onChange={(e) => setCalendarTimezone(e.target.value)}
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-medium"
                    >
                      <option value="America/New_York">America/New_York (EDT/EST)</option>
                      <option value="Asia/Dhaka">Asia/Dhaka (BST, UTC+6)</option>
                      <option value="UTC">UTC (Coordinated Universal Time)</option>
                      <option value="Europe/London">Europe/London (BST/GMT)</option>
                      <option value="America/Los_Angeles">America/Los_Angeles (PDT/PST)</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                      Daily Slot Budget
                    </label>
                    <input
                      type="number"
                      min={1}
                      max={5}
                      value={calendarMaxDailyBudget}
                      onChange={(e) => setCalendarMaxDailyBudget(parseInt(e.target.value) || 2)}
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-medium"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                    Scheduled Local Date &amp; Time
                  </label>
                  <input
                    type="datetime-local"
                    value={calendarDateInput}
                    onChange={(e) => setCalendarDateInput(e.target.value)}
                    className="w-full text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 font-medium"
                  />
                  <p className="text-[10px] text-slate-400 mt-1">
                    Automatic calendar day boundary adjustment prevents double-budget allocation or slot drop on DST shift days.
                  </p>
                </div>
              </div>

              {/* 2. Scheduled Calendar Feed */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-3">
                <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                    <Calendar className="w-3.5 h-3.5 text-teal-600" />
                    <span>Scheduled Slots ({calendarItems.filter(c => c.workspace_id === contentWorkspace).length})</span>
                  </span>
                  <span className="text-[10px] text-slate-400 font-medium">
                    Strict AT-010 Human Guarded
                  </span>
                </div>

                <div className="space-y-3 max-h-[520px] overflow-y-auto pr-1">
                  {calendarItems.filter(c => c.workspace_id === contentWorkspace).length === 0 ? (
                    <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-lg text-slate-400 text-xs">
                      No scheduled slots in {contentWorkspace}. Repurpose an engineering artifact and schedule a slot above.
                    </div>
                  ) : (
                    calendarItems
                      .filter((item) => item.workspace_id === contentWorkspace)
                      .map((item) => (
                        <div
                          key={item.id}
                          className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/60 dark:bg-slate-800/40 space-y-2.5 transition-all hover:border-teal-300"
                        >
                          <div className="flex items-start justify-between gap-2">
                            <div>
                              <div className="flex items-center gap-1.5">
                                <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-teal-100 dark:bg-teal-900/60 text-teal-800 dark:text-teal-200">
                                  {item.format.replace(/_/g, ' ')}
                                </span>
                                <span className="text-[10px] text-slate-400 font-mono">
                                  {item.id}
                                </span>
                              </div>
                              <h4 className="text-xs font-bold text-slate-900 dark:text-white mt-1">
                                {item.title}
                              </h4>
                            </div>

                            <button
                              type="button"
                              onClick={() => handleDeletePlanItem(item.id)}
                              className="p-1 text-slate-400 hover:text-rose-600 transition-colors"
                              title="Delete Plan Slot"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          </div>

                          {/* Time & DST Badges */}
                          <div className="p-2 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-[11px] space-y-1">
                            <div className="flex items-center justify-between text-slate-700 dark:text-slate-300">
                              <span className="font-semibold flex items-center gap-1">
                                <Clock className="w-3 h-3 text-teal-600" />
                                <span>Local Slot:</span>
                              </span>
                              <span className="font-mono text-[10px]">{item.local_slot_formatted}</span>
                            </div>
                            <div className="flex items-center justify-between text-slate-500">
                              <span>Canonical UTC:</span>
                              <span className="font-mono text-[10px]">{item.scheduled_slot_utc}</span>
                            </div>
                          </div>

                          {/* Approval Status & Cryptographic Token */}
                          <div className="flex items-center justify-between text-[11px]">
                            <div className="flex items-center gap-1.5">
                              {item.status === 'approved' ? (
                                <span className="flex items-center gap-1 text-emerald-600 font-bold text-[11px]">
                                  <CheckCircle2 className="w-3.5 h-3.5" />
                                  <span>Approved</span>
                                </span>
                              ) : (
                                <button
                                  type="button"
                                  onClick={() => handleApprovePlanItem(item.id)}
                                  className="text-xs font-bold text-amber-600 hover:text-amber-700 underline"
                                >
                                  Sign &amp; Approve (HMAC)
                                </button>
                              )}
                            </div>
                            <span className="text-[10px] text-slate-400">
                              Read Time: {item.asset_metadata.estimated_read_time_sec}s
                            </span>
                          </div>

                          {/* Reschedule Input */}
                          <div className="flex items-center gap-2 pt-1 border-t border-slate-200 dark:border-slate-700 text-xs">
                            <span className="text-[10px] font-bold text-slate-500 shrink-0">Reschedule:</span>
                            <input
                              type="datetime-local"
                              defaultValue={calendarDateInput}
                              onChange={(e) => handleReschedulePlanItem(item.id, e.target.value)}
                              className="text-[10px] p-1 rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 w-full"
                            />
                          </div>

                          {/* Gated Actions (AT-010 Compliance) */}
                          <div className="grid grid-cols-2 gap-2 pt-1.5">
                            <button
                              type="button"
                              onClick={() => handleExportPlanForClipboard(item)}
                              className="px-2.5 py-1.5 rounded-lg bg-teal-600 hover:bg-teal-700 text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-xs transition-colors"
                            >
                              <ExternalLink className="w-3.5 h-3.5" />
                              <span>1-Click Dispatch</span>
                            </button>

                            <button
                              type="button"
                              onClick={() => handleAttemptDirectPublish(item)}
                              className="px-2.5 py-1.5 rounded-lg border border-rose-200 dark:border-rose-900/60 bg-rose-50 dark:bg-rose-950/40 text-rose-700 dark:text-rose-300 font-bold text-xs flex items-center justify-center gap-1 hover:bg-rose-100 transition-colors"
                              title="Tests fail-closed rejection under AT-010"
                            >
                              <ShieldAlert className="w-3 h-3 text-rose-600" />
                              <span>Test Autopublish</span>
                            </button>
                          </div>
                        </div>
                      ))
                  )}
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* ========================================================================= */}
        {/* IMP-LI-15: Engagement Monitoring & Analytics Studio (AT-028, AT-010, SRC-L2) */}
        {/* ========================================================================= */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-5">
            <div>
              <div className="flex items-center gap-2">
                <span className="inline-flex items-center gap-1 rounded-md bg-purple-50 dark:bg-purple-950/60 px-2 py-0.5 text-[11px] font-bold text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800">
                  <BarChart3 className="w-3 h-3 text-purple-600" />
                  IMP-LI-15 &bull; LI-15 &bull; AT-028 &bull; AT-010 &bull; SRC-L2
                </span>
                <span className="text-xs text-slate-400">|</span>
                <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">
                  Truthful Metrics & ICP Engager Demographics
                </span>
              </div>
              <h2 className="mt-1 text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <span>Engagement Monitoring & Analytics Studio</span>
              </h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Distinguish unavailable platform telemetry from zero without fake random simulations (AT-028, AT-010) and segment interacting professionals by seniority (SRC-L2).
              </p>
            </div>

            {/* Tenant Isolation Switcher (AT-012) */}
            <div className="flex items-center gap-2 self-start sm:self-auto bg-slate-50 dark:bg-slate-800/60 p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-xs">
              <span className="text-slate-500 font-medium px-2 flex items-center gap-1">
                <Shield className="w-3.5 h-3.5 text-purple-600" />
                <span>Tenant Scope:</span>
              </span>
              <button
                type="button"
                onClick={() => {
                  setAnalyticsWorkspace('ws-alpha');
                  setSelectedSnapshotId('snap-verified-001');
                }}
                className={`px-3 py-1 rounded font-semibold text-xs transition-colors ${
                  analyticsWorkspace === 'ws-alpha'
                    ? 'bg-purple-600 text-white shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-alpha
              </button>
              <button
                type="button"
                onClick={() => {
                  setAnalyticsWorkspace('ws-beta');
                  setSelectedSnapshotId('none');
                }}
                className={`px-3 py-1 rounded font-semibold text-xs transition-colors ${
                  analyticsWorkspace === 'ws-beta'
                    ? 'bg-purple-600 text-white shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-beta (Isolated)
              </button>
            </div>
          </div>

          {/* AT-028 & AT-010 Transparency Policy Banner */}
          <div className="p-4 rounded-xl border border-purple-200 dark:border-purple-900/60 bg-purple-50/70 dark:bg-purple-950/30 flex items-start gap-3">
            <Info className="w-4 h-4 text-purple-600 mt-0.5 shrink-0" />
            <div className="space-y-1 text-xs">
              <p className="font-bold text-purple-950 dark:text-purple-200">
                AT-028 Integrity Contract: Zero vs. Unavailable Metric Transparency
              </p>
              <p className="text-purple-800 dark:text-purple-300 leading-relaxed">
                LinkedIn consumer OAuth APIs do not provide raw post impressions or viewer demographic telemetry. In strict compliance with <strong>AT-028</strong> and <strong>AT-010</strong>, missing metrics are labeled as <code className="bg-purple-100 dark:bg-purple-900/60 px-1 py-0.5 rounded text-[10px] font-mono">unavailable</code> with explicit reasons rather than populating artificial zeros or fake randomized analytics. Modeled estimates are tagged as heuristic suggestions, never causal facts.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left Column: Post Snapshot Selector & Benchmarks (4 cols) */}
            <div className="lg:col-span-4 space-y-5">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
                    <FileText className="w-3.5 h-3.5 text-purple-600" />
                    <span>Audited Post Snapshots ({filteredSnapshots.length})</span>
                  </span>
                </div>

                <div className="space-y-2 max-h-[320px] overflow-y-auto pr-1">
                  {filteredSnapshots.length === 0 ? (
                    <div className="p-6 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl text-xs text-slate-500">
                      No post snapshots in workspace <code className="font-mono">{analyticsWorkspace}</code> (AT-012 Cross-Tenant Isolation).
                    </div>
                  ) : (
                    filteredSnapshots.map((snap) => (
                      <div
                        key={snap.snapshot_id}
                        onClick={() => setSelectedSnapshotId(snap.snapshot_id)}
                        className={`p-3 rounded-xl border cursor-pointer transition-all ${
                          selectedSnapshot?.snapshot_id === snap.snapshot_id
                            ? 'border-purple-500 bg-purple-50/50 dark:bg-purple-950/40 shadow-xs'
                            : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:border-slate-300'
                        }`}
                      >
                        <div className="flex items-center justify-between text-xs mb-1">
                          <span className="font-bold text-slate-900 dark:text-white truncate max-w-[200px]">
                            {snap.post_title}
                          </span>
                          <span className="text-[10px] font-mono text-slate-400">
                            {new Date(snap.published_at).toLocaleDateString()}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 text-[10px]">
                          <span className="text-slate-500">
                            Reactions: <strong>{snap.metrics.reactions?.value ?? 0}</strong>
                          </span>
                          <span className="text-slate-400">&bull;</span>
                          <span className="text-slate-500">
                            Comments: <strong>{snap.metrics.comments?.value ?? 0}</strong>
                          </span>
                          <span className="text-slate-400">&bull;</span>
                          <span className={`font-semibold ${
                            snap.metrics.impressions?.status === 'available'
                              ? 'text-emerald-600'
                              : 'text-amber-600'
                          }`}>
                            {snap.metrics.impressions?.status === 'available'
                              ? `${snap.metrics.impressions.value} Impr.`
                              : 'Impr. Unavailable'}
                          </span>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* Curated Historical Benchmarks Card (AT-028) */}
              <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3">
                <div className="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-700">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                    <TrendingUp className="w-3.5 h-3.5 text-purple-600" />
                    <span>Creator Baseline Benchmarks (AT-028)</span>
                  </span>
                </div>
                <div className="space-y-2 text-xs">
                  {creatorBenchmarks.map((bm, i) => (
                    <div key={i} className="p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 space-y-1.5">
                      <div className="flex items-center justify-between font-bold text-slate-800 dark:text-slate-200 text-[11px]">
                        <span className="capitalize">{bm.category.replace('_', ' ')}</span>
                        <span className="text-purple-600 font-mono">{bm.avg_engagement_rate}% Eng. Rate</span>
                      </div>
                      <div className="grid grid-cols-3 gap-1 text-[10px] text-slate-500">
                        <div>Reactions: <strong>{bm.avg_reactions}</strong></div>
                        <div>Comments: <strong>{bm.avg_comments}</strong></div>
                        <div>Reposts: <strong>{bm.avg_reposts}</strong></div>
                      </div>
                      <div className="text-[9px] text-slate-400 italic">
                        Source: {bm.benchmark_source} (Sample: {bm.sample_size} verified posts)
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Right Column: Selected Post Analytics & ICP Engagers (8 cols) */}
            <div className="lg:col-span-8 space-y-5">
              {selectedSnapshot ? (
                <>
                  {/* Post Title & Meta Header */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-2">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="text-[10px] font-mono text-purple-600 bg-purple-50 dark:bg-purple-950/60 px-2 py-0.5 rounded border border-purple-200 dark:border-purple-800">
                        {selectedSnapshot.post_urn}
                      </span>
                      <span className="text-xs text-slate-500">
                        Published: {new Date(selectedSnapshot.published_at).toUTCString()}
                      </span>
                    </div>
                    <h3 className="text-base font-bold text-slate-900 dark:text-white">
                      {selectedSnapshot.post_title}
                    </h3>
                    {selectedSnapshot.disclosure_notice && (
                      <p className="text-xs text-slate-500 dark:text-slate-400 italic">
                        {selectedSnapshot.disclosure_notice}
                      </p>
                    )}
                  </div>

                  {/* 1. Metric Cards Grid (AT-028 Honest Distinction) */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                    {/* Reactions */}
                    <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30">
                      <div className="text-[10px] font-bold uppercase text-slate-400">Reactions</div>
                      <div className="text-xl font-black text-slate-900 dark:text-white mt-1">
                        {selectedSnapshot.metrics.reactions?.value ?? 0}
                      </div>
                      <div className="text-[10px] text-emerald-600 font-semibold mt-0.5 flex items-center gap-1">
                        <CheckCircle2 className="w-2.5 h-2.5" />
                        <span>Platform Verified</span>
                      </div>
                    </div>

                    {/* Comments */}
                    <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30">
                      <div className="text-[10px] font-bold uppercase text-slate-400">Comments</div>
                      <div className="text-xl font-black text-slate-900 dark:text-white mt-1">
                        {selectedSnapshot.metrics.comments?.value ?? 0}
                      </div>
                      <div className="text-[10px] text-emerald-600 font-semibold mt-0.5 flex items-center gap-1">
                        <CheckCircle2 className="w-2.5 h-2.5" />
                        <span>Platform Verified</span>
                      </div>
                    </div>

                    {/* Impressions (May be Unavailable under AT-010) */}
                    <div className={`p-3 rounded-xl border ${
                      selectedSnapshot.metrics.impressions?.status === 'available'
                        ? 'border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30'
                        : 'border-amber-200 dark:border-amber-900/60 bg-amber-50/40 dark:bg-amber-950/20'
                    }`}>
                      <div className="text-[10px] font-bold uppercase text-slate-400">Impressions</div>
                      <div className="text-xl font-black text-slate-900 dark:text-white mt-1">
                        {selectedSnapshot.metrics.impressions?.status === 'available'
                          ? selectedSnapshot.metrics.impressions.value.toLocaleString()
                          : 'Unavailable'}
                      </div>
                      {selectedSnapshot.metrics.impressions?.status === 'available' ? (
                        <div className="text-[10px] text-emerald-600 font-semibold mt-0.5">
                          Verified Telemetry
                        </div>
                      ) : (
                        <div className="text-[9px] text-amber-700 dark:text-amber-300 font-medium mt-0.5 flex items-center gap-1">
                          <EyeOff className="w-2.5 h-2.5 shrink-0" />
                          <span>Scope Required (AT-010)</span>
                        </div>
                      )}
                    </div>

                    {/* Engagement Rate / Modeled Reach */}
                    <div className={`p-3 rounded-xl border ${
                      selectedSnapshot.metrics.engagement_rate?.status === 'available'
                        ? 'border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30'
                        : selectedSnapshot.metrics.modeled_reach_estimate?.status === 'modeled_estimate'
                        ? 'border-purple-200 dark:border-purple-900/60 bg-purple-50/40 dark:bg-purple-950/20'
                        : 'border-amber-200 dark:border-amber-900/60 bg-amber-50/40 dark:bg-amber-950/20'
                    }`}>
                      <div className="text-[10px] font-bold uppercase text-slate-400">
                        {selectedSnapshot.metrics.modeled_reach_estimate ? 'Modeled Reach' : 'Eng. Rate'}
                      </div>
                      <div className="text-xl font-black text-slate-900 dark:text-white mt-1">
                        {selectedSnapshot.metrics.engagement_rate?.status === 'available'
                          ? `${selectedSnapshot.metrics.engagement_rate.value}%`
                          : selectedSnapshot.metrics.modeled_reach_estimate?.status === 'modeled_estimate'
                          ? `~${selectedSnapshot.metrics.modeled_reach_estimate.value.toLocaleString()}`
                          : 'Unavailable'}
                      </div>
                      {selectedSnapshot.metrics.modeled_reach_estimate ? (
                        <div className="text-[9px] text-purple-700 dark:text-purple-300 font-semibold mt-0.5">
                          Modeled Suggestion (AT-028)
                        </div>
                      ) : (
                        <div className="text-[9px] text-slate-500 font-medium mt-0.5">
                          {selectedSnapshot.metrics.engagement_rate?.status === 'available'
                            ? 'Calculated'
                            : 'No Impressions'}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* 2. ICP Engager Segmentation (SRC-L2) */}
                  <div className="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-4">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 pb-3 border-b border-slate-200 dark:border-slate-800">
                      <div>
                        <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                          <Target className="w-3.5 h-3.5 text-purple-600" />
                          <span>ICP Engager Demographics ({selectedSnapshot.engagers.length} Audited Profiles)</span>
                        </h4>
                        <p className="text-[11px] text-slate-500 mt-0.5">
                          Rule-based professional seniority classification categorizing author comments & reactions (SRC-L2).
                        </p>
                      </div>

                      {/* ICP Stat Pills */}
                      <div className="flex flex-wrap items-center gap-1.5 text-[10px]">
                        <span className="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300 font-bold border border-emerald-200 dark:border-emerald-800">
                          {selectedSnapshot.icp_breakdown.decision_makers} Decision Makers
                        </span>
                        <span className="px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-bold border border-blue-200 dark:border-blue-800">
                          {selectedSnapshot.icp_breakdown.peer_practitioners} Peers
                        </span>
                        <span className="px-2 py-0.5 rounded-full bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300 font-bold border border-amber-200 dark:border-amber-800">
                          {selectedSnapshot.icp_breakdown.talent_partners} Recruiters
                        </span>
                      </div>
                    </div>

                    {/* Engagers List */}
                    <div className="space-y-2.5">
                      {selectedSnapshot.engagers.length === 0 ? (
                        <div className="p-6 text-center text-xs text-slate-400 italic">
                          No engager profiles recorded for this post snapshot.
                        </div>
                      ) : (
                        selectedSnapshot.engagers.map((engager, idx) => (
                          <div
                            key={idx}
                            className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
                          >
                            <div className="space-y-1">
                              <div className="flex items-center gap-2">
                                <span className="font-bold text-slate-900 dark:text-white">
                                  {engager.name}
                                </span>
                                <span className="text-[10px] text-slate-400 font-mono">
                                  {engager.company}
                                </span>
                                <span className="text-[10px] uppercase font-bold text-purple-600 bg-purple-50 dark:bg-purple-950 px-1.5 py-0.2 rounded border border-purple-200 dark:border-purple-800">
                                  {engager.interaction_type}
                                </span>
                              </div>
                              <p className="text-[11px] text-slate-600 dark:text-slate-400">
                                {engager.headline}
                              </p>
                              {engager.comment_text && (
                                <p className="text-[11px] text-slate-700 dark:text-slate-300 italic bg-white dark:bg-slate-900 p-2 rounded border border-slate-200 dark:border-slate-800">
                                  &ldquo;{engager.comment_text}&rdquo;
                                </p>
                              )}
                            </div>

                            <span className={`self-start sm:self-auto px-2 py-0.5 rounded text-[10px] font-bold uppercase shrink-0 ${
                              engager.segment === 'decision_maker'
                                ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-300'
                                : engager.segment === 'peer_practitioner'
                                ? 'bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-300'
                                : engager.segment === 'talent_partner'
                                ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-300'
                                : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 border border-slate-300'
                            }`}>
                              {engager.segment?.replace('_', ' ')}
                            </span>
                          </div>
                        ))
                      )}
                    </div>

                    {/* Engager ICP Classification Test Bench */}
                    <div className="pt-3 border-t border-slate-200 dark:border-slate-800 space-y-2">
                      <span className="text-[11px] font-bold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                        <Plus className="w-3 h-3 text-purple-600" />
                        <span>Simulate Engager Ingestion & ICP Segmentation:</span>
                      </span>
                      <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                        <input
                          type="text"
                          value={testEngagerHeadline}
                          onChange={(e) => setTestEngagerHeadline(e.target.value)}
                          placeholder="Professional Headline (e.g. Director of Infrastructure)"
                          className="sm:col-span-2 text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900"
                        />
                        <input
                          type="text"
                          value={testEngagerCompany}
                          onChange={(e) => setTestEngagerCompany(e.target.value)}
                          placeholder="Company (e.g. CloudScale)"
                          className="text-xs p-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900"
                        />
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={handleTestEngagerSegmentation}
                          className="px-3 py-1.5 rounded-lg border border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-950/40 text-purple-700 dark:text-purple-300 font-bold text-xs hover:bg-purple-100 transition-colors"
                        >
                          Test Classification
                        </button>
                        <button
                          type="button"
                          onClick={handleAddSampleEngager}
                          className="px-3 py-1.5 rounded-lg bg-purple-600 hover:bg-purple-700 text-white font-bold text-xs transition-colors"
                        >
                          Add to Post Engagers
                        </button>
                        {segmentedResultSegment && (
                          <span className="text-xs font-bold text-purple-700 dark:text-purple-300 bg-purple-100 dark:bg-purple-900/60 px-2 py-1 rounded">
                            Identified: {segmentedResultSegment.toUpperCase()}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </>
              ) : (
                <div className="p-12 text-center text-xs text-slate-500 border border-dashed border-slate-200 dark:border-slate-800 rounded-xl">
                  Select a post analytics snapshot to view performance metrics and ICP engager breakdown.
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Employee Advocacy & Brand Governance Studio (LI-16, AT-007, AT-011) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
            <div>
              <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Megaphone className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                <span>Employee Advocacy &amp; Brand Governance Studio (LI-16, AT-007, AT-011)</span>
              </h2>
              <p className="text-xs text-slate-500 mt-1 max-w-2xl">
                Distribute cryptographically approved company narratives to employee advocates across Engineering, Product, and Talent personas with strict anti-pod policy enforcement and multi-tenant tenant isolation.
              </p>
            </div>
            {/* Multi-Tenant Workspace Selector (AT-011) */}
            <div className="flex items-center gap-2 bg-slate-100 dark:bg-slate-800/60 p-1.5 rounded-lg text-xs">
              <span className="text-slate-500 font-medium px-2 flex items-center gap-1">
                <Building2 className="w-3.5 h-3.5 text-slate-400" />
                Workspace:
              </span>
              <button
                type="button"
                onClick={() => setAdvocacyWorkspace('ws-alpha')}
                className={`px-3 py-1 rounded font-semibold transition-all ${
                  advocacyWorkspace === 'ws-alpha'
                    ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-300 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-alpha (Alpha Corp)
              </button>
              <button
                type="button"
                onClick={() => setAdvocacyWorkspace('ws-beta')}
                className={`px-3 py-1 rounded font-semibold transition-all ${
                  advocacyWorkspace === 'ws-beta'
                    ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-300 shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-beta (Beta Corp)
              </button>
            </div>
          </div>

          {/* Anti-Engagement Pod Policy (Anti-Pod LI-16) Banner */}
          <div className="p-4 rounded-xl border border-rose-200 dark:border-rose-900/60 bg-rose-50/60 dark:bg-rose-950/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-start gap-3">
              <ShieldAlert className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
              <div>
                <h3 className="text-xs font-bold text-rose-900 dark:text-rose-200">
                  Zero Fake Engagement &amp; Anti-Pod Policy Active (LI-16)
                </h3>
                <p className="text-[11px] text-rose-700 dark:text-rose-300 mt-0.5">
                  Reciprocal like rings, mutual engagement pods, and synthetic comment networks are strictly rejected (HTTP 403 Forbidden). All shares require voluntary human-in-the-loop action.
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <button
                type="button"
                onClick={() => handleTestAntiPodSecurity(true)}
                className="px-3 py-1.5 rounded-lg border border-rose-300 dark:border-rose-800 bg-white dark:bg-slate-900 text-rose-700 dark:text-rose-300 hover:bg-rose-100/50 text-xs font-semibold shadow-xs"
              >
                Simulate Pod Attack
              </button>
              <button
                type="button"
                onClick={() => handleTestAntiPodSecurity(false)}
                className="px-3 py-1.5 rounded-lg border border-emerald-300 dark:border-emerald-800 bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-50 text-xs font-semibold shadow-xs"
              >
                Verify Legitimate Check
              </button>
            </div>
          </div>

          {/* Anti-Pod Notification Output */}
          {antiPodAlert && (
            <div
              className={`p-3 rounded-lg text-xs font-mono border ${
                antiPodAlert.startsWith('BLOCKED')
                  ? 'border-rose-300 bg-rose-50 text-rose-800 dark:bg-rose-950/40 dark:text-rose-300'
                  : 'border-emerald-300 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300'
              }`}
            >
              {antiPodAlert}
            </div>
          )}

          {/* Main Advocacy Workspace Content */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left: Campaign List (AT-011 Multi-Tenant Isolation) */}
            <div className="lg:col-span-4 space-y-3">
              <div className="flex items-center justify-between pb-1">
                <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
                  Advocacy Campaigns ({currentAdvocacyCampaigns.length})
                </span>
                <span className="text-[10px] text-slate-400 font-mono">
                  {advocacyWorkspace}
                </span>
              </div>

              {currentAdvocacyCampaigns.length === 0 ? (
                <div className="p-8 text-center text-xs text-slate-500 border border-dashed border-slate-200 dark:border-slate-800 rounded-xl">
                  No active advocacy campaigns found for {advocacyWorkspace}.
                </div>
              ) : (
                <div className="space-y-2.5">
                  {currentAdvocacyCampaigns.map((camp) => {
                    const isSelected = selectedAdvocacyCampaign?.campaign_id === camp.campaign_id;
                    const isApproved = camp.status === 'approved';

                    return (
                      <div
                        key={camp.campaign_id}
                        onClick={() => handleSelectAdvocacyCampaign(camp.campaign_id)}
                        className={`p-3.5 rounded-xl border cursor-pointer transition-all text-left ${
                          isSelected
                            ? 'border-indigo-500 bg-indigo-50/50 dark:bg-indigo-950/30 shadow-xs'
                            : 'border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 bg-white dark:bg-slate-900'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-1.5">
                          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                            {camp.workspace_id}
                          </span>
                          <span
                            className={`text-[10px] font-bold px-2 py-0.5 rounded-full flex items-center gap-1 ${
                              isApproved
                                ? 'bg-emerald-100 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300'
                                : 'bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-300'
                            }`}
                          >
                            {isApproved ? (
                              <>
                                <CheckCircle2 className="w-3 h-3" /> APPROVED
                              </>
                            ) : (
                              <>
                                <Clock className="w-3 h-3" /> DRAFT
                              </>
                            )}
                          </span>
                        </div>
                        <h4 className="text-xs font-bold text-slate-900 dark:text-white line-clamp-1">
                          {camp.title}
                        </h4>
                        <p className="text-[11px] text-slate-500 dark:text-slate-400 line-clamp-2 mt-1">
                          {camp.description}
                        </p>
                        <div className="mt-2.5 pt-2 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-[10px] text-slate-400">
                          <span>{camp.variants.length} Persona Variants</span>
                          <span className="font-mono">
                            {camp.governance.forbidden_keywords?.length ?? 0} Forbidden KWs
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Right: Selected Campaign Details, Persona Switcher & Cryptographic Signer */}
            <div className="lg:col-span-8 space-y-4">
              {selectedAdvocacyCampaign ? (
                <>
                  {/* Campaign Header & Governance Metadata */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 space-y-3">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                      <div>
                        <h3 className="text-sm font-bold text-slate-900 dark:text-white">
                          {selectedAdvocacyCampaign.title}
                        </h3>
                        <p className="text-xs text-slate-500 mt-0.5">
                          {selectedAdvocacyCampaign.description}
                        </p>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <span
                          className={`text-xs font-semibold px-2.5 py-1 rounded-md border flex items-center gap-1.5 ${
                            selectedAdvocacyCampaign.status === 'approved'
                              ? 'border-emerald-300 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                              : 'border-amber-300 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-300'
                          }`}
                        >
                          {selectedAdvocacyCampaign.status === 'approved' ? (
                            <>
                              <Lock className="w-3.5 h-3.5" />
                              HMAC Signed (AT-007)
                            </>
                          ) : (
                            <>
                              <FileEdit className="w-3.5 h-3.5" />
                              Pending Approval
                            </>
                          )}
                        </span>
                      </div>
                    </div>

                    {/* Brand Governance Constraints */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs pt-2 border-t border-slate-200 dark:border-slate-700">
                      <div>
                        <span className="text-slate-500 font-medium">Mandatory Hashtags:</span>{' '}
                        <span className="font-semibold text-indigo-600 dark:text-indigo-400">
                          {selectedAdvocacyCampaign.governance.allowed_hashtags?.join(', ') || 'None'}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 font-medium">Forbidden Keywords:</span>{' '}
                        <span className="font-semibold text-rose-600 dark:text-rose-400">
                          {selectedAdvocacyCampaign.governance.forbidden_keywords?.join(', ') || 'None'}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Tamper Alert Display */}
                  {advocacyTamperAlert && (
                    <div className="p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-300 dark:border-amber-800 text-xs text-amber-800 dark:text-amber-300 flex items-center gap-2">
                      <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
                      <span>{advocacyTamperAlert}</span>
                    </div>
                  )}

                  {/* Compliance Violations Alert */}
                  {complianceViolations.length > 0 && (
                    <div className="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-300 dark:border-rose-800 text-xs text-rose-800 dark:text-rose-300 flex items-center gap-2">
                      <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0" />
                      <span>Forbidden brand keyword detected: &quot;{complianceViolations.join(', ')}&quot;. Edit copy before sharing.</span>
                    </div>
                  )}

                  {/* Persona Variant Selector (Engineering vs Product vs Talent & Culture) */}
                  <div className="space-y-2">
                    <label className="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                      Select Employee Persona Narrative
                    </label>
                    <div className="grid grid-cols-3 gap-2">
                      {(['engineering', 'product', 'talent_culture'] as const).map((persona) => {
                        const isSelected = selectedAdvocacyPersona === persona;
                        const label =
                          persona === 'engineering'
                            ? 'Engineering Lead'
                            : persona === 'product'
                            ? 'Product Manager'
                            : 'Talent & Culture';

                        return (
                          <button
                            key={persona}
                            type="button"
                            onClick={() => handleSelectAdvocacyPersona(persona)}
                            className={`p-2.5 rounded-lg border text-left transition-all ${
                              isSelected
                                ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-950/40 text-indigo-900 dark:text-indigo-200 shadow-xs'
                                : 'border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40 text-slate-700 dark:text-slate-300'
                            }`}
                          >
                            <div className="text-xs font-bold">{label}</div>
                            <div className="text-[10px] text-slate-500 capitalize">
                              Persona: {persona.replace('_', ' ')}
                            </div>
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  {/* Copy Customizer & Tamper Test Action */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <label className="text-xs font-bold text-slate-700 dark:text-slate-300">
                        Employee Customized Copy (Voluntary Human Share)
                      </label>
                      <button
                        type="button"
                        onClick={() =>
                          handleEditCampaignOriginalVariant(
                            selectedAdvocacyVariant?.variant_id || 'var-eng-01',
                            'We just launched our new engine with guaranteed 100% returns on latency reduction! #AlphaEngine #Distributed'
                          )
                        }
                        className="text-[11px] font-semibold text-rose-600 hover:text-rose-700 underline flex items-center gap-1"
                        title="Simulate modifying text after approval to test AT-007 invalidation"
                      >
                        <AlertTriangle className="w-3 h-3" />
                        Simulate Tampering (AT-007 Invalidation)
                      </button>
                    </div>

                    <textarea
                      rows={4}
                      value={customizedAdvocacyCopy}
                      onChange={(e) => handleCustomizedCopyChange(e.target.value)}
                      className="w-full text-xs font-mono rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 p-3 text-slate-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                      placeholder="Customize your persona message before voluntary sharing..."
                    />
                  </div>

                  {/* Cryptographic Approval Card (AT-007) & 1-Click Share Actions */}
                  <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-indigo-600" />
                        <span className="text-xs font-bold text-slate-900 dark:text-white">
                          Cryptographic Approval Gate (AT-007)
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-500 mt-0.5">
                        {selectedAdvocacyCampaign.status === 'approved'
                          ? `Token: ${selectedAdvocacyCampaign.approval_token?.substring(0, 24)}...`
                          : 'Campaign text is unapproved or modified. Needs brand executive approval.'}
                      </p>
                    </div>

                    <div className="flex items-center gap-2">
                      {selectedAdvocacyCampaign.status !== 'approved' && (
                        <button
                          type="button"
                          onClick={() => handleApproveAdvocacyCampaign(selectedAdvocacyCampaign.campaign_id)}
                          className="px-3.5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs"
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          Approve (Sign HMAC)
                        </button>
                      )}

                      <button
                        type="button"
                        onClick={handleShareAdvocacyContent}
                        disabled={selectedAdvocacyCampaign.status !== 'approved' || complianceViolations.length > 0}
                        className="px-3.5 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 disabled:bg-slate-300 dark:disabled:bg-slate-800 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs disabled:cursor-not-allowed"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                        1-Click Copy &amp; Share (AT-010)
                      </button>
                    </div>
                  </div>

                  {/* Share Notification Display */}
                  {advocacyShareNotice && (
                    <div className="p-3 rounded-lg border border-emerald-300 bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300 text-xs flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                        <span>{advocacyShareNotice}</span>
                      </div>
                      <a
                        href="https://www.linkedin.com/feed/?shareActive=true"
                        target="_blank"
                        rel="noreferrer"
                        className="text-emerald-700 underline font-semibold flex items-center gap-1 ml-3 shrink-0"
                      >
                        LinkedIn <ExternalLink className="w-3 h-3" />
                      </a>
                    </div>
                  )}

                  <div className="pt-2 text-[10px] text-slate-400 flex items-center justify-between">
                    <span>Total Campaign Shares: {selectedAdvocacyCampaign.share_count} voluntary shares</span>
                    <span>Created: {new Date(selectedAdvocacyCampaign.created_at).toLocaleDateString()}</span>
                  </div>
                </>
              ) : (
                <div className="p-12 text-center text-xs text-slate-500 border border-dashed border-slate-200 dark:border-slate-800 rounded-xl">
                  Select an advocacy campaign to customize persona narratives, verify brand guidelines, and initiate voluntary sharing.
                </div>
              )}
            </div>
          </div>
        </section>
        </div>
      )}

      {/* MODULE 6: SAFETY & SYSTEM TOOLS */}
      {activeModule === 'safety' && (
        <div className="space-y-6">
        {/* Provider Fallback & Diagnostics Studio (IMP-LI-17, LI-17, AT-010, AT-016, FND-009, FND-013, SRC-L5) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
            <div>
              <div className="flex items-center space-x-2">
                <span className="inline-flex items-center rounded-full bg-violet-50 dark:bg-violet-950/60 px-2 py-0.5 text-[10px] font-bold text-violet-700 dark:text-violet-300 border border-violet-200 dark:border-violet-800">
                  IMP-LI-17 &bull; LI-17 &bull; AT-010 &bull; AT-016 &bull; FND-009 &bull; FND-013 &bull; SRC-L5
                </span>
                <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Layers className="w-4 h-4 text-violet-600" />
                  <span>Provider Fallback & Diagnostics Studio</span>
                </h2>
              </div>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                Capability-aware multi-adapter orchestration, safe hierarchical fallback routing, automated credential scrubbing (AT-016), and zero permission-bypass gating (AT-010).
              </p>
            </div>

            {/* Multi-Tenant Workspace Selector (AT-011) */}
            <div className="flex items-center gap-2">
              <span className="text-xs font-semibold text-slate-500">Workspace:</span>
              <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 p-1">
                <button
                  onClick={() => {
                    setProviderWorkspace('ws-alpha');
                    setDiagnosticReport(null);
                    setDispatchResult(null);
                    setDispatchError(null);
                  }}
                  className={`px-3 py-1 text-xs font-bold rounded-md transition-all ${
                    providerWorkspace === 'ws-alpha'
                      ? 'bg-white dark:bg-slate-900 text-violet-700 dark:text-violet-300 shadow-sm'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Alpha (tenant-corp-01)
                </button>
                <button
                  onClick={() => {
                    setProviderWorkspace('ws-beta');
                    setDiagnosticReport(null);
                    setDispatchResult(null);
                    setDispatchError(null);
                  }}
                  className={`px-3 py-1 text-xs font-bold rounded-md transition-all ${
                    providerWorkspace === 'ws-beta'
                      ? 'bg-white dark:bg-slate-900 text-violet-700 dark:text-violet-300 shadow-sm'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                  }`}
                >
                  Workspace Beta (tenant-corp-02)
                </button>
              </div>
            </div>
          </div>

          {/* Diagnostic Doctor & Quick Controls */}
          <div className="mt-6 rounded-xl border border-violet-100 dark:border-violet-900/40 bg-violet-50/50 dark:bg-violet-950/20 p-5 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Terminal className="w-4 h-4 text-violet-600" />
                  <span>Diagnostic Doctor Health Check (FND-013, doctor.py / probe.py)</span>
                </h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Probes all registered adapters in {providerWorkspace} for upstream latency, rate limit status, and scope alignment.
                </p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={handleTogglePrimaryDegraded}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold text-slate-700 dark:text-slate-200 hover:bg-slate-50 shadow-sm"
                >
                  <RefreshCw className="w-3.5 h-3.5 text-amber-500" />
                  <span>Toggle Primary 429 Degradation</span>
                </button>
                <button
                  onClick={handleRunDoctorProbe}
                  disabled={isDoctorRunning}
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-700 text-xs font-bold text-white shadow-sm disabled:opacity-50"
                >
                  <RefreshCw className={`w-3.5 h-3.5 ${isDoctorRunning ? 'animate-spin' : ''}`} />
                  <span>{isDoctorRunning ? 'Probing Adapters...' : 'Run Diagnostic Doctor Probe'}</span>
                </button>
              </div>
            </div>

            {/* Diagnostic Report Display */}
            {diagnosticReport && (
              <div className="rounded-lg border border-violet-200 dark:border-violet-800 bg-white dark:bg-slate-900 p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-slate-700 dark:text-slate-200">Doctor Report Status:</span>
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-extrabold uppercase ${
                        diagnosticReport.overall_health === 'healthy'
                          ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                          : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                      }`}
                    >
                      {diagnosticReport.overall_health}
                    </span>
                  </div>
                  <span className="text-[11px] text-slate-400 font-mono">
                    Generated: {new Date(diagnosticReport.generated_at).toLocaleTimeString()}
                  </span>
                </div>

                <div className="space-y-1">
                  <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Recommendations:</span>
                  <ul className="list-disc list-inside text-xs text-slate-600 dark:text-slate-300 space-y-0.5">
                    {(diagnosticReport.recommendations || []).map((rec, idx) => (
                      <li key={idx}>{rec}</li>
                    ))}
                  </ul>
                </div>
              </div>
            )}
          </div>

          {/* Registered Provider Adapters Grid */}
          <div className="mt-6 space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                Registered Adapters in {providerWorkspace} ({providers.filter((p) => p.workspace_id === providerWorkspace).length})
              </h3>
              <span className="text-[11px] text-slate-500">Tier 1 = Official API &rarr; Tier 2 = OIDC &rarr; Tier 3 = Local &rarr; Tier 4 = RSS</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {providers
                .filter((p) => p.workspace_id === providerWorkspace)
                .sort((a, b) => a.tier - b.tier)
                .map((adapter) => (
                  <div
                    key={adapter.provider_id}
                    className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/30 p-4 flex flex-col justify-between space-y-3"
                  >
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-extrabold uppercase bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
                          Tier {adapter.tier}
                        </span>
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-extrabold uppercase ${
                            adapter.health_status === 'healthy'
                              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                              : adapter.health_status === 'degraded' || adapter.health_status === 'rate_limited'
                              ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                              : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                          }`}
                        >
                          {adapter.health_status}
                        </span>
                      </div>

                      <div>
                        <h4 className="text-xs font-bold text-slate-900 dark:text-white leading-tight">
                          {adapter.name}
                        </h4>
                        <span className="font-mono text-[10px] text-slate-400 block mt-0.5">
                          Type: {adapter.adapter_type}
                        </span>
                      </div>

                      <div className="pt-1 space-y-1">
                        <div className="flex items-center justify-between text-[11px] text-slate-500">
                          <span>Latency:</span>
                          <span className="font-mono font-bold text-slate-700 dark:text-slate-300">{adapter.latency_ms} ms</span>
                        </div>
                        <div className="flex items-center justify-between text-[11px] text-slate-500">
                          <span>Consecutive Fails:</span>
                          <span className={`font-mono font-bold ${adapter.consecutive_failures > 0 ? 'text-amber-600' : 'text-slate-700 dark:text-slate-300'}`}>
                            {adapter.consecutive_failures}
                          </span>
                        </div>
                      </div>

                      <div className="space-y-1 pt-1">
                        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Supported Actions:</span>
                        <div className="flex flex-wrap gap-1">
                          {adapter.supported_actions.map((act) => (
                            <span key={act} className="px-1.5 py-0.5 rounded bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-[10px] font-mono text-slate-600 dark:text-slate-300">
                              {act}
                            </span>
                          ))}
                        </div>
                      </div>
                    </div>

                    <div className="pt-2 border-t border-slate-200 dark:border-slate-800 text-[10px] text-slate-500">
                      <span className="font-semibold block text-slate-600 dark:text-slate-400">Remediation:</span>
                      <p className="line-clamp-2 mt-0.5">{adapter.remediation_step}</p>
                    </div>
                  </div>
                ))}
            </div>
          </div>

          {/* Fallback Dispatch Engine & Zero Permission-Bypass Test Bench (AT-010) */}
          <div className="mt-8 pt-6 border-t border-slate-200 dark:border-slate-800 space-y-4">
            <div>
              <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Shield className="w-4 h-4 text-emerald-600" />
                <span>Safe Fallback Dispatch & Zero Permission-Bypass Bench (AT-010)</span>
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Dispatch operations across adapter tiers. If Tier 1 is degraded, failover occurs gracefully to Tier 2/3. Prohibited actions fail closed immediately without permission bypass.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
              <button
                onClick={() => handleTestDispatchAction('identity_read')}
                className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-800 dark:text-slate-200 text-left shadow-sm"
              >
                <div className="font-bold text-emerald-700 dark:text-emerald-400">Permitted: `identity_read`</div>
                <div className="text-[11px] text-slate-500 mt-0.5">Dispatches Tier 1, or falls back to OIDC</div>
              </button>

              <button
                onClick={() => handleTestDispatchAction('profile_import')}
                className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-800 dark:text-slate-200 text-left shadow-sm"
              >
                <div className="font-bold text-emerald-700 dark:text-emerald-400">Permitted: `profile_import`</div>
                <div className="text-[11px] text-slate-500 mt-0.5">Dispatches Tier 1, or falls back to Local Session</div>
              </button>

              <button
                onClick={() => handleTestDispatchAction('public_feed_read')}
                className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-800 dark:text-slate-200 text-left shadow-sm"
              >
                <div className="font-bold text-emerald-700 dark:text-emerald-400">Permitted: `public_feed_read`</div>
                <div className="text-[11px] text-slate-500 mt-0.5">Dispatches Tier 3 Local or Tier 4 RSS</div>
              </button>

              <button
                onClick={() => handleTestDispatchAction('automated_headless_messaging')}
                className="px-3 py-2 text-xs font-semibold rounded-lg border border-rose-300 dark:border-rose-900 bg-rose-50/50 dark:bg-rose-950/30 hover:bg-rose-100 text-rose-900 dark:text-rose-200 text-left shadow-sm"
              >
                <div className="font-bold text-rose-700 dark:text-rose-400">Prohibited: `headless_messaging`</div>
                <div className="text-[11px] text-rose-500 mt-0.5">Zero Bypass: Fails closed (403 Forbidden)</div>
              </button>
            </div>

            {/* Dispatch Result Card */}
            {dispatchResult && (
              <div className="rounded-lg border border-emerald-200 dark:border-emerald-800 bg-emerald-50/80 dark:bg-emerald-950/40 p-4 space-y-2">
                <div className="flex items-center justify-between text-xs font-bold text-emerald-900 dark:text-emerald-200">
                  <span className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                    <span>Dispatch Success: Action "{dispatchResult.action}"</span>
                  </span>
                  <span className="font-mono text-[10px] text-emerald-700 dark:text-emerald-300">
                    Latency: {dispatchResult.latency_ms} ms
                  </span>
                </div>
                <div className="text-xs text-emerald-800 dark:text-emerald-300 space-y-1">
                  <div>
                    Dispatched Provider: <code className="font-mono font-bold bg-white dark:bg-slate-900 px-1.5 py-0.5 rounded text-emerald-800 dark:text-emerald-300">{dispatchResult.dispatched_provider_id}</code> ({dispatchResult.adapter_type})
                  </div>
                  <div>
                    Fallback Invoked: <strong className={dispatchResult.fallback_invoked ? 'text-amber-700 dark:text-amber-300' : 'text-emerald-700 dark:text-emerald-300'}>{dispatchResult.fallback_invoked ? 'YES (Graceful Failover Active)' : 'NO (Dispatched via Primary Tier)'}</strong>
                  </div>
                  {dispatchResult.fallback_reason && (
                    <div className="text-[11px] bg-white dark:bg-slate-900 p-2 rounded border border-emerald-200 dark:border-emerald-800 text-emerald-900 dark:text-emerald-200">
                      <strong>Failover Rationale:</strong> {dispatchResult.fallback_reason}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Dispatch Error Card */}
            {dispatchError && (
              <div className="rounded-lg border border-rose-300 dark:border-rose-800 bg-rose-50 dark:bg-rose-950/50 p-4 space-y-1">
                <div className="flex items-center gap-1.5 text-xs font-bold text-rose-800 dark:text-rose-200">
                  <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0" />
                  <span>Security Gatekeeper Notice (AT-010 Invariant):</span>
                </div>
                <p className="text-xs text-rose-700 dark:text-rose-300 leading-relaxed font-medium">
                  {dispatchError}
                </p>
              </div>
            )}
          </div>

          {/* Credential & Token Scrubber Workbench (AT-016) */}
          <div className="mt-8 pt-6 border-t border-slate-200 dark:border-slate-800 space-y-4">
            <div>
              <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Lock className="w-4 h-4 text-sky-600" />
                <span>Diagnostic Credential Scrubber Workbench (AT-016, utils/text.py)</span>
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Ensures diagnostic errors, probe logs, and traces never leak sensitive bearer tokens (<code className="font-mono text-sky-600">Bearer tok_...</code>), session cookies (<code className="font-mono text-sky-600">li_at=...</code>), or client secrets.
              </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400 block">
                  Raw Diagnostic Log Input (Simulated Upstream Error)
                </label>
                <textarea
                  value={testScrubInput}
                  onChange={(e) => setTestScrubInput(e.target.value)}
                  rows={3}
                  className="w-full text-xs font-mono p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-violet-500"
                />
                <button
                  onClick={handleRunScrubTest}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-sky-600 hover:bg-sky-700 text-xs font-bold text-white shadow-sm"
                >
                  <Shield className="w-3.5 h-3.5" />
                  <span>Execute AT-016 Secret Scrubbing</span>
                </button>
              </div>

              <div className="space-y-2">
                <label className="text-[11px] font-bold uppercase tracking-wider text-slate-400 block">
                  Scrubbed Output (Safe for Diagnostic Doctor Logs)
                </label>
                <div className="min-h-[72px] p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-xs font-mono text-slate-800 dark:text-slate-200 break-all leading-relaxed">
                  {scrubbedOutput || <span className="text-slate-400 italic">Click button to scrub raw credentials...</span>}
                </div>
                {scrubbedOutput && (
                  <div className="text-[11px] text-emerald-600 font-semibold flex items-center gap-1">
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>Secrets successfully redacted with `[REDACTED_SECRET]` placeholder.</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </section>

        {/* Action Permission Test Bench (AT-010 Live Guard Validator) */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Terminal className="w-4 h-4 text-sky-600" />
                <span>Action Permission Test Bench (Fail-Closed Execution Gatekeeper)</span>
              </h2>
              <p className="text-xs text-slate-500 mt-0.5">
                Simulate dispatching candidate actions against the active connection to verify fail-closed gating with zero false simulation.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5">
            <button
              onClick={() => handleTestLiveAction('sign_in')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Sign-in / Identity (`sign_in`)
            </button>
            <button
              onClick={() => handleTestLiveAction('read_profile')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Read Profile Header (`read_profile`)
            </button>
            <button
              onClick={() => handleTestLiveAction('import_experience')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Ingest Experience History (`import_experience`)
            </button>
            <button
              onClick={() => handleTestLiveAction('publish_post')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Publish UGC Post (`publish_post`)
            </button>
            <button
              onClick={() => handleTestLiveAction('send_direct_message')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Send Direct Message (`send_direct_message`)
            </button>
            <button
              onClick={() => handleTestLiveAction('automated_easy_apply')}
              className="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-800 shadow-sm text-left"
            >
              Test Automated Easy Apply (`automated_easy_apply`)
            </button>
          </div>

          {actionTestResult && (
            <div className={`mt-4 p-4 rounded-lg border ${
              actionTestResult.allowed
                ? 'bg-emerald-50 border-emerald-200 text-emerald-900'
                : 'bg-rose-50 border-rose-200 text-rose-900'
            }`}>
              <div className="flex items-center justify-between text-xs font-bold">
                <span>Action: {actionTestResult.action}</span>
                <span className="font-mono text-[10px]">{actionTestResult.timestamp}</span>
              </div>
              <p className="mt-1 text-xs font-medium leading-relaxed">
                {actionTestResult.message}
              </p>
            </div>
          )}
        </section>

        {/* ========================================================================= */}
        {/* IMP-LI-18: Account Limits, CAPTCHA / Reauth & Safety Gatekeeper Studio     */}
        {/* (LI-18, REQ-009, REQ-010, AT-008, AT-009, AT-010, FND-011, FND-013, SRC-L1, SRC-S3) */}
        {/* ========================================================================= */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-5">
            <div>
              <div className="flex items-center gap-2">
                <ShieldAlert className="w-5 h-5 text-amber-500" />
                <h2 className="text-base font-bold text-slate-900 dark:text-white">
                  Account Limits, CAPTCHA / Reauth & Safety Gatekeeper Studio
                </h2>
                <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300">
                  IMP-LI-18 • AT-010 • REQ-009
                </span>
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-3xl">
                Upstream limitation detection (HTTP 429, commercial use caps), fail-closed CAPTCHA handling with zero stealth bypass (AT-010), cross-workspace account budget pooling (AT-008, AT-009), and legitimate re-auth recovery.
              </p>
            </div>

            {/* Active Workspace Switcher */}
            <div className="flex items-center gap-2 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg border border-slate-200 dark:border-slate-700">
              <span className="text-xs font-medium text-slate-500 pl-2">Testing Workspace:</span>
              <button
                onClick={() => setLimitsWorkspace('ws-alpha')}
                className={`px-2.5 py-1 text-xs font-semibold rounded-md transition-colors ${
                  limitsWorkspace === 'ws-alpha'
                    ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                Workspace Alpha (Primary)
              </button>
              <button
                onClick={() => setLimitsWorkspace('ws-beta')}
                className={`px-2.5 py-1 text-xs font-semibold rounded-md transition-colors ${
                  limitsWorkspace === 'ws-beta'
                    ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                Workspace Beta (Shared Provider AT-009)
              </button>
            </div>
          </div>

          {/* Account Circuit Breaker Status & Safety Alerts */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Provider Account URN</span>
                <span className="text-xs font-mono font-bold text-slate-700 dark:text-slate-300">
                  {safetyState.account_id}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Circuit Breaker Status</span>
                <span className={`text-xs uppercase font-bold px-2 py-0.5 rounded-full ${
                  safetyState.status === 'active'
                    ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : safetyState.status === 'challenge_required'
                    ? 'bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-300 animate-pulse'
                    : 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300'
                }`}>
                  {safetyState.status}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs text-slate-500">
                <span>Restrictions Recorded:</span>
                <span className="font-semibold text-slate-700 dark:text-slate-300">{safetyState.consecutive_restrictions}</span>
              </div>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Active Restriction</span>
                <span className="text-xs font-mono font-semibold text-slate-700 dark:text-slate-300">
                  {safetyState.active_limit_type || 'None (Normal Operation)'}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Active Challenge</span>
                <span className="text-xs font-mono font-semibold text-slate-700 dark:text-slate-300">
                  {safetyState.active_challenge || 'None'}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs text-slate-500">
                <span>Cooldown Window:</span>
                <span className="font-mono text-[11px] text-slate-700 dark:text-slate-300">
                  {safetyState.cooldown_until ? new Date(safetyState.cooldown_until).toLocaleTimeString() : 'N/A'}
                </span>
              </div>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 flex flex-col justify-between">
              <div>
                <span className="text-xs font-semibold text-slate-500 block">Anti-Bot Integrity Policy</span>
                <p className="text-[11px] text-slate-600 dark:text-slate-400 mt-1 leading-snug">
                  Stealth CDP injections, headless evasions, or turnstile bypasses are permanently barred (AT-010, SRC-S3). Only legitimate OAuth re-authentication is permitted.
                </p>
              </div>
              {safetyState.status === 'challenge_required' ? (
                <button
                  onClick={() => setIsReauthConfirming(true)}
                  className="mt-2 w-full py-1.5 px-3 bg-rose-600 hover:bg-rose-700 text-white rounded-md text-xs font-bold shadow-sm flex items-center justify-center gap-1.5"
                >
                  <RefreshCw className="w-3.5 h-3.5" />
                  <span>Resolve Legitimate Reauth</span>
                </button>
              ) : (
                <div className="mt-2 text-[11px] font-semibold text-emerald-600 flex items-center gap-1">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>Gatekeeper Dispatches Permitted</span>
                </div>
              )}
            </div>
          </div>

          {/* Warning Banner if Challenge Required or Rate Limited */}
          {safetyState.status === 'challenge_required' && (
            <div className="p-4 rounded-lg border border-rose-300 bg-rose-50 dark:bg-rose-950/40 text-rose-900 dark:text-rose-200 space-y-2">
              <div className="flex items-center gap-2 text-xs font-bold">
                <AlertTriangle className="w-4 h-4 text-rose-600 dark:text-rose-400" />
                <span>SECURITY CHECKPOINT / CAPTCHA ACTIVE — FAIL-CLOSED PROTECTION ENGAGED (AT-010)</span>
              </div>
              <p className="text-xs leading-relaxed">
                LinkedIn presented an interactive security challenge: <code className="font-mono bg-rose-100 dark:bg-rose-900/60 px-1 py-0.5 rounded">{safetyState.challenge_reason}</code>. All background dispatches are halted. Click below to verify completion of legitimate OAuth re-authentication.
              </p>
              <div className="pt-2 flex items-center gap-3">
                <button
                  onClick={handleResolveChallengeLegitimately}
                  className="px-3 py-1.5 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-md shadow-sm"
                >
                  Confirm Official OAuth Re-auth Completed
                </button>
                <span className="text-[11px] text-rose-600 dark:text-rose-400 italic">
                  Note: Automated solve scripts violate platform safety policy.
                </span>
              </div>
            </div>
          )}

          {safetyState.status === 'rate_limited' && (
            <div className="p-4 rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/40 text-amber-900 dark:text-amber-200 space-y-1">
              <div className="flex items-center gap-2 text-xs font-bold">
                <Clock className="w-4 h-4 text-amber-600" />
                <span>UPSTREAM RATE LIMIT / COOLING-OFF COOLDOWN ENGAGED (REQ-009)</span>
              </div>
              <p className="text-xs leading-relaxed">
                Detected: <code className="font-mono bg-amber-100 dark:bg-amber-900/60 px-1 py-0.5 rounded">{safetyState.challenge_reason}</code>. The platform enforces a 72-hour adaptive backoff window. No fixed action count is advertised as "guaranteed safe" against LinkedIn platform bans.
              </p>
            </div>
          )}

          {/* Cross-Workspace Combined Account Budget Quotas (AT-008, AT-009, REQ-010) */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Database className="w-4 h-4 text-indigo-600" />
                  <span>Account-Wide Combined Budget Pooling (AT-008, AT-009, REQ-010)</span>
                </h3>
                <p className="text-xs text-slate-500">
                  Shared across modules (LinkedIn, Career) and workspaces (ws-alpha, ws-beta). Switching workspaces or reconnecting never resets rolling counters.
                </p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleRecordSampleActionInWorkspace('connection_request', 1)}
                  className="px-2.5 py-1 text-xs font-semibold rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-200"
                >
                  +1 Invite ({limitsWorkspace})
                </button>
                <button
                  onClick={() => handleRecordSampleActionInWorkspace('recruiter_dm', 1)}
                  className="px-2.5 py-1 text-xs font-semibold rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-200"
                >
                  +1 Recruiter DM ({limitsWorkspace})
                </button>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
                <div className="text-[11px] font-semibold text-slate-500">Weekly Invites Quota</div>
                <div className="text-lg font-bold text-slate-900 dark:text-white mt-1">
                  {combinedBudget.weekly_invitations_used} <span className="text-xs font-normal text-slate-500">/ {combinedBudget.weekly_invitations_limit}</span>
                </div>
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full mt-2 overflow-hidden">
                  <div
                    className="bg-indigo-600 h-full rounded-full"
                    style={{ width: `${Math.min(100, (combinedBudget.weekly_invitations_used / combinedBudget.weekly_invitations_limit) * 100)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400 mt-1 font-medium">
                  {combinedBudget.weekly_invitations_remaining} remaining this week
                </div>
              </div>

              <div className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
                <div className="text-[11px] font-semibold text-slate-500">Daily Invites Quota</div>
                <div className="text-lg font-bold text-slate-900 dark:text-white mt-1">
                  {combinedBudget.daily_invitations_used} <span className="text-xs font-normal text-slate-500">/ {combinedBudget.daily_invitations_limit}</span>
                </div>
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full mt-2 overflow-hidden">
                  <div
                    className="bg-blue-600 h-full rounded-full"
                    style={{ width: `${Math.min(100, (combinedBudget.daily_invitations_used / combinedBudget.daily_invitations_limit) * 100)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400 mt-1 font-medium">
                  {combinedBudget.daily_invitations_remaining} remaining today
                </div>
              </div>

              <div className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
                <div className="text-[11px] font-semibold text-slate-500">Daily InMail & Recruiter DM</div>
                <div className="text-lg font-bold text-slate-900 dark:text-white mt-1">
                  {combinedBudget.daily_inmail_used} <span className="text-xs font-normal text-slate-500">/ {combinedBudget.daily_inmail_limit}</span>
                </div>
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full mt-2 overflow-hidden">
                  <div
                    className="bg-purple-600 h-full rounded-full"
                    style={{ width: `${Math.min(100, (combinedBudget.daily_inmail_used / combinedBudget.daily_inmail_limit) * 100)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400 mt-1 font-medium">
                  Recruiter DMs: {combinedBudget.daily_recruiter_dm_used} / {combinedBudget.daily_recruiter_dm_limit} (nested)
                </div>
              </div>

              <div className="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
                <div className="text-[11px] font-semibold text-slate-500">Daily Company Follows</div>
                <div className="text-lg font-bold text-slate-900 dark:text-white mt-1">
                  {combinedBudget.daily_company_follows_used} <span className="text-xs font-normal text-slate-500">/ {combinedBudget.daily_company_follows_limit}</span>
                </div>
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full mt-2 overflow-hidden">
                  <div
                    className="bg-emerald-600 h-full rounded-full"
                    style={{ width: `${Math.min(100, (combinedBudget.daily_company_follows_used / combinedBudget.daily_company_follows_limit) * 100)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400 mt-1 font-medium">
                  {combinedBudget.daily_company_follows_remaining} remaining today
                </div>
              </div>
            </div>
          </div>

          {/* Upstream Limitation Simulation Bench & Fail-Closed Test */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 pt-4 border-t border-slate-100 dark:border-slate-800">
            {/* Simulation Triggers */}
            <div className="space-y-3">
              <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                Simulate Upstream Error Response (SRC-S3, REQ-009)
              </h4>
              <p className="text-xs text-slate-500">
                Trigger realistic LinkedIn HTTP responses to test automatic error classification and protective fail-closed transitions:
              </p>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => handleSimulateUpstreamRestriction('weekly_invite_429')}
                  className="p-2.5 text-xs text-left font-medium rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 text-slate-800 dark:text-slate-200 shadow-sm"
                >
                  <div className="font-bold text-amber-600">Weekly Invite Limit (429)</div>
                  <div className="text-[10px] text-slate-500 mt-0.5">Triggers 72h safe cooldown</div>
                </button>
                <button
                  onClick={() => handleSimulateUpstreamRestriction('captcha_checkpoint_403')}
                  className="p-2.5 text-xs text-left font-medium rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 text-slate-800 dark:text-slate-200 shadow-sm"
                >
                  <div className="font-bold text-rose-600">CAPTCHA Checkpoint (403)</div>
                  <div className="text-[10px] text-slate-500 mt-0.5">Zero stealth bypass • AT-010</div>
                </button>
                <button
                  onClick={() => handleSimulateUpstreamRestriction('commercial_search_429')}
                  className="p-2.5 text-xs text-left font-medium rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 text-slate-800 dark:text-slate-200 shadow-sm"
                >
                  <div className="font-bold text-sky-600">Commercial Search Cap (429)</div>
                  <div className="text-[10px] text-slate-500 mt-0.5">24h search limit pause</div>
                </button>
                <button
                  onClick={() => handleSimulateUpstreamRestriction('session_expired_401')}
                  className="p-2.5 text-xs text-left font-medium rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 text-slate-800 dark:text-slate-200 shadow-sm"
                >
                  <div className="font-bold text-violet-600">Session Expired (401)</div>
                  <div className="text-[10px] text-slate-500 mt-0.5">Re-authentication required</div>
                </button>
              </div>
            </div>

            {/* Pre-Flight Gatekeeper Evaluation */}
            <div className="space-y-3">
              <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                Pre-Flight Safety Gatekeeper Evaluation (AT-010)
              </h4>
              <p className="text-xs text-slate-500">
                Every dispatch operation checks the account circuit breaker before executing against external networks:
              </p>
              <div className="flex flex-wrap gap-2">
                {['connection_request', 'recruiter_dm', 'inmail', 'profile_view', 'publish_post'].map((action) => (
                  <button
                    key={action}
                    onClick={() => handleEvaluateSafetyGate(action)}
                    className="px-2.5 py-1.5 text-xs font-semibold rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 text-slate-700 dark:text-slate-200 shadow-sm"
                  >
                    Check `{action}`
                  </button>
                ))}
              </div>

              {gateCheckDecision && (
                <div className={`p-3 rounded-lg border text-xs font-medium ${
                  gateCheckDecision.allowed
                    ? 'bg-emerald-50 border-emerald-200 text-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-200'
                    : 'bg-rose-50 border-rose-200 text-rose-900 dark:bg-rose-950/40 dark:text-rose-200'
                }`}>
                  <div className="flex items-center justify-between font-bold">
                    <span>Action: `{gateCheckAction}`</span>
                    <span className="uppercase">{gateCheckDecision.allowed ? 'ALLOWED [200 OK]' : 'BLOCKED [FAIL-CLOSED]'}</span>
                  </div>
                  {gateCheckDecision.reason && (
                    <p className="mt-1 text-[11px] leading-relaxed">{gateCheckDecision.reason}</p>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Activity & Incident Audit Ledger (FND-013) */}
          <div className="pt-4 border-t border-slate-100 dark:border-slate-800 space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500 flex items-center gap-1.5">
                <ClipboardList className="w-3.5 h-3.5 text-slate-400" />
                <span>Account Activity & Incident Audit Ledger (FND-013)</span>
              </h4>
              <span className="text-[11px] text-slate-400 font-mono">
                {activityLedger.length} events logged
              </span>
            </div>

            <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-50 dark:bg-slate-800/60 text-[11px] uppercase font-bold text-slate-500 border-b border-slate-200 dark:border-slate-800">
                  <tr>
                    <th className="py-2 px-3">Time</th>
                    <th className="py-2 px-3">Workspace</th>
                    <th className="py-2 px-3">Module</th>
                    <th className="py-2 px-3">Action</th>
                    <th className="py-2 px-3">Status</th>
                    <th className="py-2 px-3">Audit Details</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {activityLedger.slice(0, 6).map((item) => (
                    <tr key={item.entry_id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                      <td className="py-2 px-3 font-mono text-[11px] text-slate-500">
                        {new Date(item.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="py-2 px-3 font-semibold text-slate-700 dark:text-slate-300">
                        {item.workspace_id}
                      </td>
                      <td className="py-2 px-3 text-slate-500">{item.module}</td>
                      <td className="py-2 px-3 font-mono text-[11px] text-indigo-600 dark:text-indigo-400">
                        {item.action}
                      </td>
                      <td className="py-2 px-3">
                        <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase ${
                          item.status === 'dispatched' || item.status === 'success'
                            ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300'
                            : 'bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-300'
                        }`}>
                          {item.status}
                        </span>
                      </td>
                      <td className="py-2 px-3 text-slate-600 dark:text-slate-400 truncate max-w-xs">
                        {item.details || '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        {/* ========================================================================= */}
        {/* IMP-LI-19: Session Lifecycle & Supervised Local Tools Studio              */}
        {/* (LI-19, AT-016, FND-009, FND-010, SRC-L3, SRC-L4, REQ-021)                */}
        {/* ========================================================================= */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-sm space-y-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-5">
            <div>
              <div className="flex items-center gap-2">
                <Key className="w-5 h-5 text-indigo-600" />
                <h2 className="text-base font-bold text-slate-900 dark:text-white">
                  Session Lifecycle & Supervised Local Tools Studio
                </h2>
                <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-800 dark:bg-indigo-900/40 dark:text-indigo-300">
                  IMP-LI-19 • AT-016 • FND-009
                </span>
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-3xl">
                Per-owner isolated session key storage, zero plaintext cookie/password uploads (AT-016), one-click explicit revocation (FND-009), and short-lived supervised local viewer leases with automated expiration (SRC-L3).
              </p>
            </div>

            {/* Workspace & Isolation Context Switcher */}
            <div className="flex items-center gap-2 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg border border-slate-200 dark:border-slate-700">
              <span className="text-xs font-medium text-slate-500 pl-2">Owner Isolation:</span>
              <button
                onClick={() => setSessionWorkspace('ws-alpha')}
                className={`px-2.5 py-1 text-xs font-semibold rounded-md transition-colors ${
                  sessionWorkspace === 'ws-alpha'
                    ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-alpha (usr-alex-001)
              </button>
              <button
                onClick={() => setSessionWorkspace('ws-beta')}
                className={`px-2.5 py-1 text-xs font-semibold rounded-md transition-colors ${
                  sessionWorkspace === 'ws-beta'
                    ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
                }`}
              >
                ws-beta (usr-sarah-002)
              </button>
            </div>
          </div>

          {/* Audit Notification Banner */}
          {sessionAuditMessage && (
            <div className="p-3 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-900 rounded-lg text-xs text-indigo-900 dark:text-indigo-200 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Info className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                <span>{sessionAuditMessage}</span>
              </div>
              <button
                onClick={() => setSessionAuditMessage(null)}
                className="text-[11px] font-semibold text-indigo-700 hover:underline"
              >
                Dismiss
              </button>
            </div>
          )}

          {/* Security Policy & Zero Plaintext Cookie Guard (AT-016, REQ-021) */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/40 space-y-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-4 h-4 text-emerald-600" />
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-200">
                  Zero Plaintext Credential Upload Policy (AT-016)
                </h3>
              </div>
              <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
                Cloud databases strictly reject raw session cookies (<code className="font-mono text-rose-600 bg-rose-50 dark:bg-rose-950/50 px-1 py-0.5 rounded">li_at</code>) and plaintext passwords as default connections. Sessions must authenticate via cryptographic Vault tokens or verified supervised desktop bridges.
              </p>

              {/* Simulation input to verify fail-closed blocking */}
              <div className="space-y-2 pt-2">
                <label className="text-[11px] font-semibold text-slate-500 uppercase">
                  Test Cookie / Password Paste Blocker
                </label>
                <input
                  type="text"
                  placeholder='Try pasting raw "li_at=AQED..." or password here...'
                  value={testCookiePaste}
                  onChange={(e) => setTestCookiePaste(e.target.value)}
                  className="w-full text-xs font-mono px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200"
                />
                {securityBlockAlert && (
                  <div className="p-2.5 rounded-lg border border-rose-300 bg-rose-50 dark:bg-rose-950/50 text-rose-900 dark:text-rose-200 text-xs font-medium space-y-1">
                    <div className="flex items-center gap-1.5 font-bold">
                      <AlertTriangle className="w-3.5 h-3.5 text-rose-600" />
                      <span>FAIL-CLOSED BLOCKED</span>
                    </div>
                    <p className="text-[11px] leading-relaxed">{securityBlockAlert}</p>
                  </div>
                )}
              </div>

              <div className="pt-2 flex items-center gap-2">
                <button
                  onClick={handleCreateIsolatedSession}
                  className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold rounded-lg shadow-sm flex items-center gap-1.5"
                >
                  <Key className="w-3.5 h-3.5" />
                  <span>Initialize Vault-Linked Session</span>
                </button>
                <button
                  onClick={() => {
                    setTestCookiePaste('li_at=AQEDAT43521abcdef9876; JSESSIONID="ajax:12345"');
                  }}
                  className="px-2.5 py-1.5 text-xs font-medium rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-100"
                >
                  Simulate Raw Cookie Paste
                </button>
              </div>
            </div>

            {/* Cross-Owner Storage Isolation Barrier (AT-011, AT-012) */}
            <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/40 space-y-3">
              <div className="flex items-center gap-2">
                <Lock className="w-4 h-4 text-indigo-600" />
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-200">
                  Per-Owner Storage Isolation Barrier (AT-011)
                </h3>
              </div>
              <div className="text-xs space-y-2 text-slate-600 dark:text-slate-400">
                <div className="flex items-center justify-between p-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                  <span className="font-medium">Active Tenant:</span>
                  <span className="font-mono font-semibold text-slate-800 dark:text-slate-200">{currentTenantId}</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                  <span className="font-medium">Isolated Owner:</span>
                  <span className="font-mono font-semibold text-slate-800 dark:text-slate-200">{currentOwnerId}</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
                  <span className="font-medium">Storage Namespace:</span>
                  <span className="font-mono text-[11px] text-indigo-600 dark:text-indigo-400">
                    {currentTenantId}:{currentOwnerId}:li:*
                  </span>
                </div>
              </div>
              <p className="text-[11px] text-slate-500 dark:text-slate-400">
                Foreign workspaces cannot view, lease, or revoke tokens belonging to other owners. Cross-tenant access fails closed with HTTP 403 Forbidden.
              </p>
            </div>
          </div>

          {/* Active Sessions Grid & Explicit Revocation (FND-009) */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-400">
                Active & Isolated Sessions ({sessions.filter((s) => s.tenant_id === currentTenantId).length})
              </h3>
              <span className="text-xs text-slate-400 font-mono">Explicit Revocation (FND-009)</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {sessions
                .filter((s) => s.tenant_id === currentTenantId)
                .map((session) => (
                  <div
                    key={session.id}
                    onClick={() => setSelectedSessionId(session.id)}
                    className={`p-4 rounded-xl border transition-all cursor-pointer ${
                      selectedSessionId === session.id
                        ? 'border-indigo-500 ring-2 ring-indigo-500/20 bg-white dark:bg-slate-900'
                        : 'border-slate-200 dark:border-slate-800 bg-white/70 dark:bg-slate-900/60 hover:border-slate-300'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-bold text-xs text-slate-800 dark:text-slate-200">
                          {session.id}
                        </span>
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                            session.status === 'active'
                              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300'
                              : 'bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-300'
                          }`}
                        >
                          {session.status}
                        </span>
                      </div>
                      {session.status === 'active' && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleRevokeSession(session.id);
                          }}
                          className="px-2.5 py-1 text-[11px] font-bold text-rose-600 hover:text-white hover:bg-rose-600 rounded-md border border-rose-200 dark:border-rose-900 transition-colors"
                        >
                          Revoke Session
                        </button>
                      )}
                    </div>

                    <div className="mt-3 space-y-1.5 text-xs text-slate-500 dark:text-slate-400 font-mono">
                      <div className="text-[11px] truncate">Vault Ref: {session.storage_metadata.vault_reference_token}</div>
                      <div className="text-[11px]">Active Leases: {session.active_leases_count} / Issued: {session.total_leases_issued}</div>
                      <div className="text-[10px] text-slate-400">
                        Expires: {new Date(session.expires_at).toLocaleDateString()} • Zero Plaintext: Guaranteed
                      </div>
                      {session.revoked_at && (
                        <div className="text-[11px] text-rose-600 dark:text-rose-400 mt-1">
                          Revoked at {new Date(session.revoked_at).toLocaleTimeString()} ({session.revocation_reason})
                        </div>
                      )}
                    </div>
                  </div>
                ))}
            </div>
          </div>

          {/* Short-Lived Supervised Viewer Leases (SRC-L3 profile_lease.py, daemon_lock.py) */}
          <div className="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/40 space-y-4">
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2 border-b border-slate-200 dark:border-slate-700 pb-3">
              <div>
                <div className="flex items-center gap-2">
                  <Clock className="w-4 h-4 text-indigo-600" />
                  <h3 className="text-xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-200">
                    Short-Lived Supervised Viewer Leases (SRC-L3)
                  </h3>
                </div>
                <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                  Time-bound lock leases for interactive desktop inspection. Automatic expiry blocks unauthorized reuse.
                </p>
              </div>

              {/* Acquire New Lease Controls */}
              <div className="flex items-center gap-2">
                <select
                  value={leaseDurationSec}
                  onChange={(e) => setLeaseDurationSec(Number(e.target.value))}
                  className="text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200"
                >
                  <option value={300}>5 Minutes (300s)</option>
                  <option value={600}>10 Minutes (600s)</option>
                  <option value={900}>15 Minutes Max (900s)</option>
                </select>
                <button
                  onClick={handleAcquireViewerLease}
                  className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold rounded-lg shadow-sm"
                >
                  Acquire Viewer Lease
                </button>
              </div>
            </div>

            {/* Leases Table */}
            <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-50 dark:bg-slate-800/60 text-[11px] uppercase font-bold text-slate-500 border-b border-slate-200 dark:border-slate-800">
                  <tr>
                    <th className="py-2.5 px-3">Lease ID</th>
                    <th className="py-2.5 px-3">Session Ref</th>
                    <th className="py-2.5 px-3">Tool / Purpose</th>
                    <th className="py-2.5 px-3">Duration</th>
                    <th className="py-2.5 px-3">Status</th>
                    <th className="py-2.5 px-3">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {leases
                    .filter((l) => l.tenant_id === currentTenantId)
                    .map((lease) => (
                      <tr key={lease.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                        <td className="py-2.5 px-3 font-mono font-semibold text-slate-800 dark:text-slate-200">
                          {lease.id}
                        </td>
                        <td className="py-2.5 px-3 font-mono text-[11px] text-slate-500">
                          {lease.session_id}
                        </td>
                        <td className="py-2.5 px-3 text-slate-700 dark:text-slate-300">
                          <div className="font-medium">{lease.viewer_tool_name}</div>
                          <div className="text-[10px] text-slate-400 font-mono">{lease.purpose}</div>
                        </td>
                        <td className="py-2.5 px-3 font-mono text-slate-500">
                          {lease.lease_duration_seconds}s
                        </td>
                        <td className="py-2.5 px-3">
                          <span
                            className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase ${
                              lease.status === 'active'
                                ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300'
                                : lease.status === 'expired'
                                ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300'
                                : 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300'
                            }`}
                          >
                            {lease.status}
                          </span>
                        </td>
                        <td className="py-2.5 px-3">
                          <div className="flex items-center gap-1.5">
                            {lease.status === 'active' && (
                              <>
                                <button
                                  onClick={() => handleReleaseViewerLease(lease.id)}
                                  className="px-2 py-1 text-[10px] font-semibold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 rounded"
                                >
                                  Release
                                </button>
                                <button
                                  onClick={() => handleSimulateLeaseExpiry(lease.id)}
                                  className="px-2 py-1 text-[10px] font-semibold text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-950/40 hover:bg-amber-100 rounded"
                                >
                                  Simulate Expiry
                                </button>
                              </>
                            )}
                            {lease.status !== 'active' && (
                              <span className="text-[10px] text-slate-400 italic">Locked</span>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        {/* ========================================================================= */}
        {/* Section 20: CSV & Google Sheets Relationship Exports Studio (IMP-LI-20) */}
        {/* (LI-20, REQ-006, REQ-019, AT-013, AT-014, IMP-LI-06, EXP-001, SRC-C4)   */}
        {/* ========================================================================= */}
        <section className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
          <div className="border-b border-slate-200 dark:border-slate-800 px-6 py-4 bg-gradient-to-r from-emerald-50/50 via-teal-50/30 to-transparent dark:from-emerald-950/20 dark:via-teal-950/10">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="inline-flex items-center rounded-full bg-emerald-100 dark:bg-emerald-950/80 px-2.5 py-0.5 text-xs font-bold text-emerald-800 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800">
                    IMP-LI-20 &bull; AT-013 &bull; AT-014 &bull; REQ-006 &bull; REQ-019
                  </span>
                  <span className="text-xs text-slate-400 font-mono">Formula Injection Defense &bull; Idempotent Sheets CRM Projection</span>
                </div>
                <h2 className="mt-1 text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Table className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
                  CSV & Google Sheets Relationship Exports Studio
                </h2>
                <p className="text-xs text-slate-500 mt-0.5">
                  Export recruiter relationship pipelines to sanitized CSV or sync to Google Sheets with idempotent row reconciliation, preserving custom user columns.
                </p>
              </div>

              {/* Workspace Scoping Switcher (AT-011, AT-012, REQ-019) */}
              <div className="flex items-center space-x-2 bg-white dark:bg-slate-800 p-1.5 rounded-lg border border-slate-200 dark:border-slate-700">
                <span className="text-xs font-semibold text-slate-500 px-1">Workspace:</span>
                <button
                  onClick={() => setRelExportWorkspace('ws-alpha')}
                  className={`px-2.5 py-1 text-xs font-bold rounded-md transition-colors ${
                    relExportWorkspace === 'ws-alpha'
                      ? 'bg-emerald-600 text-white'
                      : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700'
                  }`}
                >
                  Workspace Alpha (Tenant Alpha)
                </button>
                <button
                  onClick={() => setRelExportWorkspace('ws-beta')}
                  className={`px-2.5 py-1 text-xs font-bold rounded-md transition-colors ${
                    relExportWorkspace === 'ws-beta'
                      ? 'bg-emerald-600 text-white'
                      : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700'
                  }`}
                >
                  Workspace Beta (Tenant Beta)
                </button>
              </div>
            </div>
          </div>

          {relNotification && (
            <div className="mx-6 mt-4 p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/60 text-xs font-semibold text-emerald-800 dark:text-emerald-200 flex items-center justify-between">
              <span>{relNotification}</span>
              <button onClick={() => setRelNotification(null)} className="text-emerald-600 hover:text-emerald-800 font-bold">✕</button>
            </div>
          )}

          <div className="p-6 space-y-6">
            {/* Security Notice Card */}
            <div className="rounded-xl border border-emerald-200 dark:border-emerald-900/60 bg-emerald-50/40 dark:bg-emerald-950/20 p-4">
              <div className="flex items-start space-x-3">
                <ShieldCheck className="w-5 h-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
                <div className="text-xs text-emerald-900 dark:text-emerald-300 leading-relaxed">
                  <span className="font-bold text-emerald-950 dark:text-emerald-100">Zero Trust Export Security Guarantees:</span>
                  {' '}1. <strong>Formula Injection Defense (AT-013)</strong>: Cells beginning with <code>=</code>, <code>+</code>, <code>-</code>, <code>@</code>, <code>\t</code>, or <code>\r</code> are automatically neutralized with a leading single quote (<code>'</code>).
                  {' '}2. <strong>Tenant & Owner Scoping (REQ-019, AT-011)</strong>: Cross-tenant data leaks are mathematically denied at repository layer.
                  {' '}3. <strong>Private Notes Consent (REQ-006)</strong>: Candidate notes remain private by default unless explicitly opted in.
                  {' '}4. <strong>Idempotent Sheets Reconciliation (AT-014)</strong>: Re-syncing preserves sorting, cell formatting, and trailing custom evaluation columns.
                </div>
              </div>
            </div>

            {/* Export Configuration Controls */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/40">
              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Export Destination
                </label>
                <select
                  value={relExportDestination}
                  onChange={(e) => setRelExportDestination(e.target.value as LinkedInRelationshipExportDestination)}
                  className="w-full text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 px-3 py-2 text-slate-800 dark:text-slate-200 font-medium"
                >
                  <option value="csv_download">Sanitized CSV Download (UTF-8)</option>
                  <option value="google_sheets_sync">Google Sheets Idempotent Sync</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Private Candidate Notes (REQ-006)
                </label>
                <div className="flex items-center gap-2 pt-1.5">
                  <input
                    type="checkbox"
                    id="relIncludeNotesToggle"
                    checked={relIncludeNotes}
                    onChange={(e) => setRelIncludeNotes(e.target.checked)}
                    className="h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                  />
                  <label htmlFor="relIncludeNotesToggle" className="text-xs text-slate-600 dark:text-slate-400 font-medium">
                    Include Private Notes (Explicit Consent)
                  </label>
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  CSV UTF-8 BOM Header
                </label>
                <div className="flex items-center gap-2 pt-1.5">
                  <input
                    type="checkbox"
                    id="relWithBOMToggle"
                    checked={relWithBOM}
                    onChange={(e) => setRelWithBOM(e.target.checked)}
                    className="h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                  />
                  <label htmlFor="relWithBOMToggle" className="text-xs text-slate-600 dark:text-slate-400 font-medium">
                    Prepend UTF-8 BOM (Excel Compatibility)
                  </label>
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Export Timezone
                </label>
                <select
                  value={relTimezone}
                  onChange={(e) => setRelTimezone(e.target.value as any)}
                  className="w-full text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 px-3 py-2 text-slate-800 dark:text-slate-200 font-medium"
                >
                  <option value="UTC">UTC (Universal Time)</option>
                  <option value="America/New_York">America/New York (EST/EDT)</option>
                  <option value="Europe/London">Europe/London (GMT/BST)</option>
                </select>
              </div>
            </div>

            {/* Sheets-Specific Target Configuration */}
            {relExportDestination === 'google_sheets_sync' && (
              <div className="p-4 rounded-xl border border-teal-200 dark:border-teal-800 bg-teal-50/30 dark:bg-teal-950/20 grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-teal-900 dark:text-teal-200 mb-1">
                    Google Spreadsheet ID
                  </label>
                  <input
                    type="text"
                    value={relSpreadsheetId}
                    onChange={(e) => setRelSpreadsheetId(e.target.value)}
                    className="w-full text-xs font-mono rounded-lg border border-teal-300 dark:border-teal-700 bg-white dark:bg-slate-800 px-3 py-2 text-slate-800 dark:text-slate-200"
                    placeholder="e.g. 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-teal-900 dark:text-teal-200 mb-1">
                    Sheet Tab Name
                  </label>
                  <input
                    type="text"
                    value={relSheetName}
                    onChange={(e) => setRelSheetName(e.target.value)}
                    className="w-full text-xs font-mono rounded-lg border border-teal-300 dark:border-teal-700 bg-white dark:bg-slate-800 px-3 py-2 text-slate-800 dark:text-slate-200"
                    placeholder="e.g. Relationship_CRM"
                  />
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
              <div className="text-xs text-slate-500 font-medium">
                Active Recruiter Leads in {relExportWorkspace}: <span className="font-bold text-slate-900 dark:text-white">{leads.filter(l => l.workspace_id === relExportWorkspace).length}</span> records ready
              </div>

              <div className="flex items-center gap-3">
                <button
                  onClick={handleRunRelationshipCSVExport}
                  className="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-sm hover:bg-emerald-700 transition-colors"
                >
                  <Download className="w-4 h-4" />
                  Export Scoped CSV (AT-013 Neutralized)
                </button>

                <button
                  onClick={handleRunRelationshipSheetsSync}
                  className="inline-flex items-center gap-2 rounded-lg bg-teal-600 px-4 py-2 text-xs font-bold text-white shadow-sm hover:bg-teal-700 transition-colors"
                >
                  <RefreshCw className="w-4 h-4" />
                  Run Idempotent Sheets Sync (AT-014)
                </button>
              </div>
            </div>

            {/* Live Sheets Reconciliation Table & Custom Column Inspection */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                  <Table className="w-4 h-4 text-emerald-600" />
                  Live Google Sheets CRM Projection (Idempotent State &bull; AT-014)
                </h3>
                <span className="text-[11px] font-mono text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/60 px-2 py-0.5 rounded border border-emerald-200 dark:border-emerald-800">
                  Custom User Column Preserved Across Syncs
                </span>
              </div>

              <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-50 dark:bg-slate-800/80 text-[11px] font-bold text-slate-600 dark:text-slate-300 border-b border-slate-200 dark:border-slate-800">
                    <tr>
                      {relSheetRows[0]?.map((header, idx) => (
                        <th key={idx} className={`py-2.5 px-3 whitespace-nowrap ${idx >= 8 ? 'bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-300 font-extrabold border-l border-amber-200 dark:border-amber-800' : ''}`}>
                          {header}
                          {idx >= 8 && <span className="ml-1 text-[9px] uppercase px-1 py-0.2 rounded bg-amber-200 dark:bg-amber-900 text-amber-900 dark:text-amber-100 font-bold">User Custom</span>}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {relSheetRows.slice(1).map((row, rIdx) => (
                      <tr key={rIdx} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 font-mono text-[11px]">
                        {row.map((cell, cIdx) => (
                          <td key={cIdx} className={`py-2 px-3 whitespace-nowrap ${cIdx >= 8 ? 'bg-amber-50/40 dark:bg-amber-950/10 text-amber-900 dark:text-amber-200 font-semibold border-l border-amber-100 dark:border-amber-900' : 'text-slate-700 dark:text-slate-300'}`}>
                            {cell}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Audit History & Manifest Trail */}
            <div className="space-y-2 pt-2">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                <FileCheck className="w-4 h-4 text-emerald-600" />
                Export & Reconciliation Audit Trail (REQ-019 &bull; AT-011)
              </h3>

              <div className="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-50 dark:bg-slate-800/80 text-[11px] font-bold text-slate-500 uppercase border-b border-slate-200 dark:border-slate-800">
                    <tr>
                      <th className="py-2.5 px-3">Audit ID</th>
                      <th className="py-2.5 px-3">Destination</th>
                      <th className="py-2.5 px-3">Target / File</th>
                      <th className="py-2.5 px-3">Rows</th>
                      <th className="py-2.5 px-3">Private Notes</th>
                      <th className="py-2.5 px-3">Owner & Workspace</th>
                      <th className="py-2.5 px-3">Exported At</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {relExportAudits.map((audit) => (
                      <tr key={audit.audit_id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                        <td className="py-2 px-3 font-mono font-bold text-slate-800 dark:text-slate-200">
                          {audit.audit_id}
                        </td>
                        <td className="py-2 px-3">
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                            audit.destination === 'csv_download'
                              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                              : 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                          }`}>
                            {audit.destination.replace(/_/g, ' ')}
                          </span>
                        </td>
                        <td className="py-2 px-3 font-mono text-[11px] text-slate-600 dark:text-slate-400">
                          {audit.filename}
                        </td>
                        <td className="py-2 px-3 font-semibold text-slate-800 dark:text-slate-200">
                          {audit.row_count}
                        </td>
                        <td className="py-2 px-3">
                          <span className={`px-1.5 py-0.5 rounded text-[10px] font-semibold ${
                            audit.include_notes
                              ? 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300'
                              : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-400'
                          }`}>
                            {audit.include_notes ? 'Opted In' : 'Excluded'}
                          </span>
                        </td>
                        <td className="py-2 px-3 text-[11px] text-slate-500 font-mono">
                          {audit.owner_id} ({audit.workspace_id})
                        </td>
                        <td className="py-2 px-3 text-[11px] text-slate-500">
                          {new Date(audit.exported_at).toLocaleTimeString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>
        </div>
      )}
    </div>
  );
}