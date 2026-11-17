CREATE DOMAIN code_new AS text
  CONSTRAINT code_new_check CHECK (VALUE ~ '^[a-z0-9]+$');

CREATE TYPE hue_new AS ENUM (
    'red',
    'blue',
    'green'
);

ALTER TABLE items ALTER COLUMN code DROP DEFAULT;

ALTER TABLE items ALTER COLUMN code TYPE code_new USING code::code_new;

ALTER TABLE items ALTER COLUMN code SET DEFAULT ('abc')::code_new;

ALTER TABLE items ALTER COLUMN label TYPE text USING label::text;

ALTER TABLE items ALTER COLUMN hue TYPE hue_new USING hue::text::hue_new;

DROP INDEX IF EXISTS notes_marked_idx;

CREATE INDEX IF NOT EXISTS notes_marked_idx ON notes (id) WHERE (note = 'x'::text);

DROP INDEX IF EXISTS notes_tag_idx;

CREATE INDEX IF NOT EXISTS notes_tag_idx ON notes (note);

DROP DOMAIN IF EXISTS tag_old RESTRICT;

DROP DOMAIN IF EXISTS mark_old RESTRICT;

DROP DOMAIN IF EXISTS label_old RESTRICT;

DROP DOMAIN IF EXISTS hue_tag_old RESTRICT;

DROP DOMAIN IF EXISTS hue_base_old RESTRICT;

DROP TYPE IF EXISTS hue_old RESTRICT;

DROP DOMAIN IF EXISTS code_old RESTRICT;

DROP FUNCTION IF EXISTS is_short(text);
