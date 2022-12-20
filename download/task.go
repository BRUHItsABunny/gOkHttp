package download

import (
	"context"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/cornelk/hashmap"
	"go.uber.org/atomic"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type DownloadTaskController struct {
	// Metadata
	FileName     *atomic.String `json:"fileName"`
	FileLocation *atomic.String `json:"fileLocation"`
	FileURL      *atomic.String `json:"fileURL"`
	Resumable    *atomic.Bool   `json:"resumable"`
	// Totals
	TaskStats *DownloadChunkController `json:"taskStats"`
	// Chunks
	Chunks     *hashmap.Map[string, *DownloadChunkController] `json:"chunks"`
	ChunkCount *atomic.Uint64                                 `json:"chunkCount"`
}

func supportsRange(resp *http.Response) (uint64, bool) {
	supportsRanges := false
	contentLength := uint64(0)
	if resp != nil {
		if resp.Request.Method == http.MethodHead && resp.StatusCode == http.StatusOK {
			if resp.Header.Get("Accept-Ranges") == "bytes" {
				supportsRanges = true
			}
			if resp.Header.Get("Ranges-Supported") == "bytes" {
				supportsRanges = true
			}
			contentLength = uint64(resp.ContentLength)
		} else if resp.Request.Method == http.MethodGet && resp.StatusCode == http.StatusPartialContent {
			contentRange := resp.Header.Get("Content-Range")
			if contentRange != "" {
				supportsRanges = true
			}
			cLSplit := strings.Split(contentRange, "/")
			if len(cLSplit) == 2 {
				contentLength, _ = strconv.ParseUint(cLSplit[1], 10, 64)
			}
		} else {
			// File inaccessible
		}
	}
	return contentLength, supportsRanges
}

func isResumable(hClient *http.Client, fileURL string, opts ...requests.Option) (bool, uint64, error) {
	req, err := requests.MakeHEADRequest(context.Background(), fileURL, opts...)
	if err != nil {
		return false, 0, fmt.Errorf("requests.MakeHEADRequest: %w", err)
	}
	resp, err := hClient.Do(req)
	if err != nil {
		req, err = requests.MakeGETRequest(context.Background(), fileURL, opts...)
		if err != nil {
			return false, 0, fmt.Errorf("requests.MakeGETRequest: %w", err)
		}
		req.Header.Set("Range", "bytes=0-0")
		resp, err = hClient.Do(req)
	}

	if err != nil {
		return false, 0, fmt.Errorf("hClient.Do: %w", err)
	}
	length, ok := supportsRange(resp)
	return ok, length, nil
}

func NewDownloadTaskController(hClient *http.Client, global *GlobalDownloadController, fileName, fileLocation, fileURL string, threads, expectedSize uint64, opts ...requests.Option) (*DownloadTaskController, error) {
	result := &DownloadTaskController{
		FileName:     atomic.NewString(fileName),
		FileLocation: atomic.NewString(fileLocation),
		FileURL:      atomic.NewString(fileURL),
		Resumable:    atomic.NewBool(false),
		TaskStats: &DownloadChunkController{
			ChunkID:         atomic.NewUint64(0),
			ChunkStart:      0,
			ChunkStop:       0,
			F:               nil,
			FileSize:        atomic.NewUint64(0),
			DownloadedBytes: atomic.NewUint64(0),
			DeltaBytes:      atomic.NewUint64(0),
		},
		Chunks:     hashmap.New[string, *DownloadChunkController](),
		ChunkCount: atomic.NewUint64(1),
	}

	// Is the file already downloaded and complete
	var err error
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

	ok, newSize, err := isResumable(hClient, fileURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("isResumable:%w", err)
	}
	// Unable to thread
	if !ok {
		threads = 1
	}
	// Use server returned length
	if expectedSize != newSize {
		expectedSize = newSize
	}
	result.ChunkCount.Store(threads)
	result.TaskStats.FileSize.Store(expectedSize)
	chunkSize := result.TaskStats.FileSize.Load() / threads
	for i := uint64(1); i <= threads; i++ {
		chunkKey := strconv.FormatUint(i, 10)
		chunk := &DownloadChunkController{
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

	global.TotalBytes.Add(result.TaskStats.FileSize.Load())
	global.DownloadedBytes.Add(result.TaskStats.DownloadedBytes.Load())
	global.Tasks.Set(fileLocation, result)
	return result, nil
}
