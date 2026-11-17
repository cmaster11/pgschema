CREATE TYPE status_old AS ENUM ('a', 'b');

CREATE TABLE audit (
    id integer PRIMARY KEY,
    status text
);

CREATE PROCEDURE record_status(p_id integer)
    LANGUAGE sql
    BEGIN ATOMIC
        INSERT INTO audit (id, status) VALUES (p_id, ('a'::status_old)::text);
    END;
