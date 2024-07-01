BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS public.short_link (
	id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
	original_url varchar NOT NULL,
	short_url varchar NOT NULL,
	user_id varchar NOT NULL,
	is_deleted bool DEFAULT false NOT NULL,
	CONSTRAINT short_url_pk PRIMARY KEY (short_url)
);
CREATE UNIQUE INDEX IF NOT EXISTS short_link_short_url_idx ON public.short_link (short_url);

COMMIT;