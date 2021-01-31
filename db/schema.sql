--
-- PostgreSQL database dump
--

-- Dumped from database version 12.5 (Debian 12.5-1.pgdg100+1)
-- Dumped by pg_dump version 12.5

-- Started on 2021-01-30 16:11:13 PST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 203 (class 1259 OID 16408)
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: goblog
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


ALTER TABLE public.goose_db_version OWNER TO goblog;

--
-- TOC entry 202 (class 1259 OID 16406)
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: goblog
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.goose_db_version_id_seq OWNER TO goblog;

--
-- TOC entry 2921 (class 0 OID 0)
-- Dependencies: 202
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: goblog
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- TOC entry 204 (class 1259 OID 16415)
-- Name: pages; Type: TABLE; Schema: public; Owner: goblog
--

CREATE TABLE public.pages (
    id uuid NOT NULL,
    title character varying,
    body text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.pages OWNER TO goblog;

--
-- TOC entry 2782 (class 2604 OID 16411)
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: goblog
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- TOC entry 2787 (class 2606 OID 16414)
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: goblog
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- TOC entry 2789 (class 2606 OID 16424)
-- Name: pages pages_pkey; Type: CONSTRAINT; Schema: public; Owner: goblog
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);


-- Completed on 2021-01-30 16:11:13 PST

--
-- PostgreSQL database dump complete
--

