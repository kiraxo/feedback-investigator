CREATE TABLE IF NOT EXISTS agent_runs (
    id UUID PRIMARY KEY,
    status TEXT NOT NULL
        CHECK (status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED')),
    mode TEXT NOT NULL
        CHECK (mode IN ('DEMO', 'LIVE')),
    provider TEXT NOT NULL
        CHECK (provider IN ('DEMO', 'OPENAI', 'ANTHROPIC', 'GROK', 'OLLAMA')),
    feedback_count INTEGER NOT NULL
        CHECK (feedback_count >= 0),
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS agent_runs_created_at_idx
    ON agent_runs (created_at DESC);

CREATE INDEX IF NOT EXISTS agent_runs_status_idx
    ON agent_runs (status);

CREATE INDEX IF NOT EXISTS agent_runs_provider_idx
    ON agent_runs (provider);

CREATE INDEX IF NOT EXISTS agent_runs_payload_gin_idx
    ON agent_runs USING GIN (payload);