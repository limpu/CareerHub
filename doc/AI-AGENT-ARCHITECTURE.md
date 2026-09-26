# AI Agent Architecture
## Dashboard-configurable AI • Token budgets • Low-RAM / low-bandwidth operation

Before starting or continuing any work, read README.md, CONTEXT.md, PLAN.md, AI-AGENT-ARCHITECTURE.md, TASKS.md, and PROGRESS.md in full and treat them as one interconnected source of truth. Cross-reference requirements, architecture decisions, AI-agent rules, task dependencies, implementation status, and prior decisions across all six files before making changes. Do not interpret any file in isolation, silently override an existing decision, duplicate functionality, change the approved stack, or mark work complete without implementation and verification evidence. After each meaningful implementation step, update TASKS.md and PROGRESS.md so the next AI/coding session can resume accurately without losing context.

**Product:** Career / LinkedIn / Social Platform  
**Stack:** Next.js + TypeScript + Tailwind + Go + PostgreSQL + Redis + Meilisearch  
**Document version:** 1.0 • **Prepared:** 2026-09-17  
**Status:** Implementation specification — this document is an architectural specification, not application code or a live deployment.

> **Core Decision:** A shared Go AI Gateway will be implemented. Provider connections, API credentials, models, and task-wise budgets can be configured from the Dashboard. Routine tasks will be handled by deterministic services; small, bounded AI requests are dispatched only when reasoning or generative writing is strictly required. Six logical agents do not imply six physical servers, six distinct models, or six simultaneous calls.

This document serves as an extension of the existing `README.md`, `CONTEXT.md`, `PLAN.md`, `TASKS.md`, and `PROGRESS.md` files. The prior five files remain untouched in this deliverable. Guidance for merging the task/decision IDs below is documented in Section 25. Core principles regarding privacy, provider permissions, human approvals, and verified Applied/Published states remain strictly unchanged.

**Proposed numeric values represent configuration defaults and testing targets; they are not measured performance benchmarks, guaranteed provider prices, official platform limits, or production capacity warranties.**

---

## 1. Product requirements and non-negotiable rules

| ID | Requirement |
|---|---|
| REQ-AI-001 | Approved AI provider connections can be added, tested, edited, disabled, rotated, and revoked from the Dashboard. |
| REQ-AI-002 | Provide native adapters and an OpenAI-compatible adapter; unsupported APIs require dedicated new adapters. |
| REQ-AI-003 | Strictly isolate user-private, workspace-shared, and platform-managed credentials and billing. |
| REQ-AI-004 | Task-driven model selection; never route routine small tasks to unnecessarily expensive models. |
| REQ-AI-005 | Validate and reserve input/output/cost/attempt budgets prior to initiating each provider call. |
| REQ-AI-006 | Never dispatch redundant model calls for identical requests or cache hits; regeneration must remain an explicit user action. |
| REQ-AI-007 | Minimize context size without silently discarding critical candidate facts or eligibility criteria. |
| REQ-AI-008 | Local inference, browser automation, OCR, and media workers must not start without passing resource admission gates. |
| REQ-AI-009 | Enforce bounded concurrency, streamed I/O, compact queue payloads, and incremental synchronization. |
| REQ-AI-010 | AI budget limits or provider API failures must not disrupt deterministic editing, viewing, job tracking, or basic file exports. |
| REQ-AI-011 | AI suggestions must never autonomously approve external actions, platform permissions, or profile facts. |
| REQ-AI-012 | Usage metrics, estimated costs, retries, queue latency, RAM, and bandwidth must be transparently explainable on the dashboard. |

**V1 Out of Scope:** Mandatory LangGraph, CrewAI, AutoGen, Temporal, dedicated Python AI microservices, vector databases, per-agent containers, or perpetual autonomous loops. These must not be introduced without evidence-backed architectural decision records (ADRs).

---

## 2. Runtime architecture

```text
Next.js Dashboard
    |
    | authenticated request; no provider key in browser calls
    v
Go API / Domain services
    |
    +---- deterministic logic ------------------------------+
    |     auth, forms, ATS rendering orchestration,          |
    |     filtering, tracking, exports, quotas               |
    |                                                       |
    +---- AI Gateway (Go package, not mandatory new server)  |
          |                                                 |
          +-- Authorization / data-sharing policy           |
          +-- Context builder / token estimator             |
          +-- Exact-result cache / duplicate-request guard  |
          +-- Model router / capability checks              |
          +-- Budget reservation / fair queue               |
          +-- Provider adapter                              |
          +-- Output validation / usage settlement          |
          |                                                 |
          v                                                 |
      Approved external AI API                              |
      OR explicitly enabled private/local endpoint          |
          |                                                 |
          v                                                 |
      Structured draft / recommendation                     |
          |                                                 |
          +---- Approval / normal domain workflow ----------+
                            |
                            v
                    Go / isolated TS worker
                            |
                    permitted external action
```

**Durability:** PostgreSQL owns runs, input versions, credential metadata, usage, budgets, approvals, and outcomes. Redis Streams manages transport and bounded coordination. Meilisearch serves as an authorized retrieval projection; it is neither AI memory nor a permission authority. Private object storage holds documents and media. These boundaries strictly conform to ADR-001 through ADR-010 in `PLAN.md`.

**Deployment:** Initially, package the AI gateway directly within existing `services/core/cmd/api` and `cmd/worker`. Do not create a separate AI microservice until concrete load metrics demonstrate the necessity.

---

## 3. Dashboard: Settings → AI Providers

### 3.1 Add-provider flow

```text
Add AI Provider
  → Choose protocol / approved provider preset
  → Select ownership: Personal / Workspace / Platform
  → Enter API credential or choose approved secret reference
  → Select endpoint and model
  → Check supported capabilities
  → Run an optional, budgeted connection test
  → Set task routes and usage limits
  → Review which data may leave the platform
  → Save / Activate
```

### 3.2 Provider form

| Field | Behavior |
|---|---|
| Display name | Human-readable label (e.g., `My writing provider`); non-sensitive metadata. |
| Adapter type | `openai_responses`, `openai_chat_compatible`, `anthropic_native`, `gemini_native`, approved optional adapter। |
| Base URL / endpoint | Approved origin; in public cloud deployments, adding a custom origin requires administrator approval. |
| API credential | Write-only form field; stored securely and displayed only as a masked identifier post-save. Raw key is never revealed again. |
| Model ID / deployment ID | Exact provider identifier; allows manual specification. Do not assume all providers expose a working `/models` endpoint. |
| API version | Explicit, versioned parameter where required by specific adapters (e.g., Anthropic, Azure). |
| Ownership / allowed users | Personal, workspace, or platform scope; enforced strictly via server-side authorization. |
| Allowed data classes | Public content / private career / permitted messaging context / media। |
| Data residency / retention policy | References published provider policies with review dates; explicitly labeled as unknown if unverified. |
| Capability status | Verified / Declared / Unsupported / Not tested। |
| Price card | Currency, input/output/cache/media rates, source URL, verification timestamp, and billing semantics. |
| Budget | Configurable per request, daily, billing cycle, run, and credential pool. |
| Timeout / rate limits | Provider-aware timeout and rate threshold configurations; cannot exceed administrative ceilings. |
| Fallback | Disabled by default; requires explicit approved fallback model/connection and a maximum cost cap. |
| Enabled agents | Profile, Career, Application, LinkedIn, Content, Research। |

**Test connection:** Endpoint validation → authentication/capability probe → optional minimal synthetic generation. Never transmit real resumes, private messages, or complete chat history during testing. Generation probe is limited to exactly one request with a strict budget (e.g., 256 input + 64 output tokens). If a model does not support these caps, require explicit confirmation before testing. Test probes are fully metered; repeated tests must not trigger automatically on dashboard reload.

**Tooltip examples:**

> "This credential is used exclusively on the server side. It is never exposed to other users, client browsers, or AI prompts."

> "Connection tests may incur minimal provider usage. No personal resumes or sensitive user data will be transmitted."

> "If model autodetection is unavailable, enter the exact model ID from your provider dashboard. Saving credentials does not imply verified model capability."

### 3.3 Dashboard tabs

| Tab | Contents / Controls |
|---|---|
| Providers | Connections, model access, credential rotation, last successful request, known errors। |
| Agent routing | Mapping tasks to models, Economy/Balanced/Advanced mode toggles, and fallback policies. |
| Budgets | Real-time used / reserved / available allocations; user, workspace, and payer-specific views. |
| Usage | Token categories, estimated cost, cache hits, successful/failed attempts। |
| Runs | Queued, running, waiting for input, budget-blocked, cancelled, usage-pending। |
| Resource limits | Concurrency caps, browser/media worker permissions, queue depth limits, and RAM thresholds. |
| Privacy | Data sharing scopes, user consent flags, retention rules, cache invalidation, and credential revocation. |

---

## 4. Architectural definition of "Any API" support

**The architecture is provider-agnostic; however, arbitrary third-party APIs are not universally plug-and-play.** The request/response protocol, authentication scheme, and model capabilities must be compatible.

| Provider family | Implementation path | Caveat |
|---|---|---|
| OpenAI | Native Responses adapter; separately tested Chat Completions route if required | Endpoint-specific tool, usage, and output mappings. |
| Anthropic | Native Messages adapter | Custom auth headers, content blocks, and stop/usage/cache mappings. |
| Google Gemini | Native generation adapter or verified OpenAI-compatible endpoint | Capabilities may vary depending on the chosen endpoint. |
| OpenAI-compatible hosted services | Generic chat adapter + approved endpoint/model manifest | Candidate services like OpenRouter, Groq, or DeepSeek; each concrete endpoint and model must be tested and validated individually. |
| Ollama / private model server | Tested compatible adapter, explicit private-network configuration | Local model loading is disabled by default on the application server. |
| Azure / Bedrock / Other enterprise APIs | Dedicated adapter | Custom deployment names, request signing, or IAM workload identity mean a simple API key is insufficient. |
| Image / audio / video APIs | Dedicated typed media adapters | Do not assume text model credentials grant access to specialized media features. |
| Unknown vendor | New reviewed adapter or validated compatible protocol test | Never execute unreviewed user-supplied code, arbitrary curl snippets, or dynamic JavaScript mappings. |

Google documents an OpenAI compatibility interface; Ollama explicitly implements an OpenAI API subset. Therefore, capabilities such as `supports_streaming`, `supports_json_schema`, `supports_tools`, `supports_vision`, and `reports_usage` must be verified per selected endpoint and model, rather than assumed by brand name. [S4], [S5]

### 4.1 Capability manifest

Track per model and version:

```text
protocol + endpoint version
model identifier + verified date
context window + maximum generation budget
streaming / structured output / tools / vision support
supported reasoning controls
usage categories and nesting rules
cache features and pricing semantics
cancellation and status-retrieval support
request, output and response-size limits
allowed data classes / regions
compatibility test results
```

Unsupported settings must not be silently discarded. For example, if a selected model lacks strict JSON schema support, the gateway must route through a secondary validation parser or mark the task unavailable. If a model cannot enforce mandatory budget limits, it must not be enabled for automated background tasks.

---

## 5. Roles, credentials, and billing ownership

| Capability | Super Admin | Workspace Admin | User |
|---|---|---|---|
| Approved provider origins/adapters | Manage | Request addition | Request addition |
| Platform-funded credential | Manage | Use if entitled | Use if entitled |
| Workspace BYOK | Policy oversight | Add/rotate within workspace | Use if delegated |
| Personal BYOK | No default secret visibility | No default secret visibility | Add/rotate own key |
| Model allowlist / global ceilings | Manage | Narrow workspace policy | Choose within allowed policy |
| Personal budget | Policy ceiling | No unauthorized private-data access | Set lower own ceiling |
| Usage visibility | Aggregate operations | Workspace-funded usage | Own usage |
| Raw private prompt/resume access | No automatic access | No automatic access | Own data |

**BYOK = Bring Your Own Key.** Three billing modes are supported: `platform_managed`, `workspace_byok`, and `personal_byok`. The responsible payer must be displayed prominently in the UI. If a personal key fails, the system must never silently fall back to platform funds or another workspace member's credential.

The platform meters only the usage routed through its own gateway. If the same API key is utilized across third-party applications, external billing cannot be tracked within this dashboard. The UI must clearly separate BYOK direct provider invoices from platform hosting/subscription tiers.

### 5.1 Secret handling

API credentials must be transmitted via TLS directly to the Go backend and protected using envelope encryption in the secure credential store. Master encryption keys must reside strictly outside the database and source repository. Secrets must never appear in application logs, traces, Redis payloads, support bundles, analytics events, or model prompts. Never expose keys in `NEXT_PUBLIC_*` environment variables, browser localStorage, or client-side requests.

Modifying a base URL, ownership scope, or adapter requires mandatory connection re-validation. Credentials must never be automatically forwarded to modified origins. When a key is revoked, queued and new dispatch attempts are immediately blocked; inflight requests must be reconciled independently.

### 5.2 Custom endpoint security and egress controls

Never accept arbitrary endpoints in public cloud deployments to prevent turning the server into an open proxy. Enforce an approved HTTPS origin allowlist, strict port/path constraints, TLS certificate validation, and controlled egress. HTTP redirects are disabled by default. Validate DNS and resolved IPs upon connection; prevent loopback, private, link-local, and cloud-metadata addresses (IPv4/IPv6) as well as DNS rebinding attacks using combined application and network firewall rules. [S8]

For self-hosted local Ollama endpoints, exact network exceptions may be granted under explicit operator control. In a multi-tenant public cloud, users must not be permitted to point endpoints to localhost or internal network subnets. Accessing a user's local machine from the cloud requires a dedicated, authenticated tunneling connector—specifying `localhost` in the URL is strictly prohibited.

---

## 6. Six logical agents

| Agent | Responsibility | AI Output | Prohibited Autonomous Actions |
|---|---|---|---|
| Profile Agent | Structures extracted resume text into profile facts; identifies missing fields | Candidate facts + source spans + clarifying questions | Cannot silently overwrite user-confirmed profile data |
| Career Agent | Tailors resumes, generates match analyses, drafts cover letters and prep notes | Grounded suggestions / versioned drafts | Cannot fabricate unverified qualifications or achievements |
| Application Agent | Drafts answers to unmapped free-text job application questions | Draft response + supporting fact IDs + unresolved fields | Cannot guess legal eligibility or submit applications autonomously |
| LinkedIn Agent | Optimizes headline, about section, notes, and professional responses | Before/after diffs and editable drafts | Cannot connect accounts, dispatch messages, or update profiles without human confirmation |
| Content Agent | Generates tone-aware posts, hooks, article drafts, and media briefs | Versioned content drafts and asset plans | Cannot publish content unreviewed or trigger unbounded media generation |
| Research Agent | Synthesizes industry trends, discussions, and competitor insights from approved sources | Cited reports with explicit coverage boundaries | Cannot execute autonomous unbounded crawls or infer private individual identities |

An agent is defined formally as: `task key + prompt version + output schema + data scope + model route + budgets + validators`. An operation like a headline rewrite must never trigger invocations across unrelated agents.

**Shared writing primitives:** The LinkedIn Agent and Content Agent reuse identical text-generation gateway and prompt primitives; separate modules must never trigger duplicate generation passes for the same content piece.

### 6.1 Tasks with zero AI invocations

User signup/login, role/package enforcement, manual profile forms, required field validation, deterministic ATS layout generation, PDF/DOCX rendering, job deduplication, standard database filtering, updating confirmed Applied statuses, calendar tracking, publishing state reconciliation, CSV/Sheets export, quota checks, numerical analytics, transactional notifications, and credential health checks require **0 LLM calls** by default.

Resume ingestion extracts text deterministically first. LLM calls are reserved solely for semantic taxonomy mapping and resolving ambiguous sections. Missing mandatory fields detected by deterministic rules must never trigger model calls.

### 6.2 Orchestrator

In V1, standard UI actions and workflows explicitly dispatch tasks; no dynamic LLM router is needed. If optional conversational orchestration is introduced later, it must enforce a maximum of 4 planned steps, 3 total provider attempts per run, and strict budgetary caps. Recursive agent-to-agent delegation and infinite critic/rewrite loops are prohibited.

If a model proposes an action, the Go domain policy must validate it independently. A model returning `requires_approval: false` in its payload does not constitute an authorized permission grant.

---

## 7. AI request lifecycle

```text
1. Authenticate user, workspace and resource ownership
2. Freeze task/profile/job/content/prompt versions
3. Validate idempotency key and look for existing run
4. Resolve approved model + payer + capabilities + privacy policy
5. Build the smallest sufficient authorized context
6. Check exact-result cache under the resolved route/version
7. Estimate tokens, bytes, deadline and maximum attempt cost
8. Atomically reserve all relevant budgets; enqueue reference
9. Fair scheduler admits request when a worker slot is free
10. Recheck credential, permission, cancellation and stale inputs
11. Call provider once; record provider request ID when available
12. Validate schema, completeness, source references and facts
13. Persist output and settle reported/estimated usage
14. Return draft or needs-input state; cache eligible output
15. External action remains a separate approved domain workflow
```

Each retry generates a new `attempt_id` under the same persistent `run_id`. Database transactions must never remain open while awaiting external network responses. Output generation, user approval, and external platform actions are strictly decoupled events.

### 7.1 Result envelope — illustrative contract

```json
{
  "run_id": "run_example",
  "task": "career.resume_tailor",
  "status": "needs_review",
  "profile_version_id": "profile_v7",
  "prompt_version": "resume_tailor_v1",
  "result_ref": "artifact_reference",
  "supporting_fact_ids": ["fact_12", "fact_18"],
  "missing_information": [],
  "usage_status": "reported",
  "cache_hit": false,
  "external_action_performed": false
}
```

This schema illustrates the intended design contract; it does not represent an already deployed live endpoint.

---

## 8. Model routing: Economy default

| Mode | Policy |
|---|---|
| Economy — default | The lowest estimated-cost approved model that passes task evaluations, operating with the smallest sufficient context. |
| Balanced | User/workspace selected route; bounded one-step escalation allowed if explicitly configured। |
| Advanced | User explicitly requests larger context or higher-capability models; requires explicit cost estimate and user confirmation. |
| Manual / AI paused | Deterministic profile editing, template-based resumes/exports, and existing records remain fully operational; AI generation disabled. |

`economy_text`, `quality_text`, `vision_optional`, and `media_optional` serve as configurable aliases rather than hardcoded commercial model identifiers. A low-cost model must not be routed to production simply because it is cheap unless it successfully passes relevant multilingual fixtures and factual validation checks.

**Fallback defaults:** Disabled by default. If enabled, allow at most one alternate attempt within the same task budget. The system must never silently switch to a more expensive model tier, an unapproved vendor, an altered data residency jurisdiction, or an alternate payer. If an initial request times out after potential upstream dispatch, automatic fallback is prohibited to prevent duplicate billing risks.

---

## 9. Token budget: task-wise hard caps

All input caps encompass the system prompt, tool/schema declarations, conversational history, and retrieved context chunks. The adapter manifest defines how reasoning or hidden thinking tokens are billed against the generation cap. A short visible completion does not guarantee low cost—for example, OpenAI reasoning tokens are billed as output tokens. [S2]

### 9.1 Initial Economy task budgets

| Task | Max input tokens / attempt | Max billable generation tokens / attempt | Normal calls | Max provider attempts / run |
|---|---:|---:|---:|---:|
| Profile semantic extraction | 6,000 | 2,000 | 0–1 | 2 |
| Job-fit explanation, one shortlisted job | 1,600 | 300 | 0–1 | 2 |
| Resume tailoring | 4,000 | 1,800 | 1 | 2 |
| Cover letter | 2,500 | 700 | 1 | 2 |
| Application answer, one unknown question | 1,200 | 250 | 0–1 | 2 |
| LinkedIn profile optimization | 5,000 | 1,600 | 1 | 2 |
| Connection/message/comment draft | 1,200 | 250 | 1 | 2 |
| One social post | 2,500 | 800 | 1 | 2 |
| Five short content ideas | 2,000 | 800 | 1 | 2 |
| Research brief from selected evidence | 8,000 | 1,800 | 1 | 3 |
| Analytics insight from SQL aggregates | 1,500 | 500 | 1 | 2 |

**The maximum attempt limit encompasses all schema repairs, retries, fallbacks, and chunk iterations.** This cap applies across the entire run rather than resetting per stage. If a research task requires 3 attempts across preprocessing and synthesis, all calls are counted against the budget. If extensive documents or media scripts exceed standard caps, the system must request approval for an expanded workflow estimate rather than silently truncating and falsely presenting a "complete" result.

If a model's minimum generation or reasoning requirement exceeds configured caps, route to an alternative validated model or prompt for explicit budget approval. Unsupported parameters such as `temperature`, `reasoning_effort`, or `max_tokens` must not be sent indiscriminately to incompatible provider APIs.

### 9.2 Reduce calls before reducing quality

Prioritize caches, structured records, and deterministic business rules first. Never rely on an LLM's self-reported confidence as verification of factual accuracy. Enforce deterministic schema validation and anchor facts against verified source identifiers. Allow at most one budgeted JSON repair pass on syntax errors; missing candidate facts require human clarification rather than model hallucination.

Standard form editing must not trigger automatic AI runs. Field autosave may be debounced, but saving records and initiating paid generation remain distinct user actions. "Regenerate" actions create a new version and track new usage rather than destructively overwriting previous results.

### 9.3 Job discovery example

An **illustrative workflow**, not a guaranteed efficiency benchmark:

```text
100 discovered jobs
  → normalize + dedupe + explicit hard filters in Go
  → deterministic preliminary ranking
  → user can inspect all jobs and scoring criteria
  → AI explains only the selected/top 10
  → resume + cover letter only for the 2 the user chooses
```

In this pipeline, 90 unselected jobs incur zero AI explanation calls, and resumes or cover letters are never batch-generated for all 100 listings. Preliminary rankings remain transparent, and listings with ambiguous phrasing are never permanently hidden.

---

## 10. Monetary budgets, metering, and overspend prevention

### 10.1 Budget scopes

Spending limits are enforced across nested scopes: `request → run → user → workspace → credential pool/payer → platform`. Each scope tracks consumed and reserved allocations independently. Daily, rolling window, and monthly reset semantics must be explicitly defined. Routing requests through alternate modules or paths cannot circumvent shared payer caps.

**Suggested UX:** Emit an 80% quota warning, a 95% prominent notice, and block new billable calls at 100%. Notifications must be dispatched deterministically without AI. Specific production currency thresholds are not mandated by this document; the Super Admin configures limits based on validated price cards and subscription packages.

### 10.2 Before-call reservation

```text
available(scope) = limit(scope) - settled(scope) - unresolved_reserved(scope)

reserve_before_send = conservative billable input estimate
                    + permitted generation upper bound
                    + cache-write / storage upper bound when applicable
                    + explicitly bounded tool/media charges
```

All relevant budget reservations must occur atomically within a brief PostgreSQL transaction. Redis serves as an ephemeral coordination accelerator, not the durable ledger of record. If available balance is insufficient, external provider calls are rejected immediately. Retries require a separate atomic reservation.

Normalize raw provider usage categories into non-overlapping billing line items. If a provider bundles reasoning tokens into reported `output_tokens`, avoid double-charging by counting reasoning separately. Distinguish cache read/write tokens from standard input tokens based on vendor semantics. Represent currency using integer micro-units or precise database decimals, never floating-point types.

### 10.3 Accounting states

| State | Meaning |
|---|---|
| `estimated` | Pre-execution token and cost reservation; not a finalized invoice amount. |
| `reported` | Metric calculated directly from usage reported in the provider API response. |
| `pending_reconciliation` | Request was dispatched, but final status or usage could not be confirmed. |
| `reconciled` | Ledger updated following provider verification or controlled audit reconciliation. |

If an operation outcome is indeterminate, never release reserved funds by assuming zero cost simply because an in-memory TTL expired. A request may have completed on the provider side despite a local network timeout; keep usage in a pending state, suppress blind retries, and apply a bounded reconciliation policy. Cancellation does not refund already incurred provider charges.

**Spending cap boundaries:** The gateway regulates internal request admission; it cannot replace the provider's upstream billing enforcement. Fluctuations in price cards, token estimation variances, unmetered usage flags, or sharing keys across external apps preclude absolute guarantees against upstream invoice totals. Upstream spending limits must also be configured on the provider console.

Routes with unverified pricing must be disabled for platform-funded automated runs. Personal BYOK connections may permit an operator-approved `token_only` tracking mode, provided the UI does not falsely guarantee monetary caps. Unattended workflows must remain disabled for models with unknown context, output, or usage semantics.

---

## 11. Context minimization and safe memory

### 11.1 Context builder

Fetch only task-relevant fields rather than hydrating the entire candidate profile or historical database. If drafting a cover letter requires specific work experience, skills, and verified employer details, exclude sensitive identifiers like passport numbers, complete postal addresses, or private message threads.

Parse and version raw resume text once. Subsequent downstream tasks ingest structured `profile_facts` and provenance IDs. Scraped job HTML must undergo text extraction and sanitization first; scripts, stylesheets, navigation bars, and repetitive boilerplate must never enter the model prompt.

Meilisearch may retrieve top authorized candidate matches; full records must be re-authorized and hydrated from Go and PostgreSQL. Vector search and embeddings are not mandatory in the initial release. Persistent memory is modeled as versioned structured database records, not unbounded in-memory conversational transcripts.

### 11.2 Long inputs

Do not infer exact token counts from raw UTF-8 byte lengths; specifically, avoid simplistic `characters / 4` heuristics on mixed Bengali and English prose. Utilize provider-compatible tokenizers or official token-counting endpoints where available. Otherwise, employ a conservative estimation heuristic with documented safety margins.

For extensive inputs, implement section-aware selection or chunking while recording `coverage: full | partial` along with any omitted sections. If fact extraction was incomplete, the profile must not be marked complete. Maintain source citation spans alongside prompt inputs to facilitate suggestion verification.

### 11.3 Chat memory

Stateless form tasks must not transmit conversational history. For interactive chat workflows, default to bounded recent exchanges, a compact snapshot of verified facts, and relevant source evidence (e.g., maximum 6 recent messages plus a 600-token summary, staying strictly within the task input ceiling).

Avoid regenerating summaries on every message turn; trigger updates to the persisted summary only when message thresholds are crossed. Summary calls consume from the active run and user budget. Verified candidate facts must never be reconstructed from conversational summaries; they must remain anchored in original structured records.

---

## 12. Caching: distinguishing cost, tokens, and bandwidth

### 12.1 Exact-result application cache

If a validated existing output exists for the identical authorized task, input fingerprint, and model version, serve it directly without dispatching a new provider call. These cached responses consume zero new model tokens; the UI displays historical usage records.

Cache key constituents:

```text
workspace + owner/resource visibility + authorization revision
agent/task + canonical input hash
profile/job/document/voice versions
prompt + schema + validation versions
model + adapter + generation settings
provider-data-policy / consent version
locale + explicit regeneration nonce when requested
```

Secret credentials must never enter cache keys. Store hashed or HMAC-derived identifiers; private payloads must never reside in shared global cache namespaces. Re-check permissions prior to reading cache entries and invalidate on user deletion or revocation. One user's resume generation output must never be served to another user.

Store large cached artifacts in private object storage, maintain metadata references in PostgreSQL, and keep only compact lookup keys in Redis. Never maintain an unbounded in-memory cache for artifact reuse. Enforce entry counts, maximum byte limits, and scheduled cleanup eviction routines.

**Prohibited from success caching:** API errors, partial generations, expired authorizations, indeterminate provider results, application submissions, message dispatches, or social publishing side effects.

### 12.2 Provider prompt cache

On supported providers, stable system instructions and schemas can be formatted as reusable prompt prefixes. Cache boundaries and pricing mechanics must be handled according to vendor specifications; identical system prompts do not guarantee an upstream cache hit. OpenAI and Anthropic document specific prefix-length criteria, and Gemini context caching is dependent on model and endpoint configurations. [S1], [S3], [S13]

An upstream provider prompt cache hit does not mean the final completion is pre-computed. New generated tokens remain fully billable, and prompt cache writes or storage retention may incur separate charges. Furthermore, prompt caching still requires transmitting the request payload, meaning it does not automatically reduce upstream network egress.

Exact-result caching, compact context windows, and eliminating redundant calls represent the primary avenues for efficiency. Never downsize budget reservations based on hypothetical prompt cache discounts; reconcile against actual reported usage after the response is received.

---

## 13. Retry, fallback, and schema failure policies

| Failure | Required handling |
|---|---|
| Invalid credential / revoked key | Suspend connection; notify account owner; suppress repeated retries. |
| Unsupported model/capability | Return descriptive setup error; select compatible validated model. |
| Invalid input/context limit | Trim to smaller sufficient context or request expanded budget approval; suppress endless retries. |
| Explicit provider rate limit | Respect response backoff/reset; bounded reschedule। |
| Connection failed before dispatch is certain | One bounded retry allowed within remaining attempts/budget। |
| Timeout after possible dispatch | Mark usage as pending reconciliation; suppress automatic duplicate paid calls. |
| Invalid structured result | Reject incomplete result; at most one budgeted repair if appropriate। |
| Missing fact / unsafe inferred answer | Transition to `needs_input`; never output fabricated guesses. |
| Safety refusal | Surface user-facing notice; do not attempt to bypass filters by switching providers. |
| User cancellation | Stop future steps; cancel supported request best-effort; settle incurred usage। |

While strict structured output is valuable, applications must gracefully handle safety refusals, truncated outputs, and application-level factual errors. Valid JSON syntax does not guarantee factual truth. [S9]

Budget constraints, provider circuit breakers, and queue saturation must terminate retry loops promptly. Chaining 5 sequential fallback providers in a global loop is strictly prohibited.

---

## 14. RAM optimization and worker admission

### 14.1 Low-resource defaults

| Setting | Initial design default |
|---|---|
| AI execution | Hosted API; no local model loaded on app host। |
| Concurrent AI calls | 1 per user; 2 installation-wide in Economy pilot। |
| Agent processes | Zero dedicated per-agent processes; definitions loaded on demand। |
| Document/OCR/render jobs | Shared heavy-job semaphore; initially 1 job at a time। |
| Browser worker | Off unless permitted feature explicitly enabled; maximum 1 active session in low-resource profile। |
| Media generation/transcoding | Off by default; separate opt-in queue and resource budget। |
| Queue payload | References only; no PDFs, base64 images, secrets or whole prompts। |
| In-process hot cache | Bounded; illustrative maximum 8 MiB per Go process। |
| Task registry | Ingest only selected prompts and schemas; never load all upstream skill files into memory. |

**Architectural trade-offs:** Reducing concurrency increases queue latency for background tasks. Cloud provider APIs keep model weights off local RAM, but introduce network latency, external data transfers, and per-token costs. Local inference demands substantial server RAM/VRAM, where concurrent contexts expand memory footprints significantly, as documented in Ollama's resource guides. [S6]

### 14.2 Process controls

In Go, employ bounded worker pools, reusable HTTP transport clients, finite connection pools, strict request deadlines, and streamed I/O. Note that `GOMEMLIMIT` acts as a **soft limit** for Go runtime memory, not a hard cap on overall container RSS or external Chromium processes. Container and OS-level memory limits and allocation monitoring must be configured independently. [S7]

The Node worker V8 heap ceiling does not account for total process RSS; Chromium instances and native buffers must be monitored separately. `next dev` is not a production runtime; builds should execute on dedicated CI hosts. Unbounded `Promise.all` batches, spawning goroutines per record, and reading full datasets into memory are prohibited.

### 14.3 Illustrative resource envelope

The following memory allocations illustrate initial container limits on a constrained **8-GiB test host**; they are not minimum hardware warranties or performance benchmarks. Browser and media workers are disabled by default, and resource-heavy jobs are serialized. Tune limits based on observed RSS, CPU, and disk I/O.

| Process | Example container ceiling | Additional setting / note |
|---|---:|---|
| Next.js production web | 512 MiB | Excludes build-time memory spikes. |
| Go API | 256 MiB | Example `GOMEMLIMIT=192MiB`। |
| Go worker | 256 MiB | Example `GOMEMLIMIT=192MiB`। |
| PostgreSQL | 1,024 MiB | Bounded connection pools, restricted work memory, and batch queries. |
| Redis | 256 MiB | Configured with `maxmemory=128mb`; reserves remainder for overhead and persistence buffers. |
| Meilisearch | 1,024 MiB | Start with one indexing thread and explicit indexing-memory budget। |
| On-demand TS document helper | 1,024 MiB | Measure renderer/native subprocess peak; one heavy job। |
| Optional browser process group | 1,536 MiB | Not always running; separate account isolation। |

These example ceilings sum to 5,888 MiB, reserving remaining host memory for the operating system, filesystem page caches, native runtimes, and safety headroom. Setting ceilings does not guarantee isolation if subprocesses are unmanaged. Environments with lower memory may require remote browser workers or reduced concurrency; verify via load testing before production deployment.

### 14.4 Pressure response

Halt admission of resource-intensive tasks at a configurable warning threshold. At a critical threshold, suspend background research, indexing, and media pipelines to ensure core UI and API stability. Recommended baseline thresholds: warning at 75% and critical at 85% of effective memory limits, incorporating hysteresis buffers before resuming. Idle browser worker shutdown threshold is recommended at 60 seconds, resuming from persisted states when needed.

If an active user session requires browser-based approval, do not trigger idle shutdowns abruptly; maintain bounded interactive leases. Emergency kill switches halt pending jobs immediately; they cannot reverse external platform actions that have already executed.

---

## 15. Redis, PostgreSQL, and Meilisearch controls

### Redis

Maintain the existing architecture's **Redis Streams** pipeline; do not introduce conflicting proprietary queue formats. The PostgreSQL outbox remains the authoritative source of truth. Workers process small, bounded message batches and acknowledge only after durable transaction completion.

Configure queue and lock instances with a `noeviction` policy so memory pressure does not silently drop locks or queue payloads. Under `noeviction`, Redis rejects writes when capacity is exhausted; the application must handle this by pausing new task dispatches cleanly. [S10]

Selecting different Redis database numbers does not isolate eviction policies. Expire cache entries using explicit application-level TTLs and size policies; provision separate Redis instances for caching if workloads dictate. Before trimming streams, verify that consumer acknowledgments and PostgreSQL state are synchronized; `MAXLEN` must never truncate pending, unacknowledged jobs.

### PostgreSQL

Persist runs, usage metrics, and entity references in relational tables; never store large media binaries or base64 blobs in rows. Enforce paginated queries, restricted column projections, composite tenant/timestamp indexes, brief transaction lifecycles, and aggregated reporting tables. Avoid persisting individual token deltas per row; batch incremental progress updates. Ensure usage settlement operations are strictly idempotent.

### Meilisearch

Index only explicitly designated searchable fields in Meilisearch. Never index raw resumes, complete private inboxes, or full scraped HTML documents. The Go service facade must authorize search queries and facet filters; tombstone records ensure deleted entities are not hydrated even during indexing latency.

Explicitly configure indexing threads and memory allocations; note that this indexing budget is distinct from total process RAM limits. Validate actual deployment configurations against the pinned documentation for your Meilisearch release. [S11]

---

## 16. Bandwidth optimization

| Source of pressure | Required control |
|---|---|
| Repeated resume upload | Upload once, record content hash, reuse owner-scoped document ID; do not create cross-user existence oracle। |
| Inter-service file transfers | Utilize direct signed storage upload/download URLs where possible; enforce streamed chunking and bounded temporary file storage. |
| Oversized AI prompts | Clean text, selected facts, small evidence snippets; no raw HTML/base64 without task need। |
| Document/image vision | Only necessary pages/crops; explicit opt-in; text extraction first। |
| Browser loading | Only needed page/resources; unnecessary video/ads/preloads off where task permits। |
| Crawling | Scheduled incremental fetch, conditional requests when supported, domain budgets, bounded page counts। |
| Dashboard data | Pagination, field selection, lazy-loading heavy tables/charts। |
| Progress events | One authorized SSE connection per active client view; coalesce UI deltas। |
| Reconnect | Resume by run/event ID, never start another generation। |
| Repeated exports | CSV streamed; Sheets changed rows only; no whole-sheet rewrite on every edit। |
| Media previews | Thumbnails first, no autoplay, full-resolution asset only on demand। |

### 16.1 Initial byte limits

| Item | Proposed starting limit |
|---|---:|
| Ordinary AI JSON request body | 128 KiB uncompressed; task tokenizer check additionally required |
| Ordinary AI text/JSON response | 256 KiB after decompression |
| Remote HTML/text page | 2 MiB after decompression |
| Normal UI API response | 100 KiB target; paginate larger datasets |
| Queue message envelope | 4 KiB maximum |
| Resume upload | 10 MiB; magic-byte/type validation |
| DOCX expanded archive data | 50 MiB maximum; reject zip bombs/external fetches |
| Thumbnail | 250 KiB target |
| Research evidence | Up to 5 selected pages in initial Economy run |

Never silently truncate large, legitimate input documents; route them to a dedicated large-document pipeline with explicit cost estimates. Token and byte budgets are separate constraints; small archive sizes do not guarantee safe decompressed payloads. Enforce document parser sandboxing, page and processing time limits, and image pixel dimension constraints for OCR.

Trigger OCR only when deterministic text extraction yields insufficient content, processing only necessary pages (initial batch of 5 pages, requiring user confirmation for further processing). Provide a manual text entry fallback if the user opts out of OCR.

### 16.2 Progress streaming

Server-Sent Events (SSE) provide an interactive typing experience, but streaming does not reduce generated token billing. Provider streaming returns partial generation tokens; the gateway must handle client reconnection and cancellation mechanics reliably. [S12]

Recommended client coalescing: 500 ms debounce window, bounded buffer, and 30-second keep-alive heartbeats. For slow clients, prevent buffers from growing unboundedly; terminate lagging connections and resume from persisted states. Background browser tabs must suspend routine polling. Never persist raw streamed text as an approved artifact until final schema validation completes.

### 16.3 Bandwidth visibility

Track network metrics across distinct dimensions: provider request/response bytes, remote scraping bytes, browser worker transfers, user uploads/downloads, and media payload bytes. Application payload bytes differ from cloud provider egress bills due to TLS framing, proxy headers, CDN overhead, and internal routing. Display `unknown` rather than misleading zero values if measurements are unobserved.

---

## 17. Browser, OCR, and media workers

These components are not autonomous AI agents; they are isolated execution tools. The AI gateway dispatches tasks to them using typed contracts only when explicitly necessary. Account permissions and browser security policies are maintained without alteration.

Browser navigation is not a default pipeline: prioritize official authorized APIs or imported datasets whenever available. Sensitive screenshots from authentication pages, private profiles, or active sessions must never be written to general logs. Continuous live video feeds are disabled by default, requiring explicit interactive user sessions with time-bounded leases.

Media pipelines begin by generating text briefs or scripts. Dispatches to media providers occur only after the user approves explicit cost estimates and resource reservations. Successfully sending a prompt does not constitute a completed media asset. Audio and video generation must not draw from text token quotas; track media units transparently as duration seconds, image counts, or rendering minutes.

Local model deployment is an advanced, self-hosted operational configuration. Constrain local execution to a single loaded model, single concurrency, bounded context windows, and automated idle unloading; track model download size and runtime memory footprints separately. Never automatically download large model weights onto primary web/API servers. Ensure manual workflows remain functional if endpoints are unreachable.

---

## 18. Prompt, tool, and factual safety

Prompts must be maintained as versioned source files; never inject large raw repository contexts or full blueprints into runtime prompts. Ingest only the specific task instructions, schemas, and verified reference facts required for execution.

Candidate resumes, scraped job postings, job descriptions, and messages are **untrusted data**, not system instructions. Embedded directives within these inputs must never alter user permissions, model routing, credential hosts, database queries, file paths, or message recipients. All available tools must be strictly typed and validated server-side.

AI models must never receive direct SQL credentials, shell execution access, or database administrative privileges. The gateway must never dynamically auto-install third-party APIs or tool definitions from the internet. Internal function endpoints re-authorize ownership and tenant scope independently. Tool arguments specifying URLs or data sources must strictly derive from allowlisted task contexts.

Output validators independently verify: schema validity, field completeness, source traceability, factual consistency, and editorial style. Missing facts trigger clarifying questions; truncated completions are marked incomplete; model refusals are reported transparently. Short reasoning summaries may be requested, but unbounded hidden reasoning traces must not be mandated in prompts or outputs.

---

## 19. Data model extensions

Reuse existing relational credential, workflow, and usage schemas; do not create duplicate or overlapping ledger tables. The table definitions below illustrate target migration schemas rather than already deployed structures.

| Entity | Purpose / key fields |
|---|---|
| `ai_connections` | owner scope, approved origin, adapter, credential reference, privacy policy, status, last verification |
| `ai_models` | exact model ID, connection, capabilities, context/output limits, metadata revision |
| `ai_price_versions` | provider/model/currency/billing line items, source, verified_at, effective_at |
| `ai_agent_definitions` | task, prompt/schema/validator version, input scope, tool allowlist |
| `ai_routes` | task/mode to model mapping, payer, bounded fallback, policy version |
| `ai_budget_policies` | scope, token/money/request/byte limits, reset windows, warning thresholds |
| Existing `action_runs` / attempts | AI-specific metadata, input snapshot refs, idempotency key, provider request ID |
| Existing reservations / usage events | attempt-based reserve/settle, raw normalized usage, monetary snapshot, reconciliation state |
| `ai_output_cache` | exact key, owner scope, output reference, invalidation versions, expiry |
| `ai_evaluation_runs` | synthetic cases, model/prompt revision, expected checks, measured results |
| `ai_resource_events` | limited metrics for queue, RAM, network, execution time and backpressure |

Store secrets in an encrypted credential vault and media artifacts in object storage. Avoid transforming arbitrary agent labels, profile attributes, or package names into high-cardinality telemetry tags. Large execution logs must be paginated and subject to retention pruning policies.

### 19.1 Provider adapter contract — Go design sketch

```go
// Illustrative interface: concrete types and provider SDK mappings are implementation tasks.
type ProviderAdapter interface {
    ValidateConfig(ctx context.Context, cfg ConnectionConfig) error
    Capabilities(ctx context.Context, model ModelRef) (Capabilities, error)
    Estimate(ctx context.Context, req GenerationRequest) (UsageEstimate, error)
    Generate(ctx context.Context, req GenerationRequest, sink EventSink) (GenerationResult, error)
}
```

Capabilities like `ListModels`, token counting endpoints, batch execution, and cancellation are not universally available across all providers; model them as optional capability interfaces. The core `Generate` method enforces budget bounds and accurately reports partial usage. HTTP transport configurations, authentication injection, and SSRF mitigations remain shared.

---

## 20. API contracts and UI behavior

| Method / route | Purpose |
|---|---|
| `GET /api/v1/ai/connections` | Authorized connection list; no raw secrets |
| `POST /api/v1/ai/connections` | Add a connection with write-only credential |
| `PATCH /api/v1/ai/connections/{id}` | Versioned edit; endpoint change triggers verification |
| `POST /api/v1/ai/connections/{id}/test` | Explicit budgeted synthetic test |
| `POST /api/v1/ai/connections/{id}/rotate` | Replace credential safely |
| `POST /api/v1/ai/connections/{id}/disable` | Stop new/queued dispatch |
| `GET /api/v1/ai/models` | Scoped catalog and verified capability status |
| `PUT /api/v1/ai/routes/{task}` | Task/mode/model policy |
| `GET /api/v1/ai/budgets` | Consumed, reserved, remaining by permitted scope |
| `PUT /api/v1/ai/budgets/{scope}` | Authorized policy change within ceilings |
| `POST /api/v1/ai/estimate` | Local/cached preflight estimate; not an LLM generation |
| `POST /api/v1/ai/runs` | Submit task intent with Idempotency-Key |
| `GET /api/v1/ai/runs/{id}` | Durable status / result reference |
| `GET /api/v1/ai/runs/{id}/events` | Scoped SSE progress |
| `POST /api/v1/ai/runs/{id}/cancel` | Stop future work; best-effort provider cancel |
| `GET /api/v1/ai/usage` | Scoped aggregated usage with estimates distinguished |

If `estimate` queries a provider token-counting endpoint, it is not a generation request but counts against network and rate limits. If no such endpoint exists, apply a conservative local estimate. The gateway does not expose an arbitrary prompt proxy; clients send `task`, `resource_ids`, and validated parameters, and the Go backend hydrates authorized data.

**Run status values:** `queued`, `running`, `needs_input`, `needs_review`, `budget_blocked`, `resource_wait`, `failed`, `cancel_requested`, `cancelled`, and `completed`. Provider usage reconciliation status is tracked in a separate field, allowing completed artifacts to remain pending reconciliation.

---

## 21. Proposed configuration example

This YAML illustrates the intended configuration schema; the parser and validation routines are specifications. Production monetary thresholds must not be activated without operator confirmation. Secret credentials must never be stored in this file.

```yaml
ai:
  enabled: true
  default_mode: economy
  managed_spending_enabled: false
  local_inference_enabled: false
  autonomous_orchestrator_enabled: false

  concurrency:
    per_user: 1
    per_installation: 2
    document_jobs: 1
    browser_sessions: 0
    media_jobs: 0

  requests:
    max_body_kib: 128
    max_response_kib: 256
    max_attempts_standard: 2
    max_attempts_research: 3
    timeout_seconds: 60
    retry_on_ambiguous_dispatch: false
    automatic_fallback_enabled: false
    automatic_paid_upgrade: false

  budgets:
    require_verified_price_for_managed_routes: true
    user_daily_usd: null
    user_monthly_usd: null
    installation_daily_usd: null
    warning_percent: 80
    urgent_warning_percent: 95
    stop_new_requests_percent: 100
    keep_unknown_usage_reserved: true

  context:
    normal_input_tokens: 4000
    normal_generation_tokens: 1000
    recent_chat_messages: 6
    summary_tokens: 600
    retrieval_top_k: 5
    send_raw_documents_by_default: false
    embeddings_enabled: false

  cache:
    exact_results_enabled: true
    enforce_owner_and_version_scope: true
    provider_prompt_cache: capability_gated
    result_pointer_ttl_hours: 24
    in_process_max_mib: 8
    payloads_in_redis: false

  resources:
    heavy_job_slots: 1
    memory_warning_percent: 75
    memory_critical_percent: 85
    max_research_pages_per_run: 5
    max_remote_page_mib: 2
    queue_envelope_max_kib: 4
    idle_browser_shutdown_seconds: 60

  ui:
    stream_coalesce_ms: 500
    sse_keepalive_seconds: 30
    autoplay_media: false
    ai_autorun_on_form_edit: false

  security:
    public_cloud_custom_origins_require_approval: true
    follow_redirects: false
    private_network_targets_default_allowed: false
    store_raw_prompts_in_logs: false
    cross_tenant_cache: false
```

`null` indicates that no production monetary policy has been configured—it does not imply unlimited spend. Platform-managed routes in this state are blocked. Operating personal BYOK in token-only tracking mode requires explicit user consent and separate policies. Per-task limits resolve against Section 9 routing rules; global defaults must never override task-specific safety caps.

---

## 22. Dashboard metrics and tooltips

### User view

Active provider/model, billing entity, run estimates, input/output token usage, remaining allowance, cache hit indicators, artifact version, queue state, and abort controls. When BYOK credentials are used, display a "Provider-billed usage estimate" disclaimer.

### Admin view

Workspace expenditure and reservations, allocation quotas, shared provider health status, queue depth, approved models, and task-level aggregated outcomes. Candidate resumes and private prompts remain hidden by default.

### Super Admin view

Platform-wide aggregate usage, categorized connection errors, hardware resource saturation, price-card validity ages, pending reconciliation counts, adapter verification results, and emergency kill switches. Do not log raw prompt contents for routine monitoring.

### Useful tooltips

> **Economy mode:** "Prioritizes deterministic rules and cached outputs first. When AI is necessary, routes to the most cost-effective tested model for this specific task."

> **Token limit:** "Inputs, system instructions, schemas, and model generation limits are accounted for separately. Short completions may still consume reasoning token budgets."

> **Cache reused:** "Served existing verified results for this input version; no new model generation call was dispatched."

> **API spend:** "Represents usage tracked through this application only. If this credential is used across other platforms, provider invoices will differ."

> **Low-RAM mode:** "Restricts concurrent task executions. Non-urgent background jobs wait in queue to prevent unnecessary local model loading or headless browser instances."

> **Cancel:** "Halts subsequent execution stages. If the provider has already processed the request upstream, usage charges may still apply."

> **Custom API:** "Can be used if the API format, endpoint, model ID, and security parameters are fully compatible. Valid credentials do not guarantee all platform capabilities are supported."

---

## 23. Implementation task list

All listed software tasks are **TODO**. Integrate them into existing `TASKS.md` under the `AIARC-*` prefix; do not renumber existing `FND-*` task IDs.

| Task | Work | Dependencies | Acceptance evidence |
|---|---|---|---|
| AIARC-001 | AI connections/model/budget schemas + API contracts | FND-002, FND-005, FND-009 | Migration, ownership tests, OpenAPI validation |
| AIARC-002 | Approved endpoint transport + vault wiring | AIARC-001 | SSRF, TLS, redirect, key redaction/rotation tests |
| AIARC-003 | Provider add/test/manage dashboard | AIARC-002, FND-014 | Personal/workspace/platform flows; masked keys |
| AIARC-004 | OpenAI-compatible adapter | AIARC-002 | Mock stream, non-stream, usage, capability contract tests |
| AIARC-005 | Native OpenAI/Anthropic/Gemini adapters | AIARC-004 | Each adapter's typed auth/usage/schema behavior; no assumed compatibility |
| AIARC-006 | Task registry, prompt/schema versions, model routes | AIARC-001, AIARC-004 | Valid route manifests; unsupported capabilities rejected |
| AIARC-007 | Atomic monetary/token reservations + usage settlement | AIARC-001, FND-007, FND-011 | Concurrent budget tests; no zero-cost unknown outcomes |
| AIARC-008 | Owner/version-scoped cache + single-flight | AIARC-006, AIARC-007 | Duplicate request yields one provider dispatch |
| AIARC-009 | Bounded context builder + source/fact validators | AIARC-006, FND-012 | Token/byte/coverage and factual fixtures |
| AIARC-010 | Durable AI workers, cancellation, SSE and recovery | AIARC-007, AIARC-008, FND-008 | Crash/reconnect/timeout/redelivery tests |
| AIARC-011 | Profile/Career/Application task implementations | AIARC-009, AIARC-010 | Confirmed-fact/no-invention/resume review tests |
| AIARC-012 | LinkedIn/Content/Research task implementations | AIARC-009, AIARC-010 | Grounded drafts/cited research; no direct external writes |
| AIARC-013 | Quota/usage/resource dashboards + tooltips | AIARC-010 | Visibility boundaries; estimated/reported separation |
| AIARC-014 | CPU/RAM/queue/byte backpressure controls | AIARC-010 | Controlled saturation tests and graceful pause |
| AIARC-015 | Optional local inference/private endpoint support | AIARC-002, AIARC-014 | Operator allowlist, isolated resource admission |
| AIARC-016 | Optional media/browser adapter gating | AIARC-007, AIARC-014 | Separate billing/resource reservations; no default launch |
| AIARC-017 | Security, correctness and cost regression suite | AIARC-011, AIARC-012, AIARC-013, AIARC-014 | AI-AT checks below; actual reports |
| AIARC-018 | Blueprint integration and documented handoff | AIARC-001 | Add links/requirements/tasks; PROGRESS reflects actual work only |

**Initial AI vertical slice:** Configure one compatible provider → verify via synthetic test probe → confirm candidate profile manually → generate one budgeted resume tailoring suggestion → review output → track usage → verify identical repeat request serves cached result. Live LinkedIn interactions are not required.

---

## 24. Acceptance tests

| ID | Required test |
|---|---|
| AI-AT-001 | Personal credentials are never exposed in other users' or admin API responses, application logs, HTML markups, or export files. |
| AI-AT-002 | Unapproved origins, HTTP redirects, cloud metadata IP destinations, and DNS rebinding requests are strictly blocked. |
| AI-AT-003 | Updating an endpoint base URL never forwards previously stored credentials to unverified hosts. |
| AI-AT-004 | Ten concurrent identical submissions produce one logical run and at most one initial dispatch। |
| AI-AT-005 | Different owners/profile versions cannot reuse private cached output। |
| AI-AT-006 | If a required budget reservation exceeds the remaining allowance, external provider calls are rejected immediately. |
| AI-AT-007 | Concurrent requests that collectively exceed payer spending caps are blocked from admission. |
| AI-AT-008 | Token accounting for reasoning and cached inputs does not double-count nested total usage figures. |
| AI-AT-009 | Network timeouts occurring after dispatch mark usage as pending reconciliation; blind retries or immediate refunds are prohibited. |
| AI-AT-010 | Retries, schema repair passes, and fallback invocations consume from the same unified run attempt and cost budgets. |
| AI-AT-011 | Incompatibilities in provider capabilities return explicit, descriptive errors rather than silently ignoring safety caps. |
| AI-AT-012 | Mixed Bengali/English long inputs adhere to token and byte boundaries; partial extractions are flagged honestly. |
| AI-AT-013 | Model-generated hallucinations or unknown candidate facts are never recorded as confirmed profile records. |
| AI-AT-014 | AI-generated flags or parameters cannot bypass human approval gates to execute messages, job applications, or social posts. |
| AI-AT-015 | Page refreshes, SSE reconnects, or opening duplicate browser tabs do not trigger duplicate paid generation cycles. |
| AI-AT-016 | Lagging SSE consumers maintain bounded server memory buffers and disconnect safely with recoverable state markers. |
| AI-AT-017 | Redis memory exhaustion or restart does not corrupt persistent budgets or runs; recovery routines never duplicate paid requests. |
| AI-AT-018 | Headless browser, OCR, and media workers are disabled by default and require explicit resource admission; models never auto-download. |
| AI-AT-019 | Under high memory pressure, heavy background queues are paused while standard form editing and tracking remain operational. |
| AI-AT-020 | Byte, archive, and page length restrictions reject oversized or malicious files, keeping parser memory bounded. |
| AI-AT-021 | Revoking credentials or deleting user data promptly purges pending queued jobs and associated cache entries. |
| AI-AT-022 | Absence of AI credentials or exhausted budgets does not prevent manual profile edits, CSV imports, or deterministic resume exports. |
| AI-AT-023 | Model routing benchmarks evaluate factual correctness, schema compliance, and multilingual quality, not merely token unit prices. |
| AI-AT-024 | Advanced media pipelines require explicit human confirmation displaying estimated costs, payer identity, and resource allocations. |

**Metrics required in load-test reports:** p50 and p95 latency, peak/RSS memory per process, CPU utilization, queue dwell time, provider calls per completed task, input/output tokens, exact-match cache hit rates, observed network bytes, timeout and repair rates, and unresolved billing counts.

Test fixtures must use synthetic data; real resumes and credentials must never enter version control. Automated unit, contract, and load test suites must be verified through actual test runner execution rather than assumed.

---

## 25. Integration with existing blueprints and developer handoff

| Existing file | Required update during integration |
|---|---|
| `README.md` | Reference this document, document AI provider setup workflows, and clarify that it extends the core documentation suite. |
| `CONTEXT.md` | REQ-AI-001–REQ-AI-012 append; dashboard BYOK scopes, Economy default, no direct AI side effects। |
| `PLAN.md` | Gateway inside Go modular monolith, capability adapters, durable spend reservations, bounded resources add। |
| `TASKS.md` | AIARC-001–AIARC-018 append; existing IDs/dependencies preserve; AI-AT mapping include। |
| `PROGRESS.md` | Distinguish architectural specification from code implementation; record actual task execution and test evidence. |

### Architecture decisions to record

| ID | Decision |
|---|---|
| ADR-AI-001 | One shared Go gateway, six logical task agents; no mandatory multi-agent runtime। |
| ADR-AI-002 | Dashboard-configurable providers through reviewed typed adapters, not arbitrary executable integrations। |
| ADR-AI-003 | Deterministic-first routing, exact caching, task budgets and bounded retries। |
| ADR-AI-004 | PostgreSQL reservation/usage truth; Redis transport recoverable। |
| ADR-AI-005 | Hosted API default; local models/browser/media opt-in and resource-gated। |
| ADR-AI-006 | Provider/credential/data-sharing permissions remain independent from AI generation and package entitlement। |

### AI coding-agent instructions

```text
Read README.md, CONTEXT.md, PLAN.md, TASKS.md, PROGRESS.md,
and AI-AGENT-ARCHITECTURE.md. Inspect the actual code and git state.

This is a specification, not proof that any component exists.
Respect the fixed stack and existing ownership/approval rules.
Start with the next unblocked task; AI tasks depend on foundation work.
Implement one bounded feature/vertical slice and record actual evidence.

Do not add a new agent framework or separate per-agent server by default.
Do not hardcode commercial model IDs/prices as universal assumptions.
Do not expose provider keys or run arbitrary user-supplied API scripts.
Do not automatically send production resumes/messages to test providers.
Count every model attempt, cache write, paid tool and media operation.
Do not mark a cancelled/unknown provider request as unbilled.
Keep regular forms/tracking/export available without AI.

Update TASKS.md and PROGRESS.md with changed files, tests actually run,
remaining risks, resource measurements and the next exact task ID.
```

---

## 26. Sources and verification boundaries

Prepared against the supplied five-file blueprint and the user's current requirements. Official references below were checked on **2026-09-17**. Provider names, interfaces, pricing and limits must be rechecked when implementing/pinning actual versions. No commercial pricing figures, token savings percentage, RAM benchmark or live API success is asserted here.

[S1]: https://developers.openai.com/api/docs/guides/prompt-caching
[S2]: https://developers.openai.com/api/docs/guides/reasoning
[S3]: https://platform.claude.com/docs/en/build-with-claude/prompt-caching
[S4]: https://ai.google.dev/gemini-api/docs/openai
[S5]: https://docs.ollama.com/api/openai-compatibility
[S6]: https://docs.ollama.com/faq
[S7]: https://go.dev/doc/gc-guide
[S8]: https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html
[S9]: https://developers.openai.com/api/docs/guides/structured-outputs
[S10]: https://redis.io/docs/latest/develop/reference/eviction/
[S11]: https://www.meilisearch.com/docs/resources/self_hosting/configuration/reference
[S12]: https://developers.openai.com/api/docs/guides/streaming-responses
[S13]: https://ai.google.dev/gemini-api/docs/caching

| Reference | What it supports |
|---|---|
| [S1 — OpenAI prompt caching][S1] | Prefix/breakpoint and model-specific cache rules; cache savings not unconditional। |
| [S2 — OpenAI reasoning][S2] | Reasoning usage contributes to billable output/context; generation cap mapping matters। |
| [S3 — Anthropic prompt caching][S3] | Stable-prefix cache boundaries and cache write/read accounting। |
| [S4 — Gemini compatibility][S4] | Documented compatible interface; selected endpoint/model still needs tests। |
| [S5 — Ollama compatibility][S5] | Subset compatibility rather than universal OpenAI feature equivalence। |
| [S6 — Ollama FAQ][S6] | Local-model concurrency/context impacts memory। |
| [S7 — Go GC guide][S7] | Runtime memory tuning and soft-limit caveats। |
| [S8 — OWASP SSRF guidance][S8] | Custom endpoint validation plus network-layer protections। |
| [S9 — Structured outputs][S9] | Schema output handling must include refusal and incomplete-response branches। |
| [S10 — Redis eviction][S10] | Noeviction rejects new writes under pressure; cache policy must fit control data। |
| [S11 — Meilisearch configuration][S11] | Explicit indexing configuration and resource controls। |
| [S12 — OpenAI streaming][S12] | Incremental API response delivery; not a promise of lower generation usage। |
| [S13 — Gemini caching][S13] | Cache behavior depends on model and API path। |

**Completion state:** Architectural specification delivered. Software implementation, live provider integrations, billing charges, application automated tests, and production deployment are not performed by this specification document.
