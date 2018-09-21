/*
    sql query for creating the table which we are using in seed project
*/    
CREATE TABLE "users"
(
    id bigint NOT NULL DEFAULT id_generator(),
	name character varying(100) COLLATE pg_catalog."default",
    email character varying(200) COLLATE pg_catalog."default",
    password character varying(100) COLLATE pg_catalog."default",
	token character varying(100) COLLATE pg_catalog."default",
    is_active bit varying,
    CONSTRAINT user_ppkey PRIMARY KEY (id),
    CONSTRAINT uuniqueemail UNIQUE (email)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

/*
    sql query for 1 sample entry
*/
insert into users(Name,Email,Token,is_Active,Password) values ('test','test@test.com',null,1::bit,md5('test'))
