CREATE DOMAIN code_new AS text
    CHECK (VALUE ~ '^[a-z0-9]+$');

CREATE TYPE hue_new AS ENUM ('red', 'blue', 'green');

CREATE TABLE items (
    id integer PRIMARY KEY,
    code code_new DEFAULT 'abc'::code_new,
    label text,
    hue hue_new
);

CREATE TABLE notes (
    id integer PRIMARY KEY,
    note text
);

CREATE INDEX notes_tag_idx ON notes (note);

CREATE INDEX notes_marked_idx ON notes (id) WHERE note = 'x';
