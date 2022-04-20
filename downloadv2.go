package gokhttp

import (
	"context"
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/dustin/go-humanize"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:generate go-enum -f=$GOFILE --marshal --nocase --flag --names

// Make x ENUM(Initializing,Downloading,Merging,Done,Error)
type TaskStatus int

type DownloadTrackerV2Chunk struct {
	Total     *atomic.Uint64
	TotalDone *atomic.Uint64
	Done      *atomic.Uint64
}
type DownloadTrackerV2Task struct {
	Total        *atomic.Uint64
	TotalDone    *atomic.Uint64
	EpochDone    *atomic.Uint64 // Every epoch, Swap(0) to keep track of speed
	Status       *atomic.Uint64
	IsThread     *atomic.Bool // To track whether this task is a sub-task
	IsResumable  *atomic.Bool
	LockOSThread *atomic.Bool
	Name         *atomic.String // Thread-safe representation of Name (file name)
	URL          *atomic.String // Thread-safe representation of URL
	Chunks       *atomic.Uint64 // Amount of chunks in this task
	// Map of sub-tasks where we track their individual process as well
	ChunkData  *hashmap.HashMap
	chunkStart int64
	chunkStop  int64
	Error      *atomic.Error
	mainTask   *DownloadTrackerV2Task
}

type DownloadStatusV2 struct {
	Status     int
	Percentage float64
	Speed      int64
	Chunks     map[int]*DownloadStatusV2
	Total      int64
	TotalDone  int64
}

func (t *DownloadTrackerV2Task) Percentage() (float64, int64, int64) {
	status := int(t.Status.Load())
	switch TaskStatus(status) {
	case TaskStatusInitializing:
		return 0, 0, 0
	default:
		startStop := float64(t.chunkStop - t.chunkStart)
		total := t.Total.Load()
		totalDone := t.TotalDone.Load()
		current := float64(int64(total - totalDone))

		return (startStop - current) / startStop, int64(total), int64(totalDone)
	}
}

func (t *DownloadTrackerV2Task) Speed() int64 {
	return int64(t.EpochDone.Swap(0))
}

func (t *DownloadTrackerV2Task) Progress() (*DownloadStatusV2, error) {
	var err error
	percentage, total, totalDone := t.Percentage()
	result := &DownloadStatusV2{
		Status:     int(t.Status.Load()),
		Percentage: percentage,
		Total:      total,
		TotalDone:  totalDone,
		Speed:      t.Speed(),
		Chunks:     make(map[int]*DownloadStatusV2),
	}

	threads := t.Chunks.Load()
	for i := uint64(0); i < threads; i++ {
		subTask, ok := t.ChunkData.GetHashedKey(uintptr(i))
		if ok {
			threadTask := subTask.(*DownloadTrackerV2Task)
			percentage, total, totalDone = threadTask.Percentage()
			result.Chunks[int(i)] = &DownloadStatusV2{
				Status:     int(threadTask.Status.Load()),
				Percentage: percentage,
				Total:      total,
				TotalDone:  totalDone,
				Speed:      threadTask.Speed(),
				Chunks:     nil,
			}
		}
	}

	err = t.Error.Load()

	return result, err
}

func parseDuration(format string, duration time.Duration) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(duration).Format(format)
}

func (t *DownloadTrackerV2Task) Render(at time.Time) (*DownloadStatusV2, error) {
	var (
		taskStats, taskStats2 string // Name, status, speed, ETA
		taskFragmentStats     string // speed, fragmentNumber, percentage
		taskBar               string // progress bar
		taskFragmentStatsSub  string
		amount                int
	)
	progress, err := t.Progress()
	eta := time.Duration(0)
	if progress.Speed > 0 {
		eta = time.Duration((progress.Total-progress.TotalDone)/progress.Speed) * time.Second
	}

	taskStats += fmt.Sprintf("%s - File: %s - Status: %s",
		at.Format(time.RFC3339), t.Name.Load(), TaskStatus(int(t.Status.Load())).String())
	taskStats2 += fmt.Sprintf("%s out of %s (%.2f%%) at %s/s - ETA: %s", humanize.Bytes(uint64(progress.TotalDone)), humanize.Bytes(uint64(progress.Total)),
		progress.Percentage*100, humanize.Bytes(uint64(progress.Speed)), parseDuration("15:04:05", eta))
	threads := t.Chunks.Load()
	if len(progress.Chunks) > 1 {
		// [XXXXXXXXXXX][XXXXXXXXXXX][XXXXXXXXXXX][XXXXXXXXXXXX]
		// Also wipe fragment deltas
		for i := uint64(0); i < threads; i++ {
			v := progress.Chunks[int(i)]
			amount = int(math.Round(v.Percentage * 10.0 * 1.7))
			taskFragmentStatsSub = fmt.Sprintf("[%.2f%%|%s/s]", v.Percentage*100, humanize.Bytes(uint64(v.Speed)))
			padLen := 19 - len(taskFragmentStatsSub)
			for j := 0; j < padLen; j++ {
				taskFragmentStatsSub += " "
			}
			taskFragmentStats += taskFragmentStatsSub
			taskBar += "["

			for i := 0; i < 17; i++ {
				if i < amount {
					taskBar += "X"
				} else {
					taskBar += "="
				}
			}
			taskBar += "]"
		}
	} else {
		taskBar += "["
		amount = int(math.Round(progress.Percentage * 100.0 / 2))
		for i := 0; i < 50; i++ {
			if i < amount {
				taskBar += "X"
			} else {
				taskBar += "="
			}
		}
		taskBar += "]"
	}

	fmt.Println(taskStats)
	fmt.Println(taskStats2)
	if len(taskFragmentStats) > 0 {
		fmt.Println(taskFragmentStats)
	}
	fmt.Println(taskBar)
	return progress, err
}

func NewTask(urlStr, fileName string, chunks uint64, lockOSThread bool) *DownloadTrackerV2Task {
	result := &DownloadTrackerV2Task{
		Status:       atomic.NewUint64(uint64(TaskStatusInitializing)),
		IsThread:     atomic.NewBool(false),
		LockOSThread: atomic.NewBool(lockOSThread),
		IsResumable:  atomic.NewBool(false),
		Name:         atomic.NewString(fileName),
		URL:          atomic.NewString(urlStr),
		Total:        atomic.NewUint64(0),
		TotalDone:    atomic.NewUint64(0),
		EpochDone:    atomic.NewUint64(0),
		Chunks:       atomic.NewUint64(chunks),
		ChunkData:    &hashmap.HashMap{},
		Error:        atomic.NewError(nil),
	}

	for i := uint64(0); i < chunks; i++ {
		name := result.Name.Load() + ".fragment." + strconv.FormatUint(i, 10)
		result.ChunkData.SetHashedKey(uintptr(i), &DownloadTrackerV2Task{
			Status:       atomic.NewUint64(uint64(TaskStatusInitializing)),
			IsThread:     atomic.NewBool(true),
			LockOSThread: result.LockOSThread,
			IsResumable:  result.IsResumable,
			Name:         atomic.NewString(name),
			URL:          result.URL,
			Total:        atomic.NewUint64(0),
			TotalDone:    atomic.NewUint64(0),
			EpochDone:    atomic.NewUint64(0),
			Chunks:       atomic.NewUint64(1),
			// Sub tasks don't need more sub tasks
			ChunkData: nil,
			Error:     result.Error,
			mainTask:  result,
		})
	}

	return result
}

func (t *DownloadTrackerV2Task) readyTask(resp *http.Response) {
	//var err error
	// Check for Content-Length and Range support
	supportsRanges := false
	fileSize := resp.ContentLength
	if resp.Request.Method == "HEAD" {
		if resp.Header.Get("Accept-Ranges") == "bytes" {
			supportsRanges = true
		}
		if resp.Header.Get("Ranges-Supported") == "bytes" {
			supportsRanges = true
		}
	} else {
		// GET?
		contentRange := resp.Header.Get("Content-Range")
		if contentRange != "" {
			supportsRanges = true
			fileSize, _ = strconv.ParseInt(strings.Split(contentRange, "/")[1], 10, 64)
		}
	}

	t.chunkStop = fileSize
	t.IsResumable.Store(supportsRanges)
	// Fallback for not support ranges onto single threaded
	if !supportsRanges {
		t.Chunks.Store(1)
	}
	t.Total.Store(uint64(fileSize))
	// If its threaded AND ranges are supported, proceed with threaded download
	threads := t.Chunks.Load()
	info, exists := fileExists(t.Name.String())
	if !exists {
		if threads > 0 {
			// Check for ALL existing fragments
			for thread := uint64(0); thread < threads; thread++ {
				// Check what the start and end range is for this task
				start := uint64(fileSize) / threads * thread
				var stop uint64
				if thread == threads {
					stop = uint64(fileSize)
				} else {
					stop = uint64(fileSize)/threads*(thread+1) - 1
				}
				subTask, ok := t.ChunkData.GetHashedKey(uintptr(thread))
				if ok {
					threadTask := subTask.(*DownloadTrackerV2Task)
					threadTask.chunkStop = int64(stop)
					threadTask.chunkStart = int64(start)
					threadTask.TotalDone.Store(start)
					threadTask.Total.Store(stop)
					// Check if file exists and update task with existing progress accordingly
					info, exists = fileExists(threadTask.Name.String())
					if exists {
						// DONE - Skip to post renaming logic (main task merger)
						threadTask.Status.Store(uint64(TaskStatusDone))
						threadTask.TotalDone.Store(threadTask.Total.Load())
						if threadTask.mainTask != nil {
							threadTask.mainTask.TotalDone.Add(uint64(info.Size()))
						}
					} else {
						info, exists = fileExists(threadTask.Name.String() + ".tmp")
						if exists {
							if supportsRanges {
								// Resume
								threadTask.TotalDone.Add(uint64(info.Size()))
								if threadTask.mainTask != nil {
									threadTask.mainTask.TotalDone.Add(uint64(info.Size()))
								}
							} else {
								// Delete and start over
							}
						}
					}

					if threads == 1 {
						if thread == 0 {
							t.TotalDone.Store(threadTask.TotalDone.Load())
							if threadTask.Status.Load() == uint64(TaskStatusDone) {
								t.Status.Store(uint64(TaskStatusMerging))
							}
						} else {
							// Remove fragments
							t.ChunkData.DelHashedKey(uintptr(thread))
						}
					}

					if threadTask.TotalDone.Load() >= threadTask.Total.Load() {
						threadTask.Status.Store(uint64(TaskStatusDone))
					} else {
						threadTask.Status.Store(uint64(TaskStatusDownloading))
					}
				}
			}
		}
	} else {
		t.Status.Store(uint64(TaskStatusDone))
	}
}

func (c *HttpClient) CheckResourceV2(task *DownloadTrackerV2Task, parameters url.Values, headers map[string]string) error {
	var (
		err  error
		req  *http.Request
		resp *http.Response
	)
	// Send HEAD request if task isn't a thread
	if !task.IsThread.Load() {
		req, err = c.MakeHEADRequest(task.URL.Load(), parameters, headers)
		if err == nil {
			resp, err = c.Client.Do(req)
			if err == nil {
				// check status code too
				task.readyTask(resp)
			} else {
				// Check first using a GET request, should not give error but it should either have or not have the right headers
				req, err = c.MakeGETRequest(task.URL.Load(), parameters, headers)
				if err == nil {
					req.Header.Add("Range", "bytes=0-0")
					resp, err = c.Client.Do(req)
					if err == nil {
						// if PARTIAL CONTENT, else also single thread fallback
						task.readyTask(resp)
					} else {
						task.Chunks.Store(1)
					}
				}
			}
		}
	}
	return err
}

func (c *HttpClient) DownloadFileV2(task *DownloadTrackerV2Task, parameters url.Values, headers map[string]string) error {
	var err error
	// BOOTSTRAP, maybe just store entire task object?
	// Make DownloadClient, it has no HTTP timeout
	client := GetHTTPDownloadClient(c.ClientOptions)
	err = c.CheckResourceV2(task, parameters, headers)
	w := "c.CheckResourceV2"

	if err == nil {
		if task.IsThread.Load() {
			// It's a thread, download
			if task.LockOSThread.Load() {
				runtime.LockOSThread()
			}

			status := TaskStatus(task.Status.Load())
			if status == TaskStatusDownloading {
				var (
					err2    error
					resp    *HttpResponse
					written int64
					f       *os.File
				)
				// Make the Request
				req := client.makeDownloadRequestV2(task, parameters, headers)
				// Start Downloading
				resp, err = client.Do(req)
				w = "client.Do"
				if err == nil {
					// Open file, if non-existent then create file
					f, err = os.OpenFile(task.Name.Load()+".tmp", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
					w = "os.OpenFile"
					if err == nil {
						// Start reading and report back on each read
						w = "io.CopyN"
						for {
							written, err = io.CopyN(f, resp.Body, 1024)
							task.TotalDone.Add(uint64(written))
							task.EpochDone.Add(uint64(written))
							if task.mainTask != nil {
								task.mainTask.TotalDone.Add(uint64(written))
								task.mainTask.EpochDone.Add(uint64(written))
							}

							if task.TotalDone.Load() >= task.Total.Load()-1 {
								err = nil
								break
							}
							if err != nil {
								break
							}
						}
						// Close file and response after the loop finishes
						err2 = f.Close()
						if err == nil {
							w = "f.Close"
							err = err2
						}

						_ = resp.Body.Close()
						err2 = os.Rename(f.Name(), task.Name.Load())
						if err == nil {
							w = "os.Rename"
							err = err2
						}

						if err == nil {
							task.Status.Store(uint64(TaskStatusDone))
						} else {
							task.Error.Store(fmt.Errorf("%s: %s: %w", w, task.Name.Load(), err))
							task.Status.Store(uint64(TaskStatusError))
						}
					}
				}
			}

			if task.LockOSThread.Load() {
				runtime.UnlockOSThread()
			}
		} else {
			// Its not a thread, launch the threads
			var (
				ok      bool
				subTask interface{}
			)
			status := task.Status.Load()
			errGroup, _ := errgroup.WithContext(context.Background())

			if status == uint64(TaskStatusInitializing) {
				task.Status.Store(uint64(TaskStatusDownloading))
				for i := uint64(0); i < task.Chunks.Load(); i++ {
					subTask, ok = task.ChunkData.GetHashedKey(uintptr(i))
					if ok {
						threadTask := subTask.(*DownloadTrackerV2Task)
						errGroup.Go(func() error {
							err := c.DownloadFileV2(threadTask, parameters, headers)
							return err
						})
					}
				}
			}

			err = errGroup.Wait()
			if err == nil {
				// Now merge
				status = task.Status.Load()
				if status == uint64(TaskStatusDownloading) {
					task.Status.Store(uint64(TaskStatusMerging))
					status = task.Status.Load()
				}

				if status == uint64(TaskStatusMerging) {
					var (
						file, subFile *os.File
					)
					for i := uint64(0); i < task.Chunks.Load(); i++ {
						subTask, ok = task.ChunkData.GetHashedKey(uintptr(i))
						if ok {
							threadTask := subTask.(*DownloadTrackerV2Task)
							if i == 0 {
								err = os.Rename(threadTask.Name.Load(), task.Name.Load())
								if err == nil {
									file, err = os.OpenFile(task.Name.Load(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
								}
							} else {
								if file != nil && err == nil {
									subFile, err = os.Open(threadTask.Name.Load())
									if err == nil {
										_, err = io.Copy(file, subFile)
										if err == nil {
											_ = subFile.Close()
											err = os.Remove(threadTask.Name.Load())
										}
									}
								}
							}
						}

						if err != nil {
							break
						}
					}
				}

			}

			if err != nil {
				task.Status.Store(uint64(TaskStatusError))
				task.Error.Store(err)
			} else {
				task.Status.Store(uint64(TaskStatusDone))
			}
		}
	}

	return err
}

func (c *HttpClient) makeDownloadRequestV2(task *DownloadTrackerV2Task, parameters url.Values, headers map[string]string) *http.Request {
	var (
		err error
		req *http.Request
	)
	// Set context
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "GET", task.URL.Load(), nil)
	} else {
		req, err = http.NewRequest("GET", task.URL.Load(), nil)
	}
	if checkError(err) {
		// Ready from params, cookies and headers stored within the Client
		req = c.readyRequest(req)
		// Ready from params and headers from arguments
		query := req.URL.Query()
		for k, v := range parameters {
			for _, e := range v {
				query.Add(k, e)
			}
		}
		req.URL.RawQuery = query.Encode()
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		if task.IsResumable.Load() {
			byteRange := fmt.Sprintf("bytes=%d-%d", task.TotalDone.Load(), task.Total.Load())
			req.Header.Add("Range", byteRange)
		}
		return req
	}
	return nil
}
