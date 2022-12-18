package download

import (
	"context"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/client"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func Test_Download(t *testing.T) {
	clientOpts := []client.Option{
		// client.NewProxyOption("http://127.0.0.1:8888"),
	}
	hClient, err := client.NewHTTPClient(clientOpts...)
	if err != nil {
		t.Error(err)
	}

	headers := http.Header{
		"User-Agent": {"bunny"},
	}
	values := url.Values{
		"param1": {"bunny"},
	}
	reqOpts := []requests.Option{
		requests.NewHeaderOption(headers),
		requests.NewURLParamOption(values),
	}

	fileURL := "http://ipv4.download.thinkbroadband.com/200MB.zip"
	global := NewGlobalDownloadController(time.Second * time.Duration(3))
	global.PollIP(hClient)
	task, err := NewDownloadTaskController(hClient, global, "200MB.zip", "200MB.zip", fileURL, 3, 0, reqOpts...)
	if err != nil {
		t.Error(err)
	}

	go func() {
		err := DownloadTask(context.Background(), global, task, hClient, reqOpts...)
		if err != nil {
			t.Error(err)
		}
	}()
	global.TotalFiles.Inc()
	for {
		if global.GraceFulStop.Load() || global.IdleTimeoutExceeded() {
			global.Stop()
			break
		}
		fmt.Println(global.Tick(true))
		time.Sleep(time.Second)
	}
}

func Test_DownloadDaemon(t *testing.T) {
	clientOpts := []client.Option{
		// client.NewProxyOption("http://127.0.0.1:8888"),
	}
	hClient, err := client.NewHTTPClient(clientOpts...)
	if err != nil {
		t.Error(err)
	}

	headers := http.Header{
		"User-Agent": {"bunny"},
	}
	values := url.Values{
		"param1": {"bunny"},
	}
	reqOpts := []requests.Option{
		requests.NewHeaderOption(headers),
		requests.NewURLParamOption(values),
	}

	fileURL := "http://ipv4.download.thinkbroadband.com/200MB.zip"
	global := NewGlobalDownloadController(time.Second * time.Duration(3))
	task, err := NewDownloadTaskController(hClient, global, "200MB.zip", "200MB.zip", fileURL, 3, 0, reqOpts...)
	if err != nil {
		t.Error(err)
	}

	go func() {
		err := DownloadTask(context.Background(), global, task, hClient, reqOpts...)
		if err != nil {
			t.Error(err)
		}
	}()
	global.TotalFiles.Inc()
	for {
		if global.GraceFulStop.Load() || global.IdleTimeoutExceeded() {
			global.Stop()
			break
		}
		fmt.Println(global.Tick(false))
		time.Sleep(time.Second)
	}
}

func Test_DownloadIdle(t *testing.T) {
	global := NewGlobalDownloadController(time.Minute)
	for {
		if global.GraceFulStop.Load() || global.IdleTimeoutExceeded() {
			global.Stop()
			break
		}
		fmt.Println(global.Tick(true))
		time.Sleep(time.Second)
	}
}
