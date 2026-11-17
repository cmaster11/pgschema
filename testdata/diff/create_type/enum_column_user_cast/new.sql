CREATE TYPE pri AS ENUM ('low', 'high');

CREATE FUNCTION pri_rank(v pri) RETURNS integer
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT CASE v WHEN 'low' THEN 1 ELSE 2 END $$;

CREATE CAST (pri AS integer) WITH FUNCTION pri_rank(pri);

CREATE TYPE lvl AS ENUM ('minor', 'major');

CREATE FUNCTION pri_to_lvl(v pri) RETURNS lvl
    LANGUAGE sql IMMUTABLE
    AS $$ SELECT CASE v WHEN 'low' THEN 'minor'::lvl ELSE 'major'::lvl END $$;

CREATE CAST (pri AS lvl) WITH FUNCTION pri_to_lvl(pri);

CREATE TABLE tasks (
    id integer PRIMARY KEY,
    p integer
);

CREATE TABLE issues (
    id integer PRIMARY KEY,
    p lvl
);
