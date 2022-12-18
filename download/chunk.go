package download

import (
	"context"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"go.uber.org/atomic"
	"io"
	"net/http"
	"os"
)

type DownloadChunkController struct {
	// Metadata
	ChunkID    *atomic.Uint64 `json:"chunkID"`
	ChunkStart uint64         `json:"-"`
	ChunkStop  uint64         `json:"-"`
	F          *os.File       `json:"-"`
	// Totals
	FileSize *atomic.Uint64 `json:"fileSize"`
	// Download stats
	DownloadedBytes *atomic.Uint64 `json:"downloadedBytes"`
	DeltaBytes      *atomic.Uint64 `json:"deltaBytes"`
}

func downloadChunk(ctx context.Context, global *GlobalDownloadController, task *DownloadTaskController, chunk *DownloadChunkController, hClient *http.Client, opts ...requests.Option) error {
	global.TotalThreads.Inc()
	if chunk.DownloadedBytes.Load() >= chunk.FileSize.Load() {
		return nil
	}
	req, err := requests.MakeGETRequest(ctx, task.FileURL.Load(), opts...)
	if err != nil {
		global.TotalThreads.Dec()
		return fmt.Errorf("[%s---%d] requests.MakeGETRequest: %w", task.FileName.String(), chunk.ChunkID.Load(), err)
	}

	if task.ChunkCount.Load() > 1 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk.ChunkStart, chunk.ChunkStop))
	}
	resp, err := hClient.Do(req)
	if err != nil {
		global.TotalThreads.Dec()
		err = fmt.Errorf("[%s---%d] hClient.Do: %w", task.FileName.String(), chunk.ChunkID.Load(), err)
		return err
	}

	// Thread safe downloading and tracking and graceful stop
	var written int64
	for {
		written, err = io.CopyN(chunk.F, resp.Body, 1024)
		// Store all
		chunk.DownloadedBytes.Add(uint64(written))
		chunk.DeltaBytes.Add(uint64(written))
		task.TaskStats.DownloadedBytes.Add(uint64(written))
		task.TaskStats.DeltaBytes.Add(uint64(written))
		global.DownloadedBytes.Add(uint64(written))
		global.DeltaBytes.Add(uint64(written))

		shouldStop := global.GraceFulStop.Load()
		if chunk.DownloadedBytes.Load() >= chunk.FileSize.Load() || shouldStop {
			err = nil
			_ = resp.Body.Close()
			// _ = chunk.F.Close()
			break
		}
		if err != nil {
			err = fmt.Errorf("[%s---%d] io.CopyN: %w", task.FileName.String(), chunk.ChunkID.Load(), err)
			break
		}
	}

	global.TotalThreads.Dec()
	return err
}
