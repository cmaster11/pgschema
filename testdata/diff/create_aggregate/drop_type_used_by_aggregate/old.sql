CREATE TYPE tally_old AS ENUM ('a', 'b');

CREATE FUNCTION keep_state(s anyenum, v integer) RETURNS anyenum
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT s $$;

CREATE AGGREGATE tally(integer) (
    SFUNC = keep_state,
    STYPE = tally_old,
    INITCOND = 'a'
);
