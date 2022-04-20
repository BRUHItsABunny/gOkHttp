package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/url"
	"time"
)

func main() {
	url2 := "https://speed.hetzner.de/100MB.bin"
	httpTimeout := time.Second * time.Duration(3)
	// proxyUrl, _ := url.Parse("http://127.0.0.1:8888")
	gOkHttpOptions := gokhttp.HttpClientOptions{
		JarOptions: &cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: false, Filename: ".cookies", EncryptionPassword: ""},
		Transport: &http.Transport{
			TLSHandshakeTimeout: httpTimeout,
			DisableCompression:  false,
			DisableKeepAlives:   false,
			// Proxy:               http.ProxyURL(proxyUrl),
		},
		RefererOptions: &gokhttp.RefererOptions{Update: true, Use: true},
		Timeout:        &httpTimeout,
	}
	client := gokhttp.GetHTTPDownloadClient(&gOkHttpOptions)
	// TODO: Different how, this host fluctuates too much
	// 1 thread = 5-8MB/s
	// 4 threads = 20MB/s and one chunk is always progressing faster than others (2minsish)
	// 4 threads + lockOSThread = 20MB/s ish but drops off fast even before the others finish? (last chunk downloads at less than 1mb/s) ()
	task := gokhttp.NewTask(url2, "test.bin", 4, true)

	go client.DownloadFileV2(task, url.Values{}, gokhttp.EmptyMap)
	for {

		at := <-time.After(time.Second)
		progress, err := task.Render(at)

		if gokhttp.TaskStatus(progress.Status) >= gokhttp.TaskStatusDone || err != nil {
			fmt.Println(err)
			break
		}
	}
}
