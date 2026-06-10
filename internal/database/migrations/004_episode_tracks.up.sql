create table episode_tracks (
    id          uuid        primary key default gen_random_uuid(),
    episode_id  uuid        not null references episodes(id) on delete cascade,
    position    int         not null,
    title       text        not null,
    artist      text,
    album       text,
    started_at  integer,
    ended_at   integer,
    created_at  timestamptz not null default now(),
    unique (episode_id, position)
);

create index on episode_tracks (episode_id);
