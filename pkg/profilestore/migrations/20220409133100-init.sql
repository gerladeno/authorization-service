-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up

CREATE TABLE user_model
(
    uuid     text NOT NULL
        CONSTRAINT uuid_pk PRIMARY KEY,
    phone    text NOT NULL,
    created  timestamp DEFAULT NOW(),
    updated  timestamp DEFAULT NOW()
);

CREATE INDEX phone_idx ON user_model (phone);

-- +migrate Down

DROP TABLE user_model CASCADE;