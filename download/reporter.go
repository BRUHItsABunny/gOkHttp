package download

import (
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"math"
	"strconv"
	"strings"
	"time"
)

func (global *GlobalDownloadController) Tick(humanReadable bool) string {
	result := &strings.Builder{}
	if humanReadable {
		global.fmtReadable(result)
	} else {
		global.fmtNonReadable(result)
	}

	global.DeltaBytes.Store(0)
	global.Tasks.Range(func(k string, v *DownloadTaskController) bool {
		v.TaskStats.DeltaBytes.Store(0)
		return true
	})
	global.LastTick.Store(time.Now())
	return result.String()
}

func (global *GlobalDownloadController) fmtNonReadable(result *strings.Builder) {
	/*
		Backend CLI

		{total: {}, tasks: {taskN: {}}}
	*/
	jsonBytes, _ := json.Marshal(global)
	result.Write(jsonBytes)
}

func (global *GlobalDownloadController) fmtReadable(result *strings.Builder) {
	/*
		Live CLI

		Current IP: $IP and TIME: $DATETIME ($TOTAL_THREADS threads)
		Download statistics: $TOTAL_DONE ($TOTAL_DONE_SIZE) out of $TOTAL ($TOTAL_SIZE)
		[######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)

		$FILENAME
		Download statistic: $TOTAL_DONE ($TOTAL_DONE_SIZE) out of $TOTAL ($TOTAL_SIZE)
		Total  : [######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)
		Chunk N: [######################################] ($TOTAL_DONE_PERCENT% at $TOTAL_SPEED/s, ETA: $TOTAL_TIME_LEFT)

	*/
	result.WriteString(fmt.Sprintf("Current IP: %s and last tick %s (%d threads)\n", global.CurrentIP.Load(), global.LastTick.Load().Format(time.RFC3339), global.TotalThreads.Load()))
	result.WriteString(fmt.Sprintf("Download statistics: %d (%s) out of %d (%s)\nOverall runtime: ", global.DownloadedFiles.Load(), humanize.Bytes(global.DownloadedBytes.Load()), global.TotalFiles.Load(), humanize.Bytes(global.TotalBytes.Load())))
	progress(result, global.DeltaBytes.Load(), global.DownloadedBytes.Load(), global.TotalBytes.Load(), 40)
	result.WriteString(fmt.Sprintf("Downloading %d files:\n", global.Tasks.Len()))

	global.Tasks.Range(func(k string, v *DownloadTaskController) bool {
		result.WriteString(fmt.Sprintf("%s\n", Truncate(v.FileLocation.Load(), 128, 0)))
		result.WriteString(fmt.Sprintf("Download statistic: %s out of %s\nTotal file: ", humanize.Bytes(v.TaskStats.DownloadedBytes.Load()), humanize.Bytes(v.TaskStats.FileSize.Load())))
		progress(result, v.TaskStats.DeltaBytes.Load(), v.TaskStats.DownloadedBytes.Load(), v.TaskStats.FileSize.Load(), 40)
		chunkCount := v.ChunkCount.Load()
		if chunkCount > 1 {
			for i := uint64(1); i <= chunkCount; i++ {
				chunkKey := strconv.FormatUint(i, 10)
				chunk, ok := v.Chunks.Get(chunkKey)
				if ok {
					result.WriteString(fmt.Sprintf("Chunk %s: ", chunkKey))
					progress(result, chunk.DeltaBytes.Load(), chunk.DownloadedBytes.Load(), chunk.FileSize.Load(), 40)
				}
			}
		}
		return true
	})
}

func progress(result *strings.Builder, delta, downloaded, totalSize uint64, barSize int) {
	result.WriteString("[")
	percentage := float64(downloaded) / float64(totalSize)
	pieces := int(math.Floor(float64(barSize) * percentage))
	for i := 0; i < pieces; i++ {
		result.WriteString("X")
	}
	for i := 0; i < barSize-pieces; i++ {
		result.WriteString(" ")
	}
	result.WriteString("]")
	etaSecs := math.Ceil(float64(totalSize-downloaded) / float64(delta))
	eta := time.Second * time.Duration(etaSecs)
	result.WriteString(fmt.Sprintf(" (%.2f%% at %s/s, ETA: %s)\n", percentage*100, humanize.Bytes(delta), eta.String()))
}
