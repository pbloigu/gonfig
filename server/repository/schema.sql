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

CREATE TABLE IF NOT EXISTS Measurement  (
    id integer primary key,
    application_id string,
    name string not null,
    created datetime default CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES Application(id)
    UNIQUE(application_id, name)
);

CREATE TABLE IF NOT EXISTS MeasurementValue (
    measurement_id integer not null,
    created datetime default CURRENT_TIMESTAMP,
    data text,
    FOREIGN KEY (measurement_id) REFERENCES Measurement(id)
);

CREATE TABLE IF NOT EXISTS User (
    login string PRIMARY KEY,
    password string
);
