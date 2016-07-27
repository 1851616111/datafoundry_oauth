package main

import (
	"sort"
)

type GitHubWebHook struct {
	Id int `json:"id"`
	GitHubWebHookOption
}

type GitHubWebHookOption struct {
	Name   string   `json:"name",omitempty`
	Active bool     `json:"active"`
	Events []string `json:"events,omitempty"`
	Config CConfig  `json:"config,omitempty"`
}

func (o *GitHubWebHookOption) DefaultOption() {
	o.Active = true
	o.Name = "web"
	o.Events = []string{"push", "pull_request", "status"}
	o.Config.ContentType = "json"
}

type CConfig struct {
	Url string `json:"url,omitempty"`
	ContentType string `json:"content_type"`
}

func gitHubWebHookchanged(old, new *GitHubWebHookOption) bool {

	if !stringArrayEquals(old.Events, new.Events) {
		return true
	}

	if old.Config.Url != new.Config.Url {
		return true
	}

	if old.Active != new.Active {
		return true
	}

	if old.Config != new.Config {
		return true
	}

	return false
}

func stringArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
