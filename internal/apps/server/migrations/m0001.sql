create table if not exists users
(
    id         uuid primary key,
    login      text      not null,
    pass_hash  text      not null,
    created_at timestamp not null,
    updated_at timestamp not null
);
alter table if exists users
    drop constraint if exists users_login_key;
alter table if exists users
    add constraint users_login_key unique (login);