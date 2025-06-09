CREATE SCHEMA IF NOT EXISTS weather;

CREATE TYPE weather.emails_frequency AS ENUM (
    'hourly',
    'daily'
);

CREATE TABLE IF NOT EXISTS weather.subscriptions (
    id bigserial PRIMARY KEY,
    email      character varying(255)             NOT NULL,
    city       character varying(255)             NOT NULL,
    frequency  weather.emails_frequency           NOT NULL,
    token      character varying(255)             NOT NULL UNIQUE,
    confirmed  boolean DEFAULT false              NOT NULL,
    subscribed boolean DEFAULT false              NOT NULL,

    UNIQUE(email, city, frequency)
);

CREATE INDEX IF NOT EXISTS "email_idx" ON weather.subscriptions("email");
CREATE INDEX IF NOT EXISTS "token_idx" ON weather.subscriptions("token");