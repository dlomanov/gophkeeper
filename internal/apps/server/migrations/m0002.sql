create table if not exists entries
(
    id         uuid primary key,
    user_id    uuid      not null references users,
    type       text      not null,
    meta       json,
    data       bytea     not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

create index if not exists entries_id_user_id_idx on entries (id, user_id);