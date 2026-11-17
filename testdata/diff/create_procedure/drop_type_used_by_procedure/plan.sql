CREATE TYPE status_new AS ENUM (
    'a',
    'b',
    'c'
);

DROP PROCEDURE IF EXISTS record_status(IN p_id integer);

CREATE OR REPLACE PROCEDURE record_status(
    IN p_id integer
)
LANGUAGE sql
BEGIN ATOMIC
 INSERT INTO public.audit (id, status)
   VALUES (record_status.p_id, ('a'::public.status_new)::text);
END;

DROP TYPE IF EXISTS status_old RESTRICT;
