create table if not exists entries
(
    id         uuid primary key,
    user_id    uuid      not null references users,
    key        text      not null,
    type       text      not null,
    meta       json,
    data       bytea     not null,
    version    int8      not null default 0,
    created_at timestamp not null,
    updated_at timestamp not null
);

create index if not exists entries_id_user_id_idx on entries (id, user_id);

alter table if exists entries
    drop constraint if exists entries_key_unique;
alter table if exists entries
    add constraint entries_key_unique unique (key, user_id);