package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func httpAddrMaker(addr string) string {
	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if !strings.HasPrefix(addr, "http://") {
		return fmt.Sprintf("http://%s", addr)
	}

	return addr
}

func Schemastripper(addr string) string {
	schemas := []string{"https://", "http://"}

	for _, schema := range schemas {
		if strings.HasPrefix(addr, schema) {
			return strings.TrimLeft(addr, schema)
		}
	}

	return ""
}

func httpPost(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("POST", url, body, credential...)
}

func httpPUT(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("PUT", url, body, credential...)
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

func httpAction(method, url string, body []byte, credential ...string) ([]byte, error) {
	var resp *http.Response
	var err error
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(credential[0], credential[1])
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("[http] read err %s, %s\n", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("[http] status err %s, %d\n", url, resp.StatusCode)
	}

	return b, nil
}

func generateGithubName(username string) string {
	return fmt.Sprintf("%s-github", username)
}

func generateGitlabName(username, gitlabHost string) string {
	return fmt.Sprintf("%s-gitlab-%s", username, convertDFValidateName(gitlabHost))
}

func retHttpCodef(code int, w http.ResponseWriter, format string, a ...interface{}) {
	if code != 200 {
		w.WriteHeader(code)
	}

	fmt.Fprintf(w, format, a...)
	return
}

func retHttpCode(code int, w http.ResponseWriter, a ...interface{}) {
	if code != 200 {
		w.WriteHeader(code)
	}

	fmt.Fprint(w, a...)
	return
}

type deployKey struct {
	Private string `json:"private_key"`
	Public  string `json:"public_key"`
}

func stripBearToken(authValue string) string {
	return strings.TrimSpace(strings.TrimLeft(authValue, "bearer "))
}

func etcdFormatUrl(url string) string {
	return strings.Replace(url, "://", "_", 1)
}

func convertDFValidateName(name string) string {
	return strings.Replace(name,".", "-", -1)
}