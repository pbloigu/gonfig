-- +goose Up
ALTER TABLE User RENAME TO User_Old;
CREATE TABLE User (
    login string PRIMARY KEY,
    password string NOT NULL
);

INSERT INTO User (login, password) SELECT login, password FROM User_Old;

-- +goose Down