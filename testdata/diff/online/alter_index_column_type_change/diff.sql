ALTER TABLE events ALTER COLUMN account_id TYPE bigint USING account_id::bigint;

DROP INDEX IF EXISTS events_account_idx;

CREATE INDEX IF NOT EXISTS events_account_idx ON events (account_id, created_at);

ALTER TABLE ledgers ALTER COLUMN ref TYPE bigint USING ref::bigint;

ALTER TABLE ledgers DROP CONSTRAINT ledgers_ref_check;

ALTER TABLE ledgers
ADD CONSTRAINT ledgers_ref_check CHECK (ref > 10);

DROP INDEX IF EXISTS ledgers_ref_key;

CREATE UNIQUE INDEX IF NOT EXISTS ledgers_ref_key ON ledgers (ref, id);

ALTER TABLE tasks ALTER COLUMN c TYPE text USING c::text;

DROP INDEX IF EXISTS tasks_c_idx;

CREATE INDEX IF NOT EXISTS tasks_c_idx ON tasks (c, id) WHERE (d <> 'x'::phase);

ALTER TABLE tickets ALTER COLUMN state TYPE text USING state::text;

ALTER TABLE tickets DROP CONSTRAINT tickets_state_check;

ALTER TABLE tickets
ADD CONSTRAINT tickets_state_check CHECK (state <> ''::text AND n > 0);

DROP INDEX IF EXISTS tickets_state_n_key;

CREATE UNIQUE INDEX IF NOT EXISTS tickets_state_n_key ON tickets (state, n, id);

ALTER TABLE users ALTER COLUMN code TYPE varchar(20);

DROP INDEX IF EXISTS users_code_key;

CREATE UNIQUE INDEX IF NOT EXISTS users_code_key ON users (code) WHERE ((code)::text <> ''::text);
