package gitlab

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/json"
	"errors"
	"io"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)

type HttpFactory struct {
	Get     func(url string, credential ...string) ([]byte, error)
	GetJson func(json interface{}, url string, credential ...string) error
	Post    func(url string, bodyType string, body interface{}, credential ...string) ([]byte, error)
	Put     func(url string, bodyType string, body interface{}, credential ...string) ([]byte, error)
	//PostIO  func(url string, bodyType string, body io.Reader, credential ...string) ([]byte, error)
	Delete func(url string, credential ...string) ([]byte, error)
	Decode func(data []byte, v interface{}) error
	Encode func(v interface{}) ([]byte, error)
}

func ClientFactory() *HttpFactory {
	return &HttpFactory{
		Get:     httpGet,
		GetJson: httpGetJson,
		Post:    httpPost,
		Put:     httpPut,
		Delete:  httpDelete,
		Decode:  json.Unmarshal,
		Encode:  json.Marshal,
	}
}

func UrlMaker(host, api string) string {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}

	if strings.HasSuffix(host, "/") {
		host = strings.TrimRight(host, "/")
	}

	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}

	return host + api
}

func httpPost(url string, bodyType string, body interface{}, credential ...string) ([]byte, error) {
	return httpAction("POST", url, bodyType, body, credential...)
}

func httpPut(url string, bodyType string, body interface{}, credential ...string) ([]byte, error) {
	return httpAction("PUT", url, bodyType, body, credential...)
}

func httpGetJson(s interface{}, url string, credential ...string) error {
	b, err := httpGet(url, credential...)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, s); err != nil {
		return err
	}

	return nil
}

func httpGet(url string, credential ...string) ([]byte, error) {
	var resp *http.Response
	var err error
	if len(credential) == 2 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
		}
		req.Header.Set(credential[0], credential[1])
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		switch resp.StatusCode {
		case 404:
			return nil, ErrNotFound
		case 401:
			return nil, ErrUnauthorized
		case 200:
			return ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func httpAction(method, url string, bodyType string, body interface{}, credential ...string) ([]byte, error) {

	var req *http.Request
	var err error
	switch t := body.(type) {
	case []byte:
		req, err = http.NewRequest(method, url, bytes.NewBuffer(t))
	case io.Reader:
		req, err = http.NewRequest(method, url, t)
	}

	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}

	var resp *http.Response
	req.Header.Set("Content-Type", bodyType)
	if len(credential) == 2 {
		req.Header.Set(credential[0], credential[1])
	}
	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[http] read err %s, %s\n", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("[http] status err %s, %d\n", url, resp.StatusCode)
	}

	return b, nil
}

func httpActionIO(method, url string, bodyType string, body io.Reader, credential ...string) ([]byte, error) {
	var resp *http.Response
	var err error
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if len(credential) == 2 {
		req.Header.Set(credential[0], credential[1])
	}
	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[http] read err %s, %s\n", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("[http] status err %s, %d\n", url, resp.StatusCode)
	}

	return b, nil
}

func httpDelete(url string, credential ...string) ([]byte, error) {
	var resp *http.Response
	var err error
	if len(credential) == 2 {
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
		}
		req.Header.Set(credential[0], credential[1])
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		switch resp.StatusCode {
		case 404:
			return nil, ErrNotFound
		case 200:
			return ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func IsUnauthorized(err error) bool {
	return err == ErrUnauthorized
}
