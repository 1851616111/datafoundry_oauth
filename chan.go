package main

func Done(done <-chan struct{}, rc chan<- Result, code, bodyCode int, msg string) {
	select {
	case <-done:
		return
	default:
		rc <- Result{
			code:     code,
			bodyCode: bodyCode,
			msg:      msg,
		}
	}
}

type Result struct {
	code     int
	bodyCode int
	msg      string
}
