# PLAN — Architecture and step-by-step implementation

Prepared: **2026-09-17**. This is a build specification. It does not describe software already implemented. Product requirements and repository evidence in [CONTEXT.md](CONTEXT.md) take precedence over illustrative names here. Track actual execution in [TASKS.md](TASKS.md) and [PROGRESS.md](PROGRESS.md).

Before starting or continuing any work, read README.md, CONTEXT.md, PLAN.md, AI-AGENT-ARCHITECTURE.md, TASKS.md, and PROGRESS.md in full and treat them as one interconnected source of truth. Cross-reference requirements, architecture decisions, AI-agent rules, task dependencies, implementation status, and prior decisions across all six files before making changes. Do not interpret any file in isolation, silently override an existing decision, duplicate functionality, change the approved stack, or mark work complete without implementation and verification evidence. After each meaningful implementation step, update TASKS.md and PROGRESS.md so the next AI/coding session can resume accurately without losing context.

## 1. Architecture decisions

| ID | Decision | Reason / constraint |
|---|---|---|
| ADR-001 | Go modular monolith for core API and domain workers | One authorization model, transaction boundary and shared data model; extract services only when operational evidence justifies it. |
| ADR-002 | Next.js, TypeScript and Tailwind for one responsive application | Three product sections, role-aware route groups, common UI and approval center. No second disconnected admin product. |
| ADR-003 | PostgreSQL owns durable state | Applications, approvals, scheduling, quota ledger, billing and outbox must survive Redis loss. |
| ADR-004 | Redis Streams for cross-language work delivery | Specify one versioned message envelope for Go and TypeScript. Do not assume unrelated Go/Node queue libraries share a wire format. Delivery is at least once, not exactly once. |
| ADR-005 | Meilisearch is a rebuildable projection | Go authorizes and scopes every query, including counts, facets and exports. |
| ADR-006 | Private object storage is required supporting infrastructure | PDFs, DOCX, media and artifacts do not belong in Redis or public Git. Use a local-volume adapter for self-host and S3-compatible adapter for cloud. |
| ADR-007 | Isolated TypeScript workers for document/media tools and permitted browser tasks | Preserve the chosen Go/TS stack. Browser access is not permission to automate a platform. No mandatory donor Python/PHP backend. |
| ADR-008 | One publisher, one profile service, one policy engine | LinkedIn and Social views may refer to the same content/account; no duplicate sends or independent budgets. |
| ADR-009 | External side effects are capability-gated and approval-bound | Permission, account consent, recipient consent where needed, entitlements and operational budgets are separate checks. |
| ADR-010 | OSS and cloud share a core | Managed services are adapters; no hardcoded cloud-only dependency for self-hosting. |
| ADR-011 | All donor implementations remain references until audited | Pin commit, inspect actual files/functions, identify dependencies, preserve notices and test the reimplementation. README breadth is not a reliability benchmark. |
| ADR-012 | Manual/native fallback is a supported execution path | User can finish resume, review jobs, apply externally and record outcomes without an approved LinkedIn automation integration. |
| ADR-013 | No fabricated identity or career facts | Unknown answers require clarification; facts, inferred suggestions and user confirmations are separate records. |
| ADR-014 | Releases are sliced; feature inventory is not silently reduced | Adopted, merged, deferred and deliberately rejected source ideas all remain traceable. |
| ADR-015 | Dual-Product & Dedicated Repository Split | Split into Product 1 (Career & LinkedIn in `career/`) and Product 2 (SocialSuite in `SocialSuite/`) under unified domain `yourdomain.com` with separate subscription tiers. |

Changes require an ADR amendment in this file, corresponding CONTEXT/TASKS changes and a migration plan when persisted behavior changes. An AI agent must not silently replace Go with NestJS/Python or introduce Temporal/LangGraph because an upstream repository uses it.

## 2. Intended repository layout

```text
career-platform-monorepo
├── doc/                            # Central Documentation & System Specifications (.md)
│   ├── README.md
│   ├── CONTEXT.md
│   ├── PLAN.md
│   ├── TASKS.md
│   ├── PROGRESS.md
│   └── AI-AGENT-ARCHITECTURE.md
├── career/                         # Product 1: Career & LinkedIn Platform (yourdomain.com/career)
│   ├── apps/
│   │   └── web/                    # Next.js 15 Career & LinkedIn App Router
│   ├── services/
│   │   └── core/                   # Go core service (career, linkedin, identity, workspace, audit)
│   ├── packages/
│   │   └── contracts/              # TypeScript contracts
│   ├── migrations/                 # PostgreSQL schema migrations
│   ├── deploy/                     # Docker Compose & deployment manifests
│   ├── Makefile, package.json      # Monorepo orchestrator
│   └── pnpm-workspace.yaml
└── SocialSuite/                    # Product 2: Social Media & Creator Studio (yourdomain.com/social)
    ├── apps/
    │   └── web/                    # Next.js 15 Social Suite App Router
    ├── services/
    │   └── core/                   # Go publishing scheduler & analytics service
    ├── packages/
    │   └── contracts/              # Shared social contracts
    ├── package.json
    └── README.md
```
  openapi.yaml                      # Versioned HTTP contract; generated SDK input
packages/
  ui/
  sdk/                              # Generated TypeScript client
  contracts/                        # Validated worker envelopes and JSON Schemas
migrations/
tests/
  fixtures/                         # Synthetic resumes, HTML forms, provider responses
  integration/
  contract/
  security/
  e2e/
  source-inventory/                 # Pinned upstream manifests and reviewed feature map
  benchmarks/                       # Reproducible fixtures, not inflated claims
config/
  providers/                        # Reviewed capability manifests, no credentials
  packages/                         # Draft/default entitlement definitions
scripts/
deploy/
  compose/
  self-hosted/
  cloud/
```

This layout is planned; folders and executables have not been created by the documentation pack. Go commands live above their `internal/` tree so Go import boundaries remain valid.

## 3. End-to-end boundaries

A browser request goes through authentication and Go authorization before domain access. Next.js may handle rendering and a same-origin transport facade, but must not bypass Go policy for direct PostgreSQL writes, provider calls or agent tools. Never put a provider token or Meilisearch admin key in a `NEXT_PUBLIC_*` variable.

The core flow is:

```text
Next.js UI → Go API → authorization + domain service → PostgreSQL transaction
                                                       ↓
                                                   outbox record
                                                       ↓
                                        relay → Redis Stream → worker
                                                       ↓
                                            provider / document tool
                                                       ↓
                                     internal result API → PostgreSQL
                                                       ↓
                                       search projection + UI events
```

Go workers can use core repositories directly. TypeScript workers obtain narrowly scoped task authorization and post results through authenticated internal APIs. They do not receive a full database credential or authority to modify arbitrary rows. SSE provides user-visible progress; reconnect reads authoritative run history rather than relying on missed event messages.

## 4. Identity, roles and tenant boundaries

### Authentication

Implement email verification, secure password handling through a reviewed library, password reset, session rotation/revocation, optional Google/OIDC login and optional MFA. Super Admin MFA is required in cloud production. Do not hardcode seed passwords. Bootstrap the first Super Admin through a one-time operator action and disable that bootstrap path afterward.

### Authorization

`super_admin` is an installation/platform permission; `admin` and `user` are workspace memberships. A person may belong to multiple workspaces. The currently selected workspace is explicit, but all requests revalidate membership server-side. UI visibility alone is never access control.

Resource classes: `owner_private`, `workspace_shared`, `delegated`. Personal career records default to owner-private even inside a workspace. Admin manages team/shared content and billing, not every member's private resume. Delegation is explicit, scoped, revocable and logged. Support access requires a time-limited, consented reason and a read/write capability decision; it is not invisible impersonation.

Sensitive operations include package publication, permission changes, credential revocation, data export, account deletion and destructive moderation. Require fresh authorization and audit context. Tenant IDs, actor IDs and source ownership are derived from verified session context, not trusted client input.

## 5. Storage model

The following are entity groups, not final SQL migrations. Define exact columns, constraints and indexes during FND-002. Use UTC instants, IANA timezones for schedules, explicit currencies, and date precision for partial employment dates.

| Domain | Entities | Essential invariants |
|---|---|---|
| Identity | users, auth_identities, sessions, workspaces, memberships, role_permissions, resource_grants | Unique provider identity; no cross-workspace privilege inheritance; scoped delegation. |
| Profile | profiles, profile_versions, profile_facts, profile_sources, profile_reviews, job_preferences | Confirmed versions immutable; extracted and approved facts distinguishable; source spans and consent stored. |
| Documents | uploads, document_artifacts, resume_versions, cover_letter_versions, document_checks | Owner-private blobs; immutable source/profile/template versions; rendering status and parse-back report. |
| Career | jobs, job_sources, saved_jobs, match_reports, applications, application_events, application_artifacts, interview_records | Canonical job identity separate from source URL; unique applicant/canonical-job application unless explicitly reviewed reapplication. |
| Relationships | professional_contacts, contact_sources, contact_grants, conversations, message_drafts, relationship_events | Provenance, permissions and recipient consent; no inferred private identity as verified fact. |
| Accounts | provider_accounts, provider_credentials, provider_capabilities, account_health_events | Stable external account identity; encrypted secrets; independently approved capability manifests. |
| Content | brand_profiles, voice_versions, content_items, content_versions, media_assets, content_variants, calendar_entries, publish_attempts | One content record across module views; per-provider receipts; immutable approved revisions. |
| Research | research_runs, source_documents, evidence_items, trend_watches, topic_briefs, report_versions | Source URL/time/coverage; permitted access; statements distinguish evidence from inference. |
| Workflows | workflow_definitions, workflow_versions, action_runs, action_steps, action_attempts, approvals, outbox_events, scheduled_jobs | Idempotency, versioned inputs, leases/fencing, explicit unknown-outcome state. |
| Policy | consent_records, suppression_entries, policy_versions, quota_reservations, action_ledger | Nested daily/weekly budgets, all-or-nothing reservation, revocation before execution. |
| Commerce | packages, package_versions, subscriptions, subscription_events, entitlement_grants, usage_events, billing_events | Published versions immutable; webhook/event uniqueness; no package override of provider policy. |
| Export | export_jobs, export_artifacts, sheets_connections, sheets_sync_targets, sheets_sync_items, sheets_sync_attempts | Stable record keys; one-way ownership; no blind append after an ambiguous response. |
| Operations | audit_events, notifications, incident_records, retention_policies, deletion_jobs, deletion_tombstones | Scoped logs; no tokens/raw sensitive documents; purge/search/object-store propagation. |

Use foreign keys that retain workspace/owner consistency. Add database constraints for uniqueness and valid state transitions where possible, plus application authorization. PostgreSQL row-level policies are a defense in depth, not a substitute for correct service checks. Migration tests must cover pooled-connection tenant context leakage and reset behavior.

### Search projections

Create separate jobs, companies, professional contacts, content and research projections as needed. Every indexed object carries its workspace/owner visibility identifiers and only required fields. Personal resumes and private message bodies are not globally indexed. The Go search facade enforces access filters for results, counts and facets. Do not allow clients to replace the mandatory filter.

Outbox updates record projection version and deletion tombstones. Indexing is eventually consistent; authoritative record details always come from Go/PostgreSQL. Track index lag, failures and rebuild progress. Deleting an account/document removes its search copies too.

## 6. Durable workflows and queue contract

### Canonical state

PostgreSQL stores the run, input snapshot, current step, scheduled time, attempts, approval and outcome. Redis transport can be reconstructed from due, unfinished database work. Cache or transport loss must not erase an application or turn uncertain delivery into a fresh submission.

A versioned envelope includes `schema_version`, `message_id`, `run_id`, `step_id`, `attempt_id`, `workspace_id`, `actor_id`, `kind`, `payload_ref`, `input_version`, `correlation_id`, `not_before`, and `created_at`. Messages contain references instead of raw resumes, cookies or provider keys.

### Outbox and scheduling

Create domain changes and an outbox row in one transaction. A relay dispatches and marks deliveries with idempotent bookkeeping. A scheduler leases due work using transactional row locking, for example `FOR UPDATE SKIP LOCKED`. Never hold a database transaction open during a provider/browser call.

Workers persist an attempt and obtain a fencing token before work. A completion with an expired token cannot overwrite a newer attempt. Redis consumer-group acknowledgments occur after durable completion or durable recovery-state recording. A crashed consumer can be reclaimed, but a non-idempotent external write is reconciled before another attempt.

### Retry categories

| Category | Required behavior |
|---|---|
| Transient failure before request dispatch | Bounded retry with exponential backoff, jitter and a maximum elapsed budget. |
| Provider 429 | Respect provider reset/Retry-After, reduce queue pressure and show reschedule time. Never rotate identities to bypass it. |
| Authentication expired | Pause affected account, request reauthentication and recheck all scopes. |
| Permission denied / unavailable capability | Do not fall back to an unauthorized scraper; offer supported manual/import path. |
| CAPTCHA or security challenge | Pause; notify user. Do not add automated challenge solving or stealth bypass. |
| Provider timeout after potential submit | `needs_confirmation`; look for a receipt/remote state. Never assume failure or retry blindly. |
| Invalid profile answer | `needs_input`; do not invent a value. |
| Invalid or changed approval | Return to review. |
| Permanent data/schema failure | Explicit failure with actionable validation details; no endless retry. |
| Redis/policy service unavailable | Fail closed for side effects; keep read-only/local editing available where possible. |

Provide cancel, pause and resume. Cancel stops future unsent steps; it cannot undo a post, message or application already accepted externally. Show that distinction in the UI.

## 7. Account policy and action reservations

Use the policy and tooltip values in CONTEXT, not an invented universal daily limit.

Evaluation order:

1. Authenticate actor; authorize workspace/resource/account use.
2. Verify approved provider capability and current scopes for this action.
3. Verify account consent and recipient consent/suppression where applicable.
4. Verify entitlement and operational policy; unsupported actions remain disabled.
5. Validate immutable approval against actor, account, recipient, content/document revision, job URL/ID, policy version and expiry.
6. Resolve account health; challenges, revocation and emergency pause block work.
7. Atomically reserve all relevant windows and nested budgets; persist the attempt and reservation.
8. Immediately before external dispatch, recheck revocation/approval freshness and persist the dispatch checkpoint.
9. Record confirmed success or explicit uncertainty. Settle reservations according to what actually happened.

The daily and weekly counters use the **same stable provider account key across Career, LinkedIn and Social**. Attaching the same account to another workspace, upgrading a package or reconnecting must not reset its history. Recruiter outreach consumes both the recruiter sub-budget and the general DM budget. Two workers cannot each take the last remaining slot.

A reservation counts toward availability while in flight. A stale reservation is not automatically refunded when the external result is uncertain; reconcile first. A known pre-dispatch failure can release it. Store audit reasons for policy changes, reservation release and administrative reduction. Users may lower their limits. Admin cannot increase them above global/provider policy; platform permission remains a separate boolean gate, not just a number.

## 8. Profile import, fact review and ATS-friendly output

### Upload pipeline

Begin with a short-lived authorized upload token. Validate extension, MIME signature, maximum size and filename; store in quarantine with a random key. Reject unsupported/encrypted inputs with actionable instructions. Limit DOCX decompression and external references; avoid executing macros, embedded scripts or active document content. Scan uploads and isolate parsers with time/memory limits and restricted networking.

Extract text first. A scanned PDF without usable text enters `needs_ocr`; disclose processing, seek user consent where applicable and show uncertain OCR fields. No repeated OCR loop. Treat resume content as untrusted data, including prompt-like instructions.

The extractor outputs a schema with field value, source document/version, page or paragraph evidence, confidence and missing fields. Users review a side-by-side original/extracted view, merge duplicates and confirm. Keep original facts separate from AI rewrites. A missing graduation year or project metric stays missing until supplied.

### Manual entry

Use the same schema and validation as uploads. Autosave section drafts, handle optimistic concurrency, preserve partial dates and let users return later. New uploads propose changes; they must not overwrite confirmed facts silently. Career preferences are separate from factual experience.

### Rendering

The master resume is a typed content specification built from a confirmed profile version. Use conventional headings, reading order, selectable text, simple layout, restrained typography and no essential information encoded only in icons/images. PDF and DOCX renderers consume the same spec; plain-text preview is available.

Run parse-back checks on generated output: contact retention, section order, dates, bullet text and no missing/duplicated essential facts. Visual inspection covers overflow, clipped text and pagination. These are our formatting/quality tests, not an ATS vendor certification or guarantee of ranking.

### Tailoring

A tailored resume references profile version, master template version, job-description hash and model/prompt version. It may select/reorder confirmed skills/projects and rewrite language without adding facts. Every new quantitative assertion must point to a confirmed fact. Show diff and re-review before application. Editing facts invalidates dependent approvals; existing submitted application artifacts remain immutable history.

## 9. Career state machines

### Opportunity and application states

A saved job has its own shortlist state. An application has its own status; do not make every discovered job an application.

```text
Job: discovered → saved | skipped | archived

Application:
  draft → preparing → needs_input → ready_for_review → approved
       → submitting → applied → recruiter_response → interview → offer
                         └────→ rejected | withdrawn
                  └→ needs_confirmation
                  └→ failed_before_submit
```

Some transitions skip optional stages, but the server enforces permitted transitions and evidence. `applied` requires `confirmation_type = provider_confirmed | user_confirmed`, `applied_at`, actor, canonical job ID and immutable document bundle. A provider receipt may contain an ID, confirmation URL or safe captured confirmation. A user can mark a previously submitted external application with an explicit “I applied” action; label it **User-confirmed**, not provider-verified.

Clicking Apply, opening an external page, filling a form or saving a screenshot is not proof of submission. A submission timeout is not automatic failure. Corrections are appended as status events, not destructive history edits. Reapplication to a duplicate/changed listing requires explicit review.

### Matching

Normalize source IDs/URLs, salary currency and period, location, employment type, posting date and source freshness. Use strict source/job-ID dedupe and explainable similarity suggestions for cross-source duplicates; human review resolves uncertain matches. Preserve all source references after merge.

Use normalized weighted job fit, missing information and reasons. Do not inherit Career Agent's displayed weights unchanged: its documented dimension weights sum to 90 even though it describes a 0–100 score. Define our own tested denominator. A fit score is not hiring probability or a universal ATS score. Unknown salary/sponsorship does not pass a hard filter by guessing.

### Application adapters

Separate question interpretation from field manipulation. Reuse inspected Shadow DOM/label-mapping/loop-detection patterns only in authorized test fixtures or approved execution environments. Unknown/sensitive answers enter user review. Automatic submission is available only on explicitly supported and permitted providers; otherwise provide document download, answer copy and native Apply link.

## 10. LinkedIn implementation boundaries

OIDC authentication, full-profile import, publishing, inbox and application submission are different capabilities. Sign-in alone supplies neither complete career history nor permission to automate DMs or applications. See the official sources in CONTEXT.

Onboarding offers approved OAuth where suitable, own-profile export/upload, pasted sections and manual entry. An unavailable “Scan profile” action explains the supported fallback instead of claiming successful scanning. An optional local connector is not automatically compliant and never enables forbidden cloud functionality.

Store profile snapshots and advice separately. Each recommendation contains current text, proposed text, rationale, relevant confirmed facts and user decision. Never silently update LinkedIn. Resume/profile/LinkedIn differences are suggested corrections, not proof which source is right.

Build content/voice/story-bank tools as draft services. Networking screens use contacts, drafts, due follow-ups, consent and recorded outcomes. Inbox and profile fetch buttons are conditionally available from the account capability registry. Connection/follow/DM counts reflect only known application activity; explain outside activity is not fully observed.

## 11. Social modules and ownership

### Publishing / Management

One content item can have many revisions, media assets and per-platform variants. A calendar entry points to a specific revision and target account. Provider adapters validate actual content/media constraints and required permissions. Team approval is based on a specific version; an edit returns it to review.

Publishing proceeds through `draft → in_review → approved → scheduled → dispatching → published | failed | needs_confirmation`. An asynchronous upload creates `provider_processing`; it is not yet a published post. Multi-platform results are independent and visible; a partial failure never silently republishes successful channels. Use receipts and stable external IDs. Conditional comments require their own approval/capability, not inherited permission to publish any reply.

### AI Content Studio

Build voice/account memory, URL-to-draft, hook/framework selection, graphics briefs, scripts and repurposing first. Later introduce image render/edit, audio/video transformation and generative-provider adapters with per-job budgets, cancellation and deliverables. Distinguish a generated prompt from an actual PNG/video output.

Keep advanced Easel themes in the feature inventory: serial writing, paper explainers, e-commerce imagery, knowledge cards, charts/mind maps, music, subtitles, dubbing, short drama and regional formats. Do not claim support before that module's output test passes. Voice cloning requires ownership/permission verification and must not support impersonation.

### Research / Intelligence

Every brief has a question, source coverage, retrieval time, evidence references, extraction method and caveats. Compare outliers to creator-specific baselines, not just raw views. Treat missing/private data as missing. Route trend results to a draft brief; never automatically spam trending topics. Distinguish source-supported insight from model speculation.

### Outreach / Growth

Use a versioned sequence engine with working hours, per-recipient stop rules, deduplication, opt-out, account budgets and approval. Professional contact information comes from authorized/supplied sources. Discovery is not consent for unsolicited automated outreach. No fake engagement or account cycling. Bulk functions have preview and a shared hard stop.

### Scraping Infrastructure

Use allowed-source fetch adapters, policy-aware robots handling, throttling, response fixtures, checkpoints, progress and schema validation. Semantic extraction runs after fetching. Untrusted HTML cannot issue agent commands. Sanitization alone is not a complete prompt-injection defense. Restrict URLs/redirects/DNS resolution to prevent server-side request forgery and private-network access.

### OSINT / Identity Discovery

Keep this optional and separate from ordinary career onboarding. Support a user's own or authorized brand footprint with limited purpose, allowed sites, provenance, uncertain-match review, watch/diff and exports. Do not build facial identity search, private-account enumeration or sensitive inference. The S12 source remains recorded as a rejected biometric mechanism with reusable reporting concepts; it is not silently omitted.

## 12. Export design

### CSV

Export jobs, applications, profile sections, relationship records, content metrics and research results through the same permission checks as the UI. Offer a column/record preview, filters, timezone/currency labels and UTF-8 output. Neutralize formula-triggering spreadsheet cells; preserve original content internally. For very large exports, use streaming/chunked jobs and expiring private download URLs.

### Google Sheets

Use user-granted OAuth and least-privilege scopes appropriate to the selected spreadsheet flow. A service-account deployment may be an advanced self-host option, never a secret bundled into a public client. Store refresh tokens encrypted; handle revocation without deleting local records.

Initial sync is **one-way: platform → sheet**. Each exported entity has a stable ID column and sync version. Use raw values so user text is not interpreted as formulas. Do not assume row number remains fixed when a user sorts the sheet. Locate rows by stable ID or developer metadata, maintain a mapping and reconcile edits. Document which columns the app owns; preserve separate user-notes columns.

Batch operations within current provider limits. After an uncertain append/update response, inspect IDs/state before retrying. Do not duplicate rows merely because a network timeout occurred. Show last synced time, pending rows, failures and reconnect action. Manual “Sync now,” scheduled sync and CSV fallback are all required. External Sheets are user-owned exports; disclose deletion/retention behavior and offer explicit app-owned row removal, not a claim that local deletion erases every downstream copy.

## 13. API contract outline

All endpoint names below are proposed and must be finalized in the OpenAPI contract. Long operations return `202` with a run ID and status URL. Resource edits require a revision/ETag; validation conflicts return actionable structured errors. List endpoints are cursor-paginated.

| Route family | Capabilities / examples |
|---|---|
| `/v1/auth`, `/v1/me`, `/v1/workspaces` | Session, identity, membership, preferences; no privilege escalation through editable role fields. |
| `/v1/profile`, `/v1/profile/imports`, `/v1/profile/reviews` | Upload/manual drafts, field review, version confirmation. |
| `/v1/resumes`, `/v1/resumes/{id}/render`, `/v1/resumes/{id}/tailor` | Versioned document creation, parse-back check, diff, download. |
| `/v1/jobs`, `/v1/job-searches`, `/v1/matches` | Discovery, saved filters, dedupe, fit reasons. |
| `/v1/applications`, `/v1/applications/{id}/confirm`, `/v1/applications/{id}/events` | Prepare, approve, manual confirmation, history; creation is not submission. |
| `/v1/accounts`, `/v1/accounts/{id}/capabilities`, `/v1/accounts/{id}/policy` | Connection status, granted actions, quotas, pause/revoke. |
| `/v1/linkedin/profile-imports`, `/v1/linkedin/recommendations` | Own-profile analysis, consistency diff, accepted/rejected advice. |
| `/v1/contacts`, `/v1/conversations`, `/v1/message-drafts` | Scoped professional relationships and allowed inbox functions. |
| `/v1/content`, `/v1/calendar`, `/v1/publish-runs` | Drafts, variants, media, approvals, scheduled per-channel publishing. |
| `/v1/research`, `/v1/watches`, `/v1/footprint-reports` | Evidence-based briefs, notifications, optional authorized scans. |
| `/v1/approvals`, `/v1/runs`, `/v1/runs/{id}/events` | Review decisions, durable progress, cancel/pause/resume. |
| `/v1/exports`, `/v1/sheets-targets` | CSV/Sheets jobs, sync status, explicit disconnect. |
| `/v1/usage`, `/v1/subscriptions` | Workspace usage, invoices/provider references, plan changes. |
| `/v1/admin/*` | Authorized workspace management only. |
| `/v1/platform/*` | Super Admin package versions, global policy, diagnostics and audit. |
| `/internal/v1/worker-results` | Authenticated, task-scoped, revision/fencing-checked results. |
| `/webhooks/{provider}` | Signature-verified, timestamp-checked, deduplicated events with durable processing. |

Every external side effect carries an idempotency key and an approval reference. GET endpoints must not submit applications or publish content. For exports/search errors, do not leak another user's record existence. Provider secrets and stack traces must not appear in public error messages.

## 14. Package and cloud behavior

Model package definitions separately from immutable published package versions. Fields include name, visibility, billing interval, currency/price configuration, module entitlements, seats, connected accounts, AI/document/media budgets, storage, exports and worker concurrency. Leave production price amounts unset until approved; a demo plan must visibly be a demo.

A plan version may be draft, published or retired. Existing subscriptions reference a version and an explicit migration/renewal policy. Upgrades change entitlements only after verified billing state; downgrades explain paused excess work and retain data according to policy. Payment failures do not permit new billable work after grace conditions, but basic access/data export remain available.

Use a payment-provider interface; select and verify actual provider APIs during implementation. Verify webhook signature, deduplicate provider event IDs and reconcile out-of-order updates. Record usage as idempotent events with reserved/settled cost so retries do not charge twice. Show usage provenance and estimate-versus-actual differences.

Super Admin can create/publish packages and set provider/global limits. Workspace Admin can manage their subscription and lower operational budgets, not rewrite platform packages. Package “unlimited” may apply to an internal feature but never removes a provider cap or permission gate.

## 15. Security, AI and operational controls

Required controls include encrypted credentials, secret rotation, scoped tokens, CSRF protection, XSS prevention, upload isolation, SSRF protection, dependency scanning, throttled auth, audit logs and deletion propagation. Keep prompts and model outputs out of public logs when they contain resumes or private messages.

The agent only calls registered tools with schema validation. Webpages, resumes, job descriptions and messages are untrusted data. They cannot grant privileges, select arbitrary URLs, expose tokens or alter approval policy. Use minimal authorized facts per task; never send the entire workspace memory to a model. Provide configurable retention/provider consent and record which model processed which artifact.

Observability: correlation IDs, per-provider error categories, run age, outbox lag, index lag, budget utilization, expired approvals, ambiguous writes, webhook failures, storage usage and redacted audit trails. Provide a platform kill switch plus account/sequence/run pause controls. Avoid displaying a made-up “ban risk score”; show observed warnings and known state instead.

Backups must include PostgreSQL and private artifacts, with encryption and documented restore exercises. Treat Redis/search as rebuildable. Restore operations must replay deletion tombstones so previously deleted personal data is not casually reintroduced. Self-hosted maintenance and cloud operations use the same retention/account-disconnect semantics.

## 16. Build phases and release slices

| Phase | Deliverable | Exit gate |
|---|---|---|
| P0 — Source and contract preparation | Pinned source manifests, feature inventory/dispositions, boundary/contract definitions | No planned/beta/paid-only feature represented as verified upstream code. |
| P1 — Foundation | Monorepo, migrations, auth, roles, tenants, storage, queues, policy skeleton, three-section shell | Cross-user access tests and durable run smoke tests pass. |
| P2 — Profile and resume | Upload/manual review, profile versions, ATS-friendly master/tailored output | Synthetic PDF/DOCX parse-back, no fabricated facts, secure upload tests. |
| P3 — Career workflow | Jobs, filtering, matching, application preparation, manual confirmation, CSV | Full user journey without any live social automation. |
| P4 — LinkedIn setup and Sheets | Capability-aware own-profile import/advice, conflict review, reliable one-way Sheets | Unavailable API permissions produce a clear fallback; sync retry creates no duplicate rows. |
| P5 — Administration and cloud | Three dashboards, package builder, entitlements, billing adapter, usage and support | Role boundaries and billing event replay tested; no package raises platform permission. |
| P6 — LinkedIn workspace | Drafts, voice/story bank, networking queue, consented relationship tracking and analytics | Shared limits/approvals are enforced; unauthorized live actions stay disabled. |
| P7 — Social release | Publishing, media library, content drafts, research, trends, approved provider adapters | Per-channel receipts, edit invalidation and partial-failure recovery tests. |
| P8 — Advanced modules | Media production, authorized growth, scrape reliability, optional footprint tools | Each module has its own cost, permission, privacy and output-quality gate. |
| P9 — Production hardening | E2E/security/restore/load checks, self-host docs, cloud release and rollback | All release-blocking tests and source-disposition coverage pass with evidence. |

A useful first release is P1–P4 with manual/native application paths. This is a release sequence, not deletion of Social or advanced requirements. The source-feature registry and TASKS preserve later work.

## 17. Test and acceptance strategy

Tests listed here are required future checks, not already executed results. Use synthetic resumes, fixture sites, fake provider adapters and approved sandboxes. No tests submit real job applications or unsolicited social messages.

| Test ID | Acceptance condition |
|---|---|
| AT-001 | Upload PDF, upload DOCX and manual entry produce the same confirmed profile schema; import uncertainty remains visible. |
| AT-002 | Generated PDF/DOCX preserve confirmed contact/experience facts and reading order; overflow is caught by render/parse-back tests. |
| AT-003 | Unknown skill years, salary, sponsorship or legal answers do not become fabricated defaults. |
| AT-004 | Same job from multiple sources is linked/deduplicated; an ambiguous match is reviewable. |
| AT-005 | Clicking Apply does not mark Applied; confirmed receipt/user action does; timeout enters needs_confirmation. |
| AT-006 | Crash after external dispatch does not trigger an unreviewed duplicate submission. |
| AT-007 | Resume/content/recipient edit invalidates a previously approved action. |
| AT-008 | Two concurrent workers cannot consume the same last budget slot; recruiter DM consumes nested limits. |
| AT-009 | Same provider account across modules/workspaces shares budgets; reconnect/upgrade does not reset history. |
| AT-010 | Unauthorized action remains disabled despite a high package quota; 429/auth/challenge pause correctly. |
| AT-011 | Admin cannot read another member's private resume without an explicit grant; Super Admin support access is audited and consented. |
| AT-012 | Search results, counts, facets, autocomplete and downloads cannot leak another tenant's data. |
| AT-013 | Formula-like values in CSV/Sheets are safe; export preserves scope and correctly labels timezone/currency. |
| AT-014 | Sheets retry after uncertain response or user row sorting does not duplicate or overwrite unrelated rows. |
| AT-015 | A channel's successful publish is not repeated after another channel fails; processing is not prematurely Published. |
| AT-016 | Disconnect/revocation cancels pending account work and removes usable credentials. |
| AT-017 | Billing webhook replay/out-of-order events neither double-charge usage nor grant invalid entitlements. |
| AT-018 | DST, rolling windows and user timezone changes do not multiply daily budgets or lose scheduled tasks. |
| AT-019 | Malicious document/web instructions cannot exfiltrate secrets, call unauthorized tools or bypass approvals. |
| AT-020 | Upload bombs, parser timeouts, unauthorized URLs and private-network redirects are rejected. |
| AT-021 | Outbox redelivery, Redis restart and expired worker leases preserve durable state without double side effects. |
| AT-022 | Account deletion propagates to search, storage and owned background work; downstream export limitations are disclosed. |
| AT-023 | Keyboard navigation, responsive layouts and screen-reader status feedback cover the guided onboarding and review workflow. |
| AT-024 | Every audited upstream feature has a mapped adopted/merged/deferred/rejected disposition with source and reason. |
| AT-025 | An image prompt is not reported as an image asset; failed media generation cannot be published as a completed artifact. |
| AT-026 | Footprint results retain uncertainty; no face identity search or sensitive-attribute inference is exposed. |
| AT-027 | Package changes preserve data/export access and clearly explain unsupported/paused features. |
| AT-028 | Analytics distinguish unavailable data from zero and modeled suggestions from proven causal impact. |

The source selection benchmark should compare candidates on the same frozen fixtures: field recognition, fact preservation, duplicate handling, recovery, latency/cost and maintainability. Report measured results with environment/commit/date, not invented star counts or arbitrary 8/10 scores. Do not benchmark platform-ban avoidance.

## 18. Coding and release protocol

At session start, read the five files and current git state. Pick the next unblocked task, inspect its source evidence and state the task ID. Make a bounded change. Run the relevant tests and record command, exit code and result. A generated test file is not a passed test. A screenshot/mock component is not implemented backend behavior.

Before marking DONE: implementation exists, acceptance checks passed, migration/security impact reviewed, feature IDs mapped and PROGRESS updated. If external access is unavailable, mark the integration task BLOCKED and maintain a working mock/manual path. Never fabricate a live success to clear the task.

Before public release: verify pinned dependencies and compatibility, source provenance/notice obligations, secret scans, sample data privacy, exact installation commands, restore instructions, self-host portability and approved provider capabilities. Do not publish credentials, real resumes, raw sessions or source snapshots containing personal data.

Current package scope: **documentation only**. The five files are the AI coder's handoff, not a finished executable product.
