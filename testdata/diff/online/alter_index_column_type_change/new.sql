CREATE TABLE public.events (
    id integer NOT NULL,
    account_id bigint NOT NULL,
    created_at timestamp with time zone
);

CREATE INDEX events_account_idx ON public.events (account_id, created_at);

CREATE TABLE public.users (
    id integer NOT NULL,
    code varchar(20) NOT NULL
);

CREATE UNIQUE INDEX users_code_key ON public.users (code) WHERE code <> '';

CREATE TABLE public.ledgers (
    id integer NOT NULL,
    ref bigint NOT NULL,
    CONSTRAINT ledgers_ref_check CHECK (ref > 10)
);

CREATE UNIQUE INDEX ledgers_ref_key ON public.ledgers (ref, id);

CREATE TYPE public.ticket_state AS ENUM ('open', 'closed');

CREATE TABLE public.tickets (
    id integer NOT NULL,
    state text NOT NULL,
    n integer NOT NULL,
    CONSTRAINT tickets_state_check CHECK (state <> '' AND n > 0)
);

CREATE UNIQUE INDEX tickets_state_n_key ON public.tickets (state, n, id);

CREATE TYPE public.phase AS ENUM ('x', 'y');

CREATE TABLE public.tasks (
    id integer NOT NULL,
    c text,
    d public.phase
);

CREATE INDEX tasks_c_idx ON public.tasks (c, id) WHERE d <> 'x'::public.phase;
