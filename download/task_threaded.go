package gokhttp_download

import (
	"context"
	"errors"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/cornelk/hashmap"
	"github.com/dustin/go-humanize"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ThreadedChunk struct {
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

type ThreadedDownloadTask struct {
	// Tracking ref
	Global  *GlobalDownloadTracker    `json:"-"`
	HClient *http.Client              `json:"-"`
	ReqOpts []gokhttp_requests.Option `json:"-"`
	// Metadata
	TaskType     DownloadType    `json:"taskType"`
	TaskVersion  DownloadVersion `json:"taskVersion"`
	FileName     *atomic.String  `json:"fileName"`
	FileLocation *atomic.String  `json:"fileLocation"`
	FileURL      *atomic.String  `json:"fileURL"`
	Resumable    *atomic.Bool    `json:"resumable"`
	// Totals
	TaskStats *ThreadedChunk `json:"taskStats"`
	// Chunks
	Chunks     *hashmap.Map[string, *ThreadedChunk] `json:"chunks"`
	ChunkCount *atomic.Uint64                       `json:"chunkCount"`
}

func (tt *ThreadedDownloadTask) Download(ctx context.Context) error {
	errGr, ctx := errgroup.WithContext(ctx)
	chunkCount := tt.ChunkCount.Load()
	for i := uint64(1); i <= chunkCount; i++ {
		chunkKey := strconv.FormatUint(i, 10)
		errGr.Go(func() error {
			chunk, ok := tt.Chunks.Get(chunkKey)
			if ok {
				err := tt.downloadChunk(ctx, chunk)
				if err != nil {
					return fmt.Errorf("tt.downloadChunk: %w", err)
				}
				return nil
			} else {
				return errors.New("cannot find chunk object")
			}
		})
	}
	tt.Global.Add(1)

	// Block until done
	err := errGr.Wait()
	if err != nil {
		tt.Global.Done()
		return fmt.Errorf("errGr.Wait: %w", err)
	}
	tt.Global.Done()
	tt.Global.DownloadedFiles.Inc()
	tt.Global.Tasks.Del(tt.FileLocation.Load())

	// Merge
	threads := tt.ChunkCount.Load()
	if threads > 1 {
		for i := uint64(1); i <= tt.ChunkCount.Load(); i++ {
			chunkKey := strconv.FormatUint(i, 10)
			chunk, ok := tt.Chunks.Get(chunkKey)
			if !ok {
				panic("missing chunk during merger")
			}

			_, err := chunk.F.Seek(0, io.SeekStart)
			if err != nil {
				return fmt.Errorf("chunk.F.Seek: %w", err)
			}

			_, err = io.Copy(tt.TaskStats.F, chunk.F)
			if err != nil {
				return fmt.Errorf("io.Copy: %w", err)
			}
			err = chunk.F.Close()
			if err != nil {
				return fmt.Errorf("chunk.F.Close: %w", err)
			}
			err = os.Remove(tt.FileLocation.Load() + ".part" + chunkKey)
			if err != nil {
				return fmt.Errorf("os.Remove: %w", err)
			}
		}
	}
	err = tt.TaskStats.F.Close()
	if err != nil {
		return fmt.Errorf("task.TaskStats.F.Close: %w", err)
	}

	return nil
}

func (tt *ThreadedDownloadTask) Type() DownloadType {
	return tt.TaskType
}

func progress(result *strings.Builder, delta, downloaded, totalSize uint64, barSize int) {
	result.WriteString("[")
	percentage := float64(downloaded) / float64(totalSize)
	pieces := int(math.Floor(float64(barSize) * percentage))
	for i := 0; i < pieces; i++ {
		result.WriteString("X")
	}
	for i := 0; i < barSize-pieces; i++ {
		result.WriteString(" ")
	}
	result.WriteString("]")
	etaSecs := math.Ceil(float64(totalSize-downloaded) / float64(delta))
	eta := time.Second * time.Duration(etaSecs)
	result.WriteString(fmt.Sprintf(" (%.2f%% at %s/s, ETA: %s)\n", percentage*100, humanize.Bytes(delta), eta.String()))
}

func (tt *ThreadedDownloadTask) Progress(sb *strings.Builder) error {
	sb.WriteString(fmt.Sprintf("%s\n", Truncate(tt.FileLocation.Load(), 128, 0)))
	sb.WriteString(fmt.Sprintf("Download statistic: %s out of %s\nTotal file: ", humanize.Bytes(tt.TaskStats.DownloadedBytes.Load()), humanize.Bytes(tt.TaskStats.FileSize.Load())))
	progress(sb, tt.TaskStats.DeltaBytes.Load(), tt.TaskStats.DownloadedBytes.Load(), tt.TaskStats.FileSize.Load(), 40)
	chunkCount := tt.ChunkCount.Load()
	if chunkCount > 1 {
		for i := uint64(1); i <= chunkCount; i++ {
			chunkKey := strconv.FormatUint(i, 10)
			chunk, ok := tt.Chunks.Get(chunkKey)
			if ok {
				sb.WriteString(fmt.Sprintf("Chunk %s: ", chunkKey))
				progress(sb, chunk.DeltaBytes.Load(), chunk.DownloadedBytes.Load(), chunk.FileSize.Load(), 40)
			}
		}
	}
	return nil
}

func (tt *ThreadedDownloadTask) ResetDelta() {
	tt.TaskStats.DeltaBytes.Store(0)
	for i := uint64(1); i <= tt.ChunkCount.Load(); i++ {
		chunkKey := strconv.FormatUint(i, 10)
		chunk, ok := tt.Chunks.Get(chunkKey)
		if !ok {
			panic("missing chunk during merger")
		}

		chunk.DeltaBytes.Store(0)
	}
}

func (tt *ThreadedDownloadTask) isResumable(ctx context.Context) error {
	fileURL := tt.FileURL.Load()
	req, err := gokhttp_requests.MakeHEADRequest(ctx, fileURL, tt.ReqOpts...)
	if err != nil {
		return fmt.Errorf("requests.MakeHEADRequest: %w", err)
	}
	resp, err := tt.HClient.Do(req)
	if err != nil {
		req, err = gokhttp_requests.MakeGETRequest(ctx, fileURL, tt.ReqOpts...)
		if err != nil {
			return fmt.Errorf("requests.MakeGETRequest: %w", err)
		}
		req.Header.Set("Range", "bytes=0-0")
		resp, err = tt.HClient.Do(req)
	}

	if err != nil {
		return fmt.Errorf("tt.HClient.Do: %w", err)
	}

	length, ok := supportsRange(resp)
	tt.Resumable.Store(ok)
	if ok {
		predictedLength := tt.TaskStats.FileSize.Load()
		// fmt.Println(fmt.Sprintf("isResumable: length: %d and predictedLength: %d", length, predictedLength))
		if length != 0 && length != predictedLength {
			tt.TaskStats.FileSize.Store(length)
		}
	} else {
		tt.ChunkCount.Store(1)
	}
	return nil
}

func NewThreadedDownloadTask(ctx context.Context, hClient *http.Client, global *GlobalDownloadTracker, fileLocation, fileURL string, threads, expectedSize uint64, opts ...gokhttp_requests.Option) (*ThreadedDownloadTask, error) {
	result := &ThreadedDownloadTask{
		ReqOpts:      opts,
		HClient:      hClient,
		Global:       global,
		TaskType:     DownloadTypeThreaded,
		TaskVersion:  DownloadVersionV1,
		FileName:     atomic.NewString(filepath.Base(fileLocation)),
		FileLocation: atomic.NewString(fileLocation),
		FileURL:      atomic.NewString(fileURL),
		Resumable:    atomic.NewBool(false),
		TaskStats: &ThreadedChunk{
			ChunkID:         atomic.NewUint64(0),
			ChunkStart:      0,
			ChunkStop:       0,
			F:               nil,
			FileSize:        atomic.NewUint64(0),
			DownloadedBytes: atomic.NewUint64(0),
			DeltaBytes:      atomic.NewUint64(0),
		},
		Chunks:     hashmap.New[string, *ThreadedChunk](),
		ChunkCount: atomic.NewUint64(1),
	}

	// Is the dir accessible and is the file already downloaded and complete
	err := os.MkdirAll(filepath.Dir(result.FileLocation.Load()), 0600)
	if err != nil {
		return nil, fmt.Errorf("os.MkdirAll: %w", err)
	}
	result.TaskStats.F, err = os.OpenFile(fileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile: %w", err)
	}
	stats, err := result.TaskStats.F.Stat()
	if err != nil {
		return nil, fmt.Errorf("f.Stat: %w", err)
	}
	result.TaskStats.DownloadedBytes.Store(uint64(stats.Size()))
	if stats.Size() > 0 {
		threads = 1
	}

	result.TaskStats.FileSize.Store(expectedSize)
	err = result.isResumable(ctx)
	if err != nil {
		return nil, fmt.Errorf("isResumable:%w", err)
	}

	result.ChunkCount.Store(threads)
	chunkSize := result.TaskStats.FileSize.Load() / threads
	for i := uint64(1); i <= threads; i++ {
		chunkKey := strconv.FormatUint(i, 10)
		chunk := &ThreadedChunk{
			ChunkID:         atomic.NewUint64(i),
			FileSize:        atomic.NewUint64(0),
			DownloadedBytes: atomic.NewUint64(0),
			DeltaBytes:      atomic.NewUint64(0),
			ChunkStart:      0,
			ChunkStop:       0,
			F:               nil,
		}
		rangeStart := chunkSize * (i - 1)
		rangeStop := chunkSize*(i) - 1
		if i == threads {
			rangeStop = result.TaskStats.FileSize.Load()
		}
		chunk.ChunkStart = rangeStart
		chunk.ChunkStop = rangeStop
		chunk.FileSize.Store(rangeStop - rangeStart)
		chunkFileName := fileLocation + ".part" + chunkKey
		if threads > 1 {
			// size detection
			chunk.F, err = os.OpenFile(chunkFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
			if err != nil {
				return nil, fmt.Errorf("os.OpenFile: %w", err)
			}
			stats, err = chunk.F.Stat()
			if err != nil {
				return nil, fmt.Errorf("f.Stat: %w", err)
			}
			chunk.DownloadedBytes.Add(uint64(stats.Size()))
			chunk.ChunkStart += uint64(stats.Size())
			result.TaskStats.DownloadedBytes.Add(uint64(stats.Size()))
		} else {
			// Single threaded
			chunk.F = result.TaskStats.F
			chunk.FileSize.Store(result.TaskStats.FileSize.Load())
			chunk.DownloadedBytes.Store(result.TaskStats.DownloadedBytes.Load())
			chunk.ChunkStart += chunk.DownloadedBytes.Load()
			chunk.ChunkStop = chunk.FileSize.Load()
		}
		result.Chunks.Set(chunkKey, chunk)
	}

	global.TotalFiles.Inc()
	global.TotalBytes.Add(result.TaskStats.FileSize.Load())
	global.DownloadedBytes.Add(result.TaskStats.DownloadedBytes.Load())
	global.Tasks.Set(fileLocation, result)
	return result, nil
}

func (tt *ThreadedDownloadTask) downloadChunk(ctx context.Context, chunk *ThreadedChunk) error {
	tt.Global.TotalThreads.Inc()
	if chunk.DownloadedBytes.Load() >= chunk.FileSize.Load() {
		tt.Global.TotalThreads.Dec()
		return nil
	}
	req, err := gokhttp_requests.MakeGETRequest(ctx, tt.FileURL.Load(), tt.ReqOpts...)
	if err != nil {
		tt.Global.TotalThreads.Dec()
		return fmt.Errorf("[%s:%d] requests.MakeGETRequest: %w", tt.FileName.String(), chunk.ChunkID.Load(), err)
	}
	req.Close = true

	if tt.Resumable.Load() && (chunk.ChunkStart > 0 || chunk.ChunkStop < tt.TaskStats.FileSize.Load()) {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk.ChunkStart, chunk.ChunkStop))
	}
	resp, err := tt.HClient.Do(req)
	if err != nil {
		tt.Global.TotalThreads.Dec()
		err = fmt.Errorf("[%s:%d] tt.HClient.Do: %w", tt.FileName.String(), chunk.ChunkID.Load(), err)
		return err
	}

	// Thread safe downloading and tracking and graceful stop
	var written int64
	for {
		written, err = io.CopyN(chunk.F, resp.Body, 1024)
		// Store all
		chunk.DownloadedBytes.Add(uint64(written))
		chunk.DeltaBytes.Add(uint64(written))
		tt.TaskStats.DownloadedBytes.Add(uint64(written))
		tt.TaskStats.DeltaBytes.Add(uint64(written))
		tt.Global.DownloadedBytes.Add(uint64(written))
		tt.Global.DeltaBytes.Add(uint64(written))

		shouldStop := tt.Global.GraceFulStop.Load()
		if chunk.DownloadedBytes.Load() >= chunk.FileSize.Load() || shouldStop {
			err = nil
			_ = resp.Body.Close()
			// _ = chunk.F.Close()
			break
		}
		if err != nil {
			err = fmt.Errorf("[%s:%d] io.CopyN: %w", tt.FileName.String(), chunk.ChunkID.Load(), err)
			break
		}
	}

	tt.Global.TotalThreads.Dec()
	return err
}
