-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up

CREATE TABLE user_model
(
    id       text NOT NULL
        CONSTRAINT uuid_pk PRIMARY KEY,
    username text NOT NULL,
    phone    text NOT NULL,
    created  timestamp DEFAULT NOW(),
    updated  timestamp DEFAULT NOW()
);

CREATE INDEX phone_idx ON user_model (phone);

-- +migrate Down

DROP TABLE user_model CASCADE;