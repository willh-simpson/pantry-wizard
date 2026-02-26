CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE recipe_popularity (
    recipe_id UUID PRIMARY KEY,
    like_count INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);