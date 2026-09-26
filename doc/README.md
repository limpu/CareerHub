# Career / LinkedIn / Social Platform

> **Status: implementation blueprint, not a running application.** Prepared 2026-09-17. These five files specify the product and guide an AI coding agent. Application source, deployment configuration, tests and live integrations have NOT been created or executed by this documentation pack.

Before starting or continuing any work, read README.md, CONTEXT.md, PLAN.md, AI-AGENT-ARCHITECTURE.md, TASKS.md, and PROGRESS.md in full and treat them as one interconnected source of truth. Cross-reference requirements, architecture decisions, AI-agent rules, task dependencies, implementation status, and prior decisions across all six files before making changes. Do not interpret any file in isolation, silently override an existing decision, duplicate functionality, change the approved stack, or mark work complete without implementation and verification evidence. After each meaningful implementation step, update TASKS.md and PROGRESS.md so the next AI/coding session can resume accurately without losing context.

## Dual-Product Architecture & Platform Structure

The platform is structured into **two distinct flagship commercial products** served under a unified master domain (`yourdomain.com`) with separate subscription models and dedicated project directories:

```text
career-platform-monorepo
├── doc/               # Central Architecture, Specifications & Governance (.md)
├── career/            # Product 1: Career & LinkedIn Platform (yourdomain.com/career)
└── SocialSuite/       # Product 2: Social Media & Creator Studio (yourdomain.com/social)
```

| Product | Route | Target Audience | Primary Features | Subscription Tier |
|---|---|---|---|---|
| **Product 1: Career & LinkedIn OS** (`career/`) | `yourdomain.com/career` | Job Seekers, Candidates, Working Professionals | Verified profile, pure-Go ATS resumes, multi-board job discovery (Greenhouse, Lever, Ashby, Indeed), application evidence vault, Sheets/CSV exports, LinkedIn profile sync & networking. | **Career & LinkedIn Pro** |
| **Product 2: SocialSuite** (`SocialSuite/`) | `yourdomain.com/social` | Creators, Social Media Managers, Agencies, Brands | Multi-account visual publishing calendar, AI content studio (voice builder, hooks, PAS/AIDA frameworks), trend/outlier research, comment mining, and audience growth engine. | **Social / Creator Pro** |

Users may subscribe independently to **Career Pro**, **Creator Pro**, or an optional **All-Access Bundle**. Career features and Social creator workflows maintain distinct tenant and permission boundaries.


## Fixed stack

| Component | Responsibility |
|---|---|
| Next.js + TypeScript + Tailwind | Responsive web interface; App Router routes; accessible components; request/display state. |
| Go API | Authentication/authorization, domain rules, tenant isolation, billing entitlements, policies and authoritative writes. |
| Go workers | Durable workflows, search indexing, exports, notifications, discovery and provider API jobs. |
| Isolated TypeScript workers | Document/media tooling and permitted browser tasks when needed; never an alternative owner of core business rules. |
| PostgreSQL | Authoritative product state, immutable versions, approvals, billing ledger, job state and transactional outbox. |
| Redis | Transport/queues, short-lived coordination, cache and atomic action budget checks; recover from PostgreSQL. |
| Meilisearch | Derived scoped search projections; never authorization or source of truth. |
| Private object storage | Uploaded resumes, generated documents/media and sanitized evidence. Local volume for self-host; S3-compatible adapter for cloud. |

No mandatory Python/PHP service, LangGraph, Temporal, Supabase or other donor-repository backend is introduced without a documented architecture decision. Media transcoders may be isolated system dependencies rather than business-language changes.

## Roles

**Super Admin** controls the installation/platform: packages, entitlements, provider policy, global diagnostics and tightly audited support operations. **Admin** manages a workspace's team, shared assets, approvals, subscription and operational settings. **User** manages their own career records and permitted shared resources. Private career documents remain owner-private unless explicitly delegated. See the permission matrix in [CONTEXT.md](CONTEXT.md).

## Distribution

Community/self-hosted and paid managed cloud use the same product core. Cloud adds managed infrastructure, metered compute/provider services, billing, backups and support. Core records and basic export are not held hostage to an expired subscription. Self-hosting does not remove platform permissions, consent requirements or security checks.

Pricing amounts and production package quotas are not approved in this pack. Starter package names and quota fields are specified, but package activation requires explicit configuration.

## Read these files

| File | Purpose |
|---|---|
| [README.md](README.md) | Product overview, user journey, stack, current status, how to use this pack. |
| [CONTEXT.md](CONTEXT.md) | Source of truth for requirements, roles, behavior, limits, repository evidence and feature selections. |
| [PLAN.md](PLAN.md) | Architecture, storage/API contracts, state machines, phases, test strategy and release gates. |
| [TASKS.md](TASKS.md) | Stable implementation task IDs, dependencies and acceptance criteria. |
| [PROGRESS.md](PROGRESS.md) | Actual completed work, current state, blockers and next coder entry point. |

**AI coder reading order:** README → CONTEXT → PLAN → TASKS → PROGRESS. Then inspect actual repository state; do not assume code exists because a document describes it.

## Start a coding session

```text
Read README.md, CONTEXT.md, PLAN.md, TASKS.md and PROGRESS.md.
Inspect the real repository and git state. State the next unblocked task ID.
Implement one bounded task or a dependency-complete vertical slice.
Use the fixed stack and approved decisions. Source-document claims are not proof
of working code or platform permission. Do not silently drop features or swap stacks.
Do not send live applications, posts or messages during development/testing.
Run relevant checks and record actual results. Update TASKS.md and PROGRESS.md,
including changed files, tests, remaining risks and the exact next task.
Never mark a task done from a mock screen, generated code or an unexecuted test.
```

## Development setup — intended, not currently runnable

Implementation will provide a monorepo, lockfiles, Go modules, database migrations, a `.env.example` without secrets, seed fixtures and Docker Compose for web/API/workers/PostgreSQL/Redis/Meilisearch/object storage. Do not run guessed startup commands from this pack: no `package.json`, `go.mod`, executable or Compose configuration exists here yet.

The eventual developer experience should expose `make bootstrap`, `make dev`, `make test`, `make lint`, `make migrate` and `make verify-docs`, with commands implemented and validated in foundation tasks before they are advertised as working.

## Operating principles

- An ATS-friendly resume is a formatting and parsing goal, not a guarantee that every ATS will accept it or rank it highly.
- LinkedIn sign-in does not grant universal profile, messaging or job-application access. No daily quota can guarantee protection against account restrictions. See [official LinkedIn guidance](https://www.linkedin.com/help/linkedin/answer/a1341387) and [OIDC scope documentation](https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2).
- Automatic side effects require an independently allowed connector, explicit user permission and an active policy. Manual-native fallbacks are real product features.
- Never fabricate career facts. Never mark Applied/Published/Sent without provider evidence or clearly labeled user confirmation.
- A selected repository is a feature/workflow reference, not a claim that its code has been benchmarked or can be copied without review.

## Source inventory

The requested grouping contains **29 placements but 28 unique repositories**: Career 8 + LinkedIn 6 + Social 15, with `career-agent` appearing in both Career and LinkedIn. The complete linked inventory and source limitations are in CONTEXT.md.

## What this pack does not claim

It is not a production release, a completed code-level audit of all upstream files, a performance benchmark, a security certification, a universal ATS certification or an assurance against social-platform bans. All software implementation tasks are initially open. The documentation milestones alone are complete.
