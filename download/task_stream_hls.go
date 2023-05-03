package gokhttp_download

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/BRUHItsABunny/gOkHttp/requests"
	"github.com/BRUHItsABunny/gOkHttp/responses"
	"github.com/cornelk/hashmap"
	"github.com/dustin/go-humanize"
	"github.com/etherlabsio/go-m3u8/m3u8"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StreamStats struct {
	DownloadedSegments *atomic.Uint64 `json:"downloadedSegments"`
	DownloadedDuration *atomic.Int64  `json:"downloadedDuration"`
	DownloadedBytes    *atomic.Uint64 `json:"downloadedBytes"`
}

type StreamHLSTask struct {
	// Tracking ref
	Global  *GlobalDownloadTracker    `json:"-"`
	HClient *http.Client              `json:"-"`
	ReqOpts []gokhttp_requests.Option `json:"-"`

	Muxer     *StreamMuxer `json:"-"`
	TaskStats *StreamStats `json:"taskStats"`

	TaskType     DownloadType                    `json:"taskType"`
	TaskVersion  DownloadVersion                 `json:"taskVersion"`
	FileName     *atomic.String                  `json:"fileName"`
	FileLocation *atomic.String                  `json:"fileLocation"`
	PlayListUrl  *atomic.String                  `json:"playListUrl"`
	BaseUrl      *url.URL                        `json:"-"`
	Opts         []gokhttp_requests.Option       `json:"-"`
	SegmentChan  chan *m3u8.SegmentItem          `json:"-"`
	BufferChan   chan *bytes.Buffer              `json:"-"`
	SegmentCache *hashmap.Map[string, time.Time] `json:"-"`
	SaveSegments bool                            `json:"-"`
}

func (st *StreamHLSTask) Download(ctx context.Context) error {
	errGr, ctx := errgroup.WithContext(ctx)
	errGr.Go(func() error {
		err := st.getSegments(ctx)
		if err != nil {
			return fmt.Errorf("st.getSegments: %w", err)
		}
		return nil
	})
	errGr.Go(func() error {
		err := st.downloadSegments(ctx)
		if err != nil {
			return fmt.Errorf("st.downloadSegments: %w", err)
		}
		return nil
	})
	errGr.Go(func() error {
		err := st.mergeSegments()
		if err != nil {
			return fmt.Errorf("st.mergeSegments: %w", err)
		}
		return nil
	})
	errGr.Go(func() error {
		st.cleanUp()
		return nil
	})

	// Block until done
	st.Global.Add(1)
	st.Global.TotalThreads.Inc()
	err := errGr.Wait()
	if err != nil {
		st.Global.Done()
		return fmt.Errorf("errGr.Wait: %w", err)
	}
	st.Global.Done()
	st.Global.TotalThreads.Dec()

	err = st.Muxer.F.Close()
	if err != nil {
		return fmt.Errorf("st.Muxer.F.Close: %w", err)
	}

	return nil
}

func (st *StreamHLSTask) Type() DownloadType {
	return st.TaskType
}

func (st *StreamHLSTask) Progress(sb *strings.Builder) error {
	sb.WriteString(fmt.Sprintf("%s\n", Truncate(st.FileLocation.Load(), 128, 0)))
	sb.WriteString(fmt.Sprintf("Downloading stream: %s in data, %s in playtime and %d segments\n", humanize.Bytes(st.TaskStats.DownloadedBytes.Load()), (time.Duration(st.TaskStats.DownloadedDuration.Load()) * time.Millisecond).Round(time.Second).String(), st.TaskStats.DownloadedSegments.Load()))
	return nil
}

func (st *StreamHLSTask) ResetDelta() {
	// nothing? since i don't record per-tick data?
}

func NewStreamHLSTask(global *GlobalDownloadTracker, hClient *http.Client, playlistUrl, fileLocation string, saveSegments bool, opts ...gokhttp_requests.Option) (*StreamHLSTask, error) {
	if !strings.HasSuffix(fileLocation, ".ts") {
		fileLocation += ".ts"
	}
	fileName := filepath.Base(fileLocation)

	parsedUrl, err := url.Parse(playlistUrl)
	if err != nil {
		return nil, fmt.Errorf("url.Parse: %w", err)
	}

	pathSplit := strings.Split(parsedUrl.Path, "/")
	baseUrl := &url.URL{
		Scheme:     parsedUrl.Scheme,
		Opaque:     parsedUrl.Opaque,
		User:       parsedUrl.User,
		Host:       parsedUrl.Host,
		Path:       strings.Join(pathSplit[:len(pathSplit)-1], "/"),
		OmitHost:   parsedUrl.OmitHost,
		ForceQuery: parsedUrl.ForceQuery,
		RawQuery:   parsedUrl.RawQuery,
		Fragment:   parsedUrl.Fragment,
	}

	f, err := os.OpenFile(fileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile: %w", err)
	}

	result := &StreamHLSTask{
		Global:      global,
		HClient:     hClient,
		ReqOpts:     opts,
		TaskType:    DownloadTypeLiveHLS,
		TaskVersion: DownloadVersionV1,
		Muxer:       NewStreamMuxer(f),
		TaskStats: &StreamStats{
			DownloadedSegments: atomic.NewUint64(0),
			DownloadedDuration: atomic.NewInt64(0),
			DownloadedBytes:    atomic.NewUint64(0),
		},
		FileName:     atomic.NewString(fileName),
		FileLocation: atomic.NewString(fileLocation),
		SegmentChan:  make(chan *m3u8.SegmentItem, 0),
		BufferChan:   make(chan *bytes.Buffer, 0),
		PlayListUrl:  atomic.NewString(playlistUrl),
		BaseUrl:      baseUrl,
		Opts:         opts,
		SegmentCache: hashmap.New[string, time.Time](),
		SaveSegments: saveSegments,
	}
	global.Tasks.Set(fileLocation, result)
	return result, err
}

func (st *StreamHLSTask) cleanUp() {
	ticker := time.Tick(time.Second)
	tickerClean := time.Tick(time.Minute)
	for {
		select {
		case <-ticker:
			break
		case <-tickerClean:
			st.SegmentCache.Range(func(key string, val time.Time) bool {
				if time.Minute < time.Now().Sub(val) {
					st.SegmentCache.Del(key)
				}
				return true
			})
			break
		}

		if st.Global.GraceFulStop.Load() {
			break
		}
	}
}

func (st *StreamHLSTask) getSegments(ctx context.Context) error {
	var (
		req      *http.Request
		resp     *http.Response
		respText string
		playList *m3u8.Playlist
		err      error
	)

	playlistUrl := st.PlayListUrl.Load()
	for !(playList != nil && !playList.IsMaster()) {
		req, err = gokhttp_requests.MakeGETRequest(ctx, playlistUrl, st.Opts...)
		if err != nil {
			return fmt.Errorf("requests.MakeGETRequest: %w", err)
		}
		resp, err = st.HClient.Do(req)
		if err != nil {
			return fmt.Errorf("hClient.Do: %w", err)
		}
		respText, err = gokhttp_responses.ResponseText(resp)
		if err != nil {
			return fmt.Errorf("responses.ResponseBytes: %w", err)
		}
		playList, err = m3u8.ReadString(respText)
		if err != nil {
			return fmt.Errorf("m3u8.ReadString: %w", err)
		}

		if !playList.IsMaster() {
			break
		}

		targetChunkStream := &m3u8.PlaylistItem{Bandwidth: 0}
		for _, chunkStream := range playList.Playlists() {
			if targetChunkStream.Bandwidth < chunkStream.Bandwidth {
				targetChunkStream = chunkStream
			}
		}

		if targetChunkStream.Bandwidth == 0 {
			return errors.New("stream can't have 0 bandwidth")
		}
		playlistUrl = buildURLFromBase(st.BaseUrl, targetChunkStream.URI)
	}

	ticker := time.Tick(time.Second)
	for {
		select {
		case <-ticker:
			req, err = gokhttp_requests.MakeGETRequest(ctx, playlistUrl, st.Opts...)
			if err != nil {
				return fmt.Errorf("[targetStream] requests.MakeGETRequest: %w", err)
			}
			resp, err = st.HClient.Do(req)
			if err != nil {
				return fmt.Errorf("[targetStream] hClient.Do: %w", err)
			}
			respText, err = gokhttp_responses.ResponseText(resp)
			if err != nil {
				return fmt.Errorf("[targetStream] responses.ResponseBytes: %w", err)
			}
			playList, err = m3u8.ReadString(respText)
			if err != nil {
				return fmt.Errorf("[targetStream] m3u8.ReadString: %w", err)
			}

			for _, chunk := range playList.Segments() {
				if st.Global.GraceFulStop.Load() {
					break
				}

				_, ok := st.SegmentCache.Get(chunk.Segment)
				if !ok {
					st.SegmentCache.Set(chunk.Segment, time.Now())
					st.SegmentChan <- chunk

				}
			}
			// Detect end or just stop after duration of not getting new chunks

			break
		}

		if st.Global.GraceFulStop.Load() {
			break
		}
	}

	return nil
}

func (st *StreamHLSTask) mergeSegments() error {

	isFirst := true
	ticker := time.Tick(time.Second)
	for {
		select {
		case <-ticker:
			break
		case newBuffer := <-st.BufferChan:
			if isFirst {
				err := st.Muxer.AddStreams(newBuffer)
				if err != nil {
					return fmt.Errorf("controller.Muxer.AddStreams: %w", err)
				}
				isFirst = false
			}
			err := st.Muxer.Demuxer.Input(newBuffer)
			if err != nil {
				return fmt.Errorf("controller.Muxer.Demuxer.Input: %w", err)
			}
			break
		}

		if st.Global.GraceFulStop.Load() {
			break
		}
	}

	return nil
}

func (st *StreamHLSTask) downloadSegments(ctx context.Context) error {
	ticker := time.Tick(time.Second)
	for {
		select {
		case <-ticker:
			break
		case newChunk := <-st.SegmentChan:
			req, err := gokhttp_requests.MakeGETRequest(ctx, buildURLFromBase(st.BaseUrl, newChunk.Segment), st.Opts...)
			if err != nil {
				return fmt.Errorf("requests.MakeGETRequest: %w", err)
			}
			resp, err := st.HClient.Do(req)
			if err != nil {
				return fmt.Errorf("st.HClient.Do: %w", err)
			}
			respBytes, err := gokhttp_responses.ResponseBytes(resp)
			if err != nil {
				return fmt.Errorf("responses.ResponseBytes: %w", err)
			}

			buf := bytes.NewBuffer(respBytes)

			st.TaskStats.DownloadedSegments.Inc()
			st.TaskStats.DownloadedBytes.Add(uint64(buf.Len()))
			st.TaskStats.DownloadedDuration.Add(int64(newChunk.Duration * 1500))
			st.Global.TotalBytes.Add(uint64(buf.Len()))
			st.Global.DownloadedBytes.Add(uint64(buf.Len()))

			st.BufferChan <- buf
			if st.SaveSegments {
				fileLocation := filepath.Dir(st.FileLocation.Load())
				f, err := os.OpenFile(fileLocation+newChunk.Segment, os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					return fmt.Errorf("os.OpenFile: %w", err)
				}
				_, err = f.Write(respBytes)
				if err != nil {
					return fmt.Errorf("f.Write: %w", err)
				}
			}
			break
		}

		if st.Global.GraceFulStop.Load() {
			break
		}
	}

	return nil
}

func buildURLFromBase(baseUrl *url.URL, path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}

	baseUrlClone, _ := url.Parse(baseUrl.String())
	baseUrlClone.Path += "/" + path
	return baseUrlClone.String()
}
