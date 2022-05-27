-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up

create table user_model
(
    uuid    text not null
        constraint uuid_pk
            primary key,
    phone   text not null
        constraint unique_phone
            unique,
    created timestamp default now(),
    updated timestamp default now()
);

create index phone_idx
    on user_model (phone);
CREATE INDEX phone_idx ON user_model (phone);

-- +migrate Down

DROP TABLE user_model CASCADE;