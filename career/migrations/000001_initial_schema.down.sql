-- Migration 000001 Rollback: Clean teardown in reverse dependency order

-- 6. Provider Accounts & Content
DROP TABLE IF EXISTS content_items CASCADE;
DROP TABLE IF EXISTS quota_reservations CASCADE;
DROP TABLE IF EXISTS provider_accounts CASCADE;
DROP TYPE IF EXISTS post_status CASCADE;
DROP TYPE IF EXISTS provider_type CASCADE;

-- 5. Workflow & Outbox
DROP TABLE IF EXISTS scheduled_jobs CASCADE;
DROP TABLE IF EXISTS outbox_events CASCADE;
DROP TABLE IF EXISTS approvals CASCADE;
DROP TABLE IF EXISTS action_runs CASCADE;
DROP TYPE IF EXISTS run_state CASCADE;

-- 4. Career & Applications
DROP TABLE IF EXISTS application_events CASCADE;
DROP TABLE IF EXISTS applications CASCADE;
DROP TABLE IF EXISTS saved_jobs CASCADE;
DROP TABLE IF EXISTS jobs CASCADE;
DROP TYPE IF EXISTS confirmation_type CASCADE;
DROP TYPE IF EXISTS application_status CASCADE;

-- 3. Documents & Resumes
DROP TABLE IF EXISTS resume_versions CASCADE;
DROP TABLE IF EXISTS uploads CASCADE;

-- 2. Profiles & Facts
DROP TABLE IF EXISTS job_preferences CASCADE;
DROP TABLE IF EXISTS profile_facts CASCADE;
DROP TABLE IF EXISTS profiles CASCADE;

-- 1. Identity & Workspace
DROP TABLE IF EXISTS resource_grants CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS memberships CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS workspace_role CASCADE;
DROP TYPE IF EXISTS platform_role CASCADE;