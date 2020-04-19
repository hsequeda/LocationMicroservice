create table admin
(
    id           serial  not null
        constraint admin_pk
            primary key,
    username     varchar not null,
    passwordhash varchar not null
);

alter table admin
    owner to postgres;

create table "user"
(
    id           serial           not null
        constraint user_pk
            primary key,
    refreshtoken varchar          not null,
    latitude     double precision not null,
    longitude    double precision not null,
    h3index      bigint[]         not null,
    category     varchar          not null,
    admin_id     integer          not null
        constraint user_admin_id_fk
            references admin
            on delete cascade
);

alter table "user"
     owner to postgres;

create unique index user_id_uindex
    on "user" (id);

create unique index admin_id_uindex
    on admin (id);

create unique index admin_passwordhash_uindex
    on admin (passwordhash);

create unique index admin_username_uindex
    on admin (username);

insert into "admin"(username, passwordhash)
values ('root', '$2a$14$CvCHdfqfKx0jnWq9v2/Fb.OL/rd3ZaLmwPsTIIPtLQ7HtexIBuaH2');
