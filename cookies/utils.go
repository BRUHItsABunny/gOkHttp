package cookies

import (
	"net/http"
	"net/url"
)

func SaveCookies(client *http.Client) error {
	return client.Jar.(*Jar).Save()
}

func SetCookiesWithURLS(client *http.Client, cookieMap map[*url.URL][]*http.Cookie) error {
	for uri, cookieSlice := range cookieMap {
		client.Jar.SetCookies(uri, cookieSlice)
	}
	return SaveCookies(client)
}

func SetCookiesWithStrings(client *http.Client, cookieMap map[string][]*http.Cookie) error {
	var (
		urlObj *url.URL
		err    error
	)
	for uri, cookieSlice := range cookieMap {
		urlObj, err = url.Parse(uri)
		if err != nil {
			return err
		}
		client.Jar.SetCookies(urlObj, cookieSlice)
	}
	return SaveCookies(client)
}
