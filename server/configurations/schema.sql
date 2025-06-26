CREATE TABLE IF NOT EXISTS Application (
    id text primary key,
    name text NOT NULL,
    api_key text UNIQUE NOT NULL,
    hostname text,
    ip text
);

CREATE TABLE IF NOT EXISTS Configuration (
    application_id string,
    created datetime default CURRENT_TIMESTAMP,
    data text,
    is_latest integer,
    FOREIGN KEY (application_id) REFERENCES Application(id)
);

CREATE TABLE IF NOT EXISTS User (
    login string PRIMARY KEY,
    password string
);
