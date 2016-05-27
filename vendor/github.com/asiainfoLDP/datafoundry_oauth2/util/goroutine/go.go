package goroutine

import (
	"sync"
)

func Go(n int, goFunc func(time int) interface{}, resultChan chan interface{}) {
	var wg sync.WaitGroup
	wg.Add(n)

	for gon := 1; gon <= n; gon++ {
		go func(wg *sync.WaitGroup, time int) {
			defer wg.Done()

			res := goFunc(time)

			if res != nil {
				resultChan <- res
			}

			return
		}(&wg, gon)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()
}
