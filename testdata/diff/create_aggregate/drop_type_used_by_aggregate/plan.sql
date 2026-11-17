CREATE TYPE tally_new AS ENUM (
    'a',
    'b'
);

DROP AGGREGATE IF EXISTS tally(integer);

CREATE AGGREGATE tally(integer) (
    SFUNC = keep_state,
    STYPE = tally_new,
    INITCOND = 'a'
);

DROP TYPE IF EXISTS tally_old RESTRICT;
