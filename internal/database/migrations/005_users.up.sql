create table users (
    id            uuid        primary key default gen_random_uuid(),
    station_id    uuid        not null references stations(id) on delete cascade,
    email         text        not null unique,
    password_hash text        not null,
    role          text        not null default 'admin' check (role in ('admin', 'dj')),
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now()
);

create index on users (station_id);
create index on users (email);
