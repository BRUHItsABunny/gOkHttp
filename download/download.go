package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"strconv"
)

func DownloadTask(ctx context.Context, global *GlobalDownloadController, task *DownloadTaskController, hClient *http.Client, opts ...requests.Option) error {
	errGr, ctx := errgroup.WithContext(ctx)
	chunkCount := task.ChunkCount.Load()
	for i := uint64(1); i <= chunkCount; i++ {
		chunkKey := strconv.FormatUint(i, 10)
		errGr.Go(func() error {
			chunk, ok := task.Chunks.Get(chunkKey)
			if ok {
				return downloadChunk(ctx, global, task, chunk, hClient, opts...)
			} else {
				return errors.New("cannot find chunk object")
			}
		})

		// chunkKey := strconv.FormatUint(i, 10)
		// chunk, ok := task.Chunks.Get(chunkKey)
		// if ok {
		// 	errGr.Go(func() error {
		// 		return downloadChunk(ctx, global, task, chunk, hClient, opts...)
		// 	})
		// 	// TODO: Better fix for chunk not properly downloading cuz of implicit arg feeding in threading scenario
		// 	time.Sleep(time.Millisecond)
		// } else {
		// 	return errors.New("cannot find chunk object")
		// }
	}
	global.Add(1)

	// Block until done
	err := errGr.Wait()
	if err != nil {
		return fmt.Errorf("errGr.Wait: %w", err)
	}
	global.Done()
	global.DownloadedFiles.Inc()
	global.Tasks.Del(task.FileLocation.Load())

	// Merge
	threads := task.ChunkCount.Load()
	if threads > 1 {
		for i := uint64(1); i <= task.ChunkCount.Load(); i++ {
			chunkKey := strconv.FormatUint(i, 10)
			chunk, ok := task.Chunks.Get(chunkKey)
			if !ok {
				panic("missing chunk during merger")
			}

			_, err := chunk.F.Seek(0, io.SeekStart)
			if err != nil {
				return fmt.Errorf("chunk.F.Seek: %w", err)
			}

			_, err = io.Copy(task.TaskStats.F, chunk.F)
			if err != nil {
				return fmt.Errorf("io.Copy: %w", err)
			}
			err = chunk.F.Close()
			if err != nil {
				return fmt.Errorf("chunk.F.Close: %w", err)
			}
			err = os.Remove(task.FileLocation.Load() + ".part" + chunkKey)
			if err != nil {
				return fmt.Errorf("os.Remove: %w", err)
			}
		}
	}
	err = task.TaskStats.F.Close()
	if err != nil {
		return fmt.Errorf("task.TaskStats.F.Close: %w", err)
	}

	return nil
}
