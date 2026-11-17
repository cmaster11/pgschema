CREATE TYPE status_new AS ENUM ('a', 'b', 'c');

CREATE TYPE mood_new AS ENUM ('calm', 'busy');

CREATE TYPE level_new AS ENUM ('low', 'mid', 'high');

CREATE TYPE priority_new AS ENUM ('low', 'mid', 'high');

CREATE TYPE settings AS (
    level level_new
);

CREATE FUNCTION default_priority() RETURNS priority_new
    LANGUAGE sql
    AS $$ SELECT 'low'::priority_new $$;

CREATE TABLE orders (
    id integer PRIMARY KEY,
    status status_new DEFAULT 'a'::status_new NOT NULL,
    history status_new[],
    kind text,
    mood mood_new,
    note text DEFAULT 'soft'
);

CREATE TABLE shipments (
    id integer PRIMARY KEY,
    status status_new,
    grade text,
    CONSTRAINT status_not_b CHECK (status <> 'b'::status_new)
);

CREATE INDEX shipments_a_idx ON shipments (id) WHERE status = 'a'::status_new;

CREATE TABLE notes (
    id integer PRIMARY KEY,
    note text
);

ALTER TABLE notes ENABLE ROW LEVEL SECURITY;

CREATE POLICY notes_own ON notes
    USING (note = 'own')
    WITH CHECK (note = 'me');

CREATE FUNCTION notes_touch() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN RETURN NEW; END $$;

CREATE TRIGGER notes_touch BEFORE UPDATE ON notes
    FOR EACH ROW WHEN (NEW.note = 'on')
    EXECUTE FUNCTION notes_touch();
