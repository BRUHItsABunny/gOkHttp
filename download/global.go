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

func NewGlobalDownloadController() *GlobalDownloadController {
	return &GlobalDownloadController{
		WaitGroup:       sync.WaitGroup{},
		GraceFulStop:    atomic.NewBool(false),
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

func (global *GlobalDownloadController) PollIP(hClient *http.Client) {
	for {
		time.Sleep(time.Minute)
		if global.GraceFulStop.Load() {
			break
		}
		global.CurrentIP.Store(GetCurrentIPAddress(hClient))
	}
}
