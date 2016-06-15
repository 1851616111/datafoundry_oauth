package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"time"

	gitlabapi "github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/cache"
	etcdutil "github.com/asiainfoLDP/datafoundry_oauth2/util/etcd"
)

const (
	IdlerTimerSec     = 20
	CacheFrequencySec = 60
)

var scale uint8

//func init() {
//	go func() {
//		scale = uint8(initAvgIdle())
//		fmt.Printf("init cpu idle: %d\n", scale)
//	}()
//}

type pair struct {
	key   string
	value string
}

func runGitLabCacheController() {

	const (
		producerPeriod = 60 * time.Second
		consumerPeriod = 1000 * time.Millisecond
	)

	ret := make(chan interface{}, 1)
	reduceChan := make(chan string, 2000)

	producer := func() {
		if rsp, err := db.getDir("/df_service"); err != nil {
			log.Printf("controller looper %ds for gitlab orgs err %v\n", producerPeriod, err)
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
			m := new(map[string]string)
			json.Unmarshal([]byte(p), m)
			host, user, privateToken := (*m)["host"], (*m)["user"], (*m)["private_token"]
			if host == "" || user == "" || privateToken == "" {
				return
			}

			projects, err := glApi.Project(host, privateToken).ListProjects()
			if err != nil {
				fmt.Printf("controller loop get gitlab project err %v\n", err)
				return
			}

			if len(projects) == 0 {
				return
			}

			p_orgs := gitlabapi.ConverOrgProjects(projects)
			if len(p_orgs) > 0 {
				go CacheMan.HCacheObject("gitlab://"+host, "user_"+user+"@orgs_repos", p_orgs)
			}

			p_owner := gitlabapi.ConverOwnerProjects(projects)
			if len(p_owner) > 0 {
				go CacheMan.HCacheObject("gitlab://"+host, "user_"+user+"@owner_repos", p_owner)
			}

		default:
		}
	}

	cacher := cache.NewCacher(cache.ProducerFn(producer), cache.MiddlerFn(middler), cache.ConsumerFn(consumer), producerPeriod, 0, consumerPeriod)
	cacher.Run()
}

//func initAvgIdle() float32 {
//	done := make(chan struct{}, 1)
//	defer close(done)
//	newout := chanutil.TimeReader(time.Second*IdlerTimerSec, pprof.GetStat(pprof.Line_CPU, done))
//
//	return averageIdle(newout)
//}
//
//func averageIdle(c <-chan interface{}) float32 {
//	length := len(c)
//	sum := float32(0.0)
//
//	for i := range c {
//		if cpu, ok := i.(*pprof.CPU); ok {
//			sum += cpu.Idle
//		}
//	}
//
//	return sum / float32(length)
//}
