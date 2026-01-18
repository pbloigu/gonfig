-- +goose Up
INSERT INTO User (login, password) VALUES ('admin', 'password') ON CONFLICT DO NOTHING;
INSERT INTO Application (id, name, api_key) VALUES ('app1', 'app1', 'app1') ON CONFLICT DO NOTHING;
INSERT INTO Configuration(application_id, is_latest, data) VALUES ('app1', 1, '') ON CONFLICT DO NOTHING;
-- +goose Down