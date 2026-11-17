CREATE TABLE public.events (
    id integer NOT NULL,
    account_id integer NOT NULL,
    created_at timestamp with time zone
);

CREATE INDEX events_account_idx ON public.events (account_id);

CREATE TABLE public.users (
    id integer NOT NULL,
    code varchar(10) NOT NULL
);

CREATE UNIQUE INDEX users_code_key ON public.users (code);

CREATE TABLE public.ledgers (
    id integer NOT NULL,
    ref integer NOT NULL,
    CONSTRAINT ledgers_ref_check CHECK (ref > 0)
);

CREATE UNIQUE INDEX ledgers_ref_key ON public.ledgers (ref);

CREATE TYPE public.ticket_state AS ENUM ('open', 'closed');

CREATE TABLE public.tickets (
    id integer NOT NULL,
    state public.ticket_state NOT NULL,
    n integer NOT NULL,
    CONSTRAINT tickets_state_check CHECK (state IS NOT NULL AND n > 0)
);

CREATE UNIQUE INDEX tickets_state_n_key ON public.tickets (state, n);

CREATE TYPE public.phase AS ENUM ('x', 'y');

CREATE TABLE public.tasks (
    id integer NOT NULL,
    c public.phase,
    d public.phase
);

CREATE INDEX tasks_c_idx ON public.tasks (c) WHERE d <> 'x'::public.phase;
