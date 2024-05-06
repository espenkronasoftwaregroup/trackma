CREATE TABLE IF NOT EXISTS traffic
(
    domain varchar not null,
    event_name varchar not null,
    duration bigint,
    timestamp  timestamp not null,
    user_agent varchar   not null,
    referrer   varchar,
    path       varchar   not null,
    group_id   varchar,
    query_params jsonb,
    country varchar not null
);
