package main

import (
	"fmt"
	"gokhttp"
	"gokhttp/cookies"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"time"
)

func main() {
	gOkHttpPinner := gokhttp.GetSSLPinner()
	err := gOkHttpPinner.AddPin("google.com", false, "sha256\\BOGUS") // using sha256\\f8NnEFZxQ4ExFOhSN7EiFWtiudZQVD2oY60uauV/n78= will yield actual HTML code
	gOkHttpOptions := gokhttp.HttpClientOptions{
		JarOptions: &cookies.JarOptions{PublicSuffixList: publicsuffix.List, NoPersist: true},
		Transport: &http.Transport{
			TLSHandshakeTimeout: 15 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
		RefererOptions:    &gokhttp.RefererOptions{Update: false, Use: false},
		SSLPinningOptions: &gOkHttpPinner,
	}
	gOkHttpClient := gokhttp.GetHTTPClient(&gOkHttpOptions)
	if err == nil {
		req, err := gOkHttpClient.MakeGETRequest("https://google.com", map[string]string{}, map[string]string{})
		if err == nil {
			response, err := gOkHttpClient.Do(req)
			if err == nil {
				body, err := response.Bytes()
				if err == nil {
					fmt.Println(string(body))
				} else {
					fmt.Println(err)
				}
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
}