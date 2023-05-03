package gokhttp_download

import (
	"net/http"
	"strconv"
	"strings"
)

func supportsRange(resp *http.Response) (uint64, bool) {
	supportsRanges := false
	contentLength := uint64(0)
	if resp != nil {
		if resp.Request.Method == http.MethodHead && resp.StatusCode == http.StatusOK {
			if resp.Header.Get("Accept-Ranges") == "bytes" {
				supportsRanges = true
			}
			if resp.Header.Get("Ranges-Supported") == "bytes" {
				supportsRanges = true
			}
			contentLength = uint64(resp.ContentLength)
		} else if resp.Request.Method == http.MethodGet && resp.StatusCode == http.StatusPartialContent {
			contentRange := resp.Header.Get("Content-Range")
			if contentRange != "" {
				supportsRanges = true
			}
			cLSplit := strings.Split(contentRange, "/")
			if len(cLSplit) == 2 {
				contentLength, _ = strconv.ParseUint(cLSplit[1], 10, 64)
			}
		} else {
			// File inaccessible
		}
	}
	return contentLength, supportsRanges
}
