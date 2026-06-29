create table stations (
    id                        uuid        primary key default gen_random_uuid(),
    name                      text        not null,
    slug                      text        not null unique,
    timezone                  text        not null default 'UTC',
    season_end_date           date,
    default_show_horizon_days int         not null default 90,
    logo_url                  text,
    tracklist_webhook_url     text,
    created_at                timestamptz not null default now(),
    updated_at                timestamptz not null default now()
);

create table api_keys (
    id         uuid        primary key default gen_random_uuid(),
    station_id uuid        not null references stations(id) on delete cascade,
    key_hash   text        not null unique,
    created_at timestamptz not null default now()
);

create table channels (
    id                   uuid        primary key default gen_random_uuid(),
    station_id           uuid        not null references stations(id) on delete cascade,
    name                 text        not null,
    slug                 text        not null unique,
    mount                text        not null default '/main',
    harbor_password_hash text,
    gap_fill_enabled     bool        not null default false,
    repeat_prefill_days  int         not null default 7,
    created_at           timestamptz not null default now(),
    updated_at           timestamptz not null default now()
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
    contact_email    text,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now()
);

create table ical_feeds (
    id             uuid        primary key default gen_random_uuid(),
    station_id     uuid        not null references stations(id) on delete cascade,
    channel_id     uuid        not null references channels(id) on delete cascade,
    type           text        not null default 'ical' check (type in ('ical', 'caldav')),
    url            text        not null,
    username       text,
    password       text,
    calendar_path  text,
    last_synced_at timestamptz,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now()
);

create table episodes (
    id             uuid        primary key default gen_random_uuid(),
    channel_id     uuid        not null references channels(id) on delete cascade,
    show_id        uuid        references shows(id) on delete set null,
    title          text        not null,
    description    text        not null default '',
    image_ref      text,
    color          text,
    start_time     timestamptz not null,
    end_time       timestamptz not null,
    type           text        not null check (type in ('live', 'recorded', 'external', 'playlist')),
    source_adapter text        not null,
    source_ref     text        not null,
    original_start timestamptz,
    ical_uid       text,
    ical_feed_id   uuid        references ical_feeds(id) on delete set null,
    auto_filled    bool        not null default false,
    repeat_of      uuid        references episodes(id) on delete set null,
    contact_email  text,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now(),
    constraint episodes_end_after_start check (end_time > start_time)
);

create table gap_fill_rules (
    id           uuid        primary key default gen_random_uuid(),
    channel_id   uuid        not null references channels(id) on delete cascade,
    priority     int         not null default 0,
    time_from    text,
    time_to      text,
    type         text        not null check (type in ('playlist', 'show-repeat')),
    source_ref   text        not null,
    created_at   timestamptz not null default now()
);

create table repeat_airings (
    id         uuid        primary key default gen_random_uuid(),
    episode_id uuid        not null references episodes(id) on delete cascade,
    channel_id uuid        not null references channels(id) on delete cascade,
    aired_at   timestamptz not null default now()
);

create table sent_reminders (
    episode_id uuid primary key references episodes(id) on delete cascade,
    sent_at    timestamptz not null default now()
);

create table email_templates (
    id         uuid        primary key default gen_random_uuid(),
    station_id uuid        not null references stations(id) on delete cascade,
    type       text        not null,
    subject    text        not null,
    body       text        not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (station_id, type)
);

create index on api_keys  (station_id);
create index on api_keys  (key_hash);
create index on channels  (station_id);
create index on shows     (channel_id);
create index on episodes  (channel_id);
create index on episodes  (show_id) where show_id is not null;
create index on episodes  (start_time);
create unique index on episodes (ical_feed_id, ical_uid) where ical_uid is not null;
create index on ical_feeds (channel_id);
create index on gap_fill_rules (channel_id, priority);
create index on repeat_airings (channel_id, aired_at);

create table tracklist_tokens (
    id         uuid        primary key default gen_random_uuid(),
    episode_id uuid        not null references episodes(id) on delete cascade,
    token_hash text        not null unique,
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);

create index on tracklist_tokens (token_hash);
create index on tracklist_tokens (episode_id);

create table sent_tracklist_emails (
    episode_id uuid primary key references episodes(id) on delete cascade,
    sent_at    timestamptz not null default now()
);
