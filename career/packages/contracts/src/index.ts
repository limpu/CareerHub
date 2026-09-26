// Task & Queue Envelope Contracts
export type PlatformSection = 'career' | 'linkedin' | 'social';

export type UserRole = 'super_admin' | 'admin' | 'user';

export interface BaseTaskEnvelope<T = unknown> {
  id: string;
  section: PlatformSection;
  taskType: string;
  actorId: string;
  workspaceId: string;
  payload: T;
  createdAt: string;
  correlationId: string;
}

export interface DocumentTaskPayload {
  documentId: string;
  sourceUrl: string;
  mimeType: 'application/pdf' | 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
  targetFormat?: 'text' | 'markdown' | 'json';
}

export interface TaskResult<T = unknown> {
  taskId: string;
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    retryable: boolean;
  };
  completedAt: string;
}

// Career Profile Contracts (IMP-CAR-01, IMP-CAR-02)
export type PrivacyTier = 'public' | 'delegated' | 'owner_private';
export type SkillProficiency = 'unspecified' | 'beginner' | 'intermediate' | 'advanced' | 'expert';
export type LanguageProficiency = 'basic' | 'conversational' | 'professional' | 'fluent' | 'native';

export interface ContactInfo {
  full_name: string;
  email: string;
  phone?: string;
  location?: string;
  headline?: string;
  summary?: string;
  citizenship?: string;
  work_authorization?: string;
}

export interface ProfileLink {
  id: string;
  label: string;
  url: string;
  link_type: 'github' | 'linkedin' | 'portfolio' | 'blog' | 'other';
  confirmed: boolean;
}

export interface ExperienceItem {
  id: string;
  title: string;
  company: string;
  location?: string;
  start_date?: string;
  end_date?: string;
  is_current: boolean;
  description?: string;
  highlights: string[];
  skills_used: string[];
  confirmed: boolean;
}

export interface EducationItem {
  id: string;
  institution: string;
  degree: string;
  field_of_study?: string;
  start_date?: string;
  end_date?: string;
  grade?: string;
  highlights: string[];
  confirmed: boolean;
}

export interface SkillItem {
  id: string;
  name: string;
  category: 'technical' | 'soft' | 'tool' | 'domain';
  proficiency: SkillProficiency;
  years_of_experience?: number;
  confirmed: boolean;
}

export interface ProjectItem {
  id: string;
  title: string;
  role?: string;
  url?: string;
  start_date?: string;
  end_date?: string;
  description: string;
  highlights: string[];
  technologies: string[];
  confirmed: boolean;
}

export interface CertificateItem {
  id: string;
  name: string;
  issuer: string;
  issue_date?: string;
  expiry_date?: string;
  credential_id?: string;
  credential_url?: string;
  confirmed: boolean;
}

export interface LanguageItem {
  id: string;
  language: string;
  proficiency: LanguageProficiency;
  confirmed: boolean;
}

export interface UserCareerConsent {
  data_processing_consent: boolean;
  job_matching_consent: boolean;
  third_party_sharing_consent: boolean;
  consent_timestamp: string;
  consent_version: string;
}

export interface CareerPrivacySettings {
  profile_visibility: PrivacyTier;
  allow_recruiter_view: boolean;
  share_with_workspace_admin: boolean;
  anonymize_before_job_search: boolean;
}

export interface ProfileFieldAudit {
  id: string;
  field: string;
  old_value?: string;
  new_value?: string;
  changed_by: string;
  changed_at: string;
  reason?: string;
}

export interface CompletenessReport {
  score: number;
  missing_sections: string[];
  populated_count: number;
  total_sections_count: number;
}

export interface MasterCareerProfile {
  id: string;
  user_id: string;
  contact: ContactInfo;
  links: ProfileLink[];
  experiences: ExperienceItem[];
  education: EducationItem[];
  skills: SkillItem[];
  projects: ProjectItem[];
  certificates: CertificateItem[];
  languages: LanguageItem[];
  consent: UserCareerConsent;
  privacy: CareerPrivacySettings;
  audit_trail: ProfileFieldAudit[];
  completeness: CompletenessReport;
  created_at: string;
  updated_at: string;
}

// Career Preferences Contracts (IMP-CAR-03, REQ-003, AT-003)
export type WorkMode = 'remote' | 'hybrid' | 'onsite';
export type JobType = 'full_time' | 'part_time' | 'contract' | 'internship' | 'temporary' | 'other';
export type SponsorshipPreference = 'unspecified' | 'requires_sponsorship' | 'authorized_no_sponsorship' | 'will_not_sponsor_accepted';
export type SalaryInterval = 'annual' | 'monthly' | 'hourly';

export interface SalaryPreference {
  minimum_amount: number;
  target_amount?: number;
  maximum_amount?: number;
  currency: string;
  interval: SalaryInterval;
  is_explicitly_set: boolean;
}

export interface ExclusionRules {
  excluded_companies: string[];
  excluded_keywords: string[];
  excluded_industries: string[];
  excluded_locations: string[];
  block_staffing_agencies: boolean;
}

export interface CareerPreferences {
  id: string;
  user_id: string;
  target_roles: string[];
  target_locations: string[];
  work_modes: WorkMode[];
  job_types: JobType[];
  salary: SalaryPreference;
  sponsorship: SponsorshipPreference;
  exclusions: ExclusionRules;
  open_to_relocation: boolean;
  notice_period_days: number;
  created_at: string;
  updated_at: string;
}

// Master Resume Generation & Parse-Back Contracts (IMP-CAR-04, CAR-04, AT-002)
export type MasterResumeFormat = 'pdf' | 'docx' | 'txt';
export type ResumeTemplateType = 'single_column_modern' | 'single_column_classic' | 'single_column_minimal';

export interface GeneratedResumeSummary {
  id: string;
  user_id: string;
  format: MasterResumeFormat;
  template: ResumeTemplateType;
  file_name: string;
  mime_type: string;
  content_length: number;
  checksum_sha256: string;
  deterministic_section_order: string[];
  parse_back_verified: boolean;
  created_at: string;
}

export interface GenerateResumeRequest {
  format?: MasterResumeFormat;
  template?: ResumeTemplateType;
}

export interface ParseBackVerificationResult {
  success: boolean;
  format: string;
  extracted_words: number;
  matched_facts_count: number;
  missing_facts?: string[];
  warnings?: string[];
  verified_at: string;
}

// Job-Specific Tailored Documents & Approval Gatekeeper Contracts (IMP-CAR-05, REQ-016, AT-003, FND-010, AT-007)
export type ApprovalStatus = 'pending_approval' | 'approved' | 'rejected';

export interface JobTarget {
  id: string;
  title: string;
  company: string;
  location?: string;
  url?: string;
  description?: string;
  required_skills: string[];
  keywords?: string[];
}

export interface ResumeDiffSummary {
  emphasized_skills: string[];
  prioritized_highlights: string[];
  unmatched_job_requirements: string[]; // Disclosed missing requirements per REQ-016
  total_confirmed_facts_used: number;
}

export interface TailoredResumeSummary {
  id: string;
  user_id: string;
  job_id: string;
  job_target: JobTarget;
  format: MasterResumeFormat;
  template: ResumeTemplateType;
  original_profile_id: string;
  file_name: string;
  mime_type: string;
  content_length: number;
  checksum_sha256: string;
  diff_summary: ResumeDiffSummary;
  approval_status: ApprovalStatus;
  approval_token: string;
  approved_at?: string;
  parse_back_verified: boolean;
  created_at: string;
}

export interface CoverLetterSummary {
  id: string;
  user_id: string;
  job_id: string;
  job_target: JobTarget;
  recipient_name: string;
  format: MasterResumeFormat;
  salutation: string;
  opening_paragraph: string;
  body_paragraphs: string[];
  closing_paragraph: string;
  signoff: string;
  full_text: string;
  file_name: string;
  mime_type: string;
  content_length: number;
  checksum_sha256: string;
  approval_status: ApprovalStatus;
  approval_token: string;
  approved_at?: string;
  parse_back_verified: boolean;
  created_at: string;
}

export interface TailorResumeRequest {
  job: JobTarget;
  format?: MasterResumeFormat;
  template?: ResumeTemplateType;
}

export interface ApproveTailoredResumeRequest {
  resume_id: string;
  approval_token: string;
}

export interface CreateCoverLetterRequest {
  job: JobTarget;
  format?: MasterResumeFormat;
  recipient_name?: string;
}

export interface ApproveCoverLetterRequest {
  letter_id: string;
  approval_token: string;
}

// Resume Feedback & Skills Demand Contracts (IMP-CAR-06, CAR-06, AT-028)
export type FeedbackSeverity = 'critical' | 'warning' | 'info';

export interface FeedbackIssue {
  section: string;
  severity: FeedbackSeverity;
  message: string;
  recommendation: string;
  affected_item?: string;
}

export interface FeedbackSuggestion {
  category: string;
  current_text: string;
  suggested_improvement: string;
  rationale: string;
}

export interface ResumeFeedbackReport {
  id: string;
  user_id: string;
  profile_id: string;
  overall_score: number;
  ats_readability_score: number;
  impact_score: number;
  quantification_rate: number;
  action_verb_density: number;
  section_scores: Record<string, number>;
  strengths: string[];
  critical_issues: FeedbackIssue[];
  improvement_suggestions: FeedbackSuggestion[];
  analyzed_at: string;
}

export interface MarketSkillDemandItem {
  skill_name: string;
  category: string;
  demand_level: 'critical' | 'high' | 'moderate' | 'niche';
  demand_percentile: number;
  growth_yoy: number;
  evidence_date: string;
  sample_size: number;
  source_citation: string;
  is_possessed: boolean;
  status: 'possessed_and_in_demand' | 'market_demand_gap';
}

export interface SkillsDemandAnalysis {
  role_category: string;
  total_market_skills_analyzed: number;
  candidate_possessed_count: number;
  candidate_gap_count: number;
  possessed_skills: MarketSkillDemandItem[];
  market_gaps: MarketSkillDemandItem[];
  analysis_date: string;
  data_availability_status: string;
}

// Multi-Board Job Discovery Contracts (IMP-CAR-07, CAR-07, AT-004, AT-010)
export type BoardSource = 'greenhouse' | 'lever' | 'ashby' | 'linkedin' | 'indeed' | 'google';
export type CompensationPeriod = 'hourly' | 'monthly' | 'yearly';

export interface NormalizedCompensation {
  is_disclosed: boolean;
  min_amount?: number;
  max_amount?: number;
  currency?: string;
  period?: CompensationPeriod;
  is_estimated?: boolean;
}

export interface JobLocation {
  raw_location: string;
  city?: string;
  state?: string;
  country?: string;
  is_remote: boolean;
  is_hybrid: boolean;
}

export interface DiscoveredJob {
  id: string;
  source: BoardSource;
  source_job_id: string;
  canonical_url: string;
  direct_apply_url?: string;
  title: string;
  company: string;
  company_logo_url?: string;
  location: JobLocation;
  job_type: string;
  description: string;
  required_skills: string[];
  compensation: NormalizedCompensation;
  date_posted: string;
  discovered_at: string;
  is_direct_employer: boolean;
}

export interface BoardCapabilities {
  board_name: BoardSource;
  display_name: string;
  is_direct_ats: boolean;
  supports_keyword_search: boolean;
  supports_location_filter: boolean;
  supports_remote_filter: boolean;
  supports_salary_filter: boolean;
  requires_authentication: boolean;
  rate_limit_per_minute: number;
  notes: string;
}

export interface JobSearchQuery {
  keywords?: string;
  location?: string;
  remote_only?: boolean;
  job_types?: string[];
  sources?: BoardSource[];
  limit_per_board?: number;
}

// Fine-Grained Job Filters & Audit Contracts (IMP-CAR-08, CAR-08, AT-004)
export type FilterExecutionMode = 'upstream_native' | 'post_filtered' | 'unsupported';

export interface AdvancedJobFilter {
  max_age_days?: number;
  work_modes?: ('remote' | 'hybrid' | 'on_site')[];
  job_types?: JobType[];
  company_inclusions?: string[];
  company_exclusions?: string[];
  title_keywords_must_include?: string[];
  title_keywords_must_exclude?: string[];
  minimum_annual_salary?: number;
  salary_currency?: string;
  include_undisclosed_salary: boolean;
}

export interface FilterAuditReport {
  total_candidates_input: number;
  total_results_retained: number;
  disqualified_by_age: number;
  disqualified_by_mode: number;
  disqualified_by_type: number;
  disqualified_by_company: number;
  disqualified_by_salary: number;
  filter_evaluations: Record<string, FilterExecutionMode>;
}

export interface FilterJobsRequest {
  jobs: DiscoveredJob[];
  filter: AdvancedJobFilter;
}

export interface FilterJobsResponse {
  filtered_jobs: DiscoveredJob[];
  audit_report: FilterAuditReport;
}

// Explainable Job Match Contracts (IMP-CAR-09, CAR-09, AT-003, AT-028)
export type MatchStatus = 'matched' | 'mismatch' | 'unknown';
export type HardGateType = 'company_exclusion' | 'work_authorization' | 'work_mode_relocation' | 'minimum_experience';
export type FitTier = 'strong_match' | 'moderate_match' | 'low_match' | 'ineligible';

export interface HardGateResult {
  gate_type: HardGateType;
  gate_name: string;
  passed: boolean;
  is_unknown: boolean;
  reason: string;
}

export interface ScoringWeights {
  skills_weight: number;
  title_experience_weight: number;
  location_work_mode_weight: number;
  compensation_weight: number;
}

export interface RequirementMatch {
  category: string;
  requirement: string;
  status: MatchStatus;
  candidate_evidence: string;
  confidence: number;
}

export interface MatchScoreBreakdown {
  skills_score: number;
  skills_weighted_score: number;
  title_experience_score: number;
  title_experience_weighted_score: number;
  location_work_mode_score: number;
  location_work_mode_weighted_score: number;
  compensation_score: number;
  compensation_weighted_score: number;
  weights_used: ScoringWeights;
}

export interface JobMatchResult {
  job_id: string;
  algorithm_version: string;
  evaluated_at: string;
  is_eligible: boolean;
  overall_score: number;
  fit_tier: FitTier;
  hard_gates: HardGateResult[];
  breakdown: MatchScoreBreakdown;
  matched_skills: string[];
  missing_skills: string[];
  unknown_requirements: string[];
  explanations: string[];
  requirement_matches: RequirementMatch[];
}

export interface EvaluateJobMatchRequest {
  job: DiscoveredJob;
  weights?: ScoringWeights;
}

export interface EvaluateBatchJobMatchesRequest {
  jobs: DiscoveredJob[];
  weights?: ScoringWeights;
}

export interface EvaluateBatchJobMatchesResponse {
  matches: JobMatchResult[];
  total: number;
}

// Saved Jobs, Shortlists & Multi-Ledger Dedupe (IMP-CAR-10, CAR-10, AT-004, C8)
export type SavedJobStatus = 'saved' | 'shortlisted' | 'ready_to_apply' | 'applied' | 'archived';

export interface SavedJob {
  id: string;
  user_id: string;
  job_id: string;
  canonical_url: string;
  fingerprint: string;
  title: string;
  company: string;
  location: JobLocation;
  job_type: string;
  compensation: NormalizedCompensation;
  required_skills: string[];
  direct_apply_url?: string;
  status: SavedJobStatus;
  notes: string;
  priority: number;
  tags: string[];
  saved_at: string;
  updated_at: string;
  applied_at?: string;
}

export type ExclusionType = 'job_canonical_url' | 'job_fingerprint' | 'company_name';

export interface JobExclusion {
  id: string;
  user_id: string;
  type: ExclusionType;
  value: string;
  reason: string;
  job_title?: string;
  company_name?: string;
  created_at: string;
}

export interface ApplicationRecord {
  id: string;
  user_id: string;
  job_id: string;
  canonical_url: string;
  fingerprint?: string;
  title: string;
  company: string;
  status: string;
  stage?: ApplicationStage;
  submission_mode?: string;
  portal_type?: string;
  is_verified_applied?: boolean;
  applied_at: string;
  created_at?: string;
  updated_at?: string;
  notes?: string;
  verification_details?: AppliedVerificationDetails;
  interview_rounds?: InterviewRound[];
  timeline?: ApplicationTimelineEvent[];
  metadata?: Record<string, any>;
}

export type DedupeMatchLevel = 'exact' | 'ambiguous' | 'none';

export interface DedupeMatch {
  match_level: DedupeMatchLevel;
  target_ledger: string;
  existing_id: string;
  existing_title: string;
  existing_company: string;
  similarity_score: number;
  reason: string;
  is_reviewable: boolean;
}

export interface JobDedupeResult {
  job_id: string;
  canonical_url: string;
  fingerprint: string;
  matches: DedupeMatch[];
  is_duplicate: boolean;
  is_ambiguous: boolean;
  is_excluded: boolean;
  action_taken: 'retained' | 'linked' | 'flagged_for_review' | 'filtered_out';
}

export interface SaveJobRequest {
  job: DiscoveredJob;
  status?: SavedJobStatus;
  notes?: string;
  priority?: number;
  tags?: string[];
}

export interface UpdateSavedJobRequest {
  status?: SavedJobStatus;
  notes?: string;
  priority?: number;
  tags?: string[];
  applied_at?: string;
}

export interface AddExclusionRequest {
  type: ExclusionType;
  value: string;
  reason: string;
  job_title?: string;
  company_name?: string;
}

export interface RecordApplicationRequest {
  job: DiscoveredJob;
  status: string;
  submission_mode: string;
  notes: string;
  applied_at?: string;
}

export interface CheckDedupeRequest {
  jobs: DiscoveredJob[];
}

export interface CheckDedupeResponse {
  results: JobDedupeResult[];
  total: number;
}



// Human-Reviewed Application Workflow (IMP-CAR-11, CAR-11, REQ-005, REQ-015, AT-003, AT-005, AT-007)
export type ApplicationWorkflowStatus =
  | 'draft'
  | 'ready_for_review'
  | 'reviewing'
  | 'approved'
  | 'dispatched'
  | 'needs_confirmation'
  | 'applied'
  | 'cancelled';

export type ApplicationExecutionMode = 'native_manual' | 'assistant_prefill';

export interface ApplicationQuestionAnswer {
  question_id: string;
  question_text: string;
  field_type: string;
  options?: string[];
  required: boolean;
  answer_value: string;
  is_confirmed: boolean;
  needs_input: boolean;
  category: string;
  explanation_note?: string;
}

export interface ApplicationMaterialBundle {
  resume_id: string;
  resume_checksum: string;
  resume_file_name: string;
  cover_letter_id?: string;
  cover_letter_checksum?: string;
  cover_letter_file_name?: string;
  questions: ApplicationQuestionAnswer[];
  bundle_checksum_sha256: string;
}

export interface SubmissionReceipt {
  receipt_id: string;
  provider_reference?: string;
  submission_url: string;
  confirmed_at: string;
  confirmed_by_user: boolean;
  notes?: string;
}

export interface ApplicationReviewSession {
  id: string;
  user_id: string;
  workspace_id: string;
  job_id: string;
  job_title: string;
  company: string;
  apply_url: string;
  source_board: string;
  execution_mode: ApplicationExecutionMode;
  bundle: ApplicationMaterialBundle;
  status: ApplicationWorkflowStatus;
  approval_token?: string;
  approved_at?: string;
  approver_id?: string;
  dispatched_at?: string;
  timeout_duration_secs: number;
  receipt?: SubmissionReceipt;
  application_record_id?: string;
  created_at: string;
  updated_at: string;
}

export interface PrepareApplicationRequest {
  job: DiscoveredJob;
  resume_id?: string;
  cover_letter_id?: string;
  execution_mode?: ApplicationExecutionMode;
}

export interface UpdateWorkflowAnswersRequest {
  session_id: string;
  answers: ApplicationQuestionAnswer[];
}

export interface ApproveWorkflowRequest {
  session_id: string;
}

export interface DispatchWorkflowRequest {
  session_id: string;
  execution_mode?: ApplicationExecutionMode;
}

export interface ConfirmSubmissionRequest {
  session_id: string;
  receipt: SubmissionReceipt;
}

// Question & Field Recognition Engine (IMP-CAR-12, CAR-12, AT-003, AT-019, SRC-C3, SRC-C4, SRC-C6)
export type FieldType = 'text' | 'textarea' | 'number' | 'select' | 'radio' | 'checkbox' | 'file';

export type FieldCategory =
  | 'contact'
  | 'experience'
  | 'education'
  | 'skills'
  | 'work_authorization'
  | 'sponsorship'
  | 'salary'
  | 'notice_period'
  | 'portfolio'
  | 'demographic'
  | 'custom';

export type ConditionOperator = 'equals' | 'not_equals' | 'contains' | 'in';
export type ConditionAction = 'show' | 'hide' | 'enable' | 'disable' | 'require';
export type ConditionState = 'active' | 'hidden' | 'disabled';

export interface FieldCondition {
  parent_field_id: string;
  operator: ConditionOperator;
  expected_value: any;
  action: ConditionAction;
}

export interface FieldValidation {
  min_length?: number;
  max_length?: number;
  min_value?: number;
  max_value?: number;
  pattern?: string;
  allowed_extensions?: string[];
}

export interface FormField {
  field_id: string;
  label: string;
  sanitized_label?: string;
  type: FieldType;
  required: boolean;
  placeholder?: string;
  help_text?: string;
  options?: string[];
  category: FieldCategory;
  target_sub_key?: string;
  is_sensitive?: boolean;
  validation?: FieldValidation;
  condition?: FieldCondition;
  security_flags?: string[];
}

export interface FormSection {
  section_id: string;
  title: string;
  description?: string;
  fields: FormField[];
}

export interface RecognizedForm {
  form_id: string;
  title: string;
  source_url?: string;
  provider?: string;
  sections?: FormSection[];
  fields?: FormField[];
}

export interface FieldResolution {
  field_id: string;
  label: string;
  type: FieldType;
  category: FieldCategory;
  resolved_value?: any;
  needs_input: boolean;
  confidence: number;
  provenance: string;
  resolution_reason: string;
  condition_state: ConditionState;
  security_flags?: string[];
}

export interface RecognizeFieldsRequest {
  form_id: string;
  title?: string;
  provider?: string;
  fields?: FormField[];
  sections?: FormSection[];
}

export interface RecognizeFieldsResponse {
  form_id: string;
  total_fields: number;
  recognized_form: RecognizedForm;
  security_alerts?: string[];
}

export interface ResolveAnswersRequest {
  form: RecognizedForm;
  current_answers?: Record<string, any>;
}

export interface ResolveAnswersResponse {
  form_id: string;
  resolutions: FieldResolution[];
  needs_input_count: number;
  resolved_count: number;
  zero_fabrication: boolean;
}

export interface EvaluateConditionsRequest {
  fields: FormField[];
  answers: Record<string, any>;
}

export interface EvaluateConditionsResponse {
  states: Record<string, ConditionState>;
}

// ==========================================
// Multi-step Wizard, Shadow DOM & Loop Detection Contracts (IMP-CAR-13, CAR-13, AT-003, AT-006, FND-009, SRC-C4)
// ==========================================

export type FormStepStatus = 'pending' | 'active' | 'completed' | 'blocked' | 'skipped';

export interface FormStep {
  step_id: string;
  step_number: number;
  title: string;
  description?: string;
  fields: FormField[];
  shadow_host_selector?: string;
  status: FormStepStatus;
  is_pierced: boolean;
  pierced_selector?: string;
  step_fingerprint?: string;
  scrubbed_evidence?: string;
}

export interface FormStateMachine {
  wizard_id: string;
  title: string;
  provider: string;
  steps: FormStep[];
  current_step_index: number;
  status: FormStepStatus;
  step_history: string[];
  halt_reason?: string;
  max_loop_threshold: number;
}

export interface InitWizardRequest {
  wizard_id: string;
  title: string;
  provider: string;
  steps: FormStep[];
}

export interface InitWizardResponse {
  state_machine: FormStateMachine;
  current_step: FormStep;
}

export interface AdvanceWizardStepRequest {
  state_machine: FormStateMachine;
  answers: Record<string, any>;
  raw_evidence?: string;
}

export interface AdvanceWizardStepResponse {
  state_machine: FormStateMachine;
  active_step?: FormStep;
  is_completed: boolean;
  loop_detected: boolean;
  halt_reason?: string;
}

export interface PreviousWizardStepRequest {
  state_machine: FormStateMachine;
}

export interface PreviousWizardStepResponse {
  state_machine: FormStateMachine;
  active_step: FormStep;
}

export interface TraverseShadowDOMRequest {
  raw_html: string;
  shadow_host_selector: string;
}

export interface TraverseShadowDOMResponse {
  pierced_selector: string;
  fields: FormField[];
}

// ==========================================
// External & Indeed Applications Contracts (IMP-CAR-14, CAR-14, AT-005, AT-006, AT-010, SRC-C3)
// ==========================================

export type PortalType =
  | 'indeed_easy_apply'
  | 'linkedin_easy_apply'
  | 'greenhouse_direct'
  | 'lever_direct'
  | 'ashby_direct'
  | 'workday_external'
  | 'taleo_external'
  | 'successfactors_external'
  | 'generic_external';

export type PortalSupportLevel = 'native_easy_apply' | 'assisted_manual';

export interface PortalCapability {
  can_auto_fill: boolean;
  requires_external_redirect: boolean;
  manual_fallback_mandatory: boolean;
  description: string;
  known_limitations?: string;
}

export interface ExternalPortalInfo {
  url: string;
  portal_type: PortalType;
  support_level: PortalSupportLevel;
  display_name: string;
  capabilities: PortalCapability;
  truth_advertising_disclosure: string;
  redirect_target_url?: string;
}

export interface ClipboardItem {
  key: string;
  label: string;
  value: string;
  category: string;
  is_pii: boolean;
}

export interface ExternalApplicationBundle {
  bundle_id: string;
  session_id: string;
  job_id: string;
  job_title: string;
  company_name: string;
  portal_info: ExternalPortalInfo;
  clipboard_items: ClipboardItem[];
  formatted_clipboard_text: string;
  resume_document_id?: string;
  cover_letter_text?: string;
  status: 'prepared' | 'dispatched' | 'applied' | 'dismissed';
  dispatched_at?: string;
  applied_at?: string;
  notes?: string;
}

export interface ClassifyPortalRequest {
  url: string;
  redirect_target_url?: string;
}

export interface ClassifyPortalResponse {
  portal_info: ExternalPortalInfo;
}

export interface PrepareExternalBundleRequest {
  job_id: string;
  session_id?: string;
  portal_url: string;
  redirect_target_url?: string;
  custom_notes?: string;
}

export interface PrepareExternalBundleResponse {
  bundle: ExternalApplicationBundle;
}

export interface DispatchExternalApplicationRequest {
  bundle_id: string;
  session_id?: string;
}

export interface ConfirmExternalApplicationRequest {
  bundle_id: string;
  confirmation_notes?: string;
  confirmation_receipt?: string;
}

export interface ConfirmExternalApplicationResponse {
  bundle: ExternalApplicationBundle;
  application_record_id: string;
}

// ==========================================
// Application Status, Verified Applied & Reconciliation (IMP-CAR-15, CAR-15, REQ-005, REQ-016, AT-005, AT-006)
// ==========================================

export type ApplicationStage =
  | 'interested'
  | 'preparing'
  | 'dispatched'
  | 'needs_confirmation'
  | 'applied'
  | 'interviewing'
  | 'offered'
  | 'rejected'
  | 'withdrawn'
  | 'archived';

export type AppliedVerificationType =
  | 'provider_receipt'
  | 'user_attestation'
  | 'email_confirmation';

export interface AppliedVerificationDetails {
  verified_at: string;
  verification_type: AppliedVerificationType;
  receipt_id: string;
  provider_reference?: string;
  attestation_notes?: string;
  verifier_id: string;
}

export interface InterviewRound {
  round_id: string;
  stage_name: string;
  scheduled_at: string;
  completed_at?: string;
  interviewer?: string;
  notes?: string;
  outcome?: 'passed' | 'pending' | 'failed';
}

export interface ApplicationTimelineEvent {
  event_id: string;
  application_record_id: string;
  timestamp: string;
  from_stage: ApplicationStage;
  to_stage: ApplicationStage;
  trigger: string;
  actor_id: string;
  notes?: string;
  verification?: AppliedVerificationDetails;
}

export type ReconciliationAction = 'confirm_applied' | 'mark_abandoned' | 'retry_dispatch';

export type EnrichedApplicationRecord = ApplicationRecord;

export interface VerifyAppliedMarkRequest {
  application_record_id: string;
  verification_type: AppliedVerificationType;
  receipt_id: string;
  provider_reference?: string;
  attestation_notes?: string;
}

export interface VerifyAppliedMarkResponse {
  record: EnrichedApplicationRecord;
  timeline_event: ApplicationTimelineEvent;
}

export interface ReconcileApplicationRequest {
  session_id: string;
  action: ReconciliationAction;
  receipt?: {
    receipt_id: string;
    provider_reference?: string;
    submission_url?: string;
    confirmed_at?: string;
    confirmed_by_user?: boolean;
    notes?: string;
  };
  notes?: string;
}

export interface ReconcileApplicationResponse {
  session: any;
  record?: EnrichedApplicationRecord;
  timeline_event: ApplicationTimelineEvent;
}

export interface UpdateApplicationStageRequest {
  application_record_id: string;
  new_stage: ApplicationStage;
  notes?: string;
}

export interface UpdateApplicationStageResponse {
  record: EnrichedApplicationRecord;
  timeline_event: ApplicationTimelineEvent;
}

export interface GetApplicationAuditLedgerResponse {
  record: EnrichedApplicationRecord;
  timeline: ApplicationTimelineEvent[];
}

// --- Per-Application Evidence / Document Bundle Types (IMP-CAR-16, CAR-16, AT-005, AT-011, SRC-C2, SRC-C3, SRC-C4) ---

export interface JobSnapshot {
  title: string;
  company: string;
  location: string;
  job_type: string;
  required_skills: string[];
  direct_apply_url: string;
  description_snippet: string;
  captured_at: string;
}

export interface ResumeArtifactSnapshot {
  resume_id: string;
  version: string;
  format: string;
  content_checksum: string;
  document_path: string;
  content_snippet?: string;
}

export interface CoverLetterSnapshot {
  cover_letter_id: string;
  title: string;
  content_checksum: string;
  document_path: string;
  content_snippet?: string;
}

export interface SubmittedAnswer {
  field_id: string;
  field_label: string;
  canonical_type: string;
  value: string;
  was_autofilled: boolean;
  confirmed_by_candidate: boolean;
  is_custom_fact: boolean;
}

export interface ScrubbedArtifactItem {
  item_id: string;
  item_type: string;
  description: string;
  scrubbed_content: string;
  original_checksum: string;
  scrubbed_at: string;
}

export interface ApplicationEvidenceBundle {
  bundle_id: string;
  application_record_id: string;
  user_id: string;
  workspace_id?: string;
  job_snapshot: JobSnapshot;
  resume_artifact: ResumeArtifactSnapshot;
  cover_letter_artifact?: CoverLetterSnapshot;
  question_answers: SubmittedAnswer[];
  submission_timestamp: string;
  confirmation_type: AppliedVerificationType;
  provider_reference: string;
  scrubbed_artifacts: ScrubbedArtifactItem[];
  integrity_checksum: string;
  privacy_tier: string;
  sealed_at: string;
}

export interface CreateEvidenceBundleRequest {
  application_record_id: string;
  job_snapshot: JobSnapshot;
  resume_artifact: ResumeArtifactSnapshot;
  cover_letter_artifact?: CoverLetterSnapshot;
  question_answers: SubmittedAnswer[];
  submission_timestamp: string;
  confirmation_type: AppliedVerificationType;
  provider_reference: string;
  raw_artifacts?: ScrubbedArtifactItem[];
}

export interface CreateEvidenceBundleResponse {
  bundle: ApplicationEvidenceBundle;
  integrity_checksum: string;
  sealed_at: string;
}

export interface GetEvidenceBundleRequest {
  bundle_id?: string;
  application_record_id?: string;
  is_workspace_admin?: boolean;
  has_explicit_grant?: boolean;
}

export interface GetEvidenceBundleResponse {
  bundle: ApplicationEvidenceBundle;
  is_integrity_valid: boolean;
  integrity_checksum: string;
}

export interface VerifyBundleIntegrityRequest {
  bundle_id: string;
}

export interface VerifyBundleIntegrityResponse {
  bundle_id: string;
  is_valid: boolean;
  expected_checksum: string;
  computed_checksum: string;
}

// ==========================================
// Google Sheets Export & Sync Contracts (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014, SRC-C4)
// ==========================================

export type SheetsSyncMode = 'one_way_upsert' | 'append_only' | 'snapshot_overwrite';

export interface SheetsSyncConfig {
  user_id?: string;
  spreadsheet_id: string;
  sheet_name: string;
  timezone: string;
  sync_mode: SheetsSyncMode;
  include_notes: boolean;
  updated_at?: string;
}

export interface SheetsSyncResult {
  sync_id: string;
  spreadsheet_id: string;
  sheet_name: string;
  mode: SheetsSyncMode;
  total_applications: number;
  matched_count: number;
  appended_count: number;
  updated_count: number;
  unmodified_count: number;
  formula_injection_sanitized_count: number;
  rows_checksum: string;
  synced_at: string;
}

export interface SheetsExportPayload {
  spreadsheet_id: string;
  sheet_name: string;
  headers: string[];
  rows: string[][];
  csv_content: string;
  generated_at: string;
  row_count: number;
}

export interface SyncApplicationsToGoogleSheetsRequest {
  existing_sheet_rows?: string[][];
}

export interface SyncApplicationsToGoogleSheetsResponse {
  result: SheetsSyncResult;
  reconciled_rows: string[][];
}

// ==========================================
// Multi-Domain CSV & Additional Exports Contracts (IMP-CAR-18, CAR-18, REQ-006, AT-013, FND-005, FND-006, SRC-C2, SRC-C7, SRC-C8)
// ==========================================

export type ExportDatasetType = 'jobs' | 'applications' | 'contacts' | 'analysis' | 'all_bundle';

export interface RecruiterContact {
  id: string;
  user_id: string;
  name: string;
  role_title: string;
  company: string;
  email: string;
  linkedin_url?: string;
  status: 'lead' | 'contacted' | 'in_conversation' | 'archived';
  linked_application_id?: string;
  last_interaction_at?: string;
  next_follow_up_at?: string;
  notes?: string;
  created_at: string;
}

export interface CareerFunnelMetric {
  metric_key: string;
  category: 'funnel' | 'skills' | 'market' | 'sourcing';
  label: string;
  value: string;
  unit: string;
  benchmark?: string;
  notes?: string;
}

export interface ExportFilterOptions {
  dataset_type: ExportDatasetType;
  start_date?: string;
  end_date?: string;
  stage_filter?: string;
  platform_filter?: string;
  selected_columns?: string[];
  with_bom?: boolean;
  include_headers?: boolean;
  timezone?: string;
}

export interface CSVExportManifest {
  export_id: string;
  user_id: string;
  dataset_type: ExportDatasetType;
  filename: string;
  headers: string[];
  row_count: number;
  csv_content: string;
  checksum_sha256: string;
  with_bom: boolean;
  generated_at: string;
}

export interface ExportAuditRecord {
  audit_id: string;
  user_id: string;
  dataset_type: ExportDatasetType;
  filename: string;
  row_count: number;
  format: string;
  with_bom: boolean;
  exported_at: string;
}

// ==========================================
// Daily Reports & Reminders (IMP-CAR-19, CAR-19, AT-018)
// ==========================================

export type ReminderType = 'follow_up' | 'interview' | 'profile_incomplete' | 'custom' | 'job_review';
export type ReminderPriority = 'low' | 'medium' | 'high' | 'urgent';
export type ReminderStatus = 'pending' | 'snoozed' | 'dismissed' | 'completed';

export interface CareerReminder {
  id: string;
  user_id: string;
  workspace_id: string;
  type: ReminderType;
  priority: ReminderPriority;
  title: string;
  description: string;
  target_id?: string;
  due_at: string;
  snoozed_until?: string;
  status: ReminderStatus;
  created_at: string;
  updated_at: string;
}

export interface FollowUpRecommendation {
  application_id: string;
  company: string;
  role: string;
  applied_at: string;
  days_since_applied: number;
  suggested_action: string;
  recruiter_contact?: string;
}

export interface UpcomingInterview {
  id: string;
  company: string;
  role: string;
  round: string;
  scheduled_at: string;
  meeting_link?: string;
  prep_notes?: string;
}

export interface IncompleteProfileAlert {
  section: string;
  issue: string;
  impact: string;
  recommended_action: string;
  severity: 'low' | 'medium' | 'high';
}

export interface JobDiscoveryHighlight {
  job_id: string;
  title: string;
  company: string;
  location: string;
  match_score: number;
  direct_apply_url?: string;
  disclosed_salary?: string;
}

export interface DailyActivitySummary {
  applied_today: number;
  interviews_scheduled: number;
  follow_ups_due: number;
  active_applications_total: number;
  momentum_score: number;
}

export interface DailyReport {
  id: string;
  user_id: string;
  workspace_id: string;
  report_date: string;
  timezone: string;
  activity: DailyActivitySummary;
  follow_ups: FollowUpRecommendation[];
  interviews: UpcomingInterview[];
  incomplete_profile_items: IncompleteProfileAlert[];
  job_highlights: JobDiscoveryHighlight[];
  generated_at: string;
}

export interface NotificationChannels {
  in_app: boolean;
  webhook_enabled: boolean;
  webhook_url?: string;
  telegram_enabled: boolean;
  telegram_chat_id?: string;
}

export interface DailyReportConfig {
  user_id: string;
  workspace_id: string;
  enabled: boolean;
  scheduled_hour: number;
  scheduled_minute: number;
  timezone: string;
  delivery_days: 'weekdays' | 'daily';
  channels: NotificationChannels;
  categories: Record<string, boolean>;
  next_delivery_utc: string;
  updated_at: string;
}

// --- Hiring Posts & Recruiter Leads Contracts (IMP-CAR-20, CAR-20, AT-019) ---

export type LeadReviewStatus = 'pending_review' | 'reviewed' | 'approved' | 'rejected';
export type LeadOutreachStatus = 'draft' | 'ready_to_send' | 'copied_to_clipboard' | 'contacted' | 'replied' | 'archived';
export type OutreachTemplateType = 'linkedin_connect' | 'linkedin_inmail' | 'email_intro';

export interface ExtractedHiringRole {
  role_title: string;
  seniority: string;
  location: string;
  is_remote: boolean;
  tech_stack: string[];
  apply_instructions?: string;
  confidence: number;
}

export interface HiringPost {
  id: string;
  user_id: string;
  source_platform: string;
  post_url: string;
  author_name: string;
  author_title: string;
  author_linkedin_url: string;
  company: string;
  raw_content: string;
  cleaned_content: string;
  hiring_keywords_found: string[];
  extracted_roles: ExtractedHiringRole[];
  recruiter_email?: string;
  security_flags?: string[];
  imported_at: string;
  status: string;
}

export interface RecruiterLead {
  id: string;
  user_id: string;
  contact_id: string;
  hiring_post_id: string;
  recruiter_name: string;
  recruiter_title: string;
  company: string;
  recruiter_email?: string;
  linkedin_url: string;
  role_interest: string;
  source_post_url: string;
  review_status: LeadReviewStatus;
  outreach_status: LeadOutreachStatus;
  human_reviewed: boolean;
  human_reviewed_at?: string;
  reviewed_by?: string;
  review_notes?: string;
  linked_application_id?: string;
  draft_subject?: string;
  draft_message?: string;
  security_alerts?: string[];
  created_at: string;
  updated_at: string;
}

export interface OutreachDraft {
  template_type: OutreachTemplateType;
  subject?: string;
  body: string;
  character_count: number;
  personalized_highlights: string[];
  warnings?: string[];
}

export interface IngestHiringPostRequest {
  source_platform: string;
  post_url: string;
  author_name: string;
  author_title: string;
  author_linkedin_url: string;
  company: string;
  raw_content: string;
}

// Run Controls & Crash Recovery Contracts (IMP-CAR-21, CAR-21, AT-006, AT-021, REQ-017, FND-011)
export type CareerRunType = 'discovery_run' | 'tailoring_batch' | 'application_session' | 'feed_scan';

export type CareerRunStatus =
  | 'queued'
  | 'running'
  | 'paused'
  | 'draining'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'needs_reconciliation'
  | 'crashed';

export type StageSafetyLevel = 'safe_to_retry' | 'uncertain_write_requires_reconciliation';

export type RunControlAction = 'pause' | 'resume' | 'cancel' | 'drain';

export interface RunStageCheckpoint {
  stage_name: string;
  safety: StageSafetyLevel;
  started_at: string;
  completed_at?: string;
  status: string;
  items_processed: number;
  diagnostics?: string;
}

export interface RunDiagnosticLog {
  timestamp: string;
  level: string;
  stage: string;
  message: string;
  fencing_token: number;
}

export interface CareerRun {
  id: string;
  user_id: string;
  workspace_id: string;
  run_type: CareerRunType;
  status: CareerRunStatus;
  current_stage: string;
  stage_safety: StageSafetyLevel;
  fencing_token: number;
  items_total: number;
  items_processed: number;
  items_succeeded: number;
  items_failed: number;
  hourly_limit: number;
  recent_submissions: string[];
  checkpoints: RunStageCheckpoint[];
  audit_log: RunDiagnosticLog[];
  heartbeat_at: string;
  lease_expires_at: string;
  requires_reconciliation: boolean;
  reconciliation_notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCareerRunRequest {
  workspace_id?: string;
  run_type?: CareerRunType;
  total_items?: number;
  hourly_limit?: number;
}

export interface RunControlRequest {
  action: RunControlAction;
  reason?: string;
}

export interface ReconcileRunRequest {
  resolution: 'confirm_completed' | 'abandon_attempt' | 'force_retry';
  mark_succeeded: boolean;
  notes: string;
}

// PII Controls & LLM Provider Choices Contracts (IMP-CAR-22, CAR-22, REQ-023, AT-019)
export type PIIMinimizationLevel = 'full_redaction' | 'partial_mask' | 'strict_skills_only';

export interface PIIRedactionPolicy {
  minimization_level: PIIMinimizationLevel;
  redact_names: boolean;
  redact_emails: boolean;
  redact_phones: boolean;
  redact_links: boolean;
  redact_locations: boolean;
  redact_compensation: boolean;
  custom_pii_terms?: string[];
}

export interface PIIRedactionResult {
  original_length: number;
  redacted_length: number;
  redacted_text: string;
  token_map: Record<string, string>;
  detected_items_count: number;
  detected_categories: Record<string, number>;
  security_alerts?: string[];
  disclaimer: string;
}

export interface UserProviderConsent {
  id: string;
  user_id: string;
  provider_id: string;
  provider_name: string;
  status: 'active' | 'revoked';
  allowed_tasks: string[];
  zero_training_affirmed: boolean;
  consented_at: string;
  revoked_at?: string;
  retention_days: number;
}

export interface RedactPIIRequest {
  text: string;
  policy?: PIIRedactionPolicy;
}

export interface RehydratePIIRequest {
  anonymized_text: string;
  token_map: Record<string, string>;
}

export interface GrantProviderConsentRequest {
  provider_id: string;
  provider_name: string;
  allowed_tasks: string[];
  zero_training_affirmed: boolean;
  retention_days: number;
}

// ============================================================================
// Interview Preparation & Outcomes (IMP-CAR-23, CAR-23, AT-003)
// ============================================================================

export type QuestionCategory =
  | 'behavioral_star'
  | 'technical'
  | 'system_design'
  | 'role_fit'
  | 'reverse_questions';

export type RoundOutcome = 'pending' | 'passed' | 'failed' | 'skipped';

export interface CompanyBrief {
  company_name: string;
  domain?: string;
  industry?: string;
  known_size?: string;
  known_tech_stack?: string[];
  mission?: string;
  culture_notes?: string[];
  unverified_fields?: string[];
  synthesized_at: string;
}

export interface StarGuidance {
  situation: string;
  task: string;
  action: string;
  result: string;
}

export interface InterviewQuestion {
  id: string;
  category: QuestionCategory;
  difficulty: string;
  question: string;
  context_or_goal: string;
  suggested_talking_points: string[];
  grounded_fact_ids?: string[];
  identified_skill_gaps?: string[];
  star_guidance?: StarGuidance;
}

export interface InterviewPreparationPack {
  pack_id: string;
  application_record_id: string;
  candidate_id: string;
  job_title: string;
  company_brief: CompanyBrief;
  questions: InterviewQuestion[];
  preparation_checklist: string[];
  created_at: string;
}

export interface InterviewSchedule {
  schedule_id: string;
  application_record_id: string;
  round_id: string;
  stage_name: string;
  scheduled_at: string;
  timezone: string;
  format: string;
  meeting_link?: string;
  interviewer_name?: string;
  prep_reminder_at: string;
  follow_up_reminder_at: string;
  outcome: RoundOutcome;
  created_at: string;
}

export interface InterviewSessionNote {
  note_id: string;
  application_record_id: string;
  round_id: string;
  candidate_id: string;
  pre_interview_notes?: string;
  questions_asked_by_candidate?: string[];
  post_interview_reflections?: string;
  interviewer_name?: string;
  interviewer_title?: string;
  self_rating: number; // 1-5
  follow_up_actions?: string;
  updated_at: string;
}

export interface OutcomeFunnel {
  total_applications: number;
  stage_counts: Record<string, number>;
  conversion_rates: Record<string, number>;
  average_days_in_stage: Record<string, number>;
  overall_offer_rate: number;
  overall_reject_rate: number;
  computed_at: string;
}

export interface GenerateInterviewPrepRequest {
  application_record_id: string;
  job_title: string;
  required_skills: string[];
  company: CompanyBrief;
}

export interface ScheduleInterviewRoundRequest {
  application_record_id: string;
  round_id: string;
  stage_name: string;
  scheduled_at: string;
  timezone?: string;
  format?: string;
  meeting_link?: string;
  interviewer_name?: string;
}

export interface RecordInterviewNoteRequest {
  application_record_id: string;
  round_id: string;
  pre_interview_notes?: string;
  questions_asked_by_candidate?: string[];
  post_interview_reflections?: string;
  interviewer_name?: string;
  interviewer_title?: string;
  self_rating: number;
  follow_up_actions?: string;
}

export interface UpdateRoundOutcomeRequest {
  schedule_id: string;
  outcome: RoundOutcome;
}

// Career Consistency Check Contracts (IMP-CAR-24, CAR-24, REQ-004, AT-003, AT-007)
export type FactSource = 'master_resume' | 'master_profile' | 'linkedin_profile';

export type MismatchedFieldType =
  | 'job_title'
  | 'employment_date'
  | 'company_name'
  | 'skill_coverage'
  | 'education';

export type MismatchSeverity = 'high' | 'medium' | 'low';

export type ResolutionAction =
  | 'retain_resume'
  | 'retain_linkedin'
  | 'retain_profile'
  | 'custom_override';

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

export interface LinkedInSnapshotExperience {
  id?: string;
  company: string;
  title: string;
  start_date?: string;
  end_date?: string;
  is_current?: boolean;
}

export interface LinkedInProfileSnapshot {
  snapshot_id: string;
  user_id: string;
  headline: string;
  summary?: string;
  experiences: LinkedInSnapshotExperience[];
  skills: string[];
  imported_at: string;
  source_mode: string;
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

export interface ResolveConsistencyMismatchRequest {
  report_id: string;
  mismatch_id: string;
  decision: ResolutionDecision;
}

// LinkedIn Integration & Capabilities Contracts (IMP-LI-01, LI-01, REQ-004, REQ-021, AT-010, AT-016)
export type LinkedInScopeDomain = 'identity' | 'profile_read' | 'content_publish' | 'messaging' | 'easy_apply';

export type LinkedInCapabilityStatus =
  | 'granted'
  | 'denied'
  | 'unavailable_partner_only'
  | 'unsupported_platform';

export type LinkedInConnectionStatus =
  | 'connected'
  | 'partially_authorized'
  | 'reauth_required'
  | 'disconnected';

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
  profile_picture?: string;
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

export interface ConnectLinkedInRequest {
  member_id?: string;
  display_name?: string;
  email?: string;
  granted_scopes: string[];
  raw_token?: string;
  expires_in_hours?: number;
}

export interface CheckLinkedInActionRequest {
  action: string;
}

export interface CheckLinkedInActionResponse {
  action: string;
  allowed: boolean;
  capability?: InspectedLinkedInCapability;
  error?: string;
}

// LinkedIn Own-Profile Scan & Ingestion Contracts (IMP-LI-02, LI-02, REQ-004, AT-001, AT-010)
export type LinkedInIngestionSourceMode =
  | 'archive_zip'
  | 'manual_paste'
  | 'enterprise_api'
  | 'oidc_basic_only';

export interface LinkedInImportedExperience {
  company_name: string;
  title: string;
  description?: string;
  location?: string;
  start_date?: string;
  end_date?: string;
  is_current: boolean;
}

export interface LinkedInImportedEducation {
  school_name: string;
  degree_name?: string;
  field_of_study?: string;
  start_date?: string;
  end_date?: string;
  notes?: string;
}

export interface LinkedInImportedCertification {
  name: string;
  authority?: string;
  url?: string;
  license_number?: string;
  start_date?: string;
  end_date?: string;
}

export interface LinkedInImportedProfile {
  import_id: string;
  user_id: string;
  source_mode: LinkedInIngestionSourceMode;
  display_name: string;
  email?: string;
  headline?: string;
  summary?: string;
  location?: string;
  industry?: string;
  websites?: string[];
  experiences: LinkedInImportedExperience[];
  education: LinkedInImportedEducation[];
  skills: string[];
  certifications?: LinkedInImportedCertification[];
  requires_fallback: boolean;
  recommended_fallback?: string;
  truth_in_advertising_note?: string;
  imported_at: string;
  status: 'imported' | 'merged' | 'pending_review';
}

export interface ImportPastedLinkedInProfileRequest {
  raw_text: string;
}

export interface MergeLinkedInProfileRequest {
  import_id?: string;
  merge_headline?: boolean;
  merge_summary?: boolean;
  merge_experiences?: boolean;
  merge_education?: boolean;
  merge_skills?: boolean;
}

export interface MergeLinkedInProfileResponse {
  status: string;
  import_id: string;
  snapshot_id: string;
  source_mode: LinkedInIngestionSourceMode;
  merged_at: string;
  experiences: number;
  education: number;
  skills: number;
}

// LinkedIn Profile Optimization Contracts (IMP-LI-03, LI-03, REQ-004, AT-003, AT-007)
export type LinkedInOptimizationSection =
  | 'headline'
  | 'about'
  | 'experience'
  | 'featured'
  | 'skills';

export type LinkedInSuggestionApprovalState =
  | 'pending'
  | 'approved'
  | 'rejected'
  | 'custom_edited';

export interface LinkedInSectionSuggestion {
  suggestion_id: string;
  section: LinkedInOptimizationSection;
  before: string;
  after: string;
  rationale: string;
  source_facts_used: string[];
  impact_score: number;
  approval_state: LinkedInSuggestionApprovalState;
  custom_override?: string;
  approved_at?: string;
}

export interface LinkedInProfileOptimizationReport {
  report_id: string;
  user_id: string;
  candidate_name: string;
  target_role: string;
  current_headline: string;
  current_about: string;
  overall_profile_score: number;
  suggestions: LinkedInSectionSuggestion[];
  generated_at: string;
}

export interface GenerateOptimizationReportRequest {
  target_role?: string;
  current_headline?: string;
  current_about?: string;
  verified_source_facts?: string[];
}

export interface ApproveOptimizationSuggestionRequest {
  report_id: string;
  suggestion_id: string;
}

export interface CustomEditOptimizationSuggestionRequest {
  report_id: string;
  suggestion_id: string;
  custom_text: string;
}

// LinkedIn Records Engine Contracts (IMP-LI-04, LI-04, AT-011, AT-012, FND-009, FND-012, SRC-L3, SRC-L4)
export type LinkedInEntityType = 'person' | 'company' | 'job' | 'post';

export interface LinkedInPersonRecord {
  record_id: string;
  workspace_id: string;
  tenant_id: string;
  entity_type: 'person';
  vanity_slug: string;
  full_name: string;
  headline: string;
  current_company: string;
  current_role: string;
  location: string;
  connection_tier: string;
  public_profile_url: string;
  skills: string[];
  version: number;
  observed_at: string;
}

export interface LinkedInCompanyRecord {
  record_id: string;
  workspace_id: string;
  tenant_id: string;
  entity_type: 'company';
  universal_name: string;
  company_name: string;
  domain: string;
  industry: string;
  size_tier: string;
  headquarters: string;
  specialties: string[];
  follower_count: number;
  verified: boolean;
  version: number;
  observed_at: string;
}

export interface LinkedInJobRecord {
  record_id: string;
  workspace_id: string;
  tenant_id: string;
  entity_type: 'job';
  job_id: string;
  title: string;
  company_name: string;
  workplace_type: string;
  location: string;
  employment_type: string;
  description_snippet: string;
  applicant_count: number;
  posted_at: string;
  version: number;
  observed_at: string;
}

export interface LinkedInPostRecord {
  record_id: string;
  workspace_id: string;
  tenant_id: string;
  entity_type: 'post';
  urn: string;
  author_name: string;
  author_vanity: string;
  commentary: string;
  media_type: string;
  reactions_count: number;
  comments_count: number;
  shares_count: number;
  published_at: string;
  version: number;
  observed_at: string;
}

export interface LinkedInRecordFilter {
  workspace_id: string;
  tenant_id: string;
  query?: string;
  limit?: number;
}

// LinkedIn Professional Discovery Contracts (IMP-LI-05, LI-05, AT-010, FND-011, SRC-L1, SRC-L3)
export type LinkedInUserGoal = 'networking' | 'recruiting' | 'partnerships' | 'benchmarking';

export interface LinkedInDiscoveryCriteria {
  target_role: string;
  role_alternatives?: string[];
  industries?: string[];
  locations?: string[];
  current_companies?: string[];
  connection_tiers?: string[];
  company_size_tiers?: string[];
  exclusions?: string[];
  user_goal: LinkedInUserGoal;
  custom_filters?: Record<string, string>;
}

export interface LinkedInBooleanQuery {
  raw_query: string;
  role_clause: string;
  company_clause: string;
  exclusion_clause: string;
  linkedin_search_url: string;
}

export interface LinkedInCanonicalURL {
  raw_url: string;
  normalized_url: string;
  entity_type: 'person' | 'company';
  slug: string;
}

export interface LinkedInDiscoveryResult {
  criteria: LinkedInDiscoveryCriteria;
  boolean_query: LinkedInBooleanQuery;
  matched_persons: LinkedInPersonRecord[];
  matched_companies: LinkedInCompanyRecord[];
  total_vault_matches: number;
  platform_notice: string;
  generated_at: string;
}

export interface LinkedInDiscoverySearchRequest {
  workspace_id?: string;
  criteria: LinkedInDiscoveryCriteria;
}

export interface LinkedInNormalizeURLRequest {
  raw_url: string;
}

// Recruiter & Hiring Lead Workspace Contracts (IMP-LI-06, LI-06, AT-011)
export type LinkedInLeadStatus =
  | 'new'
  | 'contacted'
  | 'in_dialogue'
  | 'interview_scheduled'
  | 'offer_pending'
  | 'closed'
  | 'archived';

export type LinkedInOutreachStage =
  | 'draft'
  | 'ready_to_send'
  | 'copied_to_clipboard'
  | 'contacted'
  | 'replied'
  | 'archived';

export type LinkedInReminderStatus = 'pending' | 'completed' | 'snoozed';

export interface LinkedInLeadNote {
  id: string;
  lead_id: string;
  author: string;
  content: string;
  created_at: string;
}

export interface LinkedInFollowUpReminder {
  due_date: string;
  message: string;
  status: LinkedInReminderStatus;
  completed_at?: string;
}

export interface LinkedInRecruiterLead {
  id: string;
  workspace_id: string;
  tenant_id: string;
  recruiter_name: string;
  recruiter_title: string;
  company: string;
  linkedin_url: string;
  source_entity_type?: 'person_record' | 'post_record' | 'manual';
  source_entity_id?: string;
  status: LinkedInLeadStatus;
  outreach_stage: LinkedInOutreachStage;
  related_job_id?: string;
  related_application_id?: string;
  notes: LinkedInLeadNote[];
  reminder?: LinkedInFollowUpReminder;
  created_at: string;
  updated_at: string;
}

export interface LinkedInRecruiterLeadFilter {
  workspace_id?: string;
  tenant_id?: string;
  company?: string;
  status?: LinkedInLeadStatus;
  search_query?: string;
}

export interface LinkedInOutreachDraftPayload {
  subject: string;
  body: string;
  character_count: number;
  within_limit: boolean;
  direct_chat_url: string;
}

// Connection Note Drafts & Queue Contracts (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1)
export type LinkedInQueueItemStatus =
  | 'pending_approval'
  | 'approved'
  | 'copied_to_clipboard'
  | 'completed_manually'
  | 'rejected'
  | 'skipped';

export interface LinkedInConnectionBudget {
  workspace_id: string;
  tenant_id: string;
  daily_limit: number;
  daily_used: number;
  weekly_limit: number;
  weekly_used: number;
  last_reset_date: string;
}

export interface LinkedInConnectionQueueItem {
  id: string;
  workspace_id: string;
  tenant_id: string;
  recipient_id?: string;
  recipient_name: string;
  recipient_title: string;
  recipient_company: string;
  recipient_linkedin_url: string;
  note_text: string;
  character_count: number;
  within_limit: boolean;
  status: LinkedInQueueItemStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  context_factors?: string[];
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface LinkedInConnectionQueueFilter {
  workspace_id?: string;
  tenant_id?: string;
  status?: LinkedInQueueItemStatus;
  search_query?: string;
}

export interface LinkedInDraftNoteRequest {
  candidate_name?: string;
  target_role?: string;
  recipient_name: string;
  recipient_company?: string;
  recipient_title?: string;
  key_overlap?: string;
}

export interface LinkedInDraftNoteResponse {
  note_text: string;
  factors: string[];
  character_count: number;
  within_limit: boolean;
}

export interface LinkedInEnqueueNoteRequest {
  workspace_id: string;
  tenant_id?: string;
  recipient_id?: string;
  recipient_name: string;
  recipient_title: string;
  recipient_company: string;
  recipient_linkedin_url: string;
  note_text: string;
  context_factors?: string[];
}

export interface LinkedInApproveQueueItemRequest {
  id: string;
  approver?: string;
}

export interface LinkedInEditQueueNoteRequest {
  id: string;
  note_text: string;
}

export interface LinkedInConfirmSentRequest {
  id: string;
}

// =========================================================================
// IMP-LI-08: Company-Follow Planning Contracts (LI-08, AT-010, FND-011, SRC-L1)
// =========================================================================

export type CompanyFollowPriority = 'high' | 'medium' | 'low';
export type CompanyWatchlistStatus = 'active' | 'paused' | 'archived';
export type CompanyPlanItemStatus = 'planned' | 'ready_for_manual_follow' | 'manual_followed' | 'skipped';
export type BatchPlanStatus = 'draft' | 'in_progress' | 'completed';

export interface CompanyWatchlistItem {
  item_id: string;
  workspace_id: string;
  tenant_id: string;
  company_record_id?: string;
  company_name: string;
  universal_name: string;
  domain?: string;
  company_page_url: string;
  priority: CompanyFollowPriority;
  target_reason: string;
  tags: string[];
  status: CompanyWatchlistStatus;
  created_at: string;
  updated_at: string;
}

export interface CompanyFollowBudget {
  workspace_id: string;
  tenant_id: string;
  daily_limit: number;
  weekly_limit: number;
  daily_used: number;
  weekly_used: number;
  last_reset_date: string;
  weekly_reset_date: string;
  is_locked: boolean;
  lock_reason?: string;
}

export interface CompanyFollowPlanItem {
  item_id: string;
  company_record_id?: string;
  company_name: string;
  universal_name: string;
  company_page_url: string;
  priority: CompanyFollowPriority;
  target_reason: string;
  pacing_interval_sec: number;
  status: CompanyPlanItemStatus;
  scheduled_for: string;
  followed_at?: string;
  notes?: string;
}

export interface BatchCompanyFollowPlan {
  plan_id: string;
  workspace_id: string;
  tenant_id: string;
  plan_name: string;
  items: CompanyFollowPlanItem[];
  total_count: number;
  followed_count: number;
  status: BatchPlanStatus;
  created_at: string;
}

export interface AddCompanyWatchlistRequest {
  workspace_id: string;
  company_record_id?: string;
  company_name: string;
  universal_name: string;
  domain?: string;
  priority: CompanyFollowPriority;
  target_reason: string;
  tags: string[];
  status?: CompanyWatchlistStatus;
}

export interface CreateBatchFollowPlanRequest {
  workspace_id: string;
  plan_name: string;
  item_ids: string[];
  pacing_interval_sec: number;
}

export interface ConfirmCompanyFollowRequest {
  plan_id: string;
  item_id: string;
}

export interface SkipCompanyFollowRequest {
  plan_id: string;
  item_id: string;
  reason?: string;
}

export interface UpdateCompanyFollowBudgetRequest {
  workspace_id: string;
  daily_limit?: number;
  weekly_limit?: number;
}

// ==========================================
// Inbox & Conversation Triage Contracts (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3)
// ==========================================

export type LinkedInInboxThreadType = 'recruiter' | 'peer_connection' | 'inmail' | 'general';
export type LinkedInInboxClassification = 'interview_invitation' | 'recruiter_inquiry' | 'networking' | 'follow_up_response' | 'spam_or_promo';
export type LinkedInReplyDraftStatus = 'draft' | 'approved' | 'copied_to_clipboard' | 'rejected';
export type LinkedInReplyDraftTone = 'concise_scheduling' | 'professional_enthusiastic' | 'polite_decline' | 'inquiry_clarification';

export interface LinkedInInboxMessage {
  message_id: string;
  thread_id: string;
  sender_name: string;
  sender_type: 'self' | 'other';
  content: string;
  sent_at: string;
  is_read?: boolean;
}

export interface LinkedInConversationThread {
  thread_id: string;
  workspace_id: string;
  tenant_id: string;
  participant_name: string;
  participant_vanity?: string;
  participant_headline?: string;
  participant_company?: string;
  participant_avatar_url?: string;
  subject: string;
  last_message_snippet: string;
  last_message_at: string;
  unread_count: number;
  thread_type: LinkedInInboxThreadType;
  classification: LinkedInInboxClassification;
  active_followup_planned: boolean;
  messages: LinkedInInboxMessage[];
}

export interface LinkedInReplyDraft {
  draft_id: string;
  thread_id: string;
  workspace_id: string;
  tenant_id: string;
  tone: LinkedInReplyDraftTone;
  suggested_text: string;
  rationale: string;
  verified_facts_used: string[];
  character_count: number;
  status: LinkedInReplyDraftStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface GenerateReplyDraftRequest {
  thread_id: string;
  tone: LinkedInReplyDraftTone;
  candidate_availability?: string;
  verified_facts?: string[];
}

export interface ApproveReplyDraftRequest {
  draft_id: string;
}

export interface EditReplyDraftRequest {
  draft_id: string;
  new_text: string;
}

export interface CopyReplyDraftRequest {
  draft_id: string;
}

export interface RejectReplyDraftRequest {
  draft_id: string;
  reason?: string;
}

export interface AddThreadMessageRequest {
  thread_id: string;
  message: LinkedInInboxMessage;
}

// ==========================================
// Comments, Replies and Thread Sweep (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, FND-010, FND-015, SRC-L1, SRC-L2)
// ==========================================

export type LinkedInPostSweepCategory =
  | 'technical_discussion'
  | 'hiring_announcement'
  | 'thought_leadership'
  | 'industry_news'
  | 'general_discussion';

export type LinkedInCommentDraftAngle =
  | 'insightful_addition'
  | 'engaging_question'
  | 'supportive_perspective';

export type LinkedInCommentDraftStatus =
  | 'draft'
  | 'approved'
  | 'copied_to_clipboard'
  | 'rejected';

export interface LinkedInSweptPostComment {
  comment_id: string;
  author_name: string;
  author_headline?: string;
  content: string;
  created_at: string;
  likes_count: number;
  is_high_priority: boolean;
}

export interface LinkedInSweptTargetPost {
  post_id: string;
  workspace_id: string;
  tenant_id: string;
  author_name: string;
  author_headline?: string;
  author_company?: string;
  post_url: string;
  content: string;
  category: LinkedInPostSweepCategory;
  existing_comments_count: number;
  comments?: LinkedInSweptPostComment[];
  swept_at: string;
}

export interface LinkedInCommentDraft {
  draft_id: string;
  post_id: string;
  workspace_id: string;
  tenant_id: string;
  target_comment_id?: string;
  angle: LinkedInCommentDraftAngle;
  comment_text: string;
  rationale: string;
  verified_facts_used: string[];
  character_count: number;
  status: LinkedInCommentDraftStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface GenerateCommentDraftsRequest {
  post_id: string;
  mode?: string;
  candidate_facts?: string[];
}

export interface ApproveCommentDraftRequest {
  draft_id: string;
  approved_by?: string;
}

export interface EditCommentDraftRequest {
  draft_id: string;
  new_text: string;
}

export interface CopyCommentDraftRequest {
  draft_id: string;
}

export interface RejectCommentDraftRequest {
  draft_id: string;
  reason?: string;
}

// ==========================================
// LinkedIn Post Writing, Hooks and Audits (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2)
// ==========================================

export type LinkedInPostAngle =
  | 'contrarian_insight'
  | 'lesson_learned_breakdown'
  | 'technical_deep_dive'
  | 'milestone_celebration'
  | 'actionable_guide';

export type LinkedInHookType =
  | 'question'
  | 'contrarian'
  | 'data_scale'
  | 'story'
  | 'punchy';

export type LinkedInPostDraftStatus =
  | 'draft'
  | 'approved'
  | 'copied_to_clipboard'
  | 'rejected';

export type LinkedInAuditSeverity = 'info' | 'warning' | 'blocking';

export interface LinkedInHookVariant {
  hook_type: LinkedInHookType;
  hook_text: string;
  rationale: string;
}

export interface LinkedInAuditIssue {
  code: string;
  message: string;
  severity: LinkedInAuditSeverity;
  offending_snippet?: string;
  suggested_fix?: string;
}

export interface LinkedInEditorialAuditReport {
  passed: boolean;
  readability_score: number;
  character_count: number;
  line_break_density: number;
  paragraph_count: number;
  estimated_read_time_sec: number;
  buzzwords_count: number;
  virality_claims_count: number;
  ai_bypass_claims_count: number;
  unverified_metrics_count: number;
  issues: LinkedInAuditIssue[];
  audited_at: string;
  disclaimer: string;
}

export interface LinkedInPostDraft {
  draft_id: string;
  workspace_id: string;
  tenant_id: string;
  topic: string;
  angle: LinkedInPostAngle;
  selected_hook_type: LinkedInHookType;
  selected_hook_text: string;
  body_text: string;
  call_to_action: string;
  full_post_text: string;
  verified_facts_used: string[];
  available_hooks: LinkedInHookVariant[];
  latest_audit?: LinkedInEditorialAuditReport;
  status: LinkedInPostDraftStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface GeneratePostDraftRequest {
  workspace_id?: string;
  topic: string;
  angle: LinkedInPostAngle;
  verified_facts: string[];
  target_audience?: string;
  preferred_hook_type?: LinkedInHookType;
}

export interface ApprovePostDraftRequest {
  draft_id: string;
  approved_by?: string;
}

export interface EditPostDraftRequest {
  draft_id: string;
  selected_hook_type?: string;
  selected_hook_text?: string;
  full_post_text: string;
}

export interface CopyPostDraftRequest {
  draft_id: string;
}

export interface RejectPostDraftRequest {
  draft_id: string;
  reason?: string;
}

// ==========================================
// Humanizer and Reusable Voice (IMP-LI-12, LI-12, AT-003, AT-007, AT-010, FND-010, FND-015, SRC-L2, SRC-S4)
// ==========================================

export type LinkedInCadenceStyle =
  | 'punchy_staccato'
  | 'balanced_rhythm'
  | 'analytical_deep';

export type LinkedInPerspective =
  | 'first_person_singular'
  | 'collective_team'
  | 'neutral_practitioner';

export type LinkedInVoiceProfileStatus =
  | 'draft'
  | 'approved'
  | 'archived';

export interface LinkedInVoiceProfile {
  profile_id: string;
  workspace_id: string;
  tenant_id: string;
  profile_name: string;
  target_audience: string;
  formality: number; // 1 to 5
  technical_depth: number; // 1 to 5
  cadence_style: LinkedInCadenceStyle;
  perspective: LinkedInPerspective;
  preferred_terms: string[];
  blacklisted_terms: string[];
  status: LinkedInVoiceProfileStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface LinkedInSlopReplacement {
  tier: string;
  offending_phrase: string;
  suggested_replacement: string;
  category: string;
  rationale: string;
}

export interface LinkedInHumanizerAuditIssue {
  rule_code?: string;
  code?: string;
  message: string;
  severity: string; // "blocking" | "warning"
  offending_snippet: string;
  suggested_fix: string;
}

export interface LinkedInHumanizeResult {
  result_id: string;
  workspace_id: string;
  tenant_id: string;
  voice_profile_id?: string;
  raw_text: string;
  cleaned_text: string;
  tier1_slop_replacements: LinkedInSlopReplacement[];
  tier2_cadence_notes: string[];
  tier3_issues: LinkedInHumanizerAuditIssue[];
  slop_score_before: number;
  slop_score_after: number;
  readability_score_before: number;
  readability_score_after: number;
  passed: boolean;
  status: string; // "draft" | "approved" | "copied_to_clipboard"
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
}

export interface CreateVoiceProfileRequest {
  profile_id?: string;
  workspace_id?: string;
  profile_name: string;
  target_audience: string;
  formality: number;
  technical_depth: number;
  cadence_style: LinkedInCadenceStyle;
  perspective: LinkedInPerspective;
  preferred_terms: string[];
  blacklisted_terms: string[];
}

export interface ApproveVoiceProfileRequest {
  profile_id: string;
  approver?: string;
}

export interface HumanizeDraftTextRequest {
  raw_text: string;
  voice_profile_id?: string;
  workspace_id?: string;
  candidate_verified_facts?: string[];
}

export interface ApproveHumanizedDraftRequest {
  result_id: string;
  approver?: string;
}

export interface EditHumanizedTextRequest {
  result_id: string;
  updated_text: string;
  candidate_verified_facts?: string[];
}

export interface CopyHumanizedDraftRequest {
  result_id: string;
}

// =========================================================================
// Story Bank & Guided Interviewer (IMP-LI-13, LI-13, AT-003, SRC-L2)
// =========================================================================

export type LinkedInStoryCategory =
  | 'turning_point'
  | 'scar_or_failure'
  | 'breakthrough_win'
  | 'contrarian_belief'
  | 'mentorship_culture';

export type LinkedInStoryProvenanceType =
  | 'interview_session'
  | 'manual_entry'
  | 'imported_milestone';

export type LinkedInStoryStatus =
  | 'draft'
  | 'reviewed'
  | 'approved'
  | 'archived';

export interface LinkedInStoryProvenance {
  source_type: LinkedInStoryProvenanceType;
  interview_session_id?: string;
  author_name: string;
  timestamp: string;
  grounded_fact_ids?: string[];
}

export interface LinkedInStoryNarrative {
  hook_summary: string;
  context_background: string;
  challenge_conflict: string;
  action_taken: string;
  quantified_outcome: string;
  lesson_learned: string;
}

export interface LinkedInStoryAuditIssue {
  field: string;
  metric: string;
  severity: 'blocking' | 'advisory';
  message: string;
}

export interface LinkedInStoryAuditReport {
  has_blocking_issues: boolean;
  issues: LinkedInStoryAuditIssue[];
  readability_score: number;
  grounded_fact_count: number;
}

export interface LinkedInStoryEntry {
  id: string;
  workspace_id: string;
  tenant_id: string;
  title: string;
  category: LinkedInStoryCategory;
  provenance: LinkedInStoryProvenance;
  narrative: LinkedInStoryNarrative;
  tags: string[];
  approval_status: LinkedInStoryStatus;
  approval_token?: string;
  last_approved_at?: string;
  audit_report?: LinkedInStoryAuditReport;
  created_at: string;
  updated_at: string;
}

export interface LinkedInInterviewPrompt {
  id: string;
  category: LinkedInStoryCategory;
  title: string;
  question: string;
  context_placeholder: string;
  follow_up_probes: string[];
}

export interface LinkedInInterviewQA {
  prompt_id: string;
  question: string;
  candidate_answer: string;
  follow_up_probe?: string;
  candidate_follow_up_answer?: string;
}

export interface LinkedInInterviewSession {
  id: string;
  workspace_id: string;
  tenant_id: string;
  category: LinkedInStoryCategory;
  status: 'in_progress' | 'completed' | 'abandoned';
  questions_and_answers: LinkedInInterviewQA[];
  synthesized_story_id?: string;
  created_at: string;
  updated_at: string;
}

export interface StartInterviewSessionRequest {
  workspace_id: string;
  category: LinkedInStoryCategory;
}

export interface RecordInterviewAnswerRequest {
  session_id: string;
  prompt_id: string;
  answer: string;
}

export interface RecordInterviewFollowUpRequest {
  session_id: string;
  prompt_id: string;
  follow_up_answer: string;
}

export interface SynthesizeInterviewStoryRequest {
  session_id: string;
  author_name: string;
  verified_facts?: string[];
}

export interface CreateStoryEntryRequest {
  entry: Partial<LinkedInStoryEntry>;
  verified_facts?: string[];
}

export interface ApproveStoryEntryRequest {
  story_id: string;
  secret_key?: string;
  verified_facts?: string[];
}

export interface EditStoryEntryRequest {
  story_id: string;
  narrative: LinkedInStoryNarrative;
  verified_facts?: string[];
}

export interface CopyStoryEntryRequest {
  story_id: string;
  secret_key?: string;
}

export interface CopyStoryEntryResponse {
  story: LinkedInStoryEntry;
  markdown: string;
  copied: boolean;
  linkedin_deep_link: string;
}

export interface ArchiveStoryEntryRequest {
  story_id: string;
}

// =========================================================================
// Content Planning & Repurposing Contracts (IMP-LI-14, LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
// =========================================================================

export type LinkedInArtifactType =
  | 'technical_blog'
  | 'github_release'
  | 'architecture_rfc'
  | 'incident_postmortem'
  | 'benchmark_report';

export type LinkedInRepurposedFormat =
  | 'single_thought'
  | 'carousel_outline'
  | 'actionable_checklist'
  | 'contrarian_breakdown'
  | 'interview_qa_spotlight';

export type LinkedInContentPlanStatus =
  | 'draft'
  | 'reviewed'
  | 'approved'
  | 'scheduled'
  | 'archived';

export interface LinkedInSourceArtifact {
  id: string;
  workspace_id: string;
  tenant_id?: string;
  title: string;
  artifact_type: LinkedInArtifactType;
  raw_text: string;
  author: string;
  tags: string[];
  created_at?: string;
}

export interface LinkedInRepurposedDraft {
  id: string;
  workspace_id: string;
  tenant_id: string;
  source_artifact_id: string;
  source_title: string;
  format: LinkedInRepurposedFormat;
  title: string;
  content_body: string;
  hook: string;
  tags: string[];
  character_count: number;
  status: LinkedInContentPlanStatus;
  approval_token?: string;
  last_approved_at?: string;
  scheduled_slot_utc?: string;
  created_at: string;
  updated_at: string;
}

export interface LinkedInContentAssetMetadata {
  slide_count?: number;
  estimated_read_time_sec: number;
  canonical_url?: string;
  hashtags?: string[];
}

export interface LinkedInContentPlanItem {
  id: string;
  workspace_id: string;
  tenant_id: string;
  draft_id: string;
  title: string;
  format: LinkedInRepurposedFormat;
  scheduled_slot_utc: string;
  user_timezone: string;
  local_slot_formatted: string;
  status: LinkedInContentPlanStatus;
  approval_token?: string;
  last_approved_at?: string;
  asset_metadata: LinkedInContentAssetMetadata;
  direct_publish_blocked: boolean;
  clipboard_export_available: boolean;
  compose_url: string;
  created_at: string;
  updated_at: string;
}

export interface RepurposeSourceArtifactRequest {
  workspace_id: string;
  source: LinkedInSourceArtifact;
  format?: LinkedInRepurposedFormat;
}

export interface RepurposeSourceArtifactResponse {
  draft?: LinkedInRepurposedDraft;
  drafts: LinkedInRepurposedDraft[];
  count: number;
}

export interface ListRepurposedDraftsResponse {
  drafts: LinkedInRepurposedDraft[];
  count: number;
}

export interface UpdateRepurposedDraftRequest {
  draft_id: string;
  title: string;
  content_body: string;
}

export interface ApproveRepurposedDraftRequest {
  draft_id: string;
  secret_key?: string;
}

export interface ScheduleContentPlanRequest {
  workspace_id: string;
  draft_id: string;
  requested_local_time: string;
  user_timezone: string;
  max_daily_budget?: number;
}

export interface ListContentCalendarResponse {
  items: LinkedInContentPlanItem[];
  count: number;
}

export interface UpdateContentPlanItemRequest {
  item_id: string;
  title?: string;
  new_local_time?: string;
  user_timezone?: string;
  max_daily_budget?: number;
}

export interface ApproveContentPlanItemRequest {
  item_id: string;
  secret_key?: string;
}

export interface ExportPlanClipboardRequest {
  item_id: string;
  secret_key?: string;
}

export interface ExportPlanClipboardResponse {
  item: LinkedInContentPlanItem;
  markdown: string;
  compose_url: string;
  exported_at: string;
}

// ==========================================
// Engagement Monitoring & Analytics (IMP-LI-15, LI-15, AT-028, AT-010, AT-012, SRC-L2)
// ==========================================

export type LinkedInMetricStatus = 'available' | 'unavailable' | 'modeled_estimate';
export type LinkedInEngagerSegment = 'decision_maker' | 'peer_practitioner' | 'talent_partner' | 'other_network';
export type LinkedInInteractionType = 'reaction' | 'comment' | 'repost';

export interface LinkedInMetricItem {
  value: number;
  status: LinkedInMetricStatus;
  unavailable_reason?: string;
  is_modeled: boolean;
  model_basis?: string;
}

export interface LinkedInEngagerProfile {
  author_urn: string;
  name: string;
  headline: string;
  company: string;
  interaction_type: LinkedInInteractionType;
  comment_text?: string;
  segment?: LinkedInEngagerSegment;
  observed_at?: string;
}

export interface LinkedInICPBreakdown {
  decision_makers: number;
  peer_practitioners: number;
  talent_partners: number;
  other_network: number;
  high_value_engagers_count: number;
}

export interface LinkedInPostAnalyticsSnapshot {
  snapshot_id: string;
  workspace_id: string;
  tenant_id: string;
  post_urn: string;
  post_title: string;
  published_at: string;
  metrics: Record<string, LinkedInMetricItem>;
  engagers: LinkedInEngagerProfile[];
  icp_breakdown: LinkedInICPBreakdown;
  disclosure_notice?: string;
  observed_at: string;
  created_at: string;
  updated_at: string;
}

export interface LinkedInPerformanceBenchmark {
  category: string;
  sample_size: number;
  avg_reactions: number;
  avg_comments: number;
  avg_reposts: number;
  avg_engagement_rate: number;
  benchmark_source: string;
}

export interface ListPostAnalyticsResponse {
  snapshots: LinkedInPostAnalyticsSnapshot[];
  total: number;
}

export interface SegmentEngagersRequest {
  engagers: LinkedInEngagerProfile[];
}

export interface SegmentEngagersResponse {
  engagers: LinkedInEngagerProfile[];
  icp_breakdown: LinkedInICPBreakdown;
}

export interface CreatorBenchmarksResponse {
  benchmarks: LinkedInPerformanceBenchmark[];
  disclosure: string;
}

// ==========================================
// Employee Advocacy & Brand Governance (IMP-LI-16, LI-16, AT-007, AT-011)
// ==========================================

export type LinkedInAdvocacyCampaignStatus = 'draft' | 'approved' | 'archived';
export type LinkedInEmployeePersona = 'engineering' | 'product' | 'talent_culture' | 'general';

export interface LinkedInAdvocacyCopyVariant {
  variant_id: string;
  persona: LinkedInEmployeePersona;
  headline: string;
  suggested_text: string;
  target_tags?: string[];
}

export interface LinkedInBrandGovernance {
  guidelines: string;
  allowed_hashtags?: string[];
  forbidden_keywords?: string[];
}

export interface LinkedInAdvocacyCampaign {
  campaign_id: string;
  tenant_id: string;
  workspace_id: string;
  title: string;
  description: string;
  governance: LinkedInBrandGovernance;
  variants: LinkedInAdvocacyCopyVariant[];
  status: LinkedInAdvocacyCampaignStatus;
  approval_token?: string;
  approved_by?: string;
  approved_at?: string;
  share_count: number;
  created_at: string;
  updated_at: string;
}

export interface LinkedInEmployeeShareEvent {
  share_id: string;
  tenant_id: string;
  workspace_id: string;
  campaign_id: string;
  variant_id: string;
  employee_id: string;
  customized_text: string;
  shared_at: string;
  platform: string;
  share_method: string;
}

export interface CreateAdvocacyCampaignRequest {
  title: string;
  description: string;
  governance: LinkedInBrandGovernance;
  variants: LinkedInAdvocacyCopyVariant[];
}

export interface UpdateAdvocacyCampaignRequest {
  campaign_id: string;
  title: string;
  description: string;
  governance: LinkedInBrandGovernance;
  variants: LinkedInAdvocacyCopyVariant[];
}

export interface ApproveAdvocacyCampaignRequest {
  campaign_id: string;
  approver_id: string;
  secret_key?: string;
}

export interface ShareAdvocacyRequest {
  campaign_id: string;
  variant_id: string;
  employee_id: string;
  customized_text?: string;
  secret_key?: string;
}

export interface ShareAdvocacyResponse {
  success: boolean;
  share_id: string;
  clipboard_text: string;
  compose_deep_link: string;
  message: string;
}

export interface ListAdvocacyCampaignsResponse {
  campaigns: LinkedInAdvocacyCampaign[];
  total: number;
}

export interface ListEmployeeSharesResponse {
  shares: LinkedInEmployeeShareEvent[];
  total: number;
}

export interface CoordinatedEngagementPodRequest {
  campaign_id: string;
  target_post_urn: string;
  action: string;
  participating_employee_ids: string[];
  synthetic_comments?: string[];
}

export interface AntiPodCheckResponse {
  blocked: boolean;
  error: string;
  policy: string;
}

// ==========================================
// Provider Fallback & Diagnostics (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
// ==========================================

export type LinkedInProviderAdapterType =
  | 'official_enterprise_api'
  | 'consumer_oidc_app'
  | 'local_supervised_session'
  | 'rss_public_feed';

export type LinkedInProviderHealthStatus =
  | 'healthy'
  | 'degraded'
  | 'unauthorized'
  | 'rate_limited'
  | 'offline';

export interface LinkedInProviderAdapter {
  provider_id: string;
  name: string;
  adapter_type: LinkedInProviderAdapterType;
  tier: number;
  supported_actions: string[];
  expected_scopes: string[];
  granted_scopes: string[];
  health_status: LinkedInProviderHealthStatus;
  latency_ms: number;
  consecutive_failures: number;
  last_checked_at: string;
  last_error_message?: string;
  remediation_step?: string;
  tenant_id: string;
  workspace_id: string;
  created_at: string;
  updated_at: string;
}

export interface LinkedInDiagnosticProbeResult {
  provider_id: string;
  name: string;
  adapter_type: LinkedInProviderAdapterType;
  health_status: LinkedInProviderHealthStatus;
  latency_ms: number;
  missing_scopes: string[];
  error_message?: string;
  remediation_step?: string;
  probed_at: string;
}

export interface LinkedInProviderDiagnosticReport {
  workspace_id: string;
  tenant_id: string;
  overall_health: 'healthy' | 'degraded' | 'critical';
  total_providers: number;
  healthy_providers: number;
  degraded_providers: number;
  probes: LinkedInDiagnosticProbeResult[];
  recommendations?: string[];
  generated_at: string;
}

export interface LinkedInDispatchActionRequest {
  tenant_id?: string;
  workspace_id?: string;
  action: string;
  required_permission: string;
  payload?: Record<string, unknown>;
}

export interface LinkedInProviderDispatchResult {
  dispatched_provider_id: string;
  adapter_type: LinkedInProviderAdapterType;
  action: string;
  fallback_invoked: boolean;
  fallback_reason?: string;
  success: boolean;
  latency_ms?: number;
  result_data?: Record<string, unknown>;
  dispatched_at: string;
}

export interface ListProvidersResponse {
  providers: LinkedInProviderAdapter[];
  total: number;
}

export interface ScrubDiagnosticSecretsRequest {
  raw_text: string;
}

export interface ScrubDiagnosticSecretsResponse {
  raw_length: number;
  scrubbed_text: string;
  scrubbed_length: number;
}

// =========================================================================
// Platform Limits, CAPTCHA / Reauth & Safety Gatekeeper Contracts
// (IMP-LI-18, LI-18, REQ-009, REQ-010, AT-008, AT-009, AT-010, FND-011, FND-013, FND-014, SRC-L1, SRC-S3)
// =========================================================================

export type LinkedInPlatformLimitType =
  | 'weekly_invitations'
  | 'daily_messages'
  | 'daily_recruiter_dm'
  | 'profile_searches'
  | 'company_follows'
  | 'rate_limit_429'
  | 'security_checkpoint'
  | 'session_expired';

export type LinkedInChallengeType =
  | 'captcha'
  | 'email_pin'
  | 'sms_2fa'
  | 'credential_reauth'
  | 'commercial_use_cap'
  | 'weekly_invitation_limit';

export type LinkedInAccountSafetyStatus =
  | 'active'
  | 'paused'
  | 'challenge_required'
  | 'rate_limited'
  | 'revoked';

export interface LinkedInAccountSafetyState {
  account_id: string;
  tenant_id: string;
  workspace_id: string;
  status: LinkedInAccountSafetyStatus;
  active_limit_type?: LinkedInPlatformLimitType;
  active_challenge?: LinkedInChallengeType;
  challenge_reason?: string;
  cooldown_until?: string;
  consecutive_restrictions: number;
  last_incident_at?: string;
  last_resolved_at?: string;
  updated_at: string;
}

export interface LinkedInPlatformRestrictionIncident {
  incident_id: string;
  account_id: string;
  tenant_id: string;
  workspace_id: string;
  status_code: number;
  raw_error_message_scrubbed: string;
  limit_type: LinkedInPlatformLimitType;
  challenge_type?: LinkedInChallengeType;
  action_attempted: string;
  cooldown_hours: number;
  occurred_at: string;
}

export interface LinkedInActivityLogEntry {
  entry_id: string;
  account_id: string;
  tenant_id: string;
  workspace_id: string;
  module: string;
  action: string;
  amount: number;
  status: string;
  details?: string;
  timestamp: string;
}

export interface LinkedInCombinedAccountBudgetReport {
  account_id: string;
  tenant_id: string;
  computed_at: string;
  daily_invitations_limit: number;
  daily_invitations_used: number;
  daily_invitations_remaining: number;
  weekly_invitations_limit: number;
  weekly_invitations_used: number;
  weekly_invitations_remaining: number;
  daily_inmail_limit: number;
  daily_inmail_used: number;
  daily_inmail_remaining: number;
  daily_recruiter_dm_limit: number;
  daily_recruiter_dm_used: number;
  daily_recruiter_dm_remaining: number;
  daily_company_follows_limit: number;
  daily_company_follows_used: number;
  daily_company_follows_remaining: number;
  contributing_workspaces: string[];
}

export interface LinkedInResolveChallengeRequest {
  account_id?: string;
  resolution_method: string;
  verified_by_user: boolean;
  notes: string;
}

export interface LinkedInSafetyGateDecision {
  allowed: boolean;
  status: LinkedInAccountSafetyStatus;
  reason?: string;
}

export interface LinkedInDetectRestrictionRequest {
  account_id?: string;
  workspace_id?: string;
  action_attempted: string;
  status_code: number;
  raw_error: string;
}

export interface LinkedInDetectRestrictionResponse {
  incident: LinkedInPlatformRestrictionIncident;
  state: LinkedInAccountSafetyState;
}

export interface LinkedInListActivityLogResponse {
  account_id: string;
  total: number;
  activities: LinkedInActivityLogEntry[];
}

// =========================================================================
// Session Lifecycle, Storage Isolation & Optional Local Tools Contracts
// (IMP-LI-19, LI-19, AT-011, AT-012, AT-016, FND-009, FND-010, SRC-L3, SRC-L4)
// =========================================================================

export type LinkedInSessionLifecycleStatus =
  | 'active'
  | 'leasing'
  | 'expired'
  | 'revoked'
  | 'security_isolated';

export type LinkedInViewerLeaseLifecycleStatus =
  | 'active'
  | 'expired'
  | 'released'
  | 'revoked';

export interface LinkedInSessionStorageMetadata {
  vault_identifier: string;
  key_namespace: string;
  isolation_mode: string;
  zero_raw_cookies_stored: boolean;
  zero_plaintext_credentials: boolean;
  vault_reference_token: string;
  created_at: string;
  updated_at: string;
}

export interface LinkedInSessionRecord {
  id: string;
  tenant_id: string;
  owner_id: string;
  workspace_id: string;
  status: LinkedInSessionLifecycleStatus;
  storage_metadata: LinkedInSessionStorageMetadata;
  user_agent_signature: string;
  ip_hash: string;
  active_leases_count: number;
  total_leases_issued: number;
  last_used_at: string;
  expires_at: string;
  revoked_at?: string;
  revocation_reason?: string;
  created_at: string;
}

export interface ShortLivedViewerLease {
  id: string;
  session_id: string;
  tenant_id: string;
  owner_id: string;
  workspace_id: string;
  status: LinkedInViewerLeaseLifecycleStatus;
  purpose: string;
  viewer_tool_name: string;
  leased_at: string;
  expires_at: string;
  released_at?: string;
  lease_duration_seconds: number;
  client_nonce: string;
}

export interface CreateSessionRequest {
  tenant_id?: string;
  owner_id?: string;
  workspace_id?: string;
  vault_token: string;
  user_agent: string;
  ip_address: string;
  ttl_hours?: number;
  raw_cookie_paste?: string;
  raw_password_paste?: string;
}

export interface CreateSessionResponse {
  session: LinkedInSessionRecord;
}

export interface ListSessionsResponse {
  sessions: LinkedInSessionRecord[];
  total: number;
}

export interface RevokeSessionRequest {
  session_id: string;
  tenant_id?: string;
  owner_id?: string;
  reason?: string;
}

export interface RevokeSessionResponse {
  session: LinkedInSessionRecord;
  message: string;
}

export interface AcquireViewerLeaseRequest {
  session_id: string;
  tenant_id?: string;
  owner_id?: string;
  workspace_id?: string;
  purpose: string;
  viewer_tool_name: string;
  duration_seconds?: number;
}

export interface AcquireViewerLeaseResponse {
  lease: ShortLivedViewerLease;
}

export interface ReleaseViewerLeaseRequest {
  lease_id: string;
  tenant_id?: string;
  owner_id?: string;
}

export interface ReleaseViewerLeaseResponse {
  lease: ShortLivedViewerLease;
  message: string;
}

// Relationship CSV & Google Sheets Exports (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014)
export type LinkedInRelationshipExportDestination = 'csv_download' | 'google_sheets_sync';

export interface LinkedInRelationshipExportFilter {
  workspace_id: string;
  tenant_id?: string;
  status_filter?: string;
  company_filter?: string;
  include_notes: boolean;
  include_person_records: boolean;
  selected_columns?: string[];
  with_bom: boolean;
  timezone?: string;
}

export interface LinkedInRelationshipExportManifest {
  export_id: string;
  workspace_id: string;
  tenant_id: string;
  destination: LinkedInRelationshipExportDestination;
  filename: string;
  row_count: number;
  byte_size: number;
  checksum: string;
  with_bom: boolean;
  headers: string[];
  csv_content: string;
  generated_at: string;
}

export interface LinkedInRelationshipSheetsSyncConfig {
  config_id?: string;
  workspace_id: string;
  tenant_id?: string;
  spreadsheet_id: string;
  sheet_name: string;
  timezone?: string;
  include_notes: boolean;
  sync_mode?: string;
  existing_rows?: string[][];
}

export interface LinkedInRelationshipSheetsSyncResult {
  spreadsheet_id: string;
  sheet_name: string;
  total_synced: number;
  appended_count: number;
  updated_count: number;
  unchanged_count: number;
  checksum: string;
  synced_at: string;
  message: string;
}

export interface LinkedInRelationshipExportAuditRecord {
  audit_id: string;
  workspace_id: string;
  tenant_id: string;
  owner_id: string;
  destination: LinkedInRelationshipExportDestination;
  filename: string;
  row_count: number;
  include_notes: boolean;
  exported_at: string;
}

export interface LinkedInRelationshipSheetsSyncResponse {
  result: LinkedInRelationshipSheetsSyncResult;
  reconciled_rows: string[][];
}

export interface LinkedInRelationshipExportHistoryResponse {
  audits: LinkedInRelationshipExportAuditRecord[];
  total: number;
}


