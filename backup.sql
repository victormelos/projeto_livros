--
-- PostgreSQL database dump
--

-- Dumped from database version 17.3
-- Dumped by pg_dump version 17.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
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
-- Name: livros; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.livros (
    id character varying(27) NOT NULL,
    name character varying(255) NOT NULL,
    quantity integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.livros OWNER TO postgres;

--
-- Data for Name: livros; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.livros (id, name, quantity, created_at, updated_at) FROM stdin;
2tLuBM4VQS7qBPsV22PuKE4COFz	Senhor do anéis	88	2025-02-21 13:12:19.29964+00	2025-02-21 13:12:19.29964+00
2tLuBSfBWXQdd2XWVrqQfKA4BUK	Harry Potter	50	2025-02-21 13:12:19.301916+00	2025-02-21 13:12:19.301916+00
2tLucwWdufnYxOjfDIOCSIKIi3f	Duna	65	2025-02-21 13:15:58.28104+00	2025-02-21 13:15:58.28104+00
2tLucxNH3UgfZQs3usX5F5mlmGk	1984	90	2025-02-21 13:15:58.284338+00	2025-02-21 13:15:58.284338+00
2tLucwKvfU3lR4Ze0PQ6dmYIxru	O Sol é para Todos	45	2025-02-21 13:15:58.285823+00	2025-02-21 13:15:58.285823+00
2tLucyx65Dn0ksYGTUstpTx5TwN	Dom Casmurro	110	2025-02-21 13:15:58.287314+00	2025-02-21 13:15:58.287314+00
2tLuctlf7NKVLO01jkDvBpc2QfT	A Metamorfose	30	2025-02-21 13:15:58.288787+00	2025-02-21 13:15:58.288787+00
2tLucwxBN9idNgLpAv97IFx43pc	Cem Anos de Solidão	85	2025-02-21 13:15:58.290212+00	2025-02-21 13:15:58.290212+00
2tLuczLJwDpK1x1pxPp1Ex2oMKq	O Pequeno Príncipe	150	2025-02-21 13:15:58.291567+00	2025-02-21 13:15:58.291567+00
2tLucwTzHpWPm8BD5YZ2JdTe2Kj	Os Miseráveis	60	2025-02-21 13:15:58.293082+00	2025-02-21 13:15:58.293082+00
2tLucziS5TVF89yPRzZ2F4DMHvY	A Revolução dos Bichos	70	2025-02-21 13:15:58.294598+00	2025-02-21 13:15:58.294598+00
2tLucylXKW59sLh4Vbop14luxuu	O Nome do Vento	95	2025-02-21 13:15:58.296139+00	2025-02-21 13:15:58.296139+00
2tLucyN7JkmCLFEjem5XzvtSyJd	Game of Thrones	130	2025-02-21 13:15:58.297447+00	2025-02-21 13:15:58.297447+00
2tLucx18neoo9FywNCTB8TznXET	As Crônicas de Gelo e Fogo	55	2025-02-21 13:15:58.298924+00	2025-02-21 13:15:58.298924+00
2tLucygYTZeAxTxVbqlwIS9Heos	Percy Jackson	140	2025-02-21 13:15:58.300214+00	2025-02-21 13:15:58.300214+00
2tLucs8nLpgJhGiD9nGxxrVgMA0	A Culpa é das Estrelas	25	2025-02-21 13:15:58.301727+00	2025-02-21 13:15:58.301727+00
2tLuctu1dF4MH7NHk1R0qoVQaen	O Alquimista	175	2025-02-21 13:15:58.303359+00	2025-02-21 13:15:58.303359+00
2tLucxpbHYxj6IUWoAlqxM7Wce3	Cidade dos Ossos	80	2025-02-21 13:15:58.305076+00	2025-02-21 13:15:58.305076+00
2tLucxpff2KsfFcQpUdAwJGOJbj	A Sombra do Vento	67	2025-02-21 13:15:58.306439+00	2025-02-21 13:15:58.306439+00
2tLucu6UfUUFGdXqqdaFqDqlN72	O Código Da Vinci	105	2025-02-21 13:15:58.308134+00	2025-02-21 13:15:58.308134+00
2tLucwE2XelbZknEo8CLmcm3zoq	O Silêncio dos Inocentes	40	2025-02-21 13:15:58.309465+00	2025-02-21 13:15:58.309465+00
2tLucwfCqrtgjAkp68gs6XoGmJe	Memórias Póstumas de Brás Cubas	55	2025-02-21 13:15:58.310872+00	2025-02-21 13:15:58.310872+00
2tLucyDer5l2ss3Rtxc30owdERO	O Médico e o Monstro	33	2025-02-21 13:15:58.312149+00	2025-02-21 13:15:58.312149+00
2tLucwYMLLrqkfihp9ndtgYb9Vq	Drácula	78	2025-02-21 13:15:58.313391+00	2025-02-21 13:15:58.313391+00
2tLucsQvPAWvJsr7jkXGQqDyXAB	Frankenstein	62	2025-02-21 13:15:58.314941+00	2025-02-21 13:15:58.314941+00
2tLuctpFSLCVUzqmPjtEh8x6NGu	O Retrato de Dorian Gray	91	2025-02-21 13:15:58.316275+00	2025-02-21 13:15:58.316275+00
2tLucuNPXpu3g1LoQqS1Ad8LPbL	A Odisséia	115	2025-02-21 13:15:58.317709+00	2025-02-21 13:15:58.317709+00
2tLucy37UgTOsUoXGIB7hAbfpvM	A Ilíada	73	2025-02-21 13:15:58.31895+00	2025-02-21 13:15:58.31895+00
2tLuczeXBMDA6Eq27YM5PJFeRAB	O Senhor das Moscas	58	2025-02-21 13:15:58.320383+00	2025-02-21 13:15:58.320383+00
2tLucv2qiru6XcPYm7QQ4CecCL9	Admirável Mundo Novo	82	2025-02-21 13:15:58.32177+00	2025-02-21 13:15:58.32177+00
2tLuctt5AjjkwOiVTnb9FPgCXxq	O Apanhador no Campo de Centeio	47	2025-02-21 13:15:58.323239+00	2025-02-21 13:15:58.323239+00
2tLuczo0iACnHI8vsGLg3Dqc2UB	A Menina que Roubava Livros	122	2025-02-21 13:15:58.32467+00	2025-02-21 13:15:58.32467+00
2tLucsBMuXN3gfJav4CMRWHVNMG	O Diário de Anne Frank	99	2025-02-21 13:15:58.325961+00	2025-02-21 13:15:58.325961+00
2tLuculyh6RHec0H9wZEtaxVCuY	A Casa dos Espíritos	36	2025-02-21 13:15:58.327307+00	2025-02-21 13:15:58.327307+00
2tLucu5TYB6IAnjXLqD2jolf9Aq	O Velho e o Mar	27	2025-02-21 13:15:58.328762+00	2025-02-21 13:15:58.328762+00
2tLucv3FE4ydNUOTtjFUrt6qtq7	Moby Dick	64	2025-02-21 13:15:58.330113+00	2025-02-21 13:15:58.330113+00
2tLuctCuqcur4OhWXCihKUuRuBO	Os Irmãos Karamazov	108	2025-02-21 13:15:58.331366+00	2025-02-21 13:15:58.331366+00
2tLucxR5QqjwxmHPEhnChdxo46j	Crime e Castigo	53	2025-02-21 13:15:58.332701+00	2025-02-21 13:15:58.332701+00
2tLucuGdPm9Bc0vQhJ2b1ZYAo9T	Anna Karenina	87	2025-02-21 13:15:58.335442+00	2025-02-21 13:15:58.335442+00
2tLucuvBWqUh8AZbcuNPIOCKZUt	Guerra e Paz	41	2025-02-21 13:15:58.336705+00	2025-02-21 13:15:58.336705+00
2tLucsfn1YGQBV639ayA1z1N0ry	O Conto da Aia	134	2025-02-21 13:15:58.337843+00	2025-02-21 13:15:58.337843+00
2tLucu51yCUU8aAT92lJdSELhXe	As Mil e Uma Noites	166	2025-02-21 13:15:58.339376+00	2025-02-21 13:15:58.339376+00
2tLucvy89t9HYtGl68osXBX7Q0s	O Príncipe	22	2025-02-21 13:15:58.340805+00	2025-02-21 13:15:58.340805+00
2tLucv9sANodqZxvHiKQag8jils	O Segundo Sexo	31	2025-02-21 13:15:58.342149+00	2025-02-21 13:15:58.342149+00
2tLuctgBPKIZY2wkhR52aPufav1	A Divina Comédia	143	2025-02-21 13:15:58.343281+00	2025-02-21 13:15:58.343281+00
2tLucvZqeDpPYCISxsDlIOn14Tz	Paradiso Perdido	59	2025-02-21 13:15:58.344649+00	2025-02-21 13:15:58.344649+00
2tLucwkvKva2klKD76epdtDRgCk	O Mito de Sísifo	76	2025-02-21 13:15:58.346138+00	2025-02-21 13:15:58.346138+00
2tLucvNZ93UNHdFthwS37imXDou	O Estrangeiro	94	2025-02-21 13:15:58.347714+00	2025-02-21 13:15:58.347714+00
\.


--
-- Name: livros livros_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.livros
    ADD CONSTRAINT livros_pkey PRIMARY KEY (id);


--
-- Name: idx_livros_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_livros_name ON public.livros USING btree (name);


--
-- PostgreSQL database dump complete
--

