package gokhttp

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/anaskhan96/soup"
	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func (res *HttpResponse) Text() (string, error) {

	var body []byte
	var err error

	if strings.Contains(res.Header.Get("Content-Type"), "plain/text") || strings.Contains(res.Header.Get("Content-Type"), "text/plain") {
		body, err = res.Bytes()
	} else {
		err = errors.New("content-type is: " + res.Header.Get("Content-Type"))
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}

	return string(body), err
}

func (res *HttpResponse) Bytes() ([]byte, error) {

	var body []byte
	var buffer *bytes.Buffer
	var err error

	if res.ContentLength == -1 {
		buffer = bytes.NewBuffer(body)
		_, err = buffer.ReadFrom(res.Body)
		body = buffer.Bytes()
	} else {
		body = make([]byte, res.ContentLength)
		buffer = bytes.NewBuffer(body)
		_, err = io.Copy(buffer, res.Body)
		body = buffer.Bytes()
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return body, err
}

func (res *HttpResponse) Discard() error {

	var err error

	_, err = io.Copy(ioutil.Discard, res.Body)

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}

	return err
}

func (res *HttpResponse) SaveToFile(filename string, permissions os.FileMode, overwrite, append bool) error {

	var err error
	var info os.FileInfo
	var f *os.File

	info, err = os.Stat(filename)
	if os.IsNotExist(err) {
		f, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, permissions)
		if err == nil {
			_, err = io.Copy(f, res.Body)
		}
	} else {
		if !info.IsDir() {
			switch true {
			case overwrite:
				f, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, permissions)
			case append:
				f, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, permissions)
			default:
				err = errors.New("no action performed")
			}
			if err == nil {
				_, err = io.Copy(f, res.Body)
			}
		} else {
			err = errors.New("can't treat folder as a file")
		}
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return err
}

func (res *HttpResponse) JSON() (*HttpJSONResponse, error) {

	var body []byte
	var err error

	if strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		body, err = res.Bytes()
	} else {
		err = errors.New("content-type is: " + res.Header.Get("Content-Type"))
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return &HttpJSONResponse{body}, err
}

func (res *HttpResponse) XML() (*etree.Document, error) {

	var err error
	var doc *etree.Document

	if strings.Contains(res.Header.Get("Content-Type"), "application/xml") {
		doc = etree.NewDocument()
		_, err = doc.ReadFrom(res.Body)
	} else {
		err = errors.New("content-type is: " + res.Header.Get("Content-Type"))
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return doc, err
}

func (res *HttpResponse) HTML() (*soup.Root, error) {

	var err error
	var htmlSoup soup.Root
	var body []byte

	if strings.Contains(res.Header.Get("Content-Type"), "plain/html") || strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		body, err = res.Bytes()
		if checkError(err) {
			err = res.Body.Close()
			if checkError(err) {
				htmlSoup = soup.HTMLParse(string(body))
			}
		}
	} else {
		err = errors.New("content-type is: " + res.Header.Get("Content-Type"))
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return &htmlSoup, err
}

func (res *HttpResponse) Object(o interface{}) error {

	var err error

	switch true {
	case strings.Contains(res.Header.Get("Content-Type"), "application/xml"):
		err = xml.NewDecoder(res.Body).Decode(o)
	case strings.Contains(res.Header.Get("Content-Type"), "application/json"):
		err = json.NewDecoder(res.Body).Decode(o)
	default:
		err = errors.New("content-type not supported: " + res.Header.Get("Content-Type"))
	}

	if err == nil {
		err = res.Body.Close()
	} else {
		_ = res.Body.Close()
	}
	return err
}

func (res *HttpJSONResponse) ArrayEach(cb func(value []byte, dataType jsonparser.ValueType, offset int, err error), keys ...string) (offset int, err error) {
	return jsonparser.ArrayEach(res.data, cb, keys...)
}

func (res *HttpJSONResponse) Delete(keys ...string) []byte {
	return jsonparser.Delete(res.data, keys...)
}

func (res *HttpJSONResponse) EachKey(cb func(int, []byte, jsonparser.ValueType, error), paths ...[]string) int {
	return jsonparser.EachKey(res.data, cb, paths...)
}

func (res *HttpJSONResponse) GetBoolean(keys ...string) (val bool, err error) {
	return jsonparser.GetBoolean(res.data, keys...)
}

func (res *HttpJSONResponse) GetFloat(keys ...string) (val float64, err error) {
	return jsonparser.GetFloat(res.data, keys...)
}

func (res *HttpJSONResponse) GetInt(keys ...string) (val int64, err error) {
	return jsonparser.GetInt(res.data, keys...)
}

func (res *HttpJSONResponse) GetString(keys ...string) (val string, err error) {
	return jsonparser.GetString(res.data, keys...)
}

func (res *HttpJSONResponse) GetUnsafeString(keys ...string) (val string, err error) {
	return jsonparser.GetUnsafeString(res.data, keys...)
}

func (res *HttpJSONResponse) ObjectEach(callback func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error, keys ...string) (err error) {
	return jsonparser.ObjectEach(res.data, callback, keys...)
}

func (res *HttpJSONResponse) Set(setValue []byte, keys ...string) (value []byte, err error) {
	return jsonparser.Set(res.data, setValue, keys...)
}

func (res *HttpJSONResponse) ParseString() (string, error) {
	return jsonparser.ParseString(res.data)
}
