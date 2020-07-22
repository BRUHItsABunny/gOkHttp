package gokhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
)

func (c *HttpClient) MakeHEADRequest(URL string, parameters, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "HEAD", URL, nil)
	} else {
		req, err = http.NewRequest("HEAD", URL, nil)
	}
	req = c.readyRequest(req)
	if checkError(err) {
		query := req.URL.Query()
		for k, v := range parameters {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		return req, nil
	}
	return nil, err
}

func (c HttpClient) CheckResource(task *Task, parameters, headers map[string]string) *Task {
	// Send HEAD request if task isn't a thread
	if !task.IsThread {
		req, err := c.MakeHEADRequest(task.URL, parameters, headers)
		resp, err := c.Client.Do(req)
		if err == nil {
			// Check for Content-Length and Range support
			fileSize := resp.ContentLength
			supportsRanges := false
			if resp.Header.Get("Accept-Ranges") == "bytes" {
				supportsRanges = true
			}
			if resp.Header.Get("Ranges-Supported") == "bytes" {
				supportsRanges = true
			}
			task.CanResume = supportsRanges
			task.Expected = uint64(fileSize)
			// If its threaded AND ranges are supported, proceed with threaded download
			if task.Threads > 1 && task.CanResume {
				// Check for ALL existing fragments
				result := make([]Task, 0)
				for thread := 0; thread < task.Threads; thread++ {
					// Setup sub-task to track progress
					progressChan := make(chan *TrackerMessage)
					threadTask := Task{
						Id:              task.Id,
						Name:            task.Name + ".fragment." + strconv.Itoa(thread),
						Total:           0,
						Expected:        0,
						CanResume:       task.CanResume,
						Threads:         1,
						IsThread:        true,
						URL:             task.URL,
						ProgressChannel: progressChan,
					}
					// Check what the start and end range is for this task
					start := int(task.Expected) / task.Threads * thread
					var stop int
					if thread == task.Threads {
						stop = int(task.Expected)
					} else {
						stop = int(task.Expected)/task.Threads*(thread+1) - 1
					}
					threadTask.Total = uint64(start)
					threadTask.Expected = uint64(stop)
					threadTask.FragSize = uint64(stop - start)
					// Check if file exists and update task with existing progress accordingly
					info, exists := fileExists(threadTask.Name + ".tmp")
					if exists {
						threadTask.Total += uint64(info.Size())
					}
					// Add to result
					result = append(result, threadTask)
				}
				task.ThreadObjects = result
			} else {
				// Check if file exists, if it does check if the length is below the Content-Length
				info, exists := fileExists(task.Name + ".tmp")
				if exists && task.CanResume {
					task.Total = uint64(info.Size())
				}
			}
		}
	}
	return task
}

func (c *HttpClient) DownloadFile(task *Task, parameters, headers map[string]string) error {
	var err error
	// BOOTSTRAP
	realName := task.Name
	realId := task.Id
	progressChan := task.ProgressChannel
	// Make DownloadClient, it has no HTTP timeout
	client := GetHTTPDownloadClient(c.ClientOptions)
	task = c.CheckResource(task, parameters, headers)
	taskTotal := task.Total
	taskExpected := task.Expected
	totalDelta := int64(taskExpected - taskTotal)
	realDelta := int64(taskExpected)
	progressChan <- &TrackerMessage{Total: taskTotal, Name: realName, Id: realId, Expected: taskExpected, Delta: 0, Status: StatusStart}
	_, exists := fileExists(task.Name)
	if !exists {
		if task.Threads == 1 {
			var resp *HttpResponse
			// Make the Request
			req := client.makeDownloadRequest(task, parameters, headers)
			// Start Downloading
			resp, err = client.Do(req)
			if err == nil {
				// Open file, if non-existent then create file
				f, err := os.OpenFile(realName+".tmp", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
				if err == nil {
					// Start reading and report back on each read
					var msg *TrackerMessage
					for {
						written, err := io.CopyN(f, resp.Body, 100)
						realDelta += written
						msg = &TrackerMessage{Total: uint64(realDelta), Name: realName, Id: task.Id, Expected: uint64(totalDelta), Delta: int(written), Status: StatusProgress}
						if err != nil {
							msg.Status = StatusError
							msg.Err = err
							break
						}
						progressChan <- msg
					}
					// Close file and response after the loop finishes
					err = f.Close()
					_ = resp.Body.Close()
					if realDelta >= totalDelta {
						// Only true when done
						msg.Status = StatusDone
						msg.Err = nil
						err = os.Rename(f.Name(), realName)
					}
					progressChan <- msg
				}
			}
		} else {
			// Aggregate the progress channels
			wg := sync.WaitGroup{}
			go Aggregate(task, &wg)
			// Start all threads with their respective ranges
			ready := make(chan bool)
			for _, threadTask := range task.ThreadObjects {
				// Start GoRoutine
				go c.StartThread(&threadTask, ready, parameters, headers)
				<-ready
			}
			// Block until finish based on Aggregate thread
			wg.Wait()
		}
	} else {
		err = errors.New("file exists already")
	}

	return err
}

func Aggregate(task *Task, wg *sync.WaitGroup) {
	wg.Add(1)
	done := 0
	cases := make([]reflect.SelectCase, len(task.ThreadObjects))
	for i, threadTask := range task.ThreadObjects {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(threadTask.ProgressChannel)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			// The chosen channel has been closed, so zero out the channel to disable the case
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		}
		msg := value.Interface().(*TrackerMessage)
		task.ProgressChannel <- msg
		// Track done's
		if msg.Status == StatusDone {
			done += 1
			if done == task.Threads {
				// All threads are done, merge the files and stop this goroutine
				task.ProgressChannel <- &TrackerMessage{Name: task.Name, Id: task.Id, Status: StatusMerging}
				var (
					realFile, fragFile *os.File
					err                error
				)
				for i, threadTask := range task.ThreadObjects {
					if i == 0 {
						// Open the first fragment
						realFile, err = os.OpenFile(threadTask.Name, os.O_APPEND, 0600)
						if err == nil {
							continue
						} else {
							// FATAL ERROR
							panic(err)
						}
					}
					// Append the other fragments
					fragFile, err = os.Open(threadTask.Name)
					if err != nil {
						// FATAL ERROR
						panic(err)
					} else {
						_, _ = io.Copy(realFile, fragFile)
						// Delete appended fragments
						fragFile.Close()
						os.Remove(threadTask.Name)
					}
				}
				// Close and rename first fragment
				realFile.Close()
				os.Rename(task.ThreadObjects[0].Name, task.Name)
				// Final DONE message
				msg = &TrackerMessage{
					Total:    task.Expected,
					Delta:    0,
					Status:   StatusDone,
					Name:     task.Name,
					Err:      nil,
					Expected: task.Expected,
					Id:       task.Id,
				}
				task.ProgressChannel <- msg
			}
		}
	}
	wg.Done()
}

func (c *HttpClient) StartThread(task *Task, started chan bool, parameters, headers map[string]string) error {
	var err error
	// BOOTSTRAP
	realName := task.Name
	realId := task.Id
	progressChan := task.ProgressChannel
	// Make DownloadClient, it has no HTTP timeout
	client := GetHTTPDownloadClient(c.ClientOptions)
	task = c.CheckResource(task, parameters, headers)
	taskTotal := task.Total
	taskExpected := task.Expected
	totalDelta := int64(taskExpected - taskTotal)
	realDelta := int64(task.FragSize) - totalDelta
	progressChan <- &TrackerMessage{Total: uint64(realDelta), Name: realName, Id: realId, Expected: task.FragSize, Delta: 0, IsFragment: true, Status: StatusStart}
	_, exists := fileExists(task.Name)
	if !exists {
		var resp *HttpResponse
		// Make the Request
		req := client.makeDownloadRequest(task, parameters, headers)
		started <- true
		// Start Downloading
		resp, err = client.Do(req)
		if err == nil {
			// Open file, if non-existent then create file
			f, err := os.OpenFile(realName+".tmp", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err == nil {
				// Start reading and report back on each read
				var msg *TrackerMessage
				for {
					written, err := io.CopyN(f, resp.Body, 100)
					realDelta += written
					msg = &TrackerMessage{Total: uint64(realDelta), Name: realName, Id: task.Id, Expected: uint64(totalDelta), Delta: int(written), IsFragment: true, Status: StatusProgress}
					if err != nil {
						msg.Status = StatusError
						msg.Err = err
						break
					}
					progressChan <- msg
				}
				// Close file and response after the loop finishes
				err = f.Close()
				_ = resp.Body.Close()
				if realDelta >= totalDelta {
					// Only true when done
					msg.Status = StatusDone
					msg.Err = nil
					err = os.Rename(f.Name(), realName)
				}
				progressChan <- msg
			}
		}
	} else {
		started <- true
		err = errors.New("file exists already")
	}

	return err
}

func (c *HttpClient) makeDownloadRequest(task *Task, parameters, headers map[string]string) *http.Request {
	var (
		err error
		req *http.Request
	)
	// Set context
	if c.Context != nil {
		req, err = http.NewRequestWithContext(*c.Context, "GET", task.URL, nil)
	} else {
		req, err = http.NewRequest("GET", task.URL, nil)
	}
	if checkError(err) {
		// Ready from params, cookies and headers stored within the Client
		req = c.readyRequest(req)
		// Ready from params and headers from arguments
		query := req.URL.Query()
		for k, v := range parameters {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		if task.Expected != 0 {
			byteRange := fmt.Sprintf("bytes=%d-%d", task.Total, task.Expected)
			req.Header.Add("Range", byteRange)
		}
		return req
	}
	return nil
}

/*
TODO:
	1. [X] Make single threaded download work
		2. [X] First check if file with X filename exists
		3. [X] Save the amount of bytes already downloaded
		4. [X] Visit URL and check response header for Last-Modified to be BEFORE your file's last modified (if, cool else redownload) and presence of Accept-Ranges: bytes (if, resume else, redownload)
		5. [X] Download in a way that allows you to track things such as SPEED and PROGRESS
	2. [X] Make multi threaded download work
		1. [X] Visit URL and check presence of Accept-Ranges: bytes (if, cool else, single threaded)
		2. [X] Download in a way that allows you to track things such as SPEED and PROGRESS PER THREAD (aggregate?)
		3. [X] Merge all chunks, BUT should I keep the chunks in memory at all or write and merge from disk WASTING DISK SPACE during process?
	3. [X] Confirm all features work.
*/
