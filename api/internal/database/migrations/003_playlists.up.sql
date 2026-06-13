create table playlists (
    id             uuid        primary key default gen_random_uuid(),
    station_id     uuid        not null references stations(id) on delete cascade,
    name           text        not null,
    shuffle        bool        not null default false,
    loop           bool        not null default true,
    -- when set, a background job resolves this external playlist via yt-dlp
    -- and populates playlist_items. null for locally managed playlists.
    source_adapter text,
    source_ref     text,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now(),
    check ((source_adapter is null) = (source_ref is null))
);

-- join table: which media belongs to which playlist and in what order.
-- track metadata lives on the media row; position is the only playlist-specific data.
create table playlist_items (
    id          uuid        primary key default gen_random_uuid(),
    playlist_id uuid        not null references playlists(id) on delete cascade,
    media_id    uuid        not null references media(id) on delete restrict,
    position    int         not null,
    created_at  timestamptz not null default now(),
    unique (playlist_id, position)
);

create index on playlists (station_id);
create index on playlist_items (playlist_id, position);
create index on playlist_items (media_id);
