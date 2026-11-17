ALTER TABLE notes DROP CONSTRAINT notes_own_excl;

ALTER TABLE notes
ADD CONSTRAINT notes_own_excl EXCLUDE USING btree (id WITH =) WHERE (note = 'own'::text);

DROP DOMAIN IF EXISTS scope_old RESTRICT;
