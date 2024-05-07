CREATE TABLE IF NOT EXISTS traffic
(
    domain varchar not null,
    event_name varchar not null,
    duration bigint,
    timestamp  timestamp not null,
    user_agent varchar   not null,
    referrer   varchar,
    path       varchar   not null,
    visitor_id   varchar not null,
    query_params jsonb,
    country varchar not null,
    event_data jsonb,
    status_code int not null,
    ip inet not null
);


