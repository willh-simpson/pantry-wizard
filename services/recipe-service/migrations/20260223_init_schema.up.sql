CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50), -- 'vegetable', 'protein', etc
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    instructions TEXT NOT NULL,
    author_id UUID NOT NULL,
    prep_time_min INT,
    calories INT,
    budget_tier INT CHECK (budget_tier BETWEEN 1 AND 3),
    image_url TEXT, -- S3 url
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE recipe_ingredients (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID REFERENCES ingredients(id) ON DELETE RESTRICT,
    amount DECIMAL NOT NULL,
    unit VARCHAR(20), -- 'grams', 'tbsp', etc
    PRIMARY KEY (recipe_id, ingredient_id)
);

CREATE INDEX idx_recipes_title ON recipes(title);
CREATE INDEX idx_ingredients_name ON ingredients(name);