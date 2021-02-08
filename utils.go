package gokhttp

import (
	"errors"
	"os"
	"strings"
)

func fileExists(fileName string) (os.FileInfo, bool) {
	var (
		err  error
		info os.FileInfo
	)
	info, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return info, false
	}
	if err == nil {
		return info, !info.IsDir()
	}
	return nil, false
}

func fileCreate(fileName string) (*os.File, error) {
	var (
		err  error
		file *os.File
	)
	if _, exists := fileExists(fileName); !exists {
		file, err = os.Create(fileName)
	} else {
		err = errors.New("file exists")
	}
	return file, err
}

func fileAppend(fileName string, mode os.FileMode) (*os.File, error) {
	var (
		err  error
		file *os.File
	)
	if _, exists := fileExists(fileName); exists {
		file, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, mode)
	} else {
		err = errors.New("file does not exist")
	}
	return file, err
}

func fileDelete(fileName string) error {
	var err error
	if _, exists := fileExists(fileName); exists {
		err = os.Remove(fileName)
	} else {
		err = errors.New("file does not exist")
	}
	return err
}

func DownloadStatusString(status int) string {
	result := "Unknown"
	switch status {
	case StatusStart:
		result = "Starting"
	case StatusError:
		result = "Error"
	case StatusProgress:
		result = "Downloading"
	case StatusDone:
		result = "Done"
	case StatusMerging:
		result = "Merging fragments"
	}
	return result
}

func ParseIndex(content []byte) {
	// convert to string all at once or keep as bytes and convert if needed
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") {
			continue
		}
		elem := map[string]string{}
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}
		elem["type"] = line[1:colonIndex]
		attributes := strings.Split(line[colonIndex+1:], ",")
		for _, attr := range attributes {
			if !strings.Contains(attr, "=") {
				continue
			}
			entry := strings.Split(attr, "=")
			elem[entry[0]] = strings.ReplaceAll(entry[1], "\"", "")
		}
		if strings.ToUpper(elem["type"]) == "EXT-X-STREAM-INF" {
			elem["URI"] = lines[i+1]
		}
	}
}

func ParsePlaylist(content []byte) {
	// convert to string all at once or keep as bytes and convert if needed
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") {
			continue
		}
		elem := map[string]string{}
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}
		if strings.Contains(line, ",") {

		} else {

		}
		elem[line[1:colonIndex]] = line[colonIndex+1:]
		attributes := strings.Split(line[colonIndex+1:], ",")
		for _, attr := range attributes {
			if !strings.Contains(attr, "=") {
				continue
			}
			entry := strings.Split(attr, "=")
			elem[entry[0]] = strings.ReplaceAll(entry[1], "\"", "")
		}
		if strings.ToUpper(elem["type"]) == "EXT-X-STREAM-INF" {
			elem["URI"] = lines[i+1]
		}
	}
}
