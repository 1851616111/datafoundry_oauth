package main

import (
	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	gitlabutil "github.com/asiainfoLDP/datafactory_oauth2/util"
	oapi "github.com/openshift/origin/pkg/user/api/v1"
	"io/ioutil"
	"log"
	"net/http"
)

type gitLabInfo struct {
	Host         string `json:"host"`
	User         string `json:"user"`
	PrivateToken string `json:"private_token"`
}

type gitLabBindInfo struct {
	Host string `json:"host"`
	Id   int    `json:"project_id"`
}

type gitlabDumpling struct {
	filling *gitlabapi.User
	err     error
}

type datafoundryDumpling struct {
	filling *oapi.User
	err     error
}

func parseRequestBody(r *http.Request, i interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	return nil
}

func hasDeployKey(DataFoundryHost, user, gitLabHost string) (bool, *gitlabutil.KeyPair) {
	deployKey := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlab_service/%s/deploykey", DataFoundryHost, user, etcdFormatUrl(gitLabHost))

	pair := new(gitlabutil.KeyPair)
	if err := getJson(deployKey, pair); err != nil {
		if EtcdKeyNotFound(err) {
			//不存在
			return false, nil
		}
		//异常情况,不知道是否存在
		log.Printf("get deploy key unknown err %v", err)
		return true, nil
	}
	//存在并取回结果
	return true, pair
}

func setDeployKey(DataFoundryHost, user, gitLabHost string, new *gitlabutil.KeyPair) error {
	deployKey := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlab_service/%s/deploykey", DataFoundryHost, user, etcdFormatUrl(gitLabHost))
	return db.set(deployKey, new)
}
