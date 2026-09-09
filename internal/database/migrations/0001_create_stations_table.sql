CREATE TABLE IF NOT EXISTS stations (
    station_id VARCHAR(32) NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    short_name VARCHAR(32) NOT NULL,
    lat FLOAT NOT NULL,
    lon FLOAT NOT NULL,
    address VARCHAR NOT NULL,
    post_code VARCHAR(16) NOT NULL,
    rental_methods TEXT NOT NULL,
    capacity INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME
);
