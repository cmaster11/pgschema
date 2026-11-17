CREATE TYPE tally_new AS ENUM ('a', 'b');

CREATE FUNCTION keep_state(s anyenum, v integer) RETURNS anyenum
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT s $$;

CREATE AGGREGATE tally(integer) (
    SFUNC = keep_state,
    STYPE = tally_new,
    INITCOND = 'a'
);
