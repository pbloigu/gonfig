-- +goose Up
CREATE TABLE IF NOT EXISTS Series (
  id int NOT NULL AUTO_INCREMENT,
  application_id varchar(36) NOT NULL,
  name varchar(128) NOT NULL,
  created datetime DEFAULT current_timestamp(),
  PRIMARY KEY (id),
  UNIQUE KEY application_name (application_id,name)
);

-- +goose Down
