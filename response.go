package gokhttp

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/anaskhan96/soup"
	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"strings"
)

func (res *HttpResponse) Text() (string, error) {
	if strings.Contains(res.Header.Get("Content-Type"), "plain/text") || strings.Contains(res.Header.Get("Content-Type"), "text/plain") {
		body, err := ioutil.ReadAll(res.Body)
		if checkError(err) {
			err = res.Body.Close()
			if checkError(err) {
				return string(body), err
			}
		}
		return "", err
	}
	return "", errors.New("Content-Type is: " + res.Header.Get("Content-Type"))
}

func (res *HttpResponse) Bytes() ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if checkError(err) {
		err = res.Body.Close()
		if checkError(err) {
			return body, err
		}
	}
	return nil, err
}

func (res *HttpResponse) JSON() (*HttpJSONResponse, error) {
	if strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		body, err := ioutil.ReadAll(res.Body)
		if checkError(err) {
			err = res.Body.Close()
			if checkError(err) {
				return &HttpJSONResponse{body}, err
			}
		}
		return nil, err
	}
	return nil, errors.New("Content-Type is: " + res.Header.Get("Content-Type"))
}

func (res *HttpResponse) XML() (*etree.Document, error) {
	if strings.Contains(res.Header.Get("Content-Type"), "application/xml") {
		body, err := ioutil.ReadAll(res.Body)
		if checkError(err) {
			err = res.Body.Close()
			if checkError(err) {
				doc := etree.NewDocument()
				err = doc.ReadFromBytes(body)
				if checkError(err) {
					return doc, err
				}
			}
		}
		return nil, err
	}
	return nil, errors.New("Content-Type is: " + res.Header.Get("Content-Type"))
}

func (res *HttpResponse) HTML() (*soup.Root, error) {
	if strings.Contains(res.Header.Get("Content-Type"), "plain/html") || strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		body, err := ioutil.ReadAll(res.Body)
		if checkError(err) {
			err = res.Body.Close()
			if checkError(err) {
				htmlSoup := soup.HTMLParse(string(body))
				return &htmlSoup, err
			}
		}
		return nil, err
	}
	return nil, errors.New("Content-Type is: " + res.Header.Get("Content-Type"))
}

func (res *HttpResponse) Object(o interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if checkError(err) {
		err = res.Body.Close()
		if checkError(err) {
			switch true {
			case strings.Contains(res.Header.Get("Content-Type"), "application/xml"):
				err = xml.Unmarshal(body, o)
				return err
			case strings.Contains(res.Header.Get("Content-Type"), "application/json"):
				err = json.Unmarshal(body, o)
				return err
			default:
				return errors.New("Content-Type not supported: " + res.Header.Get("Content-Type"))
			}
		}
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
