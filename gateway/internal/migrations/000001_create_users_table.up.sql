CREATE TABLE users (
   login VARCHAR(255) PRIMARY KEY,
   password_hash TEXT NOT NULL,
   created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_login ON users(login);