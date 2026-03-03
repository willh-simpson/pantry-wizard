CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE recipe_scores (
    recipe_id UUID PRIMARY KEY,
    like_count BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    save_count BIGINT DEFAULT 0,
    total_score DOUBLE PRECISION DEFAULT 0.0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);