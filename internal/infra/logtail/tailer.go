package logtail

import (
	"context"
	"io"
)

type Tailer interface {
	Tail(ctx context.Context, path string, lines int) (io.ReadCloser, error)
}
