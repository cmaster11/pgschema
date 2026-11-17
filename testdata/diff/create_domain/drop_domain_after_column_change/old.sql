CREATE FUNCTION is_short(v text) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT length(v) < 10 $$;

CREATE DOMAIN code_old AS text
    CHECK (VALUE ~ '^[a-z]+$');

CREATE DOMAIN label_old AS text
    CHECK (is_short(VALUE));

CREATE TYPE hue_old AS ENUM ('red', 'blue');

CREATE DOMAIN hue_base_old AS hue_old;

CREATE DOMAIN hue_tag_old AS hue_base_old;

CREATE DOMAIN tag_old AS text;

CREATE DOMAIN mark_old AS text;

CREATE TABLE items (
    id integer PRIMARY KEY,
    code code_old DEFAULT 'abc'::code_old,
    label label_old,
    hue hue_tag_old
);

CREATE TABLE notes (
    id integer PRIMARY KEY,
    note text
);

CREATE INDEX notes_tag_idx ON notes ((note::tag_old));

CREATE INDEX notes_marked_idx ON notes (id) WHERE note::mark_old = 'x'::mark_old;
