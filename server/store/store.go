package store

import (
	"context"
	"database/sql"
	"errors"

	trackstar "github.com/autonomouskoi/trackstar/pb"
)

type DB interface {
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	Rebind(string) string
	SelectContext(context.Context, any, string, ...any) error
	Close() error
}

type Store struct {
	db DB
}

func New(db DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) SessionsList(ctx context.Context, userID string) ([]int64, error) {
	query := s.db.Rebind(`
SELECT DISTINCT(started) FROM track_updates
	WHERE user_id = ?
	ORDER BY started DESC
`)
	sessions := []int64{}
	err := s.db.SelectContext(ctx, &sessions, query, userID)
	return sessions, err
}

type trackUpdate struct {
	UserID  string `db:"user_id"`
	Started int64  `db:"started"`
	DeckID  string `db:"deck_id"`
	Artist  string `db:"artist"`
	Title   string `db:"title"`
	When    int64  `db:"played_when"`
	Index   int32  `db:"idx"`
}

func (s *Store) SessionGet(ctx context.Context, userID string, started int64) ([]*trackstar.TrackUpdate, error) {
	query := s.db.Rebind(`
SELECT deck_id, artist, title, played_when, idx FROM track_updates
	WHERE user_id = ? AND started = ?
	ORDER BY idx ASC
`)
	matches := []*trackUpdate{}
	if err := s.db.SelectContext(ctx, &matches, query, userID, started); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*trackstar.TrackUpdate{}, nil
		}
		return nil, err
	}
	updates := make([]*trackstar.TrackUpdate, len(matches))
	for i, match := range matches {
		updates[i] = &trackstar.TrackUpdate{
			DeckId: match.DeckID,
			Track: &trackstar.Track{
				Artist: match.Artist,
				Title:  match.Title,
			},
			When:  match.When,
			Index: match.Index,
		}
	}
	return updates, nil
}

func (s *Store) SessionDelete(_ context.Context, userID string, started int64) error {
	return errors.New("not implemented")
}

func (s *Store) AddTrackUpdate(ctx context.Context, userID string, sessionStarted int64, tu *trackstar.TrackUpdate) error {
	stmt := s.db.Rebind(`
INSERT INTO track_updates (
	user_id,
	started,
	deck_id,
	artist,
	title,
	played_when,
	idx
) VALUES (
	:user_id,
	:started,
	:deck_id,
	:artist,
	:title,
	:played_when,
	:idx
)`)
	_, err := s.db.NamedExecContext(ctx, stmt, &trackUpdate{
		UserID:  userID,
		Started: sessionStarted,
		DeckID:  tu.GetDeckId(),
		Artist:  tu.GetTrack().GetArtist(),
		Title:   tu.GetTrack().GetTitle(),
		When:    tu.GetWhen(),
		Index:   tu.GetIndex(),
	})
	return err
}
