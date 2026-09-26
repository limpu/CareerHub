-- Migration 000001: Initial Schema
-- Authoritative PostgreSQL Storage Blueprint
-- Compliant with PLAN.md Section 5 and AT-011 Invariants

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. IDENTITY & WORKSPACE DOMAIN
-- ============================================================================

CREATE TYPE platform_role AS ENUM ('super_admin', 'operator', 'standard');
CREATE TYPE workspace_role AS ENUM ('admin', 'user');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    platform_role platform_role NOT NULL DEFAULT 'standard',
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE memberships (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role workspace_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Scoped Resource Grants: AT-011 enforces that admin access to a workspace
-- DOES NOT automatically grant access to another member's private career records.
CREATE TABLE resource_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    granter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grantee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL, -- 'resume', 'profile', 'application'
    resource_id UUID NOT NULL,
    permission VARCHAR(50) NOT NULL,    -- 'read', 'review', 'comment'
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. PROFILE & CAREER FACTS DOMAIN
-- ============================================================================

CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    headline VARCHAR(255),
    summary TEXT,
    contact_phone VARCHAR(50),
    contact_location VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profile_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- 'experience', 'education', 'skill', 'project', 'certification'
    title VARCHAR(255) NOT NULL,
    organization VARCHAR(255),
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    details JSONB NOT NULL DEFAULT '[]'::jsonb,
    confirmed BOOLEAN NOT NULL DEFAULT TRUE,
    source_provenance VARCHAR(100) NOT NULL DEFAULT 'manual_entry', -- 'upload_pdf', 'upload_docx', 'manual_entry'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE job_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    target_roles TEXT[] NOT NULL DEFAULT '{}',
    work_modes TEXT[] NOT NULL DEFAULT '{}', -- 'remote', 'hybrid', 'onsite'
    locations TEXT[] NOT NULL DEFAULT '{}',
    min_salary NUMERIC(12, 2),
    salary_currency VARCHAR(10) DEFAULT 'USD',
    sponsorship_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3. DOCUMENTS & RESUME DOMAIN (OWNER-PRIVATE BY DEFAULT)
-- ============================================================================

CREATE TABLE uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(500) NOT NULL UNIQUE,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    quarantine_status VARCHAR(50) NOT NULL DEFAULT 'clean', -- 'quarantine', 'scanning', 'clean', 'rejected'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE resume_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    template_id VARCHAR(100) NOT NULL DEFAULT 'standard-ats-v1',
    content_spec JSONB NOT NULL,
    is_master BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_resume_version UNIQUE (user_id, version_number)
);

-- ============================================================================
-- 4. CAREER & APPLICATION DOMAIN
-- ============================================================================

CREATE TYPE application_status AS ENUM (
    'draft', 'preparing', 'needs_input', 'ready_for_review', 'approved',
    'submitting', 'applied', 'recruiter_response', 'interview', 'offer',
    'rejected', 'withdrawn', 'needs_confirmation', 'failed_before_submit'
);

CREATE TYPE confirmation_type AS ENUM (
    'provider_confirmed', 'user_confirmed', 'unconfirmed'
);

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(255),
    source VARCHAR(50) NOT NULL, -- 'linkedin', 'indeed', 'glassdoor', 'manual'
    title VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    work_mode VARCHAR(50),
    job_url TEXT NOT NULL,
    salary_min NUMERIC(12, 2),
    salary_max NUMERIC(12, 2),
    salary_currency VARCHAR(10),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_source_external_job UNIQUE (source, external_id)
);

CREATE TABLE saved_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    notes TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_saved_job UNIQUE (user_id, job_id)
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE RESTRICT,
    resume_version_id UUID REFERENCES resume_versions(id) ON DELETE SET NULL,
    status application_status NOT NULL DEFAULT 'draft',
    confirmation_type confirmation_type NOT NULL DEFAULT 'unconfirmed',
    applied_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE application_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status_from application_status,
    status_to application_status NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 5. WORKFLOW, OUTBOX & APPROVAL DOMAIN
-- ============================================================================

CREATE TYPE run_state AS ENUM ('queued', 'running', 'needs_approval', 'paused', 'completed', 'failed', 'cancelled');

CREATE TABLE action_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    task_type VARCHAR(100) NOT NULL,
    state run_state NOT NULL DEFAULT 'queued',
    input_payload JSONB NOT NULL,
    result_data JSONB,
    error_details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_run_id UUID NOT NULL REFERENCES action_runs(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    required BOOLEAN NOT NULL DEFAULT TRUE,
    payload_hash VARCHAR(64) NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'dispatched', 'failed'
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMPTZ
);

CREATE TABLE scheduled_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_name VARCHAR(100) NOT NULL,
    cron_spec VARCHAR(50) NOT NULL,
    next_run_at TIMESTAMPTZ NOT NULL,
    locked_by VARCHAR(100),
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 6. PROVIDER ACCOUNTS, QUOTAS & SOCIAL DOMAIN
-- ============================================================================

CREATE TYPE provider_type AS ENUM ('linkedin', 'twitter', 'facebook', 'instagram', 'youtube');
CREATE TYPE post_status AS ENUM ('draft', 'pending_approval', 'scheduled', 'published', 'failed');

CREATE TABLE provider_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider provider_type NOT NULL,
    account_uid VARCHAR(255) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_workspace_provider_account UNIQUE (workspace_id, provider, account_uid)
);

CREATE TABLE quota_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_account_id UUID NOT NULL REFERENCES provider_accounts(id) ON DELETE CASCADE,
    action_kind VARCHAR(50) NOT NULL,
    window_date DATE NOT NULL,
    amount INT NOT NULL DEFAULT 1,
    settled BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE content_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    target_platforms provider_type[] NOT NULL DEFAULT '{}',
    status post_status NOT NULL DEFAULT 'draft',
    scheduled_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE & TENANT ISOLATION
-- ============================================================================

CREATE INDEX idx_memberships_user ON memberships(user_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_profile_facts_user ON profile_facts(user_id);
CREATE INDEX idx_resume_versions_user ON resume_versions(user_id);
CREATE INDEX idx_applications_user ON applications(user_id);
CREATE INDEX idx_applications_job ON applications(job_id);
CREATE INDEX idx_action_runs_workspace ON action_runs(workspace_id);
CREATE INDEX idx_outbox_events_status ON outbox_events(status) WHERE status = 'pending';
CREATE INDEX idx_quota_reservations_acc_win ON quota_reservations(provider_account_id, window_date);
CREATE INDEX idx_content_items_workspace ON content_items(workspace_id);