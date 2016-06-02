package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	gitlabapi "github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	dfapi "github.com/openshift/origin/pkg/user/api/v1"
)

var WebHookKind = []string{"github", "gitlab"}

//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/oauth/webhooks  -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl 127.0.0.1:9443/v1/repos/source/gitlab/webhooks -d '{"host": "https://code.dataos.io", "namespace": "oauth", "build": "oauth", "repo":"43", "spec":{"url":"https://www.baidu.com"}}'  -H "Authorization:Bearer cIOXAervAeS0ErI6Ilm5vp1cYOMrAZ1ic7EA6e09GuE"
//curl 127.0.0.1:9443/v1/repos/source/github/webhooks -d '{"host": "https://github.com", "namespace": "oauth", "build": "oauth", "user":"asiainfoLDP", "repo":"datafoundry_oauth2", "spec":{"events": ["push","pull_request","status"],"config": {"url": "http://example.com/webhook"}}}'  -H "Authorization:Bearer cIOXAervAeS0ErI6Ilm5vp1cYOMrAZ1ic7EA6e09GuE"
func createWebHookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	token := r.Header.Get("Authorization")
	var user *dfapi.User
	var err error
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	source := ps.ByName("source")
	if !contains(WebHookKind, source) {
		retHttpCode(400, 1400, w, fmt.Sprintf("unknown source %s.", source))
		return
	}

	switch source {
	case "github":
		userInfo, err := getGithubInfo(user)
		if err != nil {
			retHttpCode(400, 1401, w, "unauthorized")
			return
		}

		newHook := new(GitHubWebHookOption)
		newHook.DefaultOption()

		hook := installHookParser(newHook)
		if err := json.NewDecoder(r.Body).Decode(hook); err != nil {
			retHttpCodef(400, 1400, w, "read req body err %v", err.Error())
			return
		}
		hook.Host = "www.github.com"

		hookStr, err := getWebHook(source, hook.Host, hook.NameSpace, hook.Build)
		if err == nil {
			//若存在,对比hook是否变化
			oldHook := new(GitHubWebHook)
			json.Unmarshal([]byte(hookStr), oldHook)

			if gitHubWebHookchanged(&oldHook.GitHubWebHookOption, newHook) {
				credKey, credValue := getCredentials(userInfo)
				retWH, err := UpdateRepoWebHook(hook.User, hook.Repo, oldHook.Id, newHook, credKey, credValue)
				if err != nil {
					retHttpCodef(400, 1400, w, "update webhook err %v", err.Error())
					return
				}

				err = setWebHook(source, hook.Host, hook.NameSpace, hook.Build, retWH)
				if err != nil {
					retHttpCodef(400, 1400, w, "store webhook err %v", err.Error())
					return
				}
			}

			retHttpCode(200, 1200, w, "ok")
			return

		} else {
			//total new
			if EtcdKeyNotFound(err) {
				credKey, credValue := getCredentials(userInfo)

				retHook, err := CreateRepoWebHook(hook.User, hook.Repo, hook.Spec.(*GitHubWebHookOption), credKey, credValue)
				if err != nil {
					retHttpCodef(400, 1400, w, "create webhook err %v", err.Error())
					return
				}

				err = setWebHook(source, hook.Host, hook.NameSpace, hook.Build, retHook)
				if err != nil {
					retHttpCodef(400, 1400, w, "store webhook err %v", err.Error())
					return
				}

				retHttpCode(200, 1200, w, "ok")
				return
			}

			retHttpCodef(400, 1400, w, "get webhook info err %v", err.Error())
			return
		}

	case "gitlab":
		option, err := getGitLabOptionByDFUser(user.Name)
		if err != nil {
			if EtcdKeyNotFound(err) {
				retHttpCode(400, 1401, w, "unauthorized")
				return
			}
		}

		//default config
		newHook := new(gitlabapi.WebHookParam)
		newHook.Push_events = true
		newHook.Enable_ssl_verification = true
		hook := installHookParser(newHook)

		if err := json.NewDecoder(r.Body).Decode(hook); err != nil {
			retHttpCodef(400, 1400, w, "read req body err %v", err.Error())
			return
		}
		newHook.Id = 0

		if err := hook.validate(); err != nil {
			retHttpCode(400, 1400, w, err.Error())
			return
		}

		//查询一个namespace的某个build的webhook的记录
		hookStr, err := getWebHook(source, hook.Host, hook.NameSpace, hook.Build)
		if err == nil {
			//若存在,对比hook是否变化
			oldHook := new(gitlabapi.WebHookParam)
			json.Unmarshal([]byte(hookStr), oldHook)

			if webHookchanged(oldHook, newHook) {
				o := &gitlabapi.NewOption{
					Param:            hook.Spec,
					ParamContentType: gitlabapi.ContentType_Form,
				}

				retWH, err := glApi.WebHook(option.Host, option.PrivateToken).UpdateWebHook(hook.Repo, oldHook.Id, o)
				if err != nil {
					retHttpCodef(400, 1400, w, "update webhook err %v", err.Error())
					return
				}

				err = setWebHook(source, hook.Host, hook.NameSpace, hook.Build, retWH)
				if err != nil {
					retHttpCodef(400, 1400, w, "store webhook err %v", err.Error())
					return
				}
			}

			retHttpCode(200, 1200, w, "ok")
			return

		} else {
			//total new
			if EtcdKeyNotFound(err) {

				o := &gitlabapi.NewOption{
					Param:            hook.Spec,
					ParamContentType: gitlabapi.ContentType_Form,
				}

				retHook, err := glApi.WebHook(option.Host, option.PrivateToken).CreateWebHook(hook.Repo, o)
				if err != nil {
					retHttpCodef(400, 1400, w, "create webhook err %v", err.Error())
					return
				}

				err = setWebHook(source, hook.Host, hook.NameSpace, hook.Build, retHook)
				if err != nil {
					retHttpCodef(400, 1400, w, "store webhook err %v", err.Error())
					return
				}

				retHttpCode(200, 1200, w, "ok")
				return
			}

			retHttpCodef(400, 1400, w, "get webhook info err %v", err.Error())
			return
		}
	}
}

//curl -XDELETE 127.0.0.1:9443/v1/repos/source/gitlab/webhooks?host=https://code.dataos.io\&namespace=oauth\&build=oauth\&repo=43 -H "Authorization:Bearer uH7VUpN5c9CcL4KWFCuAAGqk4INvRb4vgpsZo0FOUFA"
//curl -XDELETE 127.0.0.1:9443/v1/repos/source/github/webhooks?namespace=oauth\&build=oauth\&user=asiainfoLDP\&repo=datafoundry_oauth2 -H "Authorization:Bearer uH7VUpN5c9CcL4KWFCuAAGqk4INvRb4vgpsZo0FOUFA"
func deleteWebHookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	source := ps.ByName("source")
	if !contains(WebHookKind, source) {
		retHttpCode(400, 1400, w, fmt.Sprintf("unknown source %s.", source))
		return
	}

	host, namespace, build, repo, gituser := r.FormValue("host"), r.FormValue("namespace"), r.FormValue("build"), r.FormValue("repo"), r.FormValue("user")

	token := r.Header.Get("Authorization")

	user, err := authDF(token)
	if err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	switch source {
	case "github":
		host = "www.github.com"
		userInfo, err := getGithubInfo(user)
		if err != nil {
			retHttpCode(400, 1401, w, "unauthorized")
			return
		}

		hookStr, err := getWebHook(source, host, namespace, build)
		if err != nil {
			if EtcdKeyNotFound(err) {
				retHttpCode(200, 200, w, "ok")
				return
			}
			retHttpCodef(404, 1404, w, "delete webhook unknown err %v", err)
			return
		}

		hook := new(GitHubWebHook)
		if err = json.Unmarshal([]byte(hookStr), hook); err != nil {
			retHttpCodef(400, 1400, w, "delete webhook err %v", err.Error())
			return
		}

		credKey, credValue := getCredentials(userInfo)
		if err := DeleteRepoWebHook(gituser, repo, hook.Id, credKey, credValue); err != nil {
			retHttpCodef(400, 1400, w, "delete webhook err %v", err.Error())
			return
		}

		if err := deleteWebHook(source, host, namespace, build); err != nil {
			retHttpCodef(400, 1400, w, "delete webhook storage err %v", err.Error())
			return
		}


	case "gitlab":

		option, err := getGitLabOptionByDFUser(user.Name)
		if err != nil {
			if EtcdKeyNotFound(err) {
				retHttpCode(200, 200, w, "ok")
				return
			}
		}

		hookStr, err := getWebHook(source, host, namespace, build)
		if err != nil {
			if EtcdKeyNotFound(err) {
				retHttpCode(404, 1404, w, "not found")
				return
			}

			retHttpCodef(404, 1404, w, "delete webhook err %v", err)
			return
		}

		//find hook
		hook := new(gitlabapi.WebHookParam)
		if err = json.Unmarshal([]byte(hookStr), hook); err != nil {
			retHttpCodef(400, 1400, w, "delete webhook err %v", err.Error())
			return
		}

		if err := glApi.WebHook(option.Host, option.PrivateToken).DeleteWebHook(repo, hook.Id); err != nil && err != gitlabapi.ErrNotFound {
			retHttpCodef(400, 1400, w, "delete webhook err %v", err.Error())
			return
		}

		if err := deleteWebHook(source, host, namespace, build); err != nil {
			retHttpCodef(400, 1400, w, "delete webhook storage err %v", err.Error())
			return
		}
	}

	retHttpCode(200, 200, w, "ok")
	return
}

type WebHook struct {
	Host      string
	NameSpace string
	Build     string
	User      string //github only
	Repo      string
	Spec      interface{} `json:"spec"`
}

func (h *WebHook) validate() error {
	invalidParam := []byte{}
	if strings.TrimSpace(h.Host) == "" {

		invalidParam = append(invalidParam, "host"...)
	}

	if strings.TrimSpace(h.NameSpace) == "" {
		if len(invalidParam) > 0 {
			invalidParam = append(invalidParam, " and "...)
		}
		invalidParam = append(invalidParam, "namespace"...)
	}

	if strings.TrimSpace(h.Build) == "" {
		if len(invalidParam) > 0 {
			invalidParam = append(invalidParam, " and "...)
		}
		invalidParam = append(invalidParam, "build"...)
	}

	if strings.TrimSpace(h.Repo) == "" {
		if len(invalidParam) > 0 {
			invalidParam = append(invalidParam, " and "...)
		}
		invalidParam = append(invalidParam, "repo"...)
	}

	if len(invalidParam) > 0 {
		return errors.New(fmt.Sprintf("param %v can not be nil.", invalidParam))
	}

	return nil
}

func installHookParser(webHook interface{}) *WebHook {
	return &WebHook{
		Spec: webHook,
	}
}

func webHookchanged(old, new *gitlabapi.WebHookParam) bool {
	if old.Url != new.Url {
		return true
	}

	if old.Push_events != new.Push_events {
		return true
	}

	if old.Issues_events != new.Issues_events {
		return true
	}

	if old.Merge_requests_events != new.Merge_requests_events {
		return true
	}

	if old.Tag_push_events != new.Tag_push_events {
		return true
	}

	if old.Note_events != new.Note_events {
		return true
	}

	if old.Enable_ssl_verification != new.Enable_ssl_verification {
		return true
	}

	return false
}
