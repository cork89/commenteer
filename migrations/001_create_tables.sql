--
-- Name: users; Type: TABLE; Schema: public;
--

CREATE TABLE IF NOT EXISTS public.users (
    user_id SERIAL NOT NULL PRIMARY KEY,
    subscription_dt_tm timestamp without time zone,
    username character varying(80) NOT NULL,
    subscribed boolean DEFAULT false NOT NULL,
    refresh_token character varying(200),
    refresh_expire_dt_tm timestamp without time zone,
    icon_url character varying(220),
    access_token text,
    remaining_uploads integer NOT NULL DEFAULT 10,
    upload_refresh_dt_tm timestamp without time zone NOT NULL DEFAULT (now() + '7 days'::interval)
);

--
-- Name: links; Type: TABLE; Schema: public;
--

CREATE TABLE IF NOT EXISTS public.links (
    link_id SERIAL NOT NULL PRIMARY KEY,
    image_url character varying(120),
    proxy_url character varying(200),
    created_date timestamp without time zone DEFAULT now() NOT NULL,
    query_id character varying(80),
    cdn_image_url character varying(200) NOT NULL,
    user_id integer NOT NULL DEFAULT 0,
    CONSTRAINT fk_user FOREIGN KEY (user_id)
        REFERENCES users(user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);


--
-- Name: comments; Type: TABLE; Schema: public;
--
CREATE TABLE IF NOT EXISTS public.comments (
    comment_id SERIAL NOT NULL PRIMARY KEY,
    link_id integer NOT NULL,
    comment text,
    author character varying(100),
    CONSTRAINT fk_link FOREIGN KEY (link_id)
        REFERENCES links(link_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
