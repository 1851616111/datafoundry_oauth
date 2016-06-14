package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/rand"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
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

func httpPATCH(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("PATCH", url, body, credential...)
}

func httpGet(url string, credential ...string) ([]byte, error) {

	var resp *http.Response
	var err error
	if len(credential) == 2 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s", url, err)
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
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func httpGetFunc(url string, f func(resp *http.Response), credential ...string) ([]byte, error) {

	var resp *http.Response
	var err error
	if len(credential) == 2 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s", url, err)
		}
		req.Header.Set(credential[0], credential[1])

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}

		if f != nil {
			f(resp)
		}

		switch resp.StatusCode {
		case 404:
			return nil, ErrNotFound
		case 200:
			return ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func httpAction(method, url string, body []byte, credential ...string) ([]byte, error) {
	fmt.Println(method, url, string(body), credential)
	var resp *http.Response
	var err error
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if len(credential) == 2 {
		req.Header.Set(credential[0], credential[1])
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("[http] read err %s, %s", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		log.Printf("request err %s", string(b))
		return nil, fmt.Errorf("[http] status err %s, %d", url, resp.StatusCode)
	}

	return b, nil
}

func httpDelete(url string, credential ...string) ([]byte, error) {
	return httpAction("DELETE", url, nil, credential...)
}

func generateGithubName(username string) string {
	return fmt.Sprintf("%s-github-%s", username, rand.String(8))
}

func generateReposDeployName(sourceKind, sourceHost string) string {
	return fmt.Sprintf("source-%s-%s-%s", sourceKind, convertDFValidateName(sourceHost), rand.String(8))
}

func retHttpCodef(code, bodyCode int, w http.ResponseWriter, format string, a ...interface{}) {

	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":"%s"}`, bodyCode, fmt.Sprintf(format, a...))

	fmt.Fprintf(w, msg)
	return
}

func retHttpCode(code int, bodyCode int, w http.ResponseWriter, a ...interface{}) {
	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":"%s"}`, bodyCode, fmt.Sprint(a...))

	fmt.Fprintf(w, msg)
	return
}

func retHttpCodeJson(code int, bodyCode int, w http.ResponseWriter, a ...interface{}) {
	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":%s}`, bodyCode, fmt.Sprint(a...))

	fmt.Fprintf(w, msg)
	return
}

type deployKey struct {
	Private string `json:"private_key"`
	Public  string `json:"public_key"`
}

func stripBearToken(authValue string) string {
	return strings.TrimSpace(strings.TrimLeft(authValue, "Bearer"))
}

func etcdFormatUrl(url string) string {
	return strings.Replace(url, "://", "_", 1)
}

func convertDFValidateName(name string) string {
	return strings.Replace(name, ".", "-", -1)
}

func contains(l []string, s string) bool {
	for _, str := range l {
		if str == s {
			return true
		}
	}
	return false
}

func getMd5(content []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(content)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func getRedisMasterAddr(sentinelAddr string) (string, string) {
	if len(sentinelAddr) == 0 {
		log.Printf("Redis sentinelAddr is nil.")
		return "", ""
	}

	conn, err := redis.DialTimeout("tcp", sentinelAddr, time.Second*10, time.Second*10, time.Second*10)
	if err != nil {
		log.Printf("redis dial timeout(\"tcp\", \"%s\", %d) error(%v)", sentinelAddr, time.Second, err)
		return "", ""
	}
	defer conn.Close()

	redisMasterPair, err := redis.Strings(conn.Do("SENTINEL", "get-master-addr-by-name", Redis_Cluster_Name))
	if err != nil {
		log.Printf("conn.Do(\"SENTINEL\", \"get-master-addr-by-name\", \"%s\") error(%v)", "mymaster", err)
		return "", ""
	}

	if len(redisMasterPair) != 2 {
		return "", ""
	}
	return redisMasterPair[0], redisMasterPair[1]
}
