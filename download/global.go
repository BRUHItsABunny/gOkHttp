package gokhttp_download

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/BRUHItsABunny/gOkHttp/responses"
	"github.com/cornelk/hashmap"
	"github.com/dustin/go-humanize"
	"go.uber.org/atomic"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GlobalDownloadTracker struct {
	// Synchronization
	sync.WaitGroup
	GraceFulStop *atomic.Bool `json:"-"`
	// Idle stats
	IdleSince   *atomic.Time   `json:"idleSince"`
	IdleValue   *atomic.Uint64 `json:"-"`
	IdleTimeout *atomic.Int64  `json:"-"`
	// MetaData
	LastTick     *atomic.Time                       `json:"lastTick"`
	CurrentIP    *atomic.String                     `json:"currentIP"`
	TotalThreads *atomic.Uint64                     `json:"totalThreads"`
	Tasks        *hashmap.Map[string, DownloadTask] `json:"tasks"`
	// Total stats
	TotalFiles *atomic.Uint64 `json:"totalFiles"`
	TotalBytes *atomic.Uint64 `json:"totalBytes"`
	// Download stats
	DownloadedFiles *atomic.Uint64 `json:"downloadedFiles"`
	DownloadedBytes *atomic.Uint64 `json:"downloadedBytes"`
	// Per tick stats
	DeltaBytes *atomic.Uint64 `json:"deltaBytes"`
}

func NewGlobalDownloadTracker(idleTimeout time.Duration) *GlobalDownloadTracker {
	return &GlobalDownloadTracker{
		WaitGroup:       sync.WaitGroup{},
		GraceFulStop:    atomic.NewBool(false),
		IdleValue:       atomic.NewUint64(0),
		IdleSince:       atomic.NewTime(time.Time{}),
		IdleTimeout:     atomic.NewInt64(int64(idleTimeout)),
		LastTick:        atomic.NewTime(time.Now()),
		CurrentIP:       atomic.NewString(""),
		TotalThreads:    atomic.NewUint64(0),
		Tasks:           hashmap.New[string, DownloadTask](),
		TotalFiles:      atomic.NewUint64(0),
		TotalBytes:      atomic.NewUint64(0),
		DownloadedFiles: atomic.NewUint64(0),
		DownloadedBytes: atomic.NewUint64(0),
		DeltaBytes:      atomic.NewUint64(0),
	}
}

func GetCurrentIPAddress(hClient *http.Client) (ip string) {
	req, err := gokhttp_requests.MakeGETRequest(context.Background(), "https://httpbin.org/get")
	if err != nil {
		return
	}
	resp, err := hClient.Do(req)
	if err != nil {
		return
	}
	data := map[string]any{}
	err = gokhttp_responses.ResponseJSON(resp, &data)

	if err != nil {
		return
	}
	ipPreCast, ok := data["origin"]
	if ok {
		ip = ipPreCast.(string)
	}
	return
}

func (global *GlobalDownloadTracker) Stop() {
	global.GraceFulStop.Store(true)
	global.Wait()
}

func (global *GlobalDownloadTracker) IdleTimeoutExceeded() bool {
	idleSince := global.IdleSince.Load()
	idleTimeout := global.IdleTimeout.Load()
	if !idleSince.IsZero() {
		return time.Now().Sub(idleSince) >= time.Duration(idleTimeout)
	}
	return false
}

func (global *GlobalDownloadTracker) PollIP(hClient *http.Client) {
	go global.pollIP(hClient)
}

func (global *GlobalDownloadTracker) pollIP(hClient *http.Client) {
	global.Add(1)
	global.CurrentIP.Store(GetCurrentIPAddress(hClient))
	i := 0
	for {
		if i >= 59 {
			i = 0
			global.CurrentIP.Store(GetCurrentIPAddress(hClient))
		}
		time.Sleep(time.Second)
		if global.GraceFulStop.Load() {
			break
		}
		i++
	}
	global.Done()
}

func (global *GlobalDownloadTracker) Tick(humanReadable bool) string {
	result := &strings.Builder{}
	if humanReadable {
		global.fmtReadable(result)
	} else {
		global.fmtNonReadable(result)
	}

	global.DeltaBytes.Store(0)
	global.Tasks.Range(func(k string, v DownloadTask) bool {
		v.ResetDelta()
		return true
	})
	global.LastTick.Store(time.Now())
	idleVal := global.DownloadedBytes.Load() + global.TotalBytes.Load()
	if global.IdleValue.Load() == idleVal {
		if global.IdleSince.Load().IsZero() {
			global.IdleSince.Store(time.Now())
		}
	} else {
		global.IdleValue.Store(idleVal)
		global.IdleSince.Store(time.Time{})
	}
	return result.String()
}

func (global *GlobalDownloadTracker) fmtNonReadable(result *strings.Builder) {
	/*
		Backend CLI

		{total: {}, tasks: {taskN: {}}}
	*/
	jsonBytes, _ := json.Marshal(global)
	result.Write(jsonBytes)
}

func (global *GlobalDownloadTracker) fmtReadable(result *strings.Builder) {
	/*
		Live CLI

		Current IP: $IP and TIME: $DATETIME ($TOTAL_THREADS threads)
		Download statistics: $TOTAL_DONE ($TOTAL_DONE_SIZE) out of $TOTAL ($TOTAL_SIZE)
		[######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)

		$FILENAME
		Download statistic: $TOTAL_DONE ($TOTAL_DONE_SIZE) out of $TOTAL ($TOTAL_SIZE)
		Total  : [######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)
		Chunk N: [######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)

	*/
	result.WriteString(fmt.Sprintf("Current IP: %s and last tick %s (%d threads)\n", global.CurrentIP.Load(), global.LastTick.Load().Format(time.RFC3339), global.TotalThreads.Load()))
	result.WriteString(fmt.Sprintf("Download statistics: %d (%s) out of %d (%s)\nOverall runtime: ", global.DownloadedFiles.Load(), humanize.Bytes(global.DownloadedBytes.Load()), global.TotalFiles.Load(), humanize.Bytes(global.TotalBytes.Load())))
	progress(result, global.DeltaBytes.Load(), global.DownloadedBytes.Load(), global.TotalBytes.Load(), 40)
	result.WriteString(fmt.Sprintf("Running %d downloads:\n", global.Tasks.Len()))

	global.Tasks.Range(func(k string, v DownloadTask) bool {
		v.Progress(result)
		return true
	})
}
