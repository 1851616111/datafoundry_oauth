package main

type NanoRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    *Owner `json:"owner,omitempty"`
	Private  bool   `json:"private"`
}

type Org struct {
	Login string `json:"login"`
}

type Owner struct {
	Login string `json:"login"`
}

type Repos []NanoRepo

type NewRepoList struct {
	Owner
	Repos `json:"repos"`
}

func (list *Repos) Convert() []NewRepoList {
	m := map[Owner]Repos{}
	l := []NewRepoList{}
	if len(*list) == 0 {
		return l
	}

	for i, v := range *list {
		o := (*list)[i]
		o.Owner = nil
		if _, exist := m[*v.Owner]; !exist {
			m[*v.Owner] = Repos{}
		}
		m[*v.Owner] = append(m[*v.Owner], o)
	}

	for owner, list := range m {
		l = append(l, NewRepoList{
			Owner: owner,
			Repos: list,
		})
	}

	return l
}
