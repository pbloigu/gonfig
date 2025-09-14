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

CREATE TABLE IF NOT EXISTS MeasurementTrigger (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id string NOT NULL,
    measurement_name string NOT NULL,
    UNIQUE(application_id, measurement_name),
    FOREIGN KEY (application_id) REFERENCES Application(id)
);

CREATE TABLE IF NOT EXISTS CronTrigger (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name string UNIQUE NOT NULL,
    expression string NOT NULL
);

CREATE TABLE IF NOT EXISTS StatusTrigger (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id string UNIQUE NOT NULL,
    FOREIGN KEY (application_id) REFERENCES Application(id)
);

CREATE TABLE IF NOT EXISTS Action (
    name string UNIQUE NOT NULL,
    script string,
    measurement_trigger_id integer,
    cron_trigger_id integer,
    status_change_trigger_id integer,
    FOREIGN KEY(measurement_trigger_id) REFERENCES MeasurementTrigger(id),
    FOREIGN KEY(cron_trigger_id) REFERENCES CronTrigger(id),
    FOREIGN KEY(status_change_trigger_id) REFERENCES StatusTrigger(id)
);




INSERT INTO User (login, password) VALUES ('admin', 'password') ON CONFLICT DO NOTHING;
INSERT INTO Application (id, name, api_key) VALUES ('app1', 'app1', 'app1') ON CONFLICT DO NOTHING;
INSERT INTO Configuration(application_id, is_latest, data) VALUES ('app1', 1, '') ON CONFLICT DO NOTHING;