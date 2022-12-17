package cookies

import (
	"fmt"
	crypto_utils "github.com/BRUHItsABunny/crypto-utils"
	"github.com/BRUHItsABunny/crypto-utils/padding"
	"github.com/BRUHItsABunny/gOkHttp/constants"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type CookieJarWrapper struct {
	sync.RWMutex
	jarKeys       map[*url.URL]struct{}
	jar           *cookiejar.Jar
	f             *os.File
	encryptionKey string
}

func NewCookieJar(fileLocation, encryptionKey string, psl cookiejar.PublicSuffixList) (*CookieJarWrapper, error) {
	var f *os.File

	if len(fileLocation) > 0 {
		dirName := filepath.Dir(fileLocation)
		err := os.MkdirAll(dirName, 0600)
		if err != nil {
			return nil, fmt.Errorf("os.MkdirAll: %w", err)
		}

		f, err = os.OpenFile(fileLocation, os.O_CREATE, 0600)
		if err != nil {
			return nil, fmt.Errorf("os.OpenFile: %w", err)
		}
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: psl})
	if err != nil {
		return nil, fmt.Errorf("cookiejar.New: %w", err)
	}

	return &CookieJarWrapper{
		RWMutex:       sync.RWMutex{},
		jarKeys:       map[*url.URL]struct{}{},
		jar:           jar,
		f:             f,
		encryptionKey: encryptionKey,
	}, nil
}

// Interface functions, required for *http.Client integration

func (j *CookieJarWrapper) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.Lock()
	j.jarKeys[u] = struct{}{}
	j.Unlock()
	j.jar.SetCookies(u, cookies)
}

func (j *CookieJarWrapper) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}

// Utility functions
/*
	Persistence
	Some design choices that were made in comparison to the previous implementation
	1. I don't have a use case for a cookie file that may get used by multiple processes at the same time
	2. Encryption algo is not exposed for modification to keep configuration simple
	3. JSON is horrible to read write data in that essentially only gets read by a machine (if it needs to be humanreadable JSON makes sense, this is not the case here)
*/

func (j *CookieJarWrapper) Save() error {
	return j.save()
}

func (j *CookieJarWrapper) save() error {
	if j.f == nil {
		return constants.ErrCookieJarNotPersistent
	}

	entries := map[string][]*http.Cookie{}
	for u, _ := range j.jarKeys {
		entries[u.String()] = j.Cookies(u)
	}
	data, err := msgpack.Marshal(&entries)
	if err != nil {
		return fmt.Errorf("msgpack.Marshal: %w", err)
	}

	// Reset file
	err = j.f.Truncate(0)
	if err != nil {
		return fmt.Errorf("j.f.Truncate: %w", err)
	}
	_, err = j.f.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("j.f.Seek: %w", err)
	}

	// Encrypt
	if j.encryptionKey != "" {
		key, iv := crypto_utils.SHA256hash([]byte(j.encryptionKey)), crypto_utils.MD5hash([]byte(j.encryptionKey))
		data, err = crypto_utils.AesCBCEncrypt(&padding.PKCS7Padding{}, data, key, iv)
		if err != nil {
			return fmt.Errorf("crypto_utils.AesCBCEncrypt: %w", err)
		}
	}

	_, err = j.f.Write(data)
	if err != nil {
		return fmt.Errorf("j.f.Write: %w", err)
	}
	return nil
}

func (j *CookieJarWrapper) Load() error {
	return j.load()
}

func (j *CookieJarWrapper) load() error {
	if j.f == nil {
		return constants.ErrCookieJarNotPersistent
	}

	// Read from file
	_, err := j.f.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("j.f.Seek: %w", err)
	}
	data, err := io.ReadAll(j.f)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	// Decrypt
	if j.encryptionKey != "" {
		key, iv := crypto_utils.SHA256hash([]byte(j.encryptionKey)), crypto_utils.MD5hash([]byte(j.encryptionKey))
		data, err = crypto_utils.AesCBCDecrypt(&padding.PKCS7Padding{}, data, key, iv)
		if err != nil {
			return fmt.Errorf("crypto_utils.AesCBCDecrypt: %w", err)
		}
	}

	entries := map[string][]*http.Cookie{}
	err = msgpack.Unmarshal(data, &entries)
	if err != nil {
		return fmt.Errorf("msgpack.Marshal: %w", err)
	}

	for key, cooks := range entries {
		uKey, err := url.Parse(key)
		if err != nil {
			return fmt.Errorf("url.Parse: %w", err)
		}
		j.SetCookies(uKey, cooks)
	}
	return nil
}
