CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    userid INT NOT NULL,
    refresh_token VARCHAR(512) NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    expiration TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days',
    createdat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userid) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
