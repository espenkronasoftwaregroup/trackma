CREATE TABLE IF NOT EXISTS events
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
    status_code int not null
);

CREATE INDEX IF NOT EXISTS events_visitor_id_index
    ON events (visitor_id);

CREATE INDEX IF NOT EXISTS events_event_name_index
    ON events (event_name);

CREATE TABLE IF NOT EXISTS monthly_traffic
(
    domain varchar not null,
    duration bigint,
    timestamp  timestamp not null,
    user_agent varchar   not null,
    referrer   varchar,
    path       varchar   not null,
    query_params jsonb,
    country varchar not null,
    status_code int not null,
    ip inet not null,
    ips inet[]
);
