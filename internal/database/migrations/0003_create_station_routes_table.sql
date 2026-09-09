CREATE TABLE IF NOT EXISTS station_routes (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    origin_station_id VARCHAR(32) NOT NULL,
    destination_station_id VARCHAR(32) NOT NULL,
    mode VARCHAR(8) NOT NULL,
    distance_meters FLOAT NOT NULL,
    duration_seconds FLOAT NOT NULL,
    created_at DATETIME,
    updated_at DATETIME,
    FOREIGN KEY (origin_station_id) REFERENCES stations (station_id) ON DELETE CASCADE,
    FOREIGN KEY (destination_station_id) REFERENCES stations (station_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_station_route_per_mode
    ON station_routes (origin_station_id, destination_station_id, mode);
