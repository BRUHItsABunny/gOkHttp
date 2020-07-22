package gokhttp

import (
	"errors"
	"os"
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
	return info, !info.IsDir()
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
