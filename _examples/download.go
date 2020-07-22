package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"github.com/dustin/go-humanize"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"time"
)

func main() {
	url2 := "https://speed.hetzner.de/100MB.bin"
	httpTimeout := time.Second * time.Duration(3)
	//proxyUrl, _ := url.Parse("http://127.0.0.1:8888")
	gOkHttpOptions := gokhttp.HttpClientOptions{
		JarOptions: &cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: false, Filename: ".cookies", EncryptionPassword: ""},
		Transport: &http.Transport{
			TLSHandshakeTimeout: httpTimeout,
			DisableCompression:  false,
			DisableKeepAlives:   false,
			//Proxy:               http.ProxyURL(proxyUrl),
		},
		RefererOptions: &gokhttp.RefererOptions{Update: true, Use: true},
		Timeout:        &httpTimeout,
	}
	client := gokhttp.GetHTTPClient(&gOkHttpOptions)
	tracker := gokhttp.DownloadTracker{URL: url2, FileName: "test.bin", Started: time.Now(), ThreadCount: 4}
	err := client.DownloadThreaded(&tracker, gokhttp.EmptyMap, gokhttp.EmptyMap)
	if err == nil {
		writtenOld := int64(0)
		writtenNew := int64(0)
		done := false
		for {
			time.Sleep(time.Second)
			writtenNew, done = tracker.IsDone()
			speed := writtenNew - writtenOld
			fmt.Println("Downloaded ", humanize.Bytes(uint64(writtenNew)), " out of ", humanize.Bytes(uint64(tracker.FileSize)), "@", humanize.Bytes(uint64(speed)), "/s")
			writtenOld = writtenNew
			if done {
				break
			}
		}
	}

	fmt.Println(err)
}
