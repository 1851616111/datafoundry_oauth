package pprof

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

var (
	percentReg = regexp.MustCompile("[0-9]{1,3}.[0-9]{1,3}")
)

//CPU usage: 1.38% user, 13.88% sys, 84.72% idle
func cpuParser(s string) *CPU {
	if len(s) == 0 {
		return nil
	}

	l := percentReg.FindAllString(s, -1)
	if len(l) == 3 {
		user, _ := strconv.ParseFloat(l[0], 32)
		sys, _ := strconv.ParseFloat(l[1], 32)
		idle, _ := strconv.ParseFloat(l[2], 32)

		return &CPU{
			User: float32(user),
			Sys:  float32(sys),
			Idle: float32(idle),
		}

	}

	return nil
}

func readLine(line Line, r *bufio.Reader, out chan interface{}, done <-chan struct{}) error {
	if line >= 14 {
		return fmt.Errorf("param line must between 1 and 13")
	}

	lineName := nameIndexMapping[line]

	go func(out chan<- interface{}, done <-chan struct{}) {
		for {
			for i := 1; i <= 13; i++ {
				select {
				case <-done:
					close(out)
					return
				default:
					if i == int(line) {
						lineText, _, err := r.ReadLine()
						if err != nil {
							log.Printf("read line err %v", err)
							continue
						}

						//todo 使用反射根据lineName取出parse的func来调用, 这里只使用的cpu
						if lineName == "Cpu" {
							cpu := cpuParser(string(lineText))
							if cpu != nil {
								out <- cpu
							}

						}
						continue
					}
					r.ReadLine()
				}
			}
		}
	}(out, done)

	return nil
}
