# CONTEXT — Product requirements and source-selection contract

> Prepared 2026-09-17. This is the authoritative product specification in this pack. Architecture changes are recorded in PLAN.md; execution state belongs in TASKS.md and PROGRESS.md. Product name is a working label. No destination GitHub repository or production deployment has been created.

Before starting or continuing any work, read README.md, CONTEXT.md, PLAN.md, AI-AGENT-ARCHITECTURE.md, TASKS.md, and PROGRESS.md in full and treat them as one interconnected source of truth. Cross-reference requirements, architecture decisions, AI-agent rules, task dependencies, implementation status, and prior decisions across all six files before making changes. Do not interpret any file in isolation, silently override an existing decision, duplicate functionality, change the approved stack, or mark work complete without implementation and verification evidence. After each meaningful implementation step, update TASKS.md and PROGRESS.md so the next AI/coding session can resume accurately without losing context.

## 1. Decisions already made

The platform is engineered into **two distinct commercial flagship products** served under a unified master domain (`yourdomain.com`) with separate subscription tiers and dedicated project repositories:
1. **Product 1: Career & LinkedIn Platform (`career/` repository)** — Located at `yourdomain.com/career` & `/linkedin`. Dedicated to job seekers, candidates, and professionals with its own **Career & LinkedIn Pro** subscription tier.
2. **Product 2: SocialSuite (`SocialSuite/` repository)** — Located at `yourdomain.com/social`. Dedicated to creators, social media managers, agencies, and brands with its own **Creator / Social Pro** subscription tier.
3. **Central Master Landing Page (`yourdomain.com`)** showcasing both products with distinct feature walkthroughs and pricing tiers.
4. Core shared foundation (`FND-001` through `FND-016`: Auth, Workspaces, Vault, Audit Ledger, and Deterministic AI Gateway) is reused across products.

Feature selections below are **engineering recommendations from inspected documentation**, not measured winners. No live benchmarks or application/post/message executions were performed. Preserve unique ideas, but distinguish useful requirements from faulty defaults, unsupported features and unsafe implementation methods.

## 2. Required outcomes

| ID | Requirement |
| --- | --- |
| REQ-001 | Exactly two flagship commercial products: Product 1 (Career & LinkedIn) and Product 2 (SocialSuite), with independent subscription tiers under a unified master landing page. |
| REQ-002 | User-first onboarding: upload PDF/DOCX or enter profile manually; autosave and review extraction. |
| REQ-003 | Canonical user-confirmed profile, ATS-friendly master resume, immutable tailored versions and download. |
| REQ-004 | Optional LinkedIn connection/own-profile import, optimization suggestions and consistency diff. |
| REQ-005 | Job discovery/match/apply preparation and verified Applied marking, including manually applied jobs. |
| REQ-006 | Google Sheets one-time export and reliable one-way sync; CSV available independently. |
| REQ-007 | Separate Super Admin, workspace Admin and User privilege/dashboards with server-side checks. |
| REQ-008 | Versioned packages, subscriptions, feature entitlements, metering, billing and usage visibility. |
| REQ-009 | No fixed action count is advertised as safe from platform bans; unavailable actions fail closed. |
| REQ-010 | Connection/follow/DM/recruiter caps, explanatory tooltips and account-wide combined budgets. |
| REQ-011 | Equivalent platform-specific Social limits, permissions, consent, duplicate protection and stop controls. |
| REQ-012 | Keep all source ideas traceable with an adopted/merged/deferred/rejected disposition; no silent feature drops. |
| REQ-013 | Chosen stack: Next.js, TypeScript, Tailwind, Go, PostgreSQL, Redis, Meilisearch; no mandatory Python/PHP backend. |
| REQ-014 | One common open-source/self-hosted core and paid managed cloud product; avoid divergent application codebases. |
| REQ-015 | Human approval for external side effects; bind approval to actor/recipient/content hash/version and expiry. |
| REQ-016 | Never invent qualifications, metrics, eligibility answers, publication success or application submission. |
| REQ-017 | PostgreSQL is authoritative; Redis and search indexes are recoverable, scoped and non-authoritative. |
| REQ-018 | Strict private-profile isolation within a workspace; admin access to a workspace does not imply access to personal resumes. |
| REQ-019 | Shared exports, search, evidence, content calendar, accounts, action policy and notification services. |
| REQ-020 | Accessible responsive dashboards, guided empty states, progress, review inbox and actionable errors. |
| REQ-021 | Platform permissions, billing entitlements, user authorization and operational limits are separate gates. |
| REQ-022 | Five persistent planning files; stable task IDs, evidence-backed completion, resumable coder handoff. |
| REQ-023 | User data portability, consent/revocation, retention, deletion propagation and scoped audit trail. |
| REQ-024 | Capability and source evidence clearly distinguish existing source claims, planned concepts and our new work. |

## 3. User types and privilege boundaries

| Capability | Super Admin (platform) | Admin (workspace) | User |
|---|---|---|---|
| Manage global packages/pricing versions | Yes, audited | No; select a subscription only | No |
| Manage provider policy/kill switches | Yes; cannot manufacture provider permission | Lower workspace limits/pause assigned accounts | Lower own limits/pause own accounts |
| Create/invite/remove workspace members | Manage platform lifecycle | Own workspace within seat limit | No, unless a separate permission is delegated |
| Change platform roles | Super Admin only; step-up authentication | Cannot grant Super Admin | No |
| Read personal resumes/career facts | Not by default; explicit time-bound support consent only | Not by default; explicit owner delegation only | Own records |
| Manage shared brand assets/calendar | Support access only when approved | Own workspace | According to resource assignment |
| Approve another user's shared content | Only a delegated/audited review scope | Yes when assigned | Only when explicitly an approver |
| Connect personal external accounts | Not on behalf of an owner without authorization | Their own or delegated workspace account | Own account |
| Export personal data | Approved support scope only | Shared/delegated data only | Own data |
| Billing/invoices | Global operational administration | Workspace subscription | View allowed billing/usage; personal plan if billing owner |
| Global tenant/queue/cost diagnostics | Yes, redacted | Own workspace only | Own runs and usage |
| Bypass policy/provider cap | Never | Never | Never |

Implementation: `platform_role` and workspace `memberships.role` are distinct scopes. A user may be Admin in one workspace and User in another. Go authorizes every request and worker action. UI hiding is not security. Super Admin creation has no public signup route; use a one-time audited bootstrap. Require MFA/step-up for privileged changes. Support sessions are time-limited, explicitly consented, visible and logged; never reveal provider tokens or passwords.

## 4. Dashboard requirements

### 4.1 User dashboard

Top navigation exposes Career / LinkedIn / Social. Global controls: workspace switcher, scoped search, review inbox, notifications, connection status, usage/budget indicator and Stop all actions. A contextual AI assistant is available without replacing explicit forms and tables.

Career home shows the next onboarding step, resume readiness issues, saved/matching jobs, confirmed application counts, follow-up reminders and export status. LinkedIn home shows profile improvement suggestions, draft/review queues, relationship tasks and available/blocked capabilities. Social home shows drafts, scheduled work, failed/needs-confirmation publishing, research briefs, content performance and pending approvals.

Useful shared additions: global search and saved filters; list/board/calendar views; bulk **internal** organization (tag/archive), but not blind bulk sending; draft autosave/version history; reusable templates; accessible keyboard controls; responsive/mobile layouts; timezone/language preferences; notification preferences; dark/light theme; customizable/dismissible nonessential widgets; scoped activity/audit history; connection diagnostics; data portability/deletion. Counts must come from real data, with useful empty states instead of fake demo metrics in production.

### 4.2 Workspace Admin dashboard

Team/seat management; roles and assigned resources; shared content/calendar; approval routing; connected account owner and permission state; workspace quotas and spend forecast; failed jobs/retry review; package/subscription/invoices; exports; audit log; data-retention settings within platform bounds. Keep users' private career data out of team analytics unless they explicitly share it.

### 4.3 Super Admin dashboard

Workspace/subscription lifecycle, versioned package editor, trial/discount configuration, billing event reconciliation, aggregate service health, queues/dead letters, provider capability/policy registry, model/provider cost visibility, usage anomalies, feature rollouts, global pause controls, support-access audit, backup/restore status and release/migration status. Reports use redacted aggregate metadata rather than raw resume text or DMs.

## 5. Packages and entitlements

Create packages in draft → validate → publish immutable version → archive. Active subscriptions refer to a fixed version; changing a price or cap creates a new version rather than silently rewriting existing contracts. Preserve a migration/grandfathering policy.

Suggested labels (not approved price points): Community, Cloud Starter, Cloud Pro, Cloud Team. Community supplies self-hosted core and configurable user-owned infrastructure; cloud plans sell managed operations and resource allocations. Do not sell an increased chance of avoiding platform restrictions.

Package fields: internal ID, slug, display name, description, billing interval/currency/price, trial policy, feature flags, seat/workspace/social-account counts, document/storage limits, AI credits or budget, permitted model classes, discovery/research quotas, export schedules, media-generation quotas, concurrency ceilings, log retention and support level. Required quotas must be explicitly set before a package goes active; `null` is not implicitly unlimited. Monetary values use minor units plus currency; usage units are named and versioned.

Separate **feature entitlement**, **resource budget**, **user approval**, **provider capability** and **rate policy**. Payment cannot enable an unavailable provider scope, increase an official platform quota or override user consent. Show used/remaining/reserved quota and reset window. Reserve before expensive work; settle actual usage once; void reservations only for proven non-executed work. Signed/deduplicated billing webhooks drive subscription changes. Downgrades stop new over-quota work, not erase documents; users retain basic export and deletion. Use trial/active/past_due/cancel_at_period_end/canceled states with tested transitions.

## 6. Profile-first onboarding and ATS-friendly resume

### 6.1 Import or manual entry

Support PDF and DOCX, plus manual forms. Default upload limit is a proposed product limit of 10 MiB, configurable; it is not an ATS rule. Validate MIME/signature/size, reject executable/macro types, protect against ZIP bombs/path traversal, sandbox extraction and malware-scan uploads. Password-protected files require an unlocked copy. Do not assume a filename proves its type.

Try native text extraction first. For scanned PDFs, explain that OCR is needed; obtain consent and expose low-confidence fields for review. Original document, extraction text, parser/model version, source page/paragraph, per-field confidence and corrections are private versioned artifacts. Do not silently overwrite a confirmed profile when importing a later document.

Profile fields: name/contact/location; professional summary; work history and achievement bullets; education; skills and skill evidence; projects; certifications; publications/awards/volunteering; languages; portfolio/GitHub/LinkedIn URLs; target roles/industries/seniority; work mode/location; employment types; salary/currency/interval; exclusions; relocation/work-authorization/sponsorship preferences. Collect sensitive legal/eligibility facts only when needed and separately from public profile exports.

Forms autosave, allow reorder/add/remove and resume later, and work without AI. Dates preserve precision (year-only is not invented January 1). Missing information is `unknown`, not an empty fact or an AI guess. User review changes extraction state from draft to confirmed. Record fact provenance and confirmation date.

### 6.2 Resume builder and quality checks

Generate a canonical master resume from confirmed facts. Conservative default: one column, clear text headings, ordinary reading order, no essential facts in images, predictable dates and contact information in body text. Add selectable-text PDF, DOCX and TXT exports. Let users control sections, bullets, page breaks, language and version names. Preserve accessibility and text order.

Quality checks include parse-back comparison of contact/roles/dates/sections, clipping/overflow, missing facts, inconsistent dates, duplicate bullets, unclear language and job-specific keyword coverage. Separately label **profile completeness**, **document parse quality** and **job match**; never market them as a universal ATS pass score or interview probability.

AI may rewrite or reorder confirmed facts. It cannot invent years, employers, degrees, certifications, work eligibility, numerical achievements or skills. Suggested missing metrics are questions/placeholders until confirmed. A tailored resume has immutable links to the profile version, job snapshot, template version, prompt/model and approved changes. Keep original and tailored files.

### 6.3 LinkedIn connection is optional and capability-specific

Official OIDC sign-in returns lite profile/email information, not general access to work history, DMs or job application submission [P-LI-OIDC]. Show separate connection capabilities. Full-profile optimization works from an approved read source or user-provided LinkedIn export/copied content when full API access is unavailable. Do not make users paste account passwords or raw cookies into ordinary cloud onboarding.

Profile scan outputs: headline/About/Experience/Featured/skills improvements; missing-information requests; source-to-resume inconsistencies; before/after diff; reason and confidence. User approval is required. Copy/paste into LinkedIn is a first-class action where editing is unsupported. A scan without data must say what was unavailable, not fabricate a complete audit.

## 7. Jobs, applications and exports

### 7.1 Opportunity model

Normalize discovery sources into one job schema; keep provider job ID, canonical URL, raw snapshot hash, captured_at, posted_at, source and extraction confidence. Salary uses minimum/maximum, currency and period, preserving unknowns. Never interpret a missing salary as zero. Keep both original salary and any explicitly documented normalization. Match hard constraints before weighted relevance; explain unknowns vs mismatches. C1's displayed weights sum to 90 despite a stated 0–100 scale; normalize deliberately and test rather than copying that discrepancy.

Master profile/resume, confirmed eligibility rules and job snapshot prepare an application. Offer manual mode first and reviewed automation only for explicitly permitted, validated connectors. A job URL opening or a submit button click is not proof of application.

### 7.2 Status truth

Separate opportunity stage from action execution. Display **Applied** only after a supported submission receipt/success response or explicit user confirmation. Record `confirmation_type=provider_confirmed|user_confirmed`; a user can add applications made outside the product. Status changes are append-only events with actor/time/reason.

If a timeout occurs after a possible external write, set `needs_confirmation`, retain a reservation and investigate; do not retry submission automatically. Block duplicate application attempts for the same owner and canonical job. Later stages (interview/offer/rejected/withdrawn) require actual imported/provider/user evidence, not AI inference. Never display Viewed unless an actual source supplies it.

### 7.3 Exports

CSV: jobs, applications, permitted relationships and analysis tables. Include stable IDs, source URL, status, match version, applied_at, confirmation_type, resume version and notes where authorized. Protect CSV formula injection, preserve Unicode and quote/newline behavior. Do not include credentials or sensitive answers by default.

Google Sheets: per-user OAuth with appropriate least-privilege access, explicit spreadsheet/tab selection, preview/field mapping, one-time export or opt-in scheduled **one-way** sync. Use application ID as stable external key. Resolve existing IDs before writes, reconcile unknown writes and never blindly append on retry. Row sorting/deletion must not corrupt the mapping. Use plain-value/RAW writes for untrusted strings. Store sync cursor and errors durably; revoked access pauses sync, not application tracking. Google Sheets is a projection, not the database [P-SHEETS].

Private files are shared through authenticated routes or short-lived links; do not put permanent public resume URLs into spreadsheets. Export jobs, download audit and revoke/delete behavior are shared across modules.

## 8. Account action policy, limits and tooltip copy

### 8.1 No-ban guarantee is not a valid product promise

LinkedIn explicitly warns against unauthorized scraping/automated website activity [P-LI-AUTOMATION, P-LI-UA]. Lower volume, delays, a proxy or an open-source script do not turn it into an authorized integration. X also imposes API, consent and action-specific automation requirements [P-X]. The product must reduce operational risk without promising immunity.

**Cloud default: unsupported/unauthorized automated action = disabled, execution cap 0.** Keep drafts, native-site links, manual completion and personal reminder budgets available. Turn on a connector action only after actual capability/access review, user authorization and tests. A browser being technically able to click is not permission.

### 8.2 Proposed initial budgets — product design, not platform-safe thresholds

The nonzero values below are conservative **internal planning/operational defaults**, not official maximums and not a safety guarantee. Manual-native actions outside our app cannot be completely measured or blocked; label counters “recorded in this app” and allow a user-reported adjustment. Unapproved automatic sending remains zero regardless of the planning limit.

| Action | Initial planning / allowed-operation budget | Automatic execution default | Extra rule |
|---|---:|---|---|
| LinkedIn connection requests | 5/day; 20/rolling 7 days | 0 unless independently approved capability | Pending-recipient dedupe, no auto re-invite. |
| LinkedIn follows | 5/day | 0 unless independently approved capability | No repeated follow/unfollow cycles. |
| LinkedIn direct messages | 10/day total | 0 unless independently approved capability | Personal + recruiter + campaign DMs share a counter. |
| Recruiter outreach | 3/day, inside the 10-DM total | 0 unless independently approved capability | Not a separate allowance that adds 3 extra DMs. |
| Job application preparation/manual tasks | 10/day | Unapproved automated submission: 0 | This is preparation capacity; execution depends on destination. |
| Social scheduled publishing | 5 posts/account/day after authorized connection | Enabled only for a supported official/approved publish capability and approved content | Provider/app/account caps and content policy may be lower. |
| Social follow actions | 5 manual planning tasks/day | 0 by default | Per-provider enablement required; no indiscriminate growth loops. |
| Social comment/reply drafts | 10/day | 0 by default | Automated replies require provider permission, consent and any special approval. |
| Social opt-in message workflows | 5/account/day if permitted | 0 until permitted + consent verified | Unsupported/unsolicited bulk sending stays disabled. |
| Automated likes / fabricated engagement | Not offered | 0 | Keep legitimate draft/review workflows instead. |
| Research reads | 100 permitted-source items/day starter budget | 0 for disallowed source | Package quota and source API limits both apply. |

Backend permission is the first gate. Then enforce the minimum of package allowance, platform policy, provider/app/account allowance, workspace cap and user cap. All must permit the action. A provider cap that is unknown is not infinite permission. For verified connectors with no published numeric cap, an explicitly approved policy defines a conservative internal budget; never infer it from marketing claims.

Use rolling 24-hour and seven-day windows where appropriate; provider endpoint windows may be per-minute or different. The UI can display “today” in the user's timezone but cannot reset provider counters at local midnight. Reconnecting, switching modules/workspaces or changing a plan must not reset an account-wide budget. Keep recruiter outreach inside message budgets. Include pending reserved operations when calculating remaining allowance.

### 8.3 Tooltips and banners (Bangla UI copy)

| UI element | Tooltip / message |
|---|---|
| Connection limit | “এটি এই অ্যাপের নির্ধারিত সীমা, LinkedIn-এর নিরাপদ বা অনুমোদিত দৈনিক সীমা নয়। কম request পাঠালেও restriction হতে পারে।” |
| Follow limit | “এই সীমা শুধু অ্যাপে রেকর্ড হওয়া কাজের জন্য। বারবার follow/unfollow করবেন না; platform warning পেলে কাজ বন্ধ করুন।” |
| Direct message limit | “সব campaign ও recruiter message একই messaging limit-এর মধ্যে গণনা হবে। প্রাপকের সম্মতি ও platform permission প্রয়োজন।” |
| Recruiter outreach | “এটি Direct Message limit-এর অংশ—আলাদা অতিরিক্ত message allowance নয়।” |
| Automated action disabled | “এই action-এর অনুমোদিত integration পাওয়া যায়নি। Draft তৈরি করে platform-এ নিজে review ও complete করুন।” |
| Safety notice | “কোনো tool, proxy বা daily limit account block/suspend না হওয়ার নিশ্চয়তা দেয় না।” |
| Counter coverage | “এই সংখ্যা শুধু আমাদের অ্যাপে রেকর্ড হওয়া কাজ দেখায়; platform-এ সরাসরি করা সব কাজ এখানে দেখা নাও যেতে পারে।” |
| Social publishing | “Publish করার আগে account permission, media format ও platform limit পরীক্ষা করা হবে। Schedule করা মানেই publish নিশ্চিত নয়।” |
| Needs confirmation | “Submit-এর ফল নিশ্চিত করা যায়নি। Duplicate এড়াতে আবার submit না করে আগে status যাচাই করুন।” |
| Pause | “এই account-এর pending action বন্ধ আছে। Warning বা permission সমস্যার সমাধান না হওয়া পর্যন্ত কাজ চালু হবে না।” |
| Package upgrade | “Package upgrade resource quota বাড়াতে পারে; platform permission বা account safety limit পরিবর্তন করে না।” |

### 8.4 Enforcement and recovery

The Go API and each worker recheck permission, entitlement, resource ownership, approval hash, account status and budget immediately before acting. Atomic reservations prevent concurrent overrun; the authoritative action ledger survives Redis loss. If the policy/counter service is unavailable, external writes fail closed. A proven non-executed failure may release reservation; an ambiguous result cannot.

Honor Retry-After and bounded backoff on rate errors. Permission/restriction/authentication errors are not ordinary retryable failures. Challenge/2FA/reauth states pause the affected account and notify its owner; do not retry login endlessly, rotate identities or bypass CAPTCHAs. Stop follow-ups on reply, opt-out or user cancellation. Re-check queued actions after any account or policy change. Show per-account pause reason, evidence, observed_at, next action and reconnect controls instead of an invented “ban probability” score.

## 9. Feature sources and selection

Legend: **Documented** means an inspected upstream document describes it; **Planned** is not implemented evidence; **Edition-dependent** needs per-file verification; **NEW** is our own requirement. All code paths still need implementation audit and fixture/live-permission validation. The primary picks below are design references, not measured performance winners.

### 9.1 Career — source-to-feature matrix

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| CAR-01 | Upload PDF/DOCX or manual forms | [C2](#source-c2), [C3](#source-c3), [C4](#source-c4) | NEW unified ingestion/profile service | Text extraction, field provenance, draft/review/confirm; original file retained privately. |
| CAR-02 | Complete user-controlled career profile | [C2](#source-c2), [C4](#source-c4), [C6](#source-c6) | NEW canonical profile | Contact, experience, education, skills, projects, certificates, languages, links; field-level corrections and consent. |
| CAR-03 | Career preferences and exclusions | [C1](#source-c1), [C2](#source-c2), [C6](#source-c6), [C7](#source-c7) | C6 configuration + C7 normalized filters | Roles, countries, work mode, salary/currency, job type, sponsorship preferences; never guess eligibility. |
| CAR-04 | ATS-friendly master resume | [C2](#source-c2), [C3](#source-c3) | NEW renderer; C2/C3 workflow references | PDF + DOCX + TXT, selectable text, conservative single-column templates, deterministic section order, independent parse-back checks. |
| CAR-05 | Job-specific resume and cover letter | [C2](#source-c2), [C3](#source-c3) | C3 tailoring + C2 document workflow | Use confirmed facts only; diff/approve; preserve immutable source/job/resume versions. |
| CAR-06 | Resume feedback and skills demand | [C3](#source-c3) | C3 | Separate market-demand suggestions from actual possessed skills; evidence and dates attached. |
| CAR-07 | Multi-board discovery | [C7](#source-c7), [C8](#source-c8) | C7; C8 abstraction/history ideas | Source adapters; canonical URLs; normalized salary interval/location; per-board capability tests. |
| CAR-08 | Age/remote/type/company filters | [C2](#source-c2), [C6](#source-c6), [C7](#source-c7), [C8](#source-c8) | C7 + C6/C8 | Honor source-specific filter limitations; post-filter where valid; display unsupported filters. |
| CAR-09 | Explainable job match | [C1](#source-c1), [C3](#source-c3), [C5](#source-c5) | C1 rules + separately evaluated LLM explanation | Hard eligibility gates, configurable weights summing to 100, unknown vs mismatch, score version and reason. |
| CAR-10 | Saved jobs, dedupe, exclusion history | [C5](#source-c5), [C7](#source-c7), [C8](#source-c8) | NEW transactional engine + C8 ideas | Deduplicate discoveries and applications separately; C5 is a proposed design, not a working donor. |
| CAR-11 | Human-reviewed application workflow | [C2](#source-c2), [C4](#source-c4), [C6](#source-c6), [L1](#source-l1) | C6 UX + C4 state-machine concepts | Review material/answers; native/manual application remains available; provider approval governs any automated execution. |
| CAR-12 | Question and field recognition | [C3](#source-c3), [C4](#source-c4), [C6](#source-c6) | C4 field model + C3 contextual logic | Text/select/radio/checkbox/file inputs; conditional fields; unknown answers become needs_input, never guessed defaults. |
| CAR-13 | Multi-step/Shadow DOM/loop detection | [C4](#source-c4) | C4 | Fixture-tested form state machine; screenshots scrubbed; no blind next/submit click loops. |
| CAR-14 | External and Indeed applications | [C3](#source-c3) | C3 reference; experimental adapter family | No universal support claim. Unsupported domains use open-in-browser/manual status flow. |
| CAR-15 | Application status and verified Applied mark | [C1](#source-c1), [C2](#source-c2), [C5](#source-c5), [C8](#source-c8) | NEW PostgreSQL ledger + C2/C8 patterns | Submission receipt or explicit user confirmation; timed-out submissions require reconciliation. |
| CAR-16 | Per-application evidence/document bundle | [C2](#source-c2), [C3](#source-c3), [C4](#source-c4) | C2 + C4 | Job snapshot, exact resume/letter/answers, time, confirmation type, provider reference, scrubbed artifacts. |
| CAR-17 | Google Sheets export/sync | [C4](#source-c4) | C4 workflow; NEW per-user OAuth sync | Use platform-owned application IDs, idempotent upsert/reconciliation, one-way sync first; no Sheets-as-database. |
| CAR-18 | CSV and additional exports | [C2](#source-c2), [C7](#source-c7), [C8](#source-c8) | C7/C8 data + C2 reports | Jobs, applications, contacts, analysis; formula-injection protection; UTF-8; resume PDF/DOCX separate. |
| CAR-19 | Daily report and reminders | [C1](#source-c1), [C3](#source-c3) | C1 digest + C3 notification ideas | Application follow-ups, incomplete profile items, interviews; scheduled user-controlled notifications. |
| CAR-20 | Hiring-post and recruiter leads | [C1](#source-c1) | C1 | Only permission-supported/imported professional data; review before contact; shared LinkedIn relationship record. |
| CAR-21 | Run controls and recovery | [C1](#source-c1), [C3](#source-c3), [C4](#source-c4), [C6](#source-c6), [C8](#source-c8) | NEW durable jobs + cited patterns | Pause/cancel/retry safe stages, diagnostics, immutable history, crash recovery; uncertain writes never blindly retried. |
| CAR-22 | PII controls and LLM provider choices | [C3](#source-c3) | NEW privacy boundary + C3 reference | Explicit provider consent; minimize fields; encryption; tested redaction; no claim of complete anonymity. |
| CAR-23 | Interview preparation and outcomes |  | NEW | Company brief, questions, user notes, reminders and outcome funnel; no invented interview/recruiter signals. |
| CAR-24 | Resume/Profile/LinkedIn consistency check | [L2](#source-l2), [L3](#source-l3), [L4](#source-l4) | NEW comparison service | Show mismatched titles/dates/skills with source and precision; user decides which fact to retain. |

**Career final pick:** C7 for discovery contracts; C1 for explainable scoring/digests; C3 + C2 for tailoring/documents; C4 + C6 for form-state/manual-review patterns; C4 for Sheets workflow; C8 for historic status/recovery ideas. Build profile ingestion, ATS rendering/validation, authoritative application ledger and exports as new shared Go/TypeScript services. C5 remains a planned-source reference, not a runtime winner.

### 9.2 LinkedIn — source-to-feature matrix

The six requested placements are L1–L5 plus **L6 = C1 (`career-agent`)**; this is not a new unique repository.

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| LI-01 | Connect and inspect granted capabilities | [L3](#source-l3), [L5](#source-l5) | NEW official connector + L3/L5 contracts | Separate identity sign-in, data-read, publishing, messaging and apply access; show denied/unavailable scopes. |
| LI-02 | Own-profile scan/import | [L3](#source-l3), [L4](#source-l4) | L4 typed schema + L3 section selection | Approved API where available, otherwise user-supplied profile export/text; OIDC alone is insufficient. |
| LI-03 | Profile optimization | [L2](#source-l2), [S4](#source-s4) | L2 | Headline/About/Experience/Featured/skills suggestions; before/after diff; source facts; copy/apply only with user approval and supported API. |
| LI-04 | Person/company/job/post records | [L3](#source-l3), [L4](#source-l4) | L4 records + L3 tool boundaries | Versioned normalized snapshots and progress; retain permitted fields only. |
| LI-05 | Professional people/company discovery | [L1](#source-l1), [L3](#source-l3) | L3 queries + L1 Boolean builder | Professional criteria, company filters and explicit user goals; no protected-attribute scoring. |
| LI-06 | Recruiter/hiring lead workspace | [C1](#source-c1), [L3](#source-l3) | C1 discovery + NEW relationship service | Source-linked leads, notes, follow-up reminder and related applications. |
| LI-07 | Connection note drafts and queue | [L1](#source-l1) | L1 drafting + NEW approvals | Daily/rolling-week budgets, recipient dedupe, manual-native execution unless approved connector exists. |
| LI-08 | Company-follow planning | [L1](#source-l1) | L1 queue concept | Watchlist, batch planning, per-account limits; execution subject to platform permission. |
| LI-09 | Inbox/conversation search and reply drafts | [L3](#source-l3), [C3](#source-c3) | L3 tool contracts + C3 triage ideas | Read own authorized inbox, draft, review, correct thread targeting, stop follow-up on replies; capability-gated. |
| LI-10 | Comments, replies and thread sweep | [L1](#source-l1), [L2](#source-l2) | L2 intelligence | Thread-aware suggestions and selected replies; no bulk auto-engagement. |
| LI-11 | Post writing, hooks and audits | [L2](#source-l2) | L2 | Original facts, angle library, hook templates, editorial checks; no reach/virality or AI-detector guarantee. |
| LI-12 | Humanizer and reusable voice | [L2](#source-l2), [S4](#source-s4) | L2 + S4 | Voice preferences and readability; forbid invented personal anecdotes/metrics. |
| LI-13 | Story bank/interviewer | [L2](#source-l2) | L2 | User-approved stories and quantified achievements with provenance; separate from objective career facts. |
| LI-14 | Content planning and repurposing | [L2](#source-l2) | L2 | Plan/draft LinkedIn content here; use the same canonical Social publisher, asset and calendar records. |
| LI-15 | Engagement monitoring and analytics | [L2](#source-l2), [L1](#source-l1) | L2 + official/imported data | Author replies, relevant professional segments, actual metric snapshots; missing analytics stays unavailable. |
| LI-16 | Employee advocacy | [L2](#source-l2) | L2 | Optional team content approvals/brand governance; no coordinated fake engagement. |
| LI-17 | Provider fallback and diagnostics | [L5](#source-l5) | L5 | Capability registry, last check, expected scopes, errors; fallback only among independently permitted paths. |
| LI-18 | Limits, CAPTCHA/reauth handling, activity | [L1](#source-l1), [S3](#source-s3) | NEW shared policy engine + L1/S3 concepts | Fail closed; stop affected account, notify user, require legitimate reauth; never bypass challenges. |
| LI-19 | Session lifecycle and optional local tools | [L3](#source-l3), [L4](#source-l4) | L3 concepts; NEW secure service boundary | Per-owner isolated storage, revocation, short-lived viewer if approved; no credential/cookie paste as default cloud connection. |
| LI-20 | CSV/Sheets relationship exports | [C4](#source-c4), [C7](#source-c7) | NEW shared export service | Private, scoped, explicit destination; notes and contacts require ownership/authorization. |

**LinkedIn final pick:** L2 for profile/content intelligence; L4 for typed profile data; L3 for tool contracts/session lifecycle references; L1 for drafting/queue/feedback UX; L5 for provider diagnostics; C1 for hiring leads. The actual cloud transport must be an approved available connector or user-import/manual-native flow. None of these repositories is proof of LinkedIn authorization.

### 9.3 Social — six required feature groups

#### Publishing / Management

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-P01 | Account/provider registry, OAuth, publish | [S1](#source-s1) | S1 | Official/provider-permitted transport; available scopes and media support shown per account. |
| SOC-P02 | Calendar, queues and timezone-aware schedule | [S1](#source-s1), [S2](#source-s2), [S15](#source-s15) | S1 core; S2/S15 UX | One shared publisher; per-platform attempts, canonical timestamps, DST tests, receipt verification. |
| SOC-P03 | Media library and post variants | [S2](#source-s2), [S15](#source-s15) | S15 + own management UI | Original/derived assets, variants, thumbnails, alt text, version-aware approvals; S2 advanced edition availability unverified. |
| SOC-P04 | Templates, hashtag groups, variables, conditional comments | [S2](#source-s2) | NEW + S2 documented concepts | Track these requirements even if not in Lite source; publish follow-up comments only when permission allows. |
| SOC-P05 | Teams, reviews, approval and notifications | [S1](#source-s1), [S7](#source-s7) | S1 + S7 | Assigned approvers; changes invalidate approval; notify without resending content. |
| SOC-P06 | API/SDK/webhooks/automation integrations | [S1](#source-s1), [S3](#source-s3) | S1 | Versioned Go API; signed webhooks; idempotency; n8n/Make examples; secrets excluded. |
| SOC-P07 | Analytics/UTM/content outcome reports | [S1](#source-s1), [S2](#source-s2), [S15](#source-s15) | S1 metrics + S15 learning | Provider-specific metric definitions; time windows and data freshness; no unsupported cross-platform metric equivalence. |
| SOC-P08 | Publishing risk and preflight checks | [S15](#source-s15) | S15 + NEW policy | Format, ownership, consent, links, platform policy; block unsupported format/permission rather than silently degrade. |

#### AI Content Studio

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-C01 | Voice, newsletter voice, brand/account profile | [S4](#source-s4), [S15](#source-s15) | S4 + S15 | Interview and samples; identity, audience, tone, language, prohibited claims; user confirms profile. |
| SOC-C02 | URL/source-to-post with edit/approve | [S7](#source-s7), [S4](#source-s4) | S7 | Evidence-bound draft for LinkedIn/X; source-text hygiene and human review. |
| SOC-C03 | Hooks, writing frameworks, content matrices | [S4](#source-s4), [L2](#source-l2), [S15](#source-s15) | S4/L2 | Hook alternatives, PAS/AIDA/BAB/STAR/SLAY-style structures, pillars x formats; no fabricated anecdotes. |
| SOC-C04 | Graphics, carousels, quotes, thumbnail briefs | [S4](#source-s4), [S15](#source-s15) | S4 briefs + S15 production design | Separate draft prompt from completed media asset; versioned brand logo, alt text, approval. |
| SOC-C05 | Long-form, script, outline, storyboard, series | [S15](#source-s15), [S4](#source-s4) | S15 | Articles, newsletters, reels, video scripts, campaign/livestream/collab plans; advanced fiction/short-drama remain optional tracked scope. |
| SOC-C06 | Image editing and visual output family | [S15](#source-s15) | S15 | Resize/crop/compress/watermark, background removal, cards, posters, diagrams, comparison cards, mind maps, data reports; provider or deterministic renderer required. |
| SOC-C07 | Audio/video production family | [S15](#source-s15) | S15 | TTS, licensed voices/music, denoise/mix, transcription/subtitles/translation, clipping/highlights, reframing, chapters, intros/outros, slideshow/beat sync, conversions. |
| SOC-C08 | Generative media and personal voice | [S15](#source-s15) | S15 reference; NEW consent gates | Optional image/video/music model adapters. Voice cloning only with verified consent/rights; not available by default. |
| SOC-C09 | Editing, repurposing and platform adaptation | [S4](#source-s4), [S5](#source-s5), [S15](#source-s15) | S15 | Reuse one approved source across native formats; no blanket duplicate auto-posting. |
| SOC-C10 | Content QA, persona check, performance memory | [S4](#source-s4), [S15](#source-s15) | S15 + S4 | Editorial scores distinct from performance predictions; learn only from real permitted metrics and user feedback. |
| SOC-C11 | Regional formats and publishing adapters | [S15](#source-s15) | S15 requirements registry | Xiaohongshu/Douyin/Kuaishou/Zhihu/Bilibili/WeChat variants retained in backlog; each needs independent API/access validation. |

#### Research / Intelligence

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-R01 | Outlier posts vs creator baseline | [S5](#source-s5) | S5 | Normalize baseline/window; preserve source metrics, sample size and caveats. |
| SOC-R02 | Transcript intelligence and comment mining | [S5](#source-s5) | S5 | Claims, questions, objections and content ideas with traceable sources; quote limits and privacy filters. |
| SOC-R03 | Competitor/creator strategy and content gaps | [S5](#source-s5), [S15](#source-s15) | S5 + S15 | Comparative evidence, positioning, native formats, public business data; separate inferred insights. |
| SOC-R04 | Ad-library research | [S5](#source-s5) | S5 | Public allowed ad sources; archive dates/creative/source; no claim of ad spend or conversions without data. |
| SOC-R05 | Influencer prospects and audience fit | [S5](#source-s5) | S5 | Public professional/brand metrics, confidence, manual outreach notes; no sensitive individual inference. |
| SOC-R06 | Listening and product demand | [S5](#source-s5) | S5 | Brand/topic mentions, pains/questions, sentiment caveats; cited brief and CSV. |
| SOC-R07 | Trend/release watches and alerts | [S8](#source-s8), [S5](#source-s5) | S8 monitoring + S5 research | Provider-authorized collection, source dedupe, novelty thresholds; Slack/Discord/email user opt-in. |
| SOC-R08 | RSS/news/events/algorithm updates/UGC | [S15](#source-s15), [L5](#source-l5) | S15 + L5 | Timestamped evidence; distinguish policy announcements from speculation; UGC reuse needs permission. |
| SOC-R09 | Research-to-content handoff | [S5](#source-s5), [S7](#source-s7), [S15](#source-s15) | NEW shared evidence/asset IDs | Never lose citation lineage when drafting, repurposing or exporting. |

#### Outreach / Growth

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-G01 | Step workflows and working hours | [S3](#source-s3) | S3 | Durable state, waits, stop conditions, timezone; reviewed content and permission-gated actions. |
| SOC-G02 | Per-account serialization and budgets | [S3](#source-s3), [L1](#source-l1) | S3 concepts + NEW policy service | One account cannot multiply its quota across campaigns/modules/workspaces. |
| SOC-G03 | Professional lead enrichment waterfall | [S3](#source-s3) | S3 | Permitted provider sequence; provenance, cache, duplicate handling, deletion and cost controls. |
| SOC-G04 | Consent, suppression and follow-up stop | [S3](#source-s3) | NEW control layer | No unsolicited automated bulk DMs. Stop on reply/opt-out/rejection/permission change. Recruiter outreach shares messaging counters. |
| SOC-G05 | Outreach drafts, reminders and outcomes | [L2](#source-l2), [C1](#source-c1), [S3](#source-s3) | NEW UX + listed references | Manual-native sending is a first-class fallback, with sent status clearly user-confirmed vs API-confirmed. |

#### Scraping Infrastructure

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-S01 | HTTP/browser fetch with structured sessions | [S10](#source-s10) | S10 concepts | Approved domains/permissions; isolated TS browser worker, Go fetchers; private-address/SSRF checks. |
| SOC-S02 | Adaptive selectors and link extraction | [S10](#source-s10) | S10 | Fixture-based selector recovery; no automatic recovery for final-submit controls without validation. |
| SOC-S03 | Crawl checkpoints, streaming, replay and throttling | [S10](#source-s10), [C8](#source-c8) | S10 + C8 history | Durable PostgreSQL state; bounded workers; Retry-After; robots by default where applicable; do not evade denial. |
| SOC-S04 | Structured AI extraction | [S9](#source-s9) | S9 | Schema validation, source spans, confidence, cost budget, local/cloud model interface. |
| SOC-S05 | Markdown, feed/sitemap and dataset exports | [S10](#source-s10) | S10 | Sanitized source documents; JSON/JSONL/CSV/XML; source provenance and delete propagation. |
| SOC-S06 | Remote browser/CDP and permitted XHR capture | [S10](#source-s10) | S10 reference | Optional isolated execution for allowed integrations; no public browser-debug port or credential exposure. |
| SOC-S07 | Provider catalog and fallback health | [S6](#source-s6), [L5](#source-l5) | L5 runtime + S6 catalog reference | Catalog entry != installed/working/authorized provider. Curate, pin and test each adapter. |

#### OSINT / Identity Discovery

| ID | Feature | Compare sources | Primary selection | Implementation requirement |
| --- | --- | --- | --- | --- |
| SOC-O01 | Owned/authorized username and brand footprint | [S11](#source-s11), [S13](#source-s13), [S14](#source-s14) | S11 baseline + S13 lightweight | Domain/site allowlist and authorization; discovered account is a candidate match, never identity proof. |
| SOC-O02 | Detection evaluation and uncertainty | [S14](#source-s14), [S11](#source-s11) | S14 testing patterns | Frozen response corpus, holdout results, precision/recall, false-positive review; avoid unsupported accuracy claims. |
| SOC-O03 | Metadata reports, graphs and export | [S11](#source-s11), [S14](#source-s14), [S12](#source-s12) | S11/S14; S12 report concept | Professional/owned data only; reviewer-confirmed links; no face-based identity search. |
| SOC-O04 | Watch, diff, checkpoints, site extensions | [S14](#source-s14), [S13](#source-s13) | S14 + S13 | Explicit user opt-in, limited scope, notifications, resumable jobs; beta features require validation. |
| SOC-O05 | Historical face-correlation capability | [S12](#source-s12) | EXCLUDED from implementation; recorded | Retain requirement disposition so it is not silently forgotten. Replace with owner-supplied links/manual verification, not biometric identity matching. |
| SOC-O06 | Sensitive inference and private-data enumeration | [S11](#source-s11) | EXCLUDED from implementation; recorded | Do not infer age/ethnicity/health/etc. or enumerate private emails/phones; never use this module for candidate screening. |


**Social final pick:** S1 for publishing/provider patterns; S15 for workspace/media/learning loop; S4 for brand voice/editorial skills; S5 for research; S7 for human review/resume; S8 for trend watches; S3 for workflow serialization/working hours/enrichment; S10 for permitted crawling; S9 for structured extraction; S11 + S14 for optional authorized footprint/detection evaluation; S13 for lightweight scan/restore ideas. S2 advanced UX is edition-dependent and may be new work. S6 is a catalog, not a runtime. S12 face matching is not an implementation pick.

### 9.4 Advanced and unique ideas must remain visible

Keep these as explicit later requirements, not invisible omissions: custom/local LLM provider selection; Telegram/Slack/Discord/webhook notifications; HTML/XLSX/JSON reports; per-job screenshots; browser loop detection; official/imported inbox triage; professional story bank; team advocacy; newsletter-specific voice; source-to-draft tracing; conditional post variants; historical CSV recovery; provider diagnostics; RSS/news/event watchlists; public ad research; baseline-aware outlier detection; site/plugin catalog; frozen response-corpus evaluation; asset/template versioning; consented personal voice; audio/video editing/conversion; short dramas/fiction/paper explainers; regional content formats and publisher adapters.

The Easel capability map describes a much larger skill inventory than its brief README examples. Audit and map every actual skill folder before declaring its inventory complete: [Easel skill-function map](https://github.com/ZJU-REAL/Easel/blob/main/docs/skill-function-mapping.md). Its document was inspected in portions, not every executable skill. Publication or media-generation capabilities still depend on configured providers and allowed access.

For every candidate feature record: upstream repository + commit SHA + file/function path, evidence level, unique sub-features, chosen target capability ID, adopted/merged/deferred/rejected disposition, reason, implementation task and acceptance test. **Release gate: 100% of the audited inventory has a disposition; 0 orphan requirements.** This is a coverage rule, not a claim that all upstream code has already been audited.

### 9.5 Behaviors deliberately not copied

C4's unknown-skill “5 years”/unknown yes-no “Yes” defaults are prohibited. Blanket stealth/anti-bot bypass, account recycling, auto-solving challenges, fake engagement and bulk unsolicited messaging are not product features. C3 partial anonymization is not a privacy guarantee. S12 facial identity correlation, sensitive demographic inference and private contact enumeration are excluded; manual owner-provided links and professional/brand footprint review replace those uses. These decisions are logged rather than silently removing the source from the inventory.

## 10. Evidence catalog — all 28 unique repositories

Every entry has a direct source link. A `blob SHA` fingerprints the retrieved document, NOT the repository's commit or the source-code audit. Pin an actual commit during source audit. Descriptions are summaries; there are no claims of live benchmark performance.

<a id="source-c1"></a>

### C1 — [Rishal6/career-agent](https://github.com/Rishal6/career-agent)

- Evidence: [README.md](https://github.com/Rishal6/career-agent/blob/046be84933a2f1b2c7d12f0cbcb81f13354266fc/README.md); inspected 2026-09-17.
- Pinned Commit SHA: `046be84933a2f1b2c7d12f0cbcb81f13354266fc` (Date: 2026-04-06T10:27:13Z, tree `c7910cab15f7a90a3cfbec7bdb82700af5789c62`).
- Status: **Audited & Pinned; macOS-specific automation with multi-portal scanner**.
- Document blob SHA: `9024a0aef20a41390fae24610067be346e4ad790`.
- Audited Files:
  - `agent/portal_scanner.py`: Public JSON API scrapers for Greenhouse (`boards-api.greenhouse.io`), Ashby (`api.ashbyhq.com`), and Lever (`api.lever.co`).
  - `agent/resume_matcher.py`: Keyword-based score matching across 6 static resume variants.
  - `agent/job_scorer.py`: 0-100 weighted fit score based on title match, skills ratio, and exclusions.
  - `agent/auto_apply.py`: Multi-step Easy Apply automation using AppleScript/JavaScript injection into active Google Chrome tab.
  - `agent/feed_scanner.py`: LinkedIn feed parsing for keyword-based hiring posts (`#hiring`, `dm me`).
  - `agent/chrome.py`: macOS-specific `osascript` executing JS via `/tmp/career_agent.js`.
  - `agent/daily_post.py`: Generates technical LinkedIn posts using Claude on AWS Bedrock.
  - `agent/daily_report.py`: Aggregates summary of discovered jobs, applications, and recruiter leads.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Public Portal Scanners | `agent/portal_scanner.py` (`fetch_greenhouse`, `fetch_ashby`, `fetch_lever`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-07` | Clean, robust public REST APIs requiring no private credentials or brittle browser DOM scraping. |
  | Explainable Job Scorer | `agent/job_scorer.py` (`score_job`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-09` | Weight-based multi-factor scoring (title, skills, location); normalized into our 100% eligibility contract with explicit unknown vs mismatch labeling (`AT-003`, `AT-028`). |
  | Feed Hiring Lead Finder | `agent/feed_scanner.py` (`extract_posts`, `HIRING_KEYWORDS`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-20` | Extracting recruiter hiring posts; incorporated into permission-gated lead discovery. |
  | Crash-Safe Job Loop | `agent/auto_apply.py` (try/except per job) | Implemented | **ADOPT / MERGE** | `REQ-017`, `AT-021` | Per-job failure isolation ensures single submission errors do not fail the full job run. |
  | Static Resume Variant Matcher | `agent/resume_matcher.py` (`match_resume`, `DEFAULT_RESUMES`) | Implemented | **MERGE / ADAPT** | `IMP-CAR-03` | Replace static hardcoded variants with verified profile facts and dynamic ATS resume tailoring. |
  | macOS AppleScript Automation | `agent/chrome.py` (`run_js`, `osascript`) | Implemented | **REJECT** | None | Fragile, macOS-only, insecure `/tmp` execution, lacks multi-tenant safety. Replaced with isolated Playwright and Go HTTP fetchers. |
  | Unreviewed Easy Apply Submit | `agent/auto_apply.py` (`click_easy_apply`, `fill_form_fields`) | Implemented | **REJECT** | `REQ-015` | Violates human-in-the-loop approval invariants (`REQ-015`, `AT-005`, `AT-007`). Submissions require explicit review token before dispatch. |
  | Direct Clipboard Post Copy | `agent/daily_post.py` (`pbcopy`) | Implemented | **REJECT** | `REQ-015` | Content must flow through approval inbox and scheduled outbox rather than raw clipboard. |


<a id="source-c2"></a>

### C2 — [attdobi/aipply](https://github.com/attdobi/aipply)

- Evidence: [README.md](https://github.com/attdobi/aipply/blob/9f90c1d84863b31256eb4ecd02c21787cf7b0e27/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `9f90c1d84863b31256eb4ecd02c21787cf7b0e27` (Date: 2026-03-25T06:14:51Z, branch `feat/initial-setup`, tree `8b2a88cde081de17a3ab6053d565a1e9da2af152`).
- Status: **Audited & Pinned; Playwright-based Easy Apply with DOCX tailoring and AI de-slop filter**.
- Document blob SHA: `a789839c3370d349dd766624c2c7ef6d8510edd0`.
- Audited Files:
  - `src/resume_tailor.py`: Surgical `.docx` document cloning replacing targeted summary and core competencies while preserving fonts, bolding, and margins.
  - `src/cover_letter_gen.py`: Formatted `.docx` cover letter generator matching candidate style.
  - `src/deslop.py`: AI telltale cleaner stripping robotic cliches ("leverage my", "excited to", "synergy", em-dashes).
  - `src/linkedin_applicant.py`: Playwright browser automation for LinkedIn Easy Apply with persistent user profile and CDP connection (`localhost:9333`).
  - `src/linkedin_scanner.py`: Job search filter and description extractor.
  - `src/tracker.py`: Application tracking ledger persisting records in `tracker.json` with screenshots.
  - `scripts/dashboard.py`: Terminal UI dashboard displaying application status, daily quotas, and log streams.
  - `scripts/scan_and_apply.py`: Unified pipeline for scanning, tailoring, and submitting.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Surgical DOCX Tailoring | `src/resume_tailor.py` (`clone_and_tailor`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-04` | Surgical paragraph replacement in DOCX preserves user typography, margins, and formatting without full document re-render. |
  | AI De-Slop / Humanizer | `src/deslop.py` (`clean_text`, `REPLACEMENTS`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-12`, `IMP-SOC-C03` | Cleans robotic AI phraseology and formatting quirks from generated resumes, letters, and social drafts. |
  | Structured Cover Letter | `src/cover_letter_gen.py` (`save_cover_letter`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-05` | Clean single-page cover letter document generation matching candidate profile facts. |
  | Playwright CDP Browser | `src/linkedin_applicant.py` (`connect_browser`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-13` | Chrome DevTools Protocol (CDP) connection to isolated browser context; integrated into our task-worker. |
  | Application Evidence Ledger | `src/tracker.py` (`add_application`, `screenshots`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-15`, `AT-005` | Preserves submission receipts, timestamped screenshots, and per-job tailored documents. |
  | Unreviewed Easy Apply Submit | `src/linkedin_applicant.py` (`apply_to_job`) | Implemented | **REJECT** | `REQ-015` | Automated final submit without explicit human review violates core invariant `REQ-015` (`AT-005`, `AT-007`). |
  | Unsafe Form Guessing | `src/linkedin_applicant.py` (auto-filling answers) | Implemented | **REJECT** | `AT-003` | Must fail closed to `needs_input` on sensitive or unconfirmed questions rather than auto-guessing defaults. |

<a id="source-c3"></a>

### C3 — [beatwad/LinkedIn-AI-Job-Applier-Ultimate](https://github.com/beatwad/LinkedIn-AI-Job-Applier-Ultimate)

- Evidence: [README.md](https://github.com/beatwad/LinkedIn-AI-Job-Applier-Ultimate/blob/release/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `34331c06214b8e0dca1aaefc90d4eb0f2c771a58` (Date: 2026-09-04T06:49:44Z, branch `release`, tree `da9ffad2a2ca4d88e0dc56ee434771c17caefc1a`).
- Status: **Audited & Pinned; Dual-board (LinkedIn & Indeed) auto-applier with resume generator, PII anonymization, and monitoring dashboard**.
- Document blob SHA: `3bb990f5a2e653496d84e2c2840b1d14afdb2b72`.
- Audited Files:
  - `src/job_manager/resume_anonymizer.py`: Replaces names, emails, phone numbers, and URLs with gender-matched dummy fixtures prior to LLM submission.
  - `src/job_manager/linkedin/easy_applier_linkedin.py`: Playwright-based multi-step LinkedIn Easy Apply workflow handling file uploads, text, radio, and select inputs.
  - `src/job_manager/indeed/easy_applier_indeed.py`: Playwright-based Indeed apply flow with CAPTCHA detection and container transitions.
  - `src/llm/apply_agent.py`: Experimental `browser-use` agent integration for non-Easy Apply vacancies.
  - `src/resume_builder/resume_generator.py`: Generates vacancy-tailored resumes using HTML/CSS templates (Faangpath, Cloyola, Josylad) and markdown output.
  - `src/job_manager/linkedin/messages_manager_linkedin.py`: LinkedIn inbox scanner, star-filtering, and dry-run auto-responder using LLM classification.
  - `src/dashboard/data_service.py` & `src/dashboard/server.py`: Local web dashboard for live monitoring, run statistics, and config editing.
  - `src/telegram/telegram_manager.py`: Real-time Telegram notification client sending run summaries and CAPTCHA alert screenshots.
  - `src/utils/runtime_control.py`: Graceful shutdown coordinator across threads/processes with draining, cleanup, and browser lifecycle controls.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Resume PII Redaction / Anonymization | `src/job_manager/resume_anonymizer.py` (`anonymize_text`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-22`, `REQ-023` | Replace client PII (names, contact, social links) with synthetic tokens before passing untrusted context to external LLMs. |
  | Multi-Style HTML/CSS Resume Generation | `src/resume_builder/resume_generator.py` (`create_resume`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-04`, `IMP-CAR-05` | High-fidelity ATS templates (single-column, clean CSS) with deterministic markdown compilation. |
  | Multi-Step Form Field Recognition | `src/job_manager/linkedin/easy_applier_linkedin.py` (`_fill_additional_questions`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-12` | Structured detection of text/numeric, select, radio, and file inputs across multi-step application wizards. |
  | Run Draining & Graceful Recovery | `src/utils/runtime_control.py` (`RuntimeController`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-21`, `REQ-017` | Clean phase-based shutdown (`running` -> `draining` -> `cleanup` -> `done`) avoiding dangling browser sessions or orphaned lock files. |
  | Inbox Triage & Reply Drafting | `src/job_manager/linkedin/messages_manager_linkedin.py` (`classify_message`) | Implemented | **ADOPT / MERGE** | `IMP-LI-09` | Read-only thread evaluation and LLM-assisted draft replies gated behind capability checks. |
  | Telegram / Webhook Alerts | `src/telegram/telegram_manager.py` (`send_captcha`, `send_report`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-19`, `FND-013` | Channel notifications for run summaries, quota warnings, and manual intervention requests. |
  | Unreviewed Easy Apply Submit | `src/job_manager/linkedin/easy_applier_linkedin.py` (`_uncheck_follow_company_and_submit`) | Implemented | **REJECT** | `REQ-015` | Fully automated final submission without explicit user review token violates core safety rule `REQ-015`. |
  | Blind Autonomous Web Surfing (`browser-use`) | `src/llm/apply_agent.py` (`ApplyAgent.run`) | Experimental / Beta | **REJECT** | `REQ-015`, `AT-010` | Autonomous unconstrained web agents on external third-party portals risk non-deterministic form submission and account compromise. |
  | CAPTCHA Auto-Solver / Evasion | `src/job_manager/indeed/easy_applier_indeed.py` (bypass routines) | Implemented | **REJECT** | `REQ-021`, `AT-010` | Evasion of platform rate-limits or bot checks violates platform safety and compliance rules; system must fail closed and alert user. |

<a id="source-c4"></a>

### C4 — [Eliasjakob/ai-job-application-agent](https://github.com/Eliasjakob/ai-job-application-agent)

- Evidence: [README.md](https://github.com/Eliasjakob/ai-job-application-agent/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `66ec9038457e2afee848db887f24c92b41fc4b50` (Date: 2026-05-16T20:41:32Z, branch `main`, tree `3373449339097ea63b519c8f2b34aeb354157bc1`).
- Status: **Audited & Pinned; Shadow-DOM Easy Apply agent with pattern-based question answering, rolling limiter, and Google Sheets logger**.
- Document blob SHA: `f5ebbd86088eb55da833e7a8b9af2e5b664c0295`.
- Audited Files:
  - `agent/apply.py`: Playwright browser automation for LinkedIn Easy Apply traversing modern Shadow DOM (`#interop-outlet`), multi-step form state progression, rolling submission rate limiter (`count_recent_submissions`), and pattern-based question answering (`infer_answer`).
  - `agent/sheets.py`: Google Sheets integration using service account credentials (`service_account.json`) logging application timestamp, company, title, job URL, and submission status.
  - `agent/parse_cv.py`: Lightweight CV parser using `pypdf` extracting candidate email, phone, postal code, and candidate name to seed profile configuration.
  - `agent/config.py`: Application parameters controlling keyword/location search, distance radius, headless toggle, dry-run mode, and submission caps.
  - `tests/test_infer_answer.py` & `tests/test_label_to_key.py`: Pure-logic unit tests for question parsing, multilingual skill matching (German/English), and token-based field resolution.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Shadow DOM & Form State Machine | `agent/apply.py` (`deepAll`, `_find_active_step_container`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-13` | Modern LinkedIn modal uses Shadow DOM (`#interop-outlet`); transparent traversal ensures robust form step detection without brittle selectors. |
  | Rolling Hourly Submission Limiter | `agent/apply.py` (`count_recent_submissions`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-21`, `FND-011` | Sliding window rate-limiting based on log/ledger timestamps prevents rapid burst submissions and quota overrun. |
  | Google Sheets Application Logger | `agent/sheets.py` (`SheetsClient.append_application`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-17` | One-way projection of applied jobs into user-owned Google Sheets with idempotency and error isolation (`P-SHEETS`). |
  | Pure-Logic Field Resolution Tests | `tests/test_infer_answer.py` (`test_experience_*`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-12` | Deterministic unit tests for token matching across multilingual field labels without needing live browser/network. |
  | Basic PDF CV Contact Extraction | `agent/parse_cv.py` (`extract_text`, `parse_full_name`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-01` | Baseline contact heuristic parser for PDF uploads, followed by mandatory user confirmation (`AT-001`). |
  | Unsafe Default Skill Years Guessing | `agent/apply.py` (`infer_answer` -> `default_years_of_experience` / '5') | Implemented | **REJECT** | `AT-003` | Defaulting unknown skills to '5 years' or arbitrary profile defaults is prohibited. Must fail closed to `needs_input`. |
  | Unsafe Yes/No Bias Guessing | `agent/apply.py` (`infer_answer` -> "Yes" default for unknown skills/questions) | Implemented | **REJECT** | `AT-003` | Blindly answering "Yes" to unknown requirements or visa/sponsorship questions is unsafe; requires explicit human confirmation. |
  | Direct Service Account Key Storage | `agent/sheets.py` (`service_account.json`) | Implemented | **REJECT** | `REQ-021`, `AT-010` | Hardcoded service account files in local repo violate least-privilege; replaced with user-consented OAuth tokens stored in secure vault. |

<a id="source-c5"></a>

### C5 — [ziash/linkedin-job-bot](https://github.com/ziash/linkedin-job-bot)

- Evidence: [README.md](https://github.com/ziash/linkedin-job-bot/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `3c379a5ad605987fd15a51dd2c0ecd2347e15d76` (Date: 2026-05-17T09:58:37Z, branch `main`, tree `60da7f91c9cb4a52fc997bc618e4768ea9c81b67`).
- Status: **Audited & Pinned; Scaffolding / Concept phase only (single README.md file)**.
- Document blob SHA: `3c2b421f0451d74dfbb4fac2ab3b131b9a95423d`.
- Audited Files:
  - `README.md`: Documents planned system architecture (`run.py`, `scraper.py`, `matcher.py`, `applier.py`, `tracker.py`, `tracker.db`, `config.yaml`), planned integration with Claude API (`claude-haiku-4-5`), proposed 1-10 fit score with explanation, and semi-automatic shortlist review flow.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Semi-Automatic Shortlist Review Step | `README.md` (`[review step]`) | Documented / Planned | **ADOPT / MERGE** | `IMP-CAR-11`, `REQ-015` | Pre-application human approval gate for job shortlists reinforces core invariant `REQ-015`. |
  | Explainable Fit Score (1-10) + Reason | `README.md` (`matcher.py` concept) | Documented / Planned | **ADOPT / MERGE** | `IMP-CAR-09` | Explainable scoring with explicit rationale adopted and normalized into our 0-100 rubric (`AT-028`). |
  | Tailored Cover Letter per Job | `README.md` (`matcher.py` concept) | Documented / Planned | **ADOPT / MERGE** | `IMP-CAR-05` | Custom cover letter generation based on matching job description facts. |
  | Deduplication Tracker Ledger | `README.md` (`tracker.py` concept) | Documented / Planned | **ADOPT / MERGE** | `IMP-CAR-10`, `IMP-CAR-15` | Prevention of duplicate applications for the same canonical job URL via persistent ledger. |
  | Unimplemented Scaffolding Python Modules | `README.md` (unimplemented files list) | Planned / Non-existent | **DEFER / EXCLUDE** | N/A | Codebase contains 0 executable python scripts; cannot serve as runtime code donor. Concepts adopted above. |
  | Unattended "Auto" Mode Mode Switch | `README.md` (`mode: auto` concept) | Documented / Planned | **REJECT** | `REQ-015` | Fully automated unattended apply mode violates human review requirement `REQ-015` (`AT-005`, `AT-007`). |

<a id="source-c6"></a>

### C6 — [GodsScion/Auto_job_applier_linkedIn](https://github.com/GodsScion/Auto_job_applier_linkedIn)

- Evidence: [README.md](https://github.com/GodsScion/Auto_job_applier_linkedIn/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `e0b2401a7a7b333eab0b518e80be5285c3ee85c5` (Date: 2026-09-10T17:10:30Z, branch `main`, tree `035882e379ff8c9501d5ce675c9635e9f8feea3d`).
- Status: **Audited & Pinned; Selenium-based LinkedIn Easy Applier with local Flask control panel, question presets, and strict config validation**.
- Document blob SHA: `cec1f07f0f893dd602fdbbe287e4398e7cc7feec`.
- Audited Files:
  - `app.py`: Local Flask web application (`127.0.0.1` control panel) allowing non-technical users to inspect settings, view logs, edit application answers, and trigger application runs.
  - `runAiBot.py`: Core application runner orchestrating Selenium WebDriver, login session reuse, LinkedIn Easy Apply modal progression, and answer population.
  - `config/questions.py` & `config/personals.py`: Granular application questions configuration (legal name, contact, portfolio, salary, visa sponsorship, years of experience presets).
  - `config/search.py` & `config/settings.py`: Search filters (keywords, locations, job types, date posted, remote/hybrid, company blacklist).
  - `modules/validator.py`: Strict type and bounds checking validating config parameters before launching any browser run.
  - `modules/clickers_and_finders.py`: Resilient XPath locator helpers (`text_xpath`) ignoring case and whitespace differences in dynamic LinkedIn form labels.
  - `tests/test_question_matching.py` & `tests/test_selectors_groundtruth.py`: Groundtruth test fixtures for Easy Apply modal fields and button selectors.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Local Control Panel & Config UI | `app.py` (`Flask` control panel) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-014`, `IMP-CAR-11` | User-friendly web UI for viewing application state, configuring preferences, and managing credentials locally. |
  | Granular Questions & Personal Presets | `config/questions.py`, `config/personals.py` | Implemented | **ADOPT / MERGE** | `IMP-CAR-02`, `IMP-CAR-12` | Pre-confirmed user fact profiles covering common application questions (work auth, salary expectation, contact details). |
  | Strict Config Schema Validation | `modules/validator.py` (`check_int`, `check_bool`) | Implemented | **ADOPT / MERGE** | `FND-011`, `IMP-CAR-03` | Fail-closed validation verifying user preferences and boundaries before initiating job runs. |
  | Resilient Dynamic Field Matchers | `modules/clickers_and_finders.py` (`text_xpath`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-12` | Robust text-normalization matching for evolving LinkedIn question wording. |
  | Groundtruth Selector Test Fixtures | `tests/fixtures/` (`06_easy_apply_modal.html`) | Implemented | **ADOPT / MERGE** | `FND-016`, `IMP-CAR-12` | HTML snapshot fixtures for testing form extraction and button identification offline. |
  | High-Volume Auto-Apply Claims (100+/hr) | `README.md` ("100+ jobs in under an hour") | Documented | **REJECT** | `REQ-009`, `REQ-021` | High-volume burst spam triggers immediate platform bans; our system enforces conservative daily budgets (`AT-008`). |
  | Unreviewed Easy Apply Submit | `runAiBot.py` (auto-click submit) | Implemented | **REJECT** | `REQ-015` | Automated submission without explicit human review violates core invariant `REQ-015` (`AT-005`, `AT-007`). |
  | Blind Default Fallbacks for Visa / Exp | `config/questions.py` (`years_of_experience = "5"`) | Implemented | **REJECT** | `AT-003` | Unconfirmed fields must fail closed to `needs_input` rather than auto-submitting static arbitrary defaults. |

<a id="source-c7"></a>

### C7 — [speedyapply/JobSpy](https://github.com/speedyapply/JobSpy)

- Evidence: [README.md](https://github.com/speedyapply/JobSpy/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `fda080a373e8226f3fd60635323f5da9af9892b1` (Date: 2026-02-18T19:39:52Z, branch `main`, tree `fbe87bb4a595a882a884e1b8c199bb6bfb1063ca`).
- Status: **Audited & Pinned; Production-grade multi-board job discovery library across LinkedIn, Indeed, Glassdoor, Google, ZipRecruiter, Bayt, Naukri, and BDJobs**.
- Document blob SHA: `31e7564211a528f5f056eeb89ad978a7d8f975dd`.
- Audited Files:
  - `jobspy/model.py`: Canonical job post domain model (`JobPost`, `Location`, `Compensation`, `JobType`, `Country`, `DescriptionFormat`), ISO country codes, compensation normalization intervals (`yearly`, `monthly`, `weekly`, `daily`, `hourly`), and multi-language employment type parsers.
  - `jobspy/__init__.py`: Multi-threaded concurrent scraper orchestration (`scrape_jobs`) executing per-board adapters across thread pools with timeout and error handling.
  - `jobspy/linkedin/__init__.py`: Public HTTP guest API and web scraper for LinkedIn job listings with pagination, filter queries, and full description enrichment.
  - `jobspy/indeed/__init__.py`: Indeed public search API and HTML scraper with compensation interval parsing and direct employer apply link resolution.
  - `jobspy/util.py`: Salary regex extraction, currency parser, markdown/plain text formatters, and retry/session builders.
  - `jobspy/glassdoor/`, `jobspy/google/`, `jobspy/ziprecruiter/`: Individual board adapters with localized country subdomains.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Canonical Multi-Board Job Schema | `jobspy/model.py` (`JobPost`, `Location`, `Compensation`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-07`, `REQ-002` | High-value normalized schema unifying disparate board structures into one typed Go model (`title`, `company`, `location`, `compensation`, `job_type`, `date_posted`, `canonical_url`). |
  | Multi-Board Discovery Adapters | `jobspy/linkedin/`, `jobspy/indeed/`, `jobspy/google/` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-07` | Modular HTTP-based scrapers without heavy browser overhead for initial discovery indexing. |
  | Normalized Compensation Intervals | `jobspy/model.py` (`CompensationInterval`, `Compensation`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-07`, `IMP-CAR-08` | Standardizes min/max amount, currency, and period (hourly, monthly, annual) while explicitly preserving unknown salaries (never defaulting to 0). |
  | Multi-Language Job Type Parser | `jobspy/model.py` (`JobType` enum mappings) | Implemented | **ADOPT / MERGE** | `IMP-CAR-08` | Comprehensive multi-lingual term dictionary resolving variations of full-time, contract, and internship across regions. |
  | Direct Employer URL Resolution | `jobspy/model.py` (`job_url_direct`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-14` | Distinguishes aggregator portal links from direct employer careers page URLs for transparent manual redirect. |
  | Pandas DataFrame / CSV Exporter | `jobspy/__init__.py` (`jobs.to_csv`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-18` | Clean tabular CSV export with quote escaping and formula injection protection (`P-SHEETS`). |
  | Unbounded Concurrency Overruns | `jobspy/__init__.py` (`ThreadPoolExecutor`) | Implemented | **REJECT** | `REQ-009`, `FND-011` | Scraping must be strictly rate-limited and quota-bounded through Go task workers rather than launching unrestricted thread bursts. |
  | Unsupported Board Assumptions | `jobspy/bayt/`, `jobspy/naukri/` (regional variants) | Implemented | **DEFER** | `IMP-CAR-07` | Focus initially on core supported boards (LinkedIn, Indeed, Google, Greenhouse/Lever); secondary regional boards deferred to phase 2. |

<a id="source-c8"></a>

### C8 — [PaulMcInnis/JobFunnel](https://github.com/PaulMcInnis/JobFunnel)

- Evidence: [readme.md](https://github.com/PaulMcInnis/JobFunnel/blob/master/readme.md); inspected 2026-09-18.
- Pinned Commit SHA: `cdc0971b52cbc80057cef45b2048f68829aeabf8` (Date: 2025-12-10T03:01:02Z, branch `master`).
- Status: **Audited & Pinned; Formally ARCHIVED tool; high-value state tracking, deduplication and master CSV lifecycle patterns**.
- Document blob SHA: `3ac881fbe04dbc18bff0229cc6691b68f5d5a527`.
- Audited Files:
  - `jobfunnel/backend/jobfunnel.py`: Core orchestrator managing scraping runs, duplicate identification, filtering, historical cache recovery, and masterlist persistence.
  - `jobfunnel/resources/enums.py`: Explicit job status lifecycle model (`JobStatus`: `NEW`, `INTERESTED`, `APPLY`, `APPLIED`, `INTERVIEWING`, `INTERVIEWED`, `REJECTED`, `ACCEPTED`, `ARCHIVE`, `DELETE`), `Remoteness`, `Locale`, and `DuplicateType`.
  - `jobfunnel/resources/resources.py`: Master CSV headers, TF-IDF text similarity thresholding (`DEFAULT_MAX_TFIDF_SIMILARITY`), user-agent pools, and minimum description lengths.
  - `jobfunnel/backend/tools/filters.py`: Keyword blacklists, company blocklists, and title/description filter engines.
  - `readme.md`: Official project archive notice citing anti-bot/CAPTCHA barriers on legacy static scrapers.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Explicit Multi-Stage Job Lifecycle | `jobfunnel/resources/enums.py` (`JobStatus`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-15`, `REQ-002` | Typed status progression (`interested`, `applied`, `interviewing`, `rejected`, `archive`) perfectly matches our domain ledger. |
  | Historical Deduplication & Blocking | `jobfunnel/backend/jobfunnel.py` (`DuplicateType`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-10` | "Never see the same job twice" persistent blocklist and cross-search deduplication ledger. |
  | Master CSV Export & Field Schema | `jobfunnel/resources/resources.py` (`CSV_HEADER`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-18` | Standardized column definitions for external spreadsheet review and offline auditing. |
  | Text Similarity Match Thresholding | `jobfunnel/resources/resources.py` (`DEFAULT_MAX_TFIDF_SIMILARITY`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-10`, `AT-004` | Fuzzy deduplication across cross-board duplicates with slightly different titles or spacing. |
  | Legacy Static HTML Scrapers | `jobfunnel/backend/scrapers/` (historical scrapers) | Implemented (Archived) | **REJECT** | `REQ-012` | Static HTML scrapers are completely obsolete on modern anti-bot sites; replaced with `JobSpy` HTTP APIs and isolated Playwright CDP. |

<a id="source-l1"></a>

### L1 — [LucasSantana-Dev/linkedin-engage](https://github.com/LucasSantana-Dev/linkedin-engage)

- Evidence: [README.md](https://github.com/LucasSantana-Dev/linkedin-engage/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `413c5ff3d876b003e40fccb7848fd3546988efc4` (Date: 2026-06-26T03:32:19-03:00, branch `main`).
- Status: **Audited & Pinned; Manifest V3 Chrome Extension for LinkedIn networking, connection notes, company follow queues, and Easy Apply pre-fill**.
- Document blob SHA: `e9183f1a1ff4d605e6d6bb157bc8c32adfddfdcd`.
- Audited Files:
  - `extension/manifest.json`: Manifest V3 service worker configuration with scoped host permissions (`https://*.linkedin.com/*`).
  - `extension/background.js`: Central state machine handling serialized storage queues, quota guards (weekly 150 invites cap), rate-limiting backoff, CAPTCHA alerts, and tab coordination.
  - `extension/content.js`: DOM interaction script handling connection button clicks, custom note modal popups, company batch-follow queue, and Easy Apply pre-fill.
  - `extension/bridge.js`: Message passing relay connecting Chrome extension background worker with active content tabs.
  - `package.json` & `CHANGELOG.md`: Version 1.40.0 release log detailing career intelligence keyword boosting, input boundary validation, and duplicate application prevention.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Personalized Connection Note Drafting | `extension/content.js` (note composer) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-07`, `REQ-016` | AI-assisted drafting of contextual connection notes using candidate background and profile context. |
  | Target Company Batch-Follow Planning | `extension/content.js` (company follow queue) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-08` | Structured queue for organizing company target lists and planning follows within account budgets. |
  | Boolean Search Query Generator | `extension/background.js` (query builder) | Implemented | **ADOPT / MERGE** | `IMP-LI-05` | Generates advanced Boolean queries (`AND`, `OR`, `NOT`, `must-match`) for high-precision people and company discovery. |
  | Easy Apply Pre-Fill with Review Before Submit | `extension/content.js` (`launchJobsAssist`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-11`, `REQ-015` | Pre-fills form fields but stops before final submission, enforcing mandatory human review before execution. |
  | Client-Side Storage Concurrency Queue | `extension/background.js` (`_profileWalkCountQueue`) | Implemented | **ADOPT / MERGE** | `FND-011`, `REQ-017` | Serializes async updates to prevent race conditions during concurrent account operations. |
  | Weekly Invite Cap (150/wk) Platform Guarantee | `README.md` ("150 invites/week guard") | Documented | **REJECT** | `REQ-009`, `REQ-021` | High weekly invite limits do not protect against account restrictions; our system enforces conservative operational caps (`AT-008`). |
  | Automated Feed Auto-Reactions | `README.md` ("auto-react with category templates") | Deprecated / Beta | **REJECT** | `REQ-021`, `AT-010` | Indiscriminate automated reactions risk account flag; social interactions must be reviewed drafts. |

<a id="source-l2"></a>

### L2 — [sergebulaev/linkedin-skills](https://github.com/sergebulaev/linkedin-skills)

- Evidence: [README.md](https://github.com/sergebulaev/linkedin-skills/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `3eb227fb9d0ab328ba2e74c41559279e377964d7` (Date: 2026-09-15T16:31:40-04:00, branch `main`).
- Status: **Audited & Pinned; Production skill bundle for Claude Code/Codex with 12 specialized LinkedIn marketing skills, humanizer, story bank, and approval gates**.
- Document blob SHA: `3a8e7c04295afd2600bb816af8c569f6e8865a29`.
- Audited Files:
  - `skills/linkedin-post-writer/SKILL.md`: Long-form post drafting utilizing 20 canonical hook formulas and founder angle library.
  - `skills/linkedin-humanizer/SKILL.md` & `references/voice-rules.md`: Multi-tier AI slop removal, banned vocabulary blacklist, reveal bridges scrub, and voice profile generator.
  - `skills/linkedin-comment-drafter/SKILL.md`: URL-based third-party post comment drafting with 1-3 variants and reaction selection.
  - `skills/linkedin-reply-handler/SKILL.md`: Thread sweeping and two-level comment reply generation.
  - `skills/linkedin-content-planner/SKILL.md`: 7-day content planning taxonomy covering themes, audiences, and posting times.
  - `skills/linkedin-interviewer/SKILL.md` & `references/story-bank.md`: Structured founder interview methodology to build durable story banks of milestones, metrics, and turning points.
  - `skills/linkedin-profile-optimizer/SKILL.md`: 7-step headline, About, and Experience section optimizer.
  - `skills/linkedin-hook-extractor/SKILL.md` & `references/hook-formulas.md`: Reverse-engineering hook formulas from viral URLs.
  - `skills/linkedin-repurposer/SKILL.md`: Converting tweets, videos, and articles into fold-optimized LinkedIn posts.
  - `skills/linkedin-engager-analytics/SKILL.md`: ICP engager segmentation and roster extraction.
  - `lib/approval.py`: Explicit `render_approval_card` gating every outbound publication until user approval.
  - `lib/url_parser.py`: Robust parsing and normalization of LinkedIn post, profile, and company URLs.
  - `references/untrusted-content.md`: Indirect prompt injection defense rules isolating untrusted third-party LinkedIn content.
  - `scripts/check_no_secrets.py`: Static scanning for API tokens and credentials.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Hook-Driven Post Writer | `skills/linkedin-post-writer/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-01`, `IMP-SOC-C01` | High-performing 2026 hook formulas (anaphora, curiosity-gap, contrarian) for engaging post generation. |
  | Multi-Tier AI Humanizer & Slop Filter | `skills/linkedin-humanizer/SKILL.md`, `references/voice-rules.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-12`, `IMP-SOC-C03` | Forensic/strict/aesthetic AI cliché scrubbing and vocabulary blacklisting for authentic professional voice. |
  | Contextual Comment & Reply Drafting | `skills/linkedin-comment-drafter/`, `skills/linkedin-reply-handler/` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-06`, `IMP-LI-10` | Intelligent 1-3 comment drafting and thread sweep replies with reaction selection. |
  | 7-Day Content Planner & Pillars | `skills/linkedin-content-planner/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-03`, `IMP-SOC-P01` | Structured pillar-based weekly planning with daily targets and posting window heuristics. |
  | Story Bank & Founder Interviewer | `skills/linkedin-interviewer/SKILL.md`, `references/story-bank.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-04` | Guided interview framework capturing career turning points, scars, and metrics into structured memory. |
  | 7-Step Profile Optimizer | `skills/linkedin-profile-optimizer/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-11` | Systematic audit and rewrites for LinkedIn headline, About narrative, and Experience metrics. |
  | Untrusted Content Injection Defense | `references/untrusted-content.md` | Implemented | **ADOPT / MERGE** | `FND-015`, `AT-019` | Passive boundary isolation protecting AI gateway against malicious prompt injection in scraped comments/posts. |
  | Human-in-the-Loop Approval Card | `lib/approval.py` (`render_approval_card`) | Implemented | **ADOPT / MERGE** | `FND-010`, `REQ-015` | Enforces explicit preview and approval cards before any post, comment, or reply can be dispatched. |
  | Canonical LinkedIn URL Parser | `lib/url_parser.py` | Implemented | **ADOPT / MERGE** | `IMP-LI-05` | Normalizes tracking tags, URNs, and permalinks into clean canonical entity references. |
  | Viral Hook Extractor & Repurposer | `skills/linkedin-hook-extractor/`, `skills/linkedin-repurposer/` | Implemented | **ADOPT / MERGE** | `IMP-LI-02`, `IMP-SOC-C05` | Deconstructs viral formats and transforms cross-platform drafts into LinkedIn layout standards. |
  | Pre-Commit Secret Scanner | `scripts/check_no_secrets.py` | Implemented | **ADOPT / MERGE** | `FND-013`, `AT-016` | Static regex scanner ensuring tokens, keys, and session secrets are never committed or logged. |
  | External Commercial Publora/Pixfaro APIs | `lib/publora_client.py`, `lib/pixfaro_client.py` | Implemented | **DEFER** | `IMP-LI-01` | Platform uses native MinIO/S3 object storage and direct social publisher adapters rather than paid proxy SaaS. |
  | External Apify Actor Scraping | `lib/apify_client.py` | Implemented | **DEFER** | `IMP-LI-09` | Defer paid Apify actors in favor of user OAuth, Playwright CDP, or official API integrations. |
  | Team Employee Advocacy Engine | `skills/linkedin-employee-advocacy/SKILL.md` | Implemented | **DEFER** | `IMP-LI-13` | Enterprise team advocacy programs deferred to multi-workspace enterprise release phase. |
  | Unreviewed Automated Publishing | Publora direct auto-dispatch without approval | Deprecated / Beta | **REJECT** | `REQ-015`, `AT-007` | Zero unreviewed automated posts or comments allowed; human approval gate is strictly mandatory. |
  | AI Detector Evasion Guarantees | Marketing claims on beating AI detectors | Documented | **REJECT** | `REQ-016`, `AT-003` | Claims of guaranteed detector evasion are scientifically unsound; humanizer is an editorial quality tool. |


<a id="source-l3"></a>

### L3 — [stickerdaniel/linkedin-mcp-server](https://github.com/stickerdaniel/linkedin-mcp-server)

- Evidence: [README.md](https://github.com/stickerdaniel/linkedin-mcp-server/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `bddded1fc9de64f0c9466b0fa1c1259023930fac` (Date: 2026-09-17T11:47:36+08:00, branch `main`).
- Status: **Audited & Pinned; FastMCP server implementation exposing structured LinkedIn tools, local browser cookie session extraction, and daemon leasing**.
- Document blob SHA: `d866a4f9408fcda32454b50c0c6e8e8a5b28a9be`.
- Audited Files:
  - `linkedin_mcp_server/server.py`: FastMCP server entry point and tool registry.
  - `linkedin_mcp_server/tools/person.py`: Own-profile and public profile retrieval (`get_my_profile`, `get_person_profile`, `search_people`, `connect_with_person`).
  - `linkedin_mcp_server/tools/company.py`: Company profiles, posts, and employee search (`get_company_profile`, `get_company_posts`, `search_companies`).
  - `linkedin_mcp_server/tools/job.py`: Job search, job details, and saved jobs retrieval (`get_job_details`, `search_jobs`, `get_saved_jobs`).
  - `linkedin_mcp_server/tools/messaging.py`: Inbox triage, conversation search, and messaging (`get_inbox`, `get_conversation`, `search_conversations`, `send_message`).
  - `linkedin_mcp_server/tools/feed.py` & `post.py`: Feed content inspection and post search.
  - `linkedin_mcp_server/browser_import/`: Local browser session discovery and cookie decryption (`discovery.py`, `extract.py`, `orchestrate.py`).
  - `linkedin_mcp_server/daemon_lock.py` & `profile_lease.py`: Concurrency control, process leasing, and singleton browser daemon management.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Structured LinkedIn Tool Schemas | `linkedin_mcp_server/tools/` (all tools) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-05`, `IMP-LI-06`, `IMP-LI-09` | Standardized parameter contracts and return shapes for person, job, and messaging endpoints. |
  | Own & Public Profile Fact Extraction | `tools/person.py` (`get_my_profile`, `get_person_profile`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-LI-11`, `IMP-CAR-02` | Extracts structured experience, education, skills, and headline into verified profile facts. |
  | Local Browser Cookie Import Discovery | `browser_import/` (`discovery.py`, `extract.py`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `IMP-LI-05` | Safe local-first cookie detection avoiding risky credential sharing across network boundaries. |
  | Browser Process Leasing & Locks | `daemon_lock.py`, `profile_lease.py` | Implemented | **ADOPT / MERGE** | `FND-008`, `FND-011` | Ensures single-tenant sequential access per LinkedIn session, preventing concurrent session corruption. |
  | Saved Jobs & Native Job Search | `tools/job.py` (`search_jobs`, `get_saved_jobs`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-07` | Supplements external board scrapers with the user's personal saved jobs list on LinkedIn. |
  | Recruiter Inbox & Message Triage | `tools/messaging.py` (`get_inbox`, `get_conversation`) | Implemented | **ADOPT / MERGE** | `IMP-LI-09` | Ingests recruiter DMs into human review inbox without automated blind replies. |
  | FastMCP Daemon Process Runner | `server.py`, `daemon.py` | Implemented | **DEFER** | `IMP-LI-05` | Standalone Python MCP daemon deferred in favor of native Go worker architecture (`FND-007`). |
  | Company Employee Directory Scraper | `tools/company.py` (`get_company_employees`) | Implemented | **DEFER** | `IMP-LI-08` | Mass employee crawling deferred to enterprise talent intelligence release. |
  | Unreviewed Outbound Connections | `tools/person.py` (`connect_with_person`) | Implemented | **REJECT** | `REQ-015`, `AT-007` | Direct autonomous connection sending is prohibited; must require explicit approval card with personalized note (`IMP-LI-07`). |
  | Autonomous Direct Messaging | `tools/messaging.py` (`send_message`) | Implemented | **REJECT** | `REQ-015`, `REQ-009`, `AT-007` | Automated unreviewed messaging violates approval gates and account quotas (`AT-008`). |
  | Raw Cookie Upload Onboarding | Cookie string upload via web inputs | Documented / Untrusted | **REJECT** | `REQ-009`, `REQ-021`, `AT-016` | Raw cookie uploads to remote servers are security anti-patterns; sessions must be imported locally or vault-encrypted. |


<a id="source-l4"></a>

### L4 — [joeyism/linkedin_scraper](https://github.com/joeyism/linkedin_scraper)

- Evidence: [README.md](https://github.com/joeyism/linkedin_scraper/blob/master/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `b1cdc1c0e85bee8764d62565d229c682e5eb81bb` (Date: 2026-04-09T19:44:58-07:00, branch `master`, v3.1.2).
- Status: **Audited & Pinned; Production async Playwright scraper featuring typed Pydantic models, granular progress callbacks, and interactive 2FA challenge handling**.
- Document blob SHA: `32f1d6287fde8d3560c1f5947792016a4e258403`.
- Audited Files:
  - `linkedin_scraper/models/person.py`: Pydantic schema for LinkedIn personal profiles (`Person`, `Experience`, `Education`, `Contact`, `Interest`).
  - `linkedin_scraper/models/company.py`: Typed company profile schema (`Company`).
  - `linkedin_scraper/models/job.py`: Structured job posting model (`Job`).
  - `linkedin_scraper/models/post.py`: Post body, author, and engagement metrics model (`Post`).
  - `linkedin_scraper/scrapers/`: Async scrapers (`person.py`, `job.py`, `job_search.py`, `company.py`, `company_posts.py`).
  - `linkedin_scraper/core/auth.py`: `wait_for_manual_login` interactive 2FA handler and `warm_up_browser` routine.
  - `linkedin_scraper/core/browser.py`: Async Playwright browser context manager with stealth configurations.
  - `linkedin_scraper/callbacks.py`: Real-time async progress tracking (`on_start`, `on_progress`, `on_success`, `on_error`).
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Typed Profile Model Schema | `models/person.py` (`Person`, `Experience`, `Education`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-02`, `IMP-LI-11` | Normalized schema for storing confirmed user experiences, educations, and skills. |
  | Typed Job Posting Schema | `models/job.py` (`Job`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-07` | Clean structured model for job titles, descriptions, requirements, and applicants. |
  | Real-Time Scraper Progress Callbacks | `callbacks.py` (`ProgressCallback`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-008`, `REQ-019` | Direct bridge between scraper execution milestones and real-time SSE stream events. |
  | Interactive Manual 2FA Recovery | `core/auth.py` (`wait_for_manual_login`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `IMP-LI-05` | Allows human operator to complete 2FA or CAPTCHA checkpoints interactively without crashing runs. |
  | Typed Company Entity Schema | `models/company.py` (`Company`) | Implemented | **ADOPT / MERGE** | `IMP-LI-08` | Structured representation of company industry, size, headquarters, and specialties. |
  | Feed Post & Engagement Metrics | `models/post.py`, `scrapers/company_posts.py` | Implemented | **ADOPT / MERGE** | `IMP-LI-06` | Extracts post bodies and engagement counts to supply context for AI comment drafting. |
  | Browser Warmup Routine | `core/auth.py` (`warm_up_browser`) | Implemented | **ADOPT / MERGE** | `IMP-CAR-13` | Navigates baseline reference sites to establish realistic cache and cookies before target visits. |
  | Unauthenticated Public Profile Scraping | `scrapers/person.py` (anonymous mode) | Implemented | **DEFER** | `IMP-LI-11` | Anonymous scraping easily fails against LinkedIn authwalls; prioritize authenticated local sessions. |
  | Bulk Headless Job Search Scraper | `scrapers/job_search.py` | Implemented | **DEFER** | `IMP-CAR-07` | Job discovery is primarily handled by JobSpy HTTP scrapers; Playwright search retained as fallback. |
  | Plaintext Credential Auto-Login | `core/auth.py` (`login_with_credentials`) | Implemented | **REJECT** | `REQ-009`, `REQ-021`, `AT-016` | Storing or passing unencrypted passwords in plaintext is prohibited; must use AES-256-GCM vault. |
  | Unbounded Headless Scraping Bursts | Unthrottled scraping loops | Documented | **REJECT** | `REQ-009`, `FND-011` | High-frequency concurrent requests trigger account flags; must be governed by account quotas (`AT-008`). |


<a id="source-l5"></a>

### L5 — [Panniantong/agent-reach](https://github.com/Panniantong/agent-reach)

- Evidence: [docs/README_en.md](https://github.com/Panniantong/agent-reach/blob/main/docs/README_en.md); inspected 2026-09-18.
- Pinned Commit SHA: `a19a171fa980a0785849596492e0af4db800c82f` (Date: 2026-09-16T00:16:24+08:00, branch `main`).
- Status: **Audited & Pinned; Multi-platform read/access router featuring ordered primary/fallback backends, self-healing diagnostic doctor, and credential scrubbing**.
- Document blob SHA: `b15b3ce6a807af6980202c09a00b6d197f0a6ded`.
- Audited Files:
  - `agent_reach/core.py`: Core dispatcher orchestrating primary and fallback backends per channel.
  - `agent_reach/doctor.py`: Diagnostics health checker running isolated probes without cascading failure.
  - `agent_reach/probe.py`: Pre-flight connectivity probe assessing channel latency and rate limits.
  - `agent_reach/cookie_extract.py`: Local desktop cookie extraction helper.
  - `agent_reach/channels/`: Modular channel integrations (`linkedin.py`, `twitter.py`, `reddit.py`, `rss.py`, `web.py`).
  - `agent_reach/utils/text.py`: `scrub_url_credentials` ensuring sensitive tokens and passwords are redacted from URLs.
  - `docs/README_en.md`: Design philosophy on ordered backends and capability tiers.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Multi-Backend Fallback Routing | `core.py`, `channels/base.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `IMP-SOC-P01` | Ordered adapter priority (e.g. Official API -> Playwright -> RSS) maximizing retrieval reliability. |
  | Diagnostic Doctor Health Checker | `agent_reach/doctor.py` (`check_all`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-013`, `FND-009` | Isolated per-channel health checks where individual errors degrade gracefully without taking down platform. |
  | URL Credential & Token Scrubber | `utils/text.py` (`scrub_url_credentials`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-013`, `AT-016` | Redacts sensitive authentication tokens and basic auth credentials from URLs and logs. |
  | Pre-Flight Channel Probe | `agent_reach/probe.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `FND-011` | Verifies channel readiness and network health before dispatching queued background tasks. |
  | Local Desktop Cookie Extractor | `agent_reach/cookie_extract.py` | Implemented | **ADOPT / MERGE** | `FND-009` | Safe local-first cookie detection for user-supervised browser session onboarding. |
  | RSS & Public Web Reader Channel | `channels/rss.py`, `channels/web.py` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C02` | Public news and RSS feed ingestion to inspire cross-platform social content generation. |
  | Regional Chinese Platform Channels | `channels/bilibili.py`, `boss.py`, `xiaohongshu.py` | Implemented | **DEFER** | `IMP-SOC-P01` | Regional specific channels deferred to internationalization phase. |
  | Audio Transcription Subsystem | `agent_reach/transcribe.py` | Implemented | **DEFER** | `IMP-SOC-M01` | Deferred in favor of central media processing pipeline. |
  | Outreach Engine Misclassification | Automated outreach / cold DM engine | Rejected Concept | **REJECT** | `REQ-012` | Agent-reach is strictly an access/read router; automated cold messaging is prohibited. |
  | Permission-Bypass Fallbacks | Fallback bypassing permission checks | Rejected Architecture | **REJECT** | `REQ-007`, `AT-011` | A fallback backend must never bypass an explicit permission denial, role boundary, or quota block. |
  | Unauthenticated Proxy Hopping | Rotating proxy networks for evasion | Documented | **REJECT** | `REQ-021`, `AT-010` | Evading security checkpoints via proxy networks is prohibited; operations must follow rate limits. |


<a id="source-s1"></a>

### S1 — [gitroomhq/postiz-app](https://github.com/gitroomhq/postiz-app)

- Evidence: [README.md](https://github.com/gitroomhq/postiz-app/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `1207941bb9002743f7cc89846fe1c633f50c1a6a` (Date: 2026-09-17T23:20:39+07:00, branch `main`).
- Status: **Audited & Pinned; Production open-source social media management monorepo with 30+ official social adapters, OAuth PKCE flow, SSRF-safe HTTP dispatchers, and post-level analytics**.
- Document blob SHA: `f982792668b90cf3c4a3d43a3bde961e70a7b6d0`.
- Audited Files:
  - `libraries/nestjs-libraries/src/integrations/social.abstract.ts`: Abstract base class for social publishers with media normalization and retry logic.
  - `libraries/nestjs-libraries/src/integrations/social/social.integrations.interface.ts`: Standard interface (`IAuthenticator`, `GenerateAuthUrlResponse`, `AnalyticsData`, `PostDetails`).
  - `libraries/nestjs-libraries/src/integrations/social/`: 30+ provider implementations (`x.provider.ts`, `linkedin.provider.ts`, `linkedin.page.provider.ts`, `facebook.provider.ts`, `instagram.provider.ts`, `threads.provider.ts`, `reddit.provider.ts`, `youtube.provider.ts`, `pinterest.provider.ts`, `tiktok.provider.ts`, `bluesky.provider.ts`, `mastodon.provider.ts`).
  - `libraries/nestjs-libraries/src/integrations/refresh.integration.service.ts`: Automatic OAuth token refresh before dispatch.
  - `libraries/nestjs-libraries/src/dtos/webhooks/ssrf.safe.dispatcher.ts`: SSRF-safe HTTP request dispatcher blocking internal network addresses.
  - `libraries/nestjs-libraries/src/short-linking/`: Short-link generation with engagement click tracking.
  - `apps/backend/` & `apps/orchestrator/`: BullMQ/Temporal scheduling orchestrator.
  - `apps/sdk/`: Public API and TypeScript client SDK.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Standardized Social Provider Interface | `social.integrations.interface.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `IMP-SOC-P01` | Clean contract for OAuth PKCE URL generation, token refresh, and dispatch in Go/TS. |
  | Multi-Platform Official Social Adapters | `integrations/social/` (major platforms) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P01`, `IMP-SOC-P02` | LinkedIn personal/page, X/Twitter, Reddit, Facebook, Instagram, Threads, Bluesky, YouTube. |
  | SSRF-Safe Outbound Request Dispatcher | `ssrf.safe.dispatcher.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-013`, `FND-009` | Blocks private IP ranges (127.0.0.1, 10.0.0.0/8, 192.168.0.0/16, AWS metadata 169.254.169.254). |
  | Channel & Post-Level Analytics Pipeline | `social.integrations.interface.ts` (`analytics`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-A01`, `IMP-SOC-A02` | Standardized impressions, likes, reposts, comments, and engagement rate schema. |
  | Short-Link Click Tracking | `short-linking/` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-A01` | Measures real conversion and outbound link clicks on published social posts. |
  | Proactive OAuth Token Refresh Guard | `refresh.integration.service.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `REQ-017` | Proactively refreshes expiring OAuth tokens prior to scheduled background dispatches. |
  | Media Dimension & Duration Pre-Flight | `social.abstract.ts` (media processing) | Implemented | **ADOPT / MERGE** | `FND-006`, `IMP-SOC-M01` | Validates platform-specific image aspect ratios and video lengths before publishing. |
  | First-Comment Link Scheduling | `social.abstract.ts` (`comment`) | Implemented | **ADOPT / MERGE** | `IMP-SOC-P01` | Automatically schedules links in the first comment to preserve algorithmic feed reach. |
  | Web3 / Niche Social Providers | `farcaster.provider.ts`, `nostr.provider.ts` | Implemented | **DEFER** | `IMP-SOC-P01` | Niche decentralization protocols deferred to phase 4 community network expansion. |
  | Full NestJS/Prisma/Temporal Stack Import | Monorepo NestJS runtime | Implemented | **DEFER** | `FND-001` | Monorepo runtime not imported; persistence and queueing implemented in Go/PostgreSQL/Redis. |
  | Unreviewed Multi-Channel Auto-Blasts | Bulk cross-posting without preview | Documented | **REJECT** | `REQ-015`, `AT-007` | Simultaneous unreviewed publishing across channels is strictly forbidden; requires human approval card. |
  | Silent Token Expiry Failures | Unhandled expired token dispatch | Documented | **REJECT** | `REQ-017`, `AT-006` | Failures due to expired credentials must transition run to needs_input with reconnect prompt. |


<a id="source-s2"></a>

### S2 — [inovector/mixpost](https://github.com/inovector/mixpost)

- Evidence: [README.md](https://github.com/inovector/mixpost/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `df57648b866310446703f5294350552b62735df5` (Date: 2026-03-16T10:06:37Z, branch `main`).
- Status: **Audited & Pinned; Self-hosted social media management software (Mixpost Lite package) with multi-account batch publishing, per-account post variants, provider rate-limit backoff, and media asset conversions**.
- Document blob SHA: `0f6f04d0b7b72fbd7e5335279b28f6fdf15c85c2`.
- Audited Files:
  - `src/Actions/PublishPost.php`: Batch orchestrator (`Bus::batch`) mapping each account to `AccountPublishPostJob`, allowing partial failures and recording per-account error receipts.
  - `src/Jobs/AccountPublishPostJob.php`: Asynchronous per-account publisher with rate limit tracking (`HasSocialProviderJobRateLimit`) and graceful service status check.
  - `src/Concerns/Job/HasSocialProviderJobRateLimit.php`: Trait handling account-level (`mixpost-{account_id}-api-limit`) and app-level (`mixpost-{platform}-api-limit`) cached rate limits with 24-hour retry window.
  - `src/Models/Post.php`: Eloquent model tracking post status (`draft`, `scheduled`, `publishing`, `published`, `failed`), timezone-aware UTC timestamps, and account relations.
  - `src/Models/PostVersion.php`: Per-account customized post content model supporting original vs account-specific overrides and media attachments.
  - `src/SocialProviderManager.php`: Social provider factory managing Lite adapters (`TwitterProvider`, `FacebookPageProvider`, `MastodonProvider`).
  - `src/Models/Media.php` & `src/MediaConversions/`: Media library asset entity with image conversions (thumbnails, web-optimized previews).
  - `src/Models/Metric.php` & `src/Models/Audience.php`: Analytics data storage for audience metrics and post-level engagement.
  - `composer.json` & `README.md`: Package metadata establishing open-source Lite edition boundaries vs proprietary Pro/Enterprise claims (workspaces, team collaboration, advanced AI plugins).
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Per-Account Post Variants & Overrides | `src/Models/PostVersion.php` (`mixpost_post_versions`) | Implemented (Lite) | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P03`, `REQ-002` | Stores original base post + per-account tailored content (custom character limits, hashtags, and media attachments). |
  | Multi-Account Batch Publish Pipeline | `src/Actions/PublishPost.php` (`__invoke`) | Implemented (Lite) | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P01`, `REQ-017` | Batch-dispatches individual account jobs with isolated failure handling (`allowFailures`), setting granular status per account. |
  | Provider Rate-Limit Throttling Trait | `src/Concerns/Job/HasSocialProviderJobRateLimit.php` | Implemented (Lite) | **ADOPT / RE-IMPLEMENT** | `FND-009`, `REQ-017` | Cache-backed distinction between account-level and app-level rate-limit cooldowns, avoiding wasteful API calls. |
  | Media Asset Storage & Thumbnail Conversions | `src/Models/Media.php`, `src/MediaConversions/` | Implemented (Lite) | **ADOPT / MERGE** | `FND-006`, `IMP-SOC-P03` | Asset entity linked to posts with automated thumbnail and dimension generation. |
  | Timezone-Aware UTC Scheduling | `src/Models/Post.php` (`scheduledAt`) | Implemented (Lite) | **ADOPT / MERGE** | `IMP-SOC-P02` | Canonical UTC storage with explicit user timezone shift prevents DST scheduling drift. |
  | Post & Audience Metric Model | `src/Models/Metric.php`, `src/Models/Audience.php` | Implemented (Lite) | **ADOPT / MERGE** | `IMP-SOC-P07`, `IMP-SOC-A01` | Normalized post engagement and follower history schema across supported networks. |
  | Closed Pro/Enterprise Workspaces & Teams | Commercial Pro product claims | Proprietary / Absent | **DEFER / RE-IMPLEMENT** | `IMP-SOC-P05` | Multi-tenant workspaces and team approvals are closed-source in Mixpost; designed natively in our `FND-005` workspace model. |
  | Conditional Follow-up Comments & Templates | Commercial Pro product claims | Documented Concept | **MERGE / ADAPT** | `IMP-SOC-P04` | Follow-up comments and hashtag groups tracked as explicit requirements and designed natively in our publishing service. |
  | Unreviewed Multi-Account Auto-Post | `PublishPost.php` (immediate batch dispatch) | Implemented (Lite) | **REJECT** | `REQ-015`, `AT-007` | Unreviewed bulk publishing violates core invariant `REQ-015`; requires pre-dispatch approval token per target account. |
  | Synchronous Media Upload in Job Loop | Inline heavy asset uploads | Documented | **REJECT** | `FND-006`, `AT-020` | Media must be quarantined, validated, and processed asynchronously via object storage prior to job dispatch. |


<a id="source-s3"></a>

### S3 — [growchief/growchief](https://github.com/growchief/growchief)

- Evidence: [README.md](https://github.com/growchief/growchief/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `abb1e37a6f5595d8d105aef5871a2eeb0c22a1dc` (Date: 2026-03-24, branch `main`).
- Status: **Audited & Pinned; API-first social outreach automation and enrichment workflow monorepo with per-account concurrency serialization, working hours manager, and multi-provider enrichment waterfall**.
- Document blob SHA: `09e8504ef871f80930a24396f418c9add618020d`.
- Audited Files:
  - `apps/orchestrator/src/workflows/workflow.throttle.ts`: Temporal workflow ensuring per-account serialization, pacing jobs with 10-minute progress deadlines (`PROGRESS_DEADLINE`), and managing queue priority.
  - `apps/orchestrator/src/utils/working.hours.manager.ts`: Active working hours validator pausing step execution until user-configured local working hour windows open.
  - `shared/server/enrichment/`: Modular contact/lead enrichment waterfall supporting Apollo, Datagma, Hunter, and RocketReach (`apollo.enrichment.ts`, `datagma.enrichment.ts`, `hunter.enrichment.ts`, `rocket.reach.enrichment.ts`).
  - `shared/server/bots/providers/linkedin/linkedin.provider.ts`: LinkedIn automation provider with built-in detection for weekly invitation limits and rate-limit warnings.
  - `shared/server/bots/bot.manager.ts` & `cdp.detection.pass.ts`: Headful/headless browser manager with stealth anti-detection hooks (`ghost-cursor`, `patchright`).
  - `schema.prisma`: Data models for bot groups, workflows, workflow nodes, leads, activities, proxies, restrictions, and credits.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Per-Account Workflow Serialization & Pacing | `workflow.throttle.ts` (`enqueue`, `PROGRESS_DEADLINE`) | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-G01`, `SOC-G02`, `FND-011`, `REQ-017` | Enforces single-lane execution per account, preventing parallel runs from overwhelming platform quotas regardless of workflow count. |
  | Timezone-Aware Working Hours Manager | `working.hours.manager.ts` (`ensureWithinWorkingHours`) | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-G01`, `IMP-SOC-P02` | Pauses scheduled steps during off-hours according to account timezone, resuming automatically during legitimate business hours. |
  | Multi-Provider Lead Enrichment Waterfall | `shared/server/enrichment/` (`apollo`, `datagma`, `hunter`) | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-G03`, `IMP-CAR-20` | Sequentially queries allowed enrichment providers to resolve verified professional contact profiles with deduplication and cost control. |
  | Platform Limit & Restriction Detection | `linkedin.provider.ts` (`list` limit matchers) | Implemented | **ADOPT / MERGE** | `IMP-LI-18`, `FND-009` | Proactively detects platform limitation messages ("weekly invitation limit", "Too Many Requests") to halt runs and protect accounts. |
  | Per-Account Dedicated Proxy Binding | `schema.prisma` (`model Proxy`, `ProxyService`) | Implemented | **ADOPT / MERGE** | `FND-008`, `IMP-SOC-S01` | Associates authorized IP proxies per connected social account to ensure consistent geographic routing. |
  | External n8n/Make Automation Endpoints | `apps/backend/src/public-api-controllers/` | Implemented | **ADOPT / MERGE** | `IMP-SOC-P06` | Versioned API webhooks enabling low-code integration without exposing raw credentials. |
  | Stealth CDP & Anti-Detection Bypass | `cdp.detection.pass.ts`, `patchright` stealth forks | Implemented | **REJECT** | `REQ-021`, `AT-010` | Evasion of browser bot detection violates compliance invariants; platform must operate within permitted rate-limits and fail closed on challenges. |
  | Unsolicited Automated Bulk Outreach | Mass cold connection requests & auto-DMs | Implemented / Promoted | **REJECT** | `REQ-015`, `SOC-G04` | Unsolicited bulk messaging violates `P-X` and `P-LI-AUTOMATION`; outreach is restricted to pre-approved drafts with explicit human approval. |
  | Full Temporal/NestJS Stack Dependency | Heavy Temporal cluster runtime | Implemented | **DEFER** | `FND-001` | Pacing, queues, and workflows implemented natively in Go task workers and Redis/Postgres outbox without external Temporal complexity. |


<a id="source-s4"></a>

### S4 — [charlie947/social-media-skills](https://github.com/charlie947/social-media-skills)

- Evidence: [README.md](https://github.com/charlie947/social-media-skills/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `8cefb5b6d03757885faa6918bd8bfaef202a83db` (Date: 2026-03-24, branch `main`).
- Status: **Audited & Pinned; Production skill toolkit for Codex and Claude featuring 17 specialized skills for brand voice onboarding, LinkedIn drafting, copywriting frameworks, visual brief generation, and post scoring**.
- Document blob SHA: `d78265d732fec629b6c4ade6f95de27a85227201`.
- Audited Files:
  - `skills/voice-builder/SKILL.md`: Structured interview workflow deriving `about-me.md` and `voice.md` from candidate/founder interview and 3-5 writing samples.
  - `skills/newsletter-voice/SKILL.md`: Newsletter-specific extension deriving `newsletter-voice.md`.
  - `skills/hook-generator/SKILL.md`: Generates 6 concise hook formulas strictly from supplied user facts and authentic experience (no fake anecdotes).
  - `skills/post-formatter/SKILL.md`: Copywriting framework transformation converting rough topics into structured posts using PAS, AIDA, BAB, STAR, or SLAY.
  - `skills/content-matrix/SKILL.md`: Cross-product matrix pairing 3-5 pillars with 8 post formats generating 24-40 post ideas.
  - `skills/gemini-carousel/SKILL.md` & `gemini-infographic/SKILL.md`: Generates slide-by-slide vertical carousel and whiteboard infographic briefs with mandatory approval gate before image generation.
  - `skills/post-scorer/SKILL.md`: Post quality review against past post history or editorial-only fallback rules.
  - `skills/reels-scripting/SKILL.md` & `skills/youtube-thumbnail/SKILL.md`: Short-form video scripting and branded video thumbnail visual briefs.
  - `skills/analytics-dashboard/SKILL.md`: Transforms LinkedIn analytics CSV export into interactive React dashboard + 5 data-backed recommendations.
  - `validate-skills.sh`: Automated contract and schema validator for skill frontmatter, triggers, and runtime dependencies.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Authentic Brand Voice Interviewer | `skills/voice-builder/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C01`, `AT-025` | Structured interview generating canonical `about-me.md` and `voice.md` grounded in author-provided writing samples. |
  | Structured Copywriting Frameworks (PAS/AIDA/BAB/STAR/SLAY) | `skills/post-formatter/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C03` | Standardized copywriting architectures ensuring compelling, logical flow in professional posts. |
  | Fact-Grounded Hook Generator | `skills/hook-generator/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C03`, `AT-003` | Extracts 6 concise hook angles based strictly on real supplied metrics and author background without fabricated claims. |
  | Content Matrix Idea Generator | `skills/content-matrix/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C03` | Systematic cross-product matrix (3-5 pillars x 8 formats) generating 24-40 diverse post ideas. |
  | Carousel & Infographic Visual Briefs | `skills/gemini-carousel/`, `gemini-infographic/` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C04`, `FND-010` | Structured prompt-ready visual specifications with mandatory approval gates separating drafting from expensive rendering. |
  | Multi-Format Video Scripting (Reels/YouTube) | `skills/reels-scripting/`, `youtube-thumbnail/` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C05` | Script breakdown with hook timing, visual direction, and thumbnail prompts. |
  | Editorial Post Scorer | `skills/post-scorer/SKILL.md` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C10` | Objective rubric evaluating hook strength, readability, whitespace rhythm, and tone consistency. |
  | Analytics CSV to React Dashboard | `skills/analytics-dashboard/SKILL.md` | Implemented | **ADOPT / MERGE** | `IMP-SOC-P07`, `IMP-SOC-A01` | Parses exported analytics CSV into actionable performance metrics and recommendations. |
  | Standalone Gemini Image Generation Runtime | Multimodal prompt outputs | Documented Concept | **MERGE / ADAPT** | `IMP-SOC-C06`, `IMP-SOC-C08` | Text prompt briefs adapted into our unified media gateway (`FND-015`), using local or provider-backed image generation with approval cards. |
  | Proprietary LinkedIn AI OS / Private Config Claims | Private closed ecosystem claims | Documented Reference | **DEFER** | `IMP-SOC-C01` | Public toolkit contains 17 skills; private Figma pipelines or maintainer setup are excluded/deferred in favor of open-standard design tokens. |
  | Unreviewed Direct Post Dispatch | Automated publishing concept | Documented | **REJECT** | `REQ-015`, `AT-007` | Automatic publishing without human sign-off violates core invariant `REQ-015`; drafting and scoring remain decoupled from final dispatch. |
  | AI Detector Evasion Advice | Detection avoidance heuristics | Documented | **REJECT** | `REQ-016`, `AT-003` | Attempting to evade statistical AI detectors is unreliable; focus is placed on natural human readability and fact-checking. |


<a id="source-s5"></a>

### S5 — [ScrapeCreators/social-media-research-skills](https://github.com/ScrapeCreators/social-media-research-skills)

- Evidence: [README.md](https://github.com/ScrapeCreators/social-media-research-skills/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `64ba7b4dea71e130d2712ffb6c1c1024b3b7c4b2` (Date: 2026-03-24, branch `main`).
- Status: **Audited & Pinned; Production social media research skill bundle with 13 specialized workflows for outlier post detection, transcript intelligence, comment mining, competitor teardowns, ad library audits, and content repurposing**.
- Document blob SHA: `e091165b832ec255f978aed55884ad2e34ac7f0c`.
- Audited Files:
  - `skills/outlier-post-finder/SKILL.md`: Normalizes creator baseline engagement, identifies statistical outliers (2x-10x median), extracts repeatable patterns, and formats swipe file.
  - `skills/transcript-intelligence/SKILL.md`: Extracts summary, hooks, claims, quotes, and content atoms from short/long-form video transcripts across 8+ platforms.
  - `skills/comment-mining/SKILL.md`: Mines audience comment threads for objections, questions, buying intent, and Voice-of-Customer (VOC) pain points.
  - `skills/competitor-social-research/SKILL.md`: Benchmarks competitor strategies, content pillar distribution, format mix, and identifies unaddressed market gaps.
  - `skills/ad-library-teardown/SKILL.md`: Evaluates active Meta, Google, and LinkedIn ad campaigns for messaging hooks, value propositions, and CTA patterns without speculating on spend.
  - `skills/trend-discovery/SKILL.md`: Uncovers niche-specific breakout topics, hashtags, and sound formats backed by timestamped evidence tables.
  - `skills/influencer-prospecting/SKILL.md`: Identifies creator prospects from public engagement data, calculates niche fit scores, and drafts manual outreach notes.
  - `skills/audience-research/SKILL.md`: Assesses creator/brand audience composition with explicit confidence tags distinguishing observed metrics from inferred signals.
  - `skills/social-listening-brief/SKILL.md` & `product-demand-research/SKILL.md`: Multi-source social listening reports synthesizing audience themes with sentiment caveats.
  - `skills/creator-profile-teardown/SKILL.md` & `content-repurposing/SKILL.md`: Deconstructs creator playbooks and adapts viral assets into cross-platform drafts while preserving citation lineage (`SOC-R09`).
  - `scripts/validate-skills.py`: Automated validation script verifying skill metadata, required environment variables, and schema conformity.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Baseline-Aware Outlier Post Finder | `skills/outlier-post-finder/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R01`, `AT-028` | Normalizes creator baseline (median views/engagement) to identify true statistical outliers with repeatable hook patterns. |
  | Video Transcript Intelligence & Atoms | `skills/transcript-intelligence/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R02`, `IMP-SOC-C02` | Extracts core claims, quotes, and content atoms from multi-platform video transcripts for evidence-grounded drafting. |
  | Voice-of-Customer (VOC) Comment Mining | `skills/comment-mining/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R02`, `SOC-R06` | Mines public comment threads to identify user objections, feature requests, and unaddressed pain points. |
  | Competitor Strategy & Content Gap Teardown | `skills/competitor-social-research/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R03` | Comparative analysis of competitor publishing pillars, formats, and frequency to discover content gaps. |
  | Multi-Platform Public Ad Library Audit | `skills/ad-library-teardown/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R04` | Evaluates public Meta/Google/LinkedIn active ad copies, hooks, and offers without unsubstantiated spend claims. |
  | Provenance-Preserving Content Repurposing | `skills/content-repurposing/SKILL.md` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R09`, `IMP-SOC-C09` | Converts raw transcripts and research into cross-platform drafts while maintaining strict citation lineage. |
  | Evidence-Bound Trend Discovery | `skills/trend-discovery/SKILL.md` | Implemented | **ADOPT / MERGE** | `SOC-R07` | Tracks emerging niche trends and formats with verifiable timestamped source links. |
  | Influencer Prospecting & Fit Scoring | `skills/influencer-prospecting/SKILL.md` | Implemented | **ADOPT / MERGE** | `SOC-R05` | Evaluates creator fit using public follower and engagement metrics for manual networking lists. |
  | Explicit Confidence Audience Research | `skills/audience-research/SKILL.md` | Implemented | **ADOPT / MERGE** | `SOC-R05`, `AT-028` | Applies explicit confidence indicators, clearly separating observed metrics from demographic inferences. |
  | Single Vendor ScrapeCreators API Lock-in | `skills/scrapecreators-api/` | Implemented | **MERGE / ADAPT** | `FND-009`, `IMP-SOC-S01` | Research workflows abstracted via our Capability Registry (`FND-009`); supports pluggable providers (ScrapeCreators, JobSpy, native fetchers). |
  | Inferred Private Contact Discovery | Phone/private email scraping | Documented | **REJECT** | `SOC-O06`, `SEC-006` | Extracting private personal contact info or using data for unauthorized candidate screening is strictly prohibited. |
  | Unsupported Ad Spend / Conversion Claims | Estimating competitor ad spend without data | Documented | **REJECT** | `SOC-R04`, `AT-028` | Speculating on private ad budgets or conversion rates without verifiable data violates factual accuracy rules. |


<a id="source-s6"></a>

### S6 — [cporter202/social-media-scraping-apis](https://github.com/cporter202/social-media-scraping-apis)

- Evidence: [README.md](https://github.com/cporter202/social-media-scraping-apis/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `18b787f1f24ad2863b6b75467a981b74c993bbda` (Date: 2026-03-24, branch `main`).
- Status: **Audited & Pinned; Curated static index catalog of 3,268 third-party social media scraping APIs and Apify actors; serves strictly as an external provider discovery index (`SOC-S07`)**.
- Document blob SHA: `491aeebe466c5865871d59f70618bf6b38d0e608`.
- Audited Files:
  - `README.md` (1.31 MB): Catalog directory listing 3,268 RapidAPI / external scraping endpoints across Instagram, Twitter/X, TikTok, LinkedIn, YouTube, Facebook, and Reddit.
  - `settings/generate_readme_clean.js`: Node.js script filtering placeholder/test actors and compiling Markdown tables from raw JSON metadata.
  - `settings/fetch_apify_actors.js`: Node.js fetch utility querying Apify store API and formatting actor metadata.
  - `settings/APIFY_ACTORS.md`: Sub-catalog listing Apify scraping actors categorized by platform.
  - `social-media-apis-3268/`: Archive storage holding scraped endpoint JSON definitions.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Provider Discovery & Scraping Catalog Index | `README.md`, `settings/APIFY_ACTORS.md` | Catalog / Directory | **ADOPT / REFERENCE** | `SOC-S07`, `IMP-SOC-S07` | Curated reference mapping third-party social APIs and Apify actors; used for discoverability and candidate adapter research. |
  | Actor Cleanliness Filter Heuristic | `settings/generate_readme_clean.js` (`shouldFilterActor`) | Implemented | **ADOPT / MERGE** | `SOC-S07`, `FND-009` | Regex heuristics filtering broken/placeholder/test actors (`/^testactor/`, `/^my actor/`) from external marketplace data. |
  | Third-Party Vendor API Endpoints | 3,268 commercial RapidAPI/Apify links | Catalog References | **DEFER / ON-DEMAND** | `IMP-SOC-S07` | Individual catalog entries are neither vetted nor built-in; evaluated and pinned individually only when an authorized integration is needed. |
  | Skool Community / "Vibe Coding" Promotion | `README.md` promotional links | External Marketing | **DEFER / EXCLUDE** | N/A | Non-technical community promotion excluded from product architecture. |
  | Wholesale Unvetted Provider Approval | Claim that 3,268 APIs are working built-ins | Catalog Claim | **REJECT** | `REQ-024`, `AT-024` | Catalog entries cannot be represented as verified, compliant, or pre-integrated platform capabilities. |
  | Hardcoded RapidAPI Keys / Secrets | Commercial endpoint authorization | Documented | **REJECT** | `REQ-009`, `REQ-021` | Direct hardcoded API keys prohibited; credentials must be provided by user/tenant and stored in AES-256 vault. |


<a id="source-s7"></a>

### S7 — [langchain-ai/social-media-agent](https://github.com/langchain-ai/social-media-agent)

- Evidence: [README.md](https://github.com/langchain-ai/social-media-agent/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `61053aacf46f5d484eb13aad947ad52a81271f13` (Date: 2026-09-08, branch `main`).
- Status: **Audited & Pinned; Production-ready multi-graph human-in-the-loop social media curation and scheduling agent; serves as primary reference pattern for review/resume workflows (`REQ-015`, `AT-007`) and duplicate URL guard (`AT-004`)**.
- Document blob SHA: `d0f0fd089b38d3b12accff27f76b2ad98e458c04`.
- Audited Architecture & Files:
  - `langgraph.json`: 14 graph configurations (`generate_post`, `upload_post`, `curate_data`, `reflection`, `repurposer`, `generate_thread`, `supervisor`, `verify_tweet`, `verify_reddit_post`, etc.).
  - `src/agents/generate-post/generate-post-graph.ts`: Core post drafting pipeline with link extraction, content verification, post condensation (<280 chars), and human interrupt gating.
  - `src/agents/shared/nodes/generate-post/human-node.ts`: Human-in-the-loop interrupt boundary (`allow_accept`, `allow_edit`, `allow_ignore`, `allow_respond`), natural language routing (`rewrite_post`, `update_date`, `rewrite_with_split_url`), and used URLs persistence.
  - `src/agents/shared/stores/post-subject-urls.ts`: Deduplication store preventing multiple posts from being generated from the same source URLs.
  - `src/agents/reflection/index.ts`: Meta-prompting feedback loop analyzing human edits and user response to continuously improve drafting guidelines.
  - `src/agents/upload-post/index.ts`: Scheduled post uploader with image buffer conversion, character limit checks, and dual-platform dispatch (Twitter/X & LinkedIn).
  - `src/clients/auth-server.ts`: Local Passport.js OAuth server for Twitter and LinkedIn user credentials.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Human-in-the-Loop Review & Resume Interrupt | `shared/nodes/generate-post/human-node.ts` (`humanNode`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-015`, `FND-010`, `IMP-SOC-P01` | Core human gate (`accept`, `edit`, `ignore`, `respond`) where no post leaves draft state without human sign-off; re-implemented in Go ledger & Next.js Inbox. |
  | Duplicate Source URL Rejection Guard | `generate-post/generate-post-graph.ts` (`checkIfUrlsArePreviouslyUsed`) | Implemented | **ADOPT / RE-IMPLEMENT** | `AT-004`, `IMP-SOC-C02` | Persistent check of `relevantLinks` against previously used URLs preventing duplicate content creation from identical source links. |
  | Natural Language Human Feedback Router | `shared/nodes/route-response.ts`, `human-node.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C03`, `FND-015` | Parses human feedback strings into concrete state actions (`rewrite_post`, `update_date`, `rewrite_with_split_url`) rather than blind regeneration. |
  | Character Limit Auto-Condensation Loop | `generate-post/nodes/condense-post.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P03`, `AT-003` | Iteratively condenses posts exceeding platform length limits (<280 chars for X/Twitter) up to 3 attempts before halting. |
  | Meta-Prompting Reflection Feedback Loop | `src/agents/reflection/index.ts` (`reflection`) | Implemented | **ADOPT / MERGE** | `IMP-SOC-C10`, `AT-025` | Learns from accepted user edits by refining workspace drafting rules without hallucinating beyond explicit feedback. |
  | Dual-Account Retweet & Reshare Dispatch | `upload-post/index.ts` (`retweetFromMainAccount`) | Implemented | **ADOPT / MERGE** | `IMP-SOC-P01`, `IMP-SOC-P04` | Orchestrates secondary retweets/reshares from brand accounts following primary user publishing. |
  | Proprietary LangGraph Cloud & LangMem Dependency | `langgraph.json`, `langmem-v0` client | Infrastructure / Cloud | **DEFER / RE-IMPLEMENT** | `FND-001`, `FND-007` | Proprietary cloud deployment and SaaS memory service replaced natively with PostgreSQL ledger, Redis Streams, and local embedding store. |
  | Arcade Cloud Auth Delegation | `src/clients/auth-server.ts` (`useArcadeAuth`) | Third-Party Auth | **DEFER / EXCLUDE** | `FND-009`, `AT-016` | External third-party token custodian replaced by our local AES-256-GCM vault with strict owner isolation. |
  | Unreviewed Direct Post Auto-Dispatch | Direct posting without interrupt | Architecture Anti-Pattern | **REJECT** | `REQ-015`, `AT-007` | Automated publishing bypassing the human review interrupt violates platform safety invariants. |
  | Hardcoded Community Branding Signatures | `upload-post/index.ts` (`ensureSignature`) | Specific Branding | **REJECT** | `REQ-016` | Hardcoded "LangChain Community Spotlight" prefixes rejected in favor of tenant-configurable brand voice signatures. |

<a id="source-s8"></a>

### S8 — [ericciarla/trendFinder](https://github.com/ericciarla/trendFinder)

- Evidence: [README.md](https://github.com/ericciarla/trendFinder/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `b8098dad5348e3d177229534a6b3a98a55fe1a2c` (Date: 2025-02-16, branch `main`).
- Status: **Audited & Pinned; Production-ready scheduled social and web trend monitoring utility; serves as reference pattern for watch-and-alert monitors (`SOC-R07`, `FND-008`), structured web extraction, and webhook notifications (`FND-013`)**.
- Document blob SHA: `40b9679fdc4d3ff78b8ff2273b05e79145a3685b`.
- Audited Architecture & Files:
  - `src/index.ts`: Scheduled cron execution runner coordinating source discovery, web/X scraping, draft generation, and webhook notifications.
  - `src/services/getCronSources.ts`: Source catalog manager reading designated news sites and influencer handles based on active API keys.
  - `src/services/scrapeSources.ts`: Dual-channel data collector querying Twitter/X v2 search API (`query=from:{username} has:media -is:retweet -is:reply`) and Firecrawl structured web extraction with Zod schema (`StorySchema`: `headline`, `link`, `date_posted`).
  - `src/services/generateDraft.ts`: Trend summarizer prompting LLMs (o3-mini / Together AI / DeepSeek) to generate markdown bulleted trend digests with source attribution links.
  - `src/services/sendDraft.ts`: Pluggable webhook notification driver delivering alerts to Discord (`SUPPRESS_EMBEDS`) or Slack.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Scheduled Trend Watch & Alert Runner | `src/index.ts`, `src/services/getCronSources.ts` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-R07`, `FND-008`, `IMP-SOC-P02` | Automated scheduled scan of industry sites and influencer handles; re-implemented via Go scheduler leases (`FND-008`). |
  | Structured Web Extraction with Schema | `src/services/scrapeSources.ts` (`StoriesSchema`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C02`, `SOC-R07` | Grounded extraction of news headlines, source URLs, and publication dates into validated schemas (`StoriesSchema`). |
  | Multi-Driver Webhook Notifications | `src/services/sendDraft.ts` (`sendDraft`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-013`, `IMP-SOC-P06` | Pluggable alert delivery to Slack and Discord with suppressed embed bloat; integrated into unified notification service. |
  | Influencer Twitter/X Search Query Filter | `src/services/scrapeSources.ts` (query filter) | Implemented | **ADOPT / MERGE** | `IMP-SOC-P01`, `SOC-R05` | High-signal search filter (`has:media -is:retweet -is:reply`) focusing on original creator updates rather than noise. |
  | LLM-Driven Bulleted Digest Generation | `src/services/generateDraft.ts` (`generateDraft`) | Implemented | **ADOPT / MERGE** | `IMP-SOC-C03`, `AT-003` | Generates concise trend briefs preserving exact source URLs for every cited claim. |
  | Single-Vendor Hardcoded Firecrawl / Together APIs | `src/services/scrapeSources.ts`, `.env.example` | Vendor Specific | **DEFER / RE-IMPLEMENT** | `FND-009`, `FND-015` | Vendor-locked APIs replaced by unified provider capability registry (`FND-009`) and AI gateway (`FND-015`). |
  | Unbounded Twitter Free Tier Polling | README warning (1 account / 15 min) | Quota Sensitivity | **REJECT** | `REQ-009`, `REQ-021`, `AT-008` | Naive polling without quota budgeting risks immediate 429 locks; must strictly obey account reservation gatekeeper. |
  | Unreviewed Direct Auto-Posting of Trend Digests | Direct publishing without review | Anti-Pattern | **REJECT** | `REQ-015`, `AT-007` | Trend digests serve strictly as notification alerts and draft inputs; automated auto-posting without human review is prohibited. |

<a id="source-s9"></a>

### S9 — [ScrapeGraphAI/Scrapegraph-ai](https://github.com/ScrapeGraphAI/Scrapegraph-ai)

- Evidence: [README.md](https://github.com/ScrapeGraphAI/Scrapegraph-ai/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `c75c8084fae2d4f5ba01a8c218bc1168b67e3569` (Date: 2026-09-07, branch `main`).
- Status: **Audited & Pinned; Production-ready LLM-driven graph extraction pipeline; serves as primary reference pattern for semantic web and document extraction (`IMP-SOC-C02`, `SOC-R01`), Pydantic schema-guided parsing (`FND-015`), and robots.txt compliance (`REQ-024`)**.
- Document blob SHA: `386c41b106fd030bfaf5bead79e25cd9cd1a21b1`.
- Audited Architecture & Files:
  - `scrapegraphai/graphs/smart_scraper_graph.py`: Core pipeline (`SmartScraperGraph`) dynamically assembling node graphs (`FetchNode`, `ParseNode`, `ReasoningNode`, `GenerateAnswerNode`, `ConditionalNode`) depending on HTML mode and reasoning flags.
  - `scrapegraphai/nodes/fetch_node.py`: Multi-backend document fetcher supporting Playwright browser sessions, local directory reads, and HTTP clients with storage state injection.
  - `scrapegraphai/nodes/parse_node.py`: Token-bounded chunk parser segmenting HTML/markdown with structural tags preserved.
  - `scrapegraphai/nodes/generate_answer_node.py`: Prompt-driven extraction engine enforcing strict Pydantic schema conformity.
  - `scrapegraphai/nodes/robots_node.py`: Mandatory robots.txt parser enforcing polite scraping and user-agent crawling constraints.
  - `scrapegraphai/graphs/markdownify_graph.py`: HTML-to-clean-Markdown pre-processor optimizing context windows for LLM prompts.
  - `scrapegraphai/graphs/omni_scraper_graph.py`: Multi-modal scraper extracting structured data from screenshots and document images.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Schema-Guided Prompt Extraction Graph | `smart_scraper_graph.py`, `generate_answer_node.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C02`, `FND-015`, `SOC-R01` | Pydantic/Zod schema-guided extraction of web and document contents into typed structures; native integration into our Go/TS AI Gateway. |
  | HTML to Clean Markdown Compression | `markdownify_graph.py`, `markdownify_node.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C02`, `AT-020` | Strips bloated HTML tags and scripts into token-efficient markdown representations before LLM analysis. |
  | Polite Robots.txt Crawl Compliance | `nodes/robots_node.py` (`RobotsNode`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-024`, `IMP-SOC-C02` | Explicit check respecting robots.txt crawl delays and disallow directives on target domains. |
  | Token-Bounded Structural Chunking | `nodes/parse_node.py` (`ParseNode`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C02`, `FND-015` | Splits large HTML/documents into token-aligned chunks avoiding context window overflows. |
  | Conditional Reattempt & Reasoning Graph | `smart_scraper_graph.py` (`ConditionalNode`) | Implemented | **ADOPT / MERGE** | `FND-015`, `AT-003` | Re-evaluates missing or "NA" answers with targeted prompts when initial extraction fails. |
  | Document Multiformat Ingestion (XML/JSON/CSV) | `graphs/csv_scraper_graph.py`, `json_scraper_graph.py` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C02`, `IMP-CAR-18` | Ingests non-HTML structured data formats into normalized extraction graphs. |
  | Proprietary ScrapeGraphAI Cloud SDK (`scrapegraph-py`) | `integrations/scrapegraph_py_compat.py` | Commercial SaaS | **DEFER / EXCLUDE** | N/A | Commercial cloud extraction API excluded from self-hosted core architecture. |
  | Browserless Commercial Proxy Integrations | `FetchNode` (`scrape_do`, `browser_base`) | Commercial Proxy | **DEFER / RE-IMPLEMENT** | `FND-009`, `FND-011` | Commercial scraping proxy services replaced by our unified capability registry and tenant-managed proxy settings. |
  | Stealth Scraping & Anti-Bot Bypass Evasion | Browser evasion options | Anti-Pattern | **REJECT** | `REQ-021`, `AT-010` | Evasion of bot protections violates our platform safety rules; all fetching must respect rate limits and robots.txt. |
  | Unreviewed Bulk Web Scraping | Batch scrapers without rate limits | Anti-Pattern | **REJECT** | `REQ-009`, `AT-008` | Unbounded multi-graph batch runs rejected; execution must route through transactional outbox with strict quota leases. |

<a id="source-s10"></a>

### S10 — [d4vinci/Scrapling](https://github.com/d4vinci/Scrapling)

- Evidence: [README.md](https://github.com/d4vinci/Scrapling/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `2b160ee18bfee79bb0115e2d9e9c746c8d9bf4c9` (Date: 2026-09-14, branch `main`).
- Status: **Audited & Pinned; Production-ready adaptive web scraping framework; serves as primary reference pattern for crash-safe spider checkpointing/resumption (`FND-007`, `REQ-017`), latency-adaptive autothrottling (`FND-011`, `REQ-009`), and design-resilient adaptive selector recovery (`IMP-CAR-13`, `IMP-SOC-C02`)**.
- Document blob SHA: `6f8d8405b5691adb305aeae5d7a83f543ede1756`.
- Audited Architecture & Files:
  - `scrapling/parser.py`: High-performance HTML/XML parser (`Selector`) with adaptive element relocation storing element fingerprints in SQLite and computing text/attribute similarity when website layouts drift.
  - `scrapling/spiders/checkpoint.py`: Atomic crawl checkpoint manager (`CheckpointManager`, `CheckpointData`) serializing pending requests and seen URL hashes with atomic `.tmp` file rotation.
  - `scrapling/spiders/throttle.py`: Dynamic auto-throttle engine (`AutoThrottle`) adapting per-domain delays from observed latency, parsing `Retry-After` headers, and applying exponential backoff (`BLOCK_BACKOFF_FACTOR = 2.0`) upon detecting rate limits.
  - `scrapling/spiders/robotstxt.py`: Native robots.txt parser enforcing domain disallow rules and user-agent matching.
  - `scrapling/spiders/engine.py`: Concurrent multi-session spider engine with pause, resume, and graceful stop controls.
  - `scrapling/fetchers/stealth_chrome.py`: Camoufox/Playwright CDP wrapper with stealth patch options.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Atomic Spider Checkpoint & Resume | `spiders/checkpoint.py` (`CheckpointManager`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-017`, `FND-007`, `IMP-SOC-P01` | Atomic state serialization and seen URL hash deduplication enabling crash recovery without repeating completed work. |
  | Latency-Adaptive Auto-Throttle & Backoff | `spiders/throttle.py` (`AutoThrottle`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-009`, `FND-011`, `IMP-SOC-P02` | Dynamically tunes request pacing per target domain based on response latency and `Retry-After` headers with 2.0x backoff. |
  | Adaptive Layout Drift Relocation | `scrapling/parser.py` (`Selector.css(adaptive=True)`) | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-CAR-13`, `IMP-SOC-C02` | Stores element structural fingerprints to relocate target DOM elements when platforms change HTML markup/classes. |
  | Robots.txt Strict Filtering | `spiders/robotstxt.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-024`, `IMP-SOC-C02` | Respects robots.txt directives and prevents unauthorized automated crawling. |
  | Concurrent Spider Session Pooling | `spiders/engine.py` | Implemented | **ADOPT / MERGE** | `FND-008`, `IMP-SOC-P01` | Multi-session concurrency queue integrated into our Go worker lease manager. |
  | Anti-Bot Bypass Tokens & Cloudflare Turnstile Evasion | `fetchers/stealth_chrome.py`, commercial sponsors | Commercial Evasion | **REJECT** | `REQ-021`, `AT-010` | Bypassing bot protection challenges (Turnstile/DataDome/Kasada) violates our strict safety invariants; fails closed. |
  | Unbounded Concurrency Bursts | Aggressive spider scaling options | Anti-Pattern | **REJECT** | `REQ-009`, `AT-008` | Crawling must strictly operate under per-account quota budgets and reservation leases. |

<a id="source-s11"></a>

### S11 — [qeeqbox/social-analyzer](https://github.com/qeeqbox/social-analyzer)

- Evidence: [README.md](https://github.com/qeeqbox/social-analyzer/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `1ba0905e00d054aab833eb3693739c354db09e0f` (Date: 2026-01-12, branch `main`).
- Status: **Audited & Pinned; Production-ready multi-platform profile footprint analyzer; serves as reference pattern for authorized brand/user public footprint audits (`SOC-O01`, `IMP-SOC-S01`), probabilistic detection scoring, and retry wrappers (`FND-013`)**.
- Document blob SHA: `0d0f8a866e9f01894db20c778f693a8ae13e538d`.
- Audited Architecture & Files:
  - `modules/fast-scan.js`: High-speed concurrent HTTP profile scanner (`find_username_normal_wrapper`) with parallel limits (`async.parallelLimit(functions, 15)`) and 3-tier retry passes on inconclusive status codes.
  - `modules/engine.js`: Multi-signal profile detector calculating match confidence ratings (0 to 100) using title, language, response headers, and structural pattern matching.
  - `modules/extraction.js`: Public profile metadata extractor parsing open bio text, links, and public handle attributes.
  - `modules/string-analysis.js`: Username permutation and combination generator discovering handle variations.
  - `data/sites.json`: Curated dataset of 1,000+ public platform profile URL patterns, category classifications, and detection signatures.
  - `app.js` & `app.py`: Dual Node.js and Python CLI/API wrappers providing REST endpoints and JSON reporting.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Public Handle & Brand Footprint Audit | `modules/fast-scan.js`, `data/sites.json` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-O01`, `IMP-SOC-S01` | Allows users/organizations to audit availability and public presence of their brand handles across major social networks. |
  | Multi-Tier Inconclusive Retry Loop | `modules/fast-scan.js` (3-tier retry) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-009`, `IMP-SOC-P01` | Multi-tier retry logic rescuing intermittent network dropouts and false 404s before concluding absence. |
  | Probabilistic Match Confidence Rating (0-100) | `modules/engine.js` (`engine.detect`) | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-O01`, `AT-003` | Scores potential handle matches with explicit confidence tags rather than asserting definitive identity. |
  | Username Permutation Analysis | `modules/string-analysis.js` | Implemented | **ADOPT / MERGE** | `SOC-O01`, `IMP-LI-05` | Generates standard handle variations (dots, underscores, prefixes) for brand footprint scanning. |
  | Third-Party OSINT Age & Origin Inferences | `modules/name-analysis.js` | Heuristic Speculation | **REJECT** | `REQ-016`, `SEC-006` | Guessing person age, ethnicity, or personal attributes from names is prohibited; violates our strict privacy invariants. |
  | Mass Person Profiling & Cyberstalking | Invasive OSINT scanning modes | Anti-Pattern | **REJECT** | `REQ-012`, `REQ-018` | Invasive surveillance, phone/email scraping, or unauthorized tracking of third parties is strictly prohibited. |
  | Unbounded Parallel Workers (15+ threads) | `fast-scan.js` (`parallelLimit`) | Quota Sensitivity | **REJECT** | `REQ-009`, `AT-008` | High-frequency concurrent probing rejected; all outbound requests must obey account rate limits and local quotas. |

<a id="source-s12"></a>

### S12 — [Greenwolf/social_mapper](https://github.com/Greenwolf/social_mapper)

- Evidence: [README.md](https://github.com/Greenwolf/social_mapper/blob/master/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `92be8daebd850f865d26306fe9bcd20562f9364f` (Date: 2021-11-04, branch `master`).
- Status: **Audited & Pinned; Historical cross-platform footprint intelligence tool; facial recognition and mass identity linking strictly excluded (`REQ-011`, `SEC-006`, `SOC-O06`); serves as reference pattern for multi-platform footprint grid reporting (`SOC-O03`, `IMP-SOC-O03`) and resumable session reload workflows (`REQ-017`, `IMP-SOC-O01`)**.
- Document blob SHA: `6586800097bdfcbe5236ce70af24ed4096df3c24`.
- Audited Architecture & Files:
  - `social_mapper.py`: Main execution engine managing CLI argument parsing, multi-format inputs (`csv`, `imagefolder`, `company`, `socialmapper`), platform search execution loops, candidate scoring, and multi-format report exports (`.csv`, `.html`).
  - `modules/linkedinfinder.py`: Selenium-based LinkedIn profile and search parser.
  - `modules/facebookfinder.py`: Selenium Facebook profile searcher with session cookie sharing for authenticated image downloads.
  - `modules/twitterfinder.py`: Selenium Twitter profile search automation.
  - `modules/instagramfinder.py`, `pinterestfinder.py`, `vkontaktefinder.py`, `weibofinder.py`, `doubanfinder.py`: Platform-specific Selenium navigation modules.
  - Reporting logic (`social_mapper.py:1634-1685`): HTML grid generator embedding profile links, visual tooltips, and candidate comparisons.
  - Resumption logic (`social_mapper.py:1173-1238`): BeautifulSoup HTML parser extracting completed candidate states from previous partial runs to resume without re-running finished platforms.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Multi-Platform Footprint Grid & HTML/CSV Report | `social_mapper.py:1634-1685`, `1400-1450` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-O03`, `SOC-O03` | Provides clean matrix export showing confirmed authorized handles across platforms for organizational brand and security footprint auditing. |
  | Resumable Session Reload from Partial Reports | `social_mapper.py:1173-1238` (`-f socialmapper`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-017`, `IMP-SOC-O01` | Allows reloading previous partial audit runs to append newly checked platforms without re-executing completed network checks. |
  | Company Domain Handle Correlation | `social_mapper.py:1060-1090` | Implemented | **ADOPT / MERGE** | `SOC-O01`, `IMP-SOC-O01` | Associates authorized company identifiers with public profile footprints. |
  | Strictness Threshold Tiering (Fast vs. Accurate) | `social_mapper.py:967-978` (`-m`, `-t`) | Implemented | **ADOPT / MERGE** | `SOC-O02`, `AT-026` | Configurable confidence thresholds for string and handle similarity (loose/standard/strict), excluding all facial recognition. |
  | Facial Recognition & Biometric Identity Search | `social_mapper.py:138-200`, `face_recognition` | Biometric Surveillance | **REJECT** | `REQ-011`, `SEC-006`, `SOC-O06` | Facial recognition, biometric distance calculations, and cross-site image identity correlation are strictly prohibited by our product safety invariants. |
  | Mass Phishing Target List & Email Generation | `social_mapper.py:1400-1422` (`-e` flag for Gophish/Lucy) | Attack Vector | **REJECT** | `REQ-021`, `SEC-006` | Phishing campaign list building, credential harvesting preparation, and unsolicited email spraying are malicious attack patterns strictly rejected. |
  | Hardcoded Cleartext Credentials in Source | `social_mapper.py:32-67` | Security Vulnerability | **REJECT** | `REQ-009`, `REQ-021`, `AT-016` | Hardcoding social account credentials in source scripts violates fundamental security standards; our platform requires encrypted secret vaults. |
  | Headless Scraping with Disposable/Deceptive Accounts | `README.md:8-9`, selenium drivers | Anti-Pattern | **REJECT** | `REQ-021`, `AT-010` | Utilizing burner accounts to evade platform abuse detection violates platform terms of service and our ethical crawling rules. |

<a id="source-s13"></a>

### S13 — [Alfredredbird/tookie-osint](https://github.com/Alfredredbird/tookie-osint)

- Evidence: [README.md](https://github.com/Alfredredbird/tookie-osint/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `d418736b283785d50e8529afb9aec160faf752bc` (Date: 2026-09-15, branch `main`).
- Status: **Audited & Pinned; Lightweight username lookup CLI & scraper; serves as reference pattern for safe filename sanitization (`SEC-006`, `FND-006`), scan resumption points (`REQ-017`, `IMP-SOC-O01`), and schema-driven field extraction (`IMP-SOC-C02`, `SOC-O03`)**.
- Document blob SHA: `913b02cf100ba48c81b0dc962e413c38a997a47f`.
- Audited Architecture & Files:
  - `brib.py`: Main CLI execution script orchestrating argument parsing (`argparse`), thread pool execution (`ThreadPoolExecutor`), restore loading, and webhook notifications.
  - `modules/modules.py`: Core utility functions including `_safe_filename` (regex-based path traversal prevention filtering characters outside `[A-Za-z0-9._-]`), `load_sites`, `scan_site` (HTTP status evaluation `200 <= code <= 305`), and multi-format exporters (`write_txt`, `write_csv`, `write_json`).
  - `modules/files.py`: State checkpointing (`make_restore`, `load_restore` parsing `.tookie` restore files to resume interrupted runs from last attempted target) and Discord/Slack webhook dispatch (`send_webhook`).
  - `modules/webscraper.py`: Headless Selenium Chrome driver (`check_site`, `extract_fields`) extracting configured DOM fields via CSS/XPath selectors.
  - `sites/sites.json` & `sites/fields.json`: Declarative platform URL directories and metadata extraction selectors.
  - Roadmapped / Unimplemented features confirmed in code inspection: Tor searching (planned/unimplemented), WebUI (`brib.py:133-134` throws not implemented error), Phone OSINT (unimplemented), Custom plugins (unimplemented), and Email OSINT (unimplemented).
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | Safe Filename Sanitization & Path Traversal Guard | `modules/modules.py:_safe_filename` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-006`, `SEC-006` | Enforces strict path traversal defenses on user-supplied handles/identifiers before using them in filesystem export operations. |
  | Linear Scan Resumption Point / State Marker | `modules/files.py:make_restore`, `load_restore` | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-017`, `IMP-SOC-O01` | Records checkpoint URL/state in local marker files to resume multi-site audit scans without restarting from index 0. |
  | Declarative Field Metadata Extraction Schema | `modules/webscraper.py:extract_fields`, `sites/fields.json` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-C02`, `SOC-O03` | Uses JSON-driven CSS and XPath selector maps to harvest structured public profile attributes from dynamic web pages. |
  | Multi-Format Output Export (JSON/CSV) | `modules/modules.py:write_csv`, `write_json` | Implemented | **ADOPT / MERGE** | `IMP-SOC-O03` | Serializes structured audit results into standardized CSV and JSON formats for reporting and downstream analytics. |
  | Asynchronous Scan Webhook Notifications | `modules/files.py:send_webhook` | Implemented (Beta) | **ADOPT / MERGE** | `FND-013`, `IMP-SOC-P06` | Dispatches webhook event payloads to Slack/Discord upon scan completion or candidate discovery. |
  | Unimplemented Roadmap Features (Tor/Phone/WebUI/Email) | `README.md:165-174`, `brib.py:133-134` | Unimplemented / Planned | **DEFER** | `REQ-024`, `AT-024` | Advertised roadmap capabilities are confirmed unimplemented in codebase; deferred from product dependencies. |
  | Unthrottled Username Probing / Thread Bursts | `brib.py` (unbounded concurrency) | Anti-Pattern | **REJECT** | `REQ-009`, `AT-008` | Unthrottled outbound request bursts rejected; all platform footprint checks must execute under account quota leases and rate limits. |
  | Phone Number & Sensitive Demographic Enumeration | `README.md` (Phone OSINT roadmap) | Invasive Surveillance | **REJECT** | `REQ-011`, `SEC-006`, `SOC-O06` | Personal phone number enumeration and sensitive demographic profiling are strictly prohibited by our platform privacy invariants. |
  | Non-Consensual Target Stalking / Harassment | Offensive OSINT Use Cases | Anti-Pattern | **REJECT** | `REQ-011`, `REQ-021` | Tooling is restricted to authorized brand footprint discovery and security audits; third-party surveillance is strictly prohibited. |

<a id="source-s14"></a>

### S14 — [arxhr007/Aliens_eye](https://github.com/arxhr007/Aliens_eye)

- Evidence: [README.md](https://github.com/arxhr007/Aliens_eye/blob/main/README.md); inspected 2026-09-18.
- Pinned Commit SHA: `8f8d05724e1b970b529b3015a994d26f18dbe295` (Date: 2026-09-06, branch `main`).
- Status: **Audited & Pinned; Advanced multi-platform footprint scanner with ML-blended detection; serves as reference pattern for SSRF guards on avatar/URL fetching (`SEC-006`, `FND-006`), frozen response corpus & precision/recall evaluation (`SOC-O02`, `IMP-SOC-O02`), atomic scan checkpoints (`REQ-017`, `IMP-SOC-O01`), and Model Context Protocol agent tool integration (`FND-016`)**.
- Document blob SHA: `cfd244bb83f3c8e312bad0f558eb52332046e5bd`.
- Audited Architecture & Files:
  - `src/aliens_eye/cli.py`: Comprehensive CLI driver supporting multi-username discovery, watch schedules, variation levels, and multi-format exporters.
  - `src/aliens_eye/core/detector.py`: ML + heuristic blended detection engine (`Detector.predict`) weighting scikit-learn models and 30 structural DOM/status features via sigmoid probability scaling.
  - `src/aliens_eye/core/correlate.py`: Cross-site profile clusterer featuring a hardened SSRF protection guard (`is_fetchable_avatar`) that resolves host IPs and blocks loopback, private networks, and cloud metadata endpoints (`169.254.169.254`).
  - `src/aliens_eye/core/checkpoint.py`: Checkpoint manager serializing scan states to JSONL to resume aborted executions.
  - `src/aliens_eye/core/domains.py`: Brand domain presence evaluator checking TLD registration and DNS availability.
  - `src/aliens_eye/selfcheck.py` & `src/aliens_eye/eval/`: Offline evaluation framework testing detection models against a frozen response corpus (2,284 captures) with precision, recall, F1, and false-positive rate metrics.
  - `src/aliens_eye/mcp_server.py`: Native Model Context Protocol (FastMCP) server exposing tools (`scan_username`, `correlate`, `read_report`) over stdio for LLM coding agents.
  - `src/aliens_eye/core/exporter.py` & `pdf_report.py`: Multi-format reporting engine generating JSON, CSV, HTML, Markdown, PDF, GEXF, and Mermaid graphs.
- Feature Inventory & Disposition Table (`REQ-012`, `REQ-024`, `AT-024`):
  | Component / Feature | File / Function | Classification | Disposition | Target Task | Rationale |
  |---|---|---|---|---|---|
  | SSRF Guard on Target/Avatar URL Fetching | `core/correlate.py:is_fetchable_avatar` | Implemented | **ADOPT / RE-IMPLEMENT** | `SEC-006`, `FND-006`, `IMP-SOC-O01` | Resolves target URLs and rejects loopback, private RFC1918 subnets, and cloud metadata IPs (`169.254.169.254`) before performing any HTTP GET. |
  | Frozen Response Corpus & Precision/Recall Evaluation | `selfcheck.py`, `eval/` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-O02`, `IMP-SOC-O02`, `AT-026` | Evaluates detection accuracy against offline frozen response corpora to provide empirical precision, recall, and false-positive benchmarks without flaky live tests. |
  | ML + Heuristic Blended Confidence Scoring | `core/detector.py:predict` | Implemented | **ADOPT / RE-IMPLEMENT** | `SOC-O02`, `AT-003` | Evaluates multi-signal DOM and header features with calibrated probabilities rather than naive HTTP status codes. |
  | Atomic Scan Checkpoint & State Resume | `core/checkpoint.py` (`--resume`) | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-017`, `IMP-SOC-O01` | Records progress incrementally in JSONL logs allowing interrupted scans to resume without repeating finished requests. |
  | MCP Server Interface for AI Agent Tools | `mcp_server.py:serve` | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-016` | Exposes scanning and report inspection endpoints directly via FastMCP stdio interface to autonomous LLM agents. |
  | Multi-Format Graph & Visual Reports | `core/exporter.py` (Mermaid/GEXF/HTML) | Implemented | **ADOPT / MERGE** | `IMP-SOC-O03`, `SOC-O03` | Produces visual cluster graphs and shareable reports for brand footprint audits. |
  | Brand Domain TLD Live Verification | `core/domains.py` (`--domains`) | Implemented | **ADOPT / MERGE** | `SOC-O01`, `IMP-SOC-O01` | Validates domain registration alongside social handles to ensure complete brand footprint coverage. |
  | Recursive Traversal & Arbitrary Person Stalking | `cli.py` (`--recurse-depth N`) | Anti-Pattern | **REJECT** | `REQ-011`, `REQ-021`, `SOC-O06` | Recursive automated tracking and crawling linked third-party identities is strictly prohibited by our platform safety invariants. |
  | Unthrottled 840+ Site Concurrent Bursts | `cli.py` (unbounded concurrency) | Anti-Pattern | **REJECT** | `REQ-009`, `AT-008` | High-frequency concurrent probing across hundreds of external services rejected; must obey per-account rate limits and reservation leases. |
  | Inferred Sensitive Demographics & Speculation | Invasive OSINT Attributes | Anti-Pattern | **REJECT** | `REQ-011`, `SEC-006`, `SOC-O06` | Extracting or inferring private demographic attributes is strictly excluded from our product requirements. |

<a id="source-s15"></a>

### S15 — [ZJU-REAL/Easel](https://github.com/ZJU-REAL/Easel)

- Evidence: [README_EN.md](https://github.com/ZJU-REAL/Easel/blob/main/README_EN.md), [`docs/skill-function-mapping.md`](https://github.com/ZJU-REAL/Easel/blob/main/docs/skill-function-mapping.md), [`profiles/_template/`](https://github.com/ZJU-REAL/Easel/tree/main/profiles/_template), [`skills/shared/scripts/content_guard.py`](https://github.com/ZJU-REAL/Easel/blob/main/skills/shared/scripts/content_guard.py), [`persona_gate.py`](https://github.com/ZJU-REAL/Easel/blob/main/skills/shared/scripts/persona_gate.py), [`manifest.py`](https://github.com/ZJU-REAL/Easel/blob/main/skills/shared/scripts/manifest.py); inspected 2026-09-18.
- Pinned commit: `406438c30835a67ad9a1daaab99bcf8f4024ec61` (Release `v0.2.0`, branch `main`).
- Architecture: 5-layer agentic content generation & publishing workflow engine containing **114 modular skills** across 5 distinct operational layers (`Foundation: 6`, `Discover: 9`, `Plan: 16`, `Produce: 51`, `Publish: 20`, `Attribute: 11`), powered by a 6-dimensional account profile system (`identity`, `style`, `audience`, `platforms`, `preferences`, `memory`), deterministic pre-dispatch safety scanner (`content_guard.py`), persona consistency gate (`persona_gate.py`), and atomic inter-layer manifest contract (`manifest.py`).
- Status: **Audited upstream codebase; core architecture adopted; proprietary cloud generation APIs deferred; unreviewed direct social dispatch rejected.**

| Component / Feature | Upstream Evidence / File | Upstream Status | Platform Disposition | Target Task / Requirement | Architectural Rationale & Implementation Notes |
|---|---|---|---|---|---|
| **6-Dimensional Account Profile System** | `profiles/_template/` (`identity.md`, `style.md`, `audience.md`, `platforms.md`, `preferences.md`, `memory.md`) | Implemented | **ADOPT / RE-IMPLEMENT** | `FND-005`, `IMP-SOC-P05`, `IMP-SOC-C01`, `REQ-002` | Adopts multi-dimensional structured profile partitioning into our canonical profile store: identity (positioning for discovery/plan), style (tone/visual cadence for production), audience (personas/pain points for plan/produce), platforms (native constraints/rules for publish), preferences (hard boundaries/red-lines across all layers), and memory (continuous feedback & postmortem insights). |
| **Outbound Content Security Guard** | `skills/shared/scripts/content_guard.py:guard_or_die` | Implemented | **ADOPT / RE-IMPLEMENT** | `SEC-006`, `FND-006`, `REQ-016`, `IMP-SOC-P08` | Implements deterministic fail-closed pre-dispatch scanner rejecting secrets (`sk-*`, bearer tokens), internal domains, proxy IPs (`10.*`, `192.168.*`, `172.16-31.*`), private file paths, and environment variable references with exit code 7 before any public social post dispatch. |
| **Persona Consistency Audit Gate** | `skills/shared/scripts/persona_gate.py:classify` | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-016`, `IMP-SOC-P01`, `IMP-SOC-C10` | Implements two-stage persona alignment gate (LLM persona deviation scoring combined with deterministic threshold classifier `>= 80 pass`, `< 80 warn`) recorded directly into stage manifest before scheduling or publishing. |
| **Standardized Project Manifest Pipeline** | `skills/shared/scripts/manifest.py`, `output_paths.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `REQ-017`, `FND-006`, `FND-010`, `IMP-SOC-P02` | Adopts structured inter-layer contract (`.easel.json` schema: topic, profile, timestamps, title, summary, platform, kind, status, tags, deliverables, step audit trail) ensuring downstream generation steps read verified upstream artifacts rather than ungrounded agent speculation. |
| **Unified Content Calendar & Scheduling** | `skills/shared/scripts/calendar_ops.py`, `skill-content-calendar` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P02`, `FND-010`, `REQ-011` | Implements centralized calendar matrix scheduling topics across dates, platforms, and formats with timezone awareness and durable state tracking. |
| **Deterministic Media Processing Operations** | `skills/shared/scripts/image_ops.py`, `audio_ops.py`, `video_ops.py` | Implemented | **ADOPT / RE-IMPLEMENT** | `IMP-SOC-P03`, `IMP-SOC-C06`, `IMP-SOC-C07` | Adopts local deterministic FFmpeg/Pillow utilities for resizing, padding, format conversion, audio normalization, denoise, and watermarking without cloud vendor roundtrips. |
| **Hook & Writing Framework Matrix** | `skills/openclaw/skill-hook-generator/`, `skill-content-matrix/` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C03`, `AT-003` | Adopts proven attention hook formulas and content pillar × format matrices with character length verification and structured alternatives. |
| **Multi-Platform Repurposing Engine** | `skills/openclaw/skill-content-repurposing/` | Implemented | **ADOPT / MERGE** | `IMP-SOC-C09`, `IMP-SOC-P03` | Reconstructs single core topics into platform-native variants (vertical cards, short scripts, articles, carousel outlines) while strictly preserving source citation lineage. |
| **Content Performance Review & Postmortem** | `skills/openclaw/skill-content-postmortem/`, `skill-post-scorer/` | Implemented | **ADOPT / MERGE** | `IMP-SOC-P07`, `IMP-SOC-C10`, `AT-025` | Extracts structural patterns from high-performing posts and records verified learnings back into the profile memory layer. |
| **Regional Platform Publishing Adapters** | `skills/shared/scripts/xhs_publish.py`, `douyin_publish.py`, `web_publisher.py` | Implemented | **ADOPT / MERGE** | `IMP-SOC-P01`, `IMP-SOC-C11` | Adopts regional platform parameter schemas and dry-run preflight validation for Xiaohongshu, Douyin, Zhihu, and WeChat official accounts. |
| **Proprietary Chinese Cloud Generation APIs** | DashScope, Wan, Seedance, Kuaishou Kling | External Service | **DEFER** | `FND-015`, `IMP-SOC-C08` | Vendor-locked proprietary generative video and music models deferred to pluggable AI Gateway adapters (`FND-015`) rather than hardcoded core dependencies. |
| **Unreviewed Direct Auto-Publishing** | Automated browser `--exec` dispatch without user gate | Anti-Pattern | **REJECT** | `REQ-015`, `AT-007`, `REQ-009` | Publishing directly to external platforms without explicit, human-reviewed approval is strictly prohibited by our system safety invariants. |

## 11. Official policy / service evidence

- **P-LI-AUTOMATION** — [LinkedIn prohibited software](https://www.linkedin.com/help/linkedin/answer/a1341387), reviewed 2026-09-17: LinkedIn prohibits unauthorized scraping and automated website activity; restriction/suspension remains possible.
- **P-LI-OIDC** — [LinkedIn OpenID Connect](https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2), reviewed 2026-09-17: Sign-in obtains lite profile/email permissions, not full career history or general messaging/apply permission.
- **P-LI-UA** — [LinkedIn User Agreement](https://www.linkedin.com/legal/user-agreement), reviewed 2026-09-17: Provider permissions and restrictions are release dependencies; product quotas cannot override them.
- **P-X** — [X automation rules](https://help.x.com/en/rules-and-policies/x-automation), reviewed 2026-09-17: API and consent requirements; no unsolicited automated bulk DMs; AI reply bots require explicit prior written approval in the reviewed policy.
- **P-SHEETS** — [Google Sheets usage limits](https://developers.google.com/workspace/sheets/api/limits), reviewed 2026-09-17: Separate per-user/project read/write quotas; batch writes and bounded exponential backoff; check actual project quotas.
- **P-MIXPOST** — [Mixpost Lite documentation](https://docs.mixpost.app/lite/), reviewed 2026-09-17: Distinguish public Lite package from separately described Pro/Enterprise features.

Meta/Instagram documentation could not be fetched successfully in this session (rate-limited). No Instagram action quota is asserted here. Each provider adapter must obtain current official requirements before release. Do not substitute older repository quota claims.

## 12. Data and AI boundaries

Treat uploaded resumes, public webpages, social comments, repository content and LLM output as untrusted input. They cannot grant permissions, request secrets, change system instructions or cause external side effects. Record source references for factual summaries. Minimize LLM inputs; users select/consent to providers. Do not use private resumes, DMs or documents as training data by default.

Keep public brand voice separate from private legal/eligibility answers. Private records remain private even in a team workspace. Support audit logs never contain full tokens or raw sensitive fields. Deletion revokes tokens, cancels queued work and removes derived search/exports according to the documented retention policy. Backups must be subject to retention and delete-on-restore rules.

## 13. Scope of this handoff

Complete: product/design documentation and documentation-level repository review. Not complete: all-source code audit, measured benchmarks, application code, tests, deployments, OAuth app approvals, billing configuration or production packages. TASKS.md and PROGRESS.md deliberately reflect that reality.

<!-- Reference links for official evidence labels used above. -->
[P-LI-AUTOMATION]: https://www.linkedin.com/help/linkedin/answer/a1341387
[P-LI-OIDC]: https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2
[P-LI-UA]: https://www.linkedin.com/legal/user-agreement
[P-X]: https://help.x.com/en/rules-and-policies/x-automation
[P-SHEETS]: https://developers.google.com/workspace/sheets/api/limits
[P-MIXPOST]: https://docs.mixpost.app/lite/
