package gokhttp_download

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	crypto_utils "github.com/BRUHItsABunny/crypto-utils"
	gokhttp_client "github.com/BRUHItsABunny/gOkHttp/client"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func hashSHA(fileLocation, expectedHash string) (string, error) {
	f, err := os.Open(fileLocation)
	if err != nil {
		return "", fmt.Errorf(": %w", err)
	}
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf(": %w", err)
	}

	hashBytes := crypto_utils.SHA256hash(fileBytes)
	hash := hex.EncodeToString(hashBytes)
	if hash != expectedHash {
		return hash, errors.New("hash not match")
	}
	return hash, nil
}

func initTestEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("godotenv.Load: %w", err))
	}
}

func initHClient() (*http.Client, error) {
	clientOpts := []gokhttp_client.Option{}
	if os.Getenv("USE_PROXY") == "true" {
		clientOpts = append(clientOpts, gokhttp_client.NewProxyOption(os.Getenv("OPT_PROXY")))
	}
	return gokhttp_client.NewHTTPClient(clientOpts...)
}

func TestThreadedDownloadTask_Download(t *testing.T) {
	initTestEnv()
	hClient, err := initHClient()
	if err != nil {
		t.Error(err)
	}

	headers := http.Header{
		"User-Agent": {"bunny"},
	}
	values := url.Values{
		"param1": {"bunny"},
	}
	reqOpts := []gokhttp_requests.Option{
		gokhttp_requests.NewHeaderOption(headers),
		gokhttp_requests.NewURLParamOption(values),
	}

	global := NewGlobalDownloadTracker(time.Second * time.Duration(3))
	global.PollIP(hClient)
	task, err := NewThreadedDownloadTask(context.Background(), hClient, global, os.Getenv("THREADED_TEST_NAME"), os.Getenv("THREADED_TEST"), 3, 0, reqOpts...)
	if err != nil {
		t.Error(err)
	}

	go func() {
		err := task.Download(context.Background())
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

	hash, err := hashSHA(os.Getenv("THREADED_TEST_NAME"), os.Getenv("THREADED_SHA_256"))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(fmt.Sprintf("hash: %s", hash))

}

func TestStreamHLSTask_Download(t *testing.T) {
	initTestEnv()
	hClient, err := initHClient()
	if err != nil {
		t.Error(err)
	}

	global := NewGlobalDownloadTracker(time.Second * time.Duration(10))
	global.PollIP(hClient)
	task, err := NewStreamHLSTask(global, hClient, os.Getenv("HLS_TEST"), fmt.Sprintf("%s_%d.ts", os.Getenv("HLS_TEST_NAME"), time.Now().Unix()), os.Getenv("HLS_SAVE_SEGMENTS") == "true")
	if err != nil {
		t.Error(err)
	}

	go func() {
		err := task.Download(context.Background())
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		ticker := time.Tick(time.Second)
		timeout := time.Tick(time.Duration(2) * time.Minute)
		for {
			select {
			case <-ticker:
				break
			case <-timeout:
				global.GraceFulStop.Store(true)
				break
			}

			if global.GraceFulStop.Load() {
				break
			}
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

func TestStreamMuxer(t *testing.T) {
	// TODO: Collect segments and a glitching final muxed result.
}
