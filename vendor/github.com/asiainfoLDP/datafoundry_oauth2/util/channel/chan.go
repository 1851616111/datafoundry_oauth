package channel

import "time"

func TimeReader(d time.Duration, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{}, cap(in))
	defer close(out)
	c := time.After(d)

	for {
		select {
		case <-c:
			return out
		default:

			//in case of in channel closed.
			v, ok := <-in
			if ok {
				out <- v
			} else {
				return out
			}
		}
	}
}
