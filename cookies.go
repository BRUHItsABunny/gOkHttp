package gokhttp

import (
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"net/http"
	"net/url"
)

func (c *HttpClient) SaveCookies() error {
	// Copied from GokHTTP library's .Do(), maybe add an easier way to access, save and set the cookies in GokHTTP
	return c.Client.Jar.(*cookies.Jar).Save()
}

func (c *HttpClient) SetCookiesWithURLS(cookieMap map[*url.URL][]*http.Cookie) error {
	for uri, cookieSlice := range cookieMap {
		c.Client.Jar.SetCookies(uri, cookieSlice)
	}
	return c.SaveCookies()
}

func (c *HttpClient) SetCookiesWithStrings(cookieMap map[string][]*http.Cookie) error {
	var (
		urlObj *url.URL
		err    error
	)
	for uri, cookieSlice := range cookieMap {
		urlObj, err = url.Parse(uri)
		if err != nil {
			return err
		}
		c.Client.Jar.SetCookies(urlObj, cookieSlice)
	}
	return c.SaveCookies()
}
