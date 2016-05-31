package main

import (
	"github.com/julienschmidt/httprouter"
	//dfapi "github.com/openshift/origin/pkg/user/api/v1"
	"net/http"

	//"encoding/json"
)

type WebHook struct {
	Host string
	Repo string
	Spec interface{}
}

func newWebHook(webHook interface{}) *WebHook {
	return &WebHook{
		Spec: webHook,
	}
}

//const WebHookKind = []string{"github", "gitlab"}
//
////curl 127.0.0.1:9443/v1/repos/gitlab/webhook -d '{"host": "https://code.dataos.io", "repo":"43", "spec":{"url":"https://dev.dataos.io:8443/oapi/v1/namespaces/oauth/buildconfigs/oauth/webhooks/3oMcTgjWowVBayrnqIK2/generic"}}'  -H "Authorization:bearer i1TerZwHQSsveIrHs53wr6lKdzxbJL2mVNCu8fs5Ao0"
func webHookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//token := r.Header.Get("Authorization")
	//var user *dfapi.User
	//var err error
	//if user, err = authDF(token); err != nil {
	//	retHttpCodef(401, 1401, w, "auth err %s", err.Error())
	//	return
	//}
	//
	//source := ps["source"]
	//
	//switch source {
	//case "gitlab":
	//	option, err := getGitLabOptionByDFUser(user.Name)
	//	if err != nil {
	//		if EtcdKeyNotFound(err) {
	//			retHttpCode(400, 1401, w, "unauthorized")
	//		} else {
	//			retHttpCodef(400, 1400, w, "get user info err %s", err.Error())
	//		}
	//
	//		return
	//	}
	//
	//	wh := newWebHook(new(gitlabapi.WebHookParam))
	//
	//	if err := json.NewDecoder(r.Body).Decode(wh); err != nil {
	//		retHttpCodef(400, 1400, w, "read req body err %v", err.Error())
	//		return
	//	}
	//
	//	_, err := glApi.WebHook(option.Host, option.PrivateToken).CreateWebHook(wh.Spec.(*gitlabapi.WebHookParam))
	//	if err != nil {
	//		retHttpCodef(400, 1400, w, "get projects err %v", err.Error())
	//		return
	//	}
	//
	//default:
	//
	//}

}
