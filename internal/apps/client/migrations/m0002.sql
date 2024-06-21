create table if not exists entries
(
    id             text primary key,
    key            text      not null,
    type           text      not null,
    meta           text,
    data           blob      not null,
    global_version int8      not null default 0,
    version        int8      not null default 0,
    created_at     text not null,
    updated_at     text not null,
    unique (key)
);

create table if not exists entries_sync
(
    id text primary key,
    created_at text not null
);