CREATE TABLE IF NOT EXISTS parsing_history (
    id SERIAL PRIMARY KEY,
    search_query VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    items_found INTEGER DEFAULT 0,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP,
    error_message TEXT
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    parsing_history_id INTEGER REFERENCES parsing_history(id),
    name VARCHAR(500) NOT NULL,
    item_url TEXT NOT NULL,
    price DECIMAL(10,2),
    price_currency VARCHAR(10),
    img_url TEXT,
    parsed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
CREATE INDEX IF NOT EXISTS idx_parsing_history_query ON parsing_history(search_query);