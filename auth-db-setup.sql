CREATE TABLE tokens (
  idtokens SERIAL PRIMARY KEY,
  createDateTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  username VARCHAR(45) NOT NULL,
  token VARCHAR(45) NOT NULL,
  CONSTRAINT token_UNIQUE UNIQUE (token)
);
CREATE TABLE users (
  idusers SERIAL PRIMARY KEY,
  username VARCHAR(45) NOT NULL,
  password VARCHAR(100) NOT NULL,
  CONSTRAINT username_unique UNIQUE (username)
);
INSERT INTO users (username, password) 
VALUES ('jokedleadmin', 'password');
INSERT INTO tokens (username, token)
VALUES ('jokedleadmin', 'token');

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO gorm;

-- Grant permissions on existing auth tables
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gorm;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gorm;

-- Grant permissions on future tables (important for new tables)
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO gorm;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO gorm;

-- For simple tokens and users tables:
GRANT SELECT, INSERT, UPDATE, DELETE ON tokens TO gorm;
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO gorm;
GRANT USAGE, SELECT ON tokens_idtokens_seq TO gorm;
GRANT USAGE, SELECT ON users_idusers_seq TO gorm;