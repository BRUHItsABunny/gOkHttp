package client

import (
	"github.com/BRUHItsABunny/gOkHttp/cookies"
	"net/http"
)

type JarOption struct {
	Jar *cookies.CookieJarWrapper
}

func NewJarOption(jar *cookies.CookieJarWrapper) *JarOption {
	return &JarOption{Jar: jar}
}

func (o *JarOption) Execute(client *http.Client) error {
	client.Jar = o.Jar
	return nil
}
