CREATE TABLE public.events (
	id varchar(100) NOT NULL,
	name varchar(100) NOT NULL,
	status varchar(10) NOT NULL,
	create_date timestamptz NOT NULL,
	update_date timestamptz NOT NULL,
	delete_date timestamptz NULL,
	CONSTRAINT events_pk PRIMARY KEY (id)
);

