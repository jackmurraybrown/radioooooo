-- ✮⋆‧° shows — recurring programme identity, owns the rrule and brand
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

create index on shows (channel_id);

-- ⋆˙⟡ link episodes to their parent show
alter table episodes add column show_id uuid references shows(id) on delete set null;
alter table episodes add column original_start timestamptz;
