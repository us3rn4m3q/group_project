-- информация о сайтах
CREATE TABLE IF NOT EXISTS sites (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- история парсинга
CREATE TABLE IF NOT EXISTS parsing_history (
    id SERIAL PRIMARY KEY,
    site_id INTEGER REFERENCES sites(id),
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP,
    error_message TEXT,
    items_parsed INTEGER DEFAULT 0
);

-- результаты парсинга
CREATE TABLE IF NOT EXISTS parsed_items (
    id SERIAL PRIMARY KEY,
    parsing_history_id INTEGER REFERENCES parsing_history(id),
    site_id INTEGER REFERENCES sites(id),
    title TEXT,
    url TEXT,
    parsed_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- оптимизация запросов
CREATE INDEX IF NOT EXISTS idx_parsing_history_site_id ON parsing_history(site_id);
CREATE INDEX IF NOT EXISTS idx_parsed_items_parsing_history_id ON parsed_items(parsing_history_id);
CREATE INDEX IF NOT EXISTS idx_parsed_items_site_id ON parsed_items(site_id);