-- ⊹ ࣪ ˖ aggregate listener stats — no IPs stored, ever
create table listener_stats (
    id           uuid        primary key default gen_random_uuid(),
    channel_id   uuid        not null references channels(id) on delete cascade,
    episode_id   uuid        references episodes(id) on delete set null,
    hour         timestamptz not null,
    country_code text        not null default 'XX',
    listeners    int         not null default 0,
    peak         int         not null default 0,
    samples      int         not null default 0,
    created_at   timestamptz not null default now(),
    constraint listener_stats_unique unique (channel_id, hour, country_code)
);

create index on listener_stats (channel_id, hour);
create index on listener_stats (episode_id) where episode_id is not null;
