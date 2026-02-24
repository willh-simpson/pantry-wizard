CREATE TABLE interactions (
    id UUID PRIMARY KEY,
    user_id UUID,
    recipe_id UUID,
    event_type VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_interaction_stream ON interactions(created_at DESC);