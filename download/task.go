package gokhttp_download

import (
	"context"
	"strings"
)

// DownloadType x ENUM(Threaded, LiveHLS)
//
//go:generate go-enum --marshal --ptr
type DownloadType int

// DownloadVersion x ENUM(v1)
type DownloadVersion int

type DownloadTask interface {
	Download(ctx context.Context) error
	Type() DownloadType
	Progress(sb *strings.Builder) error
	ResetDelta()
}
