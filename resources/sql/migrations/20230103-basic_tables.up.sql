CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    email VARCHAR(32) NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT current_timestamp NOT NULL,
    last_updated_at TIMESTAMP DEFAULT current_timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS apps (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(32) NOT NULL,
    auth_key VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    UNIQUE(owner_id, name)
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    email VARCHAR(32) NOT NULL,
    password VARCHAR(128) NOT NULL,
    auth_key VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    UNIQUE(app_id, email)
);

CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(64) NOT NULL,
    filepath VARCHAR(128) NOT NULL,
    contents BYTEA,
    file_size INT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    app_id INT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS actions (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(32) NOT NULL,
    script TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS app_memberships (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    client_id INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

