-- ============================================================================
-- Jokedle Database Setup Script for PostgreSQL
-- ============================================================================

-- Create the jokedle database (run this as superuser first)
-- CREATE DATABASE jokedle_db;
-- \c jokedle_db;

-- Ensure we're using the public schema
SET search_path TO public;

-- Drop tables if they exist (for clean setup)
DROP TABLE IF EXISTS jokesubmission CASCADE;
DROP TABLE IF EXISTS sequences CASCADE;
DROP TABLE IF EXISTS jokes CASCADE;

-- ============================================================================
-- JOKES TABLE - Main jokes storage
-- ============================================================================
CREATE TABLE jokes (
    jokeid SERIAL PRIMARY KEY,
    setup VARCHAR(255) NOT NULL,
    punchline VARCHAR(50) NOT NULL,
    formattedpunchline TEXT,
    source VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add comments for documentation
COMMENT ON TABLE jokes IS 'Main jokes table storing all approved jokes';
COMMENT ON COLUMN jokes.jokeid IS 'Primary key - unique joke identifier';
COMMENT ON COLUMN jokes.setup IS 'The joke setup/question (max 255 chars)';
COMMENT ON COLUMN jokes.punchline IS 'The joke punchline/answer (max 50 chars)';
COMMENT ON COLUMN jokes.formattedpunchline IS 'HTML/Markdown formatted punchline';
COMMENT ON COLUMN jokes.source IS 'Source of the joke (optional, max 45 chars)';

-- ============================================================================
-- SEQUENCES TABLE - Manages joke sequences (like "Joke of the Day") 
-- ============================================================================
CREATE TABLE sequences (
    id SERIAL PRIMARY KEY,
    sequence_name VARCHAR(100) NOT NULL UNIQUE,
    sequence_nbr INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add comments
COMMENT ON TABLE sequences IS 'Manages joke sequences and current position';
COMMENT ON COLUMN sequences.sequence_name IS 'Name of the sequence (e.g., "JokeOfDay")';
COMMENT ON COLUMN sequences.sequence_nbr IS 'Current joke ID for this sequence';

-- ============================================================================
-- JOKE SUBMISSIONS TABLE - User-submitted jokes awaiting approval
-- ============================================================================
CREATE TABLE jokesubmission (
    submissionid SERIAL PRIMARY KEY,
    setup VARCHAR(255) NOT NULL,
    punchline VARCHAR(50) NOT NULL,
    source VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by VARCHAR(100),
    reviewed_at TIMESTAMP WITH TIME ZONE
);

-- Add comments
COMMENT ON TABLE jokesubmission IS 'User-submitted jokes pending approval';
COMMENT ON COLUMN jokesubmission.status IS 'Status: pending, approved, rejected';
COMMENT ON COLUMN jokesubmission.reviewed_by IS 'Admin who reviewed the submission';

-- ============================================================================
-- INDEXES for performance optimization
-- ============================================================================

-- Jokes table indexes
CREATE INDEX idx_jokes_jokeid ON jokes(jokeid);
CREATE INDEX idx_jokes_source ON jokes(source);
CREATE INDEX idx_jokes_created_at ON jokes(created_at DESC);

-- Sequences table indexes  
CREATE INDEX idx_sequences_name ON sequences(sequence_name);
CREATE INDEX idx_sequences_nbr ON sequences(sequence_nbr);

-- Submissions table indexes
CREATE INDEX idx_submissions_created ON jokesubmission(created_at DESC);
CREATE INDEX idx_submissions_status ON jokesubmission(status);

-- ============================================================================
-- TRIGGERS for automatic timestamp updates
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables that have updated_at
CREATE TRIGGER update_jokes_updated_at 
    BEFORE UPDATE ON jokes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sequences_updated_at 
    BEFORE UPDATE ON sequences 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- INITIAL DATA SETUP
-- ============================================================================

-- Insert the default "Joke of the Day" sequence
INSERT INTO sequences (sequence_name, sequence_nbr) VALUES 
('JokeOfDay', 1);

-- Insert some sample jokes
INSERT INTO jokes (setup, punchline, formattedpunchline, source) VALUES
('Why don''t scientists trust atoms?', 
 'Because they make up everything!', 
 'Because they **make up** everything!', 
 'Classic'),

('What do you call a fake noodle?', 
 'An impasta!', 
 'An **impasta**!', 
 'Pun'),

('Why did the scarecrow win an award?', 
 'He was outstanding in his field!', 
 'He was **outstanding** in his field!', 
 'Farm Humor'),

('What do you call a dinosaur that crashes his car?', 
 'Tyrannosaurus Wrecks!', 
 'Tyrannosaurus **Wrecks**!', 
 'Dinosaur'),

('Why don''t eggs tell jokes?', 
 'They''d crack each other up!', 
 'They''d **crack** each other up!', 
 'Food'),

('What did the ocean say to the beach?', 
 'Nothing, it just waved!', 
 'Nothing, it just **waved**!', 
 'Nature'),

('Why did the math book look so sad?', 
 'Because it had too many problems!', 
 'Because it had too many **problems**!', 
 'Math'),

('What do you call a bear with no teeth?', 
 'A gummy bear!', 
 'A **gummy** bear!', 
 'Animal'),

('Why did the coffee file a police report?', 
 'It got mugged!', 
 'It got **mugged**!', 
 'Coffee'),

('What do you call a sleeping bull?', 
 'A bulldozer!', 
 'A **bull-dozer**!', 
 'Animal');

-- Insert some sample joke submissions
INSERT INTO jokesubmission (setup, punchline, source) VALUES
('Why did the bicycle fall over?', 
 'Because it was two-tired!', 
 'User Submitted'),

('What do you call a fish wearing a crown?', 
 'A king fish!', 
 'Community');

-- Update the sequence to point to the first joke
UPDATE sequences SET sequence_nbr = 1 WHERE sequence_name = 'JokeOfDay';

-- ============================================================================
-- USEFUL VIEWS (Optional)
-- ============================================================================

-- View to get jokes currently in sequences  
CREATE VIEW jokes_in_sequence AS
SELECT j.*, s.sequence_name 
FROM jokes j
JOIN sequences s ON j.jokeid = s.sequence_nbr;

-- View to get submission statistics
CREATE VIEW submission_stats AS
SELECT 
    status,
    COUNT(*) as count,
    MIN(created_at) as earliest_submission,
    MAX(created_at) as latest_submission
FROM jokesubmission 
GROUP BY status;

-- ============================================================================
-- GRANT PERMISSIONS (Optional - adjust based on your user setup)
-- ============================================================================

-- Grant permissions to your application user (replace 'gorm' with your username)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gorm;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gorm;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Check that everything was created successfully
SELECT 'Tables created:' as info;
SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;

SELECT 'Sample data count:' as info;
SELECT 'jokes' as table_name, COUNT(*) as count FROM jokes
UNION ALL
SELECT 'sequences' as table_name, COUNT(*) as count FROM sequences  
UNION ALL
SELECT 'jokesubmission' as table_name, COUNT(*) as count FROM jokesubmission;

SELECT 'Current joke sequence:' as info;
SELECT sequence_name, sequence_nbr FROM sequences;

-- Test query that your application will use
SELECT 'Jokes in sequence test:' as info;
SELECT j.jokeid, j.setup, j.punchline 
FROM jokes j 
WHERE j.jokeid IN (SELECT sequence_nbr FROM sequences);

-- ============================================================================
-- End of setup script
-- ============================================================================

COMMIT;