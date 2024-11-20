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


--
-- Name: useractions; Type: TABLE; Schema: public;
--
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'action_type_enum') THEN
        CREATE TYPE action_type_enum AS ENUM ('like', 'follow');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'target_type_enum') THEN
        CREATE TYPE target_type_enum AS ENUM ('link', 'user');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS public.useractions (
    action_id SERIAL NOT NULL PRIMARY KEY,
    user_id integer NOT NULL,
    action_type action_type_enum NOT NULL,
    target_id integer NOT NULL,
    target_type target_type_enum NOT NULL,
    action_timestamp timestamp without time zone NOT NULL DEFAULT now(),
    active boolean DEFAULT TRUE,
    CONSTRAINT fk_user FOREIGN KEY (user_id)
        REFERENCES users(user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT useractions_unique_action
        UNIQUE (user_id, action_type, target_id, target_type)
);

---- create above / drop below ----

DROP TABLE public.useractions;
DROP TABLE public.comments;
DROP TABLE public.links;
DROP TABLE public.users;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
