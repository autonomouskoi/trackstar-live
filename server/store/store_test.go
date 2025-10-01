package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/autonomouskoi/trackstar-live/server/store"
	"github.com/autonomouskoi/trackstar-live/server/store/sqlite3"
	trackstar "github.com/autonomouskoi/trackstar/pb"
)

func TestSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	db, err := sqlite3.New(":memory:")
	require.NoError(t, err, "creating database")

	userID := "test-user"

	store := store.New(db)
	sessions, err := store.SessionsList(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, sessions)

	sessionStarted := time.Now().UnixMilli()
	trackWhen := sessionStarted + 5000
	deckID := "deck-1"
	artist := "the-artist"
	title := "the-title"

	tu := &trackstar.TrackUpdate{
		DeckId: deckID,
		Track: &trackstar.Track{
			Artist: artist,
			Title:  title,
		},
		When:  trackWhen,
		Index: 1,
	}
	require.NoError(t, store.AddTrackUpdate(ctx, userID, sessionStarted, tu), "adding update")

	sessions, err = store.SessionsList(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, []int64{sessionStarted}, sessions)

	updates, err := store.SessionGet(ctx, userID, sessionStarted)
	require.NoError(t, err, "getting session")
	require.Len(t, updates, 1)
	require.Equal(t, deckID, updates[0].GetDeckId())
	require.Equal(t, trackWhen, updates[0].GetWhen())
	require.Equal(t, int32(1), updates[0].GetIndex())
	require.Equal(t, artist, updates[0].GetTrack().GetArtist())
	require.Equal(t, title, updates[0].GetTrack().GetTitle())

	require.ErrorContains(t, store.SessionDelete(ctx, userID, sessionStarted), "not implemented")
}
