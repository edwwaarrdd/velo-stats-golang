CREATE TABLE IF NOT EXISTS rides (
    ride_id INTEGER NOT NULL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    status VARCHAR(32) NOT NULL,
    duration INTEGER NOT NULL,
    bike_number VARCHAR(32) NOT NULL,
    origin_station_code VARCHAR(32) NOT NULL,
    origin_station VARCHAR NOT NULL,
    origin_slot_id VARCHAR(16) NOT NULL,
    checkout_time DATETIME NOT NULL,
    destination_station_code VARCHAR(32) NOT NULL,
    destination_station VARCHAR NOT NULL,
    destination_slot_id VARCHAR(16) NOT NULL,
    checkin_time DATETIME NOT NULL,
    distance_checked_at DATETIME,
    weather_checked_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS rides_checkout_time_index ON rides (checkout_time);
