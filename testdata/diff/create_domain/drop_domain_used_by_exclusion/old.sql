CREATE DOMAIN scope_old AS text;

CREATE TABLE notes (
    id integer PRIMARY KEY,
    note text,
    CONSTRAINT notes_own_excl EXCLUDE USING btree (id WITH =) WHERE (note::scope_old = 'own'::scope_old)
);
