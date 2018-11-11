CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        original TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

/* create index on short, and original_url */
CREATE UNIQUE INDEX url_original_idx ON urls (original);

CREATE TABLE IF NOT EXISTS url_stats (
        id SERIAL PRIMARY KEY,
        url_id INTEGER NOT NULL REFERENCES urls (id),
        date DATE NOT NULL,
        hits INTEGER NOT NULL DEFAULT 0,
        updated_at TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX url_stats_id_date ON url_stats (url_id, date);

