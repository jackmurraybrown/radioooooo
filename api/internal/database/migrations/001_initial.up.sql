create table stations (
    id         uuid        primary key default gen_random_uuid(),
    name       text        not null,
    slug       text        not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table api_keys (
    id         uuid        primary key default gen_random_uuid(),
    station_id uuid        not null references stations(id) on delete cascade,
    key_hash   text        not null unique,
    created_at timestamptz not null default now()
);

create table channels (
    id         uuid        primary key default gen_random_uuid(),
    station_id uuid        not null references stations(id) on delete cascade,
    name       text        not null,
    slug       text        not null unique,
    mount      text        not null default '/main',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table shows (
    id               uuid        primary key default gen_random_uuid(),
    channel_id       uuid        not null references channels(id) on delete cascade,
    title            text        not null,
    description      text        not null default '',
    image_ref        text,
    recurrence_rule  text        not null,
    duration_minutes int         not null,
    type             text        not null check (type in ('live', 'recorded', 'external', 'playlist')),
    source_adapter   text        not null,
    source_ref       text        not null,
    allow_repeat     bool        not null default false,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now()
);

create table episodes (
    id             uuid        primary key default gen_random_uuid(),
    channel_id     uuid        not null references channels(id) on delete cascade,
    show_id        uuid        references shows(id) on delete set null,
    title          text        not null,
    description    text        not null default '',
    image_ref      text,
    start_time     timestamptz not null,
    end_time       timestamptz not null,
    type           text        not null check (type in ('live', 'recorded', 'external', 'playlist')),
    source_adapter text        not null,
    source_ref     text        not null,
    original_start timestamptz,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now(),
    constraint episodes_end_after_start check (end_time > start_time)
);

create index on api_keys  (station_id);
create index on api_keys  (key_hash);
create index on channels  (station_id);
create index on shows     (channel_id);
create index on episodes  (channel_id);
create index on episodes  (show_id) where show_id is not null;
create index on episodes  (start_time);
