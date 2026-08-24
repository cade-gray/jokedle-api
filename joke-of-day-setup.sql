-- ============================================================================
-- JOKE OF THE DAY - Automatic daily rotation
-- ============================================================================
-- Replaces the external bash script that advanced sequences.sequence_nbr.
-- Instead of a scheduler, the first read of each day rolls the sequence
-- forward. No cron, no pg_cron extension, no restart required.
--
-- Safe to re-run: every statement is idempotent.
--
-- >>> BEFORE RUNNING: check the timezone in jokedle_today() below. It decides
-- >>> when "tomorrow" starts for your readers. It is the only value you are
-- >>> likely to need to change.
-- ============================================================================

BEGIN;

-- ============================================================================
-- TIMEZONE HELPER - single source of truth for "what day is it"
-- ============================================================================
-- STABLE, not IMMUTABLE: the result changes over time, but never within a
-- single statement.
CREATE OR REPLACE FUNCTION jokedle_today() RETURNS date AS $$
    SELECT (now() AT TIME ZONE 'America/Chicago')::date;
$$ LANGUAGE sql STABLE;

COMMENT ON FUNCTION jokedle_today() IS
    'Current date in the app''s display timezone; determines when the joke rolls over';

-- ============================================================================
-- SCHEMA - track which day the sequence last advanced
-- ============================================================================
ALTER TABLE sequences
    ADD COLUMN IF NOT EXISTS rolled_on date NOT NULL DEFAULT jokedle_today();

COMMENT ON COLUMN sequences.rolled_on IS
    'Date this sequence last advanced; blocks a second roll on the same day';

-- ============================================================================
-- MANUAL OVERRIDE - keep admin picks from being rolled away
-- ============================================================================
-- Any hand-set sequence_nbr (the admin site's POST /joke/sequence) counts as
-- today's roll, so it survives until tomorrow. This lives in a trigger rather
-- than the API so the override works without an API change.
CREATE OR REPLACE FUNCTION mark_sequence_rolled() RETURNS trigger AS $$
BEGIN
    IF NEW.sequence_nbr IS DISTINCT FROM OLD.sequence_nbr THEN
        NEW.rolled_on := jokedle_today();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS mark_sequences_rolled ON sequences;
CREATE TRIGGER mark_sequences_rolled
    BEFORE UPDATE ON sequences
    FOR EACH ROW
    EXECUTE FUNCTION mark_sequence_rolled();

-- ============================================================================
-- JOKE OF THE DAY - rolls if needed, then returns the current joke
-- ============================================================================
-- The UPDATE advances to the next jokeid and wraps to the lowest one at the
-- end. Using MIN(jokeid) rather than sequence_nbr + 1 means gaps left by
-- deleted jokes are skipped, and a sequence_nbr pointing at a joke that no
-- longer exists heals itself on the next roll.
CREATE OR REPLACE FUNCTION joke_of_day() RETURNS SETOF jokes AS $$
    UPDATE sequences s
       SET sequence_nbr = COALESCE(
               (SELECT MIN(j.jokeid) FROM jokes j WHERE j.jokeid > s.sequence_nbr),
               (SELECT MIN(j.jokeid) FROM jokes j))
     WHERE s.sequence_name = 'JokeOfDay'
       AND s.rolled_on < jokedle_today()
       -- never NULL out a NOT NULL column when the jokes table is empty
       AND EXISTS (SELECT 1 FROM jokes);

    SELECT j.*
      FROM jokes j
      JOIN sequences s ON j.jokeid = s.sequence_nbr
     WHERE s.sequence_name = 'JokeOfDay';
$$ LANGUAGE sql;

COMMENT ON FUNCTION joke_of_day() IS
    'Advances JokeOfDay if it has not rolled today, then returns that joke';

COMMIT;

-- ============================================================================
-- VERIFICATION
-- ============================================================================
SELECT 'Timezone date:' AS info, jokedle_today() AS today;

SELECT 'Sequence state:' AS info;
SELECT sequence_name, sequence_nbr, rolled_on FROM sequences;

SELECT 'Joke of the day:' AS info;
SELECT jokeid, setup, punchline FROM joke_of_day();
