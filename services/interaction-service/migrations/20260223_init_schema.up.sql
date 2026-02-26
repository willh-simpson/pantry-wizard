CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE interactions (
    id UUID PRIMARY KEY,
    user_id UUID,
    recipe_id UUID,
    event_type VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE recipe_likes (
    user_id UUID NOT NULL,
    recipe_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, recipe_id)
);

CREATE TABLE recipe_saves (
    user_id UUID NOT NULL,
    recipe_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, recipe_id)
);

CREATE TABLE recipe_views (
    user_id UUID NOT NULL,
    recipe_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, recipe_id)
);

CREATE INDEX idx_interaction_stream ON interactions(created_at DESC);
CREATE INDEX idx_likes_recipe_id ON recipe_likes(recipe_id);