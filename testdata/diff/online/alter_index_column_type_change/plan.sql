ALTER TABLE events ALTER COLUMN account_id TYPE bigint USING account_id::bigint;

CREATE INDEX CONCURRENTLY IF NOT EXISTS events_account_idx_pgschema_new ON events (account_id, created_at);

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
WHERE c.relname = 'events_account_idx_pgschema_new'
  AND n.nspname = 'public';

DROP INDEX IF EXISTS events_account_idx;

ALTER INDEX events_account_idx_pgschema_new RENAME TO events_account_idx;

ALTER TABLE ledgers ALTER COLUMN ref TYPE bigint USING ref::bigint;

ALTER TABLE ledgers DROP CONSTRAINT ledgers_ref_check;

ALTER TABLE ledgers
ADD CONSTRAINT ledgers_ref_check CHECK (ref > 10) NOT VALID;

ALTER TABLE ledgers VALIDATE CONSTRAINT ledgers_ref_check;

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS ledgers_ref_key_pgschema_new ON ledgers (ref, id);

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
WHERE c.relname = 'ledgers_ref_key_pgschema_new'
  AND n.nspname = 'public';

DROP INDEX IF EXISTS ledgers_ref_key;

ALTER INDEX ledgers_ref_key_pgschema_new RENAME TO ledgers_ref_key;

ALTER TABLE tasks ALTER COLUMN c TYPE text USING c::text;

CREATE INDEX CONCURRENTLY IF NOT EXISTS tasks_c_idx_pgschema_new ON tasks (c, id) WHERE (d <> 'x'::phase);

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
WHERE c.relname = 'tasks_c_idx_pgschema_new'
  AND n.nspname = 'public';

DROP INDEX IF EXISTS tasks_c_idx;

ALTER INDEX tasks_c_idx_pgschema_new RENAME TO tasks_c_idx;

ALTER TABLE tickets ALTER COLUMN state TYPE text USING state::text;

ALTER TABLE tickets DROP CONSTRAINT tickets_state_check;

ALTER TABLE tickets
ADD CONSTRAINT tickets_state_check CHECK (state <> ''::text AND n > 0) NOT VALID;

ALTER TABLE tickets VALIDATE CONSTRAINT tickets_state_check;

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS tickets_state_n_key_pgschema_new ON tickets (state, n, id);

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
WHERE c.relname = 'tickets_state_n_key_pgschema_new'
  AND n.nspname = 'public';

DROP INDEX IF EXISTS tickets_state_n_key;

ALTER INDEX tickets_state_n_key_pgschema_new RENAME TO tickets_state_n_key;

ALTER TABLE users ALTER COLUMN code TYPE varchar(20);

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS users_code_key_pgschema_new ON users (code) WHERE ((code)::text <> ''::text);

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
WHERE c.relname = 'users_code_key_pgschema_new'
  AND n.nspname = 'public';

DROP INDEX IF EXISTS users_code_key;

ALTER INDEX users_code_key_pgschema_new RENAME TO users_code_key;
