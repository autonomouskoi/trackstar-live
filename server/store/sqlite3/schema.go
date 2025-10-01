package sqlite3

const schema1 = `
CREATE TABLE track_updates (
	user_id      TEXT,
	started      INT,
	deck_id      TEXT,
	artist       TEXT,
	title        TEXT,
	played_when  INT,
	idx          INT
);

-- Most lookups will be by user+session
CREATE INDEX session ON track_updates (user_id, started);

-- Prevent duplicate set entries
CREATE UNIQUE INDEX unique_updates ON track_updates (user_id, started, played_when);

PRAGMA user_version=1;
PRAGMA optimize;
`
