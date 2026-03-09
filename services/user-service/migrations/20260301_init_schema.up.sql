CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    dietary_flags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_pantry (
    user_id UUID NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, ingredient_name)
);

CREATE TABLE user_shopping_list (
    user_id UUID NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, ingredient_name)
);

CREATE TABLE user_wishlist (
    user_id UUID NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, ingredient_name)
);

CREATE TABLE consumed_recipes (
    user_id UUID NOT NULL,
    recipe_id UUID NOT NULL,
    times_made INT DEFAULT 1,
    last_made_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, recipe_id)
);

CREATE TABLE consumed_ingredients (
    user_id UUID NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    times_used INT DEFAULT 1,
    PRIMARY KEY (user_id, ingredient_name)
);

CREATE TABLE shopping_list_suggestions (
    user_id UUID NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    reason VARCHAR(255), -- "used in {recipe}", "frequently used", etc.
    suggested_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, ingredient_name)
);

CREATE INDEX idx_pantry_user ON user_pantry(user_id, ingredient_name);
CREATE INDEX idx_shopping_list_user ON user_shopping_list(user_id, ingredient_name);
CREATE INDEX idx_wishlist_user ON user_wishlist(user_id, ingredient_name);