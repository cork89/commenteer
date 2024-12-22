ALTER TABLE links ALTER proxy_url TYPE character varying(400);

---- create above / drop below ----

ALTER TABLE links ALTER proxy_url TYPE character varying(200);


-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
