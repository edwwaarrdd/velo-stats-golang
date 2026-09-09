CREATE TABLE IF NOT EXISTS weather_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    ride_id INTEGER NOT NULL,
    temperature_c FLOAT NOT NULL,
    apparent_temperature_c FLOAT NOT NULL,
    precipitation_mm FLOAT NOT NULL,
    rain_mm FLOAT NOT NULL,
    snowfall_cm FLOAT NOT NULL,
    cloud_cover_percent FLOAT NOT NULL,
    wind_speed_kmh FLOAT NOT NULL,
    wind_gusts_kmh FLOAT NOT NULL,
    wind_direction_degrees FLOAT NOT NULL,
    relative_humidity_percent FLOAT NOT NULL,
    weather_code INTEGER NOT NULL,
    observed_at DATETIME NOT NULL,
    created_at DATETIME,
    updated_at DATETIME,
    FOREIGN KEY (ride_id) REFERENCES rides (ride_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS weather_records_ride_id_unique ON weather_records (ride_id);
