CREATE TYPE status_old AS ENUM ('a', 'b');

CREATE TYPE kind_old AS ENUM ('x', 'y');

CREATE TYPE flag_old AS ENUM ('on', 'off');

CREATE TYPE level_old AS ENUM ('low', 'high');

CREATE TYPE tone_old AS ENUM ('soft', 'loud');

CREATE TYPE grade_old AS ENUM ('good', 'poor');

CREATE TYPE priority_old AS ENUM ('low', 'high');

CREATE TYPE scope_old AS ENUM ('own', 'all');

CREATE TYPE owner_old AS ENUM ('me', 'them');

CREATE TYPE mark_old AS ENUM ('on', 'off');

CREATE TYPE settings AS (
    level level_old
);

CREATE FUNCTION default_priority() RETURNS priority_old
    LANGUAGE sql
    AS $$ SELECT 'low'::priority_old $$;

CREATE TABLE orders (
    id integer PRIMARY KEY,
    status status_old DEFAULT 'a'::status_old NOT NULL,
    history status_old[],
    kind kind_old,
    mood text,
    note text DEFAULT ('soft'::tone_old)::text,
    flag flag_old
);

CREATE TABLE shipments (
    id integer PRIMARY KEY,
    status status_old,
    grade text,
    CONSTRAINT grade_known CHECK (grade::grade_old <> 'poor'::grade_old),
    CONSTRAINT status_not_b CHECK (status <> 'b'::status_old)
);

CREATE INDEX shipments_a_idx ON shipments (id) WHERE status = 'a'::status_old;

CREATE INDEX shipments_b_idx ON shipments (id) WHERE status = 'b'::status_old;

CREATE TABLE notes (
    id integer PRIMARY KEY,
    note text
);

ALTER TABLE notes ENABLE ROW LEVEL SECURITY;

CREATE POLICY notes_own ON notes
    USING (note::scope_old = 'own'::scope_old)
    WITH CHECK (note::owner_old = 'me'::owner_old);

CREATE FUNCTION notes_touch() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN RETURN NEW; END $$;

CREATE TRIGGER notes_touch BEFORE UPDATE ON notes
    FOR EACH ROW WHEN (NEW.note::mark_old = 'on'::mark_old)
    EXECUTE FUNCTION notes_touch();
