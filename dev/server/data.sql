CREATE TABLE IF NOT EXISTS Measurement (
  id int NOT NULL AUTO_INCREMENT,
  application_id varchar(36) NOT NULL,
  name varchar(128) NOT NULL,
  created datetime DEFAULT current_timestamp(),
  PRIMARY KEY (id),
  UNIQUE KEY application_name (application_id,name)
);

INSERT INTO Measurement(id, application_id, name)
VALUES(1, 'app1', 'data1');

CREATE TABLE MeasurementValue_1 (
id int NOT NULL AUTO_INCREMENT,
measurement_id int NOT NULL,
created INT(11) UNSIGNED default UNIX_TIMESTAMP(),
recorded int(11) UNSIGNED,
data varchar(128),
PRIMARY KEY (id),
FOREIGN KEY (measurement_id) REFERENCES Measurement(id));

INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value1');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value2');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value3');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value4');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value5');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value6');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value7');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value8');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value9');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value10');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value11');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value12');
INSERT INTO MeasurementValue_1 (measurement_id, recorded, data)
VALUES(1, unix_timestamp(), 'value13');