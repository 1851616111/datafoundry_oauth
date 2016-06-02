package gitlab

import (
	httputil "github.com/asiainfoLDP/datafoundry_oauth2/util/http"
)

type User struct {
	Name          string `json:"name"`
	Username      string `json:"username"`
	Id            int    `json:"id"`
	AvatarUrl     string `json:"avatar_url"`
	WebUrl        string `json:"web_url"`
	Email         string `json:"email"`
	ProjectsLimit int    `json:"projects_limit"`
}

type Group struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Description     string `json:"description"`
	VisibilityLevel int    `json:"visibility_level"`
	AvatarUrl       string `json:"avatar_ur"`
	WebUrl          string `json:"web_url"`
}

type Owner struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Id        int    `json:"id"`
	AvatarUrl string `json:"avatar_url"`
	WebUrl    string `json:"web_url"`
}

type Namespace struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	OwnerId     int    `json:"owner_id"`
	Description string `json:"description"`
}

type Project struct {
	Id                int        `json:"id"`
	Public            bool       `json:"public"`
	SshUrlToRepo      string     `json:"ssh_url_to_repo"`
	Owner             *Owner     `json:"owner,omitempty"`
	Name              string     `json:"name"`
	NameWithNamespace string     `json:"name_with_namespace"`
	Namespace         *Namespace `json:"namespace"`
}

type commit struct {
	Id string `json:"id"`
	//Message string `json:"message"`
}

type Branch struct {
	Name   string `json:"name"`
	Commit commit `json:"commit"`
}

type DeployKey struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Key   string `json:"key"`
}

type NewDeployKeyOption struct {
	ProjectId int
	Param     NewDeployKeyParam
}

type NewDeployKeyParam struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

type ClientOption struct {
	Host            string
	Api             string
	CredentialKey   string
	CredentialValue string
}

type RestClient struct {
	Url        string
	Credential Credential
	Client     *HttpFactory
}

type Credential struct {
	Key   string
	Value string
}

type Session struct {
	Name         string `json:"name"`
	UserName     string `json:"username"`
	PrivateToken string `json:"private_token"`
}

type WebHookParam struct {
	Id                      int    `param:"-"`
	Url                     string `param:"url"`
	Push_events             bool   `param:"push_events"`
	Issues_events           bool   `param:"issues_events"`
	Merge_requests_events   bool   `param:"merge_requests_events"`
	Tag_push_events         bool   `param:"tag_push_events"`
	Note_events             bool   `param:"note_events"`
	Enable_ssl_verification bool   `param:"enable_ssl_verification"`
}

func (p *WebHookParam) String() string {
	return httputil.InterfaceToString(p)
}
