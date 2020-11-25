package gokhttp

import (
	"context"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"net"
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	Client         *http.Client
	RefererOptions RefererOptions
	Headers        map[string]string
	Context        *context.Context
	CancelF        context.CancelFunc
	ClientOptions  *HttpClientOptions
}

type RefererOptions struct {
	Update bool
	Use    bool
	Value  string
}

type HttpClientOptions struct {
	JarOptions        *cookies.JarOptions
	Transport         *http.Transport
	Timeout           *time.Duration
	SSLPinningOptions *SSLPinner
	RefererOptions    *RefererOptions
	RedirectPolicy    func(req *http.Request, via []*http.Request) error
	Headers           map[string]string
	Context           *context.Context
	CancelF           context.CancelFunc
}

type HttpResponse struct {
	*http.Response
}

type HttpJSONResponse struct {
	data []byte
}

type Dialer func(ctx context.Context, network, addr string) (net.Conn, error)

type SSLPin struct {
	SkipCA    bool
	Pins      []string //sha256
	Hostname  string
	Algorithm string
}

type SSLPinner struct {
	SSLPins map[string]SSLPin
}

type Task struct {
	Id              int
	Name            string
	Total           uint64 // Whole file range downloaded as of yet
	Expected        uint64 // Whole file size
	FragSize        uint64 // If it's a fragment, the total size of the fragment
	CanResume       bool
	Threads         int
	URL             string
	ThreadObjects   []Task
	IsThread        bool
	ProgressChannel chan *TrackerMessage
}

type TrackerMessage struct {
	Total      uint64
	Delta      int
	Status     int
	Name       string
	Err        error
	Expected   uint64
	Id         int
	IsFragment bool
}

func (t *Task) Write(p []byte) (int, error) {
	n := len(p)
	t.Total += uint64(n)
	t.Publish(n)
	return n, nil
}

func (t *Task) Publish(n int) {
	t.ProgressChannel <- &TrackerMessage{Total: t.Total, Delta: n, Name: t.Name, Id: t.Id, Expected: t.Expected}
	// fmt.Println(&t, t.Name + " sent report")
}

const (
	StatusStart = iota
	StatusProgress
	StatusError
	StatusDone
	StatusMerging
)

type Download struct {
	FileName string
	Size     int
	Progress int
	Delta    int
	Started  time.Time
	Finished time.Time
	Threaded bool
	Status   int
	// Threads
	Fragments map[string]Download
}

type M3U8Index struct {
	PlayLists []M3U8PlayList
	RawValues map[string]string
}

type M3U8PlayList struct {
	URL         string
	Bandwidth   int
	ResolutionX int
	ResolutionY int
	Codecs      []string
	Framerate   int
	RawValues   map[string]string
}

type M3U8FileList struct {
	Files     []M3U8File
	RawValues map[string]string
}

type M3U8File struct {
	Duration float64
	Location string
}

type DownloadReporter interface {
	OnDone(message *TrackerMessage) error
	OnStart(message *TrackerMessage) error
	OnProgress(message *TrackerMessage) error
	OnError(message *TrackerMessage) error
	OnMerging(message *TrackerMessage) error
	OnTick() error
	Processor() error
}
type DefaultDownloadReporter struct {
	Ticker     *time.Ticker
	Chan       chan *TrackerMessage
	InProgress map[int]Download
	Done       map[int]Download
	ShouldStop bool
	WaitTime   int
	WaitGroup  *sync.WaitGroup
}
