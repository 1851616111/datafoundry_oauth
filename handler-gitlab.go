package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"

	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	oapi "github.com/openshift/origin/pkg/user/api/v1"
)

var gitlab = gitlabapi.ClientFactory()

//curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm", "namespace": "oauth", "bearer":"aFmDtGfJCefPDrdha7UvBFkTbT_yp-1jaCY3C0tgM4c"}'
func gitlabHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	option := new(gitlabBindOption)
	if err := parseRequestBody(r, option); err != nil {
		retHttpCodef(400, w, "read request body err %v", err)
		return
	}

	ret_gb, ret_df := make(chan *gitlabDumpling, 1), make(chan *datafoundryDumpling, 1)

	go func(gitlab *gitlabapi.HttpFactory, option *gitlabBindOption, ret *chan *gitlabDumpling) {
		f, err := gitlab.User(option.Host, option.PrivateToken).GetUser()

		*ret <- &gitlabDumpling{
			filling: f,
			err:     err,
		}
	}(gitlab, option, &ret_gb)

	go func(option *gitlabBindOption, ret *chan *datafoundryDumpling) {
		f, err := authDF("bearer " + option.Bearer)

		*ret <- &datafoundryDumpling{
			filling: f,
			err:     err,
		}
	}(option, &ret_df)

	count := 0
	var oUser *oapi.User
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
