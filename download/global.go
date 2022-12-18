package download

import (
	"context"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/BRUHItsABunny/gOkHttp/responses"
	"github.com/cornelk/hashmap"
	"go.uber.org/atomic"
	"net/http"
	"sync"
	"time"
)

type GlobalDownloadController struct {
	// Synchronization
	sync.WaitGroup
	GraceFulStop *atomic.Bool `json:"-"`
	// Idle stats
	IdleSince   *atomic.Time   `json:"idleSince"`
	IdleValue   *atomic.Uint64 `json:"-"`
	IdleTimeout *atomic.Int64  `json:"-"`
	// MetaData
	LastTick     *atomic.Time                                  `json:"lastTick"`
	CurrentIP    *atomic.String                                `json:"currentIP"`
	TotalThreads *atomic.Uint64                                `json:"totalThreads"`
	Tasks        *hashmap.Map[string, *DownloadTaskController] `json:"tasks"`
	// Total stats
	TotalFiles *atomic.Uint64 `json:"totalFiles"`
	TotalBytes *atomic.Uint64 `json:"totalBytes"`
	// Download stats
	DownloadedFiles *atomic.Uint64 `json:"downloadedFiles"`
	DownloadedBytes *atomic.Uint64 `json:"downloadedBytes"`
	// Per tick stats
	DeltaBytes *atomic.Uint64 `json:"deltaBytes"`
}

func NewGlobalDownloadController(idleTimeout time.Duration) *GlobalDownloadController {
	return &GlobalDownloadController{
		WaitGroup:       sync.WaitGroup{},
		GraceFulStop:    atomic.NewBool(false),
		IdleValue:       atomic.NewUint64(0),
		IdleSince:       atomic.NewTime(time.Time{}),
		IdleTimeout:     atomic.NewInt64(int64(idleTimeout)),
		LastTick:        atomic.NewTime(time.Now()),
		CurrentIP:       atomic.NewString(""),
		TotalThreads:    atomic.NewUint64(0),
		Tasks:           hashmap.New[string, *DownloadTaskController](),
		TotalFiles:      atomic.NewUint64(0),
		TotalBytes:      atomic.NewUint64(0),
		DownloadedFiles: atomic.NewUint64(0),
		DownloadedBytes: atomic.NewUint64(0),
		DeltaBytes:      atomic.NewUint64(0),
	}
}

func GetCurrentIPAddress(hClient *http.Client) (ip string) {
	req, err := requests.MakeGETRequest(context.Background(), "https://httpbin.org/get")
	if err != nil {
		return
	}
	resp, err := hClient.Do(req)
	if err != nil {
		return
	}
	data := map[string]any{}
	err = responses.ResponseJSON(resp, &data)

	if err != nil {
		return
	}
	ipPreCast, ok := data["origin"]
	if ok {
		ip = ipPreCast.(string)
	}
	return
}

func (global *GlobalDownloadController) Stop() {
	global.GraceFulStop.Store(true)
	global.Wait()
}

func (global *GlobalDownloadController) IdleTimeoutExceeded() bool {
	idleSince := global.IdleSince.Load()
	idleTimeout := global.IdleTimeout.Load()
	if !idleSince.IsZero() {
		return time.Now().Sub(idleSince) >= time.Duration(idleTimeout)
	}
	return false
}

func (global *GlobalDownloadController) PollIP(hClient *http.Client) {
	go global.pollIP(hClient)
}

func (global *GlobalDownloadController) pollIP(hClient *http.Client) {
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
