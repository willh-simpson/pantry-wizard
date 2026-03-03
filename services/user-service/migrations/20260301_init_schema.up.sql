CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

CREATE INDEX idx_pantry_user ON user_pantry(user_id, ingredient_name);
CREATE INDEX idx_shopping_list_user ON user_shopping_list(user_id, ingredient_name);
CREATE INDEX idx_wishlist_user ON user_wishlist(user_id, ingredient_name);