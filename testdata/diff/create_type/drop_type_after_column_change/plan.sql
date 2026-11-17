CREATE TYPE level_new AS ENUM (
    'low',
    'mid',
    'high'
);

CREATE TYPE mood_new AS ENUM (
    'calm',
    'busy'
);

CREATE TYPE priority_new AS ENUM (
    'low',
    'mid',
    'high'
);

CREATE TYPE status_new AS ENUM (
    'a',
    'b',
    'c'
);

ALTER TYPE settings ALTER ATTRIBUTE level TYPE level_new;

CREATE OR REPLACE TRIGGER notes_touch
    BEFORE UPDATE ON notes
    FOR EACH ROW
    WHEN (((NEW.note = 'on'::text)))
    EXECUTE FUNCTION notes_touch();

ALTER POLICY notes_own ON notes USING (note = 'own') WITH CHECK (note = 'me');

ALTER TABLE orders DROP COLUMN flag;

ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;

ALTER TABLE orders ALTER COLUMN status TYPE status_new USING status::text::status_new;

ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'a'::status_new;

ALTER TABLE orders ALTER COLUMN history TYPE status_new[] USING history::text::status_new[];

ALTER TABLE orders ALTER COLUMN kind TYPE text USING kind::text;

ALTER TABLE orders ALTER COLUMN mood TYPE mood_new USING mood::mood_new;

ALTER TABLE orders ALTER COLUMN note SET DEFAULT 'soft';

ALTER TABLE shipments DROP CONSTRAINT grade_known;

ALTER TABLE shipments DROP CONSTRAINT status_not_b;

DROP INDEX IF EXISTS shipments_a_idx;

DROP INDEX IF EXISTS shipments_b_idx;

ALTER TABLE shipments ALTER COLUMN status TYPE status_new USING status::text::status_new;

ALTER TABLE shipments
ADD CONSTRAINT status_not_b CHECK (status <> 'b'::status_new) NOT VALID;

ALTER TABLE shipments VALIDATE CONSTRAINT status_not_b;

CREATE INDEX CONCURRENTLY IF NOT EXISTS shipments_a_idx ON shipments (id) WHERE (status = 'a'::status_new);

-- pgschema:wait
SELECT 
    COALESCE(i.indisvalid, false) as done,
    CASE 
        WHEN p.blocks_total > 0 THEN p.blocks_done * 100 / p.blocks_total
        ELSE 0
    END as progress
FROM pg_class c
JOIN pg_namespace n ON c.relnamespace = n.oid
LEFT JOIN pg_index i ON c.oid = i.indexrelid
LEFT JOIN pg_stat_progress_create_index p ON c.oid = p.index_relid
WHERE c.relname = 'shipments_a_idx'
  AND n.nspname = 'public';

DROP FUNCTION IF EXISTS default_priority();

CREATE OR REPLACE FUNCTION default_priority()
RETURNS priority_new
LANGUAGE sql
VOLATILE
AS $$ SELECT 'low'::priority_new
$$;

DROP TYPE IF EXISTS tone_old RESTRICT;

DROP TYPE IF EXISTS status_old RESTRICT;

DROP TYPE IF EXISTS scope_old RESTRICT;

DROP TYPE IF EXISTS priority_old RESTRICT;

DROP TYPE IF EXISTS owner_old RESTRICT;

DROP TYPE IF EXISTS mark_old RESTRICT;

DROP TYPE IF EXISTS level_old RESTRICT;

DROP TYPE IF EXISTS kind_old RESTRICT;

DROP TYPE IF EXISTS grade_old RESTRICT;

DROP TYPE IF EXISTS flag_old RESTRICT;
