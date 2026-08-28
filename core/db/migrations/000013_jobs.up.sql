-- Migration 000013: jobs (up)
-- Postgres SKIP LOCKED queue. Dual-run with Pub/Sub until the worker switches.

CREATE TABLE jobs (
    id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
    topic TEXT NOT NULL,
    consumer TEXT,
    body BYTEA NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'done', 'dead')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX jobs_claim_idx ON jobs (topic, available_at)
    WHERE status IN ('pending', 'processing');

COMMENT ON TABLE jobs IS 'Agnostic job bus. Claimed with SELECT/UPDATE FOR UPDATE SKIP LOCKED.';
COMMENT ON COLUMN jobs.consumer IS 'Worker id that currently holds the claim.';
COMMENT ON COLUMN jobs.available_at IS 'Eligible for claim when <= now(). Visibility timeout while processing; backoff after failure.';
