-- migrations/000001_init.up.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS songs (
                                     id SERIAL PRIMARY KEY,
                                     group_name TEXT NOT NULL,
                                     song_name TEXT NOT NULL,
                                     release_date DATE NOT NULL,
                                     lyrics TEXT NOT NULL,
                                     link TEXT NOT NULL,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_songs_group ON songs(group_name);
CREATE INDEX IF NOT EXISTS idx_songs_song ON songs(song_name);

-- +goose Down
DROP TABLE IF EXISTS songs;