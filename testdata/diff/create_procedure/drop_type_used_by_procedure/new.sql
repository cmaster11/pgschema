CREATE TYPE status_new AS ENUM ('a', 'b', 'c');

CREATE TABLE audit (
    id integer PRIMARY KEY,
    status text
);

CREATE PROCEDURE record_status(p_id integer)
    LANGUAGE sql
    BEGIN ATOMIC
        INSERT INTO audit (id, status) VALUES (p_id, ('a'::status_new)::text);
    END;
