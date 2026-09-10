CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT NOT NULL UNIQUE,
    bottle_balance BIGINT NOT NULL DEFAULT 0 CHECK (bottle_balance >= 0),
    last_passive_claim_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    photo_url TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '' CHECK (char_length(description) <= 2000),
    bottle_count BIGINT NOT NULL DEFAULT 0 CHECK (bottle_count >= 0),
    total_bottles_received BIGINT NOT NULL DEFAULT 0 CHECK (total_bottles_received >= 0),
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    lock_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((is_locked = FALSE) OR lock_until IS NOT NULL)
);
CREATE INDEX targets_leaderboard_idx ON targets (bottle_count DESC, created_at DESC);
CREATE INDEX targets_search_idx ON targets USING GIN (to_tsvector('simple', name || ' ' || description));

CREATE TABLE nabutilations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id),
    target_id UUID NOT NULL REFERENCES targets(id),
    cost BIGINT NOT NULL DEFAULT 1000,
    comment TEXT NOT NULL CHECK (char_length(comment) BETWEEN 1 AND 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX nabutilations_target_idx ON nabutilations (target_id, created_at DESC);

CREATE TABLE payment_events (
    provider TEXT NOT NULL,
    external_id TEXT NOT NULL,
    player_id UUID NOT NULL REFERENCES players(id),
    bottles BIGINT NOT NULL CHECK (bottles > 0),
    payload JSONB NOT NULL,
    credited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (provider, external_id)
);
