package gitlab

func (f *HttpFactory) newClient(host, api, privateToken string) *RestClient {
	return &RestClient{
		Url: UrlMaker(host, api),
		Credential: Credential{
			Key:   Gitlab_Credential_Key,
			Value: privateToken,
		},
		Client: f,
	}
}

func filterOwnerProject(pro Project) bool {
	return pro.Owner != nil
}

func filterOrgProject(pro Project) bool {
	return pro.Owner == nil
}

type ProjectList []Project

type NewOwnerProjectList struct {
	Owner       `json:"owner"`
	ProjectList `json:"repos"`
}

type NewOrgProjectList struct {
	Namespace   `json:"org"`
	ProjectList `json:"repos"`
}

func ConverOwnerProjects(pl []Project) []NewOwnerProjectList {
	m := map[Owner]ProjectList{}
	for i, v := range pl {
		if filterOwnerProject(pl[i]) {
			_, exists := m[*pl[i].Owner]
			if !exists {
				m[*pl[i].Owner] = ProjectList{}
			}
			v.Owner = nil
			v.Namespace = nil
			m[*pl[i].Owner] = append(m[*pl[i].Owner], v)
		}
	}

	npl := []NewOwnerProjectList{}
	for k, v := range m {
		npl = append(npl, NewOwnerProjectList{k, v})
	}

	return npl
}

func ConverOrgProjects(pl []Project) []NewOrgProjectList {
	m := map[Namespace]ProjectList{}
	for i, v := range pl {
		if filterOrgProject(pl[i]) {
			_, exists := m[*pl[i].Namespace]
			if !exists {
				m[*pl[i].Namespace] = ProjectList{}
			}
			v.Owner = nil
			v.Namespace = nil
			m[*pl[i].Namespace] = append(m[*pl[i].Namespace], v)
		}
	}

	npl := []NewOrgProjectList{}
	for k, v := range m {
		npl = append(npl, NewOrgProjectList{k, v})
	}

	return npl

}

func FilterDeployKeysByTitle(dks []DeployKey, filterFn func(string, string) bool, filterStr string) []DeployKey {
	ndks := []DeployKey{}
	for _, v := range dks {
		if filterFn(v.Title, filterStr) {
			ndks = append(ndks, v)
		}
	}

	return ndks
}

func Equals(a, b string) bool {
	return a == b
}
