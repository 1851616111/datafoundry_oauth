package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"time"

	"github.com/asiainfoLDP/datafoundry_oauth2/util/cache"
	//chanutil "github.com/asiainfoLDP/datafoundry_oauth2/util/channel"
	etcdutil "github.com/asiainfoLDP/datafoundry_oauth2/util/etcd"
	//"github.com/asiainfoLDP/datafoundry_oauth2/util/pprof"
)

func runGitHubCacheController() {
	const loopPeriod = 180 * time.Second

	ret := make(chan interface{}, 1)
	reduceChan := make(chan string, 5000)

	producer := func() {
		if rsp, err := db.getDir("/oauth/users"); err != nil {
			log.Printf("controller looper %ds for github orgs err %v\n", loopPeriod, err)
			return
		} else {
			if rsp != nil {
				ret <- rsp
			}

		}
	}

	middler := func() {
		select {
		case rsp := <-ret:
			etcdutil.RangeNodeFunc(rsp.(*etcd.Response).Node, func(n *etcd.Node) {
				reduceChan <- n.Value
			})
		}
	}

	consumer := func() {
		select {
		case p := <-reduceChan:
			userInfo := map[string]string{}
			json.Unmarshal([]byte(p), &userInfo)

			if userInfo["credential_key"] == "" || userInfo["credential_value"] == "" {
				return
			}

			var repos *Repos
			var err error
			if repos, err = GetOwnerRepos(userInfo); err != nil {
				fmt.Printf("controller loop get github repos err %v\n", err)
				return
			}

			if len([]NanoRepo(*repos)) == 0 {
				return
			}

			go CacheMan.HCacheObject("www.github.com", "user_"+userInfo["user"]+"@owner_repos", repos.Convert())

			var orgs []Org
			if orgs, err = GetOwnerOrgs(userInfo); err != nil {
				fmt.Printf("controller cache github orgs err %v\n", err)
				return
			}

			if len(orgs) == 0 {
				return
			}

			repos = &Repos{}
			for _, v := range orgs {
				var l *Repos
				if l, err = GetOrgReps(userInfo, v.Login); err != nil {
					fmt.Printf("[GET]/github.com/user/orgs, get org %s info err %s", v.Login, err.Error())
					continue
				}

				if len(*l) == 0 {
					continue
				}

				*repos = append(*repos, *l...)
			}

			if len(*repos) == 0 {
				return
			}

			go CacheMan.HCacheObject("www.github.com", "user_"+userInfo["user"]+"@orgs_repos", repos.Convert())

		default:
		}
	}

	cacher := cache.NewCacher(cache.ProducerFn(producer), cache.MiddlerFn(middler), cache.ConsumerFn(consumer), loopPeriod, 0, 0)
	cacher.Run()
}
