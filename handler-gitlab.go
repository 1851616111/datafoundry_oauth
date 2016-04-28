package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"

	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	api "github.com/openshift/origin/pkg/user/api/v1"
)

var gitlab = gitlabapi.ClientFactory()

//curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer aFmDtGfJCefPDrdha7UvBFkTbT_yp-1jaCY3C0tgM4c"
func gitlabHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authorization := r.Header.Get("Authorization")

	option := new(gitlabBindOption)
	if err := parseRequestBody(r, option); err != nil {
		retHttpCodef(400, w, "read request body err %v", err)
		return
	}

	// todo validate option
	ret_gb, ret_df := make(chan *gitlabDumpling, 1), make(chan *datafoundryDumpling, 1)

	go func(gitlab *gitlabapi.HttpFactory, option *gitlabBindOption, ret *chan *gitlabDumpling) {
		f, err := gitlab.User(option.Host, option.PrivateToken).GetUser()

		*ret <- &gitlabDumpling{
			filling: f,
			err:     err,
		}
	}(gitlab, option, &ret_gb)

	go func(authorization string, ret *chan *datafoundryDumpling) {
		f, err := authDF(authorization)

		*ret <- &datafoundryDumpling{
			filling: f,
			err:     err,
		}
	}(authorization, &ret_df)

	count := 0
	var oUser *api.User
res:
	for {
		select {
		case dump := <-ret_df:
			count++
			if dump.err != nil {
				retHttpCodef(401, w, "unauthorized from datafoundry,  err %v\n", dump.err)
				return
			}
			oUser = dump.filling
		case dump := <-ret_gb:
			count++
			if dump.err != nil {
				retHttpCodef(401, w, "unauthorized from gitlab %s,  err %v\n", option.Host, dump.err)
				return
			}
		default:
			if count == 2 {
				break res
			}
		}
	}

	key := fmt.Sprintf("/oauth/gitlab/dfuser/%s", oUser.Name)
	if err := db.set(key, option); err != nil {
		retHttpCodef(400, w, "store gitlab err %v", err.Error())
		return
	}

	retHttpCodef(200, w, "ok")
}

//curl http://127.0.0.1:9443/v1/gitlab/repos/owner -H "Authorization:bearer aFmDtGfJCefPDrdha7UvBFkTbT_yp-1jaCY3C0tgM4c"
//curl http://127.0.0.1:9443/v1/gitlab/repos/org -H "Authorization:bearer aFmDtGfJCefPDrdha7UvBFkTbT_yp-1jaCY3C0tgM4c"
func gitLabOwnerReposHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userType := ps.ByName("user")

	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	option, err := getGitLabOptionByDFUser(user.Name)
	if err != nil {
		retHttpCodef(400, w, "get gitlab info err %v", err.Error())
		return
	}

	projects, err := gitlab.Project(option.Host, option.PrivateToken).ListProjects()
	if err != nil {
		retHttpCodef(400, w, "get projects err %v", err.Error())
		return
	}

	var l interface{}
	switch userType {
	case "org":
		l = gitlabapi.ConverOrgProjects(projects)
	case "owner":
		l = gitlabapi.ConverOwnerProjects(projects)

	}

	b, err := json.Marshal(l)
	if err != nil {
		retHttpCodef(400, w, "convert projects err %v", err)
		return
	}

	retHttpCodef(200, w, "%s", string(b))
}

func getGitLabOptionByDFUser(name string) (*gitlabBindOption, error) {
	key := fmt.Sprintf("/oauth/gitlab/dfuser/%s", name)
	s, err := db.get(key, true, false)
	if err != nil {
		return nil, err
	}

	option := new(gitlabBindOption)
	if err := json.Unmarshal([]byte(s), option); err != nil {
		return nil, err
	}

	return option, nil
}
