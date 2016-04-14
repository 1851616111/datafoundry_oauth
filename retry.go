package main

//import (
//	"sync"
//	"time"
//)
//
//const WorkerNum = 3
//
//var Retry ErrRetryController
//
//type ErrRetryController struct {
//	sync.RWMutex
//	retry map[opt]retry
//	ticker map[opt]*time.Time
//	intervalSec int
//	increaseSec int
//}
//
//func (c *ErrRetryController) work() {
//
//	for k := range c.retry {
//		if exist := c.ticker[k]; !exist {
//
//		}
//	}
//}
//
//type retry struct {
//	opt
//	retryTimes int
//	reason error
//}
//
//type opt struct {
//	cmd string
//	pair
//}
//
//type pair struct {
//	key, value string
//}
//
//func NewErrRetryController() *ErrRetryController {
//	return &ErrRetryController{
//		retry:make(map[opt]retry , 10000),
//		intervalSec: 5,
//		increaseSec: 5,
//	}
//}
//
//func (c *ErrRetryController)Run() {
//	for i := 0 ; i < WorkerNum; i ++ {
//		go c.work()
//	}
//}
