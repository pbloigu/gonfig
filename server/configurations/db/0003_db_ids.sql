-- +goose Up
ALTER TABLE Action RENAME TO Action_Old;

CREATE TABLE Action (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description string NOT NULL,
    script string,
    series_trigger_id integer,
    cron_trigger_id integer,
    status_change_trigger_id integer,
    FOREIGN KEY(series_trigger_id) REFERENCES SeriesTrigger(id),
    FOREIGN KEY(cron_trigger_id) REFERENCES CronTrigger(id),
    FOREIGN KEY(status_change_trigger_id) REFERENCES StatusTrigger(id)
);

INSERT INTO Action (description, script, series_trigger_id, cron_trigger_id, status_change_trigger_id)
SELECT description, script, series_trigger_id, cron_trigger_id, status_change_trigger_id FROM Action_Old;

DROP TABLE Action_Old;

-- +goose Down