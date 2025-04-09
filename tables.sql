-- Script para crear la tabla "series"
CREATE TABLE IF NOT EXISTS series (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    current_episode INTEGER DEFAULT 0,
    total_episodes INTEGER,
    status VARCHAR(50),
    score INTEGER DEFAULT 0
);