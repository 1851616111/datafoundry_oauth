package cache

import (
	"github.com/asiainfoLDP/datafoundry_oauth2/util/wait"
	"time"
)

type ProducerFn func()
type MiddlerFn func()
type ConsumerFn func()

type Cacher interface {
	Run()
}

func (i *impl) Run() {
	go wait.Forever((func())(i.producer), i.pp)
	go wait.Forever((func())(i.middler), i.mp)
	go wait.Forever((func())(i.consumer), i.cp)
}

func NewCacher(pf ProducerFn, mf MiddlerFn, cf ConsumerFn, pp, mp, cp time.Duration) Cacher {
	return &impl{
		producer: pf,
		middler:  mf,
		consumer: cf,
		pp:       pp,
		mp:       mp,
		cp:       cp,
	}
}

type impl struct {
	producer ProducerFn
	middler  MiddlerFn
	consumer ConsumerFn
	pp       time.Duration
	mp       time.Duration
	cp       time.Duration
}
