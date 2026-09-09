CREATE TABLE IF NOT EXISTS failed_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    uuid VARCHAR NOT NULL,
    queue VARCHAR NOT NULL,
    job_type VARCHAR NOT NULL,
    payload TEXT NOT NULL,
    exception TEXT NOT NULL,
    failed_at DATETIME NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS failed_jobs_uuid_unique ON failed_jobs (uuid);
CREATE INDEX IF NOT EXISTS failed_jobs_queue_failed_at_index ON failed_jobs (queue, failed_at);
