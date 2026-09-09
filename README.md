# Velo Stats

Go service for tracking Velo Antwerp bike-share stations, ride history, routing and weather.

Ride history is loaded from a JSON export, station information from the public Velo Antwerp GBFS feed. Each ride is
then enriched in the background: the cycling distance between its two stations comes from the public OSRM routing
API, and the weather at its origin station and checkin time comes from the free Open-Meteo archive. The API serves
the combined data as JSON.

This is a port of the Laravel application of the same name, and serves byte-identical JSON on every endpoint.

## Setup

```
docker compose up -d --build
```

This starts five services:
- `app` – the API, served on port `8000`
- `redis` – queue backend
- `worker` – queue worker for the `default` queue
- `worker-ride-distance` – worker consuming the `ride_distance_checks` queue one job at a time, so calls to the free
  routing API are never made concurrently
- `worker-ride-weather` – worker consuming the `ride_weather_checks` queue one job at a time, so calls to the free
  Open-Meteo API are never made concurrently

Every container migrates the database on start, so no manual setup is needed.

Verify the app is up and running:

```
curl http://localhost:8000/_healthcheck
```

Should return a 200 OK response.

Stop everything with:

```
docker compose down
```

The SQLite database and the ride export are bind-mounted from `database/` and `data/`, so data survives a rebuild and
a new export can be dropped in without one. The binary itself lives in the image, so code changes need a rebuild:

```
docker compose up -d --build
```

### Loading data

A fresh database is empty. Populate it in this order:

```
docker compose exec app velo stations:load
docker compose exec app velo rides:load
docker compose exec app velo rides:check-distances
docker compose exec app velo rides:check-weather
```

The last two commands queue one job per ride and return immediately. The dedicated workers drain them one call at a
time, which takes a few minutes for a full ride history.

### Configuration

Environment variables (set in `docker-compose.yml`, and documented in `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `8000` | Port the API listens on |
| `DB_DATABASE` | `database/database.sqlite` | Path to the SQLite database |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Comma-separated origins allowed to call the API |
| `REDIS_HOST` | `127.0.0.1` | Redis host backing the queues |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | empty | Redis password, when one is set |
| `REDIS_DB` | `0` | Redis database index |
| `QUEUE_PREFIX` | `velo-stats:queues:` | Prefix for the queue keys in Redis |
| `RIDES_JSON_PATH` | `data/rides.json` | Path to the rides JSON export |

The three upstream endpoints can be overridden with `VELO_ANTWERP_STATION_INFORMATION_URL`, `OSRM_BASE_URL` and
`OPEN_METEO_ARCHIVE_URL`.

Outside Docker, a `.env` file in the working directory is read on start. Real environment variables win over it.

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/_healthcheck` | Returns `{"message": "ok"}` with a 200 status if the app is up |
| `GET` | `/rides` | Returns every ride with its basic info, distance (from the cached station route), speed (distance ÷ the exact time between check-out and check-in), the expected ride time from the cached route and how far the actual ride time was under or over it, and cached weather, most recent first |
| `GET` | `/rides/summary` | Returns aggregate stats across all rides: total rides, total/average/longest/shortest duration, and total/average distance |
| `GET` | `/rides/cost` | Returns the cost per ride, using the € 58/year subscription price prorated over the date range from the first to the last ride, plus the equivalent cost and money saved versus paying with day passes (€ 5) or week passes (€ 12) instead |
| `GET` | `/stations` | Returns every known station with its coordinates |

## Console Commands

Run against the running `app` container:

```
docker compose exec app velo <command>
```

| Command | Description |
|---|---|
| `serve` | Serve the JSON API |
| `work -queue=NAMES` | Run a queue worker, draining the comma-separated queues one job at a time |
| `migrate` | Create the database schema. Every command does this on start, so it is rarely needed on its own |
| `stations:load` | Fetches Velo Antwerp station information from the public GBFS feed and upserts it into the database |
| `rides:load [-path=PATH]` | Loads ride history from a JSON export (defaults to `data/rides.json`) and upserts it into the database |
| `tasks:dispatch-test [-message=MSG]` | Dispatches a test job that logs a message from the worker, useful for verifying the queue setup |
| `rides:check-distances` | Queues a job per unchecked ride to calculate and cache the distance between its origin and destination stations, one at a time via the `ride_distance_checks` queue |
| `rides:check-weather [-force]` | Queues a job per ride to fetch and cache the biking-relevant weather (temperature, precipitation, wind, cloud cover, humidity, weather code) at its origin station and checkin time from the free Open-Meteo API, one at a time via the `ride_weather_checks` queue. Only unchecked rides are queued by default; pass `-force` to re-fetch weather for every ride |

### Verifying the queue setup

Dispatch a test "hello world" job through the `app` container:

```
docker compose exec app velo tasks:dispatch-test -message "hello world"
```

Then check the `worker` container's logs to confirm the message was picked up and processed:

```
docker compose logs worker
```

You should see a log line containing `hello world` from the worker.

## Layout

```
cmd/velo             the single binary: the API, the workers and every command
internal/app         wiring, so each command gets the services it needs
internal/config      settings, read from the environment and an optional .env
internal/console     the commands
internal/database    the SQLite connection and the embedded migrations
internal/httpapi     routing, CORS and the endpoints
internal/jobs        the background work the queues carry
internal/queue       the Redis queue and the worker that drains it
internal/rides       ride history: model, storage, export, statistics
internal/routing     cycling routes between stations, cached per station pair
internal/stations    stations: model, storage and the GBFS feed
internal/support     rounding, JSON encoding and date formatting shared by all
internal/weather     the weather a ride was made in, cached per ride
```

Background work is pushed onto Redis lists and popped by a worker, one job at a time per worker. Jobs are run once
and are not retried; a failure is logged and recorded in the `failed_jobs` table.

## Tests

```
go test ./...
```

The suite runs against a temporary SQLite database and fakes every third-party call, so it never touches the network
or the development database. The endpoint tests assert the exact JSON bodies the Laravel application returns.
