create table media (
    id               uuid        primary key default gen_random_uuid(),
    station_id       uuid        not null references stations(id) on delete cascade,
    title            text        not null,
    artist           text,
    album            text,
    genre            text,
    year             int,
    duration         interval,
    bpm              numeric(5,2),
    isrc             text,
    artwork_ref      text,
    file_format      text        check (file_format in ('mp3', 'aac', 'm4a')),
    file_size_bytes  bigint,
    -- original source (soundcloud url, local path, etc.)
    source_adapter   text        not null,
    source_ref       text        not null,
    -- where the file lives after download; null until ready
    local_ref        text,
    download_status  text        not null default 'not_required'
        check (download_status in ('not_required', 'pending', 'downloading', 'ready', 'failed')),
    download_error   text,
    downloaded_at    timestamptz,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now()
);

create index on media (station_id);
create index on media (download_status) where download_status in ('pending', 'failed');
