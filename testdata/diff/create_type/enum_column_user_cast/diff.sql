ALTER TABLE issues ALTER COLUMN p TYPE lvl USING p::lvl;

ALTER TABLE tasks ALTER COLUMN p TYPE integer USING p::integer;
