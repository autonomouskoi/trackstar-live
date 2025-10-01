package server

import (
	"context"

	trackstar "github.com/autonomouskoi/trackstar/pb"
)

type Store interface {
	SessionsList(ctx context.Context, userID string) ([]int64, error)
	SessionGet(ctx context.Context, userID string, started int64) ([]*trackstar.TrackUpdate, error)
	SessionDelete(ctx context.Context, userID string, started int64) error

	AddTrackUpdate(ctx context.Context, userID string, sessionStarted int64, tu *trackstar.TrackUpdate) error
}
