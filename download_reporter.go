package gokhttp

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"math"
	"sync"
	"time"
)

func GetDefaultDownloadReporter(channel chan *TrackerMessage, wg *sync.WaitGroup) *DefaultDownloadReporter {
	return &DefaultDownloadReporter{Ticker: time.NewTicker(time.Second), Chan: channel, WaitGroup: wg, InProgress: map[int]Download{}, Done: map[int]Download{}}
}

func (dr *DefaultDownloadReporter) OnDone(message *TrackerMessage) error {
	var (
		err      error
		download Download
	)

	if message.IsFragment {
		download = dr.InProgress[message.Id].Fragments[message.Name]
		download.Status = message.Status
		download.Finished = time.Now()
		dr.InProgress[message.Id].Fragments[message.Name] = download
	} else {
		// Main is done, if threaded merging is also done
		download = dr.InProgress[message.Id]
		download.Status = message.Status
		download.Finished = time.Now()
		dr.Done[message.Id] = download
		delete(dr.InProgress, message.Id)
	}

	return err
}

func (dr *DefaultDownloadReporter) OnStart(message *TrackerMessage) error {
	var (
		err      error
		download Download
	)

	// Build the DB!
	if message.IsFragment {
		// Fragment Started
		download = Download{}
		download.Status = message.Status
		download.FileName = message.Name
		download.Started = time.Now()
		download.Size = int(message.Expected)
		download.Progress = int(message.Total)
		dr.InProgress[message.Id].Fragments[message.Name] = download
		// Make sure main object knows it is threaded
		download = dr.InProgress[message.Id]
		download.Progress += int(message.Total)
		download.Threaded = true
		dr.InProgress[message.Id] = download
	} else {
		download = Download{}
		download.Status = message.Status
		download.FileName = message.Name
		download.Started = time.Now()
		download.Size = int(message.Expected)
		download.Progress = int(message.Total)
		download.Fragments = map[string]Download{}
		dr.InProgress[message.Id] = download
	}

	return err
}

func (dr *DefaultDownloadReporter) OnProgress(message *TrackerMessage) error {
	var (
		err      error
		download Download
	)

	if message.IsFragment {
		// Progress fragment
		download = dr.InProgress[message.Id].Fragments[message.Name]
		download.Delta += message.Delta
		download.Progress += message.Delta
		download.Status = message.Status
		dr.InProgress[message.Id].Fragments[message.Name] = download
		// Also update the main
		download = dr.InProgress[message.Id]
		download.Delta += message.Delta
		download.Progress += message.Delta
		download.Status = message.Status
		dr.InProgress[message.Id] = download
	} else {
		// Progress main
		download = dr.InProgress[message.Id]
		download.Delta += message.Delta
		download.Progress += message.Delta
		download.Status = message.Status
		dr.InProgress[message.Id] = download
	}

	return err
}

func (dr *DefaultDownloadReporter) OnError(message *TrackerMessage) error {
	// Keep an error db?
	fmt.Println(message.Err)
	return message.Err
}

func (dr *DefaultDownloadReporter) OnMerging(message *TrackerMessage) error {
	var (
		err      error
		download Download
	)

	download = dr.InProgress[message.Id]
	download.Status = message.Status
	dr.InProgress[message.Id] = download

	return err
}

func (dr *DefaultDownloadReporter) OnTick() error {
	var (
		err error
	)

	var amount int
	if len(dr.InProgress) == 0 {
		dr.WaitTime++
		if dr.WaitTime == 10 {
			dr.ShouldStop = true
		}
	} else {
		dr.WaitTime = 0
	}
	for id, download := range dr.InProgress {
		// Filename: 100mb.bin
		fmt.Println("Filename: " + download.FileName)
		status := DownloadStatusString(download.Status)
		amount = int(math.Round(float64(download.Progress) / float64(download.Size) * float64(100) / float64(2)))
		if download.Status == StatusProgress {
			// Status: Downloading, Speed: 1.1mb/s, ETA: 00:01:52 [50%]
			eta := time.Duration((download.Size-download.Progress)/(download.Delta+1)) * time.Second
			percentage := int(math.Round(float64(download.Progress) / float64(download.Size) * float64(100)))
			speed := humanize.Bytes(uint64(download.Delta))
			download.Delta = 0
			fmt.Println(fmt.Sprintf("Status: %s, Speed: %s/s, ETA: %s [%d%%]", status, speed, eta.String(), percentage))
		} else {
			fmt.Println("Status: " + status)
		}
		barStr := ""
		if download.Threaded {
			// [XXXXXXXXXXX][XXXXXXXXXXX][XXXXXXXXXXX][XXXXXXXXXXXX]
			// Also wipe fragment deltas
			for name, thread := range download.Fragments {
				amount = int(math.Round(float64(thread.Progress) / float64(thread.Size) * float64(100) / float64(10)))
				barStr += "["
				for i := 0; i < 10; i++ {
					if i < amount {
						barStr += "X"
					} else {
						barStr += "="
					}
				}
				barStr += "]"
				thread.Delta = 0
				download.Fragments[name] = thread
			}
		} else {
			// [XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX]
			barStr += "["
			for i := 0; i < 50; i++ {
				if i < amount {
					barStr += "X"
				} else {
					barStr += "="
				}
			}
			barStr += "]"
		}
		fmt.Println(barStr)
		// Commit delta wipes
		dr.InProgress[id] = download
	}

	return err
}

func (dr *DefaultDownloadReporter) Processor() error {
	var (
		err error
	)

	for {
		select {
		case msg := <-dr.Chan:
			switch msg.Status {
			case StatusDone:
				err = dr.OnDone(msg)
			case StatusStart:
				// Build the DB!
				err = dr.OnStart(msg)
			case StatusProgress:
				err = dr.OnProgress(msg)
			case StatusError:
				// ERROR
				err = dr.OnError(msg)
			case StatusMerging:
				// Merging started
				err = dr.OnMerging(msg)
			}
			break
		case <-dr.Ticker.C:
			err = dr.OnTick()
			break
		}
		if dr.ShouldStop {
			break
		}
	}
	dr.WaitGroup.Done()

	return err
}

func UseDownloadReporter(reporter DownloadReporter) {
	go func() {
		fmt.Println(reporter.Processor())
	}()
}
