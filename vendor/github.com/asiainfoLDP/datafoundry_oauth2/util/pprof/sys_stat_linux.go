package pprof

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

//todo 使用cpu赫兹来进行使用率的统计
//top - 15:01:09 up 37 days, 16:53,  2 users,  load average: 1.29, 0.52, 0.23
//Tasks:   1 total,   0 running,   1 sleeping,   0 stopped,   0 zombie
//%Cpu(s): 25.5 us, 22.3 sy,  0.0 ni, 50.2 id,  0.0 wa,  0.0 hi,  1.1 si,  0.9 st
//KiB Mem :  3882760 total,  1121804 free,   223180 used,  2537776 buff/cache
//KiB Swap:  2097148 total,  2050732 free,    46416 used.  3151484 avail Mem
//
//PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
//4973 root      20   0  134184   8940   3720 S  94.3  0.2   0:34.50 datafoundry_oau

const (
	Line_Processes Line = iota + 1
	Line_Time
	Line_Load
	Line_CPU
	Line_SharedLibs
	Line_MemRegions
	Line_PhysMem
	Line_VM
	Line_Networks
	Line_Disk
	Line_namespace
	Line_Names
	Line_Statis
)

var nameIndexMapping = [14]string{1: "Processes", 3: "Load", 4: "Cpu", 5: "SharedLibs", 6: "MemRegions",
	7: "PhysMem", 8: "VM", 9: "Networks", 10: "Disks", 12: "Names", 13: "Statis"}

type Line uint8

func GetStat(line Line, done <-chan struct{}) chan interface{} {

	pid := os.Getpid()
	cmd := exec.Command("top", "-p", fmt.Sprintf("%d", pid))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	r := bufio.NewReader(stdout)
	cmd.Start()

	out := make(chan interface{}, 300)
	readLine(line, r, out, done)

	return out
}
