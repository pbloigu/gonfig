-- +goose Up
INSERT INTO Series(id, application_id, name)
VALUES(1, 'app1', 'data1');

CREATE TABLE SeriesValue_1 (
id int NOT NULL AUTO_INCREMENT,
series_id int NOT NULL,
created INT(11) UNSIGNED default UNIX_TIMESTAMP(),
recorded int(11) UNSIGNED,
data varchar(128),
PRIMARY KEY (id),
FOREIGN KEY (series_id) REFERENCES Series(id));

INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 100, 'value1');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 200, 'value2');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 300, 'value3');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 400, 'value4');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 500, 'value5');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() +600, 'value6');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 700, 'value7');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 800, 'value8');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 900, 'value9');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 1000, 'value10');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 1100, 'value11');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 1200, 'value12');
INSERT INTO SeriesValue_1 (series_id, recorded, data)
VALUES(1, unix_timestamp() + 1300, 'value13');

-- +goose Down