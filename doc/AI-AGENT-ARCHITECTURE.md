# AI Agent Architecture
## Dashboard-configurable AI • Token budgets • Low-RAM / low-bandwidth operation

Before starting or continuing any work, read README.md, CONTEXT.md, PLAN.md, AI-AGENT-ARCHITECTURE.md, TASKS.md, and PROGRESS.md in full and treat them as one interconnected source of truth. Cross-reference requirements, architecture decisions, AI-agent rules, task dependencies, implementation status, and prior decisions across all six files before making changes. Do not interpret any file in isolation, silently override an existing decision, duplicate functionality, change the approved stack, or mark work complete without implementation and verification evidence. After each meaningful implementation step, update TASKS.md and PROGRESS.md so the next AI/coding session can resume accurately without losing context.

**Product:** Career / LinkedIn / Social Platform  
**Stack:** Next.js + TypeScript + Tailwind + Go + PostgreSQL + Redis + Meilisearch  
**Document version:** 1.0 • **Prepared:** 2026-09-17  
**Status:** Implementation specification — এই document application code বা deployment নয়।

> **মূল সিদ্ধান্ত:** একটি shared Go AI Gateway থাকবে। Dashboard থেকে provider connection, API credential, model এবং task-wise budget configure করা যাবে। সাধারণ কাজ deterministic services করবে; reasoning বা writing দরকার হলেই ছোট, bounded AI request যাবে। ছয়টি logical agent মানে ছয়টি server, ছয়টি model বা ছয়টি simultaneous call নয়।

এই file আগের `README.md`, `CONTEXT.md`, `PLAN.md`, `TASKS.md` এবং `PROGRESS.md`-এর extension। আগের পাঁচটি file এই deliverable-এ পরিবর্তন করা হয়নি। নিচের task/decision IDs সেগুলোতে merge করার নির্দেশনা section 25-এ আছে। মূল privacy, provider permission, approval এবং confirmed Applied/Published status-এর নিয়ম অপরিবর্তিত থাকবে।

**প্রস্তাবিত সংখ্যাগুলো configuration defaults/testing targets; measured performance, provider price, official platform limit বা production capacity guarantee নয়।**

---

## 1. Product requirements ও non-negotiable rules

| ID | Requirement |
|---|---|
| REQ-AI-001 | Dashboard থেকে approved AI provider connection add, test, edit, disable, rotate ও revoke করা যাবে। |
| REQ-AI-002 | Native adapter এবং OpenAI-compatible adapter থাকবে; unsupported API-র জন্য নতুন adapter লাগবে। |
| REQ-AI-003 | User-private, workspace-shared এবং platform-managed credentials/billing আলাদা থাকবে। |
| REQ-AI-004 | Task অনুযায়ী model নির্বাচন; ছোট কাজের জন্য অকারণে expensive model নয়। |
| REQ-AI-005 | প্রতিটি provider call-এর আগে input/output/cost/attempt budget পরীক্ষা ও reserve করতে হবে। |
| REQ-AI-006 | একই request/cache hit-এর জন্য নতুন model call নয়; regenerate explicit action। |
| REQ-AI-007 | Context ছোট হবে, কিন্তু facts বা critical eligibility information silently বাদ দেওয়া যাবে না। |
| REQ-AI-008 | Local inference, browser, OCR ও media workers resource gate ছাড়া start হবে না। |
| REQ-AI-009 | Bounded concurrency, streamed I/O, small queue messages এবং incremental sync থাকবে। |
| REQ-AI-010 | AI-এর budget বা API failure non-AI editing, viewing, tracking ও basic export বন্ধ করবে না। |
| REQ-AI-011 | AI suggestions নিজেরা external actions, permissions বা profile facts approve করতে পারবে না। |
| REQ-AI-012 | Usage, approximate cost, retries, queue delay, RAM ও bandwidth dashboard-এ explainable হবে। |

**V1-এ প্রয়োজন নেই:** mandatory LangGraph, CrewAI, AutoGen, Temporal, Python AI service, vector database, per-agent container বা always-running autonomous loop। এগুলো পরবর্তীতে evidence-backed architecture decision ছাড়া যোগ করা যাবে না।

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

**Durability:** PostgreSQL owns runs, input versions, credential metadata, usage, budgets, approvals এবং outcomes। Redis Streams transport ও bounded coordination করবে। Meilisearch authorized retrieval-এর projection; এটি AI memory বা permission authority নয়। Private object storage-এ documents/media থাকবে। এই boundaries আগের `PLAN.md`-এর ADR-001–ADR-010 অনুসরণ করে।

**Deployment:** প্রথমে existing `services/core/cmd/api` এবং `cmd/worker`-এর মধ্যেই gateway package রাখবে। Load data প্রয়োজন প্রমাণ না করা পর্যন্ত আলাদা AI microservice নয়।

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
| Display name | যেমন `My writing provider`; secret নয়। |
| Adapter type | `openai_responses`, `openai_chat_compatible`, `anthropic_native`, `gemini_native`, approved optional adapter। |
| Base URL / endpoint | Approved origin; public cloud-এ new custom origin approval চাইবে। |
| API credential | Write-only form; saving-এর পরে masked identifier। Raw key পুনরায় reveal নয়। |
| Model ID / deployment ID | Provider-এর exact identifier; manual entry সম্ভব। `/models` সব provider-এ আছে ধরে নেওয়া যাবে না। |
| API version | প্রয়োজনীয় adapters-এ explicit, versioned setting। |
| Ownership / allowed users | Personal, workspace বা platform; permission server-side। |
| Allowed data classes | Public content / private career / permitted messaging context / media। |
| Data residency / retention policy | Known provider policy reference, review date; unknown হলে স্পষ্ট label। |
| Capability status | Verified / Declared / Unsupported / Not tested। |
| Price card | Currency, input/output/cache/media rate, source, verified date ও billing semantics। |
| Budget | Per request, day, billing month, run এবং credential pool। |
| Timeout / rate limits | Provider-aware settings; admin ceiling অতিক্রম নয়। |
| Fallback | Default off; explicit approved model/connection এবং max cost। |
| Enabled agents | Profile, Career, Application, LinkedIn, Content, Research। |

**Test connection:** endpoint validation → authentication/capability probe → optional tiny synthetic generation। Real resume, private conversation বা complete prompt history test-এ পাঠাবে না। Generation test সর্বোচ্চ একটি request; উদাহরণ budget 256 input + 64 billable generation tokens। Model ওই cap support না করলে test চালানোর আগে নতুন estimate/approval চাইবে। Test call-ও metered; dashboard reload-এ repeated test নয়।

**Tooltip examples:**

> “এই key শুধু server-side ব্যবহার হবে। এটি অন্য user বা AI prompt-এ প্রকাশ করা হবে না।”

> “Connection test provider usage তৈরি করতে পারে। এখানে কোনো personal resume পাঠানো হবে না।”

> “Model list পাওয়া না গেলে provider dashboard থেকে exact model ID দিন। Save করা এবং capability verified হওয়া এক বিষয় নয়।”

### 3.3 Dashboard tabs

| Tab | কী থাকবে |
|---|---|
| Providers | Connections, model access, credential rotation, last successful request, known errors। |
| Agent routing | কোন task কোন model, Economy/Balanced/Advanced mode এবং fallback policy। |
| Budgets | Used / reserved / available; user, workspace ও payer-specific view। |
| Usage | Token categories, estimated cost, cache hits, successful/failed attempts। |
| Runs | Queued, running, waiting for input, budget-blocked, cancelled, usage-pending। |
| Resource limits | Concurrent jobs, browser/media permissions, queue depth ও RAM thresholds। |
| Privacy | Allowed data sharing, consent, retention, cache purge ও disconnect। |

---

## 4. “যেকোনো API” support-এর সঠিক অর্থ

**Provider-agnostic architecture হবে; arbitrary API universalভাবে plug-and-play হবে না।** Request/response protocol, authentication ও model capability compatible হতে হবে।

| Provider family | Implementation path | Caveat |
|---|---|---|
| OpenAI | Native Responses adapter; separately tested Chat Completions route প্রয়োজন হলে | Endpoint-specific tool/usage/output mappings। |
| Anthropic | Native messages adapter | নিজের auth headers, content blocks, stop/usage/cache mapping। |
| Google Gemini | Native generation adapter অথবা verified OpenAI-compatible endpoint | নির্বাচিত endpoint-এর capability আলাদা হতে পারে। |
| OpenAI-compatible hosted services | Generic chat adapter + approved endpoint/model manifest | OpenRouter/Groq/DeepSeek-জাতীয় service candidate; প্রতিটি বাস্তব endpoint/model আলাদাভাবে test করতে হবে। |
| Ollama / private model server | Tested compatible adapter, explicit private-network configuration | Local model loading application server-এ default off। |
| Azure / Bedrock / অন্য enterprise API | Dedicated adapter | Deployment names, signed requests বা workload identity থাকলে শুধু API key যথেষ্ট নয়। |
| Image / audio / video APIs | পৃথক typed media adapter | Text model credential দিয়ে সব media feature পাওয়া যাবে ধরে নেওয়া যাবে না। |
| Unknown vendor | New reviewed adapter বা compatible protocol test | User-provided executable code, arbitrary curl বা JavaScript mapping চালানো যাবে না। |

Google একটি OpenAI compatibility interface document করে; Ollama স্পষ্টভাবে OpenAI API-র subset support বলে। তাই `supports_streaming`, `supports_json_schema`, `supports_tools`, `supports_vision`, `reports_usage` ইত্যাদি provider-এর নাম দেখে নয়, selected endpoint/model ধরে verify করবে। [S4], [S5]

### 4.1 Capability manifest

প্রতি model/version-এ রাখবে:

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

Unsupported settings silently drop করা যাবে না। উদাহরণ: selected model strict schema support না করলে valid fallback validator route ব্যবহার করবে অথবা task unavailable দেখাবে। Mandatory budget cap honor না করলে model-টিকে সেই automated task-এ enable করবে না।

---

## 5. Roles, credentials ও billing ownership

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

**BYOK = Bring Your Own Key।** তিনটি billing mode থাকবে: `platform_managed`, `workspace_byok`, `personal_byok`। UI-তে payer আগে দেখাবে। Personal key fail হলে silently platform key বা অন্য member-এর key ব্যবহার নয়।

App কেবল নিজের মাধ্যমে হওয়া usage measure করবে। একই API key অন্য app-এ ব্যবহার হলে তার billing এই dashboard-এ automatically জানা যাবে না। BYOK provider bill এবং আমাদের hosting/subscription fee আলাদাভাবে দেখাবে।

### 5.1 Secret handling

API key TLS দিয়ে Go backend-এ যাবে; envelope encryption-এ existing credential vault-এ থাকবে। Encryption root/master key database ও public repo-র বাইরে রাখতে হবে। Logs, traces, Redis messages, support export, analytics বা prompt-এ secret যাবে না। `NEXT_PUBLIC_*`, localStorage বা browser-to-provider requests-এ provider key রাখা যাবে না।

Base URL, owner scope বা provider adapter বদলালে connection revalidation বাধ্যতামূলক। নতুন origin-এ পুরোনো credential automatically forward নয়। Key revoke হলে queued এবং নতুন attempts dispatch বন্ধ; already-sent request-এর outcome আলাদা reconcile করতে হবে।

### 5.2 Custom endpoint নিরাপত্তা

Public cloud-এ arbitrary endpoint accept করে server-কে open proxy বানাবে না। Approved HTTPS origin allowlist, port/path constraints, certificate verification ও controlled egress থাকবে। Redirect default off। Connect করার সময় DNS/IP validate করবে; loopback/private/link-local/cloud-metadata destinations, IPv4/IPv6 variants এবং DNS rebinding ঠেকাতে application validation-এর সঙ্গে network policy থাকবে। [S8]

Self-hosted local Ollama endpoint-এর জন্য **operator-controlled exact network exception** দেওয়া যাবে। Public cloud tenant নিজের form দিয়ে localhost বা internal database subnet খুলতে পারবে না। Cloud থেকে user's local machine access চাইলে separately authenticated private connector প্রয়োজন; শুধু `localhost` URL যথেষ্ট নয়।

---

## 6. ছয়টি logical agent

| Agent | দায়িত্ব | AI output | যা সরাসরি করতে পারবে না |
|---|---|---|---|
| Profile Agent | Extracted resume text থেকে structured candidate profile; ambiguity/missing facts | Candidate facts + source spans + questions | Confirmed profile silently overwrite |
| Career Agent | Match explanation, resume tailoring, cover letter, interview preparation | Fact-grounded suggestions / versioned draft | Qualification বা achievement বানানো |
| Application Agent | Unknown free-text form question-এর answer draft | Answer + supporting fact IDs + unresolved fields | Eligibility guess বা submit |
| LinkedIn Agent | Headline/About/experience optimization, professional notes/comments/replies | Before/after diff বা draft | Account connect, send বা profile update অনুমতি দেওয়া |
| Content Agent | Voice-aware posts, hooks, repurposing, scripts ও media brief | Versioned text/asset plan | Unreviewed publish বা unlimited media generation |
| Research Agent | Approved evidence থেকে trends/competitor/comment synthesis | Cited report + coverage limitations | Autonomous unlimited crawl বা private identity inference |

Agent definition হলো `task key + prompt version + schema + data scope + model route + budgets + validators`। শুধু একটি headline rewrite-এর জন্য অন্য পাঁচটি agent call হবে না।

**Shared writing:** LinkedIn Agent এবং Content Agent একই text-generation gateway/prompt primitives reuse করবে; দুই module একই post-এর জন্য duplicate generation করবে না।

### 6.1 কোন কাজে AI call হবে না

Signup/login, roles/packages, manual profile forms, required-field checks, deterministic ATS layout, PDF/DOCX rendering, job dedupe, ordinary filters, confirmed Applied state, calendar, published-status reconciliation, CSV/Sheets export, quota checks, arithmetic analytics, notifications এবং credential health metadata-তে default **0 LLM calls**।

Resume parser প্রথমে text extract করবে। AI কেবল semantic mapping/ambiguous section processing-এ ব্যবহার হবে। Missing required field rule দিয়ে ধরা গেলে সেই জন্য নতুন model call নয়।

### 6.2 Orchestrator

V1-এ buttons/workflows task নির্ধারণ করবে; LLM router প্রয়োজন নেই। Optional chat orchestration পরে add করা যাবে, কিন্তু সর্বোচ্চ 4 planned steps, 3 provider attempts total এবং স্পষ্ট run budget থাকবে। Agent-to-agent recursive delegation বা endless critic/rewrite loop নয়।

Model proposed action দিলে Go policy তা independently যাচাই করবে। Model-এর `requires_approval: false` কোনো permission grant নয়।

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

প্রতি retry-তে নতুন `attempt_id`, কিন্তু একই `run_id`। Database transaction network response-এর জন্য open রাখা যাবে না। Output generate হওয়া, user approve করা এবং external action হওয়া আলাদা events।

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

এই shape design example; runnable API ইতোমধ্যে আছে এমন দাবি নয়।

---

## 8. Model routing: Economy default

| Mode | Policy |
|---|---|
| Economy — default | সংশ্লিষ্ট task-এর evaluation pass করা সবচেয়ে কম estimated-cost approved model; smallest sufficient context। |
| Balanced | User/workspace selected route; bounded one-step escalation allowed if explicitly configured। |
| Advanced | User explicitly requests larger context/stronger model; নতুন estimate ও approval প্রয়োজন। |
| Manual / AI paused | Profile editing, template-based resume/export ও existing records available; generation disabled। |

`economy_text`, `quality_text`, `vision_optional`, `media_optional` হবে configurable aliases, hardcoded commercial model IDs নয়। অর্থসাশ্রয়ী model relevant Bengali/English fixtures ও factual checks pass না করলে সেটিকে cheapest বলে production route করা যাবে না।

**Fallback defaults:** off। Enable করলে maximum one alternate attempt, একই task budget-এর ভেতরে। More expensive model, নতুন vendor, data residency change বা নতুন payer silently নির্বাচন নয়। Timeout-এর পরে first request potentially processed হলে automatic fallback বন্ধ; duplicate cost risk দেখাবে।

---

## 9. Token budget: task-wise hard caps

সব input cap-এর মধ্যে system prompt, tool/schema definitions, previous messages এবং retrieved context অন্তর্ভুক্ত। Generation cap-এ reasoning/hidden generation কীভাবে counted হয় adapter manifest তা নির্ধারণ করবে। ছোট output মানেই ছোট bill নয়—যেমন OpenAI reasoning tokens output billing-এর অংশ। [S2]

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

**Max attempts-এর মধ্যে সব repair, retry, fallback ও chunk calls counted।** এটি প্রতি stage-এর allowance নয়। Research-এর 3 attempts-এ preprocessing/synthesis থাকলে সেগুলোও counted। Long resume, large report বা video script এই caps-এ না ধরলে আলাদা expanded workflow estimate চাইবে; silently truncate করে “complete” result নয়।

Model-এর minimum generation/thinking requirement cap-এর চেয়ে বেশি হলে আরেকটি tested model route ব্যবহার অথবা budget approval চাইবে। Unsupported `temperature`, `reasoning_effort`, `max_tokens` বা provider-specific parameter সব API-তে একইভাবে পাঠানো যাবে না।

### 9.2 Reduce calls before reducing quality

প্রথমে cache/structured fields/rules ব্যবহার করবে। LLM self-reported confidence-কে সত্যতা verification ধরে নেওয়া যাবে না। Schema validation deterministic; known facts source IDs দিয়ে মিলবে। Invalid JSON-এ maximum one budgeted repair; missing career facts-এ repair নয়, user input।

Form edit-এ AI auto-run হবে না। Autosave debounce করা যাবে, কিন্তু saving এবং paid generation আলাদা user actions। “Regenerate” current result replace না করে নতুন version ও usage তৈরি করবে।

### 9.3 Job discovery example

একটি **illustrative workflow**, savings benchmark নয়:

```text
100 discovered jobs
  → normalize + dedupe + explicit hard filters in Go
  → deterministic preliminary ranking
  → user can inspect all jobs and scoring criteria
  → AI explains only the selected/top 10
  → resume + cover letter only for the 2 the user chooses
```

এতে unselected 90 job-এর জন্য AI explanation call হয় না। Resume/letter-ও 100টির জন্য তৈরি হয় না। Preliminary ranking explainable হবে এবং uncertain eligibility বা wording variation-এর কারণে relevant jobs চিরতরে লুকাবে না।

---

## 10. Monetary budgets, metering ও overspend prevention

### 10.1 Budget scopes

`request → run → user → workspace → credential pool/payer → platform` scopes-এ limit থাকবে। প্রতিটি scope-এ consumed এবং reserved আলাদা। Day/rolling window/billing month reset semantics explicit হবে। একই key/module অন্য route দিয়ে ব্যবহার করলেও relevant shared payer limit এড়াতে পারবে না।

**Suggested UX:** 80% warning, 95% prominent warning, 100% হলে নতুন billable call blocked। Notification-এর জন্য AI দরকার নেই। Production dollar amount এই document approve করে না; Super Admin approved price metadata ও package অনুযায়ী সেট করবে।

### 10.2 Before-call reservation

```text
available(scope) = limit(scope) - settled(scope) - unresolved_reserved(scope)

reserve_before_send = conservative billable input estimate
                    + permitted generation upper bound
                    + cache-write / storage upper bound when applicable
                    + explicitly bounded tool/media charges
```

All relevant reservations একটি short PostgreSQL transaction-এ atomic হবে। Redis coordination accelerator; durable budget truth নয়। Insufficient budget হলে কোনো provider call পাঠানো যাবে না। Retry-তেও নতুন reservation লাগবে।

Raw/provider usage categories normalize করে non-overlapping billing line items তৈরি করবে। Provider-এর `output_tokens`-এ reasoning included থাকলে আবার reasoning যোগ করে double charge নয়। Cache read/write-ও provider-specific semantics অনুযায়ী normal input থেকে আলাদা করবে। Money integer micro-units বা database decimal-এ; floating-point ledger নয়।

### 10.3 Accounting states

| State | Meaning |
|---|---|
| `estimated` | আগে থেকে estimated token/cost; final invoice নয়। |
| `reported` | Provider response-এর usage থেকে calculation। |
| `pending_reconciliation` | Request পাঠানো হয়েছে, কিন্তু final usage/outcome জানা নেই। |
| `reconciled` | Provider evidence/controlled accounting review দিয়ে ledger সংশোধিত। |

Unknown outcome-এ reservation TTL expire হয়েছে বলে খরচ zero ধরে release নয়। Network timeout/cancel-এর পর request billed হয়ে থাকতে পারে; usage pending রেখে নতুন retry থামাবে অথবা explicit bounded recovery policy প্রয়োগ করবে। Cancellation আগের incurred charge ফিরিয়ে দেয় না।

**Spending cap-এর সীমানা:** gateway নতুন requests-এর admission সীমিত করে; provider-এর নিজের billing enforcement-এর বিকল্প নয়। Price-card drift, token estimation error, unreported usage এবং একই key অন্য app-এ ব্যবহার হলে exact final invoice guarantee নেই। Provider-side spending controls পাওয়া গেলে সেটিও configure করতে হবে।

Unknown price-এর route platform-funded auto-run-এ disabled থাকবে। Personal BYOK-এ operator-approved `token_only` mode রাখা যেতে পারে, কিন্তু “money cap guaranteed” দেখাবে না। Unknown context/output/usage semantics থাকলে unattended AI workflow enable নয়।

---

## 11. Context minimization এবং safe memory

### 11.1 Context builder

Task-specific fields fetch করবে, entire database/profile/history নয়। Cover letter-এ selected experience, skills, verified company/job facts যথেষ্ট হলে passport, full address বা private inbox লাগবে না।

Resume raw text একবার parse ও version করবে। পরের task-এ relevant `profile_facts` এবং source IDs যাবে। Scraped HTML প্রথমে text extraction/normalization হবে; scripts, styles, navigation ও repetitive boilerplate model prompt-এ যাবে না।

Meilisearch দিয়ে authorized top matches retrieve করা যাবে; detail Go/PostgreSQL থেকে reauthorize করে hydrate করবে। প্রথম release-এ embeddings/vector search বাধ্যতামূলক নয়। Long-lived “memory” হলো versioned database facts, unlimited in-process chat transcript নয়।

### 11.2 Long inputs

UTF-8 byte size থেকে exact token count অনুমান করা যাবে না; বিশেষ করে বাংলা/English mixed text-এ fixed `characters ÷ 4` shortcut ব্যবহার নয়। Available provider-compatible tokenizer বা verified count API ব্যবহার করবে। অন্য ক্ষেত্রে conservative estimate + safety margin + documented status।

Input বড় হলে section-aware selection/chunking হবে। `coverage: full | partial` এবং excluded sections record করতে হবে। Facts extraction অসম্পূর্ণ হলে profile complete বলা যাবে না। Reduced prompt-এর পাশাপাশি source spans থাকবে যাতে suggestion verify করা যায়।

### 11.3 Chat memory

Stateless form tasks history পাঠাবে না। Chat workflow-এ default: bounded recent turns, compact confirmed-fact snapshot এবং প্রয়োজনীয় evidence। উদাহরণ: সর্বোচ্চ 6 recent messages + 600-token summary, সবই task input cap-এর মধ্যে।

প্রতি message-এ summary generation নয়; threshold ছাড়ালে persisted summary update। Summary call-ও run/user budget consumes করবে। Confirmed career facts narrative summary থেকে পুনর্গঠন নয়; original structured facts থাকবে।

---

## 12. Caching: cost, tokens এবং bandwidth আলাদা

### 12.1 Exact-result application cache

একই authorized task/input/version-এর validated existing output থাকলে provider call না করে সেটি দেখাবে। এ request-এ নতুন model tokens লাগে না; আগের generation-এর usage record থাকবে।

Cache key-এর ingredients:

```text
workspace + owner/resource visibility + authorization revision
agent/task + canonical input hash
profile/job/document/voice versions
prompt + schema + validation versions
model + adapter + generation settings
provider-data-policy / consent version
locale + explicit regeneration nonce when requested
```

Secret key cache key-তে থাকবে না। HMAC/hash দিয়ে identifiers রাখবে; private payload public/cache-shared namespace-এ নয়। Cache read-এর আগেও permission recheck এবং user deletion/revocation-এ invalidate করবে। A user's resume output অন্য user-কে match করে দেওয়া যাবে না।

Large cached artifacts private storage-এ; PostgreSQL metadata/output reference; Redis-এ ছোট pointer বা lookup key মাত্র। Exact versioned artifact reuse-এর জন্য unlimited hot RAM cache প্রয়োজন নেই। Cache entry count, byte ceiling ও cleanup job থাকবে।

**Not cacheable as reusable success:** errors, incomplete outputs, expired permission, pending provider outcome, application submit, message send বা publish side effects।

### 12.2 Provider prompt cache

Supported provider-এ stable instructions/schema reusable prefix হিসেবে রাখা যাবে। Cache boundary এবং pricing selected model/API অনুযায়ী adapter implement করবে; “same system prompt মানেই cache hit” নয়। OpenAI ও Anthropic prefix reuse-এর নিজস্ব rules document করে; Gemini caching-ও endpoint/model-dependent। [S1], [S3], [S13]

Provider cache hit মানে final answer stored আছে নয়। নতুন output generation charged হতে পারে; cache writes/storage-ও chargeable হতে পারে। Provider prompt caching-এ request body আবার পাঠাতে হতে পারে—এটি automatic bandwidth reduction নয়।

Exact-result cache, smaller context ও fewer calls হলো প্রধান সাশ্রয়ের পথ। Cache discount অনুমান করে budget reservation ছোট করবে না; observed usage থেকে settlement করবে।

---

## 13. Retry, fallback এবং schema failure policy

| Failure | Required handling |
|---|---|
| Invalid credential / revoked key | Pause connection; owner-কে notify; repeated retries নয়। |
| Unsupported model/capability | Clear setup error; matching tested model নির্বাচন। |
| Invalid input/context limit | Smaller sufficient context বা expanded-budget approval; endless retry নয়। |
| Explicit provider rate limit | Respect response backoff/reset; bounded reschedule। |
| Connection failed before dispatch is certain | One bounded retry allowed within remaining attempts/budget। |
| Timeout after possible dispatch | Usage pending; reconcile; automatic duplicate paid call নয়। |
| Invalid structured result | Reject incomplete result; at most one budgeted repair if appropriate। |
| Missing fact / unsafe inferred answer | `needs_input`; guessed answer নয়। |
| Safety refusal | User-facing status; অন্য provider-এ switch করে bypass নয়। |
| User cancellation | Stop future steps; cancel supported request best-effort; settle incurred usage। |

Strict structured output useful হলেও refusals, incomplete output এবং application-level factual errors handle করতে হবে। Valid JSON নিজে truth guarantee নয়। [S9]

Budget, provider circuit breaker এবং queue saturation retry loop-কে থামাবে। একটি global fallback chain-এ 5 provider try করা যাবে না।

---

## 14. RAM optimization এবং worker admission

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
| Task registry | Load only selected prompt/schema; all upstream skill files নয়। |

**সাশ্রয়ের trade-off:** কম parallelism-এ queued কাজের অপেক্ষা বাড়বে। Provider API ব্যবহার করলে model weights local RAM-এ থাকবে না, কিন্তু network, privacy এবং provider cost থাকবে। Local inference-এ server RAM/VRAM প্রয়োজন; parallel contexts memory আরও বাড়াতে পারে। Ollama-র documentation-ও concurrency/context-এর এই সম্পর্ক দেখায়। [S6]

### 14.2 Process controls

Go-তে bounded worker pool, reusable HTTP clients, finite connection pools, request deadlines এবং streamed I/O ব্যবহার করবে। `GOMEMLIMIT` runtime-managed memory-এর **soft limit**; RSS/container/Chrome memory-এর hard cap নয়। Container/OS limit ও allocation monitoring আলাদা লাগবে। [S7]

Node worker V8 heap limit পুরো process RSS limit নয়; Chromium ও native buffers আলাদা measure করতে হবে। `next dev` production runtime নয়; build আলাদা CI/build host-এ করা ভালো। No unbounded `Promise.all`, goroutine-per-record বা full dataset in memory।

### 14.3 Illustrative resource envelope

নিচের values একটি ছোট **8-GiB test host**-এর জন্য initial container ceilings উদাহরণ; minimum hardware promise বা benchmark নয়। Browser/media default off এবং heavy tasks serialized। Actual RSS/CPU/I/O দিয়ে tune করতে হবে।

| Process | Example container ceiling | Additional setting / note |
|---|---:|---|
| Next.js production web | 512 MiB | Build-time memory এর মধ্যে ধরা হয়নি। |
| Go API | 256 MiB | Example `GOMEMLIMIT=192MiB`। |
| Go worker | 256 MiB | Example `GOMEMLIMIT=192MiB`। |
| PostgreSQL | 1,024 MiB | Bounded connections, query memory ও work batches। |
| Redis | 256 MiB | Example `maxmemory=128mb`; overhead/persistence reserve থাকবে। |
| Meilisearch | 1,024 MiB | Start with one indexing thread and explicit indexing-memory budget। |
| On-demand TS document helper | 1,024 MiB | Measure renderer/native subprocess peak; one heavy job। |
| Optional browser process group | 1,536 MiB | Not always running; separate account isolation। |

এই example ceilings একত্রে 5,888 MiB; বাকি host memory OS, filesystem cache, native overhead ও safety headroom-এর জন্য। Ceiling allocation guarantee নয়, এবং subprocess supervision সঠিক না হলে মোট usage আলাদা হতে পারে। Lower-memory deployment-এ remote browser/render workers বা কম concurrency প্রয়োজন হতে পারে; load test ছাড়া production claim নয়।

### 14.4 Pressure response

At a configurable warning threshold, stop admitting new heavy work. At a critical threshold, pause background research/indexing/media and preserve essential UI/API access. Initial policy candidates: warning at 75% and critical at 85% of measured effective memory limit, with hysteresis before resuming। Idle browser shutdown threshold proposed 60 seconds; resume পরে persisted state থেকে।

Live user browser approval চললে idle shutdown blindly প্রয়োগ নয়; bounded interactive lease থাকবে। Kill switch নতুন কাজ থামাবে; already-performed external action undo করবে না।

---

## 15. Redis, PostgreSQL এবং Meilisearch controls

### Redis

আগের blueprint-এর **Redis Streams** থাকবে; BullMQ/Asynq-এর private formats mix করবে না। PostgreSQL outbox authoritative। Consumers small batches claim করবে, acknowledgement durable completion-এর পরে।

Control/queue Redis-এ `noeviction` policy প্রস্তাবিত, যাতে memory pressure-এ locks/queues silently evict না হয়। `noeviction` memory full হলে writes reject করতে পারে; application সেটি handle করে new dispatch pause করবে। [S10]

শুধু আলাদা Redis database number ব্যবহার করলেই eviction policy আলাদা হয় না। Cache growth-এ explicit TTL/size eviction application-controlled হবে; প্রয়োজন হলে পরে separate cache instance। Stream trim করার আগে consumer/pending state ও durable PostgreSQL recovery নিশ্চিত করতে হবে। `MAXLEN` blindly pending work মুছে দিতে পারবে না।

### PostgreSQL

Runs/usage/reference metadata relational tables-এ; large media/base64 নয়। Paginated reads, limited SQL columns, indexed tenant/time filters, short transactions এবং aggregate metrics table থাকবে। Every token delta database row হিসেবে save নয়; bounded progress batching করবে। Usage finalization idempotent হবে।

### Meilisearch

Only searchable selected fields index করবে। Raw resumes, full private inbox বা entire crawling HTML corpus নয়। Go facade query/results/facets authorize করবে; deletion tombstone eventual indexing lag-এও hydration block করবে।

Configure indexing threads এবং indexing-memory budget explicitly; এই budget total process RAM ceiling নয়। Actual settings pinned version-এর official configuration reference অনুযায়ী যাচাই করবে। [S11]

---

## 16. Bandwidth optimization

| Source of pressure | Required control |
|---|---|
| Repeated resume upload | Upload once, record content hash, reuse owner-scoped document ID; do not create cross-user existence oracle। |
| File transfer through multiple services | Direct authorized storage upload/download যেখানে সম্ভব; streamed transfer এবং bounded temp files। |
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

বড় legitimate files/tasks silently crop না করে separate large-task profile ও estimate ব্যবহার করবে। Token budget এবং byte budget আলাদা; compressed size ছোট হলেই decompressed payload safe নয়। PDF/DOCX parser quarantine, page/time limits এবং OCR pixel limitsও থাকবে।

OCR শুধু text extraction insufficient হলে প্রয়োজনীয় pages-এ; initial batch 5 pages, expanded processing opt-in। User না চাইলে manual entry fallback থাকবে।

### 16.2 Progress streaming

SSE typing experience উন্নত করতে পারে, কিন্তু streaming নিজে generated token bill কমায় না। Provider generation stream-এ partial output আসে; application reconnect/cancel behavior নিজে implement করতে হবে। [S12]

Proposed client coalescing: 500 ms window, bounded buffer, 30-second keepalive। Slow client হলে buffer endlessly grow নয়; connection close করে saved state থেকে resume। Hidden tabs routine polling বন্ধ করবে। Result validation complete হওয়ার আগে streamed text-কে approved artifact হিসেবে save নয়।

### 16.3 Bandwidth visibility

Track: provider request/response bytes, remote fetch bytes, browser bytes, user uploads/downloads এবং media transfer আলাদা dimensions। Payload bytes এবং cloud invoice egress এক জিনিস নয়; TLS/proxy/storage/CDN overhead বা unobserved paths আলাদা হতে পারে। Missing measurement-এ `unknown` দেখাবে, zero নয়।

---

## 17. Browser, OCR ও media workers

এই components logical AI agents নয়। তারা isolated tools; AI gateway shared typed task contract দিয়ে প্রয়োজন হলে dispatch করবে। Browser account access permission এই architecture বদলায় না।

Browser navigation default-এর অংশ নয়: permitted API বা stored/imported data থাকলে সেটি ব্যবহার করবে। Authentication page, private profile ও live session-এর sensitive screenshots normal logs-এ থাকবে না। Continuous live video streaming default off; explicit user session viewing এবং limited lifetime লাগবে।

Media workflow প্রথমে text brief/script; user approve করলে estimated media cost/resource reservation; তারপর selected provider। Prompt-only success-কে generated image/video success বলা যাবে না। Audio/video jobs text-token budget থেকে hidden billing করবে না; seconds/images/render-minutes আলাদাভাবে meter করবে।

Local model option advanced self-host setting। One loaded model, one parallel request, bounded context এবং idle unload policy দিয়ে শুরু করবে; model download/runtime footprint আলাদাভাবে measure করবে। App web/API host-এ automatically large model download নয়। Provider/local endpoint reachable না হলে manual workflow থাকবে।

---

## 18. Prompt, tool ও factual safety

Prompts versioned source files হবে; all 28 upstream repositories বা complete product blueprint runtime prompt-এ ঢোকানো যাবে না। Only selected task instructions/schema/reference facts load হবে।

User resume, scraped page, job description ও messages **data**, system instruction নয়। এগুলোর embedded instruction দিয়ে permissions, model routing, credential host, SQL, file path বা outbound recipient বদলানো যাবে না। Allowed tools typed এবং server-validated।

AI SQL/shell/database credentials পাবে না। Gateway নতুন APIs/function definitions internet থেকে auto-install করবে না। Internal endpoints owner/resource scope reauthorize করবে। Tool parameter-এর URL/source কেবল allowlisted task context থেকে।

Output validators আলাদা করবে: schema validity, field completeness, source traceability, factual consistency এবং editorial preference। Fact missing হলে question; output cut off হলে incomplete; provider refused হলে refusal। Short reasoning summary চাওয়া যাবে, কিন্তু লম্বা hidden reasoning transcript prompt/output requirement নয়।

---

## 19. Data model extensions

Existing shared credentials, workflow এবং usage tables reuse করবে; নতুন overlapping ledger তৈরি নয়। নিচের names indicative migrations, currently existing table দাবি নয়।

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

Secrets encrypted vault-এ, artifacts object storage-এ। Agent labels/profile facts/package name metric label বানিয়ে high-cardinality telemetry বাড়াবে না। Large record histories paginated/retention controlled হবে।

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

`ListModels`, token-count endpoint, batch execution ও cancellation সব provider-এ mandatory নয়; optional capability interfaces হবে। `Generate` budget enforce করবে এবং incomplete usage report করতে পারবে। HTTP transport settings, auth injection ও SSRF protections shared থাকবে।

---

## 20. API contracts এবং UI behavior

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

`estimate` কখনো provider count API ব্যবহার করলে তা generation নয়, কিন্তু network/rate allowance-এ counted হবে। API না থাকলে local conservative estimate। Arbitrary user-written prompt/proxy endpoint নয়; `task`, `resource_ids`, validated options পাঠাবে। Go অনুমোদিত data hydrate করবে।

**Run status:** `queued`, `running`, `needs_input`, `needs_review`, `budget_blocked`, `resource_wait`, `failed`, `cancel_requested`, `cancelled`, `completed`। Provider usage settlement status আলাদা field, যাতে completed artifact-এর accounting pending থাকতে পারে।

---

## 21. Proposed configuration example

এই YAML intended configuration contract; parser/validation এখনো implement হয়নি। Production monetary values operator-এর সিদ্ধান্ত ছাড়া activate হবে না। Secrets এই file-এ নয়।

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

`null` অর্থ production monetary policy configured নয়; unlimited নয়। Platform-managed route সেক্ষেত্রে blocked। Personal BYOK token-only mode চাইলে explicit আলাদা policy/consent লাগবে। Per-task settings section 9 থেকে routes-এ resolve হবে; global defaults task-specific safety caps override করবে না।

---

## 22. Dashboard metrics ও tooltips

### User view

Active provider/model, payer, run estimate, input/output usage, remaining allowance, cache reused badge, generation version, waiting reason এবং stop control। নিজের API key ব্যবহারে “provider-billed usage estimate” label থাকবে।

### Admin view

Workspace spend/reservations, quotas, shared provider health, queue pressure, allowed models এবং task-level aggregated outcomes। Private resume/prompts default visible নয়।

### Super Admin view

Platform aggregate usage, connection error classes, resource saturation, price-card age, pending reconciliation, adapter verification এবং kill switches। “Monitoring” করার জন্য every prompt log নয়।

### Useful tooltips

> **Economy mode:** “আগে rules ও saved result ব্যবহার হবে। AI দরকার হলে এই task-এর জন্য পরীক্ষিত কম-খরচের model নেওয়া হবে।”

> **Token limit:** “Input, instructions, schema এবং model-এর billable generation budget আলাদাভাবে ধরা হচ্ছে। ছোট উত্তর হলেও reasoning usage থাকতে পারে।”

> **Cache reused:** “এই input/version-এর আগের result ব্যবহার করা হয়েছে; নতুন generation call করা হয়নি।”

> **API spend:** “এটি এই app-এর tracked usage estimate। একই key অন্য app-এ ব্যবহার করলে provider bill বেশি হতে পারে।”

> **Low-RAM mode:** “একসঙ্গে কম কাজ চলবে। কিছু কাজ queue-তে অপেক্ষা করবে, কিন্তু local model বা browser অকারণে চালু হবে না।”

> **Cancel:** “পরবর্তী কাজ থামবে। Provider ইতোমধ্যে request process করলে সেই usage charge থাকতে পারে।”

> **Custom API:** “API format, endpoint, model ও security rules compatible হলে ব্যবহার করা যাবে। শুধু key থাকলেই সব capability পাওয়া যায় না।”

---

## 23. Implementation task list

সব নিচের software task **TODO**। Existing `TASKS.md`-এ নতুন `AIARC-*` prefix ব্যবহার করবে; `FND-*` IDs renumber নয়।

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

**প্রথম AI vertical slice:** add one compatible provider → test with synthetic input → manually confirmed profile → one budgeted resume suggestion → review → measured usage → repeat identical request returns existing result। Live LinkedIn action প্রয়োজন নেই।

---

## 24. Acceptance tests

| ID | Required test |
|---|---|
| AI-AT-001 | Personal credential অন্য user/Admin API response, logs, HTML বা exports-এ প্রকাশ পায় না। |
| AI-AT-002 | Unapproved endpoint, redirect, metadata IP ও DNS rebinding route blocked। |
| AI-AT-003 | Base-URL change পুরোনো credential অন্য host-এ পাঠায় না। |
| AI-AT-004 | Ten concurrent identical submissions produce one logical run and at most one initial dispatch। |
| AI-AT-005 | Different owners/profile versions cannot reuse private cached output। |
| AI-AT-006 | Budget-এর চেয়ে বড় reservation হলে কোনো provider call হয় না। |
| AI-AT-007 | Concurrent requests aggregate payer limit অতিক্রম করে admit হয় না। |
| AI-AT-008 | Reasoning/cache input accounting nested totals double count করে না। |
| AI-AT-009 | Timeout after dispatch keeps usage pending; blind fallback/zero-cost refund নয়। |
| AI-AT-010 | Retry/repair/fallback সব একই run-এর attempt এবং cost budget consumes করে। |
| AI-AT-011 | Provider capability mismatch clear error দেয়; silently ignored caps নয়। |
| AI-AT-012 | Bengali/English long-input fixtures token/byte caps মানে; partial extraction honest। |
| AI-AT-013 | Unknown career fact generated answer হিসেবে confirmed হয় না। |
| AI-AT-014 | AI-generated permission flag দিয়ে message/apply/publish execute হয় না। |
| AI-AT-015 | Refresh/SSE reconnect/new tab কোনো নতুন paid generation শুরু করে না। |
| AI-AT-016 | Slow SSE consumer bounded memory রাখে এবং recoverable disconnect হয়। |
| AI-AT-017 | Redis full/restart-এ budgets/runs হারায় না; recovery duplicates paid request করে না। |
| AI-AT-018 | Browser/OCR/media default disabled অথবা explicit resource-admitted; no surprise model download। |
| AI-AT-019 | Memory pressure-এ heavy queue pauses; ordinary forms/tracking work করে। |
| AI-AT-020 | Byte/zip/page limits malicious/oversized input reject করে; parser memory bounded। |
| AI-AT-021 | Revoked consent/key এবং deleted data queued work/cache থেকে কার্যকরভাবে removed। |
| AI-AT-022 | No AI key/budget exhaustion হলেও deterministic profile edit, CSV ও existing resume export চলে। |
| AI-AT-023 | Model routing benchmark includes factual accuracy/schema success/Bangla quality, শুধু token price নয়। |
| AI-AT-024 | Advanced/media workflow estimated payer/cost/resources দেখিয়ে explicit approval নেয়। |

**Load-test report-এ যা measure করবে:** p50/p95 latency, peak/RSS per process, CPU, queue wait, provider calls per completed task, input/generation tokens, exact-cache hit rate, observed bytes, timeout/repair rate ও unresolved billing count।

Fixtures synthetic হবে; credentials ও real resumes version control-এ নয়। Unit/contract/load tests documentation creation-এর অংশ হিসেবে run হয়েছে বলে দাবি করা যাবে না।

---

## 25. Existing blueprint-এ integration এবং coder handoff

| Existing file | Required update during integration |
|---|---|
| `README.md` | এই file-এর link এবং AI-provider setup/read order যোগ করবে। Five-file pack থেকে extension আছে তা জানাবে। |
| `CONTEXT.md` | REQ-AI-001–REQ-AI-012 append; dashboard BYOK scopes, Economy default, no direct AI side effects। |
| `PLAN.md` | Gateway inside Go modular monolith, capability adapters, durable spend reservations, bounded resources add। |
| `TASKS.md` | AIARC-001–AIARC-018 append; existing IDs/dependencies preserve; AI-AT mapping include। |
| `PROGRESS.md` | Document added বনাম software implemented আলাদা; actual task/test/provider status লিখবে। |

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

## 26. Sources এবং verification boundaries

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

**Completion state:** Architecture document delivered. Software implementation, real provider connections, API charges, tests of the future application এবং deployment এখনো এই document দ্বারা সম্পন্ন হয়নি।
