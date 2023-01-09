CREATE TABLE IF NOT EXISTS migrations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(32) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT current_timestamp NOT NULL
);

