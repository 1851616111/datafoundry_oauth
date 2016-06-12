package pprof

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

//Processes: 243 total, 2 running, 5 stuck, 236 sleeping, 1063 threads
//2016/06/06 15:48:00
//Load Avg: 2.41, 2.42, 2.40
//CPU usage: 1.38% user, 13.88% sys, 84.72% idle
//SharedLibs: 163M resident, 22M data, 23M linkedit.
//MemRegions: 48168 total, 3444M resident, 121M private, 1529M shared.
//PhysMem: 11G used (1984M wired), 5501M unused.
//VM: 1234G vsize, 533M framework vsize, 0(0) swapins, 0(0) swapouts.
//Networks: packets: 530655/506M in, 345744/33M out.
//Disks: 229327/6225M read, 221809/15G written.
//
//PID   COMMAND %CPU TIME     #TH #WQ #PORTS MEM   PURG CMPRS PGRP PPID STATE    BOOSTS %CPU_ME %CPU_OTHRS UID FAULTS COW MSGSENT MSGRECV SYSBSD SYSMACH CSW PAGEINS IDLEW POWER USER    #MREGS RPRVT VPRVT VSIZE KPRVT KSHRD
//3770  pprof   0.0  00:00.00 5   0   17+    696K+ 0B   0B    3770 3630 sleeping *0[1+] 0.00000 0.00000    501 590+   53+ 22+     10+     197+   59+     76+ 0       0     0.0   michael N/A    N/A   N/A   N/A   N/A   N/A

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
	cmd := exec.Command("top", "-pid", fmt.Sprintf("%d", pid))

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
