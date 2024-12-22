--
-- Name: users; Type: TABLE; Schema: public;
--

CREATE TABLE IF NOT EXISTS public.linkstyle (
    link_style_id SERIAL NOT NULL PRIMARY KEY,
    link_id integer NOT NULL,
    style_key character varying(30) NOT NULL,
    style_value character varying(30) NOT NULL,
    CONSTRAINT fk_link FOREIGN KEY (link_id)
        REFERENCES links(link_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT linkstyle_unique_style
        UNIQUE (link_id, style_key)
);

---- create above / drop below ----

DROP TABLE public.linkstyle;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
